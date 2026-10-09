package diff

import (
	"strings"
	"testing"

	"github.com/yourusername/gatekeep/internal/config"
	"github.com/yourusername/gatekeep/internal/snowflake"
)

func TestGenericObjectsPlanAndIdempotency(t *testing.T) {
	cfg, err := config.NewParser().ParseFile("../../configs/generic-objects.yaml")
	if err != nil {
		t.Fatal(err)
	}
	state := snowflake.State{Roles: []snowflake.Role{{Name: "ANALYST_ROLE"}}}
	result, err := NewDiffer(SyncModeAdditive).ComputeDiff(Input{DesiredConfig: *cfg, ActualState: state})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := NewPlanner(result).GeneratePlan()
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 8 {
		t.Fatalf("expected eight grants, got %d", len(plan))
	}
	statements := []string{}
	for _, op := range plan {
		statements = append(statements, op.SQL)
	}
	sql := strings.Join(statements, "\n")
	for _, expected := range []string{`GRANT SELECT ON DYNAMIC TABLE "ANALYTICS_DB"."PUBLIC"."DAILY_REVENUE" TO ROLE "ANALYST_ROLE"`, `GRANT REFERENCES ON SEMANTIC VIEW "ANALYTICS_DB"."PUBLIC"."REVENUE_MODEL" TO ROLE "ANALYST_ROLE"`, `GRANT USAGE ON WAREHOUSE "ANALYTICS_WH" TO ROLE "ANALYST_ROLE"`} {
		if !strings.Contains(sql, expected) {
			t.Fatalf("missing %s in %s", expected, sql)
		}
	}
	for _, grant := range result.ObjectGrantsToAdd {
		state.Grants = append(state.Grants, snowflake.Grant{GrantedOn: strings.ReplaceAll(grant.ObjectType, " ", "_"), Name: grant.ObjectName, Privilege: grant.Privilege, GranteeName: grant.ToRole})
	}
	again, err := NewDiffer(SyncModeAdditive).ComputeDiff(Input{DesiredConfig: *cfg, ActualState: state})
	if err != nil {
		t.Fatal(err)
	}
	if !again.IsEmpty() {
		t.Fatalf("existing grants generated changes: %#v", again)
	}
}

func TestQuoteGenericObjectIdentifiers(t *testing.T) {
	result := &Result{ObjectGrantsToAdd: []ObjectGrant{{Privilege: "SELECT", ObjectType: "SEMANTIC VIEW", NameParts: []string{"DB.With.Dot", `Schema"Name`, "View"}, ToRole: `Role"Name`}}}
	plan, err := NewPlanner(result).GeneratePlan()
	if err != nil {
		t.Fatal(err)
	}
	expected := `GRANT SELECT ON SEMANTIC VIEW "DB.With.Dot"."Schema""Name"."View" TO ROLE "Role""Name"`
	if plan[0].SQL != expected {
		t.Fatalf("%s", plan[0].SQL)
	}
}
