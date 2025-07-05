package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"repository-context-protocol/internal/models"
)

const (
	// Directory permissions for hybrid storage directories
	dirPermissions  = 0755
	filePermissions = 0600
)

// HybridStorage combines SQLite indexing with MessagePack chunk storage
type HybridStorage struct {
	baseDir          string
	sqliteIndex      *SQLiteIndex
	chunkSerializer  *ChunkSerializer
	chunkingStrategy ChunkingStrategy
	manifest         *models.Manifest
	manifestPath     string
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
		baseDir:      baseDir,
		manifestPath: filepath.Join(baseDir, "manifest.json"),
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
	h.sqliteIndex = NewSQLiteIndex(dbPath)
	if err := h.sqliteIndex.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize SQLite index: %w", err)
	}

	// Initialize chunk serializer
	h.chunkSerializer = NewChunkSerializer(chunksDir)

	// Initialize chunking strategy (file-based for now)
	h.chunkingStrategy = &FileBasedChunking{}

	// Load or initialize manifest
	if err := h.loadManifest(); err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

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

	// Update manifest with chunk information
	if err := h.updateManifestForChunk(chunk); err != nil {
		return fmt.Errorf("failed to update manifest for chunk %s: %w", chunk.ID, err)
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

		// Index function calls - use GetAllCalls() to handle both deprecated and new fields
		for _, call := range function.GetAllCalls() {
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
			Signature: h.buildTypeSignature(typeDef),
		}
		if err := h.sqliteIndex.InsertIndexEntry(&entry); err != nil {
			return fmt.Errorf("failed to insert type index entry: %w", err)
		}
	}
	return nil
}

// buildTypeSignature creates a signature string for a type definition
func (h *HybridStorage) buildTypeSignature(typeDef *models.TypeDef) string {
	if len(typeDef.Embedded) == 0 {
		// No inheritance - just return the class name
		return typeDef.Name
	}

	// Include inheritance information: ClassName(BaseClass1, BaseClass2)
	signature := typeDef.Name + "("
	for i, embedded := range typeDef.Embedded {
		if i > 0 {
			signature += ", "
		}
		signature += embedded
	}
	signature += ")"
	return signature
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

// QueryByTypes searches for entries by multiple types in a single query (Phase 4A.1.2)
func (h *HybridStorage) QueryByTypes(entryTypes []string) ([]QueryResult, error) {
	if h.sqliteIndex == nil || h.chunkSerializer == nil {
		return nil, fmt.Errorf("hybrid storage not initialized")
	}

	// Query SQLite index using batch method
	entries, err := h.sqliteIndex.QueryIndexEntriesByTypes(entryTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to query index entries by types: %w", err)
	}

	return h.loadChunkDataForEntries(entries)
}

// QueryByNames searches for entries by multiple names in a single query (Phase 4A.1.2)
func (h *HybridStorage) QueryByNames(names []string) ([]QueryResult, error) {
	if h.sqliteIndex == nil || h.chunkSerializer == nil {
		return nil, fmt.Errorf("hybrid storage not initialized")
	}

	// Query SQLite index using batch method
	entries, err := h.sqliteIndex.QueryIndexEntriesByNames(names)
	if err != nil {
		return nil, fmt.Errorf("failed to query index entries by names: %w", err)
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

			// Remove chunk from manifest
			if h.manifest != nil {
				delete(h.manifest.Chunks, chunkID)
				if err := h.saveManifest(); err != nil {
					return fmt.Errorf("failed to update manifest after deleting chunk: %w", err)
				}
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

// loadManifest loads the manifest from disk or creates a new one
func (h *HybridStorage) loadManifest() error {
	// Try to load existing manifest
	if data, err := os.ReadFile(h.manifestPath); err == nil {
		// Manifest exists, parse it
		h.manifest = &models.Manifest{}
		if unmarshalErr := json.Unmarshal(data, h.manifest); unmarshalErr != nil {
			return fmt.Errorf("failed to parse manifest: %w", unmarshalErr)
		}
		// Ensure Chunks map is initialized (in case JSON had null/missing chunks field)
		if h.manifest.Chunks == nil {
			h.manifest.Chunks = make(map[string]models.ChunkInfo)
		}
	} else if os.IsNotExist(err) {
		// Manifest doesn't exist, create a new one
		h.manifest = &models.Manifest{
			Version:   "1.0.0",
			Chunks:    make(map[string]models.ChunkInfo),
			UpdatedAt: time.Now(),
		}
		// Save the new manifest
		if saveErr := h.saveManifest(); saveErr != nil {
			return fmt.Errorf("failed to save new manifest: %w", saveErr)
		}
	} else {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	return nil
}

// saveManifest saves the current manifest to disk
func (h *HybridStorage) saveManifest() error {
	if h.manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	// Update timestamp
	h.manifest.UpdatedAt = time.Now()

	// Marshal to JSON
	data, err := json.MarshalIndent(h.manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Write to file
	if err := os.WriteFile(h.manifestPath, data, filePermissions); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// updateManifestForChunk adds or updates chunk information in the manifest
func (h *HybridStorage) updateManifestForChunk(chunk *models.SemanticChunk) error {
	if h.manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	// Ensure Chunks map is initialized (defensive programming)
	if h.manifest.Chunks == nil {
		h.manifest.Chunks = make(map[string]models.ChunkInfo)
	}

	// Create chunk info
	chunkInfo := models.ChunkInfo{
		Files:      chunk.Files,
		Size:       0, // We'll calculate this from the serialized chunk
		TokenCount: chunk.TokenCount,
		UpdatedAt:  chunk.CreatedAt,
	}

	// Calculate serialized chunk size if possible
	chunkPath := h.chunkSerializer.GetChunkPath(chunk.ID)
	if info, err := os.Stat(chunkPath); err == nil {
		chunkInfo.Size = int(info.Size())
	}

	// Add to manifest
	h.manifest.Chunks[chunk.ID] = chunkInfo

	// Save manifest
	return h.saveManifest()
}

// GetFunctionImplementation extracts the full function implementation from source files
// using AST span information (StartLine/EndLine) and surrounding context lines
func (h *HybridStorage) GetFunctionImplementation(functionName string, contextLines int) (*FunctionImplementation, error) {
	// Query for the function to get its location and span information
	entries, err := h.QueryByName(functionName)
	if err != nil {
		return nil, fmt.Errorf("failed to query function %s: %w", functionName, err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("function %s not found", functionName)
	}

	// Find the function entry (there might be multiple if name conflicts exist)
	var functionEntry *QueryResult
	for i := range entries {
		if entries[i].IndexEntry.Type == EntityTypeFunction && entries[i].IndexEntry.Name == functionName {
			functionEntry = &entries[i]
			break
		}
	}

	if functionEntry == nil {
		return nil, fmt.Errorf("function %s not found in query results", functionName)
	}

	// Extract implementation from source file
	return h.extractFunctionFromSource(functionEntry, contextLines)
}

// FunctionImplementation represents the extracted function implementation
type FunctionImplementation struct {
	Body         string   `json:"body"`
	ContextLines []string `json:"context_lines"`
}

// extractFunctionFromSource reads the source file and extracts the function body and context
func (h *HybridStorage) extractFunctionFromSource(entry *QueryResult, contextLines int) (*FunctionImplementation, error) {
	filePath := entry.IndexEntry.File
	startLine := entry.IndexEntry.StartLine
	endLine := entry.IndexEntry.EndLine

	// Validate line numbers
	if startLine <= 0 || endLine <= 0 || startLine > endLine {
		return &FunctionImplementation{
			Body: "// Function implementation unavailable: invalid line numbers",
			ContextLines: []string{
				"// Context unavailable: invalid AST span information",
			},
		}, nil
	}

	// Read the source file
	content, err := os.ReadFile(filePath) // #nosec G304 - File path comes from our indexed data
	if err != nil {
		return &FunctionImplementation{
			Body: fmt.Sprintf("// Function implementation unavailable: %v", err),
			ContextLines: []string{
				fmt.Sprintf("// Context unavailable: failed to read file %s", filePath),
			},
		}, nil
	}

	// Split content into lines
	lines := strings.Split(string(content), "\n")

	// Validate line numbers against actual file content
	if startLine > len(lines) || endLine > len(lines) {
		return &FunctionImplementation{
			Body: "// Function implementation unavailable: line numbers exceed file length",
			ContextLines: []string{
				fmt.Sprintf("// Context unavailable: file has %d lines, function spans %d-%d", len(lines), startLine, endLine),
			},
		}, nil
	}

	// Extract function body (convert from 1-indexed to 0-indexed)
	functionLines := lines[startLine-1 : endLine]
	functionBody := strings.Join(functionLines, "\n")

	// Extract context lines before and after the function
	var contextLinesResult []string

	// Context before function
	contextStart := max(0, startLine-1-contextLines)
	contextEnd := startLine - 1
	if contextEnd > contextStart {
		beforeContext := lines[contextStart:contextEnd]
		for i, line := range beforeContext {
			lineNum := contextStart + i + 1
			contextLinesResult = append(contextLinesResult, fmt.Sprintf("%d: %s", lineNum, line))
		}
	}

	// Add separator if we have before context
	if len(contextLinesResult) > 0 {
		contextLinesResult = append(contextLinesResult, "// --- Function body starts ---")
	}

	// Context after function
	contextStart = endLine
	contextEnd = min(len(lines), endLine+contextLines)
	if contextEnd > contextStart {
		afterContext := lines[contextStart:contextEnd]
		if len(contextLinesResult) > 0 {
			contextLinesResult = append(contextLinesResult, "// --- Function body ends ---")
		}
		for i, line := range afterContext {
			lineNum := contextStart + i + 1
			contextLinesResult = append(contextLinesResult, fmt.Sprintf("%d: %s", lineNum, line))
		}
	}

	// If no context lines were extracted, provide a minimal context
	if len(contextLinesResult) == 0 {
		contextLinesResult = []string{
			fmt.Sprintf("// Function %s in %s (lines %d-%d)", entry.IndexEntry.Name, filepath.Base(filePath), startLine, endLine),
		}
	}

	return &FunctionImplementation{
		Body:         functionBody,
		ContextLines: contextLinesResult,
	}, nil
}
