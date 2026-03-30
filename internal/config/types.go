package config

// Config represents the complete YAML configuration
type Config struct {
	Version    string      `yaml:"version" validate:"required"`
	Roles      []Role      `yaml:"roles,omitempty"`
	Users      []User      `yaml:"users,omitempty"`
	Databases  []Database  `yaml:"databases,omitempty"`
	Warehouses []Warehouse `yaml:"warehouses,omitempty"`
}

// Role represents a Snowflake role
type Role struct {
	Name        string   `yaml:"name" validate:"required"`
	ParentRoles []string `yaml:"parent_roles,omitempty"`
	Comment     string   `yaml:"comment,omitempty"`
}

// User represents a Snowflake user
type User struct {
	Name  string   `yaml:"name" validate:"required,email"`
	Roles []string `yaml:"roles,omitempty" validate:"required,min=1"`
}

// Database represents a Snowflake database with its schemas
type Database struct {
	Name    string   `yaml:"name" validate:"required"`
	Schemas []Schema `yaml:"schemas,omitempty"`
}

// Schema represents a database schema with its tables
type Schema struct {
	Name   string  `yaml:"name" validate:"required"`
	Tables []Table `yaml:"tables,omitempty"`
}

// Table represents a database table with grants
type Table struct {
	Name   string  `yaml:"name" validate:"required"`
	Grants []Grant `yaml:"grants,omitempty"`
}

// Warehouse represents a Snowflake warehouse with grants
type Warehouse struct {
	Name   string  `yaml:"name" validate:"required"`
	Grants []Grant `yaml:"grants,omitempty"`
}

// Grant represents a permission grant
type Grant struct {
	ToRole     string   `yaml:"to_role" validate:"required"`
	Privileges []string `yaml:"privileges" validate:"required,min=1"`
}

// Privilege constants for validation
const (
	// Table privileges
	PrivilegeSelect = "SELECT"
	PrivilegeInsert = "INSERT"
	PrivilegeUpdate = "UPDATE"
	PrivilegeDelete = "DELETE"

	// Warehouse privileges
	PrivilegeUsage   = "USAGE"
	PrivilegeOperate = "OPERATE"
	PrivilegeMonitor = "MONITOR"
	PrivilegeModify  = "MODIFY"

	// Schema privileges
	PrivilegeCreateTable = "CREATE TABLE"
	PrivilegeCreateView  = "CREATE VIEW"
)

// ValidTablePrivileges returns all valid table privileges
func ValidTablePrivileges() []string {
	return []string{
		PrivilegeSelect,
		PrivilegeInsert,
		PrivilegeUpdate,
		PrivilegeDelete,
	}
}

// ValidWarehousePrivileges returns all valid warehouse privileges
func ValidWarehousePrivileges() []string {
	return []string{
		PrivilegeUsage,
		PrivilegeOperate,
		PrivilegeMonitor,
		PrivilegeModify,
	}
}

// ValidSchemaPrivileges returns all valid schema privileges
func ValidSchemaPrivileges() []string {
	return []string{
		PrivilegeCreateTable,
		PrivilegeCreateView,
		PrivilegeUsage,
	}
}
