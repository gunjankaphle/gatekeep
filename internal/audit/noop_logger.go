package audit

import (
	"context"
	"time"

	"github.com/yourusername/gatekeep/internal/repository"
)

// NoOpLogger is a no-op implementation of AuditLogger
// Used when PostgreSQL is not configured
type NoOpLogger struct{}

// NewNoOpLogger creates a new no-op audit logger
func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

// StartSync is a no-op
func (l *NoOpLogger) StartSync(ctx context.Context, configPath string, configContent []byte, triggeredBy *string) (int64, error) {
	// Return a dummy ID
	return 1, nil
}

// SetSyncRunning is a no-op
func (l *NoOpLogger) SetSyncRunning(ctx context.Context, syncRunID int64) error {
	return nil
}

// CompleteSync is a no-op
func (l *NoOpLogger) CompleteSync(ctx context.Context, syncRunID int64, totalOps, successOps, failedOps int, startTime time.Time) error {
	return nil
}

// FailSync is a no-op
func (l *NoOpLogger) FailSync(ctx context.Context, syncRunID int64, errorMsg string, startTime time.Time) error {
	return nil
}

// LogOperation is a no-op
func (l *NoOpLogger) LogOperation(ctx context.Context, syncRunID int64, opType, target, sql string) (int64, error) {
	// Return a dummy ID
	return 1, nil
}

// RecordOperationSuccess is a no-op
func (l *NoOpLogger) RecordOperationSuccess(ctx context.Context, opID int64, executionTimeMs int) error {
	return nil
}

// RecordOperationFailure is a no-op
func (l *NoOpLogger) RecordOperationFailure(ctx context.Context, opID int64, errorMsg string, executionTimeMs int) error {
	return nil
}

// GetSyncHistory returns an empty list
func (l *NoOpLogger) GetSyncHistory(ctx context.Context, limit int) ([]*repository.SyncRun, error) {
	return []*repository.SyncRun{}, nil
}

// GetSyncRunDetails returns nil (not found)
func (l *NoOpLogger) GetSyncRunDetails(ctx context.Context, syncRunID int64) (*repository.SyncRun, []*repository.SyncOperation, error) {
	return nil, nil, nil
}

// CleanupOldLogs returns 0 (nothing deleted)
func (l *NoOpLogger) CleanupOldLogs(ctx context.Context) (int64, error) {
	return 0, nil
}
