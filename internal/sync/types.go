package sync

import (
	"time"

	"github.com/yourusername/gatekeep/internal/diff"
)

// Mode defines the execution mode
type Mode string

const (
	// ModeExecute executes SQL statements in Snowflake
	ModeExecute Mode = "execute"
	// ModeDryRun generates SQL without executing
	ModeDryRun Mode = "dry-run"
)

// Result represents the outcome of a sync operation
type Result struct {
	SyncID            string
	Status            Status
	OperationsTotal   int
	OperationsSuccess int
	OperationsFailed  int
	DurationMs        int64
	StartedAt         time.Time
	CompletedAt       time.Time
	Operations        []OperationResult
	ErrorMessage      string
}

// Status represents the overall sync status
type Status string

// Sync status constants
const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
	StatusPartial Status = "partial" // Some operations failed
)

// OperationResult represents the result of a single SQL operation
type OperationResult struct {
	Operation     diff.SQLOperation
	Status        OperationStatus
	ExecutionTime time.Duration
	Error         string
	ExecutedAt    time.Time
}

// OperationStatus represents the status of a single operation
type OperationStatus string

// Operation status constants
const (
	OpStatusPending OperationStatus = "pending"
	OpStatusRunning OperationStatus = "running"
	OpStatusSuccess OperationStatus = "success"
	OpStatusFailed  OperationStatus = "failed"
	OpStatusSkipped OperationStatus = "skipped"
)

// Config holds sync execution configuration
type Config struct {
	Mode                  Mode
	Workers               int           // Number of parallel workers (default: 10)
	Timeout               time.Duration // Operation timeout (default: 30s)
	FailureThreshold      float64       // Stop if failure rate exceeds this (default: 0.20 = 20%)
	ContinueOnError       bool          // Continue executing even if some operations fail
	ProgressCallback      func(OperationResult)
	CircuitBreakerEnabled bool // Enable circuit breaker (default: true)
}

// DefaultConfig returns default sync configuration
func DefaultConfig() Config {
	return Config{
		Mode:                  ModeExecute,
		Workers:               10,
		Timeout:               30 * time.Second,
		FailureThreshold:      0.20, // 20%
		ContinueOnError:       true,
		CircuitBreakerEnabled: true,
	}
}

// DryRunConfig returns configuration for dry-run mode
func DryRunConfig() Config {
	cfg := DefaultConfig()
	cfg.Mode = ModeDryRun
	return cfg
}
