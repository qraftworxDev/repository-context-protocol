package index

import (
	"os"
	"path/filepath"
	"testing"

	"repository-context-protocol/internal/models"
)

func TestGlobalCallGraph_BuildFromFiles(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "callgraph_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files with cross-file function calls
	files := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	user := CreateUser("John")
	ProcessUser(user)
	fmt.Println("Done")
}

func ProcessUser(user *User) {
	ValidateUser(user)
	SaveUser(user)
}`,
		"user.go": `package main

func CreateUser(name string) *User {
	return &User{Name: name}
}

func ValidateUser(user *User) bool {
	return user.Name != ""
}

func SaveUser(user *User) error {
	return nil
}

type User struct {
	Name string
}`,
		"utils.go": `package main

func Helper() {
	ProcessUser(&User{Name: "Helper"})
}`,
	}

	// Create file contexts (simulating what the parser would produce)
	var fileContexts []models.FileContext
	for filename, content := range files {
		filePath := filepath.Join(tempDir, filename)
		err = os.WriteFile(filePath, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}

		// Create mock file context based on the file path
		fileContext := createMockFileContext(filePath)
		fileContexts = append(fileContexts, fileContext)
	}

	// Build global call graph
	callGraph := NewGlobalCallGraph()
	err = callGraph.BuildFromFiles(fileContexts)
	if err != nil {
		t.Fatalf("Failed to build call graph: %v", err)
	}

	// Test: Get all callers of CreateUser
	callers := callGraph.GetCallers("CreateUser")
	if len(callers) != 1 {
		t.Errorf("Expected 1 caller of CreateUser, got %d", len(callers))
	}
	if len(callers) > 0 && callers[0].Caller != "main" {
		t.Errorf("Expected main to call CreateUser, got %s", callers[0].Caller)
	}

	// Test: Get all callees of main
	callees := callGraph.GetCallees("main")
	expectedCallees := []string{"CreateUser", "ProcessUser", "fmt.Println"}
	if len(callees) != len(expectedCallees) {
		t.Errorf("Expected %d callees of main, got %d", len(expectedCallees), len(callees))
	}

	// Test: Get all callers of ProcessUser (should be called from main and Helper)
	processUserCallers := callGraph.GetCallers("ProcessUser")
	if len(processUserCallers) != 2 {
		t.Errorf("Expected 2 callers of ProcessUser, got %d", len(processUserCallers))
	}

	// Verify caller names
	callerNames := make(map[string]bool)
	for _, caller := range processUserCallers {
		callerNames[caller.Caller] = true
	}
	if !callerNames["main"] || !callerNames["Helper"] {
		t.Error("Expected ProcessUser to be called by both main and Helper")
	}
}

func TestGlobalCallGraph_CrossFileAnalysis(t *testing.T) {
	// Create file contexts representing a more complex cross-file scenario
	fileContexts := []models.FileContext{
		{
			Path:     "service.go",
			Language: "go",
			Functions: []models.Function{
				{
					Name:  "StartService",
					Calls: []string{"InitDatabase", "StartServer"},
				},
				{
					Name:  "StopService",
					Calls: []string{"StopServer", "CloseDatabase"},
				},
			},
		},
		{
			Path:     "database.go",
			Language: "go",
			Functions: []models.Function{
				{
					Name:  "InitDatabase",
					Calls: []string{"ConnectDB", "MigrateSchema"},
				},
				{
					Name:  "CloseDatabase",
					Calls: []string{"DisconnectDB"},
				},
				{
					Name:  "ConnectDB",
					Calls: []string{},
				},
				{
					Name:  "DisconnectDB",
					Calls: []string{},
				},
				{
					Name:  "MigrateSchema",
					Calls: []string{"ConnectDB"},
				},
			},
		},
		{
			Path:     "server.go",
			Language: "go",
			Functions: []models.Function{
				{
					Name:  "StartServer",
					Calls: []string{"BindPort", "RegisterRoutes"},
				},
				{
					Name:  "StopServer",
					Calls: []string{"UnbindPort"},
				},
				{
					Name:  "BindPort",
					Calls: []string{},
				},
				{
					Name:  "UnbindPort",
					Calls: []string{},
				},
				{
					Name:  "RegisterRoutes",
					Calls: []string{},
				},
			},
		},
	}

	// Build global call graph
	callGraph := NewGlobalCallGraph()
	err := callGraph.BuildFromFiles(fileContexts)
	if err != nil {
		t.Fatalf("Failed to build call graph: %v", err)
	}

	// Test: ConnectDB should be called by both InitDatabase and MigrateSchema
	connectDBCallers := callGraph.GetCallers("ConnectDB")
	if len(connectDBCallers) != 2 {
		t.Errorf("Expected 2 callers of ConnectDB, got %d", len(connectDBCallers))
	}

	// Test: StartService should call functions across multiple files
	startServiceCallees := callGraph.GetCallees("StartService")
	if len(startServiceCallees) != 2 {
		t.Errorf("Expected 2 callees of StartService, got %d", len(startServiceCallees))
	}

	// Test: Get call chain depth
	chainDepth := callGraph.GetCallChainDepth("StartService", "ConnectDB")
	if chainDepth != 2 { // StartService -> InitDatabase -> ConnectDB
		t.Errorf("Expected call chain depth of 2 from StartService to ConnectDB, got %d", chainDepth)
	}

	// Test: Get all functions in call path
	callPath := callGraph.GetCallPath("StartService", "ConnectDB")
	expectedPath := []string{"StartService", "InitDatabase", "ConnectDB"}
	if len(callPath) != len(expectedPath) {
		t.Errorf("Expected call path length %d, got %d", len(expectedPath), len(callPath))
	}
	for i, expected := range expectedPath {
		if i < len(callPath) && callPath[i] != expected {
			t.Errorf("Expected call path[%d] = %s, got %s", i, expected, callPath[i])
		}
	}
}

func TestGlobalCallGraph_GetAllFunctions(t *testing.T) {
	fileContexts := []models.FileContext{
		{
			Path:     "file1.go",
			Language: "go",
			Functions: []models.Function{
				{Name: "FuncA", Calls: []string{"FuncB"}},
				{Name: "FuncB", Calls: []string{}},
			},
		},
		{
			Path:     "file2.go",
			Language: "go",
			Functions: []models.Function{
				{Name: "FuncC", Calls: []string{"FuncA"}},
				{Name: "FuncD", Calls: []string{}},
			},
		},
	}

	callGraph := NewGlobalCallGraph()
	err := callGraph.BuildFromFiles(fileContexts)
	if err != nil {
		t.Fatalf("Failed to build call graph: %v", err)
	}

	// Test: Get all functions in the call graph
	allFunctions := callGraph.GetAllFunctions()
	expectedFunctions := []string{"FuncA", "FuncB", "FuncC", "FuncD"}
	if len(allFunctions) != len(expectedFunctions) {
		t.Errorf("Expected %d functions, got %d", len(expectedFunctions), len(allFunctions))
	}

	// Verify all expected functions are present
	functionMap := make(map[string]bool)
	for _, fn := range allFunctions {
		functionMap[fn] = true
	}
	for _, expected := range expectedFunctions {
		if !functionMap[expected] {
			t.Errorf("Expected function %s not found in call graph", expected)
		}
	}
}

func TestGlobalCallGraph_GetStatistics(t *testing.T) {
	fileContexts := []models.FileContext{
		{
			Path:     "test.go",
			Language: "go",
			Functions: []models.Function{
				{Name: "Main", Calls: []string{"Helper1", "Helper2"}},
				{Name: "Helper1", Calls: []string{"Utility"}},
				{Name: "Helper2", Calls: []string{"Utility"}},
				{Name: "Utility", Calls: []string{}},
			},
		},
	}

	callGraph := NewGlobalCallGraph()
	err := callGraph.BuildFromFiles(fileContexts)
	if err != nil {
		t.Fatalf("Failed to build call graph: %v", err)
	}

	// Test: Get call graph statistics
	stats := callGraph.GetStatistics()
	if stats.TotalFunctions != 4 {
		t.Errorf("Expected 4 total functions, got %d", stats.TotalFunctions)
	}
	if stats.TotalCallRelations != 4 { // Main->Helper1, Main->Helper2, Helper1->Utility, Helper2->Utility
		t.Errorf("Expected 4 total call relations, got %d", stats.TotalCallRelations)
	}
	if stats.MaxCallDepth != 2 { // Main -> Helper1 -> Utility
		t.Errorf("Expected max call depth of 2, got %d", stats.MaxCallDepth)
	}
}

func TestGlobalCallGraph_EmptyInput(t *testing.T) {
	callGraph := NewGlobalCallGraph()

	// Test with empty file contexts
	err := callGraph.BuildFromFiles([]models.FileContext{})
	if err != nil {
		t.Errorf("Expected no error with empty input, got: %v", err)
	}

	// Test queries on empty call graph
	callers := callGraph.GetCallers("NonExistent")
	if len(callers) != 0 {
		t.Errorf("Expected 0 callers for non-existent function, got %d", len(callers))
	}

	callees := callGraph.GetCallees("NonExistent")
	if len(callees) != 0 {
		t.Errorf("Expected 0 callees for non-existent function, got %d", len(callees))
	}

	allFunctions := callGraph.GetAllFunctions()
	if len(allFunctions) != 0 {
		t.Errorf("Expected 0 functions in empty call graph, got %d", len(allFunctions))
	}
}

func TestGlobalCallGraph_CyclicCalls(t *testing.T) {
	// Test handling of cyclic function calls
	fileContexts := []models.FileContext{
		{
			Path:     "cyclic.go",
			Language: "go",
			Functions: []models.Function{
				{Name: "FuncA", Calls: []string{"FuncB"}},
				{Name: "FuncB", Calls: []string{"FuncC"}},
				{Name: "FuncC", Calls: []string{"FuncA"}}, // Creates a cycle
			},
		},
	}

	callGraph := NewGlobalCallGraph()
	err := callGraph.BuildFromFiles(fileContexts)
	if err != nil {
		t.Fatalf("Failed to build call graph with cycles: %v", err)
	}

	// Test: Verify cycle detection doesn't cause infinite loops
	chainDepth := callGraph.GetCallChainDepth("FuncA", "FuncA")
	if chainDepth == -1 {
		t.Error("Expected to detect cycle, but got -1 (not found)")
	}

	// Test: All functions should still be accessible
	allFunctions := callGraph.GetAllFunctions()
	if len(allFunctions) != 3 {
		t.Errorf("Expected 3 functions in cyclic call graph, got %d", len(allFunctions))
	}
}

// Helper function to create mock file context for testing
func createMockFileContext(filePath string) models.FileContext {
	// This is a simplified mock - in real implementation, this would come from the parser
	fileContext := models.FileContext{
		Path:      filePath,
		Language:  "go",
		Functions: []models.Function{},
	}

	// Simple parsing logic for test purposes
	if filepath.Base(filePath) == "main.go" {
		fileContext.Functions = []models.Function{
			{Name: "main", Calls: []string{"CreateUser", "ProcessUser", "fmt.Println"}},
			{Name: "ProcessUser", Calls: []string{"ValidateUser", "SaveUser"}},
		}
	} else if filepath.Base(filePath) == "user.go" {
		fileContext.Functions = []models.Function{
			{Name: "CreateUser", Calls: []string{}},
			{Name: "ValidateUser", Calls: []string{}},
			{Name: "SaveUser", Calls: []string{}},
		}
	} else if filepath.Base(filePath) == "utils.go" {
		fileContext.Functions = []models.Function{
			{Name: "Helper", Calls: []string{"ProcessUser"}},
		}
	}

	return fileContext
}
