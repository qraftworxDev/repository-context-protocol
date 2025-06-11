package index

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"repository-context-protocol/internal/models"
)

func TestChunkSerializer_SaveAndLoadChunk(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "chunk_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	serializer := &ChunkSerializer{
		baseDir: tempDir,
	}

	// Create test chunk
	now := time.Now()
	chunk := models.SemanticChunk{
		ID:    "test_chunk_001",
		Files: []string{"main.go", "utils.go"},
		FileData: []models.FileContext{
			{
				Path:     "main.go",
				Language: "go",
				Checksum: "abc123",
				ModTime:  now,
				Functions: []models.Function{
					{Name: "main", Signature: "func main()"},
				},
			},
			{
				Path:     "utils.go",
				Language: "go",
				Checksum: "def456",
				ModTime:  now,
				Functions: []models.Function{
					{Name: "Helper", Signature: "func Helper() string"},
				},
			},
		},
		TokenCount: 150,
		CreatedAt:  now,
	}

	// Save chunk
	err = serializer.SaveChunk(&chunk)
	if err != nil {
		t.Fatalf("Failed to save chunk: %v", err)
	}

	// Verify file was created
	expectedPath := filepath.Join(tempDir, "test_chunk_001.msgpack")
	if _, statErr := os.Stat(expectedPath); os.IsNotExist(statErr) {
		t.Errorf("Expected chunk file to be created at %s", expectedPath)
	}

	// Load chunk back
	loadedChunk, err := serializer.LoadChunk("test_chunk_001")
	if err != nil {
		t.Fatalf("Failed to load chunk: %v", err)
	}

	// Verify loaded chunk matches original
	if loadedChunk.ID != chunk.ID {
		t.Errorf("Expected ID %s, got %s", chunk.ID, loadedChunk.ID)
	}
	if len(loadedChunk.Files) != len(chunk.Files) {
		t.Errorf("Expected %d files, got %d", len(chunk.Files), len(loadedChunk.Files))
	}
	if loadedChunk.Files[0] != chunk.Files[0] {
		t.Errorf("Expected first file %s, got %s", chunk.Files[0], loadedChunk.Files[0])
	}
	if len(loadedChunk.FileData) != len(chunk.FileData) {
		t.Errorf("Expected %d file data entries, got %d", len(chunk.FileData), len(loadedChunk.FileData))
	}
	if loadedChunk.FileData[0].Path != chunk.FileData[0].Path {
		t.Errorf("Expected file path %s, got %s", chunk.FileData[0].Path, loadedChunk.FileData[0].Path)
	}
	if loadedChunk.TokenCount != chunk.TokenCount {
		t.Errorf("Expected token count %d, got %d", chunk.TokenCount, loadedChunk.TokenCount)
	}
}

func TestChunkSerializer_SaveChunk_CreatesDirectory(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "chunk_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Use a subdirectory that doesn't exist yet
	chunkDir := filepath.Join(tempDir, "chunks")
	serializer := &ChunkSerializer{
		baseDir: chunkDir,
	}

	chunk := models.SemanticChunk{
		ID:         "test_chunk_002",
		Files:      []string{"test.go"},
		FileData:   []models.FileContext{{Path: "test.go", Language: "go"}},
		TokenCount: 50,
		CreatedAt:  time.Now(),
	}

	// Save chunk - should create directory
	err = serializer.SaveChunk(&chunk)
	if err != nil {
		t.Fatalf("Failed to save chunk: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(chunkDir); os.IsNotExist(err) {
		t.Errorf("Expected chunk directory to be created at %s", chunkDir)
	}

	// Verify file was created
	expectedPath := filepath.Join(chunkDir, "test_chunk_002.msgpack")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected chunk file to be created at %s", expectedPath)
	}
}

func TestChunkSerializer_LoadChunk_NotFound(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "chunk_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	serializer := &ChunkSerializer{
		baseDir: tempDir,
	}

	// Try to load non-existent chunk
	_, err = serializer.LoadChunk("nonexistent_chunk")
	if err == nil {
		t.Error("Expected error when loading non-existent chunk")
	}
}

func TestChunkSerializer_EmptyChunk(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "chunk_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	serializer := &ChunkSerializer{
		baseDir: tempDir,
	}

	// Create empty chunk
	chunk := models.SemanticChunk{
		ID:         "empty_chunk",
		Files:      []string{},
		FileData:   []models.FileContext{},
		TokenCount: 0,
		CreatedAt:  time.Now(),
	}

	// Save and load empty chunk
	err = serializer.SaveChunk(&chunk)
	if err != nil {
		t.Fatalf("Failed to save empty chunk: %v", err)
	}

	loadedChunk, err := serializer.LoadChunk("empty_chunk")
	if err != nil {
		t.Fatalf("Failed to load empty chunk: %v", err)
	}

	// Verify empty chunk properties
	if loadedChunk.ID != chunk.ID {
		t.Errorf("Expected ID %s, got %s", chunk.ID, loadedChunk.ID)
	}
	if len(loadedChunk.Files) != 0 {
		t.Errorf("Expected 0 files, got %d", len(loadedChunk.Files))
	}
	if len(loadedChunk.FileData) != 0 {
		t.Errorf("Expected 0 file data entries, got %d", len(loadedChunk.FileData))
	}
	if loadedChunk.TokenCount != 0 {
		t.Errorf("Expected token count 0, got %d", loadedChunk.TokenCount)
	}
}

func TestChunkSerializer_ComplexFileData(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "chunk_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	serializer := &ChunkSerializer{
		baseDir: tempDir,
	}

	// Create chunk with complex file data
	chunk := models.SemanticChunk{
		ID:    "complex_chunk",
		Files: []string{"complex.go"},
		FileData: []models.FileContext{
			{
				Path:     "complex.go",
				Language: "go",
				Checksum: "complex123",
				Functions: []models.Function{
					{
						Name:      "ComplexFunc",
						Signature: "func ComplexFunc(param string) (string, error)",
						Parameters: []models.Parameter{
							{Name: "param", Type: "string"},
						},
						Returns: []models.Type{
							{Name: "string", Kind: "basic"},
							{Name: "error", Kind: "interface"},
						},
						StartLine: 10,
						EndLine:   20,
						Calls:     []string{"fmt.Sprintf", "errors.New"},
					},
				},
				Types: []models.TypeDef{
					{
						Name: "ComplexType",
						Kind: "struct",
						Fields: []models.Field{
							{Name: "ID", Type: "int", Tag: "`json:\"id\"`"},
							{Name: "Name", Type: "string", Tag: "`json:\"name\"`"},
						},
						StartLine: 5,
						EndLine:   8,
					},
				},
				Variables: []models.Variable{
					{Name: "globalVar", Type: "string", StartLine: 3, EndLine: 3},
				},
				Constants: []models.Constant{
					{Name: "MaxSize", Type: "int", StartLine: 1, EndLine: 1},
				},
				Imports: []models.Import{
					{Path: "fmt"},
					{Path: "errors"},
				},
				Exports: []models.Export{
					{Name: "ComplexFunc", Type: "function", Kind: "function"},
					{Name: "ComplexType", Type: "struct", Kind: "type"},
				},
			},
		},
		TokenCount: 500,
		CreatedAt:  time.Now(),
	}

	// Save and load complex chunk
	err = serializer.SaveChunk(&chunk)
	if err != nil {
		t.Fatalf("Failed to save complex chunk: %v", err)
	}

	loadedChunk, err := serializer.LoadChunk("complex_chunk")
	if err != nil {
		t.Fatalf("Failed to load complex chunk: %v", err)
	}

	// Verify complex data is preserved
	fileData := loadedChunk.FileData[0]

	// Check functions
	if len(fileData.Functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(fileData.Functions))
	}
	if fileData.Functions[0].Name != "ComplexFunc" {
		t.Errorf("Expected function name ComplexFunc, got %s", fileData.Functions[0].Name)
	}
	if len(fileData.Functions[0].Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(fileData.Functions[0].Parameters))
	}
	if len(fileData.Functions[0].Returns) != 2 {
		t.Errorf("Expected 2 returns, got %d", len(fileData.Functions[0].Returns))
	}

	// Check types
	if len(fileData.Types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(fileData.Types))
	}
	if fileData.Types[0].Name != "ComplexType" {
		t.Errorf("Expected type name ComplexType, got %s", fileData.Types[0].Name)
	}
	if len(fileData.Types[0].Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(fileData.Types[0].Fields))
	}

	// Check variables and constants
	if len(fileData.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(fileData.Variables))
	}
	if len(fileData.Constants) != 1 {
		t.Errorf("Expected 1 constant, got %d", len(fileData.Constants))
	}

	// Check imports and exports
	if len(fileData.Imports) != 2 {
		t.Errorf("Expected 2 imports, got %d", len(fileData.Imports))
	}
	if len(fileData.Exports) != 2 {
		t.Errorf("Expected 2 exports, got %d", len(fileData.Exports))
	}
}

func TestChunkSerializer_GetChunkPath(t *testing.T) {
	serializer := &ChunkSerializer{
		baseDir: "/test/chunks",
	}

	path := serializer.GetChunkPath("test_chunk_123")
	expected := "/test/chunks/test_chunk_123.msgpack"

	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}

func TestChunkSerializer_InvalidBaseDir(t *testing.T) {
	// Use an invalid base directory (read-only root on most systems)
	serializer := &ChunkSerializer{
		baseDir: "/invalid/readonly/path",
	}

	chunk := models.SemanticChunk{
		ID:         "test_chunk",
		Files:      []string{"test.go"},
		FileData:   []models.FileContext{{Path: "test.go"}},
		TokenCount: 10,
		CreatedAt:  time.Now(),
	}

	// Should fail to save to invalid directory
	err := serializer.SaveChunk(&chunk)
	if err == nil {
		t.Error("Expected error when saving to invalid directory")
	}
}
