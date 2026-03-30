package snowflake

import (
	"fmt"
	"os"
)

// LoadConfigFromEnv loads Snowflake configuration from environment variables
func LoadConfigFromEnv() (Config, error) {
	cfg := Config{
		Account:   os.Getenv("SNOWFLAKE_ACCOUNT"),
		User:      os.Getenv("SNOWFLAKE_USER"),
		Password:  os.Getenv("SNOWFLAKE_PASSWORD"),
		Database:  os.Getenv("SNOWFLAKE_DATABASE"),
		Warehouse: os.Getenv("SNOWFLAKE_WAREHOUSE"),
		Role:      os.Getenv("SNOWFLAKE_ROLE"),
	}

	// Validate required fields
	if cfg.Account == "" {
		return cfg, fmt.Errorf("SNOWFLAKE_ACCOUNT environment variable is required")
	}
	if cfg.User == "" {
		return cfg, fmt.Errorf("SNOWFLAKE_USER environment variable is required")
	}
	if cfg.Password == "" {
		return cfg, fmt.Errorf("SNOWFLAKE_PASSWORD environment variable is required")
	}

	// Set defaults for optional fields
	if cfg.Role == "" {
		cfg.Role = "ACCOUNTADMIN"
	}

	return cfg, nil
}
