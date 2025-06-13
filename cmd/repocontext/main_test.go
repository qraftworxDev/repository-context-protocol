package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"repository-context-protocol/internal/cli"
)

func TestCreateRootCommand(t *testing.T) {
	// Test that we can create the root command successfully
	rootCmd := cli.NewRootCommand()

	if rootCmd == nil {
		t.Fatal("Expected root command to be created, got nil")
	}

	if rootCmd.Use != "repocontext" {
		t.Errorf("Expected root command Use to be 'repocontext', got '%s'", rootCmd.Use)
	}

	// Test that subcommands are properly registered
	subcommands := rootCmd.Commands()
	if len(subcommands) == 0 {
		t.Fatal("Expected root command to have subcommands")
	}

	// Check for expected subcommands
	expectedCommands := []string{"init", "build", "query"}
	foundCommands := make(map[string]bool)

	for _, cmd := range subcommands {
		foundCommands[cmd.Use] = true
	}

	for _, expected := range expectedCommands {
		if !foundCommands[expected] {
			t.Errorf("Expected subcommand '%s' to be registered", expected)
		}
	}
}

func TestMainExecuteWithArgs(t *testing.T) {
	// Test that main can handle arguments (unit test)
	// We'll test the execute function directly

	rootCmd := cli.NewRootCommand()

	// Test version flag
	rootCmd.SetArgs([]string{"--version"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Expected --version to execute successfully, got error: %v", err)
	}
}

func TestMainExecuteHelp(t *testing.T) {
	// Test help execution
	rootCmd := cli.NewRootCommand()

	// Test help (no args should show help)
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Expected help to execute successfully, got error: %v", err)
	}
}

// Integration tests (require binary to be built)

func getBinaryPath() string {
	// Try to find the binary in common locations
	possiblePaths := []string{
		"./bin/repocontext",
		"../../bin/repocontext",
		"../../../bin/repocontext",
	}

	for _, path := range possiblePaths {
		if absPath, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				return absPath
			}
		}
	}

	// If not found, assume it's in PATH
	return "repocontext"
}

func TestMain_Version(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test")
	}

	binaryPath := getBinaryPath()

	// Run the binary with --version flag
	cmd := exec.Command(binaryPath, "--version")

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run command with --version: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "repocontext version") {
		t.Errorf("Expected version output to contain 'repocontext version', got: %s", outputStr)
	}
}

func TestMain_Help(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test")
	}

	binaryPath := getBinaryPath()

	// Run the binary with --help flag
	cmd := exec.Command(binaryPath, "--help")

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run command with --help: %v", err)
	}

	outputStr := string(output)
	expectedStrings := []string{
		"Repository Context Protocol CLI",
		"Available Commands:",
		"init",
		"build",
		"query",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected help output to contain '%s', got: %s", expected, outputStr)
		}
	}
}

func TestMain_InvalidCommand(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test")
	}

	binaryPath := getBinaryPath()

	// Run the binary with an invalid command
	cmd := exec.Command(binaryPath, "invalid-command")

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected command to fail with invalid command")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "unknown command") {
		t.Errorf("Expected error output to contain 'unknown command', got: %s", outputStr)
	}
}

func TestMain_NoArgs(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test")
	}

	binaryPath := getBinaryPath()

	// Run the binary with no arguments (should show help)
	cmd := exec.Command(binaryPath)

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run command with no args: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Repository Context Protocol CLI") {
		t.Errorf("Expected no-args output to show help, got: %s", outputStr)
	}
}

func TestMain_SubcommandIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test")
	}

	binaryPath := getBinaryPath()

	// Test that subcommands are accessible
	subcommands := []string{"init", "build", "query"}

	for _, subcmd := range subcommands {
		t.Run(subcmd, func(t *testing.T) {
			cmd := exec.Command(binaryPath, subcmd, "--help")

			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("Failed to run %s --help: %v", subcmd, err)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, subcmd) {
				t.Errorf("Expected %s help to contain command name, got: %s", subcmd, outputStr)
			}
		})
	}
}

func TestMainFunctionExists(t *testing.T) {
	// This test ensures that we can call the main function without panicking
	// Since main() doesn't return anything, we test by ensuring it can execute
	// We'll capture this by testing the application doesn't crash when called directly

	// This is a basic test to ensure main function can be called
	// The real test will be when we build and run the binary
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("main() function panicked: %v", r)
		}
	}()

	// We can't directly test main() since it calls os.Exit
	// But we can test the main logic by testing the root command execution
	rootCmd := cli.NewRootCommand()
	if rootCmd == nil {
		t.Fatal("Root command should be created in main function")
	}
}
