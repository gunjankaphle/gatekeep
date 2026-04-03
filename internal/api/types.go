package api

import (
	"time"

	"github.com/google/uuid"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// RoleResponse represents a role in API responses
type RoleResponse struct {
	Name        string   `json:"name"`
	ParentRoles []string `json:"parent_roles,omitempty"`
	Comment     string   `json:"comment,omitempty"`
}

// RolesListResponse represents the response for listing roles
type RolesListResponse struct {
	Roles []RoleResponse `json:"roles"`
	Count int            `json:"count"`
}

// SyncRequest represents a sync request
type SyncRequest struct {
	ConfigPath string `json:"config_path"`
	DryRun     bool   `json:"dry_run"`
}

// SyncResponse represents the response from a sync operation
type SyncResponse struct {
	SyncID             uuid.UUID       `json:"sync_id"`
	Status             string          `json:"status"`
	OperationsExecuted int             `json:"operations_executed"`
	OperationsFailed   int             `json:"operations_failed"`
	DurationMs         int64           `json:"duration_ms"`
	DetailsURL         string          `json:"details_url,omitempty"`
	Operations         []OperationInfo `json:"operations,omitempty"`
}

// OperationInfo represents information about a single operation
type OperationInfo struct {
	Type   string `json:"type"`
	Target string `json:"target"`
	SQL    string `json:"sql"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// SyncHistoryResponse represents sync history listing
type SyncHistoryResponse struct {
	SyncRuns   []SyncRunInfo `json:"sync_runs"`
	TotalCount int           `json:"total_count"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
}

// SyncRunInfo represents summary information about a sync run
type SyncRunInfo struct {
	ID                   int64      `json:"id"`
	SyncID               uuid.UUID  `json:"sync_id"`
	StartedAt            time.Time  `json:"started_at"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	Status               string     `json:"status"`
	ConfigPath           string     `json:"config_path,omitempty"`
	TotalOperations      int        `json:"total_operations"`
	SuccessfulOperations int        `json:"successful_operations"`
	FailedOperations     int        `json:"failed_operations"`
	DurationMs           *int64     `json:"duration_ms,omitempty"`
}

// SyncRunDetailResponse represents detailed sync run information
type SyncRunDetailResponse struct {
	SyncRunInfo
	Operations []OperationDetail `json:"operations"`
}

// OperationDetail represents detailed operation information
type OperationDetail struct {
	ID              int64      `json:"id"`
	OperationType   string     `json:"operation_type"`
	TargetObject    string     `json:"target_object"`
	SQLStatement    string     `json:"sql_statement"`
	Status          string     `json:"status"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
	ExecutionTimeMs *int       `json:"execution_time_ms,omitempty"`
	ExecutedAt      *time.Time `json:"executed_at,omitempty"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}
