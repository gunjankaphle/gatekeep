package snowflake

import (
	"database/sql"
	"fmt"

	sf "github.com/snowflakedb/gosnowflake"
)

// Client interface for Snowflake operations (allows mocking)
type Client interface {
	Query(query string) (*sql.Rows, error)
	Exec(query string) (sql.Result, error)
	Close() error
}

// Config holds Snowflake connection configuration
type Config struct {
	Account   string
	User      string
	Password  string
	Database  string
	Warehouse string
	Role      string
}

// client implements the Client interface
type client struct {
	db *sql.DB
}

// NewClient creates a new Snowflake client
func NewClient(cfg Config) (Client, error) {
	// Build DSN (Data Source Name)
	dsn, err := sf.DSN(&sf.Config{
		Account:   cfg.Account,
		User:      cfg.User,
		Password:  cfg.Password,
		Database:  cfg.Database,
		Warehouse: cfg.Warehouse,
		Role:      cfg.Role,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build DSN: %w", err)
	}

	// Open connection
	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Test connection
	if err := db.Ping(); err != nil {
		//nolint:errcheck // Ignore error on cleanup
		_ = db.Close()
		return nil, fmt.Errorf("failed to connect to Snowflake: %w", err)
	}

	return &client{db: db}, nil
}

// Query executes a query that returns rows
func (c *client) Query(query string) (*sql.Rows, error) {
	return c.db.Query(query)
}

// Exec executes a query that doesn't return rows
func (c *client) Exec(query string) (sql.Result, error) {
	return c.db.Exec(query)
}

// Close closes the database connection
func (c *client) Close() error {
	return c.db.Close()
}
