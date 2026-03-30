package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	// Check if we have any commands
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "sync":
		handleSync()
	case "validate":
		handleValidate()
	case "version":
		fmt.Println("gatekeep version 0.1.0-alpha (Epic 1)")
		os.Exit(0)
	default:
		printUsage()
		os.Exit(0)
	}
}

func handleSync() {
	syncCmd := flag.NewFlagSet("sync", flag.ExitOnError)
	config := syncCmd.String("config", "", "Path to YAML configuration file")
	dryRun := syncCmd.Bool("dry-run", false, "Preview changes without applying")
	format := syncCmd.String("format", "text", "Output format (text or json)")

	if err := syncCmd.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Minimal implementation for CI - real implementation in Epic 5
	if *format == "json" {
		result := map[string]interface{}{
			"status":  "success",
			"message": "GateKeep sync not yet implemented - placeholder for CI",
			"operations": []map[string]string{
				{
					"type":   "PLACEHOLDER",
					"sql":    "-- Real implementation coming in Epic 5",
					"status": "pending",
				},
			},
			"summary": map[string]int{
				"roles_created":  0,
				"grants_added":   0,
				"grants_revoked": 0,
			},
		}

		if *dryRun {
			result["mode"] = "dry-run"
		}

		if *config != "" {
			result["config"] = *config
		}

		if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("GateKeep Sync - Coming soon!")
		fmt.Printf("Config: %s\n", *config)
		fmt.Printf("Dry-run: %v\n", *dryRun)
		fmt.Println("\nReal implementation will be available in Epic 5.")
	}

	os.Exit(0)
}

func handleValidate() {
	validateCmd := flag.NewFlagSet("validate", flag.ExitOnError)
	if err := validateCmd.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("GateKeep Validate - Coming soon!")
	fmt.Println("Real implementation will be available in Epic 2.")
	os.Exit(0)
}

func printUsage() {
	fmt.Println("GateKeep CLI - Snowflake Permissions Management")
	fmt.Println("")
	fmt.Println("Usage: gatekeep <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  validate  Validate YAML configuration")
	fmt.Println("  sync      Sync permissions to Snowflake")
	fmt.Println("  version   Show version information")
	fmt.Println("")
	fmt.Println("Run 'gatekeep <command> --help' for more information on a command.")
	fmt.Println("")
	fmt.Println("Note: This is a placeholder implementation (Epic 1).")
	fmt.Println("Full functionality will be available after Epic 5.")
}
