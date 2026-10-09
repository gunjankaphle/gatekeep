package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/gatekeep/internal/diff"
	gatesync "github.com/yourusername/gatekeep/internal/sync"
)

func testConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("version: 1.0\nroles:\n  - name: TEST_ROLE\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SYNC_MODE", "")
	t.Setenv("SYNC_WORKERS", "")
	t.Setenv("SYNC_TIMEOUT", "")
	return path
}
func TestSyncDryRunJSON(t *testing.T) {
	path := testConfig(t)
	var out, errOut bytes.Buffer
	runner := func(_ context.Context, options syncOptions) (*gatesync.Result, error) {
		if !options.dryRun || options.execution.Mode != gatesync.ModeDryRun || options.mode != diff.SyncModeAdditive || options.execution.Workers != 3 {
			t.Fatalf("options: %#v", options)
		}
		return &gatesync.Result{Status: gatesync.StatusSuccess, OperationsTotal: 1, OperationsSuccess: 1, Operations: []gatesync.OperationResult{{Operation: diff.SQLOperation{Type: diff.OpCreateRole, SQL: `CREATE ROLE "TEST_ROLE"`, Target: "TEST_ROLE"}, Status: gatesync.OpStatusSuccess}}}, nil
	}
	code := run(context.Background(), []string{"sync", "--config", path, "--dry-run", "--format", "json", "--workers", "3"}, &out, &errOut, runner)
	if code != 0 {
		t.Fatalf("%d: %s", code, errOut.String())
	}
	var output cliOutput
	if err := json.Unmarshal(out.Bytes(), &output); err != nil {
		t.Fatal(err)
	}
	if output.Mode != "dry-run" || len(output.Operations) != 1 || output.Operations[0].Status != "planned" || output.Summary["roles_created"] != 1 {
		t.Fatalf("%#v", output)
	}
}
func TestSyncFailsBeforeConnecting(t *testing.T) {
	path := testConfig(t)
	for _, args := range [][]string{
		{"sync"}, {"sync", "--config", path, "--workers", "0"}, {"sync", "--config", path, "--timeout", "bad"},
		{"sync", "--config", path, "--mode", "strict"}, {"sync", "--config", path, "--mode", "unknown"},
		{"sync", "--config", "missing.yaml"}, {"sync", "--config", path, "extra"},
	} {
		var out, errOut bytes.Buffer
		code := run(context.Background(), args, &out, &errOut, func(context.Context, syncOptions) (*gatesync.Result, error) {
			t.Fatal("unexpected connection attempt")
			return nil, nil
		})
		if code == 0 {
			t.Errorf("success for %v", args)
		}
	}
}
func TestExecutionFailuresReturnNonzero(t *testing.T) {
	path := testConfig(t)
	for _, tc := range []struct {
		result *gatesync.Result
		err    error
	}{
		{nil, errors.New("connection failed")},
		{&gatesync.Result{Status: gatesync.StatusPartial, OperationsFailed: 1, ErrorMessage: "operation failed"}, nil},
		{&gatesync.Result{Status: gatesync.StatusSuccess, Operations: []gatesync.OperationResult{{Status: gatesync.OpStatusSkipped}}}, nil},
		{nil, nil},
	} {
		var out, errOut bytes.Buffer
		code := run(context.Background(), []string{"sync", "--config", path, "--format", "json"}, &out, &errOut, func(context.Context, syncOptions) (*gatesync.Result, error) { return tc.result, tc.err })
		if code != 1 {
			t.Fatalf("exit %d", code)
		}
		var output cliOutput
		if err := json.Unmarshal(out.Bytes(), &output); err != nil {
			t.Fatal(err)
		}
		if output.Status == "success" {
			t.Fatalf("failed execution reported success: %s", out.String())
		}
	}
}
func TestZeroChangeExecution(t *testing.T) {
	path := testConfig(t)
	var out, errOut bytes.Buffer
	code := run(context.Background(), []string{"sync", "--config", path, "--format", "json"}, &out, &errOut, func(_ context.Context, options syncOptions) (*gatesync.Result, error) {
		if options.execution.Mode != gatesync.ModeExecute {
			t.Fatal("execution not selected")
		}
		return &gatesync.Result{Status: gatesync.StatusSuccess}, nil
	})
	if code != 0 {
		t.Fatal(errOut.String())
	}
	var output cliOutput
	if err := json.Unmarshal(out.Bytes(), &output); err != nil {
		t.Fatal(err)
	}
	if output.Operations == nil || len(output.Operations) != 0 {
		t.Fatalf("expected empty JSON array: %s", out.String())
	}
}
func TestUnknownCommand(t *testing.T) {
	var out, errOut bytes.Buffer
	if run(context.Background(), []string{"unknown"}, &out, &errOut, nil) != 2 {
		t.Fatal("unknown command succeeded")
	}
}
