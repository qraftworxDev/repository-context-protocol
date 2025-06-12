package index

import (
	"os"
	"testing"
	"time"

	"repository-context-protocol/internal/models"
)

func TestNewQueryEngine(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "query_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := NewHybridStorage(tempDir)
	err = storage.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	engine := NewQueryEngine(storage)
	if engine == nil {
		t.Fatal("Expected query engine to be created, got nil")
	}

	if engine.storage != storage {
		t.Error("Expected query engine to have correct storage reference")
	}
}

func TestQueryEngine_SearchByName(t *testing.T) {
	// Create temporary directory and initialize storage
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	// Create test data
	setupTestData(t, storage)

	engine := NewQueryEngine(storage)

	// Test searching for a function by name
	results, err := engine.SearchByName("TestFunction")
	if err != nil {
		t.Fatalf("Failed to search by name: %v", err)
	}

	if len(results.Entries) == 0 {
		t.Error("Expected to find TestFunction, got no results")
	}

	// Verify result structure
	if results.Query != "TestFunction" {
		t.Errorf("Expected query 'TestFunction', got %s", results.Query)
	}

	if results.SearchType != "name" {
		t.Errorf("Expected search type 'name', got %s", results.SearchType)
	}
}

func TestQueryEngine_SearchByType(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestData(t, storage)

	engine := NewQueryEngine(storage)

	// Test searching for all functions
	results, err := engine.SearchByType("function")
	if err != nil {
		t.Fatalf("Failed to search by type: %v", err)
	}

	if len(results.Entries) == 0 {
		t.Error("Expected to find functions, got no results")
	}

	// Verify all results are functions
	for _, entry := range results.Entries {
		if entry.IndexEntry.Type != "function" {
			t.Errorf("Expected all results to be functions, got %s", entry.IndexEntry.Type)
		}
	}
}

func TestQueryEngine_SearchWithCallGraph(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallGraph(t, storage)

	engine := NewQueryEngine(storage)

	// Test searching with call graph options
	options := QueryOptions{
		IncludeCallers: true,
		IncludeCallees: true,
		MaxDepth:       2,
		MaxTokens:      1000,
	}

	results, err := engine.SearchByNameWithOptions("MainFunction", options)
	if err != nil {
		t.Fatalf("Failed to search with call graph: %v", err)
	}

	if len(results.Entries) == 0 {
		t.Error("Expected to find MainFunction, got no results")
	}

	// Should include callers and callees
	if len(results.CallGraph.Callers) == 0 && len(results.CallGraph.Callees) == 0 {
		t.Error("Expected call graph information to be included")
	}
}

func TestQueryEngine_SearchWithTokenLimit(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestData(t, storage)

	engine := NewQueryEngine(storage)

	// Test with very low token limit
	options := QueryOptions{
		MaxTokens: 20, // Very low limit to force truncation
	}

	results, err := engine.SearchByNameWithOptions("TestFunction", options)
	if err != nil {
		t.Fatalf("Failed to search with token limit: %v", err)
	}

	t.Logf("Token count: %d, Entries: %d, Truncated: %t", results.TokenCount, len(results.Entries), results.Truncated)

	if results.TokenCount > 20 {
		t.Errorf("Expected token count <= 20, got %d", results.TokenCount)
	}

	if !results.Truncated {
		t.Error("Expected results to be marked as truncated")
	}
}

func TestQueryEngine_SearchByPattern(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestData(t, storage)

	engine := NewQueryEngine(storage)

	// Test pattern search
	results, err := engine.SearchByPattern("Test*")
	if err != nil {
		t.Fatalf("Failed to search by pattern: %v", err)
	}

	if len(results.Entries) == 0 {
		t.Error("Expected to find entities matching Test*, got no results")
	}

	// Verify all results match pattern
	for _, entry := range results.Entries {
		if !matchesPattern(entry.IndexEntry.Name, "Test*") {
			t.Errorf("Expected result %s to match pattern Test*", entry.IndexEntry.Name)
		}
	}
}

func TestQueryEngine_SearchInFile(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestData(t, storage)

	engine := NewQueryEngine(storage)

	// Test searching within a specific file
	results, err := engine.SearchInFile("main.go")
	if err != nil {
		t.Fatalf("Failed to search in file: %v", err)
	}

	if len(results.Entries) == 0 {
		t.Error("Expected to find entities in main.go, got no results")
	}

	// Verify all results are from the specified file
	for _, entry := range results.Entries {
		if entry.IndexEntry.File != "main.go" {
			t.Errorf("Expected all results from main.go, got %s", entry.IndexEntry.File)
		}
	}
}

func TestQueryEngine_GetCallGraph(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallGraph(t, storage)

	engine := NewQueryEngine(storage)

	// Test getting call graph for a function
	callGraph, err := engine.GetCallGraph("MainFunction", 2)
	if err != nil {
		t.Fatalf("Failed to get call graph: %v", err)
	}

	if callGraph.Function != "MainFunction" {
		t.Errorf("Expected function 'MainFunction', got %s", callGraph.Function)
	}

	if len(callGraph.Callers) == 0 && len(callGraph.Callees) == 0 {
		t.Error("Expected call graph to have callers or callees")
	}
}

func TestQueryEngine_FormatResults(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestData(t, storage)

	engine := NewQueryEngine(storage)

	results, err := engine.SearchByName("TestFunction")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	// Test JSON formatting
	jsonOutput, err := engine.FormatResults(results, "json")
	if err != nil {
		t.Fatalf("Failed to format as JSON: %v", err)
	}

	if len(jsonOutput) == 0 {
		t.Error("Expected JSON output to be non-empty")
	}

	// Test text formatting
	textOutput, err := engine.FormatResults(results, "text")
	if err != nil {
		t.Fatalf("Failed to format as text: %v", err)
	}

	if len(textOutput) == 0 {
		t.Error("Expected text output to be non-empty")
	}
}

func TestQueryEngine_EstimateTokens(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestData(t, storage)

	engine := NewQueryEngine(storage)

	results, err := engine.SearchByName("TestFunction")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	tokenCount := engine.EstimateTokens(results)
	if tokenCount <= 0 {
		t.Error("Expected positive token count")
	}

	// Token count should be reasonable (not too high or too low)
	if tokenCount > 10000 {
		t.Errorf("Token count seems too high: %d", tokenCount)
	}
}

// Helper functions for test setup

func setupTestStorage(t *testing.T) (string, *HybridStorage) {
	tempDir, err := os.MkdirTemp("", "query_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	storage := NewHybridStorage(tempDir)
	err = storage.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	return tempDir, storage
}

func setupTestData(t *testing.T, storage *HybridStorage) {
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

func setupTestDataWithCallGraph(t *testing.T, storage *HybridStorage) {
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

func matchesPattern(name, pattern string) bool {
	// Simple pattern matching for testing (just prefix matching with *)
	if pattern == "*" {
		return true
	}
	if pattern != "" && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(name) >= len(prefix) && name[:len(prefix)] == prefix
	}
	return name == pattern
}
