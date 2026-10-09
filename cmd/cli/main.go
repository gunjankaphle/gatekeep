package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/yourusername/gatekeep/internal/config"
	"github.com/yourusername/gatekeep/internal/database"
	"github.com/yourusername/gatekeep/internal/diff"
	"github.com/yourusername/gatekeep/internal/snowflake"
	gatesync "github.com/yourusername/gatekeep/internal/sync"
)

type syncOptions struct {
	configPath string
	dryRun     bool
	mode       diff.SyncMode
	execution  gatesync.Config
}
type syncRunner func(context.Context, syncOptions) (*gatesync.Result, error)

type cliOperation struct {
	Type   string `json:"type"`
	SQL    string `json:"sql"`
	Target string `json:"target"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
type cliOutput struct {
	Status     string         `json:"status"`
	Mode       string         `json:"mode"`
	Config     string         `json:"config"`
	Message    string         `json:"message,omitempty"`
	Operations []cliOperation `json:"operations"`
	Summary    map[string]int `json:"summary"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr, executeSync))
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer, runner syncRunner) int {
	if len(args) == 0 {
		usage(stdout)
		return 0
	}
	switch args[0] {
	case "help", "--help", "-h":
		usage(stdout)
		return 0
	case "version":
		if _, err := fmt.Fprintln(stdout, "gatekeep version 0.1.0-alpha"); err != nil {
			return 1
		}
		return 0
	case "validate":
		flags := flag.NewFlagSet("validate", flag.ContinueOnError)
		flags.SetOutput(stderr)
		if err := flags.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			return 2
		}
		if flags.NArg() != 1 {
			diagnostic(stderr, "Usage: gatekeep validate <config.yaml>\n")
			return 2
		}
		if _, err := config.NewParser().ParseFile(flags.Arg(0)); err != nil {
			diagnostic(stderr, "Validation failed: %v\n", err)
			return 1
		}
		if _, err := fmt.Fprintf(stdout, "Configuration valid: %s\n", flags.Arg(0)); err != nil {
			return 1
		}
		return 0
	case "sync":
		return runSync(ctx, args[1:], stdout, stderr, runner)
	default:
		diagnostic(stderr, "Unknown command: %s\n", args[0])
		usage(stderr)
		return 2
	}
}

func runSync(ctx context.Context, args []string, stdout, stderr io.Writer, runner syncRunner) int {
	flags := flag.NewFlagSet("sync", flag.ContinueOnError)
	flags.SetOutput(stderr)
	path := flags.String("config", "", "Path to YAML configuration (required)")
	dry := flags.Bool("dry-run", false, "Preview SQL without executing or writing audit records")
	format := flags.String("format", "text", "Output format: text or json")
	mode := flags.String("mode", envDefault("SYNC_MODE", "additive"), "Reconciliation mode: additive or strict (strict is dry-run only)")
	workers := flags.String("workers", envDefault("SYNC_WORKERS", "10"), "Parallel worker count")
	timeout := flags.String("timeout", envDefault("SYNC_TIMEOUT", "30s"), "Timeout per executed operation")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	output := cliOutput{Status: "failed", Mode: "execute", Config: *path, Operations: []cliOperation{}, Summary: map[string]int{"roles_created": 0, "grants_added": 0, "grants_revoked": 0, "operations_total": 0, "operations_success": 0, "operations_failed": 0}}
	if *dry {
		output.Mode = "dry-run"
	}
	finish := func(message string, code int) int {
		output.Message = message
		if *format == "json" {
			if err := json.NewEncoder(stdout).Encode(output); err != nil {
				diagnostic(stderr, "Output failed: %v\n", err)
				return 1
			}
		} else {
			for _, op := range output.Operations {
				if _, err := fmt.Fprintf(stdout, "[%s] %s\n", op.Status, op.SQL); err != nil {
					return 1
				}
				if op.Error != "" {
					diagnostic(stderr, "%s\n", op.Error)
				}
			}
			if code != 0 {
				diagnostic(stderr, "%s\n", message)
			} else {
				if _, err := fmt.Fprintf(stdout, "%s: %d operations (%s)\n", output.Status, output.Summary["operations_total"], output.Mode); err != nil {
					return 1
				}
			}
		}
		return code
	}
	if *format != "text" && *format != "json" {
		diagnostic(stderr, "format must be text or json\n")
		return 2
	}
	if *path == "" || flags.NArg() != 0 {
		return finish("--config is required; unexpected positional arguments are not allowed", 2)
	}
	if *mode != "additive" && *mode != "strict" {
		return finish("mode must be additive or strict", 2)
	}
	if *mode == "strict" && !*dry {
		return finish("strict execution is unavailable until managed scope and revocation are complete; use additive mode or --dry-run", 2)
	}
	count, err := strconv.Atoi(*workers)
	if err != nil || count < 1 || count > 100 {
		return finish("workers must be an integer from 1 to 100", 2)
	}
	duration, err := time.ParseDuration(*timeout)
	if err != nil || duration <= 0 {
		return finish("timeout must be a positive duration", 2)
	}
	// Validate before opening connections or starting audit records.
	if _, validationErr := config.NewParser().ParseFile(*path); validationErr != nil {
		return finish(validationErr.Error(), 1)
	}
	execution := gatesync.DefaultConfig()
	execution.Workers = count
	execution.Timeout = duration
	if *dry {
		execution.Mode = gatesync.ModeDryRun
	}
	result, err := runner(ctx, syncOptions{configPath: *path, dryRun: *dry, mode: diff.SyncMode(*mode), execution: execution})
	if result != nil {
		output.Status = string(result.Status)
		output.Summary["operations_total"] = result.OperationsTotal
		output.Summary["operations_success"] = result.OperationsSuccess
		output.Summary["operations_failed"] = result.OperationsFailed
		for _, op := range result.Operations {
			status := string(op.Status)
			if *dry {
				status = "planned"
			}
			output.Operations = append(output.Operations, cliOperation{Type: string(op.Operation.Type), SQL: op.Operation.SQL, Target: op.Operation.Target, Status: status, Error: op.Error})
			switch op.Operation.Type {
			case diff.OpCreateRole:
				output.Summary["roles_created"]++
			case diff.OpGrantObject, diff.OpGrantRole, diff.OpGrantUserRole:
				output.Summary["grants_added"]++
			case diff.OpRevokeObject, diff.OpRevokeRole, diff.OpRevokeUserRole:
				output.Summary["grants_revoked"]++
			}
		}
	}
	if err != nil {
		output.Status = "failed"
		return finish(err.Error(), 1)
	}
	if result == nil {
		return finish("sync returned no result", 1)
	}
	if result.Status != gatesync.StatusSuccess || result.OperationsFailed > 0 {
		return finish(result.ErrorMessage, 1)
	}
	for _, op := range result.Operations {
		if op.Status != gatesync.OpStatusSuccess {
			output.Status = "partial"
			return finish("some operations did not complete", 1)
		}
	}
	return finish("", 0)
}

func executeSync(ctx context.Context, options syncOptions) (*gatesync.Result, error) {
	cfg, err := snowflake.LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}
	client, err := snowflake.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	defer func() { _ = client.Close() }() //nolint:errcheck // best-effort cleanup
	return executeWithClient(ctx, options, client)
}

func executeWithClient(ctx context.Context, options syncOptions, client snowflake.Client) (*gatesync.Result, error) {
	orchestrator := gatesync.NewOrchestrator(config.NewParser(), snowflake.NewStateReader(client), snowflake.NewExecutor(client), options.mode)
	if !options.dryRun {
		logger, pool, err := database.NewAuditLogger(ctx)
		if err != nil {
			return nil, err
		}
		if pool != nil {
			defer pool.Close()
		}
		orchestrator.WithAuditLogger(logger)
	}
	return orchestrator.Sync(ctx, options.configPath, options.execution)
}
func envDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func usage(out io.Writer) {
	diagnostic(out, "GateKeep - Snowflake Permissions Management\n\nUsage: gatekeep <validate|sync|version> [options]\nRun gatekeep sync --help for options.\n")
}

// diagnostic emits best-effort help/error text; callers already report an exit status.
func diagnostic(out io.Writer, format string, args ...interface{}) {
	_, _ = fmt.Fprintf(out, format, args...) //nolint:errcheck // failure to emit fallback diagnostics cannot be recovered
}
