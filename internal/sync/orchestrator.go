package sync

import (
	"context"
	"fmt"
	"time"

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
	}
}

// Sync performs a complete sync operation
func (o *Orchestrator) Sync(ctx context.Context, configPath string, syncConfig Config) (*Result, error) {
	startTime := time.Now()

	result := &Result{
		SyncID:    generateSyncID(),
		Status:    StatusRunning,
		StartedAt: startTime,
	}

	// Step 1: Parse and validate YAML configuration
	cfg, err := o.configParser.ParseFile(configPath)
	if err != nil {
		result.Status = StatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to parse config: %v", err)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(startTime).Milliseconds()
		return result, err
	}

	validator := config.NewValidator()
	if validationErr := validator.Validate(cfg); validationErr != nil {
		result.Status = StatusFailed
		result.ErrorMessage = fmt.Sprintf("config validation failed: %v", validationErr)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(startTime).Milliseconds()
		return result, validationErr
	}

	// Step 2: Read current Snowflake state
	actualState, err := o.stateReader.ReadState()
	if err != nil {
		result.Status = StatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to read Snowflake state: %v", err)
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(startTime).Milliseconds()
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
		return result, err
	}

	result.OperationsTotal = len(operations)

	// If no operations, return success
	if len(operations) == 0 {
		result.Status = StatusSuccess
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(startTime).Milliseconds()
		return result, nil
	}

	// Step 5: Execute operations (or dry-run)
	var opResults []OperationResult
	if syncConfig.Mode == ModeDryRun {
		opResults = o.dryRun(operations)
	} else {
		opResults = o.execute(ctx, operations, syncConfig)
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

	return result, nil
}

// execute runs operations with parallel execution
func (o *Orchestrator) execute(ctx context.Context, operations []diff.SQLOperation, syncConfig Config) []OperationResult {
	parallelExecutor := NewParallelExecutor(o.executor, syncConfig)
	return parallelExecutor.Execute(ctx, operations)
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
