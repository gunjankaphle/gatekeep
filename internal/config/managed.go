package config

import (
	"fmt"
	"strings"
)

// ValidateManagedScope rejects ambiguous or unsafe revocation boundaries.
func ValidateManagedScope(scope *ManagedScope) error {
	if scope == nil {
		return nil
	}
	if len(scope.Roles) == 0 {
		return fmt.Errorf("managed.roles must explicitly list at least one custom role")
	}
	roles := map[string]bool{}
	for _, role := range scope.Roles {
		if strings.TrimSpace(role) == "" || strings.ContainsRune(role, 0) || roles[role] {
			return fmt.Errorf("invalid or duplicate managed role: %q", role)
		}
		if IsSystemRole(role) {
			return fmt.Errorf("system role cannot be managed: %s", role)
		}
		roles[role] = true
	}
	users := map[string]bool{}
	for _, user := range scope.Users {
		if strings.TrimSpace(user) == "" || strings.ContainsRune(user, 0) || users[user] {
			return fmt.Errorf("invalid or duplicate managed user: %q", user)
		}
		users[user] = true
	}
	for _, object := range scope.Objects {
		// Reuse named-object validation; scoped privileges share the same SQL grammar.
		if err := NewValidator().validateObjects([]Object{{Type: object.Type, Name: object.Name, Grants: []Grant{{ToRole: scope.Roles[0], Privileges: object.Privileges}}}}, roles); err != nil {
			return fmt.Errorf("invalid managed object: %w", err)
		}
		// SHOW qualified names cannot safely distinguish dotted identifier parts
		// through the current state model. Reject them for revocation scopes.
		for _, part := range object.Name {
			if strings.ContainsAny(part, ".\"") {
				return fmt.Errorf("managed object identifiers containing dots or quotes are not supported yet")
			}
		}
		switch object.Type {
		case "ROLE", "USER", "ACCOUNT", "FUNCTION", "PROCEDURE", "DATABASE ROLE":
			return fmt.Errorf("managed object type %s requires separate reconciliation support", object.Type)
		}
	}
	return nil
}
