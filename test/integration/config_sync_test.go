package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/gatekeep/internal/config"
	"github.com/yourusername/gatekeep/internal/snowflake"
)

// TestConfigParserWithLocalStack tests parsing YAML configs and validating them
func TestConfigParserWithLocalStack(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		wantRoles   int
		wantUsers   int
		wantDBs     int
		expectError bool
	}{
		{
			name:        "simple config",
			configFile:  "simple_config.yaml",
			wantRoles:   3, // READ_ONLY_ROLE, ANALYST_ROLE, ENGINEER_ROLE
			wantUsers:   2, // analyst@example.com, engineer@example.com
			wantDBs:     1, // TEST_DB
			expectError: false,
		},
		{
			name:        "complex config with role hierarchy",
			configFile:  "complex_config.yaml",
			wantRoles:   7, // BASE_READ_ROLE, BASE_WRITE_ROLE, ANALYTICS_VIEWER, ANALYTICS_EDITOR, DATA_ENGINEER, DATA_SCIENTIST, ANALYTICS_ADMIN
			wantUsers:   5, // viewer, editor, engineer, scientist, admin
			wantDBs:     2, // PROD_DB, DEV_DB
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load config from test fixtures
			configPath := filepath.Join("..", "fixtures", tt.configFile)
			parser := config.NewParser()

			cfg, err := parser.ParseFile(configPath)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, cfg)

			// Validate counts
			assert.Len(t, cfg.Roles, tt.wantRoles, "unexpected number of roles")
			assert.Len(t, cfg.Users, tt.wantUsers, "unexpected number of users")
			assert.Len(t, cfg.Databases, tt.wantDBs, "unexpected number of databases")

			// Validate role names
			roleNames := make(map[string]bool)
			for _, role := range cfg.Roles {
				roleNames[role.Name] = true
			}

			t.Logf("✅ Parsed %d roles: %v", len(cfg.Roles), roleNames)
		})
	}
}

// TestSnowflakeConnectionToLocalStack tests connecting to LocalStack Snowflake
func TestSnowflakeConnectionToLocalStack(t *testing.T) {
	db, _ := SetupLocalStackSnowflake(t)
	defer db.Close()

	// Verify connection works
	ctx := context.Background()
	err := db.PingContext(ctx)
	require.NoError(t, err, "Failed to ping LocalStack Snowflake")

	// Query current database
	var currentDB string
	err = db.QueryRowContext(ctx, "SELECT CURRENT_DATABASE()").Scan(&currentDB)
	require.NoError(t, err)
	assert.NotEmpty(t, currentDB)

	t.Logf("✅ Connected to database: %s", currentDB)
}

// TestCreateRoleInLocalStack tests creating a role in LocalStack
func TestCreateRoleInLocalStack(t *testing.T) {
	db, _ := SetupLocalStackSnowflake(t)
	defer db.Close()

	ctx := context.Background()
	testRole := "TEST_INTEGRATION_ROLE"

	// Clean up if role exists
	defer CleanupSnowflake(t, db, []string{testRole})

	// Create role
	_, err := db.ExecContext(ctx, "CREATE ROLE "+testRole)
	require.NoError(t, err, "Failed to create role")

	// Verify role exists
	roles := GetCurrentRoles(t, db)
	assert.Contains(t, roles, testRole, "Role not found after creation")

	t.Logf("✅ Created and verified role: %s", testRole)
}

// TestSnowflakeStateReader tests reading Snowflake state using StateReader
func TestSnowflakeStateReader(t *testing.T) {
	db, client := SetupLocalStackSnowflake(t)
	defer db.Close()

	ctx := context.Background()
	testRoles := []string{"STATE_READER_ROLE_1", "STATE_READER_ROLE_2"}

	// Clean up
	defer CleanupSnowflake(t, db, testRoles)

	// Create test roles
	for _, role := range testRoles {
		_, err := db.ExecContext(ctx, "CREATE ROLE "+role)
		require.NoError(t, err, "Failed to create role")
	}

	// Use StateReader to read roles
	stateReader := snowflake.NewStateReader(client)
	roles, err := stateReader.ReadRoles()
	require.NoError(t, err, "Failed to read roles")

	// Verify our test roles are in the list
	roleMap := make(map[string]bool)
	for _, role := range roles {
		roleMap[role.Name] = true
	}

	for _, expectedRole := range testRoles {
		assert.True(t, roleMap[expectedRole], "Expected role %s not found in StateReader output", expectedRole)
	}

	t.Logf("✅ StateReader found %d roles including test roles", len(roles))
}

// TestFullSyncWorkflowWithLocalStack tests a complete sync workflow
func TestFullSyncWorkflowWithLocalStack(t *testing.T) {
	db, client := SetupLocalStackSnowflake(t)
	defer db.Close()

	// Load simple config
	configPath := filepath.Join("..", "fixtures", "simple_config.yaml")
	parser := config.NewParser()

	cfg, err := parser.ParseFile(configPath)
	require.NoError(t, err)

	// Extract role names from config for cleanup
	var configRoles []string
	for _, role := range cfg.Roles {
		configRoles = append(configRoles, role.Name)
	}

	// Extract user names from config for cleanup
	var configUsers []string
	for _, user := range cfg.Users {
		configUsers = append(configUsers, user.Name)
	}

	// Clean up at the end
	defer CleanupSnowflake(t, db, configRoles)
	defer CleanupUsers(t, db, configUsers)

	// Read current Snowflake state
	stateReader := snowflake.NewStateReader(client)
	currentRoles, err := stateReader.ReadRoles()
	require.NoError(t, err)

	t.Logf("📊 Current state: %d roles", len(currentRoles))
	t.Logf("📋 Desired state: %d roles", len(cfg.Roles))

	// For now, just validate that we can read state and parse config
	// Full sync testing will be added when sync orchestrator is integrated

	assert.NotNil(t, cfg)
	assert.NotNil(t, currentRoles)

	t.Logf("✅ Successfully read both desired (config) and actual (Snowflake) state")
}

// TestConfigValidation tests YAML config validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid simple config",
			yamlContent: `
version: 1.0
roles:
  - name: VALID_ROLE
    comment: "Valid role"
`,
			expectError: false,
		},
		{
			name: "missing version",
			yamlContent: `
roles:
  - name: TEST_ROLE
`,
			expectError: true,
			errorMsg:    "version",
		},
		{
			name: "empty config",
			yamlContent: `
version: 1.0
`,
			expectError: false, // Empty roles list is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write temp config file
			tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tt.yamlContent)
			require.NoError(t, err)
			tmpFile.Close()

			// Parse and validate
			parser := config.NewParser()
			_, err = parser.ParseFile(tmpFile.Name())

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
				if err != nil && tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
