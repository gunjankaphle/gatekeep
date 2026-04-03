package sync

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/yourusername/gatekeep/internal/config"
	"github.com/yourusername/gatekeep/internal/diff"
	"github.com/yourusername/gatekeep/internal/snowflake"
)

func TestOrchestrator_Sync_ConfigParseError(t *testing.T) {
	parser := config.NewParser()
	mockClient := &snowflake.MockClient{}
	stateReader := snowflake.NewStateReader(mockClient)
	executor := snowflake.NewExecutor(mockClient)

	orchestrator := NewOrchestrator(parser, stateReader, executor, diff.SyncModeStrict)

	// Try to sync with non-existent file
	result, err := orchestrator.Sync(context.Background(), "nonexistent.yaml", DefaultConfig())

	if err == nil {
		t.Error("expected error for non-existent config file")
	}

	if result.Status != StatusFailed {
		t.Errorf("expected failed status, got %s", result.Status)
	}

	if result.ErrorMessage == "" {
		t.Error("expected error message")
	}
}

func TestOrchestrator_Sync_ConfigValidationError(t *testing.T) {
	parser := config.NewParser()
	mockClient := &snowflake.MockClient{}
	stateReader := snowflake.NewStateReader(mockClient)
	executor := snowflake.NewExecutor(mockClient)

	orchestrator := NewOrchestrator(parser, stateReader, executor, diff.SyncModeStrict)

	// Create invalid config file (missing version)
	tmpfile, err := os.CreateTemp("", "invalid-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	invalidConfig := `
roles:
  - name: TEST_ROLE
`
	if _, writeErr := tmpfile.Write([]byte(invalidConfig)); writeErr != nil {
		t.Fatal(writeErr)
	}
	tmpfile.Close()

	result, err := orchestrator.Sync(context.Background(), tmpfile.Name(), DefaultConfig())

	if err == nil {
		t.Error("expected validation error")
	}

	if result.Status != StatusFailed {
		t.Errorf("expected failed status, got %s", result.Status)
	}
}

func TestOrchestrator_Sync_StateReaderError(t *testing.T) {
	parser := config.NewParser()

	// Mock client that returns error when reading state
	mockClient := &snowflake.MockClient{
		QueryFunc: func(query string) (*sql.Rows, error) {
			return nil, &snowflake.MockError{Msg: "connection failed"}
		},
	}

	stateReader := snowflake.NewStateReader(mockClient)
	executor := snowflake.NewExecutor(mockClient)

	orchestrator := NewOrchestrator(parser, stateReader, executor, diff.SyncModeStrict)

	// Create valid config file
	tmpfile, err := os.CreateTemp("", "test-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	testConfig := `
version: "1.0"
roles:
  - name: TEST_ROLE
`
	if _, writeErr := tmpfile.Write([]byte(testConfig)); writeErr != nil {
		t.Fatal(writeErr)
	}
	tmpfile.Close()

	result, err := orchestrator.Sync(context.Background(), tmpfile.Name(), DefaultConfig())

	if err == nil {
		t.Error("expected error when state reader fails")
	}

	if result.Status != StatusFailed {
		t.Errorf("expected failed status, got %s", result.Status)
	}

	if result.ErrorMessage == "" {
		t.Error("expected error message")
	}
}

func TestOrchestrator_DryRun(t *testing.T) {
	parser := config.NewParser()
	mockClient := &snowflake.MockClient{}
	stateReader := snowflake.NewStateReader(mockClient)
	executor := snowflake.NewExecutor(mockClient)

	orchestrator := NewOrchestrator(parser, stateReader, executor, diff.SyncModeStrict)

	// Try dry-run with non-existent file
	result, err := orchestrator.DryRun(context.Background(), "nonexistent.yaml")

	if err == nil {
		t.Error("expected error for non-existent file")
	}

	if result.Status != StatusFailed {
		t.Errorf("expected failed status, got %s", result.Status)
	}
}

func TestGenerateSyncID(t *testing.T) {
	id1 := generateSyncID()
	id2 := generateSyncID()

	if id1 == "" {
		t.Error("expected non-empty sync ID")
	}

	if id1 == id2 {
		// IDs might be the same if generated in the same second, but check format
		if len(id1) < 5 {
			t.Error("sync ID seems too short")
		}
	}
}

func TestOrchestrator_NewOrchestrator(t *testing.T) {
	parser := config.NewParser()
	mockClient := &snowflake.MockClient{}
	stateReader := snowflake.NewStateReader(mockClient)
	executor := snowflake.NewExecutor(mockClient)

	orchestrator := NewOrchestrator(parser, stateReader, executor, diff.SyncModeStrict)

	if orchestrator.syncMode != diff.SyncModeStrict {
		t.Errorf("expected strict sync mode, got %s", orchestrator.syncMode)
	}

	// Verify orchestrator was properly initialized
	if orchestrator.configParser == nil {
		t.Error("expected config parser to be set")
	}
}
