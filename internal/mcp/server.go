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

// RepoContextMCPServer provides MCP server functionality for repository context protocol
type RepoContextMCPServer struct {
	queryEngine *index.QueryEngine
	storage     *index.HybridStorage
	repoPath    string
}

// NewRepoContextMCPServer creates a new MCP server instance
func NewRepoContextMCPServer() *RepoContextMCPServer {
	return &RepoContextMCPServer{}
}

// Run starts the MCP server with JSON-RPC over stdin/stdout
func (s *RepoContextMCPServer) Run(ctx context.Context) error {
	// For now, return an error if no proper setup
	return fmt.Errorf("server setup incomplete - stdin/stdout JSON-RPC not yet fully implemented")
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
	if s.repoPath == "" {
		return fmt.Errorf("repository path not set")
	}

	repoContextPath := filepath.Join(s.repoPath, ".repocontext")
	if _, err := os.Stat(repoContextPath); os.IsNotExist(err) {
		return fmt.Errorf("repository not initialized - .repocontext directory not found")
	}

	// Initialize storage
	storage := index.NewHybridStorage(repoContextPath)
	if err := storage.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	s.storage = storage
	s.queryEngine = index.NewQueryEngine(storage)

	return nil
}

// validateRepository checks if the repository is properly initialized
func (s *RepoContextMCPServer) validateRepository() error {
	if s.repoPath == "" {
		return fmt.Errorf("no repository path configured")
	}

	repoContextPath := filepath.Join(s.repoPath, ".repocontext")
	if _, err := os.Stat(repoContextPath); os.IsNotExist(err) {
		return fmt.Errorf("repository not initialized - run initialize_repository first")
	}

	return nil
}

// registerQueryTools registers the query-related MCP tools
func (s *RepoContextMCPServer) registerQueryTools() {
	// Implementation will be added as we develop the tools
}

// registerRepoTools registers the repository management MCP tools
func (s *RepoContextMCPServer) registerRepoTools() {
	// Implementation will be added as we develop the tools
}

// TODO: to complete
// handleQueryByName handles the query_by_name MCP tool
// func (s *RepoContextMCPServer) handleQueryByName(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
// 	// Validate repository first
// 	if err := s.validateRepository(); err != nil {
// 		return mcp.NewToolResultError(fmt.Sprintf("Repository validation failed: %v", err)), nil
// 	}

// 	// Parse parameters
// 	name := mcp.ParseString(request, "name", "")
// 	if name == "" {
// 		return mcp.NewToolResultError("name parameter is required"), nil
// 	}

// 	// TODO: Implement actual query logic
// 	result := map[string]interface{}{
// 		"query":   name,
// 		"status":  "not_implemented",
// 		"message": "Query functionality not yet implemented",
// 	}

// 	return s.formatSuccessResponse(result), nil
// }

// TODO: to complete
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

// TODO: to complete
// handleInitializeRepository handles the initialize_repository MCP tool
// func (s *RepoContextMCPServer) handleInitializeRepository(
// 	ctx context.Context, request mcp.CallToolRequest
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
func (s *RepoContextMCPServer) formatSuccessResponse(data interface{}) *mcp.CallToolResult {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err))
	}
	return mcp.NewToolResultText(string(jsonData))
}

// formatErrorResponse formats an error response for MCP
func (s *RepoContextMCPServer) formatErrorResponse(operation string, err error) *mcp.CallToolResult {
	errorMsg := fmt.Sprintf("Operation '%s' failed: %v", operation, err)
	return mcp.NewToolResultError(errorMsg)
}
