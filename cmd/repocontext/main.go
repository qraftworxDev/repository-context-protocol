package main

import (
	"fmt"
	"os"

	"repository-context-protocol/internal/cli"
)

// Main CLI binary for repocontext tool
func main() {
	// Create the root command
	rootCmd := cli.NewRootCommand()

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		// Print error to stderr
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		// Exit with error status
		os.Exit(1)
	}
}
