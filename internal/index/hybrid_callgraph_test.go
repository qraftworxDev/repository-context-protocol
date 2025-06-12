package index

import (
	"os"
	"testing"
	"time"

	"repository-context-protocol/internal/models"
)

func TestHybridStorage_QueryCallGraph(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hybrid_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := NewHybridStorage(tempDir)

	// Initialize hybrid storage
	err = storage.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize hybrid storage: %v", err)
	}
	defer storage.Close()

	// Store file contexts with call relationships
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
					StartLine: 1,
					EndLine:   10,
					Calls:     []string{"helper", "utils.Process"},
				},
				{
					Name:      "helper",
					Signature: "func helper() string",
					StartLine: 15,
					EndLine:   20,
					Calls:     []string{"fmt.Sprintf"},
				},
			},
		},
		{
			Path:     "utils.go",
			Language: "go",
			Checksum: "utils123",
			ModTime:  time.Now(),
			Functions: []models.Function{
				{
					Name:      "Process",
					Signature: "func Process(data string) error",
					StartLine: 5,
					EndLine:   15,
					Calls:     []string{"validate", "transform"},
				},
				{
					Name:      "validate",
					Signature: "func validate(s string) bool",
					StartLine: 20,
					EndLine:   25,
				},
				{
					Name:      "transform",
					Signature: "func transform(s string) string",
					StartLine: 30,
					EndLine:   35,
				},
			},
		},
	}

	// Store all file contexts
	for i := range fileContexts {
		err = storage.StoreFileContext(&fileContexts[i])
		if err != nil {
			t.Fatalf("Failed to store file context: %v", err)
		}
	}

	// Test QueryCallsFrom
	callsFromMain, err := storage.QueryCallsFrom("main")
	if err != nil {
		t.Fatalf("Failed to query calls from main: %v", err)
	}
	if len(callsFromMain) != 2 {
		t.Errorf("Expected 2 calls from main, got %d", len(callsFromMain))
	}

	// Verify the calls include helper and utils.Process
	callTargets := make(map[string]bool)
	for _, call := range callsFromMain {
		callTargets[call.Callee] = true
	}
	if !callTargets["helper"] {
		t.Error("Expected main to call helper")
	}
	if !callTargets["utils.Process"] {
		t.Error("Expected main to call utils.Process")
	}

	// Test QueryCallsTo
	callsToHelper, err := storage.QueryCallsTo("helper")
	if err != nil {
		t.Fatalf("Failed to query calls to helper: %v", err)
	}
	if len(callsToHelper) != 1 {
		t.Errorf("Expected 1 call to helper, got %d", len(callsToHelper))
	}
	if callsToHelper[0].Caller != "main" {
		t.Errorf("Expected call to helper from main, got from %s", callsToHelper[0].Caller)
	}
}

func TestHybridStorage_QueryCallGraphWithChunkData(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hybrid_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := NewHybridStorage(tempDir)

	// Initialize hybrid storage
	err = storage.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize hybrid storage: %v", err)
	}
	defer storage.Close()

	// Store file context with function calls
	fileContext := models.FileContext{
		Path:     "service.go",
		Language: "go",
		Checksum: "service123",
		ModTime:  time.Now(),
		Functions: []models.Function{
			{
				Name:      "ProcessData",
				Signature: "func ProcessData(input string) (string, error)",
				StartLine: 10,
				EndLine:   25,
				Calls:     []string{"validateInput", "transformData"},
			},
			{
				Name:      "validateInput",
				Signature: "func validateInput(s string) error",
				StartLine: 30,
				EndLine:   35,
			},
			{
				Name:      "transformData",
				Signature: "func transformData(s string) string",
				StartLine: 40,
				EndLine:   45,
			},
		},
	}

	err = storage.StoreFileContext(&fileContext)
	if err != nil {
		t.Fatalf("Failed to store file context: %v", err)
	}

	// Query calls with chunk data
	results, err := storage.QueryCallsFromWithChunkData("ProcessData")
	if err != nil {
		t.Fatalf("Failed to query calls with chunk data: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 call results, got %d", len(results))
	}

	// Verify each result has both call relation and chunk data
	for _, result := range results {
		if result.CallRelation.Caller != "ProcessData" {
			t.Errorf("Expected call from ProcessData, got from %s", result.CallRelation.Caller)
		}
		if result.ChunkData == nil {
			t.Error("Expected chunk data to be loaded")
		}
		if len(result.ChunkData.FileData) != 1 {
			t.Errorf("Expected 1 file in chunk data, got %d", len(result.ChunkData.FileData))
		}
	}
}

func TestHybridStorage_UpdateCallGraph(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hybrid_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := NewHybridStorage(tempDir)

	// Initialize hybrid storage
	err = storage.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize hybrid storage: %v", err)
	}
	defer storage.Close()

	// Store initial file context
	initialContext := models.FileContext{
		Path:     "app.go",
		Language: "go",
		Checksum: "app123",
		ModTime:  time.Now(),
		Functions: []models.Function{
			{
				Name:      "Start",
				Signature: "func Start()",
				StartLine: 1,
				EndLine:   10,
				Calls:     []string{"initialize"},
			},
		},
	}

	err = storage.StoreFileContext(&initialContext)
	if err != nil {
		t.Fatalf("Failed to store initial context: %v", err)
	}

	// Verify initial call graph
	initialCalls, err := storage.QueryCallsFrom("Start")
	if err != nil {
		t.Fatalf("Failed to query initial calls: %v", err)
	}
	if len(initialCalls) != 1 {
		t.Errorf("Expected 1 initial call, got %d", len(initialCalls))
	}

	// Update file context with different calls
	updatedContext := models.FileContext{
		Path:     "app.go",
		Language: "go",
		Checksum: "app456", // Different checksum
		ModTime:  time.Now(),
		Functions: []models.Function{
			{
				Name:      "Start",
				Signature: "func Start()",
				StartLine: 1,
				EndLine:   15,
				Calls:     []string{"initialize", "configure", "run"},
			},
		},
	}

	err = storage.StoreFileContext(&updatedContext)
	if err != nil {
		t.Fatalf("Failed to store updated context: %v", err)
	}

	// Verify updated call graph
	updatedCalls, err := storage.QueryCallsFrom("Start")
	if err != nil {
		t.Fatalf("Failed to query updated calls: %v", err)
	}
	if len(updatedCalls) != 3 {
		t.Errorf("Expected 3 updated calls, got %d", len(updatedCalls))
	}

	// Verify the new calls
	callTargets := make(map[string]bool)
	for _, call := range updatedCalls {
		callTargets[call.Callee] = true
	}
	expectedCalls := []string{"initialize", "configure", "run"}
	for _, expected := range expectedCalls {
		if !callTargets[expected] {
			t.Errorf("Expected call to %s", expected)
		}
	}
}
