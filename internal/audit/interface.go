package audit

import (
	"context"
	"time"

	"github.com/yourusername/gatekeep/internal/repository"
)

// AuditLogger is an interface for audit logging
//
//nolint:revive // AuditLogger is intentionally named for clarity despite stuttering
type AuditLogger interface {
	// StartSync creates a new sync run and returns its ID
	StartSync(ctx context.Context, configPath string, configContent []byte, triggeredBy *string) (int64, error)

	// SetSyncRunning marks a sync run as running
	SetSyncRunning(ctx context.Context, syncRunID int64) error

	// CompleteSync marks a sync run as complete with final statistics
	CompleteSync(ctx context.Context, syncRunID int64, totalOps, successOps, failedOps int, startTime time.Time) error

	// FailSync marks a sync run as failed with an error message
	FailSync(ctx context.Context, syncRunID int64, errorMsg string, startTime time.Time) error

	// LogOperation records an operation to be executed
	LogOperation(ctx context.Context, syncRunID int64, opType, target, sql string) (int64, error)

	// RecordOperationSuccess marks an operation as successful
	RecordOperationSuccess(ctx context.Context, opID int64, executionTimeMs int) error

	// RecordOperationFailure marks an operation as failed with an error
	RecordOperationFailure(ctx context.Context, opID int64, errorMsg string, executionTimeMs int) error

	// GetSyncHistory retrieves recent sync runs
	GetSyncHistory(ctx context.Context, limit int) ([]*repository.SyncRun, error)

	// GetSyncRunDetails retrieves a sync run with all its operations
	GetSyncRunDetails(ctx context.Context, syncRunID int64) (*repository.SyncRun, []*repository.SyncOperation, error)

	// CleanupOldLogs removes audit logs older than 30 days
	CleanupOldLogs(ctx context.Context) (int64, error)
}
