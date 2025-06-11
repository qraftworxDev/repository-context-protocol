package index

import (
	"fmt"
	"os"
	"path/filepath"

	"repository-context-protocol/internal/models"
)

const (
	// Directory permissions for hybrid storage directories
	dirPermissions = 0755
)

// HybridStorage combines SQLite indexing with MessagePack chunk storage
type HybridStorage struct {
	baseDir          string
	sqliteIndex      *SQLiteIndex
	chunkSerializer  *ChunkSerializer
	chunkingStrategy ChunkingStrategy
}

// QueryResult combines index entry with chunk data
type QueryResult struct {
	IndexEntry models.IndexEntry
	ChunkData  *models.SemanticChunk
}

// CallGraphResult combines call relation with chunk data
type CallGraphResult struct {
	CallRelation models.CallRelation
	ChunkData    *models.SemanticChunk
}

// NewHybridStorage creates a new hybrid storage instance
func NewHybridStorage(baseDir string) *HybridStorage {
	return &HybridStorage{
		baseDir: baseDir,
	}
}

// Initialize sets up the hybrid storage system
func (h *HybridStorage) Initialize() error {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(h.baseDir, dirPermissions); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create chunks directory
	chunksDir := filepath.Join(h.baseDir, "chunks")
	if err := os.MkdirAll(chunksDir, dirPermissions); err != nil {
		return fmt.Errorf("failed to create chunks directory: %w", err)
	}

	// Initialize SQLite index
	dbPath := filepath.Join(h.baseDir, "index.db")
	h.sqliteIndex = &SQLiteIndex{dbPath: dbPath}
	if err := h.sqliteIndex.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize SQLite index: %w", err)
	}

	// Initialize chunk serializer
	h.chunkSerializer = NewChunkSerializer(chunksDir)

	// Initialize chunking strategy (file-based for now)
	h.chunkingStrategy = &FileBasedChunking{}

	return nil
}

// StoreFileContext stores a file context using hybrid storage
func (h *HybridStorage) StoreFileContext(fileContext *models.FileContext) error {
	if h.sqliteIndex == nil || h.chunkSerializer == nil || h.chunkingStrategy == nil {
		return fmt.Errorf("hybrid storage not initialized")
	}

	// First, delete any existing data for this file to handle updates
	if err := h.DeleteFile(fileContext.Path); err != nil {
		return fmt.Errorf("failed to delete existing file data: %w", err)
	}

	// Generate chunks using the chunking strategy
	chunks := h.chunkingStrategy.CreateChunks([]models.FileContext{*fileContext})

	// Store each chunk
	for i := range chunks {
		chunk := &chunks[i]
		if err := h.storeChunk(chunk); err != nil {
			return err
		}
	}

	return nil
}

// storeChunk saves a chunk and creates all necessary index entries
func (h *HybridStorage) storeChunk(chunk *models.SemanticChunk) error {
	// Save chunk data using MessagePack
	if err := h.chunkSerializer.SaveChunk(chunk); err != nil {
		return fmt.Errorf("failed to save chunk %s: %w", chunk.ID, err)
	}

	// Collect file paths from chunk
	filePaths := make([]string, 0, len(chunk.FileData))
	for i := range chunk.FileData {
		filePaths = append(filePaths, chunk.FileData[i].Path)
	}

	// Register chunk in SQLite
	if err := h.sqliteIndex.RegisterChunk(chunk.ID, filePaths, chunk.TokenCount, chunk.CreatedAt); err != nil {
		return fmt.Errorf("failed to register chunk %s: %w", chunk.ID, err)
	}

	// Create index entries for fast lookups
	for i := range chunk.FileData {
		fileData := &chunk.FileData[i]
		if err := h.indexFileData(fileData, chunk.ID); err != nil {
			return err
		}
	}

	return nil
}

// indexFileData creates index entries for a single file's data
func (h *HybridStorage) indexFileData(fileData *models.FileContext, chunkID string) error {
	if err := h.indexFunctions(fileData, chunkID); err != nil {
		return err
	}
	if err := h.indexTypes(fileData, chunkID); err != nil {
		return err
	}
	if err := h.indexVariables(fileData, chunkID); err != nil {
		return err
	}
	if err := h.indexConstants(fileData, chunkID); err != nil {
		return err
	}
	return nil
}

// indexFunctions creates index entries for functions and their call relations
func (h *HybridStorage) indexFunctions(fileData *models.FileContext, chunkID string) error {
	for i := range fileData.Functions {
		function := &fileData.Functions[i]
		entry := models.IndexEntry{
			Name:      function.Name,
			Type:      "function",
			File:      fileData.Path,
			StartLine: function.StartLine,
			EndLine:   function.EndLine,
			ChunkID:   chunkID,
			Signature: function.Signature,
		}
		if err := h.sqliteIndex.InsertIndexEntry(&entry); err != nil {
			return fmt.Errorf("failed to insert function index entry: %w", err)
		}

		// Index function calls
		for _, call := range function.Calls {
			relation := models.CallRelation{
				Caller:     function.Name,
				Callee:     call,
				File:       fileData.Path,
				Line:       function.StartLine, // Use function start line as call line
				CallerFile: fileData.Path,
			}
			if err := h.sqliteIndex.InsertCallRelation(relation); err != nil {
				return fmt.Errorf("failed to insert call relation: %w", err)
			}
		}
	}
	return nil
}

// indexTypes creates index entries for type definitions
func (h *HybridStorage) indexTypes(fileData *models.FileContext, chunkID string) error {
	for i := range fileData.Types {
		typeDef := &fileData.Types[i]
		entry := models.IndexEntry{
			Name:      typeDef.Name,
			Type:      typeDef.Kind,
			File:      fileData.Path,
			StartLine: typeDef.StartLine,
			EndLine:   typeDef.EndLine,
			ChunkID:   chunkID,
		}
		if err := h.sqliteIndex.InsertIndexEntry(&entry); err != nil {
			return fmt.Errorf("failed to insert type index entry: %w", err)
		}
	}
	return nil
}

// indexVariables creates index entries for variable declarations
func (h *HybridStorage) indexVariables(fileData *models.FileContext, chunkID string) error {
	for i := range fileData.Variables {
		variable := &fileData.Variables[i]
		entry := models.IndexEntry{
			Name:      variable.Name,
			Type:      "variable",
			File:      fileData.Path,
			StartLine: variable.StartLine,
			EndLine:   variable.EndLine,
			ChunkID:   chunkID,
		}
		if err := h.sqliteIndex.InsertIndexEntry(&entry); err != nil {
			return fmt.Errorf("failed to insert variable index entry: %w", err)
		}
	}
	return nil
}

// indexConstants creates index entries for constant declarations
func (h *HybridStorage) indexConstants(fileData *models.FileContext, chunkID string) error {
	for i := range fileData.Constants {
		constant := &fileData.Constants[i]
		entry := models.IndexEntry{
			Name:      constant.Name,
			Type:      "constant",
			File:      fileData.Path,
			StartLine: constant.StartLine,
			EndLine:   constant.EndLine,
			ChunkID:   chunkID,
		}
		if err := h.sqliteIndex.InsertIndexEntry(&entry); err != nil {
			return fmt.Errorf("failed to insert constant index entry: %w", err)
		}
	}
	return nil
}

// loadChunkDataForEntries is a helper function to load chunk data for index entries
func (h *HybridStorage) loadChunkDataForEntries(entries []models.IndexEntry) ([]QueryResult, error) {
	results := make([]QueryResult, 0, len(entries))
	for _, entry := range entries {
		chunkData, err := h.chunkSerializer.LoadChunk(entry.ChunkID)
		if err != nil {
			return nil, fmt.Errorf("failed to load chunk %s: %w", entry.ChunkID, err)
		}

		results = append(results, QueryResult{
			IndexEntry: entry,
			ChunkData:  &chunkData,
		})
	}
	return results, nil
}

// QueryByName searches for entries by name and returns results with chunk data
func (h *HybridStorage) QueryByName(name string) ([]QueryResult, error) {
	if h.sqliteIndex == nil || h.chunkSerializer == nil {
		return nil, fmt.Errorf("hybrid storage not initialized")
	}

	// Query SQLite index
	entries, err := h.sqliteIndex.QueryIndexEntries(name)
	if err != nil {
		return nil, fmt.Errorf("failed to query index entries: %w", err)
	}

	return h.loadChunkDataForEntries(entries)
}

// QueryByType searches for entries by type and returns results with chunk data
func (h *HybridStorage) QueryByType(entryType string) ([]QueryResult, error) {
	if h.sqliteIndex == nil || h.chunkSerializer == nil {
		return nil, fmt.Errorf("hybrid storage not initialized")
	}

	// Query SQLite index
	entries, err := h.sqliteIndex.QueryIndexEntriesByType(entryType)
	if err != nil {
		return nil, fmt.Errorf("failed to query index entries by type: %w", err)
	}

	return h.loadChunkDataForEntries(entries)
}

// QueryCallsFrom returns functions called by the given function
func (h *HybridStorage) QueryCallsFrom(functionName string) ([]models.CallRelation, error) {
	if h.sqliteIndex == nil {
		return nil, fmt.Errorf("hybrid storage not initialized")
	}

	return h.sqliteIndex.QueryCallsFrom(functionName)
}

// QueryCallsTo returns functions that call the given function
func (h *HybridStorage) QueryCallsTo(functionName string) ([]models.CallRelation, error) {
	if h.sqliteIndex == nil {
		return nil, fmt.Errorf("hybrid storage not initialized")
	}

	return h.sqliteIndex.QueryCallsTo(functionName)
}

// QueryCallsFromWithChunkData returns call relations with chunk data
func (h *HybridStorage) QueryCallsFromWithChunkData(functionName string) ([]CallGraphResult, error) {
	if h.sqliteIndex == nil || h.chunkSerializer == nil {
		return nil, fmt.Errorf("hybrid storage not initialized")
	}

	// Query call relations
	relations, err := h.sqliteIndex.QueryCallsFrom(functionName)
	if err != nil {
		return nil, fmt.Errorf("failed to query call relations: %w", err)
	}

	// Load chunk data for each relation by finding the chunk ID from the caller function
	results := make([]CallGraphResult, 0, len(relations))
	for _, relation := range relations {
		// Find the chunk ID by looking up the caller function in the index
		callerEntries, err := h.sqliteIndex.QueryIndexEntries(relation.Caller)
		if err != nil || len(callerEntries) == 0 {
			continue // Skip if we can't find the caller
		}

		// Use the first matching entry's chunk ID
		chunkData, err := h.chunkSerializer.LoadChunk(callerEntries[0].ChunkID)
		if err != nil {
			continue // Skip if we can't load the chunk
		}

		results = append(results, CallGraphResult{
			CallRelation: relation,
			ChunkData:    &chunkData,
		})
	}

	return results, nil
}

// DeleteFile removes all data associated with a file
func (h *HybridStorage) DeleteFile(filePath string) error {
	if h.sqliteIndex == nil || h.chunkSerializer == nil {
		return fmt.Errorf("hybrid storage not initialized")
	}

	// Find chunks associated with this file
	// For file-based chunking, we need to find the chunk ID for this file
	// This is a simplified approach - in practice, we might need a more sophisticated lookup
	chunkIDs, err := h.chunkSerializer.ListChunks()
	if err != nil {
		return fmt.Errorf("failed to list chunks: %w", err)
	}

	for _, chunkID := range chunkIDs {
		chunk, err := h.chunkSerializer.LoadChunk(chunkID)
		if err != nil {
			continue // Skip chunks we can't load
		}

		// Check if this chunk contains the file
		containsFile := false
		for i := range chunk.FileData {
			if chunk.FileData[i].Path == filePath {
				containsFile = true
				break
			}
		}

		if containsFile {
			// Delete the chunk from SQLite (this will cascade delete index entries and call relations)
			if err := h.sqliteIndex.DeleteChunk(chunkID); err != nil {
				return fmt.Errorf("failed to delete chunk from SQLite: %w", err)
			}

			// Delete the chunk file
			if err := h.chunkSerializer.DeleteChunk(chunkID); err != nil {
				return fmt.Errorf("failed to delete chunk file: %w", err)
			}
		}
	}

	return nil
}

// Close closes the hybrid storage and releases resources
func (h *HybridStorage) Close() error {
	if h.sqliteIndex != nil {
		if err := h.sqliteIndex.Close(); err != nil {
			return fmt.Errorf("failed to close SQLite index: %w", err)
		}
		h.sqliteIndex = nil
	}

	h.chunkSerializer = nil
	h.chunkingStrategy = nil

	return nil
}
