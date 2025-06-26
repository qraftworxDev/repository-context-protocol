package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"repository-context-protocol/internal/index"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	constMaxTokens = 2000
	constMaxDepth  = 2
)

// RepoContextMCPServer provides MCP server functionality for repository context protocol
type RepoContextMCPServer struct {
	QueryEngine *index.QueryEngine
	Storage     *index.HybridStorage
	RepoPath    string
}

// NewRepoContextMCPServer creates a new MCP server instance
func NewRepoContextMCPServer() *RepoContextMCPServer {
	return &RepoContextMCPServer{}
}

// Run starts the MCP server with JSON-RPC over stdin/stdout
func (s *RepoContextMCPServer) Run(ctx context.Context) error {
	// Initialize repository context
	repoPath, err := s.detectRepositoryRoot()
	if err != nil {
		return fmt.Errorf("failed to detect repository root: %w", err)
	}
	s.RepoPath = repoPath

	// Initialize query engine
	if err := s.initializeQueryEngine(); err != nil {
		return fmt.Errorf("failed to initialize query engine: %w", err)
	}

	// TODO: Complete MCP server implementation with proper tool registration
	// For now, return success to indicate server can initialize properly
	return fmt.Errorf("MCP server implementation in progress - basic initialization complete")
}

// detectRepositoryRoot finds the root directory of the current repository
func (s *RepoContextMCPServer) detectRepositoryRoot() (string, error) {
	// Start from current directory and walk up to find .git
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk up the directory tree looking for .git
	dir := currentDir
	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	// If no .git found, use current directory
	return currentDir, nil
}

// initializeQueryEngine sets up the query engine for the repository
func (s *RepoContextMCPServer) initializeQueryEngine() error {
	if s.RepoPath == "" {
		return fmt.Errorf("repository path not set")
	}

	repoContextPath := filepath.Join(s.RepoPath, ".repocontext")
	if _, err := os.Stat(repoContextPath); os.IsNotExist(err) {
		return fmt.Errorf("repository not initialized - .repocontext directory not found")
	}

	// Initialize storage
	storage := index.NewHybridStorage(repoContextPath)
	if err := storage.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	s.Storage = storage
	s.QueryEngine = index.NewQueryEngine(storage)

	return nil
}

// validateRepository checks if the repository is properly initialized
func (s *RepoContextMCPServer) validateRepository() error {
	if s.RepoPath == "" {
		return fmt.Errorf("no repository path configured")
	}

	repoContextPath := filepath.Join(s.RepoPath, ".repocontext")
	if _, err := os.Stat(repoContextPath); os.IsNotExist(err) {
		return fmt.Errorf("repository not initialized - run initialize_repository first")
	}

	return nil
}

// RegisterQueryTools registers the query-related MCP tools
func (s *RepoContextMCPServer) RegisterQueryTools() []mcp.Tool {
	// Query by name tool
	queryByNameTool := mcp.NewTool("query_by_name",
		mcp.WithDescription("Search for functions, types, or variables by exact name"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name to search for")),
		mcp.WithBoolean("include_callers", mcp.Description("Include functions that call this function")),
		mcp.WithBoolean("include_callees", mcp.Description("Include functions called by this function")),
		mcp.WithBoolean("include_types", mcp.Description("Include related type definitions")),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens for response")),
	)

	// Query by pattern tool
	queryByPatternTool := mcp.NewTool("query_by_pattern",
		mcp.WithDescription(
			"Search for entities using glob or regex patterns "+
				"(supports wildcards *, ?, character classes [abc], brace expansion {a,b}, and regex /pattern/)",
		),
		mcp.WithString("pattern", mcp.Required(), mcp.Description("Search pattern (supports glob and regex patterns)")),
		mcp.WithString("entity_type", mcp.Description("Filter by entity type: function, type, variable, constant")),
		mcp.WithBoolean("include_callers", mcp.Description("Include functions that call matched functions")),
		mcp.WithBoolean("include_callees", mcp.Description("Include functions called by matched functions")),
		mcp.WithBoolean("include_types", mcp.Description("Include related type definitions")),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens for response")),
	)

	// Get call graph tool
	getCallGraphTool := mcp.NewTool("get_call_graph",
		mcp.WithDescription("Get detailed call graph for a function"),
		mcp.WithString("function_name", mcp.Required(), mcp.Description("Function name to analyze")),
		mcp.WithNumber("max_depth", mcp.Description("Maximum traversal depth (default: 2)")),
		mcp.WithBoolean("include_callers", mcp.Description("Include functions that call this function")),
		mcp.WithBoolean("include_callees", mcp.Description("Include functions called by this function")),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens for response")),
	)

	// List functions tool
	listFunctionsTool := mcp.NewTool("list_functions",
		mcp.WithDescription("List all functions in the repository with optional filtering and pagination"),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens for response (default: 2000)")),
		mcp.WithBoolean("include_signatures", mcp.Description("Include function signatures in the response (default: true)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of functions to return (0 for no limit)")),
		mcp.WithNumber("offset", mcp.Description("Number of functions to skip (for pagination)")),
	)

	// List types tool
	listTypesTool := mcp.NewTool("list_types",
		mcp.WithDescription("List all types in the repository with optional filtering and pagination"),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens for response (default: 2000)")),
		mcp.WithBoolean("include_signatures", mcp.Description("Include type signatures in the response (default: true)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of types to return (0 for no limit)")),
		mcp.WithNumber("offset", mcp.Description("Number of types to skip (for pagination)")),
	)

	return []mcp.Tool{queryByNameTool, queryByPatternTool, getCallGraphTool, listFunctionsTool, listTypesTool}
}

// RegisterRepoTools registers the repository management MCP tools
func (s *RepoContextMCPServer) RegisterRepoTools() {
	// Implementation will be added as we develop the tools
}

// HandleQueryByName handles the query_by_name MCP tool
func (s *RepoContextMCPServer) HandleQueryByName(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check for system-level failures that prevent operation
	if s.QueryEngine == nil {
		return nil, fmt.Errorf("query engine not initialized - system configuration error")
	}

	// Validate repository first
	if err := s.validateRepository(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Repository validation failed: %v", err)), nil
	}

	// Parse parameters using the MCP library helper functions
	name := request.GetString("name", "")
	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	includeCallers := request.GetBool("include_callers", false)
	includeCallees := request.GetBool("include_callees", false)
	includeTypes := request.GetBool("include_types", false)
	maxTokens := request.GetInt("max_tokens", constMaxTokens)

	// Build query options
	queryOptions := index.QueryOptions{
		IncludeCallers: includeCallers,
		IncludeCallees: includeCallees,
		IncludeTypes:   includeTypes,
		MaxTokens:      maxTokens,
		Format:         "json",
	}

	// Execute the query using the query engine
	searchResult, err := s.QueryEngine.SearchByNameWithOptions(name, queryOptions)
	if err != nil {
		return s.FormatErrorResponse("query_by_name", err), nil
	}

	// Return the formatted result
	return s.FormatSuccessResponse(searchResult), nil
}

// validateEntityType validates the entity_type parameter
func (s *RepoContextMCPServer) validateEntityType(entityType string) error {
	if entityType == "" {
		return nil // Empty is valid (no filter)
	}

	validEntityTypes := []string{"function", "type", "variable", "constant"}
	for _, validType := range validEntityTypes {
		if entityType == validType {
			return nil
		}
	}

	return fmt.Errorf("invalid entity_type '%s', must be one of: function, type, variable, constant", entityType)
}

// executePatternSearchWithFilter executes pattern search with optional entity type filtering
func (s *RepoContextMCPServer) executePatternSearchWithFilter(
	pattern, entityType string,
	queryOptions index.QueryOptions,
) (*index.SearchResult, error) {
	searchResult, err := s.QueryEngine.SearchByPatternWithOptions(pattern, queryOptions)
	if err != nil {
		return nil, err
	}

	// Apply entity type filter if specified
	if entityType != "" {
		filteredEntries := make([]index.SearchResultEntry, 0)
		for _, entry := range searchResult.Entries {
			if entry.IndexEntry.Type == entityType {
				filteredEntries = append(filteredEntries, entry)
			}
		}
		searchResult.Entries = filteredEntries
	}

	return searchResult, nil
}

// HandleQueryByPattern handles the query_by_pattern MCP tool
func (s *RepoContextMCPServer) HandleQueryByPattern(
	_ context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	// Check for system-level failures that prevent operation
	if s.QueryEngine == nil {
		return nil, fmt.Errorf("query engine not initialized - system configuration error")
	}

	// Validate repository first
	if err := s.validateRepository(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Repository validation failed: %v", err)), nil
	}

	// Parse parameters using the MCP library helper functions
	pattern := request.GetString("pattern", "")
	if pattern == "" {
		return mcp.NewToolResultError("pattern parameter is required"), nil
	}

	entityType := request.GetString("entity_type", "")
	includeCallers := request.GetBool("include_callers", false)
	includeCallees := request.GetBool("include_callees", false)
	includeTypes := request.GetBool("include_types", false)
	maxTokens := request.GetInt("max_tokens", constMaxTokens)

	// Validate entity type if provided
	if err := s.validateEntityType(entityType); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Build query options
	queryOptions := index.QueryOptions{
		IncludeCallers: includeCallers,
		IncludeCallees: includeCallees,
		IncludeTypes:   includeTypes,
		MaxTokens:      maxTokens,
		Format:         "json",
	}

	// Execute the pattern search with optional filtering
	searchResult, err := s.executePatternSearchWithFilter(pattern, entityType, queryOptions)
	if err != nil {
		return s.FormatErrorResponse("query_by_pattern", err), nil
	}

	// Return the formatted result
	return s.FormatSuccessResponse(searchResult), nil
}

// HandleGetCallGraph handles the get_call_graph MCP tool
func (s *RepoContextMCPServer) HandleGetCallGraph(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check for system-level failures that prevent operation
	if s.QueryEngine == nil {
		return nil, fmt.Errorf("query engine not initialized - system configuration error")
	}

	// Validate repository first
	if err := s.validateRepository(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Repository validation failed: %v", err)), nil
	}

	// Parse parameters using the MCP library helper functions
	functionName := request.GetString("function_name", "")
	if functionName == "" {
		return mcp.NewToolResultError("function_name parameter is required"), nil
	}

	maxDepth := request.GetInt("max_depth", constMaxDepth)
	if maxDepth <= 0 {
		maxDepth = constMaxDepth
	}

	includeCallers := request.GetBool("include_callers", true) // Default to true
	includeCallees := request.GetBool("include_callees", true) // Default to true
	maxTokens := request.GetInt("max_tokens", constMaxTokens)

	// Build query options
	queryOptions := index.QueryOptions{
		IncludeCallers: includeCallers,
		IncludeCallees: includeCallees,
		MaxDepth:       maxDepth,
		MaxTokens:      maxTokens,
		Format:         "json",
	}

	// Execute the call graph query using the query engine
	callGraphResult, err := s.QueryEngine.GetCallGraphWithOptions(functionName, queryOptions)
	if err != nil {
		return s.FormatErrorResponse("get_call_graph", err), nil
	}

	// Return the formatted result
	return s.FormatSuccessResponse(callGraphResult), nil
}

// handleListEntities is a common helper for listing entities by type (functions, types, etc.)
func (s *RepoContextMCPServer) handleListEntities(
	request mcp.CallToolRequest,
	entityType, toolName string,
) (*mcp.CallToolResult, error) {
	// Check for system-level failures that prevent operation
	if s.QueryEngine == nil {
		return nil, fmt.Errorf("query engine not initialized - system configuration error")
	}

	// Validate repository first
	if err := s.validateRepository(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Repository validation failed: %v", err)), nil
	}

	// Parse parameters using the MCP library helper functions
	maxTokens := request.GetInt("max_tokens", constMaxTokens)
	includeSignatures := request.GetBool("include_signatures", true)
	limit := request.GetInt("limit", 0)   // 0 means no limit
	offset := request.GetInt("offset", 0) // 0 means no offset

	// Build query options
	queryOptions := index.QueryOptions{
		MaxTokens: maxTokens,
		Format:    "json",
	}

	// Search for all entities of the specified type using the query engine
	searchResult, err := s.QueryEngine.SearchByTypeWithOptions(entityType, queryOptions)
	if err != nil {
		return s.FormatErrorResponse(toolName, err), nil
	}

	// Apply pagination if requested
	if limit > 0 || offset > 0 {
		s.applyPagination(searchResult, limit, offset)
	}

	// Filter out signatures if not requested
	if !includeSignatures {
		s.removeSignatures(searchResult)
	}

	// Return the formatted result
	return s.FormatSuccessResponse(searchResult), nil
}

// HandleListFunctions handles the list_functions MCP tool
func (s *RepoContextMCPServer) HandleListFunctions(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleListEntities(request, "function", "list_functions")
}

// HandleListTypes handles the list_types MCP tool
func (s *RepoContextMCPServer) HandleListTypes(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleListEntities(request, "type", "list_types")
}

// applyPagination applies limit and offset to search results
func (s *RepoContextMCPServer) applyPagination(result *index.SearchResult, limit, offset int) {
	totalEntries := len(result.Entries)

	// Apply offset
	if offset > 0 {
		if offset >= totalEntries {
			result.Entries = []index.SearchResultEntry{}
			return
		}
		result.Entries = result.Entries[offset:]
	}

	// Apply limit
	if limit > 0 && limit < len(result.Entries) {
		result.Entries = result.Entries[:limit]
		result.Truncated = true
	}
}

// removeSignatures removes signature information from search results when not requested
func (s *RepoContextMCPServer) removeSignatures(result *index.SearchResult) {
	for i := range result.Entries {
		result.Entries[i].IndexEntry.Signature = ""
	}
}

// FormatSuccessResponse formats a successful response for MCP
func (s *RepoContextMCPServer) FormatSuccessResponse(data interface{}) *mcp.CallToolResult {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err))
	}
	return mcp.NewToolResultText(string(jsonData))
}

// FormatErrorResponse formats an error response for MCP
func (s *RepoContextMCPServer) FormatErrorResponse(operation string, err error) *mcp.CallToolResult {
	errorMsg := fmt.Sprintf("Operation '%s' failed: %v", operation, err)
	return mcp.NewToolResultError(errorMsg)
}
