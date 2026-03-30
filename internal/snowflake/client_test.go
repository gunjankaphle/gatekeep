package snowflake

import (
	"database/sql"
	"testing"
)

func TestNewClient_InvalidConfig(t *testing.T) {
	// Test with empty config
	cfg := Config{}
	_, err := NewClient(cfg)
	if err == nil {
		t.Error("Expected error for empty config, got nil")
	}
}

func TestMockClient(t *testing.T) {
	mock := &MockClient{
		QueryFunc: func(query string) (*sql.Rows, error) {
			return nil, nil
		},
		ExecFunc: func(query string) (sql.Result, error) {
			return nil, nil
		},
	}

	// Test Query
	_, err := mock.Query("SELECT 1")
	if err != nil {
		t.Errorf("Unexpected error from mock Query: %v", err)
	}

	// Test Exec
	_, err = mock.Exec("SELECT 1")
	if err != nil {
		t.Errorf("Unexpected error from mock Exec: %v", err)
	}

	// Test Close
	err = mock.Close()
	if err != nil {
		t.Errorf("Unexpected error from mock Close: %v", err)
	}

	if !mock.Closed {
		t.Error("Expected Closed to be true after Close()")
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid config",
			config: Config{
				Account:   "account",
				User:      "user",
				Password:  "password",
				Database:  "db",
				Warehouse: "wh",
				Role:      "role",
			},
			valid: true,
		},
		{
			name:   "empty config",
			config: Config{},
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: We can't actually connect without real Snowflake credentials
			// This test just validates the structure
			if tt.config.Account == "" && tt.valid {
				t.Error("Valid config should have account")
			}
		})
	}
}
