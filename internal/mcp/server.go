package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"repository-context-protocol/internal/index"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	constMaxTokens = 2000 // TODO: this should not be static - user defined from config
	constMaxDepth  = 2

	// Phase 4.1: Server Configuration Constants
	ServerName    = "repocontext"
	ServerVersion = "1.0.0"
)

// Phase 4.1: Server Configuration
type ServerConfiguration struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	MaxTokens int    `json:"max_tokens"`
	MaxDepth  int    `json:"max_depth"`
}

// RepoContextMCPServer provides MCP server functionality for repository context protocol
type RepoContextMCPServer struct {
	QueryEngine *index.QueryEngine
	Storage     *index.HybridStorage
	RepoPath    string
	server      *server.MCPServer
	// Phase 4.2: Error Recovery Manager
	errorRecoveryMgr *ErrorRecoveryManager
}

// NewRepoContextMCPServer creates a new MCP server instance
func NewRepoContextMCPServer() *RepoContextMCPServer {
	return &RepoContextMCPServer{
		// Phase 4.2: Initialize error recovery manager
		errorRecoveryMgr: NewErrorRecoveryManager(),
	}
}

// ============================================================================
// Phase 4.1: Enhanced Server Implementation
// ============================================================================

// GetServerCapabilities returns the server capabilities configuration
func (s *RepoContextMCPServer) GetServerCapabilities() map[string]interface{} {
	return map[string]interface{}{
		"tools": map[string]interface{}{
			"listChanged": true,
		},
		"experimental": map[string]interface{}{
			"callGraphAnalysis": true,
			"contextAnalysis":   true,
			"repoManagement":    true,
		},
	}
}

// GetClientCapabilities returns the client capabilities configuration
func (s *RepoContextMCPServer) GetClientCapabilities() map[string]interface{} {
	return map[string]interface{}{
		"experimental": map[string]interface{}{
			"streamingSupport": true,
			"batchProcessing":  true,
		},
		"sampling": true,
	}
}

// RegisterAllTools orchestrates registration of all tool categories
func (s *RepoContextMCPServer) RegisterAllTools() []mcp.Tool {
	var allTools []mcp.Tool

	// Register Advanced Query Tools
	allTools = append(allTools, s.RegisterAdvancedQueryTools()...)

	// Register Repository Management Tools
	allTools = append(allTools, s.RegisterRepositoryManagementTools()...)

	// Register Enhanced Call Graph Tools
	allTools = append(allTools, s.RegisterCallGraphTools()...)

	// Register Context Analysis Tools
	allTools = append(allTools, s.RegisterContextTools()...)

	return allTools
}

// CreateMCPServer creates and configures the MCP server instance
func (s *RepoContextMCPServer) CreateMCPServer() *server.MCPServer {
	mcpServer := server.NewMCPServer(
		ServerName,
		ServerVersion,
		server.WithToolCapabilities(true),
	)

	s.server = mcpServer
	return mcpServer
}

// SetupToolHandlers registers all tools with their handlers on the MCP server
func (s *RepoContextMCPServer) SetupToolHandlers(mcpServer *server.MCPServer) error {
	// Register all tools with their handlers
	allTools := s.RegisterAllTools()

	for i := range allTools {
		handler := s.getToolHandler(allTools[i].Name)
		if handler != nil {
			mcpServer.AddTool(allTools[i], handler)
		}
	}

	return nil
}

// getToolHandler returns the appropriate handler for a given tool name
func (s *RepoContextMCPServer) getToolHandler(toolName string) func(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	switch toolName {
	// Advanced Query Tools
	case "query_by_name":
		return s.HandleAdvancedQueryByName
	case "query_by_pattern":
		return s.HandleAdvancedQueryByPattern
	case "get_call_graph":
		return s.HandleAdvancedGetCallGraph
	case "list_functions":
		return s.HandleAdvancedListFunctions
	case "list_types":
		return s.HandleAdvancedListTypes

	// Repository Management Tools
	case "initialize_repository":
		return s.HandleInitializeRepository
	case "build_index":
		return s.HandleBuildIndex
	case "get_repository_status":
		return s.HandleGetRepositoryStatus

	// Enhanced Call Graph Tools
	case "get_call_graph_enhanced":
		return s.HandleEnhancedGetCallGraph
	case "find_dependencies":
		return s.HandleFindDependencies

	// Context Analysis Tools
	case "get_function_context":
		return s.HandleGetFunctionContext
	case "get_type_context":
		return s.HandleGetTypeContext

	default:
		return nil
	}
}

// InitializeWithContext performs enhanced context initialization
func (s *RepoContextMCPServer) InitializeWithContext(ctx context.Context) error {
	// Detect repository root
	repoPath, err := s.detectRepositoryRoot()
	if err != nil {
		return fmt.Errorf("failed to detect repository root: %w", err)
	}
	s.RepoPath = repoPath

	// Initialize query engine
	if err := s.initializeQueryEngine(); err != nil {
		return fmt.Errorf("failed to initialize query engine: %w", err)
	}

	return nil
}

// InitializeServerLifecycle performs complete server lifecycle initialization
func (s *RepoContextMCPServer) InitializeServerLifecycle(ctx context.Context) (*server.MCPServer, error) {
	// Create MCP server
	mcpServer := s.CreateMCPServer()

	// Setup tool handlers
	if err := s.SetupToolHandlers(mcpServer); err != nil {
		return mcpServer, fmt.Errorf("failed to setup tool handlers: %w", err)
	}

	// Initialize repository context (non-blocking with graceful degradation)
	if err := s.InitializeWithContext(ctx); err != nil {
		// Log the error but don't fail - server can still operate with limited functionality
		fmt.Fprintf(os.Stderr, "Warning: Repository initialization failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Server will continue with limited functionality\n")
	}

	return mcpServer, nil
}

// GetServerConfiguration returns the current server configuration
func (s *RepoContextMCPServer) GetServerConfiguration() *ServerConfiguration {
	return &ServerConfiguration{
		Name:      ServerName,
		Version:   ServerVersion,
		MaxTokens: constMaxTokens,
		MaxDepth:  constMaxDepth,
	}
}

// ============================================================================
// Enhanced Run Method - Phase 4.1 Complete Implementation
// ============================================================================

// Run starts the MCP server with enhanced lifecycle management
func (s *RepoContextMCPServer) Run(ctx context.Context) error {
	// Phase 4.1: Enhanced Server Lifecycle Management
	mcpServer, err := s.InitializeServerLifecycle(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize server lifecycle: %w", err)
	}

	// Start JSON-RPC server over stdin/stdout using the correct API
	return server.ServeStdio(mcpServer)
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

// ============================================================================
// Phase 4.2: Error Recovery Integration
// ============================================================================

// ExecuteToolWithRecovery executes a tool operation with circuit breaker and retry mechanisms
func (s *RepoContextMCPServer) ExecuteToolWithRecovery(
	ctx context.Context,
	toolName string,
	operation func() (*mcp.CallToolResult, error),
) (*mcp.CallToolResult, error) {
	if s.errorRecoveryMgr == nil {
		// Fallback to direct execution if error recovery manager is not initialized
		return operation()
	}

	return s.errorRecoveryMgr.ExecuteWithRecovery(ctx, toolName, operation)
}

// GetErrorRecoveryStats returns statistics about error recovery
func (s *RepoContextMCPServer) GetErrorRecoveryStats() map[string]interface{} {
	if s.errorRecoveryMgr == nil {
		return map[string]interface{}{
			"status": "not_initialized",
		}
	}

	return s.errorRecoveryMgr.GetRecoveryStats()
}
