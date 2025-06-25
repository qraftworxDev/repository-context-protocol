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
	varMaxTokens = 2000
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

	return []mcp.Tool{queryByNameTool}
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
	maxTokens := request.GetInt("max_tokens", varMaxTokens)

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

// handleQueryByPattern handles the query_by_pattern MCP tool
// func (s *RepoContextMCPServer) handleQueryByPattern(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
// 	// Validate repository first
// 	if err := s.validateRepository(); err != nil {
// 		return mcp.NewToolResultError(fmt.Sprintf("Repository validation failed: %v", err)), nil
// 	}

// 	// Parse parameters
// 	pattern := mcp.ParseString(request, "pattern", "")
// 	if pattern == "" {
// 		return mcp.NewToolResultError("pattern parameter is required"), nil
// 	}

// 	// TODO: Implement actual pattern query logic
// 	result := map[string]interface{}{
// 		"query":   pattern,
// 		"status":  "not_implemented",
// 		"message": "Pattern query functionality not yet implemented",
// 	}

// 	return s.formatSuccessResponse(result), nil
// }

// // handleInitializeRepository handles the initialize_repository MCP tool
// func (s *RepoContextMCPServer) handleInitializeRepository(
// 	ctx context.Context, request mcp.CallToolRequest,
// ) (*mcp.CallToolResult, error) {
// 	// Parse parameters
// 	path := mcp.ParseString(request, "path", ".")

// 	// Resolve absolute path
// 	absPath, err := filepath.Abs(path)
// 	if err != nil {
// 		return mcp.NewToolResultError(fmt.Sprintf("Invalid path: %v", err)), nil
// 	}

// 	// TODO: Implement actual repository initialization
// 	result := map[string]interface{}{
// 		"path":    absPath,
// 		"status":  "not_implemented",
// 		"message": "Repository initialization not yet implemented",
// 	}

// 	return s.formatSuccessResponse(result), nil
// }

// formatSuccessResponse formats a successful response for MCP
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
