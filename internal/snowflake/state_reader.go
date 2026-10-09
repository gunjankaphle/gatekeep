package snowflake

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
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
	rows, err := sr.queryRows("SHOW ROLES")
	if err != nil {
		return nil, err
	}
	roles := make([]Role, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, Role{Name: row["name"], Comment: row["comment"], Owner: row["owner"]})
	}
	return roles, nil
}

// ReadUsers reads users and their assigned roles, propagating query failures.
func (sr *StateReader) ReadUsers() ([]User, error) {
	rows, err := sr.queryRows("SHOW USERS")
	if err != nil {
		return nil, err
	}
	users := make([]User, 0, len(rows))
	for _, row := range rows {
		user := User{Name: row["name"], Roles: []string{}}
		grants, err := sr.queryRows("SHOW GRANTS TO USER " + quoteIdentifier(user.Name))
		if err != nil {
			return nil, fmt.Errorf("failed to read user grants: %w", err)
		}
		for _, grant := range grants {
			if grant["role"] != "" {
				user.Roles = append(user.Roles, grant["role"])
			} else if grant["granted_on"] == "ROLE" {
				user.Roles = append(user.Roles, grant["name"])
			}
		}
		users = append(users, user)
	}
	return users, nil
}

// ReadGrants reads current grants for all visible roles.
func (sr *StateReader) ReadGrants() ([]Grant, error) {
	roles, err := sr.ReadRoles()
	if err != nil {
		return nil, err
	}
	return sr.fetchGrantsParallel(roles, 10)
}

// GetAllRolesWithGrants fetches roles and their grants with bounded concurrency.
func (sr *StateReader) GetAllRolesWithGrants(maxWorkers int) ([]Role, []Grant, error) {
	// 1. Get all roles first
	roles, err := sr.ReadRoles()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read roles: %w", err)
	}

	// 2. Fetch grants for all roles in parallel
	grants, err := sr.fetchGrantsParallel(roles, maxWorkers)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch grants: %w", err)
	}

	return roles, grants, nil
}

// fetchGrantsParallel fetches grants for multiple roles using a worker pool
func (sr *StateReader) fetchGrantsParallel(roles []Role, maxWorkers int) ([]Grant, error) {
	if maxWorkers <= 0 {
		maxWorkers = 10 // Default to 10 concurrent workers
	}

	// Create channels for work distribution
	type workItem struct {
		role Role
		idx  int
	}
	workChan := make(chan workItem, len(roles))
	resultChan := make(chan []Grant, len(roles))
	errorChan := make(chan error, len(roles))

	// Create worker pool
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range workChan {
				grants, err := sr.fetchGrantsForRole(work.role.Name)
				if err != nil {
					errorChan <- fmt.Errorf("failed to fetch grants for role %s: %w", work.role.Name, err)
					resultChan <- nil
					continue
				}
				resultChan <- grants
			}
		}()
	}

	// Send work to workers
	for idx, role := range roles {
		workChan <- workItem{role: role, idx: idx}
	}
	close(workChan)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results
	var allGrants []Grant
	var errors []error

	for grants := range resultChan {
		if grants != nil {
			allGrants = append(allGrants, grants...)
		}
	}

	// Collect any errors
	for err := range errorChan {
		errors = append(errors, err)
	}

	// If we have errors, return the first one (but still return partial results)
	if len(errors) > 0 {
		return allGrants, errors[0]
	}

	return allGrants, nil
}

// fetchGrantsForRole fetches all grants for a specific role
func (sr *StateReader) fetchGrantsForRole(roleName string) ([]Grant, error) {
	rows, err := sr.queryRows("SHOW GRANTS TO ROLE " + quoteIdentifier(roleName))
	if err != nil {
		return nil, err
	}
	grants := make([]Grant, 0, len(rows))
	for _, row := range rows {
		grants = append(grants, Grant{GrantedOn: row["granted_on"], GrantedTo: row["granted_to"], Name: row["name"], Privilege: row["privilege"], GranteeType: row["granted_to"], GranteeName: roleName})
	}
	return grants, nil
}

// ReadDatabases reads visible databases.
func (sr *StateReader) ReadDatabases() ([]Database, error) {
	rows, err := sr.queryRows("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	result := make([]Database, 0, len(rows))
	for _, row := range rows {
		result = append(result, Database{Name: row["name"]})
	}
	return result, nil
}

// ReadWarehouses reads visible warehouses.
func (sr *StateReader) ReadWarehouses() ([]Warehouse, error) {
	rows, err := sr.queryRows("SHOW WAREHOUSES")
	if err != nil {
		return nil, err
	}
	result := make([]Warehouse, 0, len(rows))
	for _, row := range rows {
		result = append(result, Warehouse{Name: row["name"]})
	}
	return result, nil
}

func quoteIdentifier(name string) string { return `"` + strings.ReplaceAll(name, `"`, `""`) + `"` }

// SHOW output can gain columns. Read by column name rather than a fixed position count.
func (sr *StateReader) queryRows(query string) ([]map[string]string, error) {
	rows, err := sr.client.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query %s: %w", query, err)
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // best-effort cleanup; rows.Err reports iteration errors
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := []map[string]string{}
	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		targets := make([]interface{}, len(columns))
		for i := range values {
			targets[i] = &values[i]
		}
		if err := rows.Scan(targets...); err != nil {
			return nil, fmt.Errorf("scan %s: %w", query, err)
		}
		row := make(map[string]string, len(columns))
		for i, column := range columns {
			row[strings.ToLower(column)] = values[i].String
		}
		result = append(result, row)
	}
	return result, rows.Err()
}
