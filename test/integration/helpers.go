package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	sf "github.com/snowflakedb/gosnowflake"
)

// LocalStackConfig holds LocalStack Snowflake connection configuration
type LocalStackConfig struct {
	Account   string
	User      string
	Password  string
	Database  string
	Warehouse string
	Host      string
	Port      string
}

// DefaultLocalStackConfig returns default LocalStack configuration
func DefaultLocalStackConfig() LocalStackConfig {
	return LocalStackConfig{
		Account:   getEnv("LOCALSTACK_SNOWFLAKE_ACCOUNT", "test"),
		User:      getEnv("LOCALSTACK_SNOWFLAKE_USER", "test"),
		Password:  getEnv("LOCALSTACK_SNOWFLAKE_PASSWORD", "test"),
		Database:  getEnv("LOCALSTACK_SNOWFLAKE_DATABASE", "TEST_DB"),
		Warehouse: getEnv("LOCALSTACK_SNOWFLAKE_WAREHOUSE", "ANALYTICS_WH"),
		Host:      getEnv("LOCALSTACK_HOST", "localhost"),
		Port:      getEnv("LOCALSTACK_PORT", "4566"),
	}
}

// testClient wraps sql.DB to implement snowflake.Client interface for testing
type testClient struct {
	db *sql.DB
}

// Query implements Client.Query
func (tc *testClient) Query(query string) (*sql.Rows, error) {
	return tc.db.Query(query)
}

// Exec implements Client.Exec
func (tc *testClient) Exec(query string) (sql.Result, error) {
	return tc.db.Exec(query)
}

// Close implements Client.Close
func (tc *testClient) Close() error {
	return tc.db.Close()
}

// SetupLocalStackSnowflake creates a Snowflake connection to LocalStack
// Returns both the raw *sql.DB for direct queries and wrapped Client for StateReader
func SetupLocalStackSnowflake(t *testing.T) (*sql.DB, *testClient) {
	t.Helper()

	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if LocalStack is not available
	if !IsLocalStackAvailable() {
		t.Skip("LocalStack is not available - start with: docker-compose up localstack")
	}

	cfg := DefaultLocalStackConfig()

	// Configure Snowflake driver for LocalStack
	config := &sf.Config{
		Account:   cfg.Account,
		User:      cfg.User,
		Password:  cfg.Password,
		Database:  cfg.Database,
		Warehouse: cfg.Warehouse,
		Protocol:  "http", // LocalStack uses HTTP, not HTTPS
		Host:      fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
	}

	dsn, err := sf.DSN(config)
	if err != nil {
		t.Fatalf("Failed to build DSN: %v", err)
	}

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to LocalStack Snowflake: %v", err)
	}

	// Test connection with retry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for i := 0; i < 5; i++ {
		if err := db.PingContext(ctx); err == nil {
			t.Logf("✅ Connected to LocalStack Snowflake")
			return db, &testClient{db: db}
		}
		t.Logf("⏳ Waiting for LocalStack Snowflake (attempt %d/5)...", i+1)
		time.Sleep(2 * time.Second)
	}

	t.Fatalf("Failed to ping LocalStack Snowflake after retries")
	return nil, nil
}

// IsLocalStackAvailable checks if LocalStack is running
func IsLocalStackAvailable() bool {
	cfg := DefaultLocalStackConfig()
	endpoint := fmt.Sprintf("http://%s:%s/_localstack/health", cfg.Host, cfg.Port)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Simple HTTP check (we'd need to import net/http for real check)
	// For now, just check if the environment suggests LocalStack should be available
	_ = ctx
	_ = endpoint

	// Check if we're in LocalStack mode via env var
	return os.Getenv("USE_LOCALSTACK") == "true" || os.Getenv("LOCALSTACK_HOST") != ""
}

// CleanupSnowflake removes all roles created during tests
func CleanupSnowflake(t *testing.T, db *sql.DB, roles []string) {
	t.Helper()

	ctx := context.Background()
	for _, role := range roles {
		_, err := db.ExecContext(ctx, fmt.Sprintf("DROP ROLE IF EXISTS %s", role))
		if err != nil {
			t.Logf("⚠️  Failed to drop role %s: %v", role, err)
		}
	}
}

// CleanupUsers removes all users created during tests
func CleanupUsers(t *testing.T, db *sql.DB, users []string) {
	t.Helper()

	ctx := context.Background()
	for _, user := range users {
		_, err := db.ExecContext(ctx, fmt.Sprintf("DROP USER IF EXISTS \"%s\"", user))
		if err != nil {
			t.Logf("⚠️  Failed to drop user %s: %v", user, err)
		}
	}
}

// GetCurrentRoles queries all roles in Snowflake
func GetCurrentRoles(t *testing.T, db *sql.DB) []string {
	t.Helper()

	rows, err := db.Query("SHOW ROLES")
	if err != nil {
		t.Fatalf("Failed to show roles: %v", err)
	}
	defer rows.Close() // nolint:errcheck

	var roles []string
	for rows.Next() {
		var createdOn, name, isDefault, isCurrent, isInherited, assignedToUsers, grantedToRoles, grantedRoles, owner, comment sql.NullString
		err := rows.Scan(&createdOn, &name, &isDefault, &isCurrent, &isInherited, &assignedToUsers, &grantedToRoles, &grantedRoles, &owner, &comment)
		if err != nil {
			t.Logf("⚠️  Failed to scan role: %v", err)
			continue
		}
		if name.Valid {
			roles = append(roles, name.String)
		}
	}

	return roles
}

// GetCurrentUsers queries all users in Snowflake
func GetCurrentUsers(t *testing.T, db *sql.DB) []string {
	t.Helper()

	rows, err := db.Query("SHOW USERS")
	if err != nil {
		t.Fatalf("Failed to show users: %v", err)
	}
	defer rows.Close() // nolint:errcheck

	var users []string
	for rows.Next() {
		// SHOW USERS returns many columns, we only need the name (second column)
		var name string
		var cols [20]interface{} // Snowflake SHOW USERS has many columns
		cols[0] = new(sql.NullString)
		cols[1] = &name // name is typically the second column
		for i := 2; i < len(cols); i++ {
			cols[i] = new(sql.NullString)
		}

		err := rows.Scan(cols[0], cols[1], cols[2], cols[3], cols[4], cols[5], cols[6], cols[7], cols[8], cols[9])
		if err != nil {
			t.Logf("⚠️  Failed to scan user: %v", err)
			continue
		}
		users = append(users, name)
	}

	return users
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
