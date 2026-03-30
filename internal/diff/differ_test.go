package diff

import (
	"fmt"
	"testing"

	"github.com/yourusername/gatekeep/internal/config"
	"github.com/yourusername/gatekeep/internal/snowflake"
)

func TestDiffer_ComputeDiff_EmptyStates(t *testing.T) {
	differ := NewDiffer(SyncModeStrict)

	input := DiffInput{
		DesiredConfig: config.Config{},
		ActualState:   snowflake.State{},
		Mode:          SyncModeStrict,
	}

	result, err := differ.ComputeDiff(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsEmpty() {
		t.Error("expected empty diff for empty states")
	}

	if result.Summary.TotalOperations != 0 {
		t.Errorf("expected 0 total operations, got %d", result.Summary.TotalOperations)
	}
}

func TestDiffer_ComputeDiff_CreateRoles(t *testing.T) {
	differ := NewDiffer(SyncModeStrict)

	input := DiffInput{
		DesiredConfig: config.Config{
			Version: "1.0",
			Roles: []config.Role{
				{Name: "ANALYST_ROLE"},
				{Name: "ENGINEER_ROLE"},
			},
		},
		ActualState: snowflake.State{
			Roles: []snowflake.Role{
				{Name: "EXISTING_ROLE"},
			},
		},
		Mode: SyncModeStrict,
	}

	result, err := differ.ComputeDiff(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.RolesToCreate) != 2 {
		t.Errorf("expected 2 roles to create, got %d", len(result.RolesToCreate))
	}

	expectedRoles := map[string]bool{
		"ANALYST_ROLE":  false,
		"ENGINEER_ROLE": false,
	}

	for _, role := range result.RolesToCreate {
		if _, exists := expectedRoles[role]; !exists {
			t.Errorf("unexpected role to create: %s", role)
		}
		expectedRoles[role] = true
	}

	for role, found := range expectedRoles {
		if !found {
			t.Errorf("expected role %s to be created", role)
		}
	}
}

func TestDiffer_ComputeDiff_DeleteRoles_StrictMode(t *testing.T) {
	differ := NewDiffer(SyncModeStrict)

	input := DiffInput{
		DesiredConfig: config.Config{
			Version: "1.0",
			Roles: []config.Role{
				{Name: "KEPT_ROLE"},
			},
		},
		ActualState: snowflake.State{
			Roles: []snowflake.Role{
				{Name: "KEPT_ROLE"},
				{Name: "REMOVED_ROLE"},
			},
		},
		Mode: SyncModeStrict,
	}

	result, err := differ.ComputeDiff(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.RolesToDelete) != 1 {
		t.Errorf("expected 1 role to delete, got %d", len(result.RolesToDelete))
	}

	if result.RolesToDelete[0] != "REMOVED_ROLE" {
		t.Errorf("expected REMOVED_ROLE to be deleted, got %s", result.RolesToDelete[0])
	}
}

func TestDiffer_ComputeDiff_DeleteRoles_AdditiveMode(t *testing.T) {
	differ := NewDiffer(SyncModeAdditive)

	input := DiffInput{
		DesiredConfig: config.Config{
			Version: "1.0",
			Roles: []config.Role{
				{Name: "KEPT_ROLE"},
			},
		},
		ActualState: snowflake.State{
			Roles: []snowflake.Role{
				{Name: "KEPT_ROLE"},
				{Name: "REMOVED_ROLE"},
			},
		},
		Mode: SyncModeAdditive,
	}

	result, err := differ.ComputeDiff(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.RolesToDelete) != 0 {
		t.Errorf("expected 0 roles to delete in additive mode, got %d", len(result.RolesToDelete))
	}
}

func TestDiffer_ComputeDiff_RoleHierarchy(t *testing.T) {
	differ := NewDiffer(SyncModeStrict)

	input := DiffInput{
		DesiredConfig: config.Config{
			Version: "1.0",
			Roles: []config.Role{
				{
					Name:        "CHILD_ROLE",
					ParentRoles: []string{"PARENT_ROLE"},
				},
			},
		},
		ActualState: snowflake.State{
			Roles: []snowflake.Role{
				{Name: "CHILD_ROLE"},
				{Name: "PARENT_ROLE"},
			},
		},
		Mode: SyncModeStrict,
	}

	result, err := differ.ComputeDiff(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.RoleGrantsToAdd) != 1 {
		t.Errorf("expected 1 role grant to add, got %d", len(result.RoleGrantsToAdd))
	}

	grant := result.RoleGrantsToAdd[0]
	if grant.Role != "CHILD_ROLE" || grant.ToRole != "PARENT_ROLE" {
		t.Errorf("expected grant CHILD_ROLE to PARENT_ROLE, got %s to %s", grant.Role, grant.ToRole)
	}
}

func TestDiffer_ComputeDiff_ObjectGrants(t *testing.T) {
	differ := NewDiffer(SyncModeStrict)

	input := DiffInput{
		DesiredConfig: config.Config{
			Version: "1.0",
			Databases: []config.Database{
				{
					Name: "PROD_DB",
					Schemas: []config.Schema{
						{
							Name: "PUBLIC",
							Tables: []config.Table{
								{
									Name: "CUSTOMERS",
									Grants: []config.Grant{
										{
											ToRole:     "ANALYST_ROLE",
											Privileges: []string{"SELECT"},
										},
									},
								},
								{
									Name: "ORDERS",
									Grants: []config.Grant{
										{
											ToRole:     "ANALYST_ROLE",
											Privileges: []string{"SELECT", "INSERT"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		ActualState: snowflake.State{},
		Mode:        SyncModeStrict,
	}

	result, err := differ.ComputeDiff(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 3 object grants: CUSTOMERS SELECT + ORDERS SELECT + ORDERS INSERT
	if len(result.ObjectGrantsToAdd) != 3 {
		t.Errorf("expected 3 object grants, got %d", len(result.ObjectGrantsToAdd))
	}

	// Check table grants
	expectedGrants := map[string]bool{
		"TABLE:PROD_DB.PUBLIC.CUSTOMERS:SELECT:ANALYST_ROLE": false,
		"TABLE:PROD_DB.PUBLIC.ORDERS:SELECT:ANALYST_ROLE":    false,
		"TABLE:PROD_DB.PUBLIC.ORDERS:INSERT:ANALYST_ROLE":    false,
	}

	for _, grant := range result.ObjectGrantsToAdd {
		key := fmt.Sprintf("%s:%s:%s:%s", grant.ObjectType, grant.ObjectName, grant.Privilege, grant.ToRole)
		if _, exists := expectedGrants[key]; exists {
			expectedGrants[key] = true
		}
	}

	for key, found := range expectedGrants {
		if !found {
			t.Errorf("expected grant not found: %s", key)
		}
	}
}

func TestDiffer_ComputeDiff_Users(t *testing.T) {
	differ := NewDiffer(SyncModeStrict)

	input := DiffInput{
		DesiredConfig: config.Config{
			Version: "1.0",
			Users: []config.User{
				{
					Name:  "john.doe@company.com",
					Roles: []string{"ANALYST_ROLE"},
				},
			},
		},
		ActualState: snowflake.State{
			Users: []snowflake.User{},
		},
		Mode: SyncModeStrict,
	}

	result, err := differ.ComputeDiff(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.UsersToCreate) != 1 {
		t.Errorf("expected 1 user to create, got %d", len(result.UsersToCreate))
	}

	if result.UsersToCreate[0] != "john.doe@company.com" {
		t.Errorf("unexpected user to create: %s", result.UsersToCreate[0])
	}

	if len(result.UserRoleGrantsToAdd) != 1 {
		t.Errorf("expected 1 user role grant, got %d", len(result.UserRoleGrantsToAdd))
	}

	grant := result.UserRoleGrantsToAdd[0]
	if grant.Role != "ANALYST_ROLE" || grant.ToUser != "john.doe@company.com" {
		t.Errorf("expected grant ANALYST_ROLE to john.doe@company.com, got %s to %s", grant.Role, grant.ToUser)
	}
}

func TestDiffer_ComputeDiff_UserRoleRevoke_StrictMode(t *testing.T) {
	differ := NewDiffer(SyncModeStrict)

	input := DiffInput{
		DesiredConfig: config.Config{
			Version: "1.0",
			Users: []config.User{
				{
					Name:  "john.doe@company.com",
					Roles: []string{"ANALYST_ROLE"},
				},
			},
		},
		ActualState: snowflake.State{
			Users: []snowflake.User{
				{
					Name:  "john.doe@company.com",
					Roles: []string{"ANALYST_ROLE", "ENGINEER_ROLE"},
				},
			},
		},
		Mode: SyncModeStrict,
	}

	result, err := differ.ComputeDiff(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should revoke ENGINEER_ROLE from user
	if len(result.UserRoleGrantsToRevoke) != 1 {
		t.Errorf("expected 1 user role grant to revoke, got %d", len(result.UserRoleGrantsToRevoke))
	}

	revoke := result.UserRoleGrantsToRevoke[0]
	if revoke.Role != "ENGINEER_ROLE" || revoke.ToUser != "john.doe@company.com" {
		t.Errorf("expected revoke ENGINEER_ROLE from john.doe@company.com, got %s from %s", revoke.Role, revoke.ToUser)
	}
}

func TestDiffer_ComputeDiff_SystemRoles_NotDeleted(t *testing.T) {
	differ := NewDiffer(SyncModeStrict)

	input := DiffInput{
		DesiredConfig: config.Config{
			Version: "1.0",
			Roles:   []config.Role{},
		},
		ActualState: snowflake.State{
			Roles: []snowflake.Role{
				{Name: "ACCOUNTADMIN"},
				{Name: "SYSADMIN"},
				{Name: "PUBLIC"},
				{Name: "CUSTOM_ROLE"},
			},
		},
		Mode: SyncModeStrict,
	}

	result, err := differ.ComputeDiff(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only delete CUSTOM_ROLE, not system roles
	if len(result.RolesToDelete) != 1 {
		t.Errorf("expected 1 role to delete, got %d", len(result.RolesToDelete))
	}

	if result.RolesToDelete[0] != "CUSTOM_ROLE" {
		t.Errorf("expected CUSTOM_ROLE to be deleted, got %s", result.RolesToDelete[0])
	}
}

func TestDiffer_ComputeDiff_ComplexScenario(t *testing.T) {
	differ := NewDiffer(SyncModeStrict)

	input := DiffInput{
		DesiredConfig: config.Config{
			Version: "1.0",
			Roles: []config.Role{
				{Name: "DATA_READER"},
				{Name: "DATA_WRITER", ParentRoles: []string{"DATA_READER"}},
			},
			Users: []config.User{
				{Name: "analyst@company.com", Roles: []string{"DATA_READER"}},
				{Name: "engineer@company.com", Roles: []string{"DATA_WRITER"}},
			},
			Databases: []config.Database{
				{
					Name: "ANALYTICS_DB",
					Schemas: []config.Schema{
						{
							Name: "PUBLIC",
							Tables: []config.Table{
								{
									Name: "SALES",
									Grants: []config.Grant{
										{ToRole: "DATA_READER", Privileges: []string{"SELECT"}},
										{ToRole: "DATA_WRITER", Privileges: []string{"SELECT", "INSERT"}},
									},
								},
							},
						},
					},
				},
			},
		},
		ActualState: snowflake.State{
			Roles: []snowflake.Role{
				{Name: "DATA_READER"},
				{Name: "OLD_ROLE"},
			},
			Users: []snowflake.User{
				{Name: "analyst@company.com", Roles: []string{}},
			},
			Databases: []snowflake.Database{},
		},
		Mode: SyncModeStrict,
	}

	result, err := differ.ComputeDiff(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify summary
	if result.Summary.RolesCreated != 1 { // DATA_WRITER
		t.Errorf("expected 1 role created, got %d", result.Summary.RolesCreated)
	}

	if result.Summary.RolesDeleted != 1 { // OLD_ROLE
		t.Errorf("expected 1 role deleted, got %d", result.Summary.RolesDeleted)
	}

	if result.Summary.UsersCreated != 1 { // engineer@company.com
		t.Errorf("expected 1 user created, got %d", result.Summary.UsersCreated)
	}

	if result.Summary.DatabasesCreated != 1 { // ANALYTICS_DB
		t.Errorf("expected 1 database created, got %d", result.Summary.DatabasesCreated)
	}

	if result.Summary.ObjectGrantsAdded != 3 { // SELECT for DATA_READER, SELECT+INSERT for DATA_WRITER
		t.Errorf("expected 3 object grants added, got %d", result.Summary.ObjectGrantsAdded)
	}

	if result.Summary.RoleGrantsAdded != 1 { // DATA_WRITER -> DATA_READER
		t.Errorf("expected 1 role grant added, got %d", result.Summary.RoleGrantsAdded)
	}

	if result.Summary.TotalOperations == 0 {
		t.Error("expected non-zero total operations")
	}
}
