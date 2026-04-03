package sync

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/yourusername/gatekeep/internal/audit"
	"github.com/yourusername/gatekeep/internal/config"
	"github.com/yourusername/gatekeep/internal/diff"
	"github.com/yourusername/gatekeep/internal/snowflake"
)

// Orchestrator coordinates the entire sync workflow
type Orchestrator struct {
	configParser *config.Parser
	stateReader  *snowflake.StateReader
	executor     *snowflake.Executor
	syncMode     diff.SyncMode
	auditLogger  audit.AuditLogger // Optional - can be nil or NoOpLogger
}

// NewOrchestrator creates a new sync orchestrator
func NewOrchestrator(
	configParser *config.Parser,
	stateReader *snowflake.StateReader,
	executor *snowflake.Executor,
	syncMode diff.SyncMode,
) *Orchestrator {
	return &Orchestrator{
		configParser: configParser,
		stateReader:  stateReader,
		executor:     executor,
		syncMode:     syncMode,
		auditLogger:  audit.NewNoOpLogger(), // Default to no-op
	}
}

// WithAuditLogger sets an audit logger (optional)
func (o *Orchestrator) WithAuditLogger(logger audit.AuditLogger) *Orchestrator {
	o.auditLogger = logger
	return o
}

// Sync performs a complete sync operation
func (o *Orchestrator) Sync(ctx context.Context, configPath string, syncConfig Config) (*Result, error) {
	startTime := time.Now()

	result := &Result{
		SyncID:    generateSyncID(),
		Status:    StatusRunning,
		StartedAt: startTime,
	}

	// Start audit logging (if configured)
	var syncRunID int64
	if o.auditLogger != nil && syncConfig.Mode != ModeDryRun {
		configContent, err := os.ReadFile(configPath) // nolint:gosec // configPath is user-provided config file
		if err == nil {
			syncRunID, _ = o.auditLogger.StartSync(ctx, configPath, configContent, nil) // nolint:errcheck // audit errors are non-fatal
			if syncRunID > 0 {
				_ = o.auditLogger.SetSyncRunning(ctx, syncRunID) // nolint:errcheck // audit errors are non-fatal
			}
		}
	}

	// Step 1: Parse and validate YAML configuration
	cfg, err := o.configParser.ParseFile(configPath)
	if err != nil {
		result.Status = StatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to parse config: %v", err)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(startTime).Milliseconds()

		// Log failure to audit
		if syncRunID > 0 && o.auditLogger != nil {
			_ = o.auditLogger.FailSync(ctx, syncRunID, result.ErrorMessage, startTime) // nolint:errcheck // audit errors are non-fatal
		}

		return result, err
	}

	validator := config.NewValidator()
	if validationErr := validator.Validate(cfg); validationErr != nil {
		result.Status = StatusFailed
		result.ErrorMessage = fmt.Sprintf("config validation failed: %v", validationErr)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(startTime).Milliseconds()

		// Log failure to audit
		if syncRunID > 0 && o.auditLogger != nil {
			_ = o.auditLogger.FailSync(ctx, syncRunID, result.ErrorMessage, startTime) // nolint:errcheck // audit errors are non-fatal
		}

		return result, validationErr
	}

	// Step 2: Read current Snowflake state
	actualState, err := o.stateReader.ReadState()
	if err != nil {
		result.Status = StatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to read Snowflake state: %v", err)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(startTime).Milliseconds()

		// Log failure to audit
		if syncRunID > 0 && o.auditLogger != nil {
			_ = o.auditLogger.FailSync(ctx, syncRunID, result.ErrorMessage, startTime) // nolint:errcheck // audit errors are non-fatal
		}

		return result, err
	}

	// Step 3: Compute diff
	differ := diff.NewDiffer(o.syncMode)
	diffResult, err := differ.ComputeDiff(diff.Input{
		DesiredConfig: *cfg,
		ActualState:   *actualState,
		Mode:          o.syncMode,
	})
	if err != nil {
		result.Status = StatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to compute diff: %v", err)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(startTime).Milliseconds()

		// Log failure to audit
		if syncRunID > 0 && o.auditLogger != nil {
			_ = o.auditLogger.FailSync(ctx, syncRunID, result.ErrorMessage, startTime) // nolint:errcheck // audit errors are non-fatal
		}

		return result, err
	}

	// Step 4: Generate SQL plan
	planner := diff.NewPlanner(diffResult)
	operations, err := planner.GeneratePlan()
	if err != nil {
		result.Status = StatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to generate plan: %v", err)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(startTime).Milliseconds()

		// Log failure to audit
		if syncRunID > 0 && o.auditLogger != nil {
			_ = o.auditLogger.FailSync(ctx, syncRunID, result.ErrorMessage, startTime) // nolint:errcheck // audit errors are non-fatal
		}

		return result, err
	}

	result.OperationsTotal = len(operations)

	// If no operations, return success
	if len(operations) == 0 {
		result.Status = StatusSuccess
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(startTime).Milliseconds()

		// Log completion to audit
		if syncRunID > 0 && o.auditLogger != nil {
			_ = o.auditLogger.CompleteSync(ctx, syncRunID, 0, 0, 0, startTime) // nolint:errcheck // audit errors are non-fatal
		}

		return result, nil
	}

	// Step 5: Execute operations (or dry-run)
	var opResults []OperationResult
	if syncConfig.Mode == ModeDryRun {
		opResults = o.dryRun(operations)
	} else {
		opResults = o.executeWithAudit(ctx, operations, syncConfig, syncRunID)
	}

	result.Operations = opResults

	// Step 6: Compute final status
	successCount := 0
	failedCount := 0
	for _, opResult := range opResults {
		if opResult.Status == OpStatusSuccess {
			successCount++
		} else if opResult.Status == OpStatusFailed {
			failedCount++
		}
	}

	result.OperationsSuccess = successCount
	result.OperationsFailed = failedCount

	if failedCount == 0 {
		result.Status = StatusSuccess
	} else if successCount > 0 {
		result.Status = StatusPartial
		result.ErrorMessage = fmt.Sprintf("%d operations failed", failedCount)
	} else {
		result.Status = StatusFailed
		result.ErrorMessage = "all operations failed"
	}

	result.CompletedAt = time.Now()
	result.DurationMs = time.Since(startTime).Milliseconds()

	// Log completion to audit
	if syncRunID > 0 && o.auditLogger != nil {
		_ = o.auditLogger.CompleteSync(ctx, syncRunID, len(operations), successCount, failedCount, startTime) // nolint:errcheck // audit errors are non-fatal
	}

	return result, nil
}

// execute runs operations with parallel execution
func (o *Orchestrator) execute(ctx context.Context, operations []diff.SQLOperation, syncConfig Config) []OperationResult {
	parallelExecutor := NewParallelExecutor(o.executor, syncConfig)
	return parallelExecutor.Execute(ctx, operations)
}

// executeWithAudit runs operations with audit logging
func (o *Orchestrator) executeWithAudit(ctx context.Context, operations []diff.SQLOperation, syncConfig Config, syncRunID int64) []OperationResult {
	// If no audit logger, just execute normally
	if o.auditLogger == nil || syncRunID == 0 {
		return o.execute(ctx, operations, syncConfig)
	}

	// Log all operations to audit first
	opIDs := make(map[int]int64) // map operation index to audit operation ID
	for i, op := range operations {
		opID, err := o.auditLogger.LogOperation(ctx, syncRunID, string(op.Type), op.Target, op.SQL)
		if err == nil {
			opIDs[i] = opID
		}
	}

	// Execute operations
	results := o.execute(ctx, operations, syncConfig)

	// Update audit log with results
	for i, result := range results {
		if opID, exists := opIDs[i]; exists {
			executionTimeMs := int(result.ExecutionTime.Milliseconds())
			if result.Status == OpStatusSuccess {
				_ = o.auditLogger.RecordOperationSuccess(ctx, opID, executionTimeMs) // nolint:errcheck // audit errors are non-fatal
			} else if result.Status == OpStatusFailed {
				_ = o.auditLogger.RecordOperationFailure(ctx, opID, result.Error, executionTimeMs) // nolint:errcheck // audit errors are non-fatal
			}
		}
	}

	return results
}

// dryRun simulates execution without running SQL
func (o *Orchestrator) dryRun(operations []diff.SQLOperation) []OperationResult {
	results := make([]OperationResult, len(operations))

	for i, op := range operations {
		results[i] = OperationResult{
			Operation:     op,
			Status:        OpStatusSuccess,
			ExecutionTime: 0,
			ExecutedAt:    time.Now(),
		}
	}

	return results
}

// DryRun performs a dry-run sync (no execution)
func (o *Orchestrator) DryRun(ctx context.Context, configPath string) (*Result, error) {
	return o.Sync(ctx, configPath, DryRunConfig())
}

// generateSyncID generates a unique sync ID
func generateSyncID() string {
	return fmt.Sprintf("sync-%d", time.Now().Unix())
}
