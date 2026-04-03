package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/yourusername/gatekeep/internal/repository"
)

// Logger provides high-level audit logging functionality
type Logger struct {
	repo *repository.AuditRepository
}

// NewLogger creates a new audit logger
func NewLogger(repo *repository.AuditRepository) *Logger {
	return &Logger{
		repo: repo,
	}
}

// StartSync creates a new sync run and returns its ID
func (l *Logger) StartSync(ctx context.Context, configPath string, configContent []byte, triggeredBy *string) (int64, error) {
	configHash := computeConfigHash(configContent)

	params := repository.CreateSyncRunParams{
		ConfigHash:  configHash,
		ConfigPath:  configPath,
		TriggeredBy: triggeredBy,
	}

	syncRun, err := l.repo.CreateSyncRun(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("failed to start sync: %w", err)
	}

	return syncRun.ID, nil
}

// SetSyncRunning marks a sync run as running
func (l *Logger) SetSyncRunning(ctx context.Context, syncRunID int64) error {
	return l.repo.SetSyncRunStatus(ctx, syncRunID, repository.SyncStatusRunning)
}

// CompleteSync marks a sync run as complete with final statistics
func (l *Logger) CompleteSync(ctx context.Context, syncRunID int64, totalOps, successOps, failedOps int, startTime time.Time) error {
	durationMs := time.Since(startTime).Milliseconds()

	var status repository.SyncStatus
	if failedOps == 0 {
		status = repository.SyncStatusSuccess
	} else if successOps > 0 {
		status = repository.SyncStatusPartial
	} else {
		status = repository.SyncStatusFailed
	}

	params := repository.UpdateSyncRunParams{
		Status:               status,
		TotalOperations:      totalOps,
		SuccessfulOperations: successOps,
		FailedOperations:     failedOps,
		DurationMs:           durationMs,
	}

	return l.repo.UpdateSyncRun(ctx, syncRunID, params)
}

// FailSync marks a sync run as failed with an error message
func (l *Logger) FailSync(ctx context.Context, syncRunID int64, errorMsg string, startTime time.Time) error {
	durationMs := time.Since(startTime).Milliseconds()
	errMsgPtr := &errorMsg

	params := repository.UpdateSyncRunParams{
		Status:       repository.SyncStatusFailed,
		ErrorMessage: errMsgPtr,
		DurationMs:   durationMs,
	}

	return l.repo.UpdateSyncRun(ctx, syncRunID, params)
}

// LogOperation records an operation to be executed
func (l *Logger) LogOperation(ctx context.Context, syncRunID int64, opType, target, sql string) (int64, error) {
	params := repository.CreateOperationParams{
		SyncRunID:     syncRunID,
		OperationType: opType,
		TargetObject:  target,
		SQLStatement:  sql,
	}

	op, err := l.repo.CreateOperation(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("failed to log operation: %w", err)
	}

	return op.ID, nil
}

// RecordOperationSuccess marks an operation as successful
func (l *Logger) RecordOperationSuccess(ctx context.Context, opID int64, executionTimeMs int) error {
	params := repository.UpdateOperationParams{
		Status:          repository.OperationStatusSuccess,
		ExecutionTimeMs: executionTimeMs,
	}

	return l.repo.UpdateOperation(ctx, opID, params)
}

// RecordOperationFailure marks an operation as failed with an error
func (l *Logger) RecordOperationFailure(ctx context.Context, opID int64, errorMsg string, executionTimeMs int) error {
	errMsgPtr := &errorMsg

	params := repository.UpdateOperationParams{
		Status:          repository.OperationStatusFailed,
		ErrorMessage:    errMsgPtr,
		ExecutionTimeMs: executionTimeMs,
	}

	return l.repo.UpdateOperation(ctx, opID, params)
}

// GetSyncHistory retrieves recent sync runs
func (l *Logger) GetSyncHistory(ctx context.Context, limit int) ([]*repository.SyncRun, error) {
	filter := repository.SyncRunFilter{
		Limit: limit,
	}

	return l.repo.ListSyncRuns(ctx, filter)
}

// GetSyncRunDetails retrieves a sync run with all its operations
func (l *Logger) GetSyncRunDetails(ctx context.Context, syncRunID int64) (*repository.SyncRun, []*repository.SyncOperation, error) {
	syncRun, err := l.repo.GetSyncRun(ctx, syncRunID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get sync run: %w", err)
	}

	operations, err := l.repo.GetOperationsBySyncRun(ctx, syncRunID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get operations: %w", err)
	}

	return syncRun, operations, nil
}

// CleanupOldLogs removes audit logs older than 30 days
func (l *Logger) CleanupOldLogs(ctx context.Context) (int64, error) {
	return l.repo.CleanupOldSyncRuns(ctx)
}

// computeConfigHash computes SHA256 hash of config content
func computeConfigHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}
