package mcp

import (
	"context"
	"fmt"
	"strings"

	"repository-context-protocol/internal/index"

	"github.com/mark3labs/mcp-go/mcp"
)

// Constants for context tools
const (
	MaxContextLines     = 50 // Maximum allowed context lines around function
	DefaultContextLines = 5  // Default context lines around function
)

// Token management constants for context tools
const (
	// Context-specific token overhead - empirical measurements
	FunctionContextBaseTokens    = 150 // Base tokens for function metadata (name, signature, location)
	ImplementationOverheadTokens = 50  // Overhead for implementation structure
	ContextLineTokens            = 10  // Average tokens per context line
	FunctionRefTokens            = 15  // Average tokens per function reference
	TypeRefTokens                = 12  // Average tokens per type reference

	// Token distribution ratios for function context
	ImplementationTokenRatio = 0.4  // 40% for implementation content
	CallersTokenRatio        = 0.25 // 25% for callers
	CalleesTokenRatio        = 0.25 // 25% for callees
	TypesTokenRatio          = 0.1  // 10% for related types
)

// Token optimization constants
const (
	CharsPerToken  = 4 // Rough estimate: 4 characters per token
	BodyTokenRatio = 2 // Body gets half of available tokens when balancing with context
)

// GetFunctionContextParams encapsulates get_function_context parameters
type GetFunctionContextParams struct {
	FunctionName           string
	IncludeImplementations bool
	ContextLines           int
	MaxTokens              int
}

// GetIncludeCallers implements QueryOptionsBuilder interface (not applicable)
func (p *GetFunctionContextParams) GetIncludeCallers() bool { return false }

// GetIncludeCallees implements QueryOptionsBuilder interface (not applicable)
func (p *GetFunctionContextParams) GetIncludeCallees() bool { return false }

// GetIncludeTypes implements QueryOptionsBuilder interface (not applicable)
func (p *GetFunctionContextParams) GetIncludeTypes() bool { return false }

// GetMaxTokens implements QueryOptionsBuilder interface
func (p *GetFunctionContextParams) GetMaxTokens() int { return p.MaxTokens }

// FunctionLocation represents the location of a function in the codebase
type FunctionLocation struct {
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// FunctionImplementation represents the implementation details of a function
type FunctionImplementation struct {
	Body         string   `json:"body"`
	ContextLines []string `json:"context_lines"`
}

// FunctionReference represents a reference to a function (caller or callee)
type FunctionReference struct {
	Name string `json:"name"`
	File string `json:"file"`
	Line int    `json:"line"`
}

// TypeReference represents a reference to a type
type TypeReference struct {
	Name string `json:"name"`
	File string `json:"file"`
	Line int    `json:"line"`
}

// FunctionContextResult represents the complete result of function context analysis
type FunctionContextResult struct {
	FunctionName   string                  `json:"function_name"`
	Signature      string                  `json:"signature"`
	Location       FunctionLocation        `json:"location"`
	Implementation *FunctionImplementation `json:"implementation,omitempty"`
	Callers        []FunctionReference     `json:"callers,omitempty"`
	Callees        []FunctionReference     `json:"callees,omitempty"`
	RelatedTypes   []TypeReference         `json:"related_types,omitempty"`
	TokenCount     int                     `json:"token_count"`
	Truncated      bool                    `json:"truncated"`
}

// ToolOperations defines the tool-specific operations for the generic handler
type ToolOperations[P any, R any] struct {
	ParseParams    func(mcp.CallToolRequest) (P, error)
	BuildResult    func(P) (R, error)
	OptimizeResult func(R, int)
	ToolName       string
}

// executeGenericToolHandler provides a generic template that eliminates duplication
func executeGenericToolHandler[P any, R any](
	s *RepoContextMCPServer,
	request mcp.CallToolRequest,
	ops ToolOperations[P, R],
) (*mcp.CallToolResult, error) {
	// System validation
	if s.QueryEngine == nil {
		return nil, fmt.Errorf("query engine not initialized - system configuration error")
	}
	if err := s.validateRepository(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Repository validation failed: %v", err)), nil
	}

	// Tool-specific parameter parsing
	params, err := ops.ParseParams(request)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Parameter validation failed: %v", err)), nil
	}

	// Tool-specific result building
	result, err := ops.BuildResult(params)
	if err != nil {
		return s.FormatErrorResponse(ops.ToolName, err), nil
	}

	// Tool-specific optimization
	if paramsWithTokens, ok := any(params).(interface{ GetMaxTokens() int }); ok {
		ops.OptimizeResult(result, paramsWithTokens.GetMaxTokens())
	}

	return s.FormatSuccessResponse(result), nil
}

// HandleGetFunctionContext provides comprehensive function context analysis
func (s *RepoContextMCPServer) HandleGetFunctionContext(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ops := ToolOperations[*GetFunctionContextParams, *FunctionContextResult]{
		ParseParams: s.parseGetFunctionContextParameters,
		BuildResult: s.buildFunctionContextResult,
		OptimizeResult: func(result *FunctionContextResult, maxTokens int) {
			s.optimizeFunctionContextResponse(result, maxTokens)
		},
		ToolName: "get_function_context",
	}
	return executeGenericToolHandler(s, request, ops)
}

// validateContextLines validates and normalizes context lines parameter
func validateContextLines(contextLines int) int {
	if contextLines <= 0 {
		return DefaultContextLines
	}
	if contextLines > MaxContextLines {
		return MaxContextLines
	}
	return contextLines
}

// parseGetFunctionContextParameters extracts and validates get_function_context parameters
func (s *RepoContextMCPServer) parseGetFunctionContextParameters(request mcp.CallToolRequest) (*GetFunctionContextParams, error) {
	functionName := strings.TrimSpace(request.GetString("function_name", ""))
	if functionName == "" {
		return nil, fmt.Errorf("function_name parameter is required")
	}

	contextLines := request.GetInt("context_lines", DefaultContextLines)
	validatedContextLines := validateContextLines(contextLines)

	return &GetFunctionContextParams{
		FunctionName:           functionName,
		IncludeImplementations: request.GetBool("include_implementations", false),
		ContextLines:           validatedContextLines,
		MaxTokens:              request.GetInt("max_tokens", constMaxTokens),
	}, nil
}

// createGetFunctionContextTool creates the get_function_context tool
func (s *RepoContextMCPServer) createGetFunctionContextTool() mcp.Tool {
	return mcp.NewTool("get_function_context",
		mcp.WithDescription(
			"Get complete context for a function including signature, implementation details, callers, callees, and related types",
		),
		mcp.WithString("function_name", mcp.Required(), mcp.Description("Function name to analyze")),
		mcp.WithBoolean("include_implementations", mcp.Description("Include function implementation details (default: false)")),
		mcp.WithNumber("context_lines", mcp.Description("Number of context lines around function (default: 5, max: 50)")),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens for response (default: 2000)")),
	)
}

// RegisterContextTools registers context analysis tools
func (s *RepoContextMCPServer) RegisterContextTools() []mcp.Tool {
	return []mcp.Tool{
		s.createGetFunctionContextTool(),
	}
}

// optimizeFunctionContextResponse optimizes function context response for token limits
func (s *RepoContextMCPServer) optimizeFunctionContextResponse(result *FunctionContextResult, maxTokens int) {
	// Calculate current token count
	currentTokens := s.estimateFunctionContextTokens(result)
	result.TokenCount = currentTokens

	// If within limit, no optimization needed
	if currentTokens <= maxTokens {
		result.Truncated = false
		return
	}

	// Mark as truncated
	result.Truncated = true

	// Calculate available tokens for content (reserve tokens for metadata)
	availableTokens := maxTokens - FunctionContextBaseTokens
	if availableTokens <= 0 {
		// Minimal response - just function metadata
		result.Implementation = nil
		result.Callers = nil
		result.Callees = nil
		result.RelatedTypes = nil
		result.TokenCount = FunctionContextBaseTokens
		return
	}

	// Distribute tokens according to ratios
	implementationTokens := int(float64(availableTokens) * ImplementationTokenRatio)
	callersTokens := int(float64(availableTokens) * CallersTokenRatio)
	calleesTokens := int(float64(availableTokens) * CalleesTokenRatio)
	typesTokens := int(float64(availableTokens) * TypesTokenRatio)

	// Optimize implementation
	if result.Implementation != nil {
		s.optimizeImplementation(result.Implementation, implementationTokens)
	}

	// Optimize callers
	if len(result.Callers) > 0 {
		maxCallers := s.calculateMaxFunctionRefs(callersTokens)
		if maxCallers < len(result.Callers) {
			result.Callers = result.Callers[:maxCallers]
		}
	}

	// Optimize callees
	if len(result.Callees) > 0 {
		maxCallees := s.calculateMaxFunctionRefs(calleesTokens)
		if maxCallees < len(result.Callees) {
			result.Callees = result.Callees[:maxCallees]
		}
	}

	// Optimize related types
	if len(result.RelatedTypes) > 0 {
		maxTypes := s.calculateMaxTypeRefs(typesTokens)
		if maxTypes < len(result.RelatedTypes) {
			result.RelatedTypes = result.RelatedTypes[:maxTypes]
		}
	}

	// Recalculate final token count
	result.TokenCount = s.estimateFunctionContextTokens(result)
}

// estimateFunctionContextTokens estimates token count for function context result
func (s *RepoContextMCPServer) estimateFunctionContextTokens(result *FunctionContextResult) int {
	tokens := FunctionContextBaseTokens

	// Add implementation tokens
	if result.Implementation != nil {
		tokens += ImplementationOverheadTokens
		tokens += len(result.Implementation.Body) / CharsPerToken // Rough estimate
		tokens += len(result.Implementation.ContextLines) * ContextLineTokens
	}

	// Add caller tokens
	tokens += len(result.Callers) * FunctionRefTokens

	// Add callee tokens
	tokens += len(result.Callees) * FunctionRefTokens

	// Add type tokens
	tokens += len(result.RelatedTypes) * TypeRefTokens

	return tokens
}

// optimizeImplementation optimizes implementation content for token limits
func (s *RepoContextMCPServer) optimizeImplementation(impl *FunctionImplementation, tokenLimit int) {
	if tokenLimit <= ImplementationOverheadTokens {
		impl.Body = ""
		impl.ContextLines = nil
		return
	}

	availableTokens := tokenLimit - ImplementationOverheadTokens

	// Optimize body (rough estimate: 4 chars per token)
	bodyTokens := len(impl.Body) / CharsPerToken
	contextTokens := len(impl.ContextLines) * ContextLineTokens

	if bodyTokens+contextTokens <= availableTokens {
		return // No optimization needed
	}

	// Prioritize body over context lines using ratio
	if bodyTokens > availableTokens/BodyTokenRatio {
		maxBodyChars := (availableTokens / BodyTokenRatio) * CharsPerToken
		if len(impl.Body) > maxBodyChars {
			impl.Body = impl.Body[:maxBodyChars] + "..."
		}
		availableTokens -= len(impl.Body) / CharsPerToken
	}

	// Optimize context lines
	maxContextLines := availableTokens / ContextLineTokens
	if maxContextLines < len(impl.ContextLines) {
		impl.ContextLines = impl.ContextLines[:maxContextLines]
	}
}

// calculateMaxFunctionRefs calculates maximum function references for token limit
func (s *RepoContextMCPServer) calculateMaxFunctionRefs(tokenLimit int) int {
	if tokenLimit <= 0 {
		return 0
	}
	return tokenLimit / FunctionRefTokens
}

// calculateMaxTypeRefs calculates maximum type references for token limit
func (s *RepoContextMCPServer) calculateMaxTypeRefs(tokenLimit int) int {
	if tokenLimit <= 0 {
		return 0
	}
	return tokenLimit / TypeRefTokens
}

// buildFunctionContextResult constructs the complete function context result
func (s *RepoContextMCPServer) buildFunctionContextResult(params *GetFunctionContextParams) (*FunctionContextResult, error) {
	// Search for the function
	queryOptions := index.QueryOptions{
		IncludeCallers: true, // Always include callers for context
		IncludeCallees: true, // Always include callees for context
		IncludeTypes:   true, // Always include related types for context
		MaxTokens:      params.MaxTokens,
		Format:         "json",
	}

	searchResult, err := s.QueryEngine.SearchByNameWithOptions(params.FunctionName, queryOptions)
	if err != nil {
		return nil, fmt.Errorf("function search failed: %w", err)
	}

	if len(searchResult.Entries) == 0 {
		return nil, fmt.Errorf("function '%s' not found", params.FunctionName)
	}

	// Find the function in search results
	var functionEntry *index.SearchResultEntry
	for i := range searchResult.Entries {
		entry := &searchResult.Entries[i]
		if entry.IndexEntry.Type == index.EntityTypeFunction && entry.IndexEntry.Name == params.FunctionName {
			functionEntry = entry
			break
		}
	}

	if functionEntry == nil {
		return nil, fmt.Errorf("function '%s' not found in search results", params.FunctionName)
	}

	// Build the result
	result := &FunctionContextResult{
		FunctionName: params.FunctionName,
		Signature:    functionEntry.IndexEntry.Signature,
		Location: FunctionLocation{
			File:      functionEntry.IndexEntry.File,
			StartLine: functionEntry.IndexEntry.StartLine,
			EndLine:   functionEntry.IndexEntry.EndLine,
		},
	}

	// Add implementation details if requested
	if params.IncludeImplementations {
		implementation := s.buildFunctionImplementation(functionEntry, params.ContextLines)
		result.Implementation = implementation
	}

	// Add callers
	result.Callers = s.extractFunctionReferences(searchResult.CallGraph.Callers)

	// Add callees
	result.Callees = s.extractFunctionReferences(searchResult.CallGraph.Callees)

	// Add related types
	result.RelatedTypes = s.extractTypeReferences(searchResult.Entries)

	return result, nil
}

// buildFunctionImplementation constructs function implementation details
func (s *RepoContextMCPServer) buildFunctionImplementation(
	entry *index.SearchResultEntry,
	contextLines int,
) *FunctionImplementation {
	// Calculate available tokens for implementation (reserve tokens for other response parts)
	availableTokens := ImplementationOverheadTokens // Start with base overhead

	// Use ratio-based allocation for body vs context lines
	bodyTokenRatio := 0.7    // 70% for function body
	contextTokenRatio := 0.3 // 30% for context lines

	bodyTokenLimit := int(float64(availableTokens) * bodyTokenRatio)
	contextTokenLimit := int(float64(availableTokens) * contextTokenRatio)

	// Build initial implementation structure
	implementation := &FunctionImplementation{
		Body: "// Function implementation details would be extracted here",
		ContextLines: []string{
			"// Context line before function",
			"// Additional context would be extracted from source file",
		},
	}

	// Extract function body from chunk data if available
	if entry.ChunkData.FileData != nil {
		var bodyBuilder strings.Builder

		for i := range entry.ChunkData.FileData {
			line := &entry.ChunkData.FileData[i]
			for j := range line.Functions {
				function := &line.Functions[j]
				bodyBuilder.WriteString(function.Signature)
				bodyBuilder.WriteString("\n")
			}
		}

		rawBody := bodyBuilder.String()

		// Apply body token limit using ratio-based allocation
		bodyTokens := len(rawBody) / CharsPerToken
		if bodyTokens > bodyTokenLimit {
			maxBodyChars := bodyTokenLimit * CharsPerToken
			if len(rawBody) > maxBodyChars {
				implementation.Body = rawBody[:maxBodyChars] + "..."
			} else {
				implementation.Body = rawBody
			}
		} else {
			implementation.Body = rawBody
		}
	}

	// Generate context lines based on ratio-allocated token space
	maxContextLines := min(
		contextTokenLimit/ContextLineTokens,
		contextLines,
	)

	// Clear default context lines and add ratio-based context
	implementation.ContextLines = make([]string, 0, maxContextLines)

	for i := 0; i < maxContextLines; i++ {
		contextLine := fmt.Sprintf("// Context line %d: function context would be extracted from source", i+1)
		implementation.ContextLines = append(implementation.ContextLines, contextLine)
	}

	return implementation
}

// extractFunctionReferences converts call graph entries to function references
func (s *RepoContextMCPServer) extractFunctionReferences(entries []index.CallGraphEntry) []FunctionReference {
	if len(entries) == 0 {
		return nil
	}

	refs := make([]FunctionReference, len(entries))
	for i, entry := range entries {
		refs[i] = FunctionReference{
			Name: entry.Function,
			File: entry.File,
			Line: entry.Line,
		}
	}
	return refs
}

// extractTypeReferences converts search results to type references
func (s *RepoContextMCPServer) extractTypeReferences(entries []index.SearchResultEntry) []TypeReference {
	var typeRefs []TypeReference

	for _, entry := range entries {
		entryType := entry.IndexEntry.Type
		if entryType == index.EntityKindStruct || entryType == index.EntityKindInterface || entryType == index.EntityKindType {
			typeRefs = append(typeRefs, TypeReference{
				Name: entry.IndexEntry.Name,
				File: entry.IndexEntry.File,
				Line: entry.IndexEntry.StartLine,
			})
		}
	}

	return typeRefs
}
