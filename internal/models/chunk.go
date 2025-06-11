package models

import "time"

// SemanticChunk represents a group of related files and their detailed semantic data
// stored as MessagePack for efficient serialization and LLM consumption
type SemanticChunk struct {
	ID         string        `json:"id"`          // Unique identifier for this chunk
	Files      []string      `json:"files"`       // List of file paths included in this chunk
	FileData   []FileContext `json:"file_data"`   // Complete FileContext data for all files
	TokenCount int           `json:"token_count"` // Estimated token count for LLM consumption
	CreatedAt  time.Time     `json:"created_at"`  // When this chunk was created
}

// Manifest tracks all chunks and provides metadata for the chunked storage system
type Manifest struct {
	Version   string               `json:"version"`    // Version of the manifest format
	Chunks    map[string]ChunkInfo `json:"chunks"`     // ChunkID -> ChunkInfo mapping
	UpdatedAt time.Time            `json:"updated_at"` // Last update timestamp
}

// ChunkInfo provides metadata about a specific chunk without loading the full chunk data
type ChunkInfo struct {
	Files      []string  `json:"files"`       // List of files in this chunk
	Size       int       `json:"size"`        // Size in bytes of the serialized chunk
	TokenCount int       `json:"token_count"` // Estimated token count
	UpdatedAt  time.Time `json:"updated_at"`  // When this chunk was last updated
}
