package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditRepository handles persistence of sync runs and operations
type AuditRepository struct {
	pool *pgxpool.Pool
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{
		pool: pool,
	}
}

// CreateSyncRun creates a new sync run record
func (r *AuditRepository) CreateSyncRun(ctx context.Context, params CreateSyncRunParams) (*SyncRun, error) {
	var metadata []byte
	var err error
	if params.Metadata != nil {
		metadata, err = json.Marshal(params.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	syncRun := &SyncRun{
		SyncID:      uuid.New(),
		Status:      SyncStatusPending,
		ConfigHash:  params.ConfigHash,
		ConfigPath:  params.ConfigPath,
		TriggeredBy: params.TriggeredBy,
		StartedAt:   time.Now(),
	}

	query := `
		INSERT INTO sync_runs (sync_id, status, config_hash, config_path, triggered_by, metadata, started_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err = r.pool.QueryRow(ctx, query,
		syncRun.SyncID,
		syncRun.Status,
		syncRun.ConfigHash,
		syncRun.ConfigPath,
		syncRun.TriggeredBy,
		metadata,
		syncRun.StartedAt,
	).Scan(&syncRun.ID, &syncRun.CreatedAt, &syncRun.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create sync run: %w", err)
	}

	if params.Metadata != nil {
		syncRun.Metadata = params.Metadata
	}

	return syncRun, nil
}

// UpdateSyncRun updates a sync run with completion status
func (r *AuditRepository) UpdateSyncRun(ctx context.Context, id int64, params UpdateSyncRunParams) error {
	completedAt := time.Now()

	query := `
		UPDATE sync_runs
		SET status = $1,
		    total_operations = $2,
		    successful_operations = $3,
		    failed_operations = $4,
		    duration_ms = $5,
		    error_message = $6,
		    completed_at = $7
		WHERE id = $8
	`

	_, err := r.pool.Exec(ctx, query,
		params.Status,
		params.TotalOperations,
		params.SuccessfulOperations,
		params.FailedOperations,
		params.DurationMs,
		params.ErrorMessage,
		completedAt,
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update sync run: %w", err)
	}

	return nil
}

// SetSyncRunStatus updates only the status of a sync run
func (r *AuditRepository) SetSyncRunStatus(ctx context.Context, id int64, status SyncStatus) error {
	query := `UPDATE sync_runs SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update sync run status: %w", err)
	}
	return nil
}

// GetSyncRun retrieves a sync run by ID
func (r *AuditRepository) GetSyncRun(ctx context.Context, id int64) (*SyncRun, error) {
	query := `
		SELECT id, sync_id, started_at, completed_at, status, config_hash, config_path,
		       total_operations, successful_operations, failed_operations, duration_ms,
		       error_message, triggered_by, metadata, created_at, updated_at
		FROM sync_runs
		WHERE id = $1
	`

	var sr SyncRun
	var metadata []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&sr.ID,
		&sr.SyncID,
		&sr.StartedAt,
		&sr.CompletedAt,
		&sr.Status,
		&sr.ConfigHash,
		&sr.ConfigPath,
		&sr.TotalOperations,
		&sr.SuccessfulOperations,
		&sr.FailedOperations,
		&sr.DurationMs,
		&sr.ErrorMessage,
		&sr.TriggeredBy,
		&metadata,
		&sr.CreatedAt,
		&sr.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("sync run not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get sync run: %w", err)
	}

	if metadata != nil {
		if err := json.Unmarshal(metadata, &sr.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &sr, nil
}

// GetSyncRunBySyncID retrieves a sync run by sync_id (UUID)
func (r *AuditRepository) GetSyncRunBySyncID(ctx context.Context, syncID uuid.UUID) (*SyncRun, error) {
	query := `
		SELECT id, sync_id, started_at, completed_at, status, config_hash, config_path,
		       total_operations, successful_operations, failed_operations, duration_ms,
		       error_message, triggered_by, metadata, created_at, updated_at
		FROM sync_runs
		WHERE sync_id = $1
	`

	var sr SyncRun
	var metadata []byte

	err := r.pool.QueryRow(ctx, query, syncID).Scan(
		&sr.ID,
		&sr.SyncID,
		&sr.StartedAt,
		&sr.CompletedAt,
		&sr.Status,
		&sr.ConfigHash,
		&sr.ConfigPath,
		&sr.TotalOperations,
		&sr.SuccessfulOperations,
		&sr.FailedOperations,
		&sr.DurationMs,
		&sr.ErrorMessage,
		&sr.TriggeredBy,
		&metadata,
		&sr.CreatedAt,
		&sr.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("sync run not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get sync run: %w", err)
	}

	if metadata != nil {
		if err := json.Unmarshal(metadata, &sr.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &sr, nil
}

// ListSyncRuns retrieves sync runs with optional filters
func (r *AuditRepository) ListSyncRuns(ctx context.Context, filter SyncRunFilter) ([]*SyncRun, error) {
	query := `
		SELECT id, sync_id, started_at, completed_at, status, config_hash, config_path,
		       total_operations, successful_operations, failed_operations, duration_ms,
		       error_message, triggered_by, metadata, created_at, updated_at
		FROM sync_runs
		WHERE 1=1
	`

	args := []interface{}{}
	argNum := 1

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, *filter.Status)
		argNum++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND started_at >= $%d", argNum)
		args = append(args, *filter.StartDate)
		argNum++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND started_at <= $%d", argNum)
		args = append(args, *filter.EndDate)
		argNum++
	}

	query += " ORDER BY started_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filter.Limit)
		argNum++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filter.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list sync runs: %w", err)
	}
	defer rows.Close()

	var syncRuns []*SyncRun
	for rows.Next() {
		var sr SyncRun
		var metadata []byte

		err := rows.Scan(
			&sr.ID,
			&sr.SyncID,
			&sr.StartedAt,
			&sr.CompletedAt,
			&sr.Status,
			&sr.ConfigHash,
			&sr.ConfigPath,
			&sr.TotalOperations,
			&sr.SuccessfulOperations,
			&sr.FailedOperations,
			&sr.DurationMs,
			&sr.ErrorMessage,
			&sr.TriggeredBy,
			&metadata,
			&sr.CreatedAt,
			&sr.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan sync run: %w", err)
		}

		if metadata != nil {
			if err := json.Unmarshal(metadata, &sr.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		syncRuns = append(syncRuns, &sr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sync runs: %w", err)
	}

	return syncRuns, nil
}

// CreateOperation creates a new operation record
func (r *AuditRepository) CreateOperation(ctx context.Context, params CreateOperationParams) (*SyncOperation, error) {
	op := &SyncOperation{
		SyncRunID:     params.SyncRunID,
		OperationType: params.OperationType,
		TargetObject:  params.TargetObject,
		SQLStatement:  params.SQLStatement,
		Status:        OperationStatusPending,
	}

	query := `
		INSERT INTO sync_operations (sync_run_id, operation_type, target_object, sql_statement, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.pool.QueryRow(ctx, query,
		op.SyncRunID,
		op.OperationType,
		op.TargetObject,
		op.SQLStatement,
		op.Status,
	).Scan(&op.ID, &op.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create operation: %w", err)
	}

	return op, nil
}

// UpdateOperation updates an operation with execution results
func (r *AuditRepository) UpdateOperation(ctx context.Context, id int64, params UpdateOperationParams) error {
	executedAt := time.Now()

	query := `
		UPDATE sync_operations
		SET status = $1,
		    error_message = $2,
		    execution_time_ms = $3,
		    executed_at = $4
		WHERE id = $5
	`

	_, err := r.pool.Exec(ctx, query,
		params.Status,
		params.ErrorMessage,
		params.ExecutionTimeMs,
		executedAt,
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update operation: %w", err)
	}

	return nil
}

// GetOperationsBySyncRun retrieves all operations for a sync run
func (r *AuditRepository) GetOperationsBySyncRun(ctx context.Context, syncRunID int64) ([]*SyncOperation, error) {
	query := `
		SELECT id, sync_run_id, operation_type, target_object, sql_statement,
		       status, error_message, execution_time_ms, executed_at, created_at
		FROM sync_operations
		WHERE sync_run_id = $1
		ORDER BY id ASC
	`

	rows, err := r.pool.Query(ctx, query, syncRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get operations: %w", err)
	}
	defer rows.Close()

	var operations []*SyncOperation
	for rows.Next() {
		var op SyncOperation

		err := rows.Scan(
			&op.ID,
			&op.SyncRunID,
			&op.OperationType,
			&op.TargetObject,
			&op.SQLStatement,
			&op.Status,
			&op.ErrorMessage,
			&op.ExecutionTimeMs,
			&op.ExecutedAt,
			&op.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan operation: %w", err)
		}

		operations = append(operations, &op)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating operations: %w", err)
	}

	return operations, nil
}

// CleanupOldSyncRuns deletes sync runs older than 30 days
func (r *AuditRepository) CleanupOldSyncRuns(ctx context.Context) (int64, error) {
	query := `SELECT * FROM cleanup_old_sync_runs()`

	var deletedCount int64
	err := r.pool.QueryRow(ctx, query).Scan(&deletedCount)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old sync runs: %w", err)
	}

	return deletedCount, nil
}

// Ping checks if the database connection is healthy
func (r *AuditRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

// Close closes the database connection pool
func (r *AuditRepository) Close() {
	r.pool.Close()
}
