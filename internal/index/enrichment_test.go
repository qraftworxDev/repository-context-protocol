package index

import (
	"testing"
	"time"

	"repository-context-protocol/internal/models"
)

func TestGlobalEnrichment_EnrichFileContexts(t *testing.T) {
	// Create test file contexts representing a multi-file scenario
	fileContexts := []models.FileContext{
		{
			Path:     "main.go",
			Language: "go",
			Checksum: "main123",
			ModTime:  time.Now(),
			Functions: []models.Function{
				{
					Name:      "main",
					Signature: "func main()",
					StartLine: 5,
					EndLine:   10,
					Calls:     []string{"CreateUser", "ProcessUser"},
					CalledBy:  []string{},
				},
			},
		},
		{
			Path:     "user.go",
			Language: "go",
			Checksum: "user123",
			ModTime:  time.Now(),
			Functions: []models.Function{
				{
					Name:      "CreateUser",
					Signature: "func CreateUser(name string) *User",
					StartLine: 5,
					EndLine:   8,
					Calls:     []string{},
					CalledBy:  []string{"main"},
				},
				{
					Name:      "ProcessUser",
					Signature: "func ProcessUser(user *User)",
					StartLine: 15,
					EndLine:   20,
					Calls:     []string{"ValidateUser"},
					CalledBy:  []string{"main"},
				},
				{
					Name:      "ValidateUser",
					Signature: "func ValidateUser(user *User) bool",
					StartLine: 25,
					EndLine:   30,
					Calls:     []string{},
					CalledBy:  []string{"ProcessUser"},
				},
			},
		},
	}

	// Create enrichment processor
	enrichment := NewGlobalEnrichment()

	// Enrich the file contexts
	enrichedContexts, err := enrichment.EnrichFileContexts(fileContexts)
	if err != nil {
		t.Fatalf("Failed to enrich file contexts: %v", err)
	}

	if len(enrichedContexts) != 2 {
		t.Errorf("Expected 2 enriched contexts, got %d", len(enrichedContexts))
	}

	// Test main function enrichment
	mainFile := findFileContext(enrichedContexts, "main.go")
	if mainFile == nil {
		t.Fatal("Could not find main.go in enriched contexts")
	}

	mainFunc := findFunction(mainFile.Functions, "main")
	if mainFunc == nil {
		t.Fatal("Could not find main function")
	}

	// Main should have no local calls (all calls are cross-file)
	if len(mainFunc.LocalCalls) != 0 {
		t.Errorf("Expected 0 local calls in main, got %d", len(mainFunc.LocalCalls))
	}

	// Main should have 2 cross-file calls
	if len(mainFunc.CrossFileCalls) != 2 {
		t.Errorf("Expected 2 cross-file calls in main, got %d", len(mainFunc.CrossFileCalls))
	}

	// Verify cross-file call details
	createUserCall := findCallReference(mainFunc.CrossFileCalls, "CreateUser")
	if createUserCall == nil {
		t.Error("Expected main to have cross-file call to CreateUser")
	} else if createUserCall.File != "user.go" {
		t.Errorf("Expected CreateUser call to reference user.go, got %s", createUserCall.File)
	}

	// Test ProcessUser function enrichment
	userFile := findFileContext(enrichedContexts, "user.go")
	if userFile == nil {
		t.Fatal("Could not find user.go in enriched contexts")
	}

	processUserFunc := findFunction(userFile.Functions, "ProcessUser")
	if processUserFunc == nil {
		t.Fatal("Could not find ProcessUser function")
	}

	// ProcessUser should have 1 local call (ValidateUser)
	if len(processUserFunc.LocalCalls) != 1 {
		t.Errorf("Expected 1 local call in ProcessUser, got %d", len(processUserFunc.LocalCalls))
	}
	if processUserFunc.LocalCalls[0] != "ValidateUser" {
		t.Errorf("Expected local call to ValidateUser, got %s", processUserFunc.LocalCalls[0])
	}

	// ProcessUser should have 1 cross-file caller (main)
	if len(processUserFunc.CrossFileCallers) != 1 {
		t.Errorf("Expected 1 cross-file caller for ProcessUser, got %d", len(processUserFunc.CrossFileCallers))
	}

	mainCaller := processUserFunc.CrossFileCallers[0]
	if mainCaller.FunctionName != "main" {
		t.Errorf("Expected cross-file caller 'main', got %s", mainCaller.FunctionName)
	}
	if mainCaller.File != "main.go" {
		t.Errorf("Expected cross-file caller from main.go, got %s", mainCaller.File)
	}
}

func TestGlobalEnrichment_LocalCallsOnly(t *testing.T) {
	// Test scenario with only local calls within a single file
	fileContexts := []models.FileContext{
		{
			Path:     "utils.go",
			Language: "go",
			Checksum: "utils123",
			ModTime:  time.Now(),
			Functions: []models.Function{
				{
					Name:      "PublicFunction",
					Signature: "func PublicFunction()",
					StartLine: 5,
					EndLine:   10,
					Calls:     []string{"privateHelper"},
					CalledBy:  []string{},
				},
				{
					Name:      "privateHelper",
					Signature: "func privateHelper()",
					StartLine: 15,
					EndLine:   20,
					Calls:     []string{},
					CalledBy:  []string{"PublicFunction"},
				},
			},
		},
	}

	enrichment := NewGlobalEnrichment()
	enrichedContexts, err := enrichment.EnrichFileContexts(fileContexts)
	if err != nil {
		t.Fatalf("Failed to enrich file contexts: %v", err)
	}

	utilsFile := enrichedContexts[0]
	publicFunc := findFunction(utilsFile.Functions, "PublicFunction")
	if publicFunc == nil {
		t.Fatal("Could not find PublicFunction")
	}

	// Should have 1 local call, 0 cross-file calls
	if len(publicFunc.LocalCalls) != 1 {
		t.Errorf("Expected 1 local call, got %d", len(publicFunc.LocalCalls))
	}
	if len(publicFunc.CrossFileCalls) != 0 {
		t.Errorf("Expected 0 cross-file calls, got %d", len(publicFunc.CrossFileCalls))
	}
	if publicFunc.LocalCalls[0] != "privateHelper" {
		t.Errorf("Expected local call to privateHelper, got %s", publicFunc.LocalCalls[0])
	}

	// Test privateHelper
	privateFunc := findFunction(utilsFile.Functions, "privateHelper")
	if privateFunc == nil {
		t.Fatal("Could not find privateHelper")
	}

	// Should have 1 local caller, 0 cross-file callers
	if len(privateFunc.LocalCallers) != 1 {
		t.Errorf("Expected 1 local caller, got %d", len(privateFunc.LocalCallers))
	}
	if len(privateFunc.CrossFileCallers) != 0 {
		t.Errorf("Expected 0 cross-file callers, got %d", len(privateFunc.CrossFileCallers))
	}
	if privateFunc.LocalCallers[0] != "PublicFunction" {
		t.Errorf("Expected local caller PublicFunction, got %s", privateFunc.LocalCallers[0])
	}
}

func TestGlobalEnrichment_EmptyInput(t *testing.T) {
	enrichment := NewGlobalEnrichment()

	// Test with empty file contexts
	enrichedContexts, err := enrichment.EnrichFileContexts([]models.FileContext{})
	if err != nil {
		t.Errorf("Expected no error with empty input, got: %v", err)
	}
	if len(enrichedContexts) != 0 {
		t.Errorf("Expected 0 enriched contexts, got %d", len(enrichedContexts))
	}
}

func TestGlobalEnrichment_GetGlobalCallGraph(t *testing.T) {
	enrichment := NewGlobalEnrichment()

	// Verify we can access the underlying global call graph
	callGraph := enrichment.GetGlobalCallGraph()
	if callGraph == nil {
		t.Error("Expected global call graph to be accessible")
	}

	// Test that it's functional
	allFunctions := callGraph.GetAllFunctions()
	if allFunctions == nil {
		t.Error("Expected global call graph to be functional")
	}
}

// Helper functions for tests

func findFileContext(contexts []models.FileContext, path string) *models.FileContext {
	for i := range contexts {
		if contexts[i].Path == path {
			return &contexts[i]
		}
	}
	return nil
}

func findFunction(functions []models.Function, name string) *models.Function {
	for i := range functions {
		if functions[i].Name == name {
			return &functions[i]
		}
	}
	return nil
}

func findCallReference(calls []models.CallReference, functionName string) *models.CallReference {
	for i := range calls {
		if calls[i].FunctionName == functionName {
			return &calls[i]
		}
	}
	return nil
}
