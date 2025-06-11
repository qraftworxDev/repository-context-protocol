package index

import (
	"fmt"
	"os"
	"path/filepath"

	"repository-context-protocol/internal/models"

	"github.com/vmihailenco/msgpack/v5"
)

const (
	// File permissions for chunk files (read/write for owner only)
	chunkFilePermissions = 0600
	// Directory permissions for chunk directory (read/write/execute for owner, read/execute for group and others)
	chunkDirPermissions = 0755
)

// ChunkSerializer handles saving and loading semantic chunks using MessagePack
type ChunkSerializer struct {
	baseDir string
}

// NewChunkSerializer creates a new chunk serializer with the specified base directory
func NewChunkSerializer(baseDir string) *ChunkSerializer {
	return &ChunkSerializer{
		baseDir: baseDir,
	}
}

// SaveChunk saves a semantic chunk to disk using MessagePack serialization
func (cs *ChunkSerializer) SaveChunk(chunk *models.SemanticChunk) error {
	// Ensure the base directory exists
	if err := os.MkdirAll(cs.baseDir, chunkDirPermissions); err != nil {
		return fmt.Errorf("failed to create chunk directory: %w", err)
	}

	// Generate the file path
	filePath := cs.GetChunkPath(chunk.ID)

	// Serialize the chunk using MessagePack
	data, err := msgpack.Marshal(chunk)
	if err != nil {
		return fmt.Errorf("failed to marshal chunk: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, chunkFilePermissions); err != nil {
		return fmt.Errorf("failed to write chunk file: %w", err)
	}

	return nil
}

// LoadChunk loads a semantic chunk from disk using MessagePack deserialization
func (cs *ChunkSerializer) LoadChunk(chunkID string) (models.SemanticChunk, error) {
	var chunk models.SemanticChunk

	// Generate the file path
	filePath := cs.GetChunkPath(chunkID)

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return chunk, fmt.Errorf("failed to read chunk file: %w", err)
	}

	// Deserialize using MessagePack
	if err := msgpack.Unmarshal(data, &chunk); err != nil {
		return chunk, fmt.Errorf("failed to unmarshal chunk: %w", err)
	}

	return chunk, nil
}

// GetChunkPath returns the file path for a given chunk ID
func (cs *ChunkSerializer) GetChunkPath(chunkID string) string {
	return filepath.Join(cs.baseDir, chunkID+".msgpack")
}

// ChunkExists checks if a chunk file exists on disk
func (cs *ChunkSerializer) ChunkExists(chunkID string) bool {
	filePath := cs.GetChunkPath(chunkID)
	_, err := os.Stat(filePath)
	return err == nil
}

// DeleteChunk removes a chunk file from disk
func (cs *ChunkSerializer) DeleteChunk(chunkID string) error {
	filePath := cs.GetChunkPath(chunkID)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete chunk file: %w", err)
	}
	return nil
}

// ListChunks returns a list of all chunk IDs in the base directory
func (cs *ChunkSerializer) ListChunks() ([]string, error) {
	entries, err := os.ReadDir(cs.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read chunk directory: %w", err)
	}

	var chunkIDs []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".msgpack" {
			// Remove the .msgpack extension to get the chunk ID
			chunkID := entry.Name()[:len(entry.Name())-8]
			chunkIDs = append(chunkIDs, chunkID)
		}
	}

	return chunkIDs, nil
}
