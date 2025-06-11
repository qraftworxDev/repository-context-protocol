package index

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"repository-context-protocol/internal/models"
)

const (
	// Token estimation constants
	baseFileTokens     = 10 // Base tokens for file structure
	functionBaseTokens = 5  // Base tokens per function
	parameterTokens    = 3  // Tokens per parameter
	returnTokens       = 3  // Tokens per return value
	callTokens         = 2  // Tokens per function call
	typeBaseTokens     = 5  // Base tokens per type
	fieldTokens        = 4  // Tokens per struct field
	methodBaseTokens   = 3  // Base tokens per method
	variableTokens     = 3  // Tokens per variable
	constantTokens     = 4  // Tokens per constant
	importTokens       = 2  // Tokens per import
	exportTokens       = 2  // Tokens per export
)

// ChunkingStrategy defines how files should be grouped into semantic chunks
type ChunkingStrategy interface {
	CreateChunks(files []models.FileContext) []models.SemanticChunk
}

// FileBasedChunking creates one chunk per file
// This is the simplest strategy - each file becomes its own chunk
type FileBasedChunking struct {
	MaxTokens int // Maximum tokens per chunk (for future use)
}

// CreateChunks implements ChunkingStrategy for file-based chunking
func (f *FileBasedChunking) CreateChunks(files []models.FileContext) []models.SemanticChunk {
	var chunks []models.SemanticChunk

	for i := range files {
		file := &files[i]
		chunk := models.SemanticChunk{
			ID:         generateChunkID(file.Path),
			Files:      []string{file.Path},
			FileData:   []models.FileContext{*file},
			TokenCount: estimateTokens(file),
			CreatedAt:  time.Now(),
		}
		chunks = append(chunks, chunk)
	}

	return chunks
}

// generateChunkID creates a deterministic, unique ID for a chunk based on file path
func generateChunkID(filePath string) string {
	// Use SHA-256 hash of the file path for deterministic, unique IDs
	hash := sha256.Sum256([]byte(filePath))
	hashStr := fmt.Sprintf("%x", hash)

	// Take first 16 characters for a reasonable length ID
	// This gives us 64 bits of entropy, which should be sufficient for uniqueness
	chunkID := hashStr[:16]

	// Add a prefix to make it clear this is a chunk ID
	return "chunk_" + chunkID
}

// estimateTokens provides a rough estimate of token count for a file
// This is used for LLM context window planning
func estimateTokens(file *models.FileContext) int {
	tokens := 0

	// Base tokens for file structure (path, language, etc.)
	tokens += baseFileTokens

	// Estimate tokens for functions
	for i := range file.Functions {
		fn := &file.Functions[i]
		// Function name + signature + basic structure
		tokens += len(strings.Fields(fn.Signature)) + functionBaseTokens

		// Add tokens for parameters and returns
		tokens += len(fn.Parameters) * parameterTokens
		tokens += len(fn.Returns) * returnTokens

		// Add tokens for function calls (rough estimate)
		tokens += len(fn.Calls) * callTokens
	}

	// Estimate tokens for types
	for i := range file.Types {
		typ := &file.Types[i]
		// Type name + kind + basic structure
		tokens += typeBaseTokens

		// Add tokens for fields
		tokens += len(typ.Fields) * fieldTokens

		// Add tokens for methods
		for j := range typ.Methods {
			method := &typ.Methods[j]
			tokens += len(strings.Fields(method.Signature)) + methodBaseTokens
		}
	}

	// Estimate tokens for variables
	for range file.Variables {
		// Variable name + type
		tokens += variableTokens
	}

	// Estimate tokens for constants
	for range file.Constants {
		// Constant name + type + value
		tokens += constantTokens
	}

	// Estimate tokens for imports
	tokens += len(file.Imports) * importTokens

	// Estimate tokens for exports
	tokens += len(file.Exports) * exportTokens

	return tokens
}
