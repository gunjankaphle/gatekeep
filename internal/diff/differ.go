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
	if d.mode != SyncModeAdditive && d.mode != SyncModeStrict {
		return nil, fmt.Errorf("unknown sync mode: %s", d.mode)
	}
	if d.mode == SyncModeStrict {
		if input.DesiredConfig.Managed == nil {
			return nil, fmt.Errorf("strict mode requires an explicit managed scope")
		}
		if err := config.ValidateManagedScope(input.DesiredConfig.Managed); err != nil {
			return nil, err
		}
	}

	// Build lookup maps for actual state
	actualRoles := buildRoleMap(input.ActualState.Roles)
	actualUsers := buildUserMap(input.ActualState.Users)
	actualDatabases := buildDatabaseMap(input.ActualState.Databases)
	actualWarehouses := buildWarehouseMap(input.ActualState.Warehouses)

	// 1. Diff roles
	result.RolesToCreate = d.diffRolesToCreate(input.DesiredConfig.Roles, actualRoles)
	// Strict mode reconciles permissions only; omitted roles are never dropped.

	// 2. Diff role hierarchies (parent role grants)
	result.RoleGrantsToAdd = d.diffRoleGrantsToAdd(input.DesiredConfig.Roles, input.ActualState.Grants)
	if d.mode == SyncModeStrict {
		result.RoleGrantsToRevoke = scopedRoleRevocations(input.DesiredConfig, input.ActualState.Grants)
	}

	// 3. Diff object grants (permissions on databases, tables, warehouses, etc.)
	result.ObjectGrantsToAdd = missingObjectGrants(d.diffObjectGrantsToAdd(input.DesiredConfig), input.ActualState.Grants)
	if d.mode == SyncModeStrict {
		result.ObjectGrantsToRevoke = scopedObjectRevocations(input.DesiredConfig, input.ActualState.Grants, d.diffObjectGrantsToAdd(input.DesiredConfig))
	}

	// 4. Diff users
	result.UsersToCreate = d.diffUsersToCreate(input.DesiredConfig.Users, actualUsers)

	// 5. Diff user role assignments
	result.UserRoleGrantsToAdd = d.diffUserRoleGrantsToAdd(input.DesiredConfig.Users, actualUsers)
	if d.mode == SyncModeStrict {
		result.UserRoleGrantsToRevoke = scopedUserRevocations(input.DesiredConfig, input.ActualState.Users)
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

// diffRoleGrantsToAdd compares direct hierarchy grants and deduplicates desired edges.
func (d *Differ) diffRoleGrantsToAdd(desiredRoles []config.Role, actual []snowflake.Grant) []RoleGrant {
	existing := roleEdges(actual)
	var result []RoleGrant
	for _, role := range desiredRoles {
		for _, parent := range role.ParentRoles {
			edge := [2]string{role.Name, parent}
			if !existing[edge] {
				result = append(result, RoleGrant{Role: role.Name, ToRole: parent})
				existing[edge] = true
			}
		}
	}
	return result
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

	// Generic grants on existing objects, including multiword Snowflake object types.
	for _, object := range cfg.Objects {
		for _, grant := range object.Grants {
			for _, privilege := range grant.Privileges {
				toAdd = append(toAdd, ObjectGrant{Privilege: privilege, ObjectType: object.Type, ObjectName: strings.Join(object.Name, "."), NameParts: object.Name, ToRole: grant.ToRole})
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
		actualUser := actualUsers[user.Name]
		actualRoles := make(map[string]bool)
		for _, role := range actualUser.Roles {
			actualRoles[role] = true
		}

		for _, role := range user.Roles {
			if !actualRoles[role] {
				toAdd = append(toAdd, UserRoleGrant{
					Role:   role,
					ToUser: user.Name,
				})
				actualRoles[role] = true
			}
		}
	}

	return toAdd
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

func missingObjectGrants(desired []ObjectGrant, actual []snowflake.Grant) []ObjectGrant {
	key := func(objectType, name, privilege, role string) [4]string {
		return [4]string{normalizeObjectType(objectType), name, privilege, role}
	}
	existing := make(map[[4]string]bool)
	for _, grant := range actual {
		if !isRoleRecipient(grant) {
			continue
		}
		existing[key(grant.GrantedOn, grant.Name, grant.Privilege, grant.GranteeName)] = true
	}
	result := []ObjectGrant{}
	for _, grant := range desired {
		k := key(grant.ObjectType, grant.ObjectName, grant.Privilege, grant.ToRole)
		if !existing[k] {
			result = append(result, grant)
			existing[k] = true
		}
	}
	return result
}
