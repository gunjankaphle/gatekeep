package snowflake

import (
	"context"
	"database/sql"
	"testing"
)

func TestExecutor_Execute_Success(t *testing.T) {
	mockClient := &MockClient{
		ExecFunc: func(query string) (sql.Result, error) {
			return &MockResult{AffectedRows: 1}, nil
		},
	}

	executor := NewExecutor(mockClient)
	result := executor.Execute(context.Background(), `CREATE ROLE "TEST"`)

	if !result.Success {
		t.Error("expected success")
	}

	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}

	if result.RowsAffected != 1 {
		t.Errorf("expected 1 row affected, got %d", result.RowsAffected)
	}
}

func TestExecutor_Execute_Failure(t *testing.T) {
	mockClient := &MockClient{
		ExecFunc: func(query string) (sql.Result, error) {
			return nil, &MockError{Msg: "test error"}
		},
	}

	executor := NewExecutor(mockClient)
	result := executor.Execute(context.Background(), `CREATE ROLE "TEST"`)

	if result.Success {
		t.Error("expected failure")
	}

	if result.Error == nil {
		t.Error("expected error")
	}
}

func TestExecutor_ExecuteBatch(t *testing.T) {
	execCount := 0
	mockClient := &MockClient{
		ExecFunc: func(query string) (sql.Result, error) {
			execCount++
			return &MockResult{AffectedRows: 1}, nil
		},
	}

	executor := NewExecutor(mockClient)
	statements := []string{
		`CREATE ROLE "ROLE1"`,
		`CREATE ROLE "ROLE2"`,
		`CREATE ROLE "ROLE3"`,
	}

	results := executor.ExecuteBatch(context.Background(), statements)

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	if execCount != 3 {
		t.Errorf("expected 3 executions, got %d", execCount)
	}

	for i, result := range results {
		if !result.Success {
			t.Errorf("result %d: expected success", i)
		}
	}
}

func TestExecutor_ExecuteBatch_ContextCancelled(t *testing.T) {
	execCount := 0
	mockClient := &MockClient{
		ExecFunc: func(query string) (sql.Result, error) {
			execCount++
			return &MockResult{AffectedRows: 1}, nil
		},
	}

	executor := NewExecutor(mockClient)
	statements := []string{
		`CREATE ROLE "ROLE1"`,
		`CREATE ROLE "ROLE2"`,
		`CREATE ROLE "ROLE3"`,
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	results := executor.ExecuteBatch(ctx, statements)

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// All should be cancelled
	for i, result := range results {
		if result.Success {
			t.Errorf("result %d: expected failure due to cancelled context", i)
		}
	}

	// No statements should have executed
	if execCount != 0 {
		t.Errorf("expected 0 executions (context cancelled), got %d", execCount)
	}
}

func TestExecutor_Validate(t *testing.T) {
	mockClient := &MockClient{}
	executor := NewExecutor(mockClient)

	// Test valid SQL
	err := executor.Validate(`CREATE ROLE "TEST"`)
	if err != nil {
		t.Errorf("expected no error for valid SQL, got %v", err)
	}

	// Test empty SQL
	err = executor.Validate("")
	if err == nil {
		t.Error("expected error for empty SQL")
	}
}
