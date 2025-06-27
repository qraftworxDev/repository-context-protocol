package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"repository-context-protocol/internal/index"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	constDuration50ms      = 50
	ConstFilePermission600 = 0600
	constFilePermission755 = 0755
)

// RegisterAdvancedQueryTools registers enhanced query tools with advanced features
func (s *RepoContextMCPServer) RegisterAdvancedQueryTools() []mcp.Tool {
	return []mcp.Tool{
		s.createQueryByNameTool(),
		s.createQueryByPatternTool(),
		s.createGetCallGraphTool(),
		s.createListFunctionsTool(),
		s.createListTypesTool(),
	}
}

// createQueryByNameTool creates the enhanced query_by_name tool with advanced parameter handling
func (s *RepoContextMCPServer) createQueryByNameTool() mcp.Tool {
	return mcp.NewTool("query_by_name",
		mcp.WithDescription("Search for functions, types, or variables by exact name with advanced options"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name to search for")),
		mcp.WithBoolean("include_callers", mcp.Description("Include functions that call this function")),
		mcp.WithBoolean("include_callees", mcp.Description("Include functions called by this function")),
		mcp.WithBoolean("include_types", mcp.Description("Include related type definitions")),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens for response (default: 2000)")),
	)
}

// createQueryByPatternTool creates the enhanced query_by_pattern tool with advanced pattern support
func (s *RepoContextMCPServer) createQueryByPatternTool() mcp.Tool {
	return mcp.NewTool("query_by_pattern",
		mcp.WithDescription(
			"Search for entities using glob or regex patterns with advanced filtering "+
				"(supports wildcards *, ?, character classes [abc], brace expansion {a,b}, and regex /pattern/). "+
				"This tool is useful for searching for entities in the repository.",
		),
		mcp.WithString("pattern", mcp.Required(), mcp.Description("Search pattern (supports glob and regex patterns)")),
		mcp.WithString("entity_type", mcp.Description("Filter by entity type: function, type, variable, constant")),
		mcp.WithBoolean("include_callers", mcp.Description("Include functions that call matched functions")),
		mcp.WithBoolean("include_callees", mcp.Description("Include functions called by matched functions")),
		mcp.WithBoolean("include_types", mcp.Description("Include related type definitions")),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens for response (default: 2000)")),
	)
}

// createGetCallGraphTool creates the enhanced get_call_graph tool with depth control
func (s *RepoContextMCPServer) createGetCallGraphTool() mcp.Tool {
	return mcp.NewTool("get_call_graph",
		mcp.WithDescription("Get detailed call graph for a function with configurable depth and selective inclusion"),
		mcp.WithString("function_name", mcp.Required(), mcp.Description("Function name to analyze")),
		mcp.WithNumber("max_depth", mcp.Description("Maximum traversal depth (default: 2)")),
		mcp.WithBoolean("include_callers", mcp.Description("Include functions that call this function")),
		mcp.WithBoolean("include_callees", mcp.Description("Include functions called by this function")),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens for response (default: 2000)")),
	)
}

// createListFunctionsTool creates the enhanced list_functions tool with pagination
func (s *RepoContextMCPServer) createListFunctionsTool() mcp.Tool {
	return mcp.NewTool("list_functions",
		mcp.WithDescription("List all functions in the repository with pagination and signature control"),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens for response (default: 2000)")),
		mcp.WithBoolean("include_signatures", mcp.Description("Include function signatures in the response (default: true)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of functions to return (0 for no limit)")),
		mcp.WithNumber("offset", mcp.Description("Number of functions to skip (for pagination)")),
	)
}

// createListTypesTool creates the enhanced list_types tool with pagination
func (s *RepoContextMCPServer) createListTypesTool() mcp.Tool {
	return mcp.NewTool("list_types",
		mcp.WithDescription("List all types in the repository with pagination and signature control"),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens for response (default: 2000)")),
		mcp.WithBoolean("include_signatures", mcp.Description("Include type signatures in the response (default: true)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of types to return (0 for no limit)")),
		mcp.WithNumber("offset", mcp.Description("Number of types to skip (for pagination)")),
	)
}

// Advanced Parameter Handling

// parseQueryByNameParameters extracts and validates parameters for query_by_name with enhanced handling
func (s *RepoContextMCPServer) parseQueryByNameParameters(request mcp.CallToolRequest) (*QueryByNameParams, error) {
	name := request.GetString("name", "")
	if name == "" {
		return nil, fmt.Errorf("name parameter is required")
	}

	return &QueryByNameParams{
		Name:           name,
		IncludeCallers: request.GetBool("include_callers", false),
		IncludeCallees: request.GetBool("include_callees", false),
		IncludeTypes:   request.GetBool("include_types", false),
		MaxTokens:      request.GetInt("max_tokens", constMaxTokens),
	}, nil
}

// parseQueryByPatternParameters extracts and validates parameters for query_by_pattern with enhanced handling
func (s *RepoContextMCPServer) parseQueryByPatternParameters(request mcp.CallToolRequest) (*QueryByPatternParams, error) {
	pattern := request.GetString("pattern", "")
	if pattern == "" {
		return nil, fmt.Errorf("pattern parameter is required")
	}

	entityType := request.GetString("entity_type", "")
	if err := s.validateEntityType(entityType); err != nil {
		return nil, err
	}

	return &QueryByPatternParams{
		Pattern:        pattern,
		EntityType:     entityType,
		IncludeCallers: request.GetBool("include_callers", false),
		IncludeCallees: request.GetBool("include_callees", false),
		IncludeTypes:   request.GetBool("include_types", false),
		MaxTokens:      request.GetInt("max_tokens", constMaxTokens),
	}, nil
}

// parseGetCallGraphParameters extracts and validates parameters for get_call_graph with enhanced handling
func (s *RepoContextMCPServer) parseGetCallGraphParameters(request mcp.CallToolRequest) (*GetCallGraphParams, error) {
	functionName := request.GetString("function_name", "")
	if functionName == "" {
		return nil, fmt.Errorf("function_name parameter is required")
	}

	return &GetCallGraphParams{
		FunctionName:   functionName,
		MaxDepth:       request.GetInt("max_depth", constMaxDepth),
		IncludeCallers: request.GetBool("include_callers", false),
		IncludeCallees: request.GetBool("include_callees", false),
		MaxTokens:      request.GetInt("max_tokens", constMaxTokens),
	}, nil
}

// Query Options Integration

// buildQueryOptionsFromParams creates QueryOptions from parsed parameters with optimizations
func (s *RepoContextMCPServer) buildQueryOptionsFromParams(params QueryOptionsBuilder) index.QueryOptions {
	return index.QueryOptions{
		IncludeCallers: params.GetIncludeCallers(),
		IncludeCallees: params.GetIncludeCallees(),
		IncludeTypes:   params.GetIncludeTypes(),
		MaxTokens:      params.GetMaxTokens(),
		Format:         "json",
	}
}

// Advanced Tool Handlers with Enhanced Error Handling

// HandleAdvancedQueryByName provides enhanced query_by_name with better parameter handling
func (s *RepoContextMCPServer) HandleAdvancedQueryByName(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// System-level validation
	if s.QueryEngine == nil {
		return nil, fmt.Errorf("query engine not initialized - system configuration error")
	}

	// Repository validation
	if err := s.validateRepository(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Repository validation failed: %v", err)), nil
	}

	// Enhanced parameter parsing
	params, err := s.parseQueryByNameParameters(request)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Parameter validation failed: %v", err)), nil
	}

	// Query options integration
	queryOptions := s.buildQueryOptionsFromParams(params)

	// Execute query with enhanced error handling
	searchResult, err := s.QueryEngine.SearchByNameWithOptions(params.Name, queryOptions)
	if err != nil {
		return s.FormatErrorResponse("query_by_name", err), nil
	}

	// Response optimization
	return s.FormatSuccessResponse(searchResult), nil
}

// HandleAdvancedQueryByPattern provides enhanced query_by_pattern with better filtering
func (s *RepoContextMCPServer) HandleAdvancedQueryByPattern(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// System-level validation
	if s.QueryEngine == nil {
		return nil, fmt.Errorf("query engine not initialized - system configuration error")
	}

	// Repository validation
	if err := s.validateRepository(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Repository validation failed: %v", err)), nil
	}

	// Enhanced parameter parsing
	params, err := s.parseQueryByPatternParameters(request)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Parameter validation failed: %v", err)), nil
	}

	// Query options integration
	queryOptions := s.buildQueryOptionsFromParams(params)

	// Execute pattern search with filtering
	searchResult, err := s.executePatternSearchWithFilter(
		params.Pattern,
		params.EntityType,
		queryOptions,
	)
	if err != nil {
		return s.FormatErrorResponse("query_by_pattern", err), nil
	}

	// Response optimization
	return s.FormatSuccessResponse(searchResult), nil
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

// HandleAdvancedGetCallGraph provides enhanced get_call_graph with better parameter handling
func (s *RepoContextMCPServer) HandleAdvancedGetCallGraph(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// System-level validation
	if s.QueryEngine == nil {
		return nil, fmt.Errorf("query engine not initialized - system configuration error")
	}

	// Repository validation
	if err := s.validateRepository(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Repository validation failed: %v", err)), nil
	}

	// Enhanced parameter parsing
	params, err := s.parseGetCallGraphParameters(request)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Parameter validation failed: %v", err)), nil
	}

	// Query options integration
	queryOptions := s.buildQueryOptionsFromParams(params)
	queryOptions.MaxDepth = params.MaxDepth

	// Execute call graph query with enhanced error handling
	callGraphResult, err := s.QueryEngine.GetCallGraphWithOptions(params.FunctionName, queryOptions)
	if err != nil {
		return s.FormatErrorResponse("get_call_graph", err), nil
	}

	// Response optimization
	return s.FormatSuccessResponse(callGraphResult), nil
}

// HandleAdvancedListFunctions provides enhanced list_functions with better parameter handling
func (s *RepoContextMCPServer) HandleAdvancedListFunctions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// System-level validation
	if s.QueryEngine == nil {
		return nil, fmt.Errorf("query engine not initialized - system configuration error")
	}

	// Repository validation
	if err := s.validateRepository(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Repository validation failed: %v", err)), nil
	}

	// Enhanced parameter parsing
	params := s.parseListEntitiesParameters(request)

	// Execute list functions with enhanced error handling
	return s.executeListEntitiesWithParams("function", "list_functions", params)
}

// HandleAdvancedListTypes provides enhanced list_types with better parameter handling
func (s *RepoContextMCPServer) HandleAdvancedListTypes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// System-level validation
	if s.QueryEngine == nil {
		return nil, fmt.Errorf("query engine not initialized - system configuration error")
	}

	// Repository validation
	if err := s.validateRepository(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Repository validation failed: %v", err)), nil
	}

	// Enhanced parameter parsing
	params := s.parseListEntitiesParameters(request)

	// Execute list types with enhanced error handling
	return s.executeListEntitiesWithParams("type", "list_types", params)
}

// parseListEntitiesParameters extracts and validates parameters for list operations
func (s *RepoContextMCPServer) parseListEntitiesParameters(request mcp.CallToolRequest) *ListEntitiesParams {
	return &ListEntitiesParams{
		MaxTokens:         request.GetInt("max_tokens", constMaxTokens),
		IncludeSignatures: request.GetBool("include_signatures", true),
		Limit:             request.GetInt("limit", 0),
		Offset:            request.GetInt("offset", 0),
	}
}

// executeListEntitiesWithParams executes list operations with enhanced parameter handling
func (s *RepoContextMCPServer) executeListEntitiesWithParams(
	entityType,
	toolName string,
	params *ListEntitiesParams,
) (*mcp.CallToolResult, error) {
	// Build query options
	queryOptions := index.QueryOptions{
		MaxTokens: params.MaxTokens,
		Format:    "json",
	}

	// Search for all entities of the specified type using the query engine
	searchResult, err := s.QueryEngine.SearchByTypeWithOptions(entityType, queryOptions)
	if err != nil {
		return s.FormatErrorResponse(toolName, err), nil
	}

	// Apply pagination if requested
	if params.Limit > 0 || params.Offset > 0 {
		s.applyPagination(searchResult, params.Limit, params.Offset)
	}

	// Filter out signatures if not requested
	if !params.IncludeSignatures {
		s.removeSignatures(searchResult)
	}

	// Return the formatted result
	return s.FormatSuccessResponse(searchResult), nil
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

// Parameter Types for Enhanced Handling

// QueryByNameParams encapsulates query_by_name parameters with validation
type QueryByNameParams struct {
	Name           string
	IncludeCallers bool
	IncludeCallees bool
	IncludeTypes   bool
	MaxTokens      int
}

func (p *QueryByNameParams) GetIncludeCallers() bool { return p.IncludeCallers }
func (p *QueryByNameParams) GetIncludeCallees() bool { return p.IncludeCallees }
func (p *QueryByNameParams) GetIncludeTypes() bool   { return p.IncludeTypes }
func (p *QueryByNameParams) GetMaxTokens() int       { return p.MaxTokens }

// QueryByPatternParams encapsulates query_by_pattern parameters with validation
type QueryByPatternParams struct {
	Pattern        string
	EntityType     string
	IncludeCallers bool
	IncludeCallees bool
	IncludeTypes   bool
	MaxTokens      int
}

func (p *QueryByPatternParams) GetIncludeCallers() bool { return p.IncludeCallers }
func (p *QueryByPatternParams) GetIncludeCallees() bool { return p.IncludeCallees }
func (p *QueryByPatternParams) GetIncludeTypes() bool   { return p.IncludeTypes }
func (p *QueryByPatternParams) GetMaxTokens() int       { return p.MaxTokens }

// GetCallGraphParams encapsulates get_call_graph parameters with validation
type GetCallGraphParams struct {
	FunctionName   string
	MaxDepth       int
	IncludeCallers bool
	IncludeCallees bool
	MaxTokens      int
}

func (p *GetCallGraphParams) GetIncludeCallers() bool { return p.IncludeCallers }
func (p *GetCallGraphParams) GetIncludeCallees() bool { return p.IncludeCallees }
func (p *GetCallGraphParams) GetIncludeTypes() bool   { return false } // Not applicable for call graph
func (p *GetCallGraphParams) GetMaxTokens() int       { return p.MaxTokens }

// QueryOptionsBuilder interface for building query options from different parameter types
type QueryOptionsBuilder interface {
	GetIncludeCallers() bool
	GetIncludeCallees() bool
	GetIncludeTypes() bool
	GetMaxTokens() int
}

// ListEntitiesParams encapsulates list entity parameters with validation
type ListEntitiesParams struct {
	MaxTokens         int
	IncludeSignatures bool
	Limit             int
	Offset            int
}

func (p *ListEntitiesParams) GetIncludeCallers() bool { return false } // Not applicable for list operations
func (p *ListEntitiesParams) GetIncludeCallees() bool { return false } // Not applicable for list operations
func (p *ListEntitiesParams) GetIncludeTypes() bool   { return false } // Not applicable for list operations
func (p *ListEntitiesParams) GetMaxTokens() int       { return p.MaxTokens }

// Repository Management Tools - Phase 2.2 Implementation

// RegisterRepositoryManagementTools registers repository management tools
func (s *RepoContextMCPServer) RegisterRepositoryManagementTools() []mcp.Tool {
	return []mcp.Tool{
		s.createInitializeRepositoryTool(),
		s.createBuildIndexTool(),
		s.createGetRepositoryStatusTool(),
	}
}

// createInitializeRepositoryTool creates the initialize_repository tool
func (s *RepoContextMCPServer) createInitializeRepositoryTool() mcp.Tool {
	return mcp.NewTool("initialize_repository",
		mcp.WithDescription("Initialize a repository for semantic indexing by creating .repocontext directory structure"),
		mcp.WithString("path", mcp.Description("Path to repository directory (default: current directory)")),
	)
}

// createBuildIndexTool creates the build_index tool
func (s *RepoContextMCPServer) createBuildIndexTool() mcp.Tool {
	return mcp.NewTool("build_index",
		mcp.WithDescription("Build semantic index for the repository by parsing source code files and creating searchable chunks"),
		mcp.WithString("path", mcp.Description("Path to repository directory (default: current directory)")),
		mcp.WithBoolean("verbose", mcp.Description("Enable verbose output with detailed build statistics")),
	)
}

// createGetRepositoryStatusTool creates the get_repository_status tool
func (s *RepoContextMCPServer) createGetRepositoryStatusTool() mcp.Tool {
	return mcp.NewTool("get_repository_status",
		mcp.WithDescription(
			"Get current repository indexing status and comprehensive statistics including index size, entity counts, and build information. "+
				"This tool is useful for monitoring the repository's indexing progress and ensuring it is up to date."),
		mcp.WithString("path", mcp.Description("Path to repository directory (default: current directory)")),
	)
}

// HandleInitializeRepository handles the initialize_repository tool request
func (s *RepoContextMCPServer) HandleInitializeRepository(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse parameters
	params := s.parseInitializeRepositoryParameters(request)

	// Determine target path
	targetPath, err := s.determineInitializationPath(params.Path)
	if err != nil {
		return s.FormatErrorResponse("initialize_repository", err), nil
	}

	// Validate path exists
	if err = s.validateInitializationPath(targetPath); err != nil {
		return s.FormatErrorResponse("initialize_repository", err), nil
	}

	// Perform initialization
	result, err := s.initializeRepositoryStructure(targetPath)
	if err != nil {
		return s.FormatErrorResponse("initialize_repository", err), nil
	}

	// Return success response
	return s.FormatSuccessResponse(result), nil
}

// parseInitializeRepositoryParameters extracts and validates parameters for initialize_repository
func (s *RepoContextMCPServer) parseInitializeRepositoryParameters(request mcp.CallToolRequest) *InitializeRepositoryParams {
	path := request.GetString("path", "")

	return &InitializeRepositoryParams{
		Path: path,
	}
}

// determineInitializationPath determines the actual path to initialize
func (s *RepoContextMCPServer) determineInitializationPath(providedPath string) (string, error) {
	if providedPath == "" {
		// Use current directory
		currentDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		return currentDir, nil
	}

	// Use provided path
	return filepath.Abs(providedPath)
}

// validateInitializationPath validates that the path is suitable for initialization
func (s *RepoContextMCPServer) validateInitializationPath(path string) error {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Check if path is a directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to check path: %w", err)
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	return nil
}

// initializeRepositoryStructure creates the repository structure and returns initialization result
func (s *RepoContextMCPServer) initializeRepositoryStructure(path string) (*InitializationResult, error) {
	repoContextPath := filepath.Join(path, ".repocontext")

	// Check if already initialized
	if _, err := os.Stat(repoContextPath); err == nil {
		return &InitializationResult{
			Path:               path,
			RepoContextPath:    repoContextPath,
			AlreadyInitialized: true,
			Message:            "Repository already initialized",
			CreatedDirectories: []string{},
		}, nil
	}

	// Create .repocontext directory
	if err := os.MkdirAll(repoContextPath, constFilePermission755); err != nil {
		return nil, fmt.Errorf("failed to create .repocontext directory: %w", err)
	}

	createdDirs := []string{repoContextPath}

	// Create chunks subdirectory
	chunksPath := filepath.Join(repoContextPath, "chunks")
	if err := os.MkdirAll(chunksPath, constFilePermission755); err != nil {
		return nil, fmt.Errorf("failed to create chunks directory: %w", err)
	}

	createdDirs = append(createdDirs, chunksPath)

	// Create manifest.json file (basic structure)
	manifestPath := filepath.Join(repoContextPath, "manifest.json")
	if err := s.createInitialManifest(manifestPath); err != nil {
		return nil, fmt.Errorf("failed to create initial manifest: %w", err)
	}

	return &InitializationResult{
		Path:               path,
		RepoContextPath:    repoContextPath,
		AlreadyInitialized: false,
		Message:            "Repository initialized successfully",
		CreatedDirectories: createdDirs,
		CreatedFiles:       []string{manifestPath},
	}, nil
}

// createInitialManifest creates the initial manifest.json file
func (s *RepoContextMCPServer) createInitialManifest(manifestPath string) error {
	manifest := map[string]interface{}{
		"version":     "1.0",
		"created_at":  time.Now().UTC().Format(time.RFC3339),
		"description": "Repository context index manifest",
	}

	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, manifestJSON, ConstFilePermission600); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}

// HandleBuildIndex handles the build_index tool request
func (s *RepoContextMCPServer) HandleBuildIndex(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse parameters
	params := s.parseBuildIndexParameters(request)

	// Determine target path
	targetPath, err := s.determineBuildPath(params.Path)
	if err != nil {
		return s.FormatErrorResponse("build_index", err), nil
	}

	// Validate repository is initialized
	if err = s.validateRepositoryForBuild(targetPath); err != nil {
		return s.FormatErrorResponse("build_index", err), nil
	}

	// Perform index build
	result, err := s.buildRepositoryIndex(targetPath, params.Verbose)
	if err != nil {
		return s.FormatErrorResponse("build_index", err), nil
	}

	// Return success response
	return s.FormatSuccessResponse(result), nil
}

// parseBuildIndexParameters extracts and validates parameters for build_index
func (s *RepoContextMCPServer) parseBuildIndexParameters(request mcp.CallToolRequest) *BuildIndexParams {
	path := request.GetString("path", "")
	verbose := request.GetBool("verbose", false)

	return &BuildIndexParams{
		Path:    path,
		Verbose: verbose,
	}
}

// determineBuildPath determines the actual path for index building
func (s *RepoContextMCPServer) determineBuildPath(providedPath string) (string, error) {
	if providedPath == "" {
		// Use current directory
		currentDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		return currentDir, nil
	}

	// Use provided path
	return filepath.Abs(providedPath)
}

// validateRepositoryForBuild validates that the repository is initialized and ready for build
func (s *RepoContextMCPServer) validateRepositoryForBuild(path string) error {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Check if path is a directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to check path: %w", err)
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Check if repository is initialized
	repoContextPath := filepath.Join(path, ".repocontext")
	if _, err := os.Stat(repoContextPath); os.IsNotExist(err) {
		return fmt.Errorf("repository not initialized - run initialize_repository first")
	}

	// Check if manifest.json exists
	manifestPath := filepath.Join(repoContextPath, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("manifest.json not found - repository may be corrupted, try initialize_repository")
	}

	return nil
}

// buildRepositoryIndex performs the actual index building and returns build result
func (s *RepoContextMCPServer) buildRepositoryIndex(path string, verbose bool) (*BuildIndexResult, error) {
	// Create and initialize the IndexBuilder
	builder := index.NewIndexBuilder(path)
	if err := builder.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize index builder: %w", err)
	}
	defer func() {
		if closeErr := builder.Close(); closeErr != nil {
			// Log warning but don't fail the operation
			log.Printf("Failed to close index builder: %v", closeErr)
		}
	}()

	// Build the index
	stats, err := builder.BuildIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to build index: %w", err)
	}

	// Create result with build statistics
	result := &BuildIndexResult{
		Path:             path,
		Success:          true,
		Message:          "Index built successfully",
		FilesProcessed:   stats.FilesProcessed,
		FunctionsIndexed: stats.FunctionsIndexed,
		TypesIndexed:     stats.TypesIndexed,
		VariablesIndexed: stats.VariablesIndexed,
		ConstantsIndexed: stats.ConstantsIndexed,
		CallsIndexed:     stats.CallsIndexed,
		Duration:         stats.Duration,
		Verbose:          verbose,
	}

	return result, nil
}

// BuildIndexParams holds parameters for build_index
type BuildIndexParams struct {
	Path    string
	Verbose bool
}

// BuildIndexResult holds the result of index building
type BuildIndexResult struct {
	Path             string        `json:"path"`
	Success          bool          `json:"success"`
	Message          string        `json:"message"`
	FilesProcessed   int           `json:"files_processed"`
	FunctionsIndexed int           `json:"functions_indexed"`
	TypesIndexed     int           `json:"types_indexed"`
	VariablesIndexed int           `json:"variables_indexed"`
	ConstantsIndexed int           `json:"constants_indexed"`
	CallsIndexed     int           `json:"calls_indexed"`
	Duration         time.Duration `json:"duration"`
	Verbose          bool          `json:"verbose"`
}

// InitializeRepositoryParams holds parameters for initialize_repository
type InitializeRepositoryParams struct {
	Path string
}

// InitializationResult holds the result of repository initialization
type InitializationResult struct {
	Path               string   `json:"path"`
	RepoContextPath    string   `json:"repo_context_path"`
	AlreadyInitialized bool     `json:"already_initialized"`
	Message            string   `json:"message"`
	CreatedDirectories []string `json:"created_directories"`
	CreatedFiles       []string `json:"created_files,omitempty"`
}

// HandleGetRepositoryStatus handles the get_repository_status tool request
func (s *RepoContextMCPServer) HandleGetRepositoryStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse parameters
	params := s.parseGetRepositoryStatusParameters(request)

	// Determine target path
	targetPath, err := s.determineStatusPath(params.Path)
	if err != nil {
		return s.FormatErrorResponse("get_repository_status", err), nil
	}

	// Collect repository status and statistics
	status, err := s.collectRepositoryStatus(targetPath)
	if err != nil {
		return s.FormatErrorResponse("get_repository_status", err), nil
	}

	// Return success response
	return s.FormatSuccessResponse(status), nil
}

// parseGetRepositoryStatusParameters extracts and validates parameters for get_repository_status
func (s *RepoContextMCPServer) parseGetRepositoryStatusParameters(request mcp.CallToolRequest) *GetRepositoryStatusParams {
	path := request.GetString("path", "")

	return &GetRepositoryStatusParams{
		Path: path,
	}
}

// determineStatusPath determines the actual path for status checking
func (s *RepoContextMCPServer) determineStatusPath(providedPath string) (string, error) {
	if providedPath == "" {
		// Use current directory
		currentDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		return currentDir, nil
	}

	// Use provided path
	return filepath.Abs(providedPath)
}

// collectRepositoryStatus gathers comprehensive repository status and statistics
func (s *RepoContextMCPServer) collectRepositoryStatus(path string) (*RepositoryStatus, error) {
	// Validate path exists and is directory
	if err := s.validateStatusPath(path); err != nil {
		return nil, err
	}

	status := &RepositoryStatus{
		Path:    path,
		Message: "",
	}

	// Check if repository is initialized
	repoContextPath := filepath.Join(path, ".repocontext")
	if _, err := os.Stat(repoContextPath); os.IsNotExist(err) {
		status.IsInitialized = false
		status.IsIndexed = false
		status.Message = "Repository not initialized - run initialize_repository first"
		status.Statistics = &RepositoryStatistics{}
		return status, nil
	}

	status.IsInitialized = true

	// Check if repository is indexed
	indexPath := filepath.Join(repoContextPath, "index.db")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		status.IsIndexed = false
		status.Message = "Repository initialized but not indexed - run build_index to create index"
		status.Statistics = &RepositoryStatistics{}
		return status, nil
	}

	status.IsIndexed = true
	status.Message = "Repository fully initialized and indexed"

	// Collect detailed statistics
	statistics, err := s.collectDetailedStatistics(path, repoContextPath, indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to collect statistics: %w", err)
	}

	status.Statistics = statistics

	return status, nil
}

// validateStatusPath validates that the path is suitable for status checking
func (s *RepoContextMCPServer) validateStatusPath(path string) error {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Check if path is a directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to check path: %w", err)
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	return nil
}

// collectDetailedStatistics gathers comprehensive statistics about the repository index
func (s *RepoContextMCPServer) collectDetailedStatistics(repoPath, repoContextPath, indexPath string) (*RepositoryStatistics, error) {
	statistics := &RepositoryStatistics{}

	statistics.RepositoryPath = repoPath
	statistics.IndexPath = indexPath
	statistics.RepoContextPath = repoContextPath
	statistics.ManifestPath = filepath.Join(repoContextPath, "manifest.json")

	// Get index file size
	indexInfo, err := os.Stat(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get index file info: %w", err)
	}
	statistics.IndexSize = indexInfo.Size()

	// Get index modification time (last build time)
	statistics.LastBuildTime = indexInfo.ModTime()

	// Load manifest for additional information
	manifestPath := filepath.Join(repoContextPath, "manifest.json")
	var manifestInfo os.FileInfo
	if manifestInfo, err = os.Stat(manifestPath); err == nil {
		statistics.ManifestSize = manifestInfo.Size()
		statistics.InitializedTime = manifestInfo.ModTime()
	}

	// Initialize storage to collect entity statistics
	storage := index.NewHybridStorage(repoContextPath)
	if err = storage.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize storage for statistics: %w", err)
	}
	defer storage.Close()

	// Use query engine to get entity counts
	queryEngine := index.NewQueryEngine(storage)

	// Get function count
	functionResult, err := queryEngine.SearchByType("function")
	if err == nil && functionResult != nil {
		statistics.FunctionsIndexed = len(functionResult.Entries)
	}

	// Get type count by aggregating all type kinds
	typeKinds := []string{"struct", "interface", "type", "alias", "enum"}
	totalTypes := 0
	var typeResult *index.SearchResult
	for _, typeKind := range typeKinds {
		typeResult, err = queryEngine.SearchByType(typeKind)
		if err == nil && typeResult != nil {
			totalTypes += len(typeResult.Entries)
		}
	}
	statistics.TypesIndexed = totalTypes

	// Get variable count
	variableResult, err := queryEngine.SearchByType("variable")
	if err == nil && variableResult != nil {
		statistics.VariablesIndexed = len(variableResult.Entries)
	}

	// Get constant count
	constantResult, err := queryEngine.SearchByType("constant")
	if err == nil && constantResult != nil {
		statistics.ConstantsIndexed = len(constantResult.Entries)
	}

	// Count source files processed (approximation based on chunks directory)
	chunksPath := filepath.Join(repoContextPath, "chunks")
	if chunksInfo, err := os.ReadDir(chunksPath); err == nil {
		statistics.FilesProcessed = len(chunksInfo)
	}

	// Calculate last build duration based on file timestamps
	if !statistics.LastBuildTime.IsZero() && !statistics.InitializedTime.IsZero() {
		// Calculate absolute time difference between build and initialization
		var duration time.Duration
		if statistics.LastBuildTime.After(statistics.InitializedTime) {
			duration = statistics.LastBuildTime.Sub(statistics.InitializedTime)
		} else {
			duration = statistics.InitializedTime.Sub(statistics.LastBuildTime)
		}

		// Ensure we have a meaningful duration (minimum based on entities processed)
		totalEntities := statistics.FunctionsIndexed + statistics.TypesIndexed +
			statistics.VariablesIndexed + statistics.ConstantsIndexed
		if totalEntities > 0 {
			// Use actual duration if reasonable, otherwise estimate based on entities
			minEstimatedDuration := time.Duration(totalEntities) * constDuration50ms * time.Millisecond
			if duration > minEstimatedDuration {
				statistics.LastBuildDuration = duration
			} else {
				statistics.LastBuildDuration = minEstimatedDuration
			}
		} else {
			statistics.LastBuildDuration = time.Second // Default minimum
		}
	}

	return statistics, nil
}

// GetRepositoryStatusParams represents parameters for get_repository_status tool
type GetRepositoryStatusParams struct {
	Path string
}

// RepositoryStatus represents the complete status of a repository
type RepositoryStatus struct {
	Path          string                `json:"path"`
	IsInitialized bool                  `json:"is_initialized"`
	IsIndexed     bool                  `json:"is_indexed"`
	Message       string                `json:"message"`
	Statistics    *RepositoryStatistics `json:"statistics"`
}

// RepositoryStatistics represents detailed statistics about the repository index
type RepositoryStatistics struct {
	FilesProcessed    int           `json:"files_processed"`
	FunctionsIndexed  int           `json:"functions_indexed"`
	TypesIndexed      int           `json:"types_indexed"`
	VariablesIndexed  int           `json:"variables_indexed"`
	ConstantsIndexed  int           `json:"constants_indexed"`
	CallsIndexed      int           `json:"calls_indexed"`
	IndexSize         int64         `json:"index_size"`
	ManifestSize      int64         `json:"manifest_size"`
	LastBuildTime     time.Time     `json:"last_build_time"`
	LastBuildDuration time.Duration `json:"last_build_duration"`
	InitializedTime   time.Time     `json:"initialized_time"`

	// Additional fields for detailed statistics
	RepositoryPath  string `json:"repository_path"`
	IndexPath       string `json:"index_path"`
	RepoContextPath string `json:"repo_context_path"`
	ManifestPath    string `json:"manifest_path"`
}
