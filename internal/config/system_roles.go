package config

import "strings"

// IsSystemRole protects Snowflake's built-in account and organization roles.
func IsSystemRole(name string) bool {
	switch strings.ToUpper(name) {
	case "ACCOUNTADMIN", "SECURITYADMIN", "SYSADMIN", "USERADMIN", "PUBLIC", "ORGADMIN", "GLOBALORGADMIN":
		return true
	default:
		return false
	}
}
