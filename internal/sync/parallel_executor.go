package sync

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yourusername/gatekeep/internal/diff"
	"github.com/yourusername/gatekeep/internal/snowflake"
)

// ParallelExecutor executes SQL operations in parallel using a worker pool
type ParallelExecutor struct {
	executor *snowflake.Executor
	config   Config
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor(executor *snowflake.Executor, config Config) *ParallelExecutor {
	return &ParallelExecutor{
		executor: executor,
		config:   config,
	}
}

// Execute runs all operations in parallel with the configured number of workers
func (pe *ParallelExecutor) Execute(ctx context.Context, operations []diff.SQLOperation) []OperationResult {
	if len(operations) == 0 {
		return []OperationResult{}
	}

	// Create results slice
	results := make([]OperationResult, len(operations))

	// Create job queue
	jobs := make(chan jobItem, len(operations))
	resultsChan := make(chan jobResult, len(operations))

	// Track execution statistics for circuit breaker
	var totalExecuted, totalFailed atomic.Int64

	// Create context with timeout
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start worker pool
	var wg sync.WaitGroup
	numWorkers := pe.config.Workers
	if numWorkers <= 0 {
		numWorkers = 10 // Default
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go pe.worker(execCtx, &wg, jobs, resultsChan, &totalExecuted, &totalFailed, cancel)
	}

	// Send jobs to queue
	go func() {
		for i, op := range operations {
			jobs <- jobItem{
				index:     i,
				operation: op,
			}
		}
		close(jobs)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Process results
	for result := range resultsChan {
		results[result.index] = result.result

		// Call progress callback if provided
		if pe.config.ProgressCallback != nil {
			pe.config.ProgressCallback(result.result)
		}
	}

	return results
}

// worker processes jobs from the queue
func (pe *ParallelExecutor) worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	jobs <-chan jobItem,
	results chan<- jobResult,
	totalExecuted *atomic.Int64,
	totalFailed *atomic.Int64,
	cancel context.CancelFunc,
) {
	defer wg.Done()

	for job := range jobs {
		// Check if context is cancelled (circuit breaker triggered)
		select {
		case <-ctx.Done():
			// Mark remaining jobs as skipped
			results <- jobResult{
				index: job.index,
				result: OperationResult{
					Operation: job.operation,
					Status:    OpStatusSkipped,
					Error:     "execution stopped by circuit breaker",
				},
			}
			continue
		default:
		}

		// Execute the operation
		result := pe.executeOperation(ctx, job.operation)

		// Update counters
		executed := totalExecuted.Add(1)
		if result.Status == OpStatusFailed {
			failed := totalFailed.Add(1)

			// Check circuit breaker
			if pe.config.CircuitBreakerEnabled {
				failureRate := float64(failed) / float64(executed)
				if failureRate > pe.config.FailureThreshold && executed >= 10 {
					// Trigger circuit breaker
					fmt.Printf("Circuit breaker triggered: %.1f%% failure rate (threshold: %.1f%%)\n",
						failureRate*100, pe.config.FailureThreshold*100)
					cancel() // Cancel context to stop all workers
				}
			}
		}

		// Send result
		results <- jobResult{
			index:  job.index,
			result: result,
		}
	}
}

// executeOperation executes a single SQL operation
func (pe *ParallelExecutor) executeOperation(ctx context.Context, op diff.SQLOperation) OperationResult {
	result := OperationResult{
		Operation:  op,
		Status:     OpStatusRunning,
		ExecutedAt: time.Now(),
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, pe.config.Timeout)
	defer cancel()

	// Execute SQL
	execResult := pe.executor.Execute(execCtx, op.SQL)

	result.ExecutionTime = execResult.ExecutionTime

	if execResult.Success {
		result.Status = OpStatusSuccess
	} else {
		result.Status = OpStatusFailed
		if execResult.Error != nil {
			result.Error = execResult.Error.Error()
		}
	}

	return result
}

// jobItem represents a job in the queue
type jobItem struct {
	index     int
	operation diff.SQLOperation
}

// jobResult represents the result of a job
type jobResult struct {
	index  int
	result OperationResult
}
