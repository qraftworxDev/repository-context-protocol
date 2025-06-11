package index

import (
	"testing"
	"time"

	"repository-context-protocol/internal/models"
)

func TestFileBasedChunking_CreateChunks(t *testing.T) {
	chunker := &FileBasedChunking{
		MaxTokens: 2000,
	}

	// Create test file contexts
	file1 := models.FileContext{
		Path:     "main.go",
		Language: "go",
		Checksum: "abc123",
		ModTime:  time.Now(),
		Functions: []models.Function{
			{Name: "main", Signature: "func main()"},
		},
	}

	file2 := models.FileContext{
		Path:     "utils.go",
		Language: "go",
		Checksum: "def456",
		ModTime:  time.Now(),
		Functions: []models.Function{
			{Name: "Helper", Signature: "func Helper() string"},
		},
	}

	files := []models.FileContext{file1, file2}

	// Create chunks
	chunks := chunker.CreateChunks(files)

	// Should create one chunk per file
	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(chunks))
	}

	// Check first chunk
	chunk1 := chunks[0]
	if len(chunk1.Files) != 1 {
		t.Errorf("Expected 1 file in first chunk, got %d", len(chunk1.Files))
	}
	if chunk1.Files[0] != "main.go" {
		t.Errorf("Expected first chunk to contain main.go, got %s", chunk1.Files[0])
	}
	if len(chunk1.FileData) != 1 {
		t.Errorf("Expected 1 file data in first chunk, got %d", len(chunk1.FileData))
	}
	if chunk1.FileData[0].Path != "main.go" {
		t.Errorf("Expected file data path main.go, got %s", chunk1.FileData[0].Path)
	}

	// Check second chunk
	chunk2 := chunks[1]
	if len(chunk2.Files) != 1 {
		t.Errorf("Expected 1 file in second chunk, got %d", len(chunk2.Files))
	}
	if chunk2.Files[0] != "utils.go" {
		t.Errorf("Expected second chunk to contain utils.go, got %s", chunk2.Files[0])
	}

	// Check chunk IDs are unique
	if chunk1.ID == chunk2.ID {
		t.Error("Chunk IDs should be unique")
	}

	// Check chunk IDs are not empty
	if chunk1.ID == "" || chunk2.ID == "" {
		t.Error("Chunk IDs should not be empty")
	}
}

func TestFileBasedChunking_EmptyFiles(t *testing.T) {
	chunker := &FileBasedChunking{
		MaxTokens: 2000,
	}

	files := []models.FileContext{}
	chunks := chunker.CreateChunks(files)

	if len(chunks) != 0 {
		t.Errorf("Expected 0 chunks for empty files, got %d", len(chunks))
	}
}

func TestFileBasedChunking_SingleFile(t *testing.T) {
	chunker := &FileBasedChunking{
		MaxTokens: 1000,
	}

	file := models.FileContext{
		Path:     "single.go",
		Language: "go",
		Checksum: "single123",
		Functions: []models.Function{
			{Name: "SingleFunc", Signature: "func SingleFunc()"},
		},
		Types: []models.TypeDef{
			{Name: "SingleType", Kind: "struct"},
		},
	}

	files := []models.FileContext{file}
	chunks := chunker.CreateChunks(files)

	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk, got %d", len(chunks))
	}

	chunk := chunks[0]
	if chunk.Files[0] != "single.go" {
		t.Errorf("Expected chunk to contain single.go, got %s", chunk.Files[0])
	}

	// Check that file data is preserved
	if len(chunk.FileData[0].Functions) != 1 {
		t.Errorf("Expected 1 function in chunk data, got %d", len(chunk.FileData[0].Functions))
	}
	if chunk.FileData[0].Functions[0].Name != "SingleFunc" {
		t.Errorf("Expected function name SingleFunc, got %s", chunk.FileData[0].Functions[0].Name)
	}
}

func TestEstimateTokens_BasicCounting(t *testing.T) {
	file := models.FileContext{
		Path:     "test.go",
		Language: "go",
		Functions: []models.Function{
			{Name: "TestFunc", Signature: "func TestFunc() error"},
		},
		Types: []models.TypeDef{
			{Name: "TestType", Kind: "struct"},
		},
		Variables: []models.Variable{
			{Name: "testVar", Type: "string"},
		},
		Constants: []models.Constant{
			{Name: "TestConst", Type: "int"},
		},
	}

	tokens := estimateTokens(&file)

	// Should be greater than 0
	if tokens <= 0 {
		t.Errorf("Expected positive token count, got %d", tokens)
	}

	// Should account for functions, types, variables, constants
	// Basic estimation: each entity contributes some tokens
	expectedMinTokens := 4 // At least one token per entity type
	if tokens < expectedMinTokens {
		t.Errorf("Expected at least %d tokens, got %d", expectedMinTokens, tokens)
	}
}

func TestEstimateTokens_EmptyFile(t *testing.T) {
	file := models.FileContext{
		Path:     "empty.go",
		Language: "go",
	}

	tokens := estimateTokens(&file)

	// Empty file should have minimal tokens (just for basic structure)
	if tokens < 0 {
		t.Errorf("Expected non-negative token count, got %d", tokens)
	}
}

func TestGenerateChunkID_Uniqueness(t *testing.T) {
	id1 := generateChunkID("main.go")
	id2 := generateChunkID("utils.go")
	id3 := generateChunkID("main.go") // Same file

	// Different files should have different IDs
	if id1 == id2 {
		t.Error("Different files should generate different chunk IDs")
	}

	// Same file should generate same ID (deterministic)
	if id1 != id3 {
		t.Error("Same file should generate same chunk ID")
	}

	// IDs should not be empty
	if id1 == "" || id2 == "" {
		t.Error("Chunk IDs should not be empty")
	}
}

func TestGenerateChunkID_Format(t *testing.T) {
	id := generateChunkID("test/path/file.go")

	// Should be a reasonable length (not too short, not too long)
	if len(id) < 8 {
		t.Errorf("Chunk ID seems too short: %s", id)
	}
	if len(id) > 64 {
		t.Errorf("Chunk ID seems too long: %s", id)
	}

	// Should not contain problematic characters for filenames
	problematicChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range problematicChars {
		if contains(id, char) {
			t.Errorf("Chunk ID contains problematic character %s: %s", char, id)
		}
	}
}

func TestChunkingStrategy_Interface(t *testing.T) {
	// Test that FileBasedChunking implements ChunkingStrategy interface
	var strategy ChunkingStrategy = &FileBasedChunking{MaxTokens: 1000}

	file := models.FileContext{
		Path:     "interface_test.go",
		Language: "go",
	}

	chunks := strategy.CreateChunks([]models.FileContext{file})

	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk from interface call, got %d", len(chunks))
	}
}

// Helper function for testing
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
