package snowflake

import (
	"context"
	"fmt"
	"time"
)

// Executor executes SQL statements in Snowflake
type Executor struct {
	client Client
}

// NewExecutor creates a new SQL executor
func NewExecutor(client Client) *Executor {
	return &Executor{client: client}
}

// ExecutionResult represents the result of executing a SQL statement
type ExecutionResult struct {
	SQL           string
	Success       bool
	RowsAffected  int64
	ExecutionTime time.Duration
	Error         error
}

// Execute runs a single SQL statement
func (e *Executor) Execute(ctx context.Context, sql string) ExecutionResult {
	start := time.Now()

	result := ExecutionResult{
		SQL:     sql,
		Success: false,
	}

	// Execute the SQL statement
	sqlResult, err := e.client.Exec(sql)
	result.ExecutionTime = time.Since(start)

	if err != nil {
		result.Error = fmt.Errorf("failed to execute SQL: %w", err)
		return result
	}

	// Get rows affected (if applicable)
	rowsAffected, err := sqlResult.RowsAffected()
	if err != nil {
		// RowsAffected may not be supported for all statement types (e.g., CREATE ROLE)
		// This is not a critical error, so we continue
		result.RowsAffected = 0
	} else {
		result.RowsAffected = rowsAffected
	}

	result.Success = true
	return result
}

// ExecuteBatch executes multiple SQL statements sequentially
func (e *Executor) ExecuteBatch(ctx context.Context, statements []string) []ExecutionResult {
	results := make([]ExecutionResult, len(statements))

	for i, sql := range statements {
		// Check context cancellation
		select {
		case <-ctx.Done():
			// Context cancelled, mark remaining as skipped
			for j := i; j < len(statements); j++ {
				results[j] = ExecutionResult{
					SQL:     statements[j],
					Success: false,
					Error:   ctx.Err(),
				}
			}
			return results
		default:
			results[i] = e.Execute(ctx, sql)
		}
	}

	return results
}

// Validate checks if a SQL statement is valid without executing it
func (e *Executor) Validate(sql string) error {
	// Basic validation - check for empty statements
	if len(sql) == 0 {
		return fmt.Errorf("empty SQL statement")
	}

	// Additional validation could be added here
	// For now, we rely on Snowflake to validate the SQL

	return nil
}
