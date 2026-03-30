package diff

import (
	"fmt"
	"sort"
	"strings"
)

// SQLOperation represents a single SQL statement to execute
type SQLOperation struct {
	Type        OperationType
	SQL         string
	Target      string // Role, user, or object name
	Description string
}

// OperationType categorizes SQL operations for ordering
type OperationType string

const (
	OpCreateDatabase  OperationType = "CREATE_DATABASE"
	OpCreateWarehouse OperationType = "CREATE_WAREHOUSE"
	OpCreateRole      OperationType = "CREATE_ROLE"
	OpDeleteRole      OperationType = "DELETE_ROLE"
	OpGrantRole       OperationType = "GRANT_ROLE"
	OpRevokeRole      OperationType = "REVOKE_ROLE"
	OpGrantObject     OperationType = "GRANT_OBJECT"
	OpRevokeObject    OperationType = "REVOKE_OBJECT"
	OpCreateUser      OperationType = "CREATE_USER"
	OpGrantUserRole   OperationType = "GRANT_USER_ROLE"
	OpRevokeUserRole  OperationType = "REVOKE_USER_ROLE"
)

// ExecutionPhase groups operations that can run in parallel
type ExecutionPhase int

const (
	PhaseCreateResources ExecutionPhase = iota // Databases, warehouses
	PhaseCreateRoles                           // Create all roles
	PhaseGrantRoleHierarchy                    // GRANT ROLE TO ROLE
	PhaseGrantObjectPermissions                // GRANT privileges ON object
	PhaseCreateUsers                           // Create users
	PhaseGrantUserRoles                        // GRANT ROLE TO USER
	PhaseRevokePermissions                     // All REVOKE operations
	PhaseDeleteRoles                           // DROP ROLE
)

// Planner generates SQL statements from diff results
type Planner struct {
	diff *DiffResult
}

// NewPlanner creates a new SQL planner
func NewPlanner(diff *DiffResult) *Planner {
	return &Planner{diff: diff}
}

// GeneratePlan creates an ordered list of SQL operations
func (p *Planner) GeneratePlan() ([]SQLOperation, error) {
	var operations []SQLOperation

	// Phase 1: Create databases and warehouses (these are dependencies for grants)
	operations = append(operations, p.generateCreateDatabases()...)
	operations = append(operations, p.generateCreateWarehouses()...)

	// Phase 2: Create roles (in dependency order)
	createRoleOps, err := p.generateCreateRoles()
	if err != nil {
		return nil, fmt.Errorf("failed to generate create role operations: %w", err)
	}
	operations = append(operations, createRoleOps...)

	// Phase 3: Grant role hierarchies (parent role assignments)
	grantRoleOps, err := p.generateGrantRoles()
	if err != nil {
		return nil, fmt.Errorf("failed to generate grant role operations: %w", err)
	}
	operations = append(operations, grantRoleOps...)

	// Phase 4: Grant object permissions
	operations = append(operations, p.generateGrantObjects()...)

	// Phase 5: Create users
	operations = append(operations, p.generateCreateUsers()...)

	// Phase 6: Grant roles to users
	operations = append(operations, p.generateGrantUserRoles()...)

	// Phase 7: Revoke operations (in reverse order of grants)
	operations = append(operations, p.generateRevokeUserRoles()...)
	operations = append(operations, p.generateRevokeObjects()...)
	operations = append(operations, p.generateRevokeRoles()...)

	// Phase 8: Delete roles (after all revocations)
	operations = append(operations, p.generateDeleteRoles()...)

	return operations, nil
}

// generateCreateDatabases generates CREATE DATABASE statements
func (p *Planner) generateCreateDatabases() []SQLOperation {
	var ops []SQLOperation
	for _, db := range p.diff.DatabasesToCreate {
		ops = append(ops, SQLOperation{
			Type:        OpCreateDatabase,
			SQL:         fmt.Sprintf("CREATE DATABASE IF NOT EXISTS \"%s\"", db),
			Target:      db,
			Description: fmt.Sprintf("Create database %s", db),
		})
	}
	return ops
}

// generateCreateWarehouses generates CREATE WAREHOUSE statements
func (p *Planner) generateCreateWarehouses() []SQLOperation {
	var ops []SQLOperation
	for _, wh := range p.diff.WarehousesToCreate {
		ops = append(ops, SQLOperation{
			Type:        OpCreateWarehouse,
			SQL:         fmt.Sprintf("CREATE WAREHOUSE IF NOT EXISTS \"%s\"", wh),
			Target:      wh,
			Description: fmt.Sprintf("Create warehouse %s", wh),
		})
	}
	return ops
}

// generateCreateRoles generates CREATE ROLE statements in dependency order
func (p *Planner) generateCreateRoles() ([]SQLOperation, error) {
	var ops []SQLOperation

	// Sort roles to ensure deterministic output
	roles := make([]string, len(p.diff.RolesToCreate))
	copy(roles, p.diff.RolesToCreate)
	sort.Strings(roles)

	for _, role := range roles {
		ops = append(ops, SQLOperation{
			Type:        OpCreateRole,
			SQL:         fmt.Sprintf("CREATE ROLE IF NOT EXISTS \"%s\"", role),
			Target:      role,
			Description: fmt.Sprintf("Create role %s", role),
		})
	}

	return ops, nil
}

// generateGrantRoles generates GRANT ROLE TO ROLE statements (role hierarchy)
func (p *Planner) generateGrantRoles() ([]SQLOperation, error) {
	var ops []SQLOperation

	// Build dependency graph to ensure parent roles are granted first
	grants := p.diff.RoleGrantsToAdd

	// Sort for deterministic output
	sort.Slice(grants, func(i, j int) bool {
		if grants[i].ToRole != grants[j].ToRole {
			return grants[i].ToRole < grants[j].ToRole
		}
		return grants[i].Role < grants[j].Role
	})

	for _, grant := range grants {
		ops = append(ops, SQLOperation{
			Type:        OpGrantRole,
			SQL:         fmt.Sprintf("GRANT ROLE \"%s\" TO ROLE \"%s\"", grant.Role, grant.ToRole),
			Target:      grant.Role,
			Description: fmt.Sprintf("Grant role %s to role %s", grant.Role, grant.ToRole),
		})
	}

	return ops, nil
}

// generateGrantObjects generates GRANT privilege ON object TO ROLE statements
func (p *Planner) generateGrantObjects() []SQLOperation {
	var ops []SQLOperation

	grants := p.diff.ObjectGrantsToAdd

	// Sort for deterministic output
	sort.Slice(grants, func(i, j int) bool {
		if grants[i].ObjectType != grants[j].ObjectType {
			return grants[i].ObjectType < grants[j].ObjectType
		}
		if grants[i].ObjectName != grants[j].ObjectName {
			return grants[i].ObjectName < grants[j].ObjectName
		}
		if grants[i].ToRole != grants[j].ToRole {
			return grants[i].ToRole < grants[j].ToRole
		}
		return grants[i].Privilege < grants[j].Privilege
	})

	for _, grant := range grants {
		sql := fmt.Sprintf("GRANT %s ON %s \"%s\" TO ROLE \"%s\"",
			grant.Privilege,
			grant.ObjectType,
			grant.ObjectName,
			grant.ToRole,
		)

		ops = append(ops, SQLOperation{
			Type:        OpGrantObject,
			SQL:         sql,
			Target:      grant.ObjectName,
			Description: fmt.Sprintf("Grant %s on %s %s to role %s", grant.Privilege, grant.ObjectType, grant.ObjectName, grant.ToRole),
		})
	}

	return ops
}

// generateCreateUsers generates CREATE USER statements
func (p *Planner) generateCreateUsers() []SQLOperation {
	var ops []SQLOperation

	// Sort for deterministic output
	users := make([]string, len(p.diff.UsersToCreate))
	copy(users, p.diff.UsersToCreate)
	sort.Strings(users)

	for _, user := range users {
		ops = append(ops, SQLOperation{
			Type:        OpCreateUser,
			SQL:         fmt.Sprintf("CREATE USER IF NOT EXISTS \"%s\"", user),
			Target:      user,
			Description: fmt.Sprintf("Create user %s", user),
		})
	}

	return ops
}

// generateGrantUserRoles generates GRANT ROLE TO USER statements
func (p *Planner) generateGrantUserRoles() []SQLOperation {
	var ops []SQLOperation

	grants := p.diff.UserRoleGrantsToAdd

	// Sort for deterministic output
	sort.Slice(grants, func(i, j int) bool {
		if grants[i].ToUser != grants[j].ToUser {
			return grants[i].ToUser < grants[j].ToUser
		}
		return grants[i].Role < grants[j].Role
	})

	for _, grant := range grants {
		ops = append(ops, SQLOperation{
			Type:        OpGrantUserRole,
			SQL:         fmt.Sprintf("GRANT ROLE \"%s\" TO USER \"%s\"", grant.Role, grant.ToUser),
			Target:      grant.ToUser,
			Description: fmt.Sprintf("Grant role %s to user %s", grant.Role, grant.ToUser),
		})
	}

	return ops
}

// generateRevokeUserRoles generates REVOKE ROLE FROM USER statements
func (p *Planner) generateRevokeUserRoles() []SQLOperation {
	var ops []SQLOperation

	revokes := p.diff.UserRoleGrantsToRevoke

	// Sort for deterministic output
	sort.Slice(revokes, func(i, j int) bool {
		if revokes[i].ToUser != revokes[j].ToUser {
			return revokes[i].ToUser < revokes[j].ToUser
		}
		return revokes[i].Role < revokes[j].Role
	})

	for _, revoke := range revokes {
		ops = append(ops, SQLOperation{
			Type:        OpRevokeUserRole,
			SQL:         fmt.Sprintf("REVOKE ROLE \"%s\" FROM USER \"%s\"", revoke.Role, revoke.ToUser),
			Target:      revoke.ToUser,
			Description: fmt.Sprintf("Revoke role %s from user %s", revoke.Role, revoke.ToUser),
		})
	}

	return ops
}

// generateRevokeObjects generates REVOKE privilege ON object FROM ROLE statements
func (p *Planner) generateRevokeObjects() []SQLOperation {
	var ops []SQLOperation

	revokes := p.diff.ObjectGrantsToRevoke

	// Sort for deterministic output
	sort.Slice(revokes, func(i, j int) bool {
		if revokes[i].ObjectType != revokes[j].ObjectType {
			return revokes[i].ObjectType < revokes[j].ObjectType
		}
		if revokes[i].ObjectName != revokes[j].ObjectName {
			return revokes[i].ObjectName < revokes[j].ObjectName
		}
		if revokes[i].ToRole != revokes[j].ToRole {
			return revokes[i].ToRole < revokes[j].ToRole
		}
		return revokes[i].Privilege < revokes[j].Privilege
	})

	for _, revoke := range revokes {
		sql := fmt.Sprintf("REVOKE %s ON %s \"%s\" FROM ROLE \"%s\"",
			revoke.Privilege,
			revoke.ObjectType,
			revoke.ObjectName,
			revoke.ToRole,
		)

		ops = append(ops, SQLOperation{
			Type:        OpRevokeObject,
			SQL:         sql,
			Target:      revoke.ObjectName,
			Description: fmt.Sprintf("Revoke %s on %s %s from role %s", revoke.Privilege, revoke.ObjectType, revoke.ObjectName, revoke.ToRole),
		})
	}

	return ops
}

// generateRevokeRoles generates REVOKE ROLE FROM ROLE statements
func (p *Planner) generateRevokeRoles() []SQLOperation {
	var ops []SQLOperation

	revokes := p.diff.RoleGrantsToRevoke

	// Sort for deterministic output
	sort.Slice(revokes, func(i, j int) bool {
		if revokes[i].ToRole != revokes[j].ToRole {
			return revokes[i].ToRole < revokes[j].ToRole
		}
		return revokes[i].Role < revokes[j].Role
	})

	for _, revoke := range revokes {
		ops = append(ops, SQLOperation{
			Type:        OpRevokeRole,
			SQL:         fmt.Sprintf("REVOKE ROLE \"%s\" FROM ROLE \"%s\"", revoke.Role, revoke.ToRole),
			Target:      revoke.Role,
			Description: fmt.Sprintf("Revoke role %s from role %s", revoke.Role, revoke.ToRole),
		})
	}

	return ops
}

// generateDeleteRoles generates DROP ROLE statements
func (p *Planner) generateDeleteRoles() []SQLOperation {
	var ops []SQLOperation

	// Sort for deterministic output
	roles := make([]string, len(p.diff.RolesToDelete))
	copy(roles, p.diff.RolesToDelete)
	sort.Strings(roles)

	for _, role := range roles {
		// Skip system roles (safety check)
		if isSystemRole(role) {
			continue
		}

		ops = append(ops, SQLOperation{
			Type:        OpDeleteRole,
			SQL:         fmt.Sprintf("DROP ROLE IF EXISTS \"%s\"", role),
			Target:      role,
			Description: fmt.Sprintf("Delete role %s", role),
		})
	}

	return ops
}

// FormatPlan formats operations as human-readable text
func FormatPlan(operations []SQLOperation) string {
	if len(operations) == 0 {
		return "No operations to execute"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Execution Plan (%d operations):\n\n", len(operations)))

	currentPhase := ""
	for i, op := range operations {
		phase := getPhaseForOperation(op.Type)
		if phase != currentPhase {
			if currentPhase != "" {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("--- %s ---\n", phase))
			currentPhase = phase
		}

		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, op.Type, op.SQL))
	}

	return sb.String()
}

func getPhaseForOperation(opType OperationType) string {
	switch opType {
	case OpCreateDatabase, OpCreateWarehouse:
		return "CREATE RESOURCES"
	case OpCreateRole:
		return "CREATE ROLES"
	case OpGrantRole:
		return "GRANT ROLE HIERARCHIES"
	case OpGrantObject:
		return "GRANT OBJECT PERMISSIONS"
	case OpCreateUser:
		return "CREATE USERS"
	case OpGrantUserRole:
		return "GRANT USER ROLES"
	case OpRevokeUserRole, OpRevokeObject, OpRevokeRole:
		return "REVOKE PERMISSIONS"
	case OpDeleteRole:
		return "DELETE ROLES"
	default:
		return "OTHER"
	}
}
