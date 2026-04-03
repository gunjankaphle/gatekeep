package repository

import (
	"time"

	"github.com/google/uuid"
)

// SyncRun represents a high-level sync execution
type SyncRun struct {
	ID                   int64
	SyncID               uuid.UUID
	StartedAt            time.Time
	CompletedAt          *time.Time
	Status               SyncStatus
	ConfigHash           string
	ConfigPath           string
	TotalOperations      int
	SuccessfulOperations int
	FailedOperations     int
	DurationMs           *int64
	ErrorMessage         *string
	TriggeredBy          *string
	Metadata             map[string]interface{}
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// SyncOperation represents an individual operation within a sync run
type SyncOperation struct {
	ID              int64
	SyncRunID       int64
	OperationType   string
	TargetObject    string
	SQLStatement    string
	Status          OperationStatus
	ErrorMessage    *string
	ExecutionTimeMs *int
	ExecutedAt      *time.Time
	CreatedAt       time.Time
}

// SyncStatus represents the status of a sync run
type SyncStatus string

// Sync status constants
const (
	SyncStatusPending SyncStatus = "pending"
	SyncStatusRunning SyncStatus = "running"
	SyncStatusSuccess SyncStatus = "success"
	SyncStatusFailed  SyncStatus = "failed"
	SyncStatusPartial SyncStatus = "partial"
)

// OperationStatus represents the status of an individual operation
type OperationStatus string

// Operation status constants
const (
	OperationStatusPending OperationStatus = "pending"
	OperationStatusSuccess OperationStatus = "success"
	OperationStatusFailed  OperationStatus = "failed"
	OperationStatusSkipped OperationStatus = "skipped"
)

// SyncRunFilter represents filters for querying sync runs
type SyncRunFilter struct {
	Status    *SyncStatus
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
}

// SyncRunSummary represents a summary of sync operations
type SyncRunSummary struct {
	SyncID             uuid.UUID
	Status             SyncStatus
	OperationsExecuted int
	OperationsFailed   int
	DurationMs         int64
	AuditURL           string
}

// CreateSyncRunParams represents parameters for creating a new sync run
type CreateSyncRunParams struct {
	ConfigHash  string
	ConfigPath  string
	TriggeredBy *string
	Metadata    map[string]interface{}
}

// UpdateSyncRunParams represents parameters for updating a sync run
type UpdateSyncRunParams struct {
	Status               SyncStatus
	TotalOperations      int
	SuccessfulOperations int
	FailedOperations     int
	DurationMs           int64
	ErrorMessage         *string
}

// CreateOperationParams represents parameters for creating an operation
type CreateOperationParams struct {
	SyncRunID     int64
	OperationType string
	TargetObject  string
	SQLStatement  string
}

// UpdateOperationParams represents parameters for updating an operation
type UpdateOperationParams struct {
	Status          OperationStatus
	ErrorMessage    *string
	ExecutionTimeMs int
}
