package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information for the CLI
const Version = "1.0.0"

// NewRootCommand creates the root command for the repocontext CLI
func NewRootCommand() *cobra.Command {
	var showVersion bool

	rootCmd := &cobra.Command{
		Use:   "repocontext",
		Short: "Repository context protocol CLI tool",
		Long: `Repository Context Protocol CLI

A tool for creating semantic indexes of code repositories to provide
rich context for Large Language Models (LLMs). This tool parses source
code, extracts semantic information, and creates searchable indexes
that can be queried for code understanding and generation tasks.

Features:
- Initialize repository context tracking
- Build semantic indexes from source code
- Query code semantics and relationships
- Serve context via HTTP API

Use 'repocontext <command> --help' for more information about a command.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				fmt.Printf("repocontext version %s\n", Version)
				return nil
			}
			// If no subcommand is provided, show help
			return cmd.Help()
		},
	}

	// Add version flag
	rootCmd.Flags().BoolVarP(&showVersion, "version", "V", false, "Show version information")

	// Add subcommands
	rootCmd.AddCommand(NewInitCommand())
	rootCmd.AddCommand(NewBuildCommand())

	return rootCmd
}
