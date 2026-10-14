package diff

import (
	"sort"
	"strings"

	"github.com/yourusername/gatekeep/internal/config"
	"github.com/yourusername/gatekeep/internal/snowflake"
)

func normalizeObjectType(value string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(strings.ToUpper(value), "_", " ")), " ")
}

func isRoleRecipient(grant snowflake.Grant) bool {
	kind := grant.GrantedTo
	if kind == "" {
		kind = grant.GranteeType
	}
	// Older callers omit recipient type; explicit non-account-role recipients never match.
	return kind == "" || kind == "ROLE"
}

func roleEdges(actual []snowflake.Grant) map[[2]string]bool {
	edges := map[[2]string]bool{}
	for _, grant := range actual {
		if normalizeObjectType(grant.GrantedOn) == "ROLE" && grant.Privilege == "USAGE" && isRoleRecipient(grant) {
			edges[[2]string{grant.Name, grant.GranteeName}] = true
		}
	}
	return edges
}

func stringSet(values []string) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		result[value] = true
	}
	return result
}

func scopedRoleRevocations(cfg config.Config, actual []snowflake.Grant) []RoleGrant {
	managed := stringSet(cfg.Managed.Roles)
	desired := map[[2]string]bool{}
	for _, role := range cfg.Roles {
		for _, parent := range role.ParentRoles {
			desired[[2]string{role.Name, parent}] = true
		}
	}
	var result []RoleGrant
	for edge := range roleEdges(actual) {
		if managed[edge[0]] && managed[edge[1]] && !desired[edge] {
			result = append(result, RoleGrant{Role: edge[0], ToRole: edge[1]})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Role != result[j].Role {
			return result[i].Role < result[j].Role
		}
		return result[i].ToRole < result[j].ToRole
	})
	return result
}

func scopedUserRevocations(cfg config.Config, actual []snowflake.User) []UserRoleGrant {
	roles, users := stringSet(cfg.Managed.Roles), stringSet(cfg.Managed.Users)
	desired := map[[2]string]bool{}
	for _, user := range cfg.Users {
		for _, role := range user.Roles {
			desired[[2]string{role, user.Name}] = true
		}
	}
	seen := map[[2]string]bool{}
	var result []UserRoleGrant
	for _, user := range actual {
		if !users[user.Name] {
			continue
		}
		for _, role := range user.Roles {
			edge := [2]string{role, user.Name}
			if roles[role] && !desired[edge] && !seen[edge] {
				result = append(result, UserRoleGrant{Role: role, ToUser: user.Name})
				seen[edge] = true
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].ToUser != result[j].ToUser {
			return result[i].ToUser < result[j].ToUser
		}
		return result[i].Role < result[j].Role
	})
	return result
}

func scopedObjectRevocations(cfg config.Config, actual []snowflake.Grant, desired []ObjectGrant) []ObjectGrant {
	roles := stringSet(cfg.Managed.Roles)
	type objectKey struct{ kind, name string }
	objects := map[objectKey]config.ManagedObject{}
	privileges := map[objectKey]map[string]bool{}
	for _, object := range cfg.Managed.Objects {
		key := objectKey{normalizeObjectType(object.Type), strings.Join(object.Name, ".")}
		objects[key] = object
		if privileges[key] == nil {
			privileges[key] = map[string]bool{}
		}
		for _, privilege := range object.Privileges {
			privileges[key][privilege] = true
		}
	}
	wanted := map[[4]string]bool{}
	for _, grant := range desired {
		wanted[[4]string{normalizeObjectType(grant.ObjectType), grant.ObjectName, grant.Privilege, grant.ToRole}] = true
	}
	seen := map[[4]string]bool{}
	var result []ObjectGrant
	for _, grant := range actual {
		key := objectKey{normalizeObjectType(grant.GrantedOn), grant.Name}
		id := [4]string{key.kind, key.name, grant.Privilege, grant.GranteeName}
		if !isRoleRecipient(grant) || !roles[grant.GranteeName] || !privileges[key][grant.Privilege] || wanted[id] || seen[id] {
			continue
		}
		object := objects[key]
		result = append(result, ObjectGrant{ObjectType: object.Type, ObjectName: grant.Name, NameParts: object.Name, Privilege: grant.Privilege, ToRole: grant.GranteeName})
		seen[id] = true
	}
	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		if a.ObjectType != b.ObjectType {
			return a.ObjectType < b.ObjectType
		}
		if a.ObjectName != b.ObjectName {
			return a.ObjectName < b.ObjectName
		}
		if a.ToRole != b.ToRole {
			return a.ToRole < b.ToRole
		}
		return a.Privilege < b.Privilege
	})
	return result
}
