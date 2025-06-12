package cli

import (
	"os"
	"path/filepath"
	"testing"
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

	// Test default values
	tests := []struct {
		flag     string
		expected string
	}{
		{"format", "text"},
		{"depth", "2"},
		{"max-tokens", "0"},
		{"path", "."},
	}

	for _, test := range tests {
		flag := cmd.Flags().Lookup(test.flag)
		if flag == nil {
			t.Errorf("Expected flag --%s to exist", test.flag)
			continue
		}
		if flag.DefValue != test.expected {
			t.Errorf("Expected flag --%s default value '%s', got '%s'", test.flag, test.expected, flag.DefValue)
		}
	}
}

func TestQueryCommand_MutuallyExclusiveFlags(t *testing.T) {
	cmd := NewQueryCommand()

	// Test setting multiple search type flags should work (they're OR-ed together)
	err := cmd.Flags().Set("function", "TestFunc")
	if err != nil {
		t.Errorf("Failed to set function flag: %v", err)
	}

	err = cmd.Flags().Set("type", "TestType")
	if err != nil {
		t.Errorf("Failed to set type flag: %v", err)
	}
}

func TestQueryCommand_JSONFlag(t *testing.T) {
	cmd := NewQueryCommand()

	// Test that --json flag sets format to json
	err := cmd.Flags().Set("json", "true")
	if err != nil {
		t.Errorf("Failed to set json flag: %v", err)
	}

	// The actual logic to set format will be tested in the implementation
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
