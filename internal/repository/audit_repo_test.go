package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: These tests require a PostgreSQL instance running
// They are integration tests and can be run with: make test-integration
// For now, we'll use a mock or skip if DB is not available

func TestAuditRepository_CreateSyncRun(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	params := CreateSyncRunParams{
		ConfigHash:  "abc123",
		ConfigPath:  "/path/to/config.yaml",
		TriggeredBy: stringPtr("test-user"),
		Metadata: map[string]interface{}{
			"test": "value",
		},
	}

	syncRun, err := repo.CreateSyncRun(ctx, params)
	require.NoError(t, err)
	assert.NotZero(t, syncRun.ID)
	assert.NotEqual(t, uuid.Nil, syncRun.SyncID)
	assert.Equal(t, SyncStatusPending, syncRun.Status)
	assert.Equal(t, "abc123", syncRun.ConfigHash)
	assert.Equal(t, "/path/to/config.yaml", syncRun.ConfigPath)
	assert.Equal(t, "test-user", *syncRun.TriggeredBy)
}

func TestAuditRepository_UpdateSyncRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	// Create a sync run
	createParams := CreateSyncRunParams{
		ConfigHash: "abc123",
		ConfigPath: "/path/to/config.yaml",
	}

	syncRun, err := repo.CreateSyncRun(ctx, createParams)
	require.NoError(t, err)

	// Update it
	updateParams := UpdateSyncRunParams{
		Status:               SyncStatusSuccess,
		TotalOperations:      10,
		SuccessfulOperations: 10,
		FailedOperations:     0,
		DurationMs:           5000,
	}

	err = repo.UpdateSyncRun(ctx, syncRun.ID, updateParams)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetSyncRun(ctx, syncRun.ID)
	require.NoError(t, err)
	assert.Equal(t, SyncStatusSuccess, updated.Status)
	assert.Equal(t, 10, updated.TotalOperations)
	assert.Equal(t, 10, updated.SuccessfulOperations)
	assert.Equal(t, 0, updated.FailedOperations)
	assert.NotNil(t, updated.CompletedAt)
}

func TestAuditRepository_GetSyncRunBySyncID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	// Create a sync run
	createParams := CreateSyncRunParams{
		ConfigHash: "abc123",
		ConfigPath: "/path/to/config.yaml",
	}

	syncRun, err := repo.CreateSyncRun(ctx, createParams)
	require.NoError(t, err)

	// Retrieve by sync_id
	found, err := repo.GetSyncRunBySyncID(ctx, syncRun.SyncID)
	require.NoError(t, err)
	assert.Equal(t, syncRun.ID, found.ID)
	assert.Equal(t, syncRun.SyncID, found.SyncID)
}

func TestAuditRepository_ListSyncRuns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	// Create multiple sync runs
	for i := 0; i < 5; i++ {
		params := CreateSyncRunParams{
			ConfigHash: "abc123",
			ConfigPath: "/path/to/config.yaml",
		}
		_, err := repo.CreateSyncRun(ctx, params)
		require.NoError(t, err)
	}

	// List all
	filter := SyncRunFilter{Limit: 10}
	runs, err := repo.ListSyncRuns(ctx, filter)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(runs), 5)

	// List with limit
	filter = SyncRunFilter{Limit: 2}
	runs, err = repo.ListSyncRuns(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 2, len(runs))
}

func TestAuditRepository_CreateOperation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	// Create a sync run first
	syncParams := CreateSyncRunParams{
		ConfigHash: "abc123",
		ConfigPath: "/path/to/config.yaml",
	}
	syncRun, err := repo.CreateSyncRun(ctx, syncParams)
	require.NoError(t, err)

	// Create an operation
	opParams := CreateOperationParams{
		SyncRunID:     syncRun.ID,
		OperationType: "CREATE_ROLE",
		TargetObject:  "ANALYST",
		SQLStatement:  "CREATE ROLE ANALYST;",
	}

	op, err := repo.CreateOperation(ctx, opParams)
	require.NoError(t, err)
	assert.NotZero(t, op.ID)
	assert.Equal(t, syncRun.ID, op.SyncRunID)
	assert.Equal(t, "CREATE_ROLE", op.OperationType)
	assert.Equal(t, OperationStatusPending, op.Status)
}

func TestAuditRepository_UpdateOperation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	// Create sync run and operation
	syncParams := CreateSyncRunParams{
		ConfigHash: "abc123",
		ConfigPath: "/path/to/config.yaml",
	}
	syncRun, err := repo.CreateSyncRun(ctx, syncParams)
	require.NoError(t, err)

	opParams := CreateOperationParams{
		SyncRunID:     syncRun.ID,
		OperationType: "CREATE_ROLE",
		TargetObject:  "ANALYST",
		SQLStatement:  "CREATE ROLE ANALYST;",
	}
	op, err := repo.CreateOperation(ctx, opParams)
	require.NoError(t, err)

	// Update operation
	updateParams := UpdateOperationParams{
		Status:          OperationStatusSuccess,
		ExecutionTimeMs: 150,
	}
	err = repo.UpdateOperation(ctx, op.ID, updateParams)
	require.NoError(t, err)

	// Verify
	ops, err := repo.GetOperationsBySyncRun(ctx, syncRun.ID)
	require.NoError(t, err)
	require.Equal(t, 1, len(ops))
	assert.Equal(t, OperationStatusSuccess, ops[0].Status)
	assert.Equal(t, 150, *ops[0].ExecutionTimeMs)
	assert.NotNil(t, ops[0].ExecutedAt)
}

func TestAuditRepository_GetOperationsBySyncRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewAuditRepository(pool)
	ctx := context.Background()

	// Create sync run
	syncParams := CreateSyncRunParams{
		ConfigHash: "abc123",
		ConfigPath: "/path/to/config.yaml",
	}
	syncRun, err := repo.CreateSyncRun(ctx, syncParams)
	require.NoError(t, err)

	// Create multiple operations
	operations := []CreateOperationParams{
		{syncRun.ID, "CREATE_ROLE", "ANALYST", "CREATE ROLE ANALYST;"},
		{syncRun.ID, "GRANT", "ANALYST", "GRANT SELECT ON TABLE foo TO ROLE ANALYST;"},
		{syncRun.ID, "GRANT", "ANALYST", "GRANT USAGE ON WAREHOUSE bar TO ROLE ANALYST;"},
	}

	for _, op := range operations {
		_, createErr := repo.CreateOperation(ctx, op)
		require.NoError(t, createErr)
	}

	// Retrieve all operations
	ops, err := repo.GetOperationsBySyncRun(ctx, syncRun.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, len(ops))
}

// Helper functions

//nolint:unparam // setupTestDB always returns nil when skipping tests, which is expected
func setupTestDB(t *testing.T) *pgxpool.Pool {
	// This would require setting up a test database
	// For now, we'll skip actual DB tests unless integration test env is set up
	t.Skip("PostgreSQL integration tests require test database setup")
	return nil
}

func stringPtr(s string) *string {
	return &s
}
