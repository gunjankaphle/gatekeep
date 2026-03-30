package diff

import (
	"fmt"
	"strings"

	"github.com/yourusername/gatekeep/internal/config"
	"github.com/yourusername/gatekeep/internal/snowflake"
)

// Differ computes differences between desired and actual state
type Differ struct {
	mode SyncMode
}

// NewDiffer creates a new Differ
func NewDiffer(mode SyncMode) *Differ {
	return &Differ{mode: mode}
}

// ComputeDiff compares desired config vs actual Snowflake state
func (d *Differ) ComputeDiff(input Input) (*Result, error) {
	result := &Result{}

	// Build lookup maps for actual state
	actualRoles := buildRoleMap(input.ActualState.Roles)
	actualUsers := buildUserMap(input.ActualState.Users)
	actualDatabases := buildDatabaseMap(input.ActualState.Databases)
	actualWarehouses := buildWarehouseMap(input.ActualState.Warehouses)

	// 1. Diff roles
	result.RolesToCreate = d.diffRolesToCreate(input.DesiredConfig.Roles, actualRoles)
	if d.mode == SyncModeStrict {
		result.RolesToDelete = d.diffRolesToDelete(input.DesiredConfig.Roles, actualRoles)
	}

	// 2. Diff role hierarchies (parent role grants)
	result.RoleGrantsToAdd = d.diffRoleGrantsToAdd(input.DesiredConfig.Roles, actualRoles)
	if d.mode == SyncModeStrict {
		result.RoleGrantsToRevoke = d.diffRoleGrantsToRevoke(input.DesiredConfig.Roles)
	}

	// 3. Diff object grants (permissions on databases, tables, warehouses, etc.)
	result.ObjectGrantsToAdd = d.diffObjectGrantsToAdd(input.DesiredConfig)
	if d.mode == SyncModeStrict {
		result.ObjectGrantsToRevoke = d.diffObjectGrantsToRevoke(input.DesiredConfig, input.ActualState)
	}

	// 4. Diff users
	result.UsersToCreate = d.diffUsersToCreate(input.DesiredConfig.Users, actualUsers)

	// 5. Diff user role assignments
	result.UserRoleGrantsToAdd = d.diffUserRoleGrantsToAdd(input.DesiredConfig.Users, actualUsers)
	if d.mode == SyncModeStrict {
		result.UserRoleGrantsToRevoke = d.diffUserRoleGrantsToRevoke(input.DesiredConfig.Users, actualUsers)
	}

	// 6. Diff databases and warehouses
	result.DatabasesToCreate = d.diffDatabasesToCreate(input.DesiredConfig.Databases, actualDatabases)
	result.WarehousesToCreate = d.diffWarehousesToCreate(input.DesiredConfig.Warehouses, actualWarehouses)

	// Compute summary
	result.ComputeSummary()

	return result, nil
}

// diffRolesToCreate finds roles that exist in desired but not in actual
func (d *Differ) diffRolesToCreate(desiredRoles []config.Role, actualRoles map[string]snowflake.Role) []string {
	var toCreate []string
	for _, role := range desiredRoles {
		if _, exists := actualRoles[role.Name]; !exists {
			toCreate = append(toCreate, role.Name)
		}
	}
	return toCreate
}

// diffRolesToDelete finds roles that exist in actual but not in desired
func (d *Differ) diffRolesToDelete(desiredRoles []config.Role, actualRoles map[string]snowflake.Role) []string {
	desiredMap := make(map[string]bool)
	for _, role := range desiredRoles {
		desiredMap[role.Name] = true
	}

	var toDelete []string
	for roleName := range actualRoles {
		// Skip system roles
		if isSystemRole(roleName) {
			continue
		}
		if !desiredMap[roleName] {
			toDelete = append(toDelete, roleName)
		}
	}
	return toDelete
}

// diffRoleGrantsToAdd finds role hierarchy grants to add
func (d *Differ) diffRoleGrantsToAdd(desiredRoles []config.Role, actualRoles map[string]snowflake.Role) []RoleGrant {
	var toAdd []RoleGrant

	for _, role := range desiredRoles {
		for _, parentRole := range role.ParentRoles {
			// Check if this grant already exists in actual state
			if !roleGrantExists(role.Name, parentRole, actualRoles) {
				toAdd = append(toAdd, RoleGrant{
					Role:   role.Name,
					ToRole: parentRole,
				})
			}
		}
	}

	return toAdd
}

// diffRoleGrantsToRevoke finds role hierarchy grants to revoke
//
//nolint:unparam // Placeholder - will be implemented when ReadGrants (VAR-13) is complete
func (d *Differ) diffRoleGrantsToRevoke(desiredRoles []config.Role) []RoleGrant {
	// Build map of desired role grants
	desiredGrants := make(map[string]map[string]bool) // role -> parent roles
	for _, role := range desiredRoles {
		desiredGrants[role.Name] = make(map[string]bool)
		for _, parent := range role.ParentRoles {
			desiredGrants[role.Name][parent] = true
		}
	}

	var toRevoke []RoleGrant
	// This would require actual grant data from Snowflake
	// For now, return empty (will be implemented when ReadGrants is complete)
	return toRevoke
}

// diffObjectGrantsToAdd finds object grants (permissions) to add
func (d *Differ) diffObjectGrantsToAdd(cfg config.Config) []ObjectGrant {
	var toAdd []ObjectGrant

	// Table-level grants
	for _, db := range cfg.Databases {
		for _, schema := range db.Schemas {
			for _, table := range schema.Tables {
				for _, grant := range table.Grants {
					for _, priv := range grant.Privileges {
						toAdd = append(toAdd, ObjectGrant{
							Privilege:  priv,
							ObjectType: "TABLE",
							ObjectName: fmt.Sprintf("%s.%s.%s", db.Name, schema.Name, table.Name),
							ToRole:     grant.ToRole,
						})
					}
				}
			}
		}
	}

	// Warehouse-level grants
	for _, wh := range cfg.Warehouses {
		for _, grant := range wh.Grants {
			for _, priv := range grant.Privileges {
				toAdd = append(toAdd, ObjectGrant{
					Privilege:  priv,
					ObjectType: "WAREHOUSE",
					ObjectName: wh.Name,
					ToRole:     grant.ToRole,
				})
			}
		}
	}

	return toAdd
}

// diffObjectGrantsToRevoke finds object grants to revoke
func (d *Differ) diffObjectGrantsToRevoke(cfg config.Config, actual snowflake.State) []ObjectGrant {
	// This requires ReadGrants() to be implemented
	// For now, return empty
	return []ObjectGrant{}
}

// diffUsersToCreate finds users that exist in desired but not in actual
func (d *Differ) diffUsersToCreate(desiredUsers []config.User, actualUsers map[string]snowflake.User) []string {
	var toCreate []string
	for _, user := range desiredUsers {
		if _, exists := actualUsers[user.Name]; !exists {
			toCreate = append(toCreate, user.Name)
		}
	}
	return toCreate
}

// diffUserRoleGrantsToAdd finds user role grants to add
func (d *Differ) diffUserRoleGrantsToAdd(desiredUsers []config.User, actualUsers map[string]snowflake.User) []UserRoleGrant {
	var toAdd []UserRoleGrant

	for _, user := range desiredUsers {
		actualUser, exists := actualUsers[user.Name]
		var actualRoles map[string]bool
		if exists {
			actualRoles = make(map[string]bool)
			for _, role := range actualUser.Roles {
				actualRoles[role] = true
			}
		}

		for _, role := range user.Roles {
			if !exists || !actualRoles[role] {
				toAdd = append(toAdd, UserRoleGrant{
					Role:   role,
					ToUser: user.Name,
				})
			}
		}
	}

	return toAdd
}

// diffUserRoleGrantsToRevoke finds user role grants to revoke
func (d *Differ) diffUserRoleGrantsToRevoke(desiredUsers []config.User, actualUsers map[string]snowflake.User) []UserRoleGrant {
	// Build map of desired user role grants
	desiredUserRoles := make(map[string]map[string]bool) // user -> roles
	for _, user := range desiredUsers {
		desiredUserRoles[user.Name] = make(map[string]bool)
		for _, role := range user.Roles {
			desiredUserRoles[user.Name][role] = true
		}
	}

	var toRevoke []UserRoleGrant
	for userName, actualUser := range actualUsers {
		desiredRoles, userInConfig := desiredUserRoles[userName]
		if !userInConfig {
			// User not in config - revoke all roles in strict mode
			for _, role := range actualUser.Roles {
				if !isSystemRole(role) {
					toRevoke = append(toRevoke, UserRoleGrant{
						Role:   role,
						ToUser: userName,
					})
				}
			}
		} else {
			// User in config - revoke roles not in desired state
			for _, role := range actualUser.Roles {
				if !desiredRoles[role] && !isSystemRole(role) {
					toRevoke = append(toRevoke, UserRoleGrant{
						Role:   role,
						ToUser: userName,
					})
				}
			}
		}
	}

	return toRevoke
}

// diffDatabasesToCreate finds databases to create
func (d *Differ) diffDatabasesToCreate(desiredDatabases []config.Database, actualDatabases map[string]snowflake.Database) []string {
	var toCreate []string
	for _, db := range desiredDatabases {
		if _, exists := actualDatabases[db.Name]; !exists {
			toCreate = append(toCreate, db.Name)
		}
	}
	return toCreate
}

// diffWarehousesToCreate finds warehouses to create
func (d *Differ) diffWarehousesToCreate(desiredWarehouses []config.Warehouse, actualWarehouses map[string]snowflake.Warehouse) []string {
	var toCreate []string
	for _, wh := range desiredWarehouses {
		if _, exists := actualWarehouses[wh.Name]; !exists {
			toCreate = append(toCreate, wh.Name)
		}
	}
	return toCreate
}

// Helper functions

func buildRoleMap(roles []snowflake.Role) map[string]snowflake.Role {
	m := make(map[string]snowflake.Role)
	for _, role := range roles {
		m[role.Name] = role
	}
	return m
}

func buildUserMap(users []snowflake.User) map[string]snowflake.User {
	m := make(map[string]snowflake.User)
	for _, user := range users {
		m[user.Name] = user
	}
	return m
}

func buildDatabaseMap(databases []snowflake.Database) map[string]snowflake.Database {
	m := make(map[string]snowflake.Database)
	for _, db := range databases {
		m[db.Name] = db
	}
	return m
}

func buildWarehouseMap(warehouses []snowflake.Warehouse) map[string]snowflake.Warehouse {
	m := make(map[string]snowflake.Warehouse)
	for _, wh := range warehouses {
		m[wh.Name] = wh
	}
	return m
}

func roleGrantExists(role, parentRole string, actualRoles map[string]snowflake.Role) bool {
	// This would require grant data from Snowflake
	// For now, assume it doesn't exist (always add)
	// Will be properly implemented when ReadGrants is complete
	return false
}

func isSystemRole(roleName string) bool {
	systemRoles := []string{
		"ACCOUNTADMIN",
		"SECURITYADMIN",
		"SYSADMIN",
		"USERADMIN",
		"PUBLIC",
		"ORGADMIN",
	}
	upper := strings.ToUpper(roleName)
	for _, sys := range systemRoles {
		if upper == sys {
			return true
		}
	}
	return false
}
