package snowflake

import (
	"database/sql"
	"fmt"
)

// StateReader reads current state from Snowflake
type StateReader struct {
	client Client
}

// NewStateReader creates a new StateReader
func NewStateReader(client Client) *StateReader {
	return &StateReader{client: client}
}

// ReadState reads the complete current state from Snowflake
func (sr *StateReader) ReadState() (*State, error) {
	state := &State{}

	// Read roles
	roles, err := sr.ReadRoles()
	if err != nil {
		return nil, fmt.Errorf("failed to read roles: %w", err)
	}
	state.Roles = roles

	// Read users with their roles
	users, err := sr.ReadUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to read users: %w", err)
	}
	state.Users = users

	// Read grants
	grants, err := sr.ReadGrants()
	if err != nil {
		return nil, fmt.Errorf("failed to read grants: %w", err)
	}
	state.Grants = grants

	// Read databases
	databases, err := sr.ReadDatabases()
	if err != nil {
		return nil, fmt.Errorf("failed to read databases: %w", err)
	}
	state.Databases = databases

	// Read warehouses
	warehouses, err := sr.ReadWarehouses()
	if err != nil {
		return nil, fmt.Errorf("failed to read warehouses: %w", err)
	}
	state.Warehouses = warehouses

	return state, nil
}

// ReadRoles reads all roles from Snowflake
func (sr *StateReader) ReadRoles() ([]Role, error) {
	query := "SHOW ROLES"
	rows, err := sr.client.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	//nolint:errcheck // Deferred close
	defer func() { _ = rows.Close() }()

	var roles []Role
	for rows.Next() {
		var role Role
		var createdOn, owner sql.NullString
		var assignedToUsers, grantedToRoles, grantedRoles sql.NullInt64

		// SHOW ROLES returns: created_on, name, is_default, is_current, is_inherited, assigned_to_users, granted_to_roles, granted_roles, owner, comment
		scanErr := rows.Scan(
			&createdOn,
			&role.Name,
			&sql.NullBool{}, // is_default
			&sql.NullBool{}, // is_current
			&sql.NullBool{}, // is_inherited
			&assignedToUsers,
			&grantedToRoles,
			&grantedRoles,
			&owner,
			&role.Comment,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan role: %w", scanErr)
		}

		if owner.Valid {
			role.Owner = owner.String
		}

		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// ReadUsers reads all users and their assigned roles
func (sr *StateReader) ReadUsers() ([]User, error) {
	query := "SHOW USERS"
	rows, err := sr.client.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	//nolint:errcheck // Deferred close
	defer func() { _ = rows.Close() }()

	userMap := make(map[string]*User)
	for rows.Next() {
		var name, loginName, displayName, defaultRole, defaultNamespace, defaultWarehouse sql.NullString
		var createdOn, hasPassword, hasRsaPublicKey, disabled, comment, owner sql.NullString
		var lastSuccessLogin, expiresAt, lockedUntil, extAuthnDuo, extAuthnUID, bypassMfaUntil sql.NullString
		var snowflakeLock sql.NullBool

		// SHOW USERS returns many columns, we only need name
		err := rows.Scan(
			&name,
			&createdOn,
			&loginName,
			&displayName,
			&sql.NullString{}, // first_name
			&sql.NullString{}, // last_name
			&sql.NullString{}, // email
			&sql.NullString{}, // mins_to_unlock
			&sql.NullString{}, // days_to_expiry
			&comment,
			&disabled,
			&sql.NullBool{}, // must_change_password
			&snowflakeLock,
			&defaultWarehouse,
			&defaultNamespace,
			&defaultRole,
			&sql.NullString{}, // default_secondary_roles
			&extAuthnDuo,
			&extAuthnUID,
			&bypassMfaUntil,
			&lastSuccessLogin,
			&expiresAt,
			&lockedUntil,
			&hasPassword,
			&hasRsaPublicKey,
			&sql.NullString{}, // email_verified
			&owner,
		)
		if err != nil {
			// Try simpler scan if columns don't match (Snowflake version differences)
			continue
		}

		if name.Valid {
			userMap[name.String] = &User{
				Name:  name.String,
				Roles: []string{},
			}
		}
	}

	// Read role grants for each user
	for userName := range userMap {
		query := fmt.Sprintf("SHOW GRANTS TO USER \"%s\"", userName)
		roleRows, err := sr.client.Query(query)
		if err != nil {
			// User might not exist or no permissions, skip
			continue
		}

		for roleRows.Next() {
			var createdOn, privilege, grantedOn, name, grantedTo, granteeType, grantOption sql.NullString

			err := roleRows.Scan(
				&createdOn,
				&privilege,
				&grantedOn,
				&name,
				&grantedTo,
				&granteeType,
				&grantOption,
			)
			if err != nil {
				continue
			}

			// Add role to user if it's a role grant
			if grantedOn.Valid && grantedOn.String == "ROLE" && name.Valid {
				userMap[userName].Roles = append(userMap[userName].Roles, name.String)
			}
		}
		//nolint:errcheck // Cleanup
		_ = roleRows.Close()
	}

	// Convert map to slice
	var users []User
	for _, user := range userMap {
		users = append(users, *user)
	}

	return users, nil
}

// ReadGrants reads all grants from Snowflake
func (sr *StateReader) ReadGrants() ([]Grant, error) {
	// For now, return empty - will be populated when reading role grants
	// In a full implementation, you would query grants for each role
	return []Grant{}, nil
}

// ReadDatabases reads all databases from Snowflake
func (sr *StateReader) ReadDatabases() ([]Database, error) {
	query := "SHOW DATABASES"
	rows, err := sr.client.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	//nolint:errcheck // Deferred close
	defer func() { _ = rows.Close() }()

	var databases []Database
	for rows.Next() {
		var db Database
		var createdOn, owner, comment, options, retentionTime sql.NullString

		err := rows.Scan(
			&createdOn,
			&db.Name,
			&sql.NullBool{},   // is_default
			&sql.NullBool{},   // is_current
			&sql.NullString{}, // origin
			&owner,
			&comment,
			&options,
			&retentionTime,
		)
		if err != nil {
			continue
		}

		databases = append(databases, db)
	}

	return databases, nil
}

// ReadWarehouses reads all warehouses from Snowflake
func (sr *StateReader) ReadWarehouses() ([]Warehouse, error) {
	query := "SHOW WAREHOUSES"
	rows, err := sr.client.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	//nolint:errcheck // Deferred close
	defer func() { _ = rows.Close() }()

	var warehouses []Warehouse
	for rows.Next() {
		var wh Warehouse
		// SHOW WAREHOUSES has many columns, we just need name
		var dummy sql.NullString
		var dummyInt sql.NullInt64
		var dummyBool sql.NullBool

		err := rows.Scan(
			&wh.Name,
			&dummy,     // state
			&dummy,     // type
			&dummy,     // size
			&dummyInt,  // min_cluster_count
			&dummyInt,  // max_cluster_count
			&dummyInt,  // started_clusters
			&dummyInt,  // running
			&dummyInt,  // queued
			&dummyBool, // is_default
			&dummyBool, // is_current
			&dummyBool, // auto_suspend
			&dummyInt,  // auto_resume
			&dummy,     // available
			&dummy,     // provisioning
			&dummy,     // quiescing
			&dummy,     // other
			&dummy,     // created_on
			&dummy,     // resumed_on
			&dummy,     // updated_on
			&dummy,     // owner
			&dummy,     // comment
			&dummy,     // enable_query_acceleration
			&dummy,     // query_acceleration_max_scale_factor
			&dummy,     // resource_monitor
			&dummy,     // actives
			&dummy,     // pendings
			&dummy,     // failed
			&dummy,     // suspended
			&dummy,     // uuid
		)
		if err != nil {
			// Try simpler scan
			var name sql.NullString
			if scanErr := rows.Scan(&name); scanErr != nil {
				continue
			}
			if name.Valid {
				wh.Name = name.String
			} else {
				continue
			}
		}

		warehouses = append(warehouses, wh)
	}

	return warehouses, nil
}
