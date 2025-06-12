package index

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"repository-context-protocol/internal/models"
)

func TestHybridStorage_Initialize(t *testing.T) {
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

	// Verify directory structure was created
	expectedDirs := []string{
		filepath.Join(tempDir, "chunks"),
		filepath.Join(tempDir, "index.db"),
	}

	// Check chunks directory
	if _, err := os.Stat(expectedDirs[0]); os.IsNotExist(err) {
		t.Errorf("Expected chunks directory to be created at %s", expectedDirs[0])
	}

	// Check database file
	if _, err := os.Stat(expectedDirs[1]); os.IsNotExist(err) {
		t.Errorf("Expected database file to be created at %s", expectedDirs[1])
	}

	// Verify components are initialized
	if storage.sqliteIndex == nil {
		t.Error("Expected SQLite index to be initialized")
	}
	if storage.chunkSerializer == nil {
		t.Error("Expected chunk serializer to be initialized")
	}
	if storage.chunkingStrategy == nil {
		t.Error("Expected chunking strategy to be initialized")
	}
}

func TestHybridStorage_StoreFileContext(t *testing.T) {
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

	// Create test file context
	now := time.Now()
	fileContext := models.FileContext{
		Path:     "main.go",
		Language: "go",
		Checksum: "abc123",
		ModTime:  now,
		Functions: []models.Function{
			{
				Name:      "main",
				Signature: "func main()",
				StartLine: 10,
				EndLine:   15,
			},
			{
				Name:      "helper",
				Signature: "func helper() string",
				StartLine: 20,
				EndLine:   25,
				Calls:     []string{"fmt.Sprintf"},
			},
		},
		Types: []models.TypeDef{
			{
				Name:      "User",
				Kind:      "struct",
				StartLine: 5,
				EndLine:   8,
			},
		},
		Variables: []models.Variable{
			{Name: "globalVar", Type: "string", StartLine: 1, EndLine: 1},
		},
		Constants: []models.Constant{
			{Name: "MaxSize", Type: "int", StartLine: 2, EndLine: 2},
		},
	}

	// Store file context
	err = storage.StoreFileContext(&fileContext)
	if err != nil {
		t.Fatalf("Failed to store file context: %v", err)
	}

	// Verify data was stored in both SQLite and chunks
	// Check SQLite index entries
	entries, err := storage.sqliteIndex.QueryIndexEntries("main")
	if err != nil {
		t.Fatalf("Failed to query index entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 index entry for 'main', got %d", len(entries))
	}
	if entries[0].Type != "function" {
		t.Errorf("Expected type 'function', got %s", entries[0].Type)
	}

	// Check that chunk was created
	chunkIDs, err := storage.chunkSerializer.ListChunks()
	if err != nil {
		t.Fatalf("Failed to list chunks: %v", err)
	}
	if len(chunkIDs) != 1 {
		t.Errorf("Expected 1 chunk, got %d", len(chunkIDs))
	}
}

func TestHybridStorage_QueryByName(t *testing.T) {
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

	// Store test data
	fileContext := models.FileContext{
		Path:     "utils.go",
		Language: "go",
		Checksum: "def456",
		ModTime:  time.Now(),
		Functions: []models.Function{
			{
				Name:      "Helper",
				Signature: "func Helper() string",
				StartLine: 5,
				EndLine:   10,
			},
		},
	}

	err = storage.StoreFileContext(&fileContext)
	if err != nil {
		t.Fatalf("Failed to store file context: %v", err)
	}

	// Query by name
	results, err := storage.QueryByName("Helper")
	if err != nil {
		t.Fatalf("Failed to query by name: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.IndexEntry.Name != "Helper" {
		t.Errorf("Expected name 'Helper', got %s", result.IndexEntry.Name)
	}
	if result.IndexEntry.Type != "function" {
		t.Errorf("Expected type 'function', got %s", result.IndexEntry.Type)
	}
	if result.ChunkData == nil {
		t.Error("Expected chunk data to be loaded")
	}
	if len(result.ChunkData.FileData) != 1 {
		t.Errorf("Expected 1 file in chunk data, got %d", len(result.ChunkData.FileData))
	}
}

func TestHybridStorage_QueryByType(t *testing.T) {
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

	// Store multiple file contexts with different types
	fileContexts := []models.FileContext{
		{
			Path:     "main.go",
			Language: "go",
			Checksum: "abc123",
			ModTime:  time.Now(),
			Functions: []models.Function{
				{Name: "main", Signature: "func main()", StartLine: 1, EndLine: 5},
				{Name: "helper", Signature: "func helper()", StartLine: 10, EndLine: 15},
			},
		},
		{
			Path:     "types.go",
			Language: "go",
			Checksum: "def456",
			ModTime:  time.Now(),
			Types: []models.TypeDef{
				{Name: "User", Kind: "struct", StartLine: 1, EndLine: 5},
				{Name: "Config", Kind: "struct", StartLine: 10, EndLine: 15},
			},
		},
	}

	for i := range fileContexts {
		err = storage.StoreFileContext(&fileContexts[i])
		if err != nil {
			t.Fatalf("Failed to store file context: %v", err)
		}
	}

	// Query by type - functions
	functionResults, err := storage.QueryByType("function")
	if err != nil {
		t.Fatalf("Failed to query functions: %v", err)
	}
	if len(functionResults) != 2 {
		t.Errorf("Expected 2 function results, got %d", len(functionResults))
	}

	// Query by type - structs
	structResults, err := storage.QueryByType("struct")
	if err != nil {
		t.Fatalf("Failed to query structs: %v", err)
	}
	if len(structResults) != 2 {
		t.Errorf("Expected 2 struct results, got %d", len(structResults))
	}

	// Verify chunk data is loaded
	for _, result := range structResults {
		if result.ChunkData == nil {
			t.Error("Expected chunk data to be loaded")
		}
	}
}

func TestHybridStorage_DeleteFile(t *testing.T) {
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

	// Store test file
	fileContext := models.FileContext{
		Path:     "test.go",
		Language: "go",
		Checksum: "test123",
		ModTime:  time.Now(),
		Functions: []models.Function{
			{Name: "TestFunc", Signature: "func TestFunc()", StartLine: 1, EndLine: 5},
		},
	}

	err = storage.StoreFileContext(&fileContext)
	if err != nil {
		t.Fatalf("Failed to store file context: %v", err)
	}

	// Verify data exists
	results, err := storage.QueryByName("TestFunc")
	if err != nil {
		t.Fatalf("Failed to query before deletion: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result before deletion, got %d", len(results))
	}

	// Delete file
	err = storage.DeleteFile("test.go")
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify data is gone
	results, err = storage.QueryByName("TestFunc")
	if err != nil {
		t.Fatalf("Failed to query after deletion: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results after deletion, got %d", len(results))
	}
}

func TestHybridStorage_Close(t *testing.T) {
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

	// Close storage
	err = storage.Close()
	if err != nil {
		t.Errorf("Failed to close hybrid storage: %v", err)
	}

	// Verify components are properly closed
	// Attempting to use closed storage should fail gracefully
	err = storage.StoreFileContext(&models.FileContext{Path: "test.go"})
	if err == nil {
		t.Error("Expected error when using closed storage")
	}
}
