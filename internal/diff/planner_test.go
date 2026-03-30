package diff

import (
	"strings"
	"testing"
)

func TestPlanner_GeneratePlan_Empty(t *testing.T) {
	diff := &DiffResult{}
	diff.ComputeSummary()

	planner := NewPlanner(diff)
	operations, err := planner.GeneratePlan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(operations) != 0 {
		t.Errorf("expected 0 operations for empty diff, got %d", len(operations))
	}
}

func TestPlanner_GeneratePlan_CreateRole(t *testing.T) {
	diff := &DiffResult{
		RolesToCreate: []string{"ANALYST_ROLE", "ENGINEER_ROLE"},
	}
	diff.ComputeSummary()

	planner := NewPlanner(diff)
	operations, err := planner.GeneratePlan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(operations) != 2 {
		t.Errorf("expected 2 operations, got %d", len(operations))
	}

	for _, op := range operations {
		if op.Type != OpCreateRole {
			t.Errorf("expected CREATE_ROLE operation, got %s", op.Type)
		}

		if !strings.Contains(op.SQL, "CREATE ROLE") {
			t.Errorf("expected CREATE ROLE in SQL, got: %s", op.SQL)
		}
	}
}

func TestPlanner_GeneratePlan_RoleHierarchy(t *testing.T) {
	diff := &DiffResult{
		RolesToCreate: []string{"CHILD_ROLE", "PARENT_ROLE"},
		RoleGrantsToAdd: []RoleGrant{
			{Role: "CHILD_ROLE", ToRole: "PARENT_ROLE"},
		},
	}
	diff.ComputeSummary()

	planner := NewPlanner(diff)
	operations, err := planner.GeneratePlan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have: 2 CREATE ROLE + 1 GRANT ROLE
	if len(operations) != 3 {
		t.Errorf("expected 3 operations, got %d", len(operations))
	}

	// Verify CREATE ROLE operations come before GRANT ROLE
	var createRoleCount int
	var grantRoleFound bool
	for i, op := range operations {
		if op.Type == OpCreateRole {
			createRoleCount++
			if grantRoleFound {
				t.Error("CREATE ROLE operation found after GRANT ROLE")
			}
		}
		if op.Type == OpGrantRole {
			grantRoleFound = true
			if createRoleCount != 2 {
				t.Errorf("GRANT ROLE at index %d, but only %d CREATE ROLE operations before it", i, createRoleCount)
			}

			expected := `GRANT ROLE "CHILD_ROLE" TO ROLE "PARENT_ROLE"`
			if op.SQL != expected {
				t.Errorf("expected SQL: %s, got: %s", expected, op.SQL)
			}
		}
	}

	if createRoleCount != 2 {
		t.Errorf("expected 2 CREATE ROLE operations, got %d", createRoleCount)
	}

	if !grantRoleFound {
		t.Error("expected GRANT ROLE operation not found")
	}
}

func TestPlanner_GeneratePlan_ObjectGrants(t *testing.T) {
	diff := &DiffResult{
		ObjectGrantsToAdd: []ObjectGrant{
			{
				Privilege:  "SELECT",
				ObjectType: "TABLE",
				ObjectName: "DB.SCHEMA.TABLE1",
				ToRole:     "ANALYST_ROLE",
			},
			{
				Privilege:  "USAGE",
				ObjectType: "DATABASE",
				ObjectName: "DB",
				ToRole:     "ANALYST_ROLE",
			},
		},
	}
	diff.ComputeSummary()

	planner := NewPlanner(diff)
	operations, err := planner.GeneratePlan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(operations) != 2 {
		t.Errorf("expected 2 operations, got %d", len(operations))
	}

	for _, op := range operations {
		if op.Type != OpGrantObject {
			t.Errorf("expected GRANT_OBJECT operation, got %s", op.Type)
		}

		if !strings.Contains(op.SQL, "GRANT") {
			t.Errorf("expected GRANT in SQL, got: %s", op.SQL)
		}

		if !strings.Contains(op.SQL, "TO ROLE") {
			t.Errorf("expected TO ROLE in SQL, got: %s", op.SQL)
		}
	}
}

func TestPlanner_GeneratePlan_Users(t *testing.T) {
	diff := &DiffResult{
		UsersToCreate: []string{"john.doe@company.com"},
		UserRoleGrantsToAdd: []UserRoleGrant{
			{Role: "ANALYST_ROLE", ToUser: "john.doe@company.com"},
		},
	}
	diff.ComputeSummary()

	planner := NewPlanner(diff)
	operations, err := planner.GeneratePlan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have: 1 CREATE USER + 1 GRANT ROLE TO USER
	if len(operations) != 2 {
		t.Errorf("expected 2 operations, got %d", len(operations))
	}

	// Verify CREATE USER comes before GRANT
	var createUserFound, grantUserRoleFound bool
	for _, op := range operations {
		if op.Type == OpCreateUser {
			createUserFound = true
			if grantUserRoleFound {
				t.Error("CREATE USER operation found after GRANT USER ROLE")
			}

			expected := `CREATE USER IF NOT EXISTS "john.doe@company.com"`
			if op.SQL != expected {
				t.Errorf("expected SQL: %s, got: %s", expected, op.SQL)
			}
		}
		if op.Type == OpGrantUserRole {
			grantUserRoleFound = true

			expected := `GRANT ROLE "ANALYST_ROLE" TO USER "john.doe@company.com"`
			if op.SQL != expected {
				t.Errorf("expected SQL: %s, got: %s", expected, op.SQL)
			}
		}
	}

	if !createUserFound {
		t.Error("CREATE USER operation not found")
	}
	if !grantUserRoleFound {
		t.Error("GRANT USER ROLE operation not found")
	}
}

func TestPlanner_GeneratePlan_Revocations(t *testing.T) {
	diff := &DiffResult{
		UserRoleGrantsToRevoke: []UserRoleGrant{
			{Role: "OLD_ROLE", ToUser: "john.doe@company.com"},
		},
		ObjectGrantsToRevoke: []ObjectGrant{
			{
				Privilege:  "SELECT",
				ObjectType: "TABLE",
				ObjectName: "DB.SCHEMA.TABLE1",
				ToRole:     "OLD_ROLE",
			},
		},
		RoleGrantsToRevoke: []RoleGrant{
			{Role: "CHILD_ROLE", ToRole: "PARENT_ROLE"},
		},
	}
	diff.ComputeSummary()

	planner := NewPlanner(diff)
	operations, err := planner.GeneratePlan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(operations) != 3 {
		t.Errorf("expected 3 revoke operations, got %d", len(operations))
	}

	// Verify all are revoke operations
	revokeTypes := map[OperationType]bool{
		OpRevokeUserRole: false,
		OpRevokeObject:   false,
		OpRevokeRole:     false,
	}

	for _, op := range operations {
		if _, isRevoke := revokeTypes[op.Type]; !isRevoke {
			t.Errorf("expected revoke operation, got %s", op.Type)
		}
		revokeTypes[op.Type] = true

		if !strings.Contains(op.SQL, "REVOKE") {
			t.Errorf("expected REVOKE in SQL, got: %s", op.SQL)
		}
	}

	for opType, found := range revokeTypes {
		if !found {
			t.Errorf("expected %s operation not found", opType)
		}
	}
}

func TestPlanner_GeneratePlan_DeleteRoles(t *testing.T) {
	diff := &DiffResult{
		RolesToDelete: []string{"OLD_ROLE", "UNUSED_ROLE"},
	}
	diff.ComputeSummary()

	planner := NewPlanner(diff)
	operations, err := planner.GeneratePlan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(operations) != 2 {
		t.Errorf("expected 2 operations, got %d", len(operations))
	}

	for _, op := range operations {
		if op.Type != OpDeleteRole {
			t.Errorf("expected DELETE_ROLE operation, got %s", op.Type)
		}

		if !strings.Contains(op.SQL, "DROP ROLE") {
			t.Errorf("expected DROP ROLE in SQL, got: %s", op.SQL)
		}
	}
}

func TestPlanner_GeneratePlan_CorrectOrdering(t *testing.T) {
	// Complex scenario testing correct phase ordering
	diff := &DiffResult{
		DatabasesToCreate:  []string{"NEW_DB"},
		WarehousesToCreate: []string{"NEW_WH"},
		RolesToCreate:      []string{"NEW_ROLE"},
		RoleGrantsToAdd: []RoleGrant{
			{Role: "CHILD_ROLE", ToRole: "NEW_ROLE"},
		},
		ObjectGrantsToAdd: []ObjectGrant{
			{Privilege: "USAGE", ObjectType: "DATABASE", ObjectName: "NEW_DB", ToRole: "NEW_ROLE"},
		},
		UsersToCreate: []string{"new.user@company.com"},
		UserRoleGrantsToAdd: []UserRoleGrant{
			{Role: "NEW_ROLE", ToUser: "new.user@company.com"},
		},
		UserRoleGrantsToRevoke: []UserRoleGrant{
			{Role: "OLD_ROLE", ToUser: "old.user@company.com"},
		},
		RolesToDelete: []string{"DELETED_ROLE"},
	}
	diff.ComputeSummary()

	planner := NewPlanner(diff)
	operations, err := planner.GeneratePlan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify operation order by phase
	expectedPhases := []OperationType{
		OpCreateDatabase,   // Phase 1: Resources
		OpCreateWarehouse,  // Phase 1: Resources
		OpCreateRole,       // Phase 2: Roles
		OpGrantRole,        // Phase 3: Role hierarchy
		OpGrantObject,      // Phase 4: Object permissions
		OpCreateUser,       // Phase 5: Users
		OpGrantUserRole,    // Phase 6: User role grants
		OpRevokeUserRole,   // Phase 7: Revocations
		OpDeleteRole,       // Phase 8: Delete roles
	}

	if len(operations) != len(expectedPhases) {
		t.Errorf("expected %d operations, got %d", len(expectedPhases), len(operations))
	}

	for i, op := range operations {
		if i >= len(expectedPhases) {
			break
		}
		if op.Type != expectedPhases[i] {
			t.Errorf("operation %d: expected type %s, got %s", i, expectedPhases[i], op.Type)
		}
	}
}

func TestPlanner_GeneratePlan_SystemRoles_NotDeleted(t *testing.T) {
	diff := &DiffResult{
		RolesToDelete: []string{"ACCOUNTADMIN", "CUSTOM_ROLE", "PUBLIC"},
	}
	diff.ComputeSummary()

	planner := NewPlanner(diff)
	operations, err := planner.GeneratePlan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only generate DROP for CUSTOM_ROLE, skip system roles
	if len(operations) != 1 {
		t.Errorf("expected 1 operation (CUSTOM_ROLE only), got %d", len(operations))
	}

	if operations[0].Target != "CUSTOM_ROLE" {
		t.Errorf("expected CUSTOM_ROLE to be deleted, got %s", operations[0].Target)
	}
}

func TestFormatPlan(t *testing.T) {
	operations := []SQLOperation{
		{Type: OpCreateRole, SQL: `CREATE ROLE "TEST_ROLE"`, Target: "TEST_ROLE"},
		{Type: OpGrantObject, SQL: `GRANT SELECT ON TABLE "DB.SCHEMA.TABLE" TO ROLE "TEST_ROLE"`, Target: "DB.SCHEMA.TABLE"},
	}

	output := FormatPlan(operations)

	if !strings.Contains(output, "Execution Plan") {
		t.Error("expected 'Execution Plan' in output")
	}

	if !strings.Contains(output, "CREATE ROLE") {
		t.Error("expected CREATE ROLE in output")
	}

	if !strings.Contains(output, "GRANT SELECT") {
		t.Error("expected GRANT SELECT in output")
	}

	if !strings.Contains(output, "CREATE ROLES") {
		t.Error("expected phase header 'CREATE ROLES' in output")
	}
}

func TestFormatPlan_Empty(t *testing.T) {
	output := FormatPlan([]SQLOperation{})

	expected := "No operations to execute"
	if output != expected {
		t.Errorf("expected '%s', got '%s'", expected, output)
	}
}
