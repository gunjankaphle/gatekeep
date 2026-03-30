package snowflake

// State represents the current state of Snowflake
type State struct {
	Roles      []Role
	Users      []User
	Grants     []Grant
	Databases  []Database
	Warehouses []Warehouse
}

// Role represents a Snowflake role
type Role struct {
	Name    string
	Comment string
	Owner   string
}

// User represents a Snowflake user with assigned roles
type User struct {
	Name  string
	Roles []string
}

// Grant represents a permission grant in Snowflake
type Grant struct {
	GrantedOn   string // ROLE, TABLE, WAREHOUSE, etc.
	GrantedTo   string // ROLE or USER
	Name        string // Name of the object
	Privilege   string // SELECT, INSERT, USAGE, etc.
	GranteeType string // ROLE or USER
	GranteeName string // Name of the grantee
}

// Database represents a Snowflake database
type Database struct {
	Name string
}

// Warehouse represents a Snowflake warehouse
type Warehouse struct {
	Name string
}
