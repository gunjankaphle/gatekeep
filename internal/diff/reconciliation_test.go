package diff

import (
	"reflect"
	"strings"
	"testing"

	"github.com/yourusername/gatekeep/internal/config"
	"github.com/yourusername/gatekeep/internal/snowflake"
)

func TestHierarchyDirectGrantIdempotency(t *testing.T) {
	cfg := config.Config{Roles: []config.Role{{Name: "CHILD", ParentRoles: []string{"PARENT", "PARENT"}}, {Name: "PARENT"}}}
	state := snowflake.State{Roles: []snowflake.Role{{Name: "CHILD"}, {Name: "PARENT"}}}
	first, err := NewDiffer(SyncModeAdditive).ComputeDiff(Input{DesiredConfig: cfg, ActualState: state})
	if err != nil {
		t.Fatal(err)
	}
	if len(first.RoleGrantsToAdd) != 1 {
		t.Fatalf("expected one missing edge: %+v", first)
	}
	// The reverse edge and non-role recipients do not satisfy the desired grant.
	state.Grants = []snowflake.Grant{
		{GrantedOn: "ROLE", Privilege: "USAGE", Name: "PARENT", GranteeName: "CHILD", GrantedTo: "ROLE"},
		{GrantedOn: "ROLE", Privilege: "USAGE", Name: "CHILD", GranteeName: "PARENT", GrantedTo: "USER"},
	}
	wrong, err := NewDiffer(SyncModeAdditive).ComputeDiff(Input{DesiredConfig: cfg, ActualState: state})
	if err != nil || len(wrong.RoleGrantsToAdd) != 1 {
		t.Fatalf("wrong edge matched: %+v %v", wrong, err)
	}
	state.Grants = append(state.Grants, snowflake.Grant{GrantedOn: "ROLE", Privilege: "USAGE", Name: "CHILD", GranteeName: "PARENT", GrantedTo: "ROLE"})
	second, err := NewDiffer(SyncModeAdditive).ComputeDiff(Input{DesiredConfig: cfg, ActualState: state})
	if err != nil || !second.IsEmpty() {
		t.Fatalf("repeat must be empty: %+v %v", second, err)
	}
}

func TestStrictScopeIsolationAndRemovedDesiredEntries(t *testing.T) {
	cfg := config.Config{
		Managed: &config.ManagedScope{Roles: []string{"CHILD", "PARENT"}, Users: []string{"managed@example.com"}, Objects: []config.ManagedObject{{Type: "DYNAMIC TABLE", Name: []string{"DB", "S", "T"}, Privileges: []string{"SELECT", "MONITOR"}}}},
		Roles:   []config.Role{{Name: "CHILD"}, {Name: "PARENT"}},
		Objects: []config.Object{{Type: "DYNAMIC TABLE", Name: []string{"DB", "S", "T"}, Grants: []config.Grant{{ToRole: "CHILD", Privileges: []string{"SELECT"}}}}},
	}
	state := snowflake.State{
		Roles: []snowflake.Role{{Name: "CHILD"}, {Name: "PARENT"}, {Name: "OUTSIDE"}},
		Users: []snowflake.User{{Name: "managed@example.com", Roles: []string{"CHILD", "CHILD", "OUTSIDE", "ACCOUNTADMIN"}}, {Name: "outside@example.com", Roles: []string{"CHILD"}}},
		Grants: []snowflake.Grant{
			{GrantedOn: "ROLE", Privilege: "USAGE", Name: "CHILD", GranteeName: "PARENT", GrantedTo: "ROLE"},
			{GrantedOn: "ROLE", Privilege: "USAGE", Name: "CHILD", GranteeName: "OUTSIDE", GrantedTo: "ROLE"},
			{GrantedOn: "ROLE", Privilege: "USAGE", Name: "OUTSIDE", GranteeName: "PARENT", GrantedTo: "ROLE"},
			{GrantedOn: "ROLE", Privilege: "OWNERSHIP", Name: "CHILD", GranteeName: "PARENT", GrantedTo: "ROLE"},
			{GrantedOn: "DYNAMIC_TABLE", Name: "DB.S.T", Privilege: "SELECT", GranteeName: "CHILD", GrantedTo: "ROLE"},
			{GrantedOn: "DYNAMIC_TABLE", Name: "DB.S.T", Privilege: "MONITOR", GranteeName: "CHILD", GrantedTo: "ROLE"},
			{GrantedOn: "DYNAMIC_TABLE", Name: "DB.S.T", Privilege: "MONITOR", GranteeName: "CHILD", GrantedTo: "ROLE"},
			{GrantedOn: "DYNAMIC_TABLE", Name: "DB.S.T", Privilege: "OPERATE", GranteeName: "CHILD", GrantedTo: "ROLE"},
			{GrantedOn: "DYNAMIC_TABLE", Name: "DB.S.T", Privilege: "OWNERSHIP", GranteeName: "CHILD", GrantedTo: "ROLE"},
			{GrantedOn: "DYNAMIC_TABLE", Name: "DB.S.T", Privilege: "SELECT", GranteeName: "PARENT", GrantedTo: "ROLE"},
			{GrantedOn: "DYNAMIC_TABLE", Name: "DB.S.OTHER", Privilege: "MONITOR", GranteeName: "CHILD", GrantedTo: "ROLE"},
			{GrantedOn: "DYNAMIC_TABLE", Name: "DB.S.T", Privilege: "MONITOR", GranteeName: "OUTSIDE", GrantedTo: "ROLE"},
			{GrantedOn: "DYNAMIC_TABLE", Name: "DB.S.T", Privilege: "MONITOR", GranteeName: "PARENT", GrantedTo: "DATABASE_ROLE"},
		},
	}
	result, err := NewDiffer(SyncModeStrict).ComputeDiff(Input{DesiredConfig: cfg, ActualState: state})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.RoleGrantsToRevoke, []RoleGrant{{Role: "CHILD", ToRole: "PARENT"}}) {
		t.Fatalf("role scope: %+v", result.RoleGrantsToRevoke)
	}
	if !reflect.DeepEqual(result.UserRoleGrantsToRevoke, []UserRoleGrant{{Role: "CHILD", ToUser: "managed@example.com"}}) {
		t.Fatalf("user scope: %+v", result.UserRoleGrantsToRevoke)
	}
	if len(result.ObjectGrantsToRevoke) != 2 || result.ObjectGrantsToRevoke[0].Privilege != "MONITOR" || result.ObjectGrantsToRevoke[1].ToRole != "PARENT" {
		t.Fatalf("object scope: %+v", result.ObjectGrantsToRevoke)
	}
	if len(result.ObjectGrantsToAdd) != 0 || len(result.RolesToDelete) != 0 {
		t.Fatalf("unexpected additions/deletions: %+v", result)
	}
	plan, err := NewPlanner(result).GeneratePlan()
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 4 {
		t.Fatalf("plan: %+v", plan)
	}
	for _, op := range plan {
		if !strings.HasPrefix(op.SQL, "REVOKE ") {
			t.Fatalf("unexpected mutation: %+v", op)
		}
	}
	// Once the scoped revocations are applied, unrelated grants remain and the next plan is empty.
	kept := []snowflake.Grant{}
	for _, grant := range state.Grants {
		if grant.GrantedOn == "ROLE" && grant.Privilege == "USAGE" && grant.Name == "CHILD" && grant.GranteeName == "PARENT" {
			continue
		}
		if grant.GrantedTo == "ROLE" && grant.Name == "DB.S.T" && ((grant.GranteeName == "CHILD" && grant.Privilege == "MONITOR") || (grant.GranteeName == "PARENT" && grant.Privilege == "SELECT")) {
			continue
		}
		kept = append(kept, grant)
	}
	state.Grants = kept
	state.Users[0].Roles = []string{"OUTSIDE", "ACCOUNTADMIN"}
	again, err := NewDiffer(SyncModeStrict).ComputeDiff(Input{DesiredConfig: cfg, ActualState: state})
	if err != nil || !again.IsEmpty() {
		t.Fatalf("second strict diff: %+v %v", again, err)
	}
	additive, err := NewDiffer(SyncModeAdditive).ComputeDiff(Input{DesiredConfig: cfg, ActualState: state})
	if err != nil || !additive.IsEmpty() {
		t.Fatalf("additive diff: %+v %v", additive, err)
	}
}

func TestStrictRejectsMissingOrUnsafeScope(t *testing.T) {
	for _, scope := range []*config.ManagedScope{nil, {}, {Roles: []string{"ACCOUNTADMIN"}}, {Roles: []string{"GLOBALORGADMIN"}}, {Roles: []string{"R"}, Objects: []config.ManagedObject{{Type: "TABLE", Name: []string{"DB", "S", "T"}, Privileges: []string{"OWNERSHIP"}}}}} {
		if _, err := NewDiffer(SyncModeStrict).ComputeDiff(Input{DesiredConfig: config.Config{Managed: scope}}); err == nil {
			t.Fatalf("accepted unsafe scope: %+v", scope)
		}
	}
}

func TestObjectGrantIgnoresNonRoleRecipients(t *testing.T) {
	desired := []ObjectGrant{{ObjectType: "TABLE", ObjectName: "DB.S.T", Privilege: "SELECT", ToRole: "R"}}
	actual := []snowflake.Grant{{GrantedOn: "TABLE", Name: "DB.S.T", Privilege: "SELECT", GranteeName: "R", GrantedTo: "DATABASE_ROLE"}}
	if len(missingObjectGrants(desired, actual)) != 1 {
		t.Fatal("database role grant matched account role")
	}
}

func TestHierarchyAndUserSQLQuotesIdentifiers(t *testing.T) {
	result := &Result{
		RoleGrantsToAdd:        []RoleGrant{{Role: `CHILD"ROLE`, ToRole: `PARENT"ROLE`}},
		RoleGrantsToRevoke:     []RoleGrant{{Role: `CHILD"ROLE`, ToRole: `PARENT"ROLE`}},
		UserRoleGrantsToRevoke: []UserRoleGrant{{Role: `CHILD"ROLE`, ToUser: `USER"NAME`}},
	}
	plan, err := NewPlanner(result).GeneratePlan()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{`GRANT ROLE "CHILD""ROLE" TO ROLE "PARENT""ROLE"`, `REVOKE ROLE "CHILD""ROLE" FROM USER "USER""NAME"`, `REVOKE ROLE "CHILD""ROLE" FROM ROLE "PARENT""ROLE"`}
	for i, op := range plan {
		if op.SQL != expected[i] {
			t.Fatalf("unsafe identifier quoting: %s", op.SQL)
		}
	}
}

func TestUserAssignmentsDeduplicated(t *testing.T) {
	for _, exists := range []bool{false, true} {
		actual := map[string]snowflake.User{}
		if exists {
			actual["u@example.com"] = snowflake.User{Name: "u@example.com"}
		}
		desired := []config.User{{Name: "u@example.com", Roles: []string{"R", "R"}}}
		grants := NewDiffer(SyncModeAdditive).diffUserRoleGrantsToAdd(desired, actual)
		if len(grants) != 1 {
			t.Fatalf("duplicate desired assignment: %+v", grants)
		}
		actual["u@example.com"] = snowflake.User{Name: "u@example.com", Roles: []string{"R"}}
		if again := NewDiffer(SyncModeAdditive).diffUserRoleGrantsToAdd(desired, actual); len(again) != 0 {
			t.Fatalf("repeat assignment: %+v", again)
		}
	}
}
