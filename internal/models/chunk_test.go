package models

import (
	"testing"
	"time"
)

func TestSemanticChunk_Creation(t *testing.T) {
	now := time.Now()
	chunk := SemanticChunk{
		ID:         "chunk_001",
		Files:      []string{"main.go", "utils.go"},
		FileData:   []FileContext{},
		TokenCount: 1500,
		CreatedAt:  now,
	}

	if chunk.ID != "chunk_001" {
		t.Errorf("Expected ID 'chunk_001', got %s", chunk.ID)
	}
	if len(chunk.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(chunk.Files))
	}
	if chunk.Files[0] != "main.go" {
		t.Errorf("Expected first file 'main.go', got %s", chunk.Files[0])
	}
	if chunk.Files[1] != "utils.go" {
		t.Errorf("Expected second file 'utils.go', got %s", chunk.Files[1])
	}
	if chunk.TokenCount != 1500 {
		t.Errorf("Expected token count 1500, got %d", chunk.TokenCount)
	}
	if !chunk.CreatedAt.Equal(now) {
		t.Errorf("Expected created at %v, got %v", now, chunk.CreatedAt)
	}
}

func TestSemanticChunk_EmptyChunk(t *testing.T) {
	chunk := SemanticChunk{}

	if chunk.ID != "" {
		t.Errorf("Expected empty ID, got %s", chunk.ID)
	}
	if len(chunk.Files) != 0 {
		t.Errorf("Expected 0 files, got %d", len(chunk.Files))
	}
	if chunk.TokenCount != 0 {
		t.Errorf("Expected token count 0, got %d", chunk.TokenCount)
	}
	if !chunk.CreatedAt.IsZero() {
		t.Errorf("Expected zero time, got %v", chunk.CreatedAt)
	}
}

func TestManifest_Creation(t *testing.T) {
	now := time.Now()
	manifest := Manifest{
		Version:   "1.0.0",
		Chunks:    make(map[string]ChunkInfo),
		UpdatedAt: now,
	}

	if manifest.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", manifest.Version)
	}
	if manifest.Chunks == nil {
		t.Error("Expected chunks map to be initialized")
	}
	if len(manifest.Chunks) != 0 {
		t.Errorf("Expected 0 chunks, got %d", len(manifest.Chunks))
	}
	if !manifest.UpdatedAt.Equal(now) {
		t.Errorf("Expected updated at %v, got %v", now, manifest.UpdatedAt)
	}
}

func TestChunkInfo_Creation(t *testing.T) {
	now := time.Now()
	chunkInfo := ChunkInfo{
		Files:      []string{"main.go"},
		Size:       2048,
		TokenCount: 800,
		UpdatedAt:  now,
	}

	if len(chunkInfo.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(chunkInfo.Files))
	}
	if chunkInfo.Files[0] != "main.go" {
		t.Errorf("Expected file 'main.go', got %s", chunkInfo.Files[0])
	}
	if chunkInfo.Size != 2048 {
		t.Errorf("Expected size 2048, got %d", chunkInfo.Size)
	}
	if chunkInfo.TokenCount != 800 {
		t.Errorf("Expected token count 800, got %d", chunkInfo.TokenCount)
	}
	if !chunkInfo.UpdatedAt.Equal(now) {
		t.Errorf("Expected updated at %v, got %v", now, chunkInfo.UpdatedAt)
	}
}

func TestManifest_AddChunk(t *testing.T) {
	manifest := Manifest{
		Version: "1.0.0",
		Chunks:  make(map[string]ChunkInfo),
	}

	chunkInfo := ChunkInfo{
		Files:      []string{"test.go"},
		Size:       1024,
		TokenCount: 400,
		UpdatedAt:  time.Now(),
	}

	// Add chunk to manifest
	manifest.Chunks["chunk_001"] = chunkInfo

	if len(manifest.Chunks) != 1 {
		t.Errorf("Expected 1 chunk in manifest, got %d", len(manifest.Chunks))
	}

	retrievedChunk, exists := manifest.Chunks["chunk_001"]
	if !exists {
		t.Error("Expected chunk_001 to exist in manifest")
	}

	if retrievedChunk.Size != 1024 {
		t.Errorf("Expected chunk size 1024, got %d", retrievedChunk.Size)
	}
}

func TestSemanticChunk_WithFileData(t *testing.T) {
	fileContext := FileContext{
		Path:     "test.go",
		Language: "go",
		Checksum: "abc123",
	}

	chunk := SemanticChunk{
		ID:         "chunk_002",
		Files:      []string{"test.go"},
		FileData:   []FileContext{fileContext},
		TokenCount: 500,
		CreatedAt:  time.Now(),
	}

	if len(chunk.FileData) != 1 {
		t.Errorf("Expected 1 file data entry, got %d", len(chunk.FileData))
	}

	if chunk.FileData[0].Path != "test.go" {
		t.Errorf("Expected file path 'test.go', got %s", chunk.FileData[0].Path)
	}

	if chunk.FileData[0].Language != "go" {
		t.Errorf("Expected language 'go', got %s", chunk.FileData[0].Language)
	}
}
