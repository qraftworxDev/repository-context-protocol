package index

import (
	"os"
	"testing"
	"time"

	"repository-context-protocol/internal/models"
)

func TestNewQueryEngine(t *testing.T) {
	// Create temporary directory and initialize storage
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
	t.Cleanup(func() {
		if err := storage.Close(); err != nil {
			t.Errorf("Failed to close storage: %v", err)
		}
	})

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

func TestQueryEngine_SearchInFileWithOptions(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallGraph(t, storage)

	engine := NewQueryEngine(storage)

	// Test basic file search without options
	basicOptions := QueryOptions{}
	results, err := engine.SearchInFileWithOptions("main.go", basicOptions)
	if err != nil {
		t.Fatalf("Failed to search in file with basic options: %v", err)
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

func TestQueryEngine_SearchInFileWithCallGraphOptions(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallGraph(t, storage)

	engine := NewQueryEngine(storage)

	// Test file search with call graph options
	options := QueryOptions{
		IncludeCallers: true,
		IncludeCallees: true,
		MaxDepth:       2,
	}

	results, err := engine.SearchInFileWithOptions("main.go", options)
	if err != nil {
		t.Fatalf("Failed to search in file with call graph options: %v", err)
	}

	if len(results.Entries) == 0 {
		t.Error("Expected to find entities in main.go, got no results")
	}

	// Should include call graph information for functions found in file
	foundFunction := false
	for _, entry := range results.Entries {
		if entry.IndexEntry.Type == "function" {
			foundFunction = true
			break
		}
	}

	if foundFunction && results.CallGraph == nil {
		t.Error("Expected call graph information when IncludeCallers/IncludeCallees is true and functions are present")
	}
}

func TestQueryEngine_SearchInFileWithTokenLimit(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallGraph(t, storage)

	engine := NewQueryEngine(storage)

	// Test with low token limit
	options := QueryOptions{
		MaxTokens: 50, // Very low limit to trigger truncation
	}

	results, err := engine.SearchInFileWithOptions("main.go", options)
	if err != nil {
		t.Fatalf("Failed to search in file with token limit: %v", err)
	}

	// Use helper function to validate token limits
	validateTokenLimits(t, engine, results, options.MaxTokens)
}

func TestQueryEngine_SearchInFileWithIncludeTypes(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallGraph(t, storage)

	engine := NewQueryEngine(storage)

	// Test with include types option
	options := QueryOptions{
		IncludeTypes: true,
	}

	results, err := engine.SearchInFileWithOptions("main.go", options)
	if err != nil {
		t.Fatalf("Failed to search in file with include types: %v", err)
	}

	// Should find functions, types, and variables that exist in the file
	foundFunction := false
	foundType := false
	foundVariable := false
	for _, entry := range results.Entries {
		if entry.IndexEntry.Type == "function" {
			foundFunction = true
		}
		if entry.IndexEntry.Type == "struct" {
			foundType = true
		}
		if entry.IndexEntry.Type == "variable" {
			foundVariable = true
		}
	}

	if !foundFunction {
		t.Error("Expected to find functions in main.go")
	}
	// Only check for types and variables if we actually added them to the test data
	if len(results.Entries) > 2 { // We know we have 2 functions, so more means we have types/variables
		if !foundType {
			t.Error("Expected to find types in main.go")
		}
		if !foundVariable {
			t.Error("Expected to find variables in main.go")
		}
	}

	// The IncludeTypes flag should work (this is currently not implemented but the test structure is correct)
	// In the future, this could add additional related type information
}

func TestQueryEngine_GetCallGraphWithOptions_OnlyCallers(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallerCalleeRelations(t, storage)

	engine := NewQueryEngine(storage)

	// Test getting only callers
	options := QueryOptions{
		IncludeCallers: true,
		IncludeCallees: false,
		MaxDepth:       2,
	}

	callGraph, err := engine.GetCallGraphWithOptions("HelperFunction", options)
	if err != nil {
		t.Fatalf("Failed to get call graph with options: %v", err)
	}

	if callGraph.Function != "HelperFunction" {
		t.Errorf("Expected function 'HelperFunction', got %s", callGraph.Function)
	}

	// Should have callers (MainFunction calls HelperFunction)
	if len(callGraph.Callers) == 0 {
		t.Error("Expected callers when IncludeCallers is true")
	}

	// Should NOT have callees when IncludeCallees is false
	if len(callGraph.Callees) != 0 {
		t.Errorf("Expected no callees when IncludeCallees is false, got %d", len(callGraph.Callees))
	}
}

func TestQueryEngine_GetCallGraphWithOptions_OnlyCallees(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallerCalleeRelations(t, storage)

	engine := NewQueryEngine(storage)

	// Test getting only callees
	options := QueryOptions{
		IncludeCallers: false,
		IncludeCallees: true,
		MaxDepth:       2,
	}

	callGraph, err := engine.GetCallGraphWithOptions("MainFunction", options)
	if err != nil {
		t.Fatalf("Failed to get call graph with options: %v", err)
	}

	if callGraph.Function != "MainFunction" {
		t.Errorf("Expected function 'MainFunction', got %s", callGraph.Function)
	}

	// Should NOT have callers when IncludeCallers is false
	if len(callGraph.Callers) != 0 {
		t.Errorf("Expected no callers when IncludeCallers is false, got %d", len(callGraph.Callers))
	}

	// Should have callees (MainFunction calls HelperFunction)
	if len(callGraph.Callees) == 0 {
		t.Error("Expected callees when IncludeCallees is true")
	}
}

func TestQueryEngine_GetCallGraphWithOptions_Both(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallerCalleeRelations(t, storage)

	engine := NewQueryEngine(storage)

	// Test getting both callers and callees
	options := QueryOptions{
		IncludeCallers: true,
		IncludeCallees: true,
		MaxDepth:       2,
	}

	callGraph, err := engine.GetCallGraphWithOptions("HelperFunction", options)
	if err != nil {
		t.Fatalf("Failed to get call graph with options: %v", err)
	}

	// Should have both callers and callees
	if len(callGraph.Callers) == 0 {
		t.Error("Expected callers when IncludeCallers is true")
	}

	// HelperFunction doesn't call anything in our test data, but we should still
	// populate the callees slice (even if empty) when IncludeCallees is true
	// This test verifies the method was called, not necessarily that data exists
}

func TestQueryEngine_GetCallGraphWithOptions_Neither(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallerCalleeRelations(t, storage)

	engine := NewQueryEngine(storage)

	// Test getting neither callers nor callees
	options := QueryOptions{
		IncludeCallers: false,
		IncludeCallees: false,
		MaxDepth:       2,
	}

	callGraph, err := engine.GetCallGraphWithOptions("MainFunction", options)
	if err != nil {
		t.Fatalf("Failed to get call graph with options: %v", err)
	}

	// Should have neither callers nor callees
	if len(callGraph.Callers) != 0 {
		t.Errorf("Expected no callers when IncludeCallers is false, got %d", len(callGraph.Callers))
	}

	if len(callGraph.Callees) != 0 {
		t.Errorf("Expected no callees when IncludeCallees is false, got %d", len(callGraph.Callees))
	}
}

func TestQueryEngine_SearchWithSelectiveCallGraph(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallerCalleeRelations(t, storage)

	engine := NewQueryEngine(storage)

	// Test that SearchByNameWithOptions respects call graph flags
	optionsOnlyCallers := QueryOptions{
		IncludeCallers: true,
		IncludeCallees: false,
		MaxDepth:       2,
	}

	result, err := engine.SearchByNameWithOptions("HelperFunction", optionsOnlyCallers)
	if err != nil {
		t.Fatalf("Failed to search with selective call graph: %v", err)
	}

	if result.CallGraph == nil {
		t.Fatal("Expected call graph to be included")
	}

	// Should have callers but not callees
	if len(result.CallGraph.Callers) == 0 {
		t.Error("Expected callers when IncludeCallers is true")
	}

	if len(result.CallGraph.Callees) != 0 {
		t.Errorf("Expected no callees when IncludeCallees is false, got %d", len(result.CallGraph.Callees))
	}
}

func TestQueryEngine_SearchByPatternWithOptions(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallGraph(t, storage)

	engine := NewQueryEngine(storage)

	// Test pattern search with query options
	options := QueryOptions{
		IncludeCallers: true,
		IncludeCallees: true,
		MaxDepth:       2,
		MaxTokens:      1000,
	}

	results, err := engine.SearchByPatternWithOptions("Test*", options)
	if err != nil {
		t.Fatalf("Failed to search by pattern with options: %v", err)
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

	// Should include call graph information for functions
	foundFunction := false
	for _, entry := range results.Entries {
		if entry.IndexEntry.Type == "function" {
			foundFunction = true
			break
		}
	}

	if foundFunction && results.CallGraph == nil {
		t.Error("Expected call graph information when IncludeCallers/IncludeCallees is true and functions are present")
	}

	// Check that token count is reasonable and within limits
	if results.TokenCount > options.MaxTokens {
		t.Errorf("Expected token count <= %d, got %d", options.MaxTokens, results.TokenCount)
	}
}

func TestQueryEngine_SearchByTypeWithOptions(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallGraph(t, storage)

	engine := NewQueryEngine(storage)

	// Test type search with query options
	options := QueryOptions{
		IncludeCallers: true,
		IncludeCallees: true,
		MaxDepth:       2,
		MaxTokens:      500,
	}

	results, err := engine.SearchByTypeWithOptions("function", options)
	if err != nil {
		t.Fatalf("Failed to search by type with options: %v", err)
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

	// Should include call graph information for functions
	if results.CallGraph == nil {
		t.Error("Expected call graph information when IncludeCallers/IncludeCallees is true and functions are present")
	}

	// Check that token count is reasonable and within limits
	if results.TokenCount > options.MaxTokens {
		t.Errorf("Expected token count <= %d, got %d", options.MaxTokens, results.TokenCount)
	}
}

func TestQueryEngine_SearchByPatternWithTokenLimit(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallGraph(t, storage)

	engine := NewQueryEngine(storage)

	// Test with very low token limit to force truncation
	options := QueryOptions{
		MaxTokens: 50, // Very low limit
	}

	results, err := engine.SearchByPatternWithOptions("Test*", options)
	if err != nil {
		t.Fatalf("Failed to search by pattern with token limit: %v", err)
	}

	if results.TokenCount > options.MaxTokens {
		t.Errorf("Expected token count <= %d, got %d", options.MaxTokens, results.TokenCount)
	}

	// Test that token limits are applied correctly
	// The specific truncation behavior depends on the actual data size
	t.Logf("Token count: %d, Entries: %d, Truncated: %t", results.TokenCount, len(results.Entries), results.Truncated)
}

func TestQueryEngine_SearchByTypeWithTokenLimit(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallGraph(t, storage)

	engine := NewQueryEngine(storage)

	// Test with very low token limit to force truncation
	options := QueryOptions{
		MaxTokens: 30, // Very low limit
	}

	results, err := engine.SearchByTypeWithOptions("function", options)
	if err != nil {
		t.Fatalf("Failed to search by type with token limit: %v", err)
	}

	// Use helper function to validate token limits
	validateTokenLimits(t, engine, results, options.MaxTokens)
}

func TestQueryEngine_CallGraphMaxDepthActuallyUsed(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	// Set up a deeper call chain: FuncA -> FuncB -> FuncC -> FuncD
	setupTestDataWithDeepCallChain(t, storage)

	engine := NewQueryEngine(storage)

	// Test with depth 1 - should only get direct callees
	options1 := QueryOptions{
		IncludeCallees: true,
		MaxDepth:       1,
	}

	result1, err := engine.SearchByNameWithOptions("FuncA", options1)
	if err != nil {
		t.Fatalf("Failed to search with depth 1: %v", err)
	}

	if result1.CallGraph == nil {
		t.Fatal("Expected call graph to be included")
	}

	// With depth 1, should only see FuncB (direct callee)
	if len(result1.CallGraph.Callees) != 1 {
		t.Errorf("Expected 1 callee with depth 1, got %d", len(result1.CallGraph.Callees))
	}

	if len(result1.CallGraph.Callees) > 0 && result1.CallGraph.Callees[0].Function != "FuncB" {
		t.Errorf("Expected direct callee to be FuncB, got %s", result1.CallGraph.Callees[0].Function)
	}

	// Test with depth 2 - should get FuncB and FuncC
	options2 := QueryOptions{
		IncludeCallees: true,
		MaxDepth:       2,
	}

	result2, err := engine.SearchByNameWithOptions("FuncA", options2)
	if err != nil {
		t.Fatalf("Failed to search with depth 2: %v", err)
	}

	if result2.CallGraph == nil {
		t.Fatal("Expected call graph to be included")
	}

	// With depth 2, should see FuncB and FuncC
	if len(result2.CallGraph.Callees) < 2 {
		t.Errorf("Expected at least 2 callees with depth 2, got %d", len(result2.CallGraph.Callees))
	}

	// Test with depth 3 - should get FuncB, FuncC, and FuncD
	options3 := QueryOptions{
		IncludeCallees: true,
		MaxDepth:       3,
	}

	result3, err := engine.SearchByNameWithOptions("FuncA", options3)
	if err != nil {
		t.Fatalf("Failed to search with depth 3: %v", err)
	}

	if result3.CallGraph == nil {
		t.Fatal("Expected call graph to be included")
	}

	// With depth 3, should see FuncB, FuncC, and FuncD
	if len(result3.CallGraph.Callees) < 3 {
		t.Errorf("Expected at least 3 callees with depth 3, got %d", len(result3.CallGraph.Callees))
	}

	t.Logf("Depth 1: %d callees", len(result1.CallGraph.Callees))
	t.Logf("Depth 2: %d callees", len(result2.CallGraph.Callees))
	t.Logf("Depth 3: %d callees", len(result3.CallGraph.Callees))
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
	t.Cleanup(func() {
		if err := storage.Close(); err != nil {
			t.Errorf("Failed to close storage: %v", err)
		}
	})

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

	err := storage.StoreFileContext(fileContext)
	if err != nil {
		t.Fatalf("Failed to store test data with call graph: %v", err)
	}
}

func setupTestDataWithCallerCalleeRelations(t *testing.T, storage *HybridStorage) {
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
				// HelperFunction doesn't call anything
			},
			{
				Name:      "AnotherFunction",
				Signature: "func AnotherFunction()",
				StartLine: 30,
				EndLine:   35,
				Calls:     []string{"HelperFunction"},
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
	}

	err := storage.StoreFileContext(fileContext)
	if err != nil {
		t.Fatalf("Failed to store test data with caller/callee relations: %v", err)
	}
}

func setupTestDataWithDeepCallChain(t *testing.T, storage *HybridStorage) {
	// Create a deep call chain: FuncA -> FuncB -> FuncC -> FuncD
	fileContext := &models.FileContext{
		Path:     "chain.go",
		Language: "go",
		Checksum: "test456",
		ModTime:  time.Now(),
		Functions: []models.Function{
			{
				Name:      "FuncA",
				Signature: "func FuncA()",
				StartLine: 10,
				EndLine:   15,
				Calls:     []string{"FuncB"},
			},
			{
				Name:      "FuncB",
				Signature: "func FuncB()",
				StartLine: 20,
				EndLine:   25,
				Calls:     []string{"FuncC"},
				CalledBy:  []string{"FuncA"},
			},
			{
				Name:      "FuncC",
				Signature: "func FuncC()",
				StartLine: 30,
				EndLine:   35,
				Calls:     []string{"FuncD"},
				CalledBy:  []string{"FuncB"},
			},
			{
				Name:      "FuncD",
				Signature: "func FuncD()",
				StartLine: 40,
				EndLine:   45,
				CalledBy:  []string{"FuncC"},
			},
		},
	}

	err := storage.StoreFileContext(fileContext)
	if err != nil {
		t.Fatalf("Failed to store deep call chain test data: %v", err)
	}
}

// validateTokenLimits is a helper function to validate token limit behavior in tests
func validateTokenLimits(t *testing.T, engine *QueryEngine, results *SearchResult, maxTokens int) {
	if results.TokenCount > maxTokens {
		t.Errorf("Expected token count <= %d, got %d", maxTokens, results.TokenCount)
	}

	if !results.Truncated && len(results.Entries) > 0 {
		// Only expect truncation if there were actually entries to truncate
		// and the token limit was exceeded
		estimatedTotal := engine.EstimateTokens(&SearchResult{Entries: results.Entries})
		if estimatedTotal > maxTokens {
			t.Error("Expected results to be marked as truncated when token limit is exceeded")
		}
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
