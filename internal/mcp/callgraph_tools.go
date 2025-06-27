package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"repository-context-protocol/internal/index"

	"github.com/mark3labs/mcp-go/mcp"
)

// Constants for enhanced call graph tools
const (
	MaxCallGraphDepth     = 10          // Maximum allowed call graph depth
	DefaultCallGraphDepth = 2           // Default call graph depth
	ExternalCallPrefix    = "external:" // Prefix for external calls
)

// Dependency type constants
const (
	CalleeCallers = "callees"
	CallerCallees = "callers"
	Both          = "both"
)

// Entity type constants
const (
	TypeFunction = "function"
	TypeType     = "type"
)

// Token management constants with empirical rationale
const (
	// JSON structure overhead - reserve for JSON field names, brackets, metadata
	// Measured from typical responses: ~5% overhead for 2000 token responses
	JSONMetadataReserveTokens = 100

	// Per-entry overhead - measured from actual JSON serialization
	// Function entry: {"function":"name","file":"path","line":123,"chunk_data":{...}}
	FunctionEntryOverheadTokens = 20

	// Search entry: {"name":"...","type":"...","signature":"...","file":"...","start_line":123,"end_line":456,"chunk_data":{...}}
	// More complex structure requires higher overhead
	SearchEntryOverheadTokens = 30

	// Token distribution ratios for balanced analysis
	CallGraphCalleeTokenRatio  = 0.5  // 50% for callees in call graph (balanced caller/callee view)
	DependencyCalleeTokenRatio = 0.33 // 33% for callees in dependency analysis (3-way split: callees/callers/types)
	RemainingTokenSplitRatio   = 0.5  // 50% split for remaining tokens between callers and related types

	// Calculation constants
	PercentageMultiplier = 100 // Convert ratio to percentage
)

// EnhancedGetCallGraphParams encapsulates enhanced get_call_graph parameters
type EnhancedGetCallGraphParams struct {
	FunctionName    string
	MaxDepth        int
	IncludeCallers  bool
	IncludeCallees  bool
	IncludeExternal bool
	MaxTokens       int
}

// GetIncludeCallers implements QueryOptionsBuilder interface
func (p *EnhancedGetCallGraphParams) GetIncludeCallers() bool { return p.IncludeCallers }

// GetIncludeCallees implements QueryOptionsBuilder interface
func (p *EnhancedGetCallGraphParams) GetIncludeCallees() bool { return p.IncludeCallees }

// GetIncludeTypes implements QueryOptionsBuilder interface (not applicable for call graph)
func (p *EnhancedGetCallGraphParams) GetIncludeTypes() bool { return false }

// GetMaxTokens implements QueryOptionsBuilder interface
func (p *EnhancedGetCallGraphParams) GetMaxTokens() int { return p.MaxTokens }

// FindDependenciesParams encapsulates find_dependencies parameters
type FindDependenciesParams struct {
	EntityName     string
	DependencyType string // "callers", "callees", or "both"
	MaxTokens      int
}

// validateEnhancedCallGraphDepth validates and normalizes call graph depth
func validateEnhancedCallGraphDepth(depth int) int {
	if depth <= 0 {
		return DefaultCallGraphDepth
	}
	if depth > MaxCallGraphDepth {
		return MaxCallGraphDepth
	}
	return depth
}

// validateDependencyType validates the dependency type parameter
func validateDependencyType(depType string) error {
	switch strings.ToLower(depType) {
	case CalleeCallers, CallerCallees, Both, "":
		return nil
	default:
		return fmt.Errorf("invalid dependency_type '%s': must be '%s', '%s', or '%s'", depType, CalleeCallers, CallerCallees, Both)
	}
}

// parseEnhancedGetCallGraphParameters extracts and validates enhanced call graph parameters
func (s *RepoContextMCPServer) parseEnhancedGetCallGraphParameters(request mcp.CallToolRequest) (*EnhancedGetCallGraphParams, error) {
	functionName := strings.TrimSpace(request.GetString("function_name", ""))
	if functionName == "" {
		return nil, fmt.Errorf("function_name parameter is required")
	}

	maxDepth := request.GetInt("max_depth", DefaultCallGraphDepth)
	validatedDepth := validateEnhancedCallGraphDepth(maxDepth)

	return &EnhancedGetCallGraphParams{
		FunctionName:    functionName,
		MaxDepth:        validatedDepth,
		IncludeCallers:  request.GetBool("include_callers", false),
		IncludeCallees:  request.GetBool("include_callees", false),
		IncludeExternal: request.GetBool("include_external", false),
		MaxTokens:       request.GetInt("max_tokens", constMaxTokens),
	}, nil
}

// parseFindDependenciesParameters extracts and validates find dependencies parameters
func (s *RepoContextMCPServer) parseFindDependenciesParameters(request mcp.CallToolRequest) (*FindDependenciesParams, error) {
	entityName := strings.TrimSpace(request.GetString("entity_name", ""))
	if entityName == "" {
		return nil, fmt.Errorf("entity_name parameter is required")
	}

	dependencyType := strings.ToLower(strings.TrimSpace(request.GetString("dependency_type", "both")))
	if err := validateDependencyType(dependencyType); err != nil {
		return nil, err
	}

	return &FindDependenciesParams{
		EntityName:     entityName,
		DependencyType: dependencyType,
		MaxTokens:      request.GetInt("max_tokens", constMaxTokens),
	}, nil
}

// createEnhancedGetCallGraphTool creates the enhanced get_call_graph tool with external call filtering
func (s *RepoContextMCPServer) createEnhancedGetCallGraphTool() mcp.Tool {
	return mcp.NewTool("get_call_graph_enhanced",
		mcp.WithDescription(
			"Get detailed call graph for a function with enhanced depth control, external call filtering, and performance optimization",
		),
		mcp.WithString("function_name", mcp.Required(), mcp.Description("Function name to analyze")),
		mcp.WithNumber("max_depth", mcp.Description("Maximum traversal depth (default: 2, max: 10)")),
		mcp.WithBoolean("include_callers", mcp.Description("Include functions that call this function")),
		mcp.WithBoolean("include_callees", mcp.Description("Include functions called by this function")),
		mcp.WithBoolean("include_external", mcp.Description("Include external function calls (default: false)")),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens for response (default: 2000)")),
	)
}

// HandleEnhancedGetCallGraph provides enhanced get_call_graph with external filtering and performance optimization
func (s *RepoContextMCPServer) HandleEnhancedGetCallGraph(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// System-level validation
	if s.QueryEngine == nil {
		return nil, fmt.Errorf("query engine not initialized - system configuration error")
	}

	// Repository validation
	if err := s.validateRepository(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Repository validation failed: %v", err)), nil
	}

	// Enhanced parameter parsing
	params, err := s.parseEnhancedGetCallGraphParameters(request)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Parameter validation failed: %v", err)), nil
	}

	// Build query options with enhanced parameters
	queryOptions := s.buildQueryOptionsFromParams(params)
	queryOptions.MaxDepth = params.MaxDepth

	// Execute call graph query with enhanced error handling
	callGraphResult, err := s.QueryEngine.GetCallGraphWithOptions(params.FunctionName, queryOptions)
	if err != nil {
		return s.FormatErrorResponse("get_call_graph_enhanced", err), nil
	}

	// Apply external call filtering if requested
	if !params.IncludeExternal {
		s.filterExternalCalls(callGraphResult)
	}

	// Apply performance optimizations
	s.optimizeCallGraphResponse(callGraphResult, params.MaxTokens)

	// Response optimization
	return s.FormatSuccessResponse(callGraphResult), nil
}

// filterExternalCalls removes external calls from the call graph result
func (s *RepoContextMCPServer) filterExternalCalls(callGraph *index.CallGraphInfo) {
	if callGraph == nil {
		return
	}

	// Filter external callers
	filteredCallers := make([]index.CallGraphEntry, 0, len(callGraph.Callers))
	for _, caller := range callGraph.Callers {
		if !s.isExternalCall(caller.Function) {
			filteredCallers = append(filteredCallers, caller)
		}
	}
	callGraph.Callers = filteredCallers

	// Filter external callees
	filteredCallees := make([]index.CallGraphEntry, 0, len(callGraph.Callees))
	for _, callee := range callGraph.Callees {
		if !s.isExternalCall(callee.Function) {
			filteredCallees = append(filteredCallees, callee)
		}
	}
	callGraph.Callees = filteredCallees
}

// isExternalCall determines if a function call is external to the repository
func (s *RepoContextMCPServer) isExternalCall(functionName string) bool {
	// Check if function starts with external call prefix
	if strings.HasPrefix(functionName, ExternalCallPrefix) {
		return true
	}

	// Check if function is from standard library or external packages
	externalPrefixes := []string{
		"fmt.", "os.", "io.", "log.", "http.", "json.", "strings.", "strconv.",
		"time.", "context.", "sync.", "errors.", "path.", "filepath.",
		"github.com/", "golang.org/", "gopkg.in/",
	}

	for _, prefix := range externalPrefixes {
		if strings.HasPrefix(functionName, prefix) {
			return true
		}
	}

	return false
}

// optimizeCallGraphResponse applies performance optimizations to the call graph response
func (s *RepoContextMCPServer) optimizeCallGraphResponse(callGraph *index.CallGraphInfo, maxTokens int) {
	if callGraph == nil {
		return
	}

	// Estimate current token count
	currentTokens := s.estimateCallGraphTokens(callGraph)

	// If within limits, no optimization needed
	if currentTokens <= maxTokens {
		return
	}

	// Apply truncation strategy: prioritize callers over callees
	targetTokens := maxTokens - JSONMetadataReserveTokens // Reserve tokens for metadata

	// Truncate callees first if needed
	if len(callGraph.Callees) > 0 {
		maxCallees := s.calculateMaxEntries(callGraph.Callees, int(float64(targetTokens)*CallGraphCalleeTokenRatio))
		if maxCallees < len(callGraph.Callees) {
			callGraph.Callees = callGraph.Callees[:maxCallees]
		}
	}

	// Truncate callers if still over limit
	if len(callGraph.Callers) > 0 {
		remainingTokens := targetTokens - s.estimateEntriesTokens(callGraph.Callees)
		maxCallers := s.calculateMaxEntries(callGraph.Callers, remainingTokens)
		if maxCallers < len(callGraph.Callers) {
			callGraph.Callers = callGraph.Callers[:maxCallers]
		}
	}
}

// estimateCallGraphTokens estimates the token count for a call graph
func (s *RepoContextMCPServer) estimateCallGraphTokens(callGraph *index.CallGraphInfo) int {
	if callGraph == nil {
		return 0
	}

	tokens := 50 // Base metadata tokens
	tokens += s.estimateEntriesTokens(callGraph.Callers)
	tokens += s.estimateEntriesTokens(callGraph.Callees)

	return tokens
}

// estimateEntriesTokens estimates tokens for call graph entries
func (s *RepoContextMCPServer) estimateEntriesTokens(entries []index.CallGraphEntry) int {
	tokens := 0
	for _, entry := range entries {
		tokens += len(strings.Fields(entry.Function)) + FunctionEntryOverheadTokens // Function name + JSON overhead
		if entry.ChunkData != nil {
			tokens += entry.ChunkData.TokenCount
		}
	}
	return tokens
}

// calculateMaxEntries calculates maximum entries that fit within token limit
func (s *RepoContextMCPServer) calculateMaxEntries(entries []index.CallGraphEntry, tokenLimit int) int {
	currentTokens := 0
	maxEntries := 0

	for i, entry := range entries {
		entryTokens := len(strings.Fields(entry.Function)) + FunctionEntryOverheadTokens
		if entry.ChunkData != nil {
			entryTokens += entry.ChunkData.TokenCount
		}

		if currentTokens+entryTokens > tokenLimit {
			break
		}

		currentTokens += entryTokens
		maxEntries = i + 1
	}

	return maxEntries
}

// createFindDependenciesTool creates the find_dependencies tool
func (s *RepoContextMCPServer) createFindDependenciesTool() mcp.Tool {
	return mcp.NewTool("find_dependencies",
		mcp.WithDescription("Find all dependencies for a given function or type with configurable dependency type filtering"),
		mcp.WithString("entity_name", mcp.Required(), mcp.Description("Entity name to analyze (function or type)")),
		mcp.WithString("dependency_type", mcp.Description("Type of dependencies: 'callers', 'callees', or 'both' (default: 'both')")),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens for response (default: 2000)")),
	)
}

// HandleFindDependencies provides dependency analysis for functions and types
func (s *RepoContextMCPServer) HandleFindDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// System-level validation
	if s.QueryEngine == nil {
		return nil, fmt.Errorf("query engine not initialized - system configuration error")
	}

	// Repository validation
	if err := s.validateRepository(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Repository validation failed: %v", err)), nil
	}

	// Parameter parsing and validation
	params, err := s.parseFindDependenciesParameters(request)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Parameter validation failed: %v", err)), nil
	}

	// Build dependency analysis result
	dependencyResult, err := s.buildDependencyAnalysis(params)
	if err != nil {
		return s.FormatErrorResponse("find_dependencies", err), nil
	}

	// Apply token limits and optimization
	s.optimizeDependencyResponse(dependencyResult, params.MaxTokens)

	// Format and return result
	return s.FormatSuccessResponse(dependencyResult), nil
}

// DependencyAnalysisResult represents the result of dependency analysis
type DependencyAnalysisResult struct {
	EntityName        string                    `json:"entity_name"`
	EntityType        string                    `json:"entity_type"`
	DependencyType    string                    `json:"dependency_type"`
	Callers           []index.CallGraphEntry    `json:"callers,omitempty"`
	Callees           []index.CallGraphEntry    `json:"callees,omitempty"`
	RelatedTypes      []index.SearchResultEntry `json:"related_types,omitempty"`
	TotalCallers      int                       `json:"total_callers"`
	TotalCallees      int                       `json:"total_callees"`
	TotalRelatedTypes int                       `json:"total_related_types"`
	TokenCount        int                       `json:"token_count"`
	Truncated         bool                      `json:"truncated"`
}

// buildDependencyAnalysis builds comprehensive dependency analysis
func (s *RepoContextMCPServer) buildDependencyAnalysis(params *FindDependenciesParams) (*DependencyAnalysisResult, error) {
	result := &DependencyAnalysisResult{
		EntityName:     params.EntityName,
		DependencyType: params.DependencyType,
		Callers:        []index.CallGraphEntry{},
		Callees:        []index.CallGraphEntry{},
		RelatedTypes:   []index.SearchResultEntry{},
	}

	// First, determine the entity type by searching for it
	entitySearch, err := s.QueryEngine.SearchByNameWithOptions(params.EntityName, index.QueryOptions{
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find entity '%s': %w", params.EntityName, err)
	}

	if len(entitySearch.Entries) == 0 {
		return nil, fmt.Errorf("entity '%s' not found in repository", params.EntityName)
	}

	// Use the first matching entity
	entity := entitySearch.Entries[0]
	result.EntityType = entity.IndexEntry.Type

	// Build call graph analysis if it's a function
	if entity.IndexEntry.Type == TypeFunction {
		if err := s.addCallGraphDependencies(result, params); err != nil {
			return nil, err
		}
	}

	// Add related types analysis
	if err := s.addRelatedTypesDependencies(result, params); err != nil {
		return nil, err
	}

	// Calculate totals
	result.TotalCallers = len(result.Callers)
	result.TotalCallees = len(result.Callees)
	result.TotalRelatedTypes = len(result.RelatedTypes)

	return result, nil
}

// addCallGraphDependencies adds call graph dependencies to the analysis result
func (s *RepoContextMCPServer) addCallGraphDependencies(result *DependencyAnalysisResult, params *FindDependenciesParams) error {
	// Build query options based on dependency type
	queryOptions := index.QueryOptions{
		IncludeCallers: params.DependencyType == CallerCallees || params.DependencyType == Both,
		IncludeCallees: params.DependencyType == CalleeCallers || params.DependencyType == Both,
		MaxDepth:       DefaultCallGraphDepth,
		Format:         "json",
	}

	// Get call graph
	callGraph, err := s.QueryEngine.GetCallGraphWithOptions(params.EntityName, queryOptions)
	if err != nil {
		return fmt.Errorf("failed to get call graph for '%s': %w", params.EntityName, err)
	}

	// Add callers if requested
	if queryOptions.IncludeCallers {
		result.Callers = callGraph.Callers
	}

	// Add callees if requested
	if queryOptions.IncludeCallees {
		result.Callees = callGraph.Callees
	}

	return nil
}

// addRelatedTypesDependencies adds related types to the dependency analysis
func (s *RepoContextMCPServer) addRelatedTypesDependencies(result *DependencyAnalysisResult, params *FindDependenciesParams) error {
	// Search for types that might be related to this entity
	typeSearch, err := s.QueryEngine.SearchByNameWithOptions(params.EntityName, index.QueryOptions{
		IncludeTypes: true,
		Format:       "json",
	})
	if err != nil {
		return fmt.Errorf("failed to search for related types: %w", err)
	}

	// Filter for type entries
	for _, entry := range typeSearch.Entries {
		if entry.IndexEntry.Type == TypeType {
			result.RelatedTypes = append(result.RelatedTypes, entry)
		}
	}

	return nil
}

// optimizeDependencyResponse optimizes dependency analysis response for token limits
func (s *RepoContextMCPServer) optimizeDependencyResponse(result *DependencyAnalysisResult, maxTokens int) {
	// Estimate current tokens
	currentTokens := s.estimateDependencyTokens(result)
	result.TokenCount = currentTokens

	if currentTokens <= maxTokens {
		return
	}

	// Apply truncation strategy
	targetTokens := maxTokens - JSONMetadataReserveTokens // Reserve for metadata
	result.Truncated = true

	// Truncate in order of priority: callees, callers, related types
	if len(result.Callees) > 0 {
		maxCallees := s.calculateMaxEntries(result.Callees, int(float64(targetTokens)*DependencyCalleeTokenRatio))
		if maxCallees < len(result.Callees) {
			result.Callees = result.Callees[:maxCallees]
		}
	}

	if len(result.Callers) > 0 {
		remainingTokens := targetTokens - s.estimateEntriesTokens(result.Callees)
		maxCallers := s.calculateMaxEntries(result.Callers, int(float64(remainingTokens)*RemainingTokenSplitRatio))
		if maxCallers < len(result.Callers) {
			result.Callers = result.Callers[:maxCallers]
		}
	}

	if len(result.RelatedTypes) > 0 {
		remainingTokens := targetTokens - s.estimateEntriesTokens(result.Callees) - s.estimateEntriesTokens(result.Callers)
		maxTypes := s.calculateMaxSearchEntries(result.RelatedTypes, remainingTokens)
		if maxTypes < len(result.RelatedTypes) {
			result.RelatedTypes = result.RelatedTypes[:maxTypes]
		}
	}

	// Recalculate token count after optimization
	result.TokenCount = s.estimateDependencyTokens(result)
}

// estimateDependencyTokens estimates tokens for dependency analysis result
func (s *RepoContextMCPServer) estimateDependencyTokens(result *DependencyAnalysisResult) int {
	tokens := JSONMetadataReserveTokens // Base metadata tokens
	tokens += s.estimateEntriesTokens(result.Callers)
	tokens += s.estimateEntriesTokens(result.Callees)
	tokens += s.estimateSearchEntriesTokens(result.RelatedTypes)
	return tokens
}

// estimateSearchEntriesTokens estimates tokens for search result entries
func (s *RepoContextMCPServer) estimateSearchEntriesTokens(entries []index.SearchResultEntry) int {
	tokens := 0
	for _, entry := range entries {
		tokens += len(strings.Fields(entry.IndexEntry.Name)) +
			len(strings.Fields(entry.IndexEntry.Signature)) + SearchEntryOverheadTokens
		if entry.ChunkData != nil {
			tokens += entry.ChunkData.TokenCount
		}
	}
	return tokens
}

// calculateMaxSearchEntries calculates max search entries within token limit
func (s *RepoContextMCPServer) calculateMaxSearchEntries(entries []index.SearchResultEntry, tokenLimit int) int {
	currentTokens := 0
	maxEntries := 0

	for i, entry := range entries {
		entryTokens := len(strings.Fields(entry.IndexEntry.Name)) +
			len(strings.Fields(entry.IndexEntry.Signature)) + SearchEntryOverheadTokens
		if entry.ChunkData != nil {
			entryTokens += entry.ChunkData.TokenCount
		}

		if currentTokens+entryTokens > tokenLimit {
			break
		}

		currentTokens += entryTokens
		maxEntries = i + 1
	}

	return maxEntries
}

// measureActualTokens measures actual JSON serialization token count for validation
// This can be used to validate and tune our token estimation constants
func (s *RepoContextMCPServer) measureActualTokens(result interface{}) int {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		// Fallback to simple marshal if indent fails
		jsonData, _ = json.Marshal(result)
	}

	// Rough token approximation: split by whitespace and count words
	// This is a conservative estimate - actual tokenization may vary
	return len(strings.Fields(string(jsonData)))
}

// validateTokenEstimate compares estimated vs actual tokens for tuning
// This can be used during development to validate our token estimation accuracy
func (s *RepoContextMCPServer) validateTokenEstimate(estimated int, actual interface{}) (actualTokens int, accuracy float64) {
	actualTokens = s.measureActualTokens(actual)
	accuracy = float64(estimated) / float64(actualTokens)
	return actualTokens, accuracy
}

// optimizeWithActualMeasurement applies optimization using actual token measurement
// This provides more accurate optimization than estimates alone
func (s *RepoContextMCPServer) optimizeWithActualMeasurement(result interface{}, maxTokens int) bool {
	actualTokens := s.measureActualTokens(result)
	return actualTokens <= maxTokens
}

// logTokenEstimateAccuracy logs token estimation accuracy for monitoring
// This can be used to monitor and improve token estimation accuracy over time
//
//nolint:unused // Reserved for future logging implementation
func (s *RepoContextMCPServer) logTokenEstimateAccuracy(operation string, estimated, actual int) {
	accuracy := float64(estimated) / float64(actual) * PercentageMultiplier
	// In a real implementation, this would log to a proper logging system
	// For now, this is a placeholder for the pattern
	_ = fmt.Sprintf("Token estimate accuracy for %s: estimated=%d, actual=%d, accuracy=%.1f%%",
		operation, estimated, actual, accuracy)
}

// RegisterCallGraphTools registers enhanced call graph analysis tools
func (s *RepoContextMCPServer) RegisterCallGraphTools() []mcp.Tool {
	return []mcp.Tool{
		s.createEnhancedGetCallGraphTool(),
		s.createFindDependenciesTool(),
	}
}
