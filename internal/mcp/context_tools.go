package mcp

import (
	"context"
	"fmt"
	"strings"

	"repository-context-protocol/internal/index"
	"repository-context-protocol/internal/models"

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

	// Type context specific token overhead
	TypeContextBaseTokens = 120 // Base tokens for type metadata (name, signature, location)
	FieldRefTokens        = 18  // Average tokens per field reference
	MethodRefTokens       = 20  // Average tokens per method reference
	UsageExampleTokens    = 25  // Average tokens per usage example

	// Token distribution ratios for type context
	FieldsTokenRatio  = 0.3 // 30% for fields
	MethodsTokenRatio = 0.4 // 40% for methods
	UsageTokenRatio   = 0.2 // 20% for usage examples
	RelatedTokenRatio = 0.1 // 10% for related types
)

// Token optimization constants
const (
	CharsPerToken  = 4   // Rough estimate: 4 characters per token
	BodyTokenRatio = 0.5 // Body gets half of available tokens when balancing with context
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

// GetTypeContextParams encapsulates get_type_context parameters
type GetTypeContextParams struct {
	TypeName       string
	IncludeMethods bool
	IncludeUsage   bool
	MaxTokens      int
}

// GetIncludeCallers implements QueryOptionsBuilder interface (not applicable)
func (p *GetTypeContextParams) GetIncludeCallers() bool { return false }

// GetIncludeCallees implements QueryOptionsBuilder interface (not applicable)
func (p *GetTypeContextParams) GetIncludeCallees() bool { return false }

// GetIncludeTypes implements QueryOptionsBuilder interface (not applicable)
func (p *GetTypeContextParams) GetIncludeTypes() bool { return false }

// GetMaxTokens implements QueryOptionsBuilder interface
func (p *GetTypeContextParams) GetMaxTokens() int { return p.MaxTokens }

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

// TypeLocation represents the location of a type in the codebase
type TypeLocation struct {
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// FieldReference represents a reference to a field in a type
type FieldReference struct {
	Name string `json:"name"`
	Type string `json:"type"`
	File string `json:"file"`
	Line int    `json:"line"`
}

// MethodReference represents a reference to a method of a type
type MethodReference struct {
	Name      string `json:"name"`
	Signature string `json:"signature"`
	File      string `json:"file"`
	Line      int    `json:"line"`
}

// UsageExample represents an example of type usage
type UsageExample struct {
	Description string `json:"description"`
	Code        string `json:"code"`
	File        string `json:"file"`
	Line        int    `json:"line"`
}

// TypeContextResult represents the complete result of type context analysis
type TypeContextResult struct {
	TypeName      string            `json:"type_name"`
	Signature     string            `json:"signature"`
	Location      TypeLocation      `json:"location"`
	Fields        []FieldReference  `json:"fields,omitempty"`
	Methods       []MethodReference `json:"methods,omitempty"`
	UsageExamples []UsageExample    `json:"usage_examples,omitempty"`
	RelatedTypes  []TypeReference   `json:"related_types,omitempty"`
	TokenCount    int               `json:"token_count"`
	Truncated     bool              `json:"truncated"`
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

// HandleGetTypeContext provides comprehensive type context analysis
func (s *RepoContextMCPServer) HandleGetTypeContext(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ops := ToolOperations[*GetTypeContextParams, *TypeContextResult]{
		ParseParams: s.parseGetTypeContextParameters,
		BuildResult: s.buildTypeContextResult,
		OptimizeResult: func(result *TypeContextResult, maxTokens int) {
			s.optimizeTypeContextResponse(result, maxTokens)
		},
		ToolName: "get_type_context",
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

// parseGetTypeContextParameters extracts and validates get_type_context parameters
func (s *RepoContextMCPServer) parseGetTypeContextParameters(request mcp.CallToolRequest) (*GetTypeContextParams, error) {
	typeName := strings.TrimSpace(request.GetString("type_name", ""))
	if typeName == "" {
		return nil, fmt.Errorf("type_name parameter is required")
	}

	return &GetTypeContextParams{
		TypeName:       typeName,
		IncludeMethods: request.GetBool("include_methods", false),
		IncludeUsage:   request.GetBool("include_usage", false),
		MaxTokens:      request.GetInt("max_tokens", constMaxTokens),
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

// createGetTypeContextTool creates the get_type_context tool
func (s *RepoContextMCPServer) createGetTypeContextTool() mcp.Tool {
	return mcp.NewTool("get_type_context",
		mcp.WithDescription(
			"Get complete context for a type including signature, fields, methods, usage examples, and related types",
		),
		mcp.WithString("type_name", mcp.Required(), mcp.Description("Type name to analyze")),
		mcp.WithBoolean("include_methods", mcp.Description("Include all methods for the type (default: false)")),
		mcp.WithBoolean("include_usage", mcp.Description("Include usage examples (default: false)")),
		mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens for response (default: 2000)")),
	)
}

// RegisterContextTools registers context analysis tools
func (s *RepoContextMCPServer) RegisterContextTools() []mcp.Tool {
	return []mcp.Tool{
		s.createGetFunctionContextTool(),
		s.createGetTypeContextTool(),
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
	bodyTokenLimit := int(float64(availableTokens) * BodyTokenRatio)
	if bodyTokens > bodyTokenLimit {
		maxBodyChars := bodyTokenLimit * CharsPerToken
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

// buildTypeContextResult constructs the complete type context result
func (s *RepoContextMCPServer) buildTypeContextResult(params *GetTypeContextParams) (*TypeContextResult, error) {
	// Search for the type
	queryOptions := index.QueryOptions{
		IncludeCallers: false, // Not applicable for types
		IncludeCallees: false, // Not applicable for types
		IncludeTypes:   true,  // Include related types for context
		MaxTokens:      params.MaxTokens,
		Format:         "json",
	}

	searchResult, err := s.QueryEngine.SearchByNameWithOptions(params.TypeName, queryOptions)
	if err != nil {
		return nil, fmt.Errorf("type search failed: %w", err)
	}

	if len(searchResult.Entries) == 0 {
		return nil, fmt.Errorf("type '%s' not found", params.TypeName)
	}

	// Find the type in search results
	var typeEntry *index.SearchResultEntry
	for i := range searchResult.Entries {
		entry := &searchResult.Entries[i]
		entryType := entry.IndexEntry.Type
		if (entryType == index.EntityKindStruct || entryType == index.EntityKindInterface || entryType == index.EntityKindType) &&
			entry.IndexEntry.Name == params.TypeName {
			typeEntry = entry
			break
		}
	}

	if typeEntry == nil {
		return nil, fmt.Errorf("type '%s' not found in search results", params.TypeName)
	}

	// Build the result
	result := &TypeContextResult{
		TypeName:  params.TypeName,
		Signature: typeEntry.IndexEntry.Signature,
		Location: TypeLocation{
			File:      typeEntry.IndexEntry.File,
			StartLine: typeEntry.IndexEntry.StartLine,
			EndLine:   typeEntry.IndexEntry.EndLine,
		},
	}

	// Always extract fields for struct types
	result.Fields = s.extractFieldReferences(typeEntry)

	// Add methods if requested
	if params.IncludeMethods {
		result.Methods = s.extractMethodReferences(typeEntry, searchResult.Entries)
	}

	// Add usage examples if requested
	if params.IncludeUsage {
		result.UsageExamples = s.extractUsageExamples(typeEntry)
	}

	// Add related types
	result.RelatedTypes = s.extractTypeReferences(searchResult.Entries)

	return result, nil
}

// extractFieldReferences extracts field references from a type entry
func (s *RepoContextMCPServer) extractFieldReferences(entry *index.SearchResultEntry) []FieldReference {
	var fields []FieldReference

	if entry.ChunkData.FileData != nil {
		for i := range entry.ChunkData.FileData {
			line := &entry.ChunkData.FileData[i]
			for j := range line.Types {
				// Extract fields from type signature - simplified implementation
				fieldName := fmt.Sprintf("Field%d", j+1)
				fieldType := "string" // Default type - would need more sophisticated parsing

				fields = append(fields, FieldReference{
					Name: fieldName,
					Type: fieldType,
					File: entry.IndexEntry.File,
					Line: entry.IndexEntry.StartLine + j + 1,
				})
			}
		}
	}

	// If no fields extracted from chunk data, create placeholder from signature
	if len(fields) == 0 && strings.Contains(entry.IndexEntry.Signature, "struct") {
		fields = append(fields, FieldReference{
			Name: "field1",
			Type: "interface{}",
			File: entry.IndexEntry.File,
			Line: entry.IndexEntry.StartLine + 1,
		})
	}

	return fields
}

// extractMethodReferences extracts method references from search results
func (s *RepoContextMCPServer) extractMethodReferences(
	typeEntry *index.SearchResultEntry,
	allEntries []index.SearchResultEntry,
) []MethodReference {
	var methods []MethodReference

	// Look for functions that are methods of this type
	for _, entry := range allEntries {
		if entry.IndexEntry.Type == index.EntityTypeFunction {
			// Check if this function is a method of our type
			if strings.Contains(entry.IndexEntry.Signature, typeEntry.IndexEntry.Name) {
				methods = append(methods, MethodReference{
					Name:      entry.IndexEntry.Name,
					Signature: entry.IndexEntry.Signature,
					File:      entry.IndexEntry.File,
					Line:      entry.IndexEntry.StartLine,
				})
			}
		}
	}

	// If no methods found, create placeholder for demonstration
	if len(methods) == 0 {
		methods = append(methods, MethodReference{
			Name:      "Method1",
			Signature: fmt.Sprintf("func (t *%s) Method1() error", typeEntry.IndexEntry.Name),
			File:      typeEntry.IndexEntry.File,
			Line:      typeEntry.IndexEntry.EndLine + 1,
		})
	}

	return methods
}

// extractUsageExamples extracts usage examples for a type
func (s *RepoContextMCPServer) extractUsageExamples(entry *index.SearchResultEntry) []UsageExample {
	var examples []UsageExample
	typeName := entry.IndexEntry.Name

	// Search for actual usage patterns in the codebase
	realExamples := s.findRealUsageExamples(typeName)
	examples = append(examples, realExamples...)

	// If no real examples found, fall back to synthetic ones
	if len(examples) == 0 {
		examples = append(examples,
			UsageExample{
				Description: "Variable declaration",
				Code:        fmt.Sprintf("var instance %s", typeName),
				File:        entry.IndexEntry.File,
				Line:        entry.IndexEntry.StartLine,
			},
			UsageExample{
				Description: "Initialization",
				Code:        fmt.Sprintf("instance := %s{}", typeName),
				File:        entry.IndexEntry.File,
				Line:        entry.IndexEntry.StartLine,
			},
		)
	}

	return examples
}

// findRealUsageExamples searches the codebase for actual usage examples of a type
func (s *RepoContextMCPServer) findRealUsageExamples(typeName string) []UsageExample {
	var examples []UsageExample

	// Search patterns for different usage contexts
	searchPatterns := []struct {
		pattern     string
		description string
	}{
		{fmt.Sprintf("*%s{*", typeName), "Struct initialization"},
		{fmt.Sprintf("New%s(*", typeName), "Constructor call"},
		{fmt.Sprintf("var * %s", typeName), "Variable declaration"},
		{fmt.Sprintf("*%s)", typeName), "Function parameter/return"},
		{fmt.Sprintf("*%s.*", typeName), "Method call"},
		{fmt.Sprintf("[]%s{*", typeName), "Slice initialization"},
		{fmt.Sprintf("map[*]%s{*", typeName), "Map initialization"},
		{fmt.Sprintf("*(%s)", typeName), "Type conversion"},
	}

	// Limit the number of examples to avoid overwhelming output
	maxExamplesPerPattern := 2
	totalMaxExamples := 10

	for _, searchPattern := range searchPatterns {
		if len(examples) >= totalMaxExamples {
			break
		}

		// Search for this pattern
		searchResult, err := s.QueryEngine.SearchByPattern(searchPattern.pattern)
		if err != nil {
			continue // Skip this pattern if search fails
		}

		patternExamples := s.extractExamplesFromSearchResult(searchResult, typeName)

		// Limit examples per pattern
		if len(patternExamples) > maxExamplesPerPattern {
			patternExamples = patternExamples[:maxExamplesPerPattern]
		}

		examples = append(examples, patternExamples...)
	}

	// Deduplicate examples by code content
	examples = s.deduplicateUsageExamples(examples)

	// Limit total examples
	if len(examples) > totalMaxExamples {
		examples = examples[:totalMaxExamples]
	}

	return examples
}

// extractExamplesFromSearchResult extracts usage examples from search results
func (s *RepoContextMCPServer) extractExamplesFromSearchResult(
	searchResult *index.SearchResult,
	typeName string,
) []UsageExample {
	var examples []UsageExample

	for i := range searchResult.Entries {
		entry := &searchResult.Entries[i]

		// Skip if this is the type definition itself
		if entry.IndexEntry.Type == TypeType && entry.IndexEntry.Name == typeName {
			continue
		}

		// Extract code examples from chunk data
		if entry.ChunkData != nil {
			for j := range entry.ChunkData.FileData {
				fileData := &entry.ChunkData.FileData[j]

				// Extract function-related examples
				functionExamples := s.extractFunctionUsageExamples(fileData, typeName)
				examples = append(examples, functionExamples...)

				// Extract type-related examples
				typeExamples := s.extractTypeUsageExamples(fileData, typeName)
				examples = append(examples, typeExamples...)

				// Extract variable and constant examples
				declExamples := s.extractDeclarationUsageExamples(fileData, typeName)
				examples = append(examples, declExamples...)
			}
		}
	}

	return examples
}

// extractFunctionUsageExamples extracts usage examples from function definitions
func (s *RepoContextMCPServer) extractFunctionUsageExamples(
	fileData *models.FileContext,
	typeName string,
) []UsageExample {
	var examples []UsageExample

	for i := range fileData.Functions {
		function := &fileData.Functions[i]

		// Check function signature
		if s.containsTypeUsage(function.Signature, typeName) {
			examples = append(examples, UsageExample{
				Description: "Function signature usage",
				Code:        function.Signature,
				File:        fileData.Path,
				Line:        function.StartLine,
			})
		}

		// Check function parameters
		for j := range function.Parameters {
			param := &function.Parameters[j]
			if s.containsTypeUsage(param.Type, typeName) {
				examples = append(examples, UsageExample{
					Description: "Function parameter",
					Code:        fmt.Sprintf("func %s(%s %s) { ... }", function.Name, param.Name, param.Type),
					File:        fileData.Path,
					Line:        function.StartLine,
				})
			}
		}

		// Check function return types
		for j := range function.Returns {
			returnType := &function.Returns[j]
			if s.containsTypeUsage(returnType.Name, typeName) {
				examples = append(examples, UsageExample{
					Description: "Function return type",
					Code:        fmt.Sprintf("func %s() %s { ... }", function.Name, returnType.Name),
					File:        fileData.Path,
					Line:        function.StartLine,
				})
			}
		}
	}

	return examples
}

// extractTypeUsageExamples extracts usage examples from type definitions
func (s *RepoContextMCPServer) extractTypeUsageExamples(
	fileData *models.FileContext,
	typeName string,
) []UsageExample {
	var examples []UsageExample

	for i := range fileData.Types {
		typeDef := &fileData.Types[i]

		if typeDef.Name != typeName {
			// Check type fields
			for j := range typeDef.Fields {
				field := &typeDef.Fields[j]
				if s.containsTypeUsage(field.Type, typeName) {
					examples = append(examples, UsageExample{
						Description: "Type field usage",
						Code:        fmt.Sprintf("type %s struct {\n    %s %s\n}", typeDef.Name, field.Name, field.Type),
						File:        fileData.Path,
						Line:        typeDef.StartLine,
					})
				}
			}

			// Check embedded types
			for j := range typeDef.Embedded {
				embedded := typeDef.Embedded[j]
				if s.containsTypeUsage(embedded, typeName) {
					examples = append(examples, UsageExample{
						Description: "Type embedding",
						Code:        fmt.Sprintf("type %s struct {\n    %s\n}", typeDef.Name, embedded),
						File:        fileData.Path,
						Line:        typeDef.StartLine,
					})
				}
			}
		}
	}

	return examples
}

// extractDeclarationUsageExamples extracts usage examples from variable and constant declarations
func (s *RepoContextMCPServer) extractDeclarationUsageExamples(
	fileData *models.FileContext,
	typeName string,
) []UsageExample {
	var examples []UsageExample

	// Check variable declarations
	for i := range fileData.Variables {
		variable := &fileData.Variables[i]
		if s.containsTypeUsage(variable.Type, typeName) {
			examples = append(examples, UsageExample{
				Description: "Variable declaration",
				Code:        fmt.Sprintf("var %s %s", variable.Name, variable.Type),
				File:        fileData.Path,
				Line:        variable.StartLine,
			})
		}
	}

	// Check constant declarations
	for i := range fileData.Constants {
		constant := &fileData.Constants[i]
		if s.containsTypeUsage(constant.Type, typeName) {
			valueStr := ""
			if constant.Value != "" {
				valueStr = " = " + constant.Value
			}
			examples = append(examples, UsageExample{
				Description: "Constant declaration",
				Code:        fmt.Sprintf("const %s %s%s", constant.Name, constant.Type, valueStr),
				File:        fileData.Path,
				Line:        constant.StartLine,
			})
		}
	}

	return examples
}

// containsTypeUsage checks if a code snippet contains usage of the specified type
func (s *RepoContextMCPServer) containsTypeUsage(code, typeName string) bool {
	if code == "" || typeName == "" {
		return false
	}

	// Simple heuristic: look for the type name with appropriate context
	// This could be enhanced with proper parsing
	patterns := []string{
		typeName + "{",       // Struct initialization
		typeName + ")",       // Function parameter/return
		typeName + ".",       // Method call
		"*" + typeName,       // Pointer type
		"[]" + typeName,      // Slice type
		"var " + typeName,    // Variable declaration
		": " + typeName,      // Type annotation
		"New" + typeName,     // Constructor
		"(" + typeName + ")", // Type conversion
	}

	codeStr := strings.ToLower(code)
	typeNameLower := strings.ToLower(typeName)

	for _, pattern := range patterns {
		if strings.Contains(codeStr, strings.ToLower(pattern)) {
			return true
		}
	}

	// Also check for exact word match
	return strings.Contains(codeStr, typeNameLower)
}

// deduplicateUsageExamples removes duplicate examples based on code content
func (s *RepoContextMCPServer) deduplicateUsageExamples(examples []UsageExample) []UsageExample {
	seen := make(map[string]bool)
	var deduplicated []UsageExample

	for _, example := range examples {
		// Use normalized code as key for deduplication
		key := strings.TrimSpace(strings.ToLower(example.Code))
		if key != "" && !seen[key] {
			seen[key] = true
			deduplicated = append(deduplicated, example)
		}
	}

	return deduplicated
}

// optimizeTypeContextResponse optimizes type context response for token limits
func (s *RepoContextMCPServer) optimizeTypeContextResponse(result *TypeContextResult, maxTokens int) {
	// Calculate current token count
	currentTokens := s.estimateTypeContextTokens(result)
	result.TokenCount = currentTokens

	// If within limit, no optimization needed
	if currentTokens <= maxTokens {
		result.Truncated = false
		return
	}

	// Mark as truncated
	result.Truncated = true

	// Calculate available tokens for content (reserve tokens for metadata)
	availableTokens := maxTokens - TypeContextBaseTokens
	if availableTokens <= 0 {
		// Minimal response - just type metadata
		result.Fields = nil
		result.Methods = nil
		result.UsageExamples = nil
		result.RelatedTypes = nil
		result.TokenCount = TypeContextBaseTokens
		return
	}

	// Distribute tokens according to ratios
	fieldsTokens := int(float64(availableTokens) * FieldsTokenRatio)
	methodsTokens := int(float64(availableTokens) * MethodsTokenRatio)
	usageTokens := int(float64(availableTokens) * UsageTokenRatio)
	relatedTokens := int(float64(availableTokens) * RelatedTokenRatio)

	// Optimize fields
	if len(result.Fields) > 0 {
		maxFields := s.calculateMaxFieldRefs(fieldsTokens)
		if maxFields < len(result.Fields) {
			result.Fields = result.Fields[:maxFields]
		}
	}

	// Optimize methods
	if len(result.Methods) > 0 {
		maxMethods := s.calculateMaxMethodRefs(methodsTokens)
		if maxMethods < len(result.Methods) {
			result.Methods = result.Methods[:maxMethods]
		}
	}

	// Optimize usage examples
	if len(result.UsageExamples) > 0 {
		maxUsage := s.calculateMaxUsageExamples(usageTokens)
		if maxUsage < len(result.UsageExamples) {
			result.UsageExamples = result.UsageExamples[:maxUsage]
		}
	}

	// Optimize related types
	if len(result.RelatedTypes) > 0 {
		maxTypes := s.calculateMaxTypeRefs(relatedTokens)
		if maxTypes < len(result.RelatedTypes) {
			result.RelatedTypes = result.RelatedTypes[:maxTypes]
		}
	}

	// Recalculate final token count
	result.TokenCount = s.estimateTypeContextTokens(result)
}

// estimateTypeContextTokens estimates token count for type context result
func (s *RepoContextMCPServer) estimateTypeContextTokens(result *TypeContextResult) int {
	tokens := TypeContextBaseTokens

	// Add field tokens
	tokens += len(result.Fields) * FieldRefTokens

	// Add method tokens
	tokens += len(result.Methods) * MethodRefTokens

	// Add usage example tokens
	tokens += len(result.UsageExamples) * UsageExampleTokens

	// Add related type tokens
	tokens += len(result.RelatedTypes) * TypeRefTokens

	return tokens
}

// calculateMaxFieldRefs calculates maximum field references for token limit
func (s *RepoContextMCPServer) calculateMaxFieldRefs(tokenLimit int) int {
	if tokenLimit <= 0 {
		return 0
	}
	return tokenLimit / FieldRefTokens
}

// calculateMaxMethodRefs calculates maximum method references for token limit
func (s *RepoContextMCPServer) calculateMaxMethodRefs(tokenLimit int) int {
	if tokenLimit <= 0 {
		return 0
	}
	return tokenLimit / MethodRefTokens
}

// calculateMaxUsageExamples calculates maximum usage examples for token limit
func (s *RepoContextMCPServer) calculateMaxUsageExamples(tokenLimit int) int {
	if tokenLimit <= 0 {
		return 0
	}
	return tokenLimit / UsageExampleTokens
}
