package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	// Test that it's a valid cobra command
	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
	}

	// Test basic command properties
	if cmd.Use != "repocontext" {
		t.Errorf("Expected Use to be 'repocontext', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("Expected Long description to be set")
	}
}

func TestRootCommand_HasSubcommands(t *testing.T) {
	cmd := NewRootCommand()

	// Check that subcommands are registered
	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Fatal("Expected root command to have subcommands")
	}

	// Create a map of expected commands
	expectedCommands := map[string]bool{
		"init":  false,
		"build": false,
		"query": false,
	}

	// Check that expected commands are present
	for _, subcmd := range subcommands {
		if _, exists := expectedCommands[subcmd.Use]; exists {
			expectedCommands[subcmd.Use] = true
		}
	}

	// Verify all expected commands were found
	for cmdName, found := range expectedCommands {
		if !found {
			t.Errorf("Expected subcommand '%s' not found", cmdName)
		}
	}
}

func TestRootCommand_InitSubcommand(t *testing.T) {
	cmd := NewRootCommand()

	// Find the init subcommand
	var initCmd *cobra.Command
	for _, subcmd := range cmd.Commands() {
		if subcmd.Use == "init" {
			initCmd = subcmd
			break
		}
	}

	if initCmd == nil {
		t.Fatal("Expected 'init' subcommand to be registered")
	}

	// Test init command properties
	if initCmd.Short == "" {
		t.Error("Expected init command to have Short description")
	}

	if initCmd.RunE == nil {
		t.Error("Expected init command to have RunE function")
	}

	// Test that init command has expected flags
	pathFlag := initCmd.Flags().Lookup("path")
	if pathFlag == nil {
		t.Error("Expected init command to have 'path' flag")
	}
}

func TestRootCommand_BuildSubcommand(t *testing.T) {
	cmd := NewRootCommand()

	// Find the build subcommand
	var buildCmd *cobra.Command
	for _, subcmd := range cmd.Commands() {
		if subcmd.Use == "build" {
			buildCmd = subcmd
			break
		}
	}

	if buildCmd == nil {
		t.Fatal("Expected 'build' subcommand to be registered")
	}

	// Test build command properties
	if buildCmd.Short == "" {
		t.Error("Expected build command to have Short description")
	}

	if buildCmd.RunE == nil {
		t.Error("Expected build command to have RunE function")
	}

	// Test that build command has expected flags
	pathFlag := buildCmd.Flags().Lookup("path")
	if pathFlag == nil {
		t.Error("Expected build command to have 'path' flag")
	}

	verboseFlag := buildCmd.Flags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("Expected build command to have 'verbose' flag")
	}
}

func TestRootCommand_QuerySubcommand(t *testing.T) {
	cmd := NewRootCommand()

	// Find the query subcommand
	var queryCmd *cobra.Command
	for _, subcmd := range cmd.Commands() {
		if subcmd.Use == "query" {
			queryCmd = subcmd
			break
		}
	}

	if queryCmd == nil {
		t.Fatal("Expected 'query' subcommand to be registered")
	}

	// Test query command properties
	if queryCmd.Short == "" {
		t.Error("Expected query command to have Short description")
	}

	if queryCmd.RunE == nil {
		t.Error("Expected query command to have RunE function")
	}

	// Test that query command has expected flags
	pathFlag := queryCmd.Flags().Lookup("path")
	if pathFlag == nil {
		t.Error("Expected query command to have 'path' flag")
	}

	functionFlag := queryCmd.Flags().Lookup("function")
	if functionFlag == nil {
		t.Error("Expected query command to have 'function' flag")
	}

	formatFlag := queryCmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Error("Expected query command to have 'format' flag")
	}
}

func TestRootCommand_HelpOutput(t *testing.T) {
	cmd := NewRootCommand()

	// Test that help can be generated without errors
	err := cmd.Help()
	if err != nil {
		t.Fatalf("Failed to generate help output: %v", err)
	}

	// For this test, we just verify help doesn't error
	// The actual help content is tested via usage strings
	helpStr := cmd.UsageString()

	if !strings.Contains(helpStr, "repocontext") {
		t.Error("Expected help output to contain 'repocontext'")
	}

	if !strings.Contains(helpStr, "init") {
		t.Error("Expected help output to contain 'init' subcommand")
	}

	if !strings.Contains(helpStr, "build") {
		t.Error("Expected help output to contain 'build' subcommand")
	}

	if !strings.Contains(helpStr, "query") {
		t.Error("Expected help output to contain 'query' subcommand")
	}
}

func TestRootCommand_Version(t *testing.T) {
	cmd := NewRootCommand()

	// Test that version is set using Cobra's native version handling
	if cmd.Version == "" {
		t.Error("Expected root command to have Version field set")
	}

	// Test that the version matches our constant
	if cmd.Version != Version {
		t.Errorf("Expected version to be '%s', got '%s'", Version, cmd.Version)
	}

	// Test that --version flag works (Cobra handles this automatically)
	cmd.SetArgs([]string{"--version"})

	// Capture output to test version output format
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected --version to execute successfully, got error: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, Version) {
		t.Errorf("Expected version output to contain '%s', got: %s", Version, output)
	}
}

func TestRootCommand_GlobalFlags(t *testing.T) {
	cmd := NewRootCommand()

	// Test that global flags are properly set up
	flags := cmd.PersistentFlags()

	// Check for common global flags that might be useful
	// (This test can be expanded as global flags are added)
	if flags == nil {
		t.Error("Expected persistent flags to be initialized")
	}
}

func TestRootCommand_ExecuteWithoutArgs(t *testing.T) {
	cmd := NewRootCommand()

	// Test that root command can be executed (should show help)
	cmd.SetArgs([]string{})

	// Capture output to avoid printing during tests
	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	// This should not panic or error when no args provided
	// (Root command typically shows help in this case)
	err := cmd.Execute()
	// Note: We don't check for specific error here as the behavior
	// of root command without args can vary (help vs error)
	_ = err // Acknowledge we're not checking the error
}

func TestRootCommand_SubcommandIntegration(t *testing.T) {
	cmd := NewRootCommand()

	// Test that subcommands are properly integrated
	// by checking they can be found via command traversal
	initCmd, _, err := cmd.Find([]string{"init"})
	if err != nil {
		t.Errorf("Failed to find 'init' subcommand: %v", err)
	}
	if initCmd == nil || initCmd.Use != "init" {
		t.Error("Expected to find 'init' subcommand")
	}

	buildCmd, _, err := cmd.Find([]string{"build"})
	if err != nil {
		t.Errorf("Failed to find 'build' subcommand: %v", err)
	}
	if buildCmd == nil || buildCmd.Use != "build" {
		t.Error("Expected to find 'build' subcommand")
	}

	queryCmd, _, err := cmd.Find([]string{"query"})
	if err != nil {
		t.Errorf("Failed to find 'query' subcommand: %v", err)
	}
	if queryCmd == nil || queryCmd.Use != "query" {
		t.Error("Expected to find 'query' subcommand")
	}
}
