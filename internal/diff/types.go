package diff

import (
	"github.com/yourusername/gatekeep/internal/config"
	"github.com/yourusername/gatekeep/internal/snowflake"
)

// SyncMode determines how the diff engine handles resources not in YAML
type SyncMode string

const (
	// SyncModeStrict removes resources not in YAML (full reconciliation)
	SyncModeStrict SyncMode = "strict"
	// SyncModeAdditive only adds resources, never removes (append-only)
	SyncModeAdditive SyncMode = "additive"
)

// Result represents the complete diff between desired and actual state
type Result struct {
	// Roles
	RolesToCreate []string
	RolesToDelete []string // Empty in additive mode

	// Role hierarchy (GRANT ROLE x TO ROLE y)
	RoleGrantsToAdd    []RoleGrant
	RoleGrantsToRevoke []RoleGrant // Empty in additive mode

	// Object permissions (GRANT privilege ON object TO ROLE)
	ObjectGrantsToAdd    []ObjectGrant
	ObjectGrantsToRevoke []ObjectGrant // Empty in additive mode

	// Users
	UsersToCreate []string

	// User role assignments (GRANT ROLE x TO USER y)
	UserRoleGrantsToAdd    []UserRoleGrant
	UserRoleGrantsToRevoke []UserRoleGrant // Empty in additive mode

	// Databases and warehouses to create
	DatabasesToCreate  []string
	WarehousesToCreate []string

	// Summary statistics
	Summary Stats
}

// RoleGrant represents a role granted to another role (role hierarchy)
type RoleGrant struct {
	Role      string // The role being granted
	ToRole    string // The role receiving the grant (parent role)
	GrantedBy string // Who granted it (optional)
	GrantedAt string // When it was granted (optional)
}

// ObjectGrant represents a privilege granted on an object to a role
type ObjectGrant struct {
	Privilege   string // SELECT, INSERT, USAGE, etc.
	ObjectType  string // TABLE, DATABASE, SCHEMA, WAREHOUSE, etc.
	ObjectName  string // Fully qualified name (e.g., "DB.SCHEMA.TABLE")
	ToRole      string // Role receiving the grant
	GrantOption bool   // Can this role grant to others?
}

// UserRoleGrant represents a role granted to a user
type UserRoleGrant struct {
	Role   string // The role being granted
	ToUser string // The user receiving the grant
}

// Stats provides a high-level summary of changes
type Stats struct {
	TotalOperations int

	RolesCreated int
	RolesDeleted int

	RoleGrantsAdded   int
	RoleGrantsRevoked int

	ObjectGrantsAdded   int
	ObjectGrantsRevoked int

	UsersCreated int

	UserRoleGrantsAdded   int
	UserRoleGrantsRevoked int

	DatabasesCreated  int
	WarehousesCreated int
}

// Input contains the input data for the diff engine
type Input struct {
	DesiredConfig config.Config
	ActualState   snowflake.State
	Mode          SyncMode
}

// IsEmpty returns true if the diff has no changes
func (d *Result) IsEmpty() bool {
	return d.Summary.TotalOperations == 0
}

// ComputeSummary calculates summary statistics
func (d *Result) ComputeSummary() {
	d.Summary = Stats{
		RolesCreated:          len(d.RolesToCreate),
		RolesDeleted:          len(d.RolesToDelete),
		RoleGrantsAdded:       len(d.RoleGrantsToAdd),
		RoleGrantsRevoked:     len(d.RoleGrantsToRevoke),
		ObjectGrantsAdded:     len(d.ObjectGrantsToAdd),
		ObjectGrantsRevoked:   len(d.ObjectGrantsToRevoke),
		UsersCreated:          len(d.UsersToCreate),
		UserRoleGrantsAdded:   len(d.UserRoleGrantsToAdd),
		UserRoleGrantsRevoked: len(d.UserRoleGrantsToRevoke),
		DatabasesCreated:      len(d.DatabasesToCreate),
		WarehousesCreated:     len(d.WarehousesToCreate),
	}

	d.Summary.TotalOperations = d.Summary.RolesCreated +
		d.Summary.RolesDeleted +
		d.Summary.RoleGrantsAdded +
		d.Summary.RoleGrantsRevoked +
		d.Summary.ObjectGrantsAdded +
		d.Summary.ObjectGrantsRevoked +
		d.Summary.UsersCreated +
		d.Summary.UserRoleGrantsAdded +
		d.Summary.UserRoleGrantsRevoked +
		d.Summary.DatabasesCreated +
		d.Summary.WarehousesCreated
}
