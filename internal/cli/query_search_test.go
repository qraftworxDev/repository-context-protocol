package cli

import (
	"os"
	"path/filepath"
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

func TestQueryCommand_SearchWithCallGraph(t *testing.T) {
	tempDir, storage := setupTestRepository(t)
	defer os.RemoveAll(tempDir)
	defer storage.Close()

	addTestDataWithCallGraph(t, storage)

	cmd := NewQueryCommand()
	setRequiredFlags(t, cmd, tempDir, "function", "MainFunction")

	err := cmd.Flags().Set("include-callers", "true")
	if err != nil {
		t.Fatalf("Failed to set include-callers flag: %v", err)
	}

	err = cmd.Flags().Set("include-callees", "true")
	if err != nil {
		t.Fatalf("Failed to set include-callees flag: %v", err)
	}

	// Run command - should succeed
	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Expected search with call graph to succeed, got error: %v", err)
	}
}

func TestQueryCommand_SearchWithJSONOutput(t *testing.T) {
	tempDir, storage := setupTestRepository(t)
	defer os.RemoveAll(tempDir)
	defer storage.Close()

	addTestData(t, storage)

	cmd := NewQueryCommand()
	setRequiredFlags(t, cmd, tempDir, "function", "TestFunction")

	err := cmd.Flags().Set("json", "true")
	if err != nil {
		t.Fatalf("Failed to set json flag: %v", err)
	}

	// Run command - should succeed
	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Expected JSON output to succeed, got error: %v", err)
	}
}

func TestQueryCommand_SearchWithTokenLimit(t *testing.T) {
	tempDir, storage := setupTestRepository(t)
	defer os.RemoveAll(tempDir)
	defer storage.Close()

	addTestData(t, storage)

	cmd := NewQueryCommand()
	setRequiredFlags(t, cmd, tempDir, "function", "TestFunction")

	err := cmd.Flags().Set("max-tokens", "100")
	if err != nil {
		t.Fatalf("Failed to set max-tokens flag: %v", err)
	}

	// Run command - should succeed
	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Expected search with token limit to succeed, got error: %v", err)
	}
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
