package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParser_Parse_ValidConfig(t *testing.T) {
	yaml := `version: 1.0
roles:
  - name: ANALYST_ROLE
    comment: "Analysts"
  - name: ENGINEER_ROLE
    parent_roles: [ANALYST_ROLE]
    comment: "Engineers"

users:
  - name: alice@company.com
    roles: [ANALYST_ROLE]
  - name: bob@company.com
    roles: [ENGINEER_ROLE]

databases:
  - name: PROD_DB
    schemas:
      - name: PUBLIC
        tables:
          - name: CUSTOMERS
            grants:
              - to_role: ANALYST_ROLE
                privileges: [SELECT]

warehouses:
  - name: ANALYTICS_WH
    grants:
      - to_role: ANALYST_ROLE
        privileges: [USAGE]
`

	parser := NewParser()
	config, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Validate parsed config
	if config.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", config.Version)
	}

	if len(config.Roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(config.Roles))
	}

	if len(config.Users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(config.Users))
	}

	if len(config.Databases) != 1 {
		t.Errorf("Expected 1 database, got %d", len(config.Databases))
	}

	if len(config.Warehouses) != 1 {
		t.Errorf("Expected 1 warehouse, got %d", len(config.Warehouses))
	}
}

func TestParser_Parse_MissingVersion(t *testing.T) {
	yaml := `roles:
  - name: TEST_ROLE
`

	parser := NewParser()
	_, err := parser.Parse([]byte(yaml))
	if err == nil {
		t.Fatal("Expected error for missing version, got nil")
	}

	if !contains(err.Error(), "version is required") {
		t.Errorf("Expected 'version is required' error, got: %v", err)
	}
}

func TestParser_Parse_InvalidYAML(t *testing.T) {
	yaml := `version: 1.0
roles:
  - name: TEST
    invalid syntax here
`

	parser := NewParser()
	_, err := parser.Parse([]byte(yaml))
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
}

func TestParser_ParseFile(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yaml := `version: 1.0
roles:
  - name: TEST_ROLE
users:
  - name: test@example.com
    roles: [TEST_ROLE]
`

	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	parser := NewParser()
	config, err := parser.ParseFile(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", config.Version)
	}
}

func TestParser_ParseFile_NotFound(t *testing.T) {
	parser := NewParser()
	_, err := parser.ParseFile("/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}
}

func TestParser_ParseWithoutValidation(t *testing.T) {
	// This YAML has validation errors but should parse
	yaml := `version: 1.0
roles:
  - name: ROLE1
users:
  - name: invalid_email_format
    roles: [NONEXISTENT_ROLE]
`

	parser := NewParser()
	config, err := parser.ParseWithoutValidation([]byte(yaml))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", config.Version)
	}

	// Should have parsed the data even though it's invalid
	if len(config.Users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(config.Users))
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
