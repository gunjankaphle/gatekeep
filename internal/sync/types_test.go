package sync

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Mode != ModeExecute {
		t.Errorf("expected mode %s, got %s", ModeExecute, config.Mode)
	}

	if config.Workers != 10 {
		t.Errorf("expected 10 workers, got %d", config.Workers)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", config.Timeout)
	}

	if config.FailureThreshold != 0.20 {
		t.Errorf("expected 0.20 failure threshold, got %f", config.FailureThreshold)
	}

	if !config.ContinueOnError {
		t.Error("expected ContinueOnError to be true")
	}

	if !config.CircuitBreakerEnabled {
		t.Error("expected CircuitBreakerEnabled to be true")
	}
}

func TestDryRunConfig(t *testing.T) {
	config := DryRunConfig()

	if config.Mode != ModeDryRun {
		t.Errorf("expected mode %s, got %s", ModeDryRun, config.Mode)
	}

	// Other settings should match DefaultConfig
	if config.Workers != 10 {
		t.Errorf("expected 10 workers, got %d", config.Workers)
	}
}

func TestResult_StatusProgression(t *testing.T) {
	result := &Result{
		Status: StatusPending,
	}

	if result.Status != StatusPending {
		t.Errorf("expected pending status, got %s", result.Status)
	}

	result.Status = StatusRunning
	if result.Status != StatusRunning {
		t.Errorf("expected running status, got %s", result.Status)
	}

	result.Status = StatusSuccess
	if result.Status != StatusSuccess {
		t.Errorf("expected success status, got %s", result.Status)
	}
}

func TestOperationStatus_Values(t *testing.T) {
	statuses := []OperationStatus{
		OpStatusPending,
		OpStatusRunning,
		OpStatusSuccess,
		OpStatusFailed,
		OpStatusSkipped,
	}

	expected := []string{
		"pending",
		"running",
		"success",
		"failed",
		"skipped",
	}

	for i, status := range statuses {
		if string(status) != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], status)
		}
	}
}
