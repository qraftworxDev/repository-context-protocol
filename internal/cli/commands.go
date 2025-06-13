package cli

import (
	"github.com/spf13/cobra"
)

// Version information for the CLI
const Version = "1.0.0"

// NewRootCommand creates the root command for the repocontext CLI
func NewRootCommand() *cobra.Command {
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
			// If no subcommand is provided, show help
			return cmd.Help()
		},
	}

	// Set version using Cobra's native version handling
	rootCmd.Version = Version

	// Customize version output template
	rootCmd.SetVersionTemplate("repocontext version {{.Version}}\n")

	// Add subcommands
	rootCmd.AddCommand(NewInitCommand())
	rootCmd.AddCommand(NewBuildCommand())
	rootCmd.AddCommand(NewQueryCommand())

	return rootCmd
}
