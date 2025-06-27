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
	constMaxTokens = 2000 // TODO: this should not be static - user defined from config
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
	// returning an error for now indicating WIP
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
	// TODO: this should be improved - what if it's an un-initialized repository?
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
	return s.RegisterAdvancedQueryTools()
}

// RegisterRepoTools registers the repository management MCP tools
func (s *RepoContextMCPServer) RegisterRepoTools() {
	// Implementation will be added as we develop the tools
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
