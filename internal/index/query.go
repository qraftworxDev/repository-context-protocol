package index

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"repository-context-protocol/internal/models"
)

// Constants for query engine
const (
	// Entity type constants
	EntityTypeFunction = "function"
	EntityTypeType     = "type"
	EntityTypeVariable = "variable"
	EntityTypeConstant = "constant"

	// Token estimation constants
	TokenOverhead   = 10
	CallerTokens    = 20
	CalleeTokens    = 20
	MetadataTokens  = 50
	DefaultMaxDepth = 1
)

// Query engine for semantic searches

// QueryEngine provides semantic search capabilities over the indexed repository
type QueryEngine struct {
	storage *HybridStorage
}

// QueryOptions configures search behavior and result formatting
type QueryOptions struct {
	IncludeCallers bool   `json:"include_callers"` // Include functions that call the target
	IncludeCallees bool   `json:"include_callees"` // Include functions called by the target
	IncludeTypes   bool   `json:"include_types"`   // Include related type definitions
	MaxDepth       int    `json:"max_depth"`       // Maximum depth for relationship traversal
	MaxTokens      int    `json:"max_tokens"`      // Maximum tokens for LLM consumption
	Format         string `json:"format"`          // Output format: "json" or "text"
}

// SearchResult represents the result of a search operation
type SearchResult struct {
	Query      string              `json:"query"`                // Original search query
	SearchType string              `json:"search_type"`          // Type of search performed
	Entries    []SearchResultEntry `json:"entries"`              // Matching entries with chunk data
	CallGraph  *CallGraphInfo      `json:"call_graph,omitempty"` // Call graph information
	TokenCount int                 `json:"token_count"`          // Estimated token count
	Truncated  bool                `json:"truncated"`            // Whether results were truncated
	ExecutedAt time.Time           `json:"executed_at"`          // When the query was executed
}

// SearchResultEntry combines index entry with chunk data
type SearchResultEntry struct {
	IndexEntry models.IndexEntry     `json:"index_entry"` // Basic index information
	ChunkData  *models.SemanticChunk `json:"chunk_data"`  // Detailed semantic data
}

// CallGraphInfo provides call relationship information
type CallGraphInfo struct {
	Function string           `json:"function"` // Target function name
	Callers  []CallGraphEntry `json:"callers"`  // Functions that call this function
	Callees  []CallGraphEntry `json:"callees"`  // Functions called by this function
	Depth    int              `json:"depth"`    // Traversal depth used
}

// CallGraphEntry represents a single call relationship
type CallGraphEntry struct {
	Function  string                `json:"function"`   // Function name
	File      string                `json:"file"`       // File where function is defined
	Line      int                   `json:"line"`       // Line number of call
	ChunkData *models.SemanticChunk `json:"chunk_data"` // Detailed semantic data
}

// NewQueryEngine creates a new query engine with the given storage
func NewQueryEngine(storage *HybridStorage) *QueryEngine {
	return &QueryEngine{
		storage: storage,
	}
}

// SearchByName searches for entities by exact name match
func (qe *QueryEngine) SearchByName(name string) (*SearchResult, error) {
	return qe.SearchByNameWithOptions(name, QueryOptions{})
}

// SearchByNameWithOptions searches for entities by name with additional options
func (qe *QueryEngine) SearchByNameWithOptions(name string, options QueryOptions) (*SearchResult, error) {
	result := &SearchResult{
		Query:      name,
		SearchType: "name",
		ExecutedAt: time.Now(),
	}

	// Query the storage for matching entries
	queryResults, err := qe.storage.QueryByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to query by name: %w", err)
	}

	// Convert storage results to query result entries
	result.Entries = make([]SearchResultEntry, len(queryResults))
	for i, qr := range queryResults {
		result.Entries[i] = SearchResultEntry(qr)
	}

	// Add call graph information if requested
	if (options.IncludeCallers || options.IncludeCallees) && len(result.Entries) > 0 {
		// Find the first function entry to get call graph for
		for _, entry := range result.Entries {
			if entry.IndexEntry.Type == EntityTypeFunction {
				callGraph, err := qe.GetCallGraph(entry.IndexEntry.Name, options.MaxDepth)
				if err == nil {
					result.CallGraph = callGraph
				}
				break
			}
		}
	}

	// Apply token limits and estimate tokens
	qe.applyTokenLimits(result, options.MaxTokens)

	return result, nil
}

// SearchByType searches for all entities of a specific type
func (qe *QueryEngine) SearchByType(entityType string) (*SearchResult, error) {
	result := &SearchResult{
		Query:      entityType,
		SearchType: "type",
		ExecutedAt: time.Now(),
	}

	// Query the storage for matching entries
	queryResults, err := qe.storage.QueryByType(entityType)
	if err != nil {
		return nil, fmt.Errorf("failed to query by type: %w", err)
	}

	// Convert storage results to query result entries
	result.Entries = make([]SearchResultEntry, len(queryResults))
	for i, qr := range queryResults {
		result.Entries[i] = SearchResultEntry(qr)
	}

	// Estimate tokens
	result.TokenCount = qe.EstimateTokens(result)

	return result, nil
}

// SearchByPattern searches for entities matching a pattern (supports wildcards)
func (qe *QueryEngine) SearchByPattern(pattern string) (*SearchResult, error) {
	result := &SearchResult{
		Query:      pattern,
		SearchType: "pattern",
		ExecutedAt: time.Now(),
	}

	// For now, implement simple prefix matching with *
	// In a full implementation, this could use more sophisticated pattern matching
	var allEntries []SearchResultEntry

	// Get all entity types and search each
	entityTypes := []string{EntityTypeFunction, EntityTypeType, EntityTypeVariable, EntityTypeConstant}
	for _, entityType := range entityTypes {
		queryResults, err := qe.storage.QueryByType(entityType)
		if err != nil {
			continue // Skip errors and continue with other types
		}

		for _, qr := range queryResults {
			if qe.matchesPattern(qr.IndexEntry.Name, pattern) {
				allEntries = append(allEntries, SearchResultEntry(qr))
			}
		}
	}

	result.Entries = allEntries
	result.TokenCount = qe.EstimateTokens(result)

	return result, nil
}

// SearchInFile searches for all entities within a specific file
func (qe *QueryEngine) SearchInFile(filePath string) (*SearchResult, error) {
	result := &SearchResult{
		Query:      filePath,
		SearchType: "file",
		ExecutedAt: time.Now(),
	}

	// Get all entity types and filter by file
	var allEntries []SearchResultEntry
	entityTypes := []string{EntityTypeFunction, EntityTypeType, EntityTypeVariable, EntityTypeConstant}

	for _, entityType := range entityTypes {
		queryResults, err := qe.storage.QueryByType(entityType)
		if err != nil {
			continue
		}

		for _, qr := range queryResults {
			// Match file path (handle both absolute and relative paths)
			if qr.IndexEntry.File == filePath || filepath.Base(qr.IndexEntry.File) == filepath.Base(filePath) {
				allEntries = append(allEntries, SearchResultEntry(qr))
			}
		}
	}

	result.Entries = allEntries
	result.TokenCount = qe.EstimateTokens(result)

	return result, nil
}

// GetCallGraph retrieves call graph information for a function
func (qe *QueryEngine) GetCallGraph(functionName string, maxDepth int) (*CallGraphInfo, error) {
	if maxDepth <= 0 {
		maxDepth = DefaultMaxDepth
	}

	callGraph := &CallGraphInfo{
		Function: functionName,
		Depth:    maxDepth,
	}

	// Get callers (functions that call this function)
	callers, err := qe.storage.QueryCallsTo(functionName)
	if err == nil {
		for _, caller := range callers {
			// Load chunk data for the caller
			callerEntries, callerErr := qe.storage.QueryByName(caller.Caller)
			if callerErr == nil && len(callerEntries) > 0 {
				callGraph.Callers = append(callGraph.Callers, CallGraphEntry{
					Function:  caller.Caller,
					File:      caller.CallerFile,
					Line:      caller.Line,
					ChunkData: callerEntries[0].ChunkData,
				})
			}
		}
	}

	// Get callees (functions called by this function)
	callees, err := qe.storage.QueryCallsFrom(functionName)
	if err == nil {
		for _, callee := range callees {
			// Load chunk data for the callee
			calleeEntries, calleeErr := qe.storage.QueryByName(callee.Callee)
			if calleeErr == nil && len(calleeEntries) > 0 {
				callGraph.Callees = append(callGraph.Callees, CallGraphEntry{
					Function:  callee.Callee,
					File:      callee.File,
					Line:      callee.Line,
					ChunkData: calleeEntries[0].ChunkData,
				})
			}
		}
	}

	return callGraph, nil
}

// FormatResults formats query results in the specified format
func (qe *QueryEngine) FormatResults(result *SearchResult, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "json":
		return json.MarshalIndent(result, "", "  ")
	case "text", "":
		return qe.formatAsText(result), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// EstimateTokens estimates the token count for a query result
func (qe *QueryEngine) EstimateTokens(result *SearchResult) int {
	tokenCount := 0

	// Estimate tokens for each entry
	for _, entry := range result.Entries {
		// Basic entry information
		tokenCount += len(strings.Fields(entry.IndexEntry.Name)) +
			len(strings.Fields(entry.IndexEntry.Signature)) +
			TokenOverhead

		// Chunk data tokens (if present)
		if entry.ChunkData != nil {
			tokenCount += entry.ChunkData.TokenCount
		}
	}

	// Call graph tokens
	if result.CallGraph != nil {
		tokenCount += len(result.CallGraph.Callers) * CallerTokens
		tokenCount += len(result.CallGraph.Callees) * CalleeTokens
	}

	// Metadata overhead
	tokenCount += MetadataTokens

	return tokenCount
}

// Helper methods

func (qe *QueryEngine) matchesPattern(name, pattern string) bool {
	// Simple pattern matching (supports * wildcard at the end)
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(name, prefix)
	}
	return name == pattern
}

func (qe *QueryEngine) applyTokenLimits(result *SearchResult, maxTokens int) {
	if maxTokens <= 0 {
		result.TokenCount = qe.EstimateTokens(result)
		return
	}

	currentTokens := 0
	truncatedEntries := []SearchResultEntry{}

	for _, entry := range result.Entries {
		entryTokens := qe.estimateEntryTokens(&entry)
		if currentTokens+entryTokens <= maxTokens {
			truncatedEntries = append(truncatedEntries, entry)
			currentTokens += entryTokens
		} else {
			result.Truncated = true
			break
		}
	}

	result.Entries = truncatedEntries
	result.TokenCount = currentTokens
}

func (qe *QueryEngine) estimateEntryTokens(entry *SearchResultEntry) int {
	tokens := len(strings.Fields(entry.IndexEntry.Name)) +
		len(strings.Fields(entry.IndexEntry.Signature)) +
		TokenOverhead

	if entry.ChunkData != nil {
		tokens += entry.ChunkData.TokenCount
	}

	return tokens
}

func (qe *QueryEngine) formatAsText(result *SearchResult) []byte {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Query: %s (type: %s)\n", result.Query, result.SearchType))
	output.WriteString(fmt.Sprintf("Results: %d entries\n", len(result.Entries)))
	output.WriteString(fmt.Sprintf("Token count: %d\n", result.TokenCount))
	if result.Truncated {
		output.WriteString("Results truncated due to token limit\n")
	}
	output.WriteString("\n")

	for i, entry := range result.Entries {
		output.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, entry.IndexEntry.Name, entry.IndexEntry.Type))
		output.WriteString(fmt.Sprintf("   File: %s:%d-%d\n", entry.IndexEntry.File, entry.IndexEntry.StartLine, entry.IndexEntry.EndLine))
		if entry.IndexEntry.Signature != "" {
			output.WriteString(fmt.Sprintf("   Signature: %s\n", entry.IndexEntry.Signature))
		}
		output.WriteString("\n")
	}

	if result.CallGraph != nil {
		output.WriteString("Call Graph:\n")
		if len(result.CallGraph.Callers) > 0 {
			output.WriteString("  Callers:\n")
			for _, caller := range result.CallGraph.Callers {
				output.WriteString(fmt.Sprintf("    - %s (%s:%d)\n", caller.Function, caller.File, caller.Line))
			}
		}
		if len(result.CallGraph.Callees) > 0 {
			output.WriteString("  Callees:\n")
			for _, callee := range result.CallGraph.Callees {
				output.WriteString(fmt.Sprintf("    - %s (%s:%d)\n", callee.Function, callee.File, callee.Line))
			}
		}
	}

	return []byte(output.String())
}
