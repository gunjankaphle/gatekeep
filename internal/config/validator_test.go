package config

import (
	"testing"
)

func TestValidator_Validate_ValidConfig(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Roles: []Role{
			{Name: "ROLE1"},
			{Name: "ROLE2", ParentRoles: []string{"ROLE1"}},
		},
		Users: []User{
			{Name: "user@example.com", Roles: []string{"ROLE1"}},
		},
		Databases: []Database{
			{
				Name: "DB1",
				Schemas: []Schema{
					{
						Name: "SCHEMA1",
						Tables: []Table{
							{
								Name: "TABLE1",
								Grants: []Grant{
									{ToRole: "ROLE1", Privileges: []string{PrivilegeSelect}},
								},
							},
						},
					},
				},
			},
		},
		Warehouses: []Warehouse{
			{
				Name: "WH1",
				Grants: []Grant{
					{ToRole: "ROLE1", Privileges: []string{PrivilegeUsage}},
				},
			},
		},
	}

	validator := NewValidator()
	err := validator.Validate(config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestValidator_Validate_MissingVersion(t *testing.T) {
	config := &Config{
		Roles: []Role{{Name: "ROLE1"}},
	}

	validator := NewValidator()
	err := validator.Validate(config)
	if err == nil {
		t.Fatal("Expected error for missing version")
	}
}

func TestValidator_Validate_DuplicateRoleName(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Roles: []Role{
			{Name: "DUPLICATE"},
			{Name: "DUPLICATE"},
		},
	}

	validator := NewValidator()
	err := validator.Validate(config)
	if err == nil {
		t.Fatal("Expected error for duplicate role name")
	}

	if !contains(err.Error(), "duplicate role name") {
		t.Errorf("Expected 'duplicate role name' error, got: %v", err)
	}
}

func TestValidator_Validate_NonExistentParentRole(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Roles: []Role{
			{Name: "ROLE1", ParentRoles: []string{"NONEXISTENT"}},
		},
	}

	validator := NewValidator()
	err := validator.Validate(config)
	if err == nil {
		t.Fatal("Expected error for non-existent parent role")
	}

	if !contains(err.Error(), "non-existent parent role") {
		t.Errorf("Expected 'non-existent parent role' error, got: %v", err)
	}
}

func TestValidator_Validate_CyclicDependency(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Roles: []Role{
			{Name: "ROLE1", ParentRoles: []string{"ROLE2"}},
			{Name: "ROLE2", ParentRoles: []string{"ROLE1"}},
		},
	}

	validator := NewValidator()
	err := validator.Validate(config)
	if err == nil {
		t.Fatal("Expected error for cyclic dependency")
	}

	if !contains(err.Error(), "cyclic dependency") {
		t.Errorf("Expected 'cyclic dependency' error, got: %v", err)
	}
}

func TestValidator_Validate_ComplexCyclicDependency(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Roles: []Role{
			{Name: "ROLE1", ParentRoles: []string{"ROLE2"}},
			{Name: "ROLE2", ParentRoles: []string{"ROLE3"}},
			{Name: "ROLE3", ParentRoles: []string{"ROLE1"}},
		},
	}

	validator := NewValidator()
	err := validator.Validate(config)
	if err == nil {
		t.Fatal("Expected error for cyclic dependency")
	}

	if !contains(err.Error(), "cyclic dependency") {
		t.Errorf("Expected 'cyclic dependency' error, got: %v", err)
	}
}

func TestValidator_Validate_InvalidUserEmail(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Roles:   []Role{{Name: "ROLE1"}},
		Users: []User{
			{Name: "invalid_email", Roles: []string{"ROLE1"}},
		},
	}

	validator := NewValidator()
	err := validator.Validate(config)
	if err == nil {
		t.Fatal("Expected error for invalid email")
	}

	if !contains(err.Error(), "valid email") {
		t.Errorf("Expected 'valid email' error, got: %v", err)
	}
}

func TestValidator_Validate_UserNoRoles(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Roles:   []Role{{Name: "ROLE1"}},
		Users: []User{
			{Name: "user@example.com", Roles: []string{}},
		},
	}

	validator := NewValidator()
	err := validator.Validate(config)
	if err == nil {
		t.Fatal("Expected error for user with no roles")
	}

	if !contains(err.Error(), "at least one role") {
		t.Errorf("Expected 'at least one role' error, got: %v", err)
	}
}

func TestValidator_Validate_UserNonExistentRole(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Roles:   []Role{{Name: "ROLE1"}},
		Users: []User{
			{Name: "user@example.com", Roles: []string{"NONEXISTENT"}},
		},
	}

	validator := NewValidator()
	err := validator.Validate(config)
	if err == nil {
		t.Fatal("Expected error for non-existent role reference")
	}

	if !contains(err.Error(), "non-existent role") {
		t.Errorf("Expected 'non-existent role' error, got: %v", err)
	}
}

func TestValidator_Validate_DuplicateUserName(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Roles:   []Role{{Name: "ROLE1"}},
		Users: []User{
			{Name: "user@example.com", Roles: []string{"ROLE1"}},
			{Name: "user@example.com", Roles: []string{"ROLE1"}},
		},
	}

	validator := NewValidator()
	err := validator.Validate(config)
	if err == nil {
		t.Fatal("Expected error for duplicate user name")
	}

	if !contains(err.Error(), "duplicate user name") {
		t.Errorf("Expected 'duplicate user name' error, got: %v", err)
	}
}

func TestValidator_Validate_InvalidPrivilege(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Roles:   []Role{{Name: "ROLE1"}},
		Databases: []Database{
			{
				Name: "DB1",
				Schemas: []Schema{
					{
						Name: "SCHEMA1",
						Tables: []Table{
							{
								Name: "TABLE1",
								Grants: []Grant{
									{ToRole: "ROLE1", Privileges: []string{"INVALID_PRIVILEGE"}},
								},
							},
						},
					},
				},
			},
		},
	}

	validator := NewValidator()
	err := validator.Validate(config)
	if err == nil {
		t.Fatal("Expected error for invalid privilege")
	}

	if !contains(err.Error(), "invalid privilege") {
		t.Errorf("Expected 'invalid privilege' error, got: %v", err)
	}
}

func TestValidator_Validate_GrantNonExistentRole(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Roles:   []Role{{Name: "ROLE1"}},
		Warehouses: []Warehouse{
			{
				Name: "WH1",
				Grants: []Grant{
					{ToRole: "NONEXISTENT", Privileges: []string{PrivilegeUsage}},
				},
			},
		},
	}

	validator := NewValidator()
	err := validator.Validate(config)
	if err == nil {
		t.Fatal("Expected error for grant to non-existent role")
	}

	if !contains(err.Error(), "non-existent role") {
		t.Errorf("Expected 'non-existent role' error, got: %v", err)
	}
}

func TestValidator_Validate_DuplicateDatabaseName(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Databases: []Database{
			{Name: "DB1"},
			{Name: "DB1"},
		},
	}

	validator := NewValidator()
	err := validator.Validate(config)
	if err == nil {
		t.Fatal("Expected error for duplicate database name")
	}

	if !contains(err.Error(), "duplicate database name") {
		t.Errorf("Expected 'duplicate database name' error, got: %v", err)
	}
}

func TestValidator_ValidTablePrivileges(t *testing.T) {
	privs := ValidTablePrivileges()
	expected := []string{"SELECT", "INSERT", "UPDATE", "DELETE"}

	if len(privs) != len(expected) {
		t.Errorf("Expected %d table privileges, got %d", len(expected), len(privs))
	}

	for _, exp := range expected {
		found := false
		for _, priv := range privs {
			if priv == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected privilege %s not found", exp)
		}
	}
}

func TestValidator_ValidWarehousePrivileges(t *testing.T) {
	privs := ValidWarehousePrivileges()
	expected := []string{"USAGE", "OPERATE", "MONITOR", "MODIFY"}

	if len(privs) != len(expected) {
		t.Errorf("Expected %d warehouse privileges, got %d", len(expected), len(privs))
	}
}
