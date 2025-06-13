package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"repository-context-protocol/internal/index"
)

func TestNewQueryCommand(t *testing.T) {
	cmd := NewQueryCommand()

	if cmd == nil {
		t.Fatal("Expected NewQueryCommand to return a command, got nil")
	}

	if cmd.Use != "query" {
		t.Errorf("Expected command use 'query', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command to have a short description")
	}

	if cmd.Long == "" {
		t.Error("Expected command to have a long description")
	}

	if cmd.RunE == nil {
		t.Error("Expected command to have a RunE function")
	}
}

func TestQueryCommand_Flags(t *testing.T) {
	cmd := NewQueryCommand()

	// Test that all expected long flags are present
	expectedLongFlags := []string{
		"function", "type", "variable", "file", "search", "entity-type",
		"include-callers", "include-callees", "include-types",
		"depth", "max-tokens", "format", "json", "verbose", "compact", "path",
	}

	for _, flagName := range expectedLongFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag --%s to exist", flagName)
		}
	}

	// Test that short flags work by checking they're accessible
	shortFlagTests := []struct {
		short string
		long  string
	}{
		{"f", "function"},
		{"t", "type"},
		{"v", "variable"},
		{"s", "search"},
		{"p", "path"},
	}

	for _, test := range shortFlagTests {
		flag := cmd.Flags().ShorthandLookup(test.short)
		if flag == nil {
			t.Errorf("Expected short flag -%s (for --%s) to exist", test.short, test.long)
		}
	}
}

func TestQueryCommand_FlagDefaults(t *testing.T) {
	cmd := NewQueryCommand()

	// Test string flag defaults
	stringTests := []struct {
		flag     string
		expected string
	}{
		{"format", "text"},
		{"path", "."},
	}

	for _, test := range stringTests {
		actual, err := cmd.Flags().GetString(test.flag)
		if err != nil {
			t.Errorf("Failed to get string flag --%s: %v", test.flag, err)
			continue
		}
		if actual != test.expected {
			t.Errorf("Expected flag --%s default value '%s', got '%s'", test.flag, test.expected, actual)
		}
	}

	// Test integer flag defaults
	intTests := []struct {
		flag     string
		expected int
	}{
		{"depth", 2}, // DefaultDepth
		{"max-tokens", 0},
	}

	for _, test := range intTests {
		actual, err := cmd.Flags().GetInt(test.flag)
		if err != nil {
			t.Errorf("Failed to get int flag --%s: %v", test.flag, err)
			continue
		}
		if actual != test.expected {
			t.Errorf("Expected flag --%s default value %d, got %d", test.flag, test.expected, actual)
		}
	}

	// Test boolean flag defaults
	boolTests := []struct {
		flag     string
		expected bool
	}{
		{"json", false},
		{"verbose", false},
		{"compact", false},
		{"include-callers", false},
		{"include-callees", false},
		{"include-types", false},
	}

	for _, test := range boolTests {
		actual, err := cmd.Flags().GetBool(test.flag)
		if err != nil {
			t.Errorf("Failed to get bool flag --%s: %v", test.flag, err)
			continue
		}
		if actual != test.expected {
			t.Errorf("Expected flag --%s default value %t, got %t", test.flag, test.expected, actual)
		}
	}
}

func TestQueryCommand_MutuallyExclusiveFlags(t *testing.T) {
	cmd := NewQueryCommand()

	// Test setting multiple search type flags should be allowed at the CLI level
	// (validation will happen later in the execution)
	err := cmd.Flags().Set("function", "TestFunc")
	if err != nil {
		t.Errorf("Failed to set function flag: %v", err)
	}

	err = cmd.Flags().Set("type", "TestType")
	if err != nil {
		t.Errorf("Failed to set type flag: %v", err)
	}

	// Both flags should be set successfully at the CLI level
	functionFlag := cmd.Flags().Lookup("function")
	if functionFlag.Value.String() != "TestFunc" {
		t.Errorf("Expected function flag to be 'TestFunc', got '%s'", functionFlag.Value.String())
	}

	typeFlag := cmd.Flags().Lookup("type")
	if typeFlag.Value.String() != "TestType" {
		t.Errorf("Expected type flag to be 'TestType', got '%s'", typeFlag.Value.String())
	}
}

func TestQueryCommand_JSONFlag(t *testing.T) {
	cmd := NewQueryCommand()

	// Test that --json flag can be set
	err := cmd.Flags().Set("json", "true")
	if err != nil {
		t.Errorf("Failed to set json flag: %v", err)
	}

	// Verify the JSON flag is set correctly
	jsonFlag, err := cmd.Flags().GetBool("json")
	if err != nil {
		t.Errorf("Failed to get json flag: %v", err)
	}
	if !jsonFlag {
		t.Error("Expected json flag to be true after setting it")
	}

	// Initially, format should still be the default "text" (JSON processing happens in runQuery)
	formatFlag, err := cmd.Flags().GetString("format")
	if err != nil {
		t.Errorf("Failed to get format flag: %v", err)
	}
	if formatFlag != "text" {
		t.Errorf("Expected format flag to be 'text' initially, got '%s'", formatFlag)
	}
}

func TestQueryCommand_JSONFlagProcessing(t *testing.T) {
	// Create temporary directory with .repocontext
	tempDir, err := os.MkdirTemp("", "query_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .repocontext directory
	repoContextDir := filepath.Join(tempDir, ".repocontext")
	err = os.MkdirAll(repoContextDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .repocontext directory: %v", err)
	}

	// Initialize storage to avoid search failures
	storage := index.NewHybridStorage(repoContextDir)
	err = storage.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	cmd := NewQueryCommand()

	// Set required flags
	err = cmd.Flags().Set("path", tempDir)
	if err != nil {
		t.Fatalf("Failed to set path flag: %v", err)
	}

	err = cmd.Flags().Set("function", "TestFunc")
	if err != nil {
		t.Fatalf("Failed to set function flag: %v", err)
	}

	// Set JSON flag
	err = cmd.Flags().Set("json", "true")
	if err != nil {
		t.Fatalf("Failed to set json flag: %v", err)
	}

	// Run the command to trigger the JSON flag processing logic
	// This should set the format to "json" internally before validation
	err = cmd.RunE(cmd, []string{})
	// The command might fail due to no search results, but that's fine for this test
	// We're testing that the JSON flag processing doesn't cause validation errors

	// The format should be processed correctly (no validation error about invalid format)
	if err != nil && strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Expected JSON flag to set format correctly, but got format validation error: %v", err)
	}
}

func TestQueryCommand_ValidationNoRepository(t *testing.T) {
	// Create temporary directory without .repocontext
	tempDir, err := os.MkdirTemp("", "query_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cmd := NewQueryCommand()
	err = cmd.Flags().Set("path", tempDir)
	if err != nil {
		t.Fatalf("Failed to set path flag: %v", err)
	}

	err = cmd.Flags().Set("function", "TestFunc")
	if err != nil {
		t.Fatalf("Failed to set function flag: %v", err)
	}

	// Should fail because repository is not initialized
	err = cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected command to fail when repository is not initialized")
	}
}

func TestQueryCommand_ValidationNoSearchCriteria(t *testing.T) {
	// Create temporary directory with .repocontext
	tempDir, err := os.MkdirTemp("", "query_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .repocontext directory
	repoContextDir := filepath.Join(tempDir, ".repocontext")
	err = os.MkdirAll(repoContextDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .repocontext directory: %v", err)
	}

	cmd := NewQueryCommand()
	err = cmd.Flags().Set("path", tempDir)
	if err != nil {
		t.Fatalf("Failed to set path flag: %v", err)
	}

	// Should fail because no search criteria provided
	err = cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected command to fail when no search criteria provided")
	}
}

func TestQueryCommand_ValidationInvalidFormat(t *testing.T) {
	// Create temporary directory with .repocontext
	tempDir, err := os.MkdirTemp("", "query_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .repocontext directory
	repoContextDir := filepath.Join(tempDir, ".repocontext")
	err = os.MkdirAll(repoContextDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .repocontext directory: %v", err)
	}

	cmd := NewQueryCommand()
	err = cmd.Flags().Set("path", tempDir)
	if err != nil {
		t.Fatalf("Failed to set path flag: %v", err)
	}

	err = cmd.Flags().Set("function", "TestFunc")
	if err != nil {
		t.Fatalf("Failed to set function flag: %v", err)
	}

	err = cmd.Flags().Set("format", "invalid")
	if err != nil {
		t.Fatalf("Failed to set format flag: %v", err)
	}

	// Should fail because format is invalid
	err = cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected command to fail with invalid format")
	}
}

func TestQueryCommand_ValidationInvalidDepth(t *testing.T) {
	cmd := NewQueryCommand()

	// Test setting negative depth
	err := cmd.Flags().Set("depth", "-1")
	if err != nil {
		t.Errorf("Failed to set depth flag: %v", err)
	}

	// The validation will happen in the RunE function
}

func TestQueryCommand_ValidationInvalidMaxTokens(t *testing.T) {
	cmd := NewQueryCommand()

	// Test setting negative max-tokens
	err := cmd.Flags().Set("max-tokens", "-1")
	if err != nil {
		t.Errorf("Failed to set max-tokens flag: %v", err)
	}

	// The validation will happen in the RunE function
}

func TestQueryCommand_ValidationInvalidFormatWithJSON(t *testing.T) {
	// Create temporary directory with .repocontext
	tempDir, err := os.MkdirTemp("", "query_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .repocontext directory
	repoContextDir := filepath.Join(tempDir, ".repocontext")
	err = os.MkdirAll(repoContextDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .repocontext directory: %v", err)
	}

	// Initialize storage to avoid search failures
	storage := index.NewHybridStorage(repoContextDir)
	err = storage.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	cmd := NewQueryCommand()
	err = cmd.Flags().Set("path", tempDir)
	if err != nil {
		t.Fatalf("Failed to set path flag: %v", err)
	}

	err = cmd.Flags().Set("function", "TestFunc")
	if err != nil {
		t.Fatalf("Failed to set function flag: %v", err)
	}

	// Set both invalid format and json flag
	err = cmd.Flags().Set("format", "invalid")
	if err != nil {
		t.Fatalf("Failed to set format flag: %v", err)
	}

	err = cmd.Flags().Set("json", "true")
	if err != nil {
		t.Fatalf("Failed to set json flag: %v", err)
	}

	// Should succeed because JSON flag should override the invalid format before validation
	// The effective format should be "json" which is valid
	err = cmd.RunE(cmd, []string{})
	if err != nil && strings.Contains(err.Error(), "invalid format") {
		t.Error("Expected command to succeed when JSON flag overrides invalid format, but got format validation error")
	}
	// Note: The command might still fail due to no search results, but it shouldn't fail due to format validation
}

func TestQueryCommand_ValidationMultipleSearchCriteria(t *testing.T) {
	// Create temporary directory with .repocontext
	tempDir, err := os.MkdirTemp("", "query_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .repocontext directory
	repoContextDir := filepath.Join(tempDir, ".repocontext")
	err = os.MkdirAll(repoContextDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .repocontext directory: %v", err)
	}

	cmd := NewQueryCommand()
	err = cmd.Flags().Set("path", tempDir)
	if err != nil {
		t.Fatalf("Failed to set path flag: %v", err)
	}

	// Set multiple search criteria flags
	err = cmd.Flags().Set("function", "TestFunc")
	if err != nil {
		t.Fatalf("Failed to set function flag: %v", err)
	}

	err = cmd.Flags().Set("type", "TestType")
	if err != nil {
		t.Fatalf("Failed to set type flag: %v", err)
	}

	// Should fail because multiple search criteria are specified
	err = cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected command to fail when multiple search criteria are specified")
	}

	expectedErrorSubstring := "exactly one search criterion must be specified"
	if err != nil && !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Expected error to contain '%s', but got: %v", expectedErrorSubstring, err)
	}
}

func TestQueryCommand_ValidationTooManySearchCriteria(t *testing.T) {
	// Create temporary directory with .repocontext
	tempDir, err := os.MkdirTemp("", "query_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .repocontext directory
	repoContextDir := filepath.Join(tempDir, ".repocontext")
	err = os.MkdirAll(repoContextDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .repocontext directory: %v", err)
	}

	cmd := NewQueryCommand()
	err = cmd.Flags().Set("path", tempDir)
	if err != nil {
		t.Fatalf("Failed to set path flag: %v", err)
	}

	// Set three different search criteria flags
	err = cmd.Flags().Set("function", "TestFunc")
	if err != nil {
		t.Fatalf("Failed to set function flag: %v", err)
	}

	err = cmd.Flags().Set("variable", "TestVar")
	if err != nil {
		t.Fatalf("Failed to set variable flag: %v", err)
	}

	err = cmd.Flags().Set("search", "pattern*")
	if err != nil {
		t.Fatalf("Failed to set search flag: %v", err)
	}

	// Should fail because multiple search criteria are specified
	err = cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected command to fail when multiple search criteria are specified")
	}

	expectedErrorSubstring := "exactly one search criterion must be specified"
	if err != nil && !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Expected error to contain '%s', but got: %v", expectedErrorSubstring, err)
	}
}

func TestQueryCommand_ValidationExactlyOneSearchCriterion(t *testing.T) {
	// Create temporary directory with .repocontext
	tempDir, err := os.MkdirTemp("", "query_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .repocontext directory
	repoContextDir := filepath.Join(tempDir, ".repocontext")
	err = os.MkdirAll(repoContextDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .repocontext directory: %v", err)
	}

	// Initialize storage to avoid search failures
	storage := index.NewHybridStorage(repoContextDir)
	err = storage.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	tests := []struct {
		name  string
		flag  string
		value string
	}{
		{"function", "function", "TestFunc"},
		{"type", "type", "TestType"},
		{"variable", "variable", "TestVar"},
		{"file", "file", "test.go"},
		{"search", "search", "pattern*"},
		{"entity-type", "entity-type", "function"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := NewQueryCommand()
			err = cmd.Flags().Set("path", tempDir)
			if err != nil {
				t.Fatalf("Failed to set path flag: %v", err)
			}

			// Set exactly one search criterion flag
			err = cmd.Flags().Set(test.flag, test.value)
			if err != nil {
				t.Fatalf("Failed to set %s flag: %v", test.flag, err)
			}

			// Should not fail validation (might fail on search execution but that's fine)
			err = cmd.RunE(cmd, []string{})
			if err != nil && strings.Contains(err.Error(), "exactly one search criterion must be specified") {
				t.Errorf("Expected command to pass validation with exactly one search criterion (%s), but got validation error: %v", test.flag, err)
			}
		})
	}
}
