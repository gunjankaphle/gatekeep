package config

import (
	"fmt"
	"strings"
)

// Validator handles configuration validation
type Validator struct{}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates the entire configuration
func (v *Validator) Validate(config *Config) error {
	// Validate version
	if config.Version == "" {
		return fmt.Errorf("version is required")
	}

	// Build role map for reference checking
	roleMap := make(map[string]bool)
	for _, role := range config.Roles {
		if role.Name == "" {
			return fmt.Errorf("role name is required")
		}
		if roleMap[role.Name] {
			return fmt.Errorf("duplicate role name: %s", role.Name)
		}
		roleMap[role.Name] = true
	}

	// Validate role parent references and check for cycles
	if err := v.validateRoleHierarchy(config.Roles, roleMap); err != nil {
		return err
	}

	// Validate users
	if err := v.validateUsers(config.Users, roleMap); err != nil {
		return err
	}

	// Validate databases
	if err := v.validateDatabases(config.Databases, roleMap); err != nil {
		return err
	}

	// Validate warehouses
	if err := v.validateWarehouses(config.Warehouses, roleMap); err != nil {
		return err
	}

	return nil
}

// validateRoleHierarchy validates role parent references and detects cycles
func (v *Validator) validateRoleHierarchy(roles []Role, roleMap map[string]bool) error {
	for _, role := range roles {
		// Check parent role references
		for _, parentRole := range role.ParentRoles {
			if !roleMap[parentRole] {
				return fmt.Errorf("role %s references non-existent parent role: %s", role.Name, parentRole)
			}
		}

		// Check for cycles
		if err := v.detectCycle(role.Name, roles, make(map[string]bool)); err != nil {
			return err
		}
	}

	return nil
}

// detectCycle detects cyclic dependencies in role hierarchy using DFS
func (v *Validator) detectCycle(roleName string, roles []Role, visited map[string]bool) error {
	if visited[roleName] {
		return fmt.Errorf("cyclic dependency detected in role hierarchy involving: %s", roleName)
	}

	visited[roleName] = true

	// Find the role
	var currentRole *Role
	for i := range roles {
		if roles[i].Name == roleName {
			currentRole = &roles[i]
			break
		}
	}

	if currentRole == nil {
		return nil
	}

	// Check all parent roles
	for _, parentRole := range currentRole.ParentRoles {
		if err := v.detectCycle(parentRole, roles, visited); err != nil {
			return err
		}
	}

	delete(visited, roleName)
	return nil
}

// validateUsers validates user configurations
func (v *Validator) validateUsers(users []User, roleMap map[string]bool) error {
	userMap := make(map[string]bool)

	for _, user := range users {
		if user.Name == "" {
			return fmt.Errorf("user name is required")
		}

		if userMap[user.Name] {
			return fmt.Errorf("duplicate user name: %s", user.Name)
		}
		userMap[user.Name] = true

		// Validate email format (basic check)
		if !strings.Contains(user.Name, "@") {
			return fmt.Errorf("user name must be a valid email: %s", user.Name)
		}

		// Check that user has at least one role
		if len(user.Roles) == 0 {
			return fmt.Errorf("user %s must have at least one role", user.Name)
		}

		// Validate role references
		for _, roleName := range user.Roles {
			if !roleMap[roleName] {
				return fmt.Errorf("user %s references non-existent role: %s", user.Name, roleName)
			}
		}
	}

	return nil
}

// validateDatabases validates database configurations
func (v *Validator) validateDatabases(databases []Database, roleMap map[string]bool) error {
	dbMap := make(map[string]bool)

	for _, db := range databases {
		if db.Name == "" {
			return fmt.Errorf("database name is required")
		}

		if dbMap[db.Name] {
			return fmt.Errorf("duplicate database name: %s", db.Name)
		}
		dbMap[db.Name] = true

		// Validate schemas
		if err := v.validateSchemas(db.Name, db.Schemas, roleMap); err != nil {
			return err
		}
	}

	return nil
}

// validateSchemas validates schema configurations
func (v *Validator) validateSchemas(dbName string, schemas []Schema, roleMap map[string]bool) error {
	schemaMap := make(map[string]bool)

	for _, schema := range schemas {
		if schema.Name == "" {
			return fmt.Errorf("schema name is required in database %s", dbName)
		}

		if schemaMap[schema.Name] {
			return fmt.Errorf("duplicate schema name in database %s: %s", dbName, schema.Name)
		}
		schemaMap[schema.Name] = true

		// Validate tables
		if err := v.validateTables(dbName, schema.Name, schema.Tables, roleMap); err != nil {
			return err
		}
	}

	return nil
}

// validateTables validates table configurations
func (v *Validator) validateTables(dbName, schemaName string, tables []Table, roleMap map[string]bool) error {
	tableMap := make(map[string]bool)

	for _, table := range tables {
		if table.Name == "" {
			return fmt.Errorf("table name is required in %s.%s", dbName, schemaName)
		}

		if tableMap[table.Name] {
			return fmt.Errorf("duplicate table name in %s.%s: %s", dbName, schemaName, table.Name)
		}
		tableMap[table.Name] = true

		// Validate grants
		if err := v.validateGrants(table.Grants, roleMap, ValidTablePrivileges()); err != nil {
			return fmt.Errorf("invalid grant for table %s.%s.%s: %w", dbName, schemaName, table.Name, err)
		}
	}

	return nil
}

// validateWarehouses validates warehouse configurations
func (v *Validator) validateWarehouses(warehouses []Warehouse, roleMap map[string]bool) error {
	warehouseMap := make(map[string]bool)

	for _, warehouse := range warehouses {
		if warehouse.Name == "" {
			return fmt.Errorf("warehouse name is required")
		}

		if warehouseMap[warehouse.Name] {
			return fmt.Errorf("duplicate warehouse name: %s", warehouse.Name)
		}
		warehouseMap[warehouse.Name] = true

		// Validate grants
		if err := v.validateGrants(warehouse.Grants, roleMap, ValidWarehousePrivileges()); err != nil {
			return fmt.Errorf("invalid grant for warehouse %s: %w", warehouse.Name, err)
		}
	}

	return nil
}

// validateGrants validates grant configurations
func (v *Validator) validateGrants(grants []Grant, roleMap map[string]bool, validPrivileges []string) error {
	for _, grant := range grants {
		if grant.ToRole == "" {
			return fmt.Errorf("grant to_role is required")
		}

		if !roleMap[grant.ToRole] {
			return fmt.Errorf("grant references non-existent role: %s", grant.ToRole)
		}

		if len(grant.Privileges) == 0 {
			return fmt.Errorf("grant must have at least one privilege")
		}

		// Validate privilege names
		validPrivMap := make(map[string]bool)
		for _, priv := range validPrivileges {
			validPrivMap[priv] = true
		}

		for _, priv := range grant.Privileges {
			if !validPrivMap[priv] {
				return fmt.Errorf("invalid privilege: %s (valid: %v)", priv, validPrivileges)
			}
		}
	}

	return nil
}
