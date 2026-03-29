package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("GateKeep CLI - Coming soon!")
	fmt.Println("Usage: gatekeep <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  validate  Validate YAML configuration")
	fmt.Println("  sync      Sync permissions to Snowflake")
	fmt.Println("  version   Show version information")
	fmt.Println("")
	fmt.Println("Run 'gatekeep <command> --help' for more information on a command.")

	os.Exit(0)
}
