package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"repository-context-protocol/internal/index"
	"repository-context-protocol/internal/models"

	"github.com/spf13/cobra"
)

// testCaseSearch represents a search test case
type testCaseSearch struct {
	name      string
	flagName  string
	flagValue string
	setupFunc func(*testing.T, *index.HybridStorage)
}

// testQueryWithFlags represents a test configuration for query commands
type testQueryWithFlags struct {
	searchType    string            // "function", "file", etc.
	searchValue   string            // "MainFunction", "main.go", etc.
	useCallGraph  bool              // whether to set up call graph data
	flags         map[string]string // flags to set on the command
	expectedError string            // expected error message (empty if success expected)
}

func TestQueryCommand_SearchOperations(t *testing.T) {
	tests := []testCaseSearch{
		{
			name:      "SearchByFunction",
			flagName:  "function",
			flagValue: "TestFunction",
			setupFunc: addTestData,
		},
		{
			name:      "SearchByType",
			flagName:  "type",
			flagValue: "TestStruct",
			setupFunc: addTestData,
		},
		{
			name:      "SearchByVariable",
			flagName:  "variable",
			flagValue: "TestVar",
			setupFunc: addTestData,
		},
		{
			name:      "SearchByFile",
			flagName:  "file",
			flagValue: "main.go",
			setupFunc: addTestData,
		},
		{
			name:      "SearchByPattern",
			flagName:  "search",
			flagValue: "Test*",
			setupFunc: addTestData,
		},
		{
			name:      "SearchByEntityType",
			flagName:  "entity-type",
			flagValue: "function",
			setupFunc: addTestData,
		},
		{
			name:      "SearchNotFound",
			flagName:  "function",
			flagValue: "NonExistentFunction",
			setupFunc: addTestData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runSearchTest(t, tt.flagName, tt.flagValue, tt.setupFunc)
		})
	}
}

// runQueryTest is a helper function to reduce duplication in query tests
func runQueryTest(t *testing.T, config testQueryWithFlags) {
	tempDir, storage := setupTestRepository(t)
	defer os.RemoveAll(tempDir)
	defer storage.Close()

	// Set up test data based on configuration
	if config.useCallGraph {
		addTestDataWithCallGraph(t, storage)
	} else {
		addTestData(t, storage)
	}

	cmd := NewQueryCommand()
	setRequiredFlags(t, cmd, tempDir, config.searchType, config.searchValue)

	// Set additional flags
	for flagName, flagValue := range config.flags {
		err := cmd.Flags().Set(flagName, flagValue)
		if err != nil {
			t.Fatalf("Failed to set %s flag: %v", flagName, err)
		}
	}

	// Run command
	err := cmd.RunE(cmd, []string{})

	// Check result based on expected error
	if config.expectedError == "" {
		// Expect success
		if err != nil {
			t.Errorf("Expected command to succeed, got error: %v", err)
		}
	} else {
		// Expect specific error
		if err == nil {
			t.Errorf("Expected error '%s', but command succeeded", config.expectedError)
		} else if !strings.Contains(err.Error(), config.expectedError) {
			t.Errorf("Expected error containing '%s', got: %v", config.expectedError, err)
		}
	}
}

func TestQueryCommand_SearchWithCallGraph(t *testing.T) {
	runQueryTest(t, testQueryWithFlags{
		searchType:   "function",
		searchValue:  "MainFunction",
		useCallGraph: true,
		flags: map[string]string{
			"include-callers": "true",
			"include-callees": "true",
		},
	})
}

func TestQueryCommand_SearchWithJSONOutput(t *testing.T) {
	runQueryTest(t, testQueryWithFlags{
		searchType:   "function",
		searchValue:  "TestFunction",
		useCallGraph: false,
		flags: map[string]string{
			"json": "true",
		},
	})
}

func TestQueryCommand_SearchWithTokenLimit(t *testing.T) {
	runQueryTest(t, testQueryWithFlags{
		searchType:   "function",
		searchValue:  "TestFunction",
		useCallGraph: false,
		flags: map[string]string{
			"max-tokens": "100",
		},
	})
}

func TestQueryCommand_SearchFileWithCallGraph(t *testing.T) {
	runQueryTest(t, testQueryWithFlags{
		searchType:   "file",
		searchValue:  "main.go",
		useCallGraph: true,
		flags: map[string]string{
			"include-callers": "true",
			"include-callees": "true",
		},
	})
}

func TestQueryCommand_SearchFileWithTokenLimit(t *testing.T) {
	runQueryTest(t, testQueryWithFlags{
		searchType:   "file",
		searchValue:  "main.go",
		useCallGraph: true,
		flags: map[string]string{
			"max-tokens": "50",
			"json":       "true",
		},
	})
}

func TestQueryCommand_SearchFileWithIncludeTypes(t *testing.T) {
	runQueryTest(t, testQueryWithFlags{
		searchType:   "file",
		searchValue:  "main.go",
		useCallGraph: true,
		flags: map[string]string{
			"include-types": "true",
		},
	})
}

func TestQueryCommand_SearchFileWithMultipleFlags(t *testing.T) {
	runQueryTest(t, testQueryWithFlags{
		searchType:   "file",
		searchValue:  "main.go",
		useCallGraph: true,
		flags: map[string]string{
			"include-callers": "true",
			"include-callees": "true",
			"include-types":   "true",
			"max-tokens":      "1000",
			"json":            "true",
			"depth":           "3",
		},
	})
}

func TestQueryCommand_FunctionWithOnlyCallers(t *testing.T) {
	tempDir, storage := setupTestRepository(t)
	defer os.RemoveAll(tempDir)
	defer storage.Close()

	addTestDataWithCallGraph(t, storage)

	cmd := NewQueryCommand()
	setRequiredFlags(t, cmd, tempDir, "function", "HelperFunction")

	err := cmd.Flags().Set("include-callers", "true")
	if err != nil {
		t.Fatalf("Failed to set include-callers flag: %v", err)
	}

	// Explicitly set include-callees to false to test selective behavior
	err = cmd.Flags().Set("include-callees", "false")
	if err != nil {
		t.Fatalf("Failed to set include-callees flag: %v", err)
	}

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Expected function search with only callers to succeed, got error: %v", err)
	}

	output := buf.String()

	// Should contain "Callers:" section
	if !strings.Contains(output, "Callers:") {
		t.Error("Expected output to contain 'Callers:' section when include-callers is true")
	}

	// Should NOT contain "Callees:" section
	if strings.Contains(output, "Callees:") {
		t.Error("Expected output to NOT contain 'Callees:' section when include-callees is false")
	}
}

func TestQueryCommand_FunctionWithOnlyCallees(t *testing.T) {
	tempDir, storage := setupTestRepository(t)
	defer os.RemoveAll(tempDir)
	defer storage.Close()

	addTestDataWithCallGraph(t, storage)

	cmd := NewQueryCommand()
	setRequiredFlags(t, cmd, tempDir, "function", "MainFunction")

	// Explicitly set include-callers to false to test selective behavior
	err := cmd.Flags().Set("include-callers", "false")
	if err != nil {
		t.Fatalf("Failed to set include-callers flag: %v", err)
	}

	err = cmd.Flags().Set("include-callees", "true")
	if err != nil {
		t.Fatalf("Failed to set include-callees flag: %v", err)
	}

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Expected function search with only callees to succeed, got error: %v", err)
	}

	output := buf.String()

	// Should NOT contain "Callers:" section
	if strings.Contains(output, "Callers:") {
		t.Error("Expected output to NOT contain 'Callers:' section when include-callers is false")
	}

	// Should contain "Callees:" section
	if !strings.Contains(output, "Callees:") {
		t.Error("Expected output to contain 'Callees:' section when include-callees is true")
	}
}

func TestQueryCommand_FunctionWithBothCallGraph(t *testing.T) {
	tempDir, storage := setupTestRepository(t)
	defer os.RemoveAll(tempDir)
	defer storage.Close()

	addTestDataWithCallGraph(t, storage)

	cmd := NewQueryCommand()
	setRequiredFlags(t, cmd, tempDir, "function", "HelperFunction")

	err := cmd.Flags().Set("include-callers", "true")
	if err != nil {
		t.Fatalf("Failed to set include-callers flag: %v", err)
	}

	err = cmd.Flags().Set("include-callees", "true")
	if err != nil {
		t.Fatalf("Failed to set include-callees flag: %v", err)
	}

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Expected function search with both call graph flags to succeed, got error: %v", err)
	}

	output := buf.String()

	// Should contain both "Callers:" and "Callees:" sections
	if !strings.Contains(output, "Callers:") {
		t.Error("Expected output to contain 'Callers:' section when include-callers is true")
	}

	if !strings.Contains(output, "Callees:") {
		t.Error("Expected output to contain 'Callees:' section when include-callees is true")
	}
}

func TestQueryCommand_FunctionWithNoCallGraph(t *testing.T) {
	tempDir, storage := setupTestRepository(t)
	defer os.RemoveAll(tempDir)
	defer storage.Close()

	addTestDataWithCallGraph(t, storage)

	cmd := NewQueryCommand()
	setRequiredFlags(t, cmd, tempDir, "function", "MainFunction")

	// Explicitly set both flags to false
	err := cmd.Flags().Set("include-callers", "false")
	if err != nil {
		t.Fatalf("Failed to set include-callers flag: %v", err)
	}

	err = cmd.Flags().Set("include-callees", "false")
	if err != nil {
		t.Fatalf("Failed to set include-callees flag: %v", err)
	}

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Expected function search with no call graph to succeed, got error: %v", err)
	}

	output := buf.String()

	// Should NOT contain "Call Graph:" section at all
	if strings.Contains(output, "Call Graph:") {
		t.Error("Expected output to NOT contain 'Call Graph:' section when both flags are false")
	}

	// Should NOT contain "Callers:" or "Callees:" sections
	if strings.Contains(output, "Callers:") {
		t.Error("Expected output to NOT contain 'Callers:' section when include-callers is false")
	}

	if strings.Contains(output, "Callees:") {
		t.Error("Expected output to NOT contain 'Callees:' section when include-callees is false")
	}
}

func TestQueryCommand_JSONOutputCallGraphFlags(t *testing.T) {
	tempDir, storage := setupTestRepository(t)
	defer os.RemoveAll(tempDir)
	defer storage.Close()

	addTestDataWithCallGraph(t, storage)

	cmd := NewQueryCommand()
	setRequiredFlags(t, cmd, tempDir, "function", "HelperFunction")

	err := cmd.Flags().Set("include-callers", "true")
	if err != nil {
		t.Fatalf("Failed to set include-callers flag: %v", err)
	}

	err = cmd.Flags().Set("include-callees", "false")
	if err != nil {
		t.Fatalf("Failed to set include-callees flag: %v", err)
	}

	err = cmd.Flags().Set("json", "true")
	if err != nil {
		t.Fatalf("Failed to set json flag: %v", err)
	}

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Expected JSON function search with selective call graph to succeed, got error: %v", err)
	}

	output := buf.String()

	// Parse JSON to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify call_graph exists and has the expected structure
	callGraphRaw, exists := result["call_graph"]
	if !exists {
		t.Fatal("Expected call_graph field in JSON output")
	}

	callGraph, ok := callGraphRaw.(map[string]interface{})
	if !ok {
		t.Fatal("Expected call_graph to be an object")
	}

	// Should have callers but not callees in the JSON
	if _, hasCallers := callGraph["callers"]; !hasCallers {
		t.Error("Expected 'callers' field in call_graph when include-callers is true")
	}

	if callees, hasCallees := callGraph["callees"]; hasCallees {
		// If callees field exists, it should be empty or null
		if callees != nil {
			if calleesSlice, ok := callees.([]interface{}); ok && len(calleesSlice) > 0 {
				t.Error("Expected empty or null 'callees' field when include-callees is false")
			}
		}
	}
}

func TestQueryCommand_SearchByPatternWithOptions(t *testing.T) {
	runQueryTest(t, testQueryWithFlags{
		searchType:   "search",
		searchValue:  "Test*",
		useCallGraph: true,
		flags: map[string]string{
			"include-callers": "true",
			"include-callees": "true",
			"max-tokens":      "1000",
		},
	})
}

func TestQueryCommand_SearchByEntityTypeWithOptions(t *testing.T) {
	runQueryTest(t, testQueryWithFlags{
		searchType:   "entity-type",
		searchValue:  "function",
		useCallGraph: true,
		flags: map[string]string{
			"include-callers": "true",
			"include-callees": "true",
			"max-tokens":      "500",
		},
	})
}

func TestQueryCommand_SearchByPatternWithTokenLimit(t *testing.T) {
	runQueryTest(t, testQueryWithFlags{
		searchType:   "search",
		searchValue:  "Test*",
		useCallGraph: true,
		flags: map[string]string{
			"max-tokens": "50",
			"json":       "true",
		},
	})
}

func TestQueryCommand_SearchByEntityTypeWithTokenLimit(t *testing.T) {
	runQueryTest(t, testQueryWithFlags{
		searchType:   "entity-type",
		searchValue:  "function",
		useCallGraph: true,
		flags: map[string]string{
			"max-tokens": "30",
			"json":       "true",
		},
	})
}

func TestQueryCommand_SearchByPatternWithIncludeTypes(t *testing.T) {
	runQueryTest(t, testQueryWithFlags{
		searchType:   "search",
		searchValue:  "Test*",
		useCallGraph: true,
		flags: map[string]string{
			"include-types": "true",
		},
	})
}

func TestQueryCommand_SearchByEntityTypeWithIncludeTypes(t *testing.T) {
	runQueryTest(t, testQueryWithFlags{
		searchType:   "entity-type",
		searchValue:  "function",
		useCallGraph: true,
		flags: map[string]string{
			"include-types": "true",
		},
	})
}

// Helper functions

func runSearchTest(t *testing.T, flagName, flagValue string, setupFunc func(*testing.T, *index.HybridStorage)) {
	tempDir, storage := setupTestRepository(t)
	defer os.RemoveAll(tempDir)
	defer storage.Close()

	setupFunc(t, storage)

	cmd := NewQueryCommand()
	setRequiredFlags(t, cmd, tempDir, flagName, flagValue)

	// Run command - should succeed
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Expected search to succeed, got error: %v", err)
	}
}

func setRequiredFlags(t *testing.T, cmd *cobra.Command, tempDir, flagName, flagValue string) {
	err := cmd.Flags().Set("path", tempDir)
	if err != nil {
		t.Fatalf("Failed to set path flag: %v", err)
	}

	err = cmd.Flags().Set(flagName, flagValue)
	if err != nil {
		t.Fatalf("Failed to set %s flag: %v", flagName, err)
	}
}

func setupTestRepository(t *testing.T) (string, *index.HybridStorage) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "query_search_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create .repocontext directory
	repoContextDir := filepath.Join(tempDir, ".repocontext")
	err = os.MkdirAll(repoContextDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .repocontext directory: %v", err)
	}

	// Create chunks subdirectory
	chunksDir := filepath.Join(repoContextDir, "chunks")
	err = os.MkdirAll(chunksDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create chunks directory: %v", err)
	}

	// Initialize storage
	storage := index.NewHybridStorage(repoContextDir)
	err = storage.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	return tempDir, storage
}

func addTestData(t *testing.T, storage *index.HybridStorage) {
	// Create test file context
	fileContext := &models.FileContext{
		Path:     "main.go",
		Language: "go",
		Checksum: "test123",
		ModTime:  time.Now(),
		Functions: []models.Function{
			{
				Name:      "TestFunction",
				Signature: "func TestFunction() error",
				StartLine: 10,
				EndLine:   15,
			},
			{
				Name:      "AnotherFunction",
				Signature: "func AnotherFunction(s string) int",
				StartLine: 20,
				EndLine:   25,
			},
		},
		Types: []models.TypeDef{
			{
				Name:      "TestStruct",
				Kind:      "struct",
				StartLine: 5,
				EndLine:   8,
			},
		},
		Variables: []models.Variable{
			{
				Name:      "TestVar",
				Type:      "string",
				StartLine: 3,
				EndLine:   3,
			},
		},
	}

	// Store the file context
	err := storage.StoreFileContext(fileContext)
	if err != nil {
		t.Fatalf("Failed to store test data: %v", err)
	}
}

func addTestDataWithCallGraph(t *testing.T, storage *index.HybridStorage) {
	// Create test file context with call relationships
	fileContext := &models.FileContext{
		Path:     "main.go",
		Language: "go",
		Checksum: "test123",
		ModTime:  time.Now(),
		Functions: []models.Function{
			{
				Name:      "MainFunction",
				Signature: "func MainFunction()",
				StartLine: 10,
				EndLine:   15,
				Calls:     []string{"HelperFunction", "fmt.Println"},
			},
			{
				Name:      "HelperFunction",
				Signature: "func HelperFunction() string",
				StartLine: 20,
				EndLine:   25,
				CalledBy:  []string{"MainFunction"},
			},
		},
	}

	err := storage.StoreFileContext(fileContext)
	if err != nil {
		t.Fatalf("Failed to store test data with call graph: %v", err)
	}
}
