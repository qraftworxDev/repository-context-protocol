package index

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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

	// Entity kind constants
	EntityKindStruct    = "struct"
	EntityKindInterface = "interface"
	EntityKindType      = "type"
	EntityKindAlias     = "alias"
	EntityKindEnum      = "enum"
	EntityKindFunction  = "function"

	// Token estimation constants
	TokenOverhead    = 10
	CallerTokens     = 20
	CalleeTokens     = 20
	MetadataTokens   = 50
	DefaultMaxDepth  = 1
	LookAheadMatches = 2
)

// Query engine for semantic searches

// QueryEngine provides semantic search capabilities over the indexed repository
type QueryEngine struct {
	storage    *HybridStorage
	regexCache map[string]*regexp.Regexp
	regexMutex sync.RWMutex
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
	Options    *QueryOptions       `json:"-"`                    // Original query options (not serialized)
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
		storage:    storage,
		regexCache: make(map[string]*regexp.Regexp),
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
		Options:    &options,
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
				callGraph, err := qe.GetCallGraphWithOptions(entry.IndexEntry.Name, options)
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
		Options:    &QueryOptions{}, // Default empty options
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

// SearchByTypeWithOptions searches for all entities of a specific type with query options
func (qe *QueryEngine) SearchByTypeWithOptions(entityType string, options QueryOptions) (*SearchResult, error) {
	result := &SearchResult{
		Query:      entityType,
		SearchType: "type",
		ExecutedAt: time.Now(),
		Options:    &options,
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

	// Add call graph information if requested and functions are found
	if (options.IncludeCallers || options.IncludeCallees) && entityType == EntityTypeFunction && len(result.Entries) > 0 {
		// Get call graph for the first function found
		for _, entry := range result.Entries {
			if entry.IndexEntry.Type == EntityTypeFunction {
				callGraph, err := qe.GetCallGraphWithOptions(entry.IndexEntry.Name, options)
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

// SearchByPattern searches for entities matching a pattern (supports wildcards)
func (qe *QueryEngine) SearchByPattern(pattern string) (*SearchResult, error) {
	// Use the full-featured version with default options
	return qe.SearchByPatternWithOptions(pattern, QueryOptions{})
}

// SearchByPatternWithOptions searches for entities matching a pattern with query options
func (qe *QueryEngine) SearchByPatternWithOptions(pattern string, options QueryOptions) (*SearchResult, error) {
	result := &SearchResult{
		Query:      pattern,
		SearchType: "pattern",
		ExecutedAt: time.Now(),
		Options:    &options,
	}

	// For now, implement simple prefix matching with *
	// In a full implementation, this could use more sophisticated pattern matching
	var allEntries []SearchResultEntry

	// Get all entity types and search each
	// Search functions, variables, constants
	basicEntityTypes := []string{EntityTypeFunction, EntityTypeVariable, EntityTypeConstant}
	for _, entityType := range basicEntityTypes {
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

	// Search all type kinds (struct, interface, etc.) since types are stored by their specific kind
	typeKinds := []string{EntityKindStruct, EntityKindInterface, EntityKindType, EntityKindAlias, EntityKindEnum}
	for _, typeKind := range typeKinds {
		queryResults, err := qe.storage.QueryByType(typeKind)
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

	// Add call graph information if requested and functions are found
	if (options.IncludeCallers || options.IncludeCallees) && len(result.Entries) > 0 {
		// Find the first function entry to get call graph for
		for _, entry := range result.Entries {
			if entry.IndexEntry.Type == EntityTypeFunction {
				callGraph, err := qe.GetCallGraphWithOptions(entry.IndexEntry.Name, options)
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

// SearchInFile searches for all entities within a specific file
func (qe *QueryEngine) SearchInFile(filePath string) (*SearchResult, error) {
	return qe.SearchInFileWithOptions(filePath, QueryOptions{})
}

// SearchInFileWithOptions searches for all entities within a specific file with query options
func (qe *QueryEngine) SearchInFileWithOptions(filePath string, options QueryOptions) (*SearchResult, error) {
	result := &SearchResult{
		Query:      filePath,
		SearchType: "file",
		ExecutedAt: time.Now(),
		Options:    &options,
	}

	// Get all entity types and filter by file
	var allEntries []SearchResultEntry
	entityTypes := []string{EntityTypeFunction, EntityTypeVariable, EntityTypeConstant}

	// For types, we need to search for all specific type kinds (struct, interface, etc.)
	// since they are stored by their specific kind, not the generic "type"
	typeKinds := []string{EntityKindStruct, EntityKindInterface, EntityKindType, EntityKindAlias, EntityKindEnum}

	// Search for functions, variables, constants
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

	// Search for all type kinds
	for _, typeKind := range typeKinds {
		queryResults, err := qe.storage.QueryByType(typeKind)
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

	// Add call graph information if requested
	if (options.IncludeCallers || options.IncludeCallees) && len(result.Entries) > 0 {
		// Find the first function entry to get call graph for
		for _, entry := range result.Entries {
			if entry.IndexEntry.Type == EntityTypeFunction {
				callGraph, err := qe.GetCallGraphWithOptions(entry.IndexEntry.Name, options)
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

// GetCallGraph retrieves the call graph for a function
func (qe *QueryEngine) GetCallGraph(functionName string, maxDepth int) (*CallGraphInfo, error) {
	// For backward compatibility, include both callers and callees
	options := QueryOptions{
		IncludeCallers: true,
		IncludeCallees: true,
		MaxDepth:       maxDepth,
	}
	return qe.GetCallGraphWithOptions(functionName, options)
}

// GetCallGraphWithOptions retrieves the call graph for a function with selective inclusion
func (qe *QueryEngine) GetCallGraphWithOptions(functionName string, options QueryOptions) (*CallGraphInfo, error) {
	callGraph := &CallGraphInfo{
		Function: functionName,
		Callers:  []CallGraphEntry{},
		Callees:  []CallGraphEntry{},
		Depth:    options.MaxDepth,
	}

	// Use default depth of 1 if not specified or invalid
	maxDepth := options.MaxDepth
	if maxDepth <= 0 {
		maxDepth = DefaultMaxDepth
	}

	// Only retrieve callers if requested
	if options.IncludeCallers {
		callers, err := qe.populateCallGraphEntriesWithDepth(functionName, true, maxDepth, 0, make(map[string]bool))
		if err != nil {
			return nil, fmt.Errorf("failed to query callers: %w", err)
		}
		callGraph.Callers = callers
	}

	// Only retrieve callees if requested
	if options.IncludeCallees {
		callees, err := qe.populateCallGraphEntriesWithDepth(functionName, false, maxDepth, 0, make(map[string]bool))
		if err != nil {
			return nil, fmt.Errorf("failed to query callees: %w", err)
		}
		callGraph.Callees = callees
	}

	return callGraph, nil
}

// populateCallGraphEntriesWithDepth recursively populates call graph entries up to maxDepth
func (qe *QueryEngine) populateCallGraphEntriesWithDepth(
	functionName string,
	isCallers bool,
	maxDepth, currentDepth int,
	visited map[string]bool,
) ([]CallGraphEntry, error) {
	var entries []CallGraphEntry

	// Stop if we've reached max depth
	if currentDepth >= maxDepth {
		return entries, nil
	}

	// Prevent infinite loops in circular call graphs
	if visited[functionName] {
		return entries, nil
	}
	visited[functionName] = true

	if isCallers {
		// Handle callers: functions that call this function
		callers, err := qe.storage.QueryCallsTo(functionName)
		if err != nil {
			return nil, err
		}

		for _, caller := range callers {
			entry := qe.createCallGraphEntry(caller.Caller, caller.CallerFile, caller.Line)
			entries = append(entries, entry)

			// Recursively get callers of this caller
			if currentDepth+1 < maxDepth {
				subEntries, err := qe.populateCallGraphEntriesWithDepth(caller.Caller, isCallers, maxDepth, currentDepth+1, visited)
				if err == nil {
					entries = append(entries, subEntries...)
				}
			}
		}
	} else {
		// Handle callees: functions called by this function
		callees, err := qe.storage.QueryCallsFrom(functionName)
		if err != nil {
			return nil, err
		}

		for _, callee := range callees {
			entry := qe.createCallGraphEntry(callee.Callee, callee.File, callee.Line)
			entries = append(entries, entry)

			// Recursively get callees of this callee
			if currentDepth+1 < maxDepth {
				subEntries, err := qe.populateCallGraphEntriesWithDepth(callee.Callee, isCallers, maxDepth, currentDepth+1, visited)
				if err == nil {
					entries = append(entries, subEntries...)
				}
			}
		}
	}

	// Remove this function from visited set for other branches
	delete(visited, functionName)

	return entries, nil
}

// createCallGraphEntry is a helper function to create a CallGraphEntry with chunk data
func (qe *QueryEngine) createCallGraphEntry(functionName, file string, line int) CallGraphEntry {
	// Load chunk data for the function
	functionEntries, err := qe.storage.QueryByName(functionName)
	var chunkData *models.SemanticChunk
	if err == nil && len(functionEntries) > 0 {
		chunkData = functionEntries[0].ChunkData
	}

	return CallGraphEntry{
		Function:  functionName,
		File:      file,
		Line:      line,
		ChunkData: chunkData,
	}
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

// matchesPattern supports both glob and regex patterns with automatic detection
func (qe *QueryEngine) matchesPattern(name, pattern string) bool {
	// Detect pattern type and route accordingly
	if qe.isRegexPattern(pattern) {
		return qe.matchesRegex(name, pattern)
	}
	return qe.matchesGlob(name, pattern)
}

// isRegexPattern detects if pattern uses regex syntax
func (qe *QueryEngine) isRegexPattern(pattern string) bool {
	// Check for explicit regex delimiters like /pattern/
	if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") && len(pattern) > 2 {
		return true
	}

	// Simple heuristic: contains regex metacharacters not commonly used in globs
	// Exclude braces {} as they're used in glob brace expansion
	regexChars := []string{"(", ")", "^", "$", "+", "|", "\\"}
	for _, char := range regexChars {
		if strings.Contains(pattern, char) {
			return true
		}
	}

	// Check for regex-specific patterns
	regexPatterns := []string{"(?", ".+", ".*", ".?", "\\d", "\\w", "\\s", "\\p{", "\\b"}
	for _, regexPattern := range regexPatterns {
		if strings.Contains(pattern, regexPattern) {
			return true
		}
	}

	// Check for regex quantifiers like {1,3} - these are regex if they follow pattern like word{min,max}
	if strings.Contains(pattern, "{") && strings.Contains(pattern, "}") {
		// Simple heuristic: if contains digits with comma inside braces, it's likely regex quantifier
		braceStart := strings.Index(pattern, "{")
		braceEnd := strings.Index(pattern[braceStart:], "}")
		if braceEnd != -1 {
			braceContent := pattern[braceStart+1 : braceStart+braceEnd]
			if strings.Contains(braceContent, ",") || (braceContent != "" && isAllDigits(braceContent)) {
				return true
			}
		}
	}

	return false
}

// isAllDigits checks if a string contains only digits
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// matchesGlob handles shell-style glob patterns with enhanced support
func (qe *QueryEngine) matchesGlob(name, pattern string) bool {
	// Handle brace expansion first
	if strings.Contains(pattern, "{") && strings.Contains(pattern, "}") {
		return qe.matchesBraceExpansion(name, pattern)
	}

	// Handle character class negation [!...] - Go's filepath.Match has issues with this
	if strings.Contains(pattern, "[!") {
		return qe.matchesGlobWithNegation(name, pattern)
	}

	// Use filepath.Match for standard glob patterns
	matched, err := filepath.Match(pattern, name)
	if err != nil {
		// Invalid glob pattern, fall back to exact match
		return name == pattern
	}
	return matched
}

// matchesRegex handles full regular expressions with caching
func (qe *QueryEngine) matchesRegex(name, pattern string) bool {
	regex, err := qe.getCompiledRegex(pattern)
	if err != nil {
		// Invalid regex, fall back to exact match
		return name == pattern
	}
	return regex.MatchString(name)
}

// matchesGlobWithNegation handles glob patterns with character class negation [!...]
func (qe *QueryEngine) matchesGlobWithNegation(name, pattern string) bool {
	// Convert negation pattern to equivalent regex
	// [!ABC] becomes [^ABC] in regex
	regexPattern := strings.ReplaceAll(pattern, "[!", "[^")

	// Convert glob wildcards to regex equivalents
	regexPattern = strings.ReplaceAll(regexPattern, "*", ".*")
	regexPattern = strings.ReplaceAll(regexPattern, "?", ".")

	// Anchor the pattern to match the entire string
	regexPattern = "^" + regexPattern + "$"

	// Use regex matching since it handles negation properly
	regex, err := qe.getCompiledRegex(regexPattern)
	if err != nil {
		// Fall back to exact match if regex compilation fails
		return name == pattern
	}

	return regex.MatchString(name)
}

// matchesBraceExpansion handles brace expansion patterns like {option1,option2}
func (qe *QueryEngine) matchesBraceExpansion(name, pattern string) bool {
	// Find the first brace group
	openBrace := strings.Index(pattern, "{")
	closeBrace := strings.Index(pattern[openBrace:], "}")
	if openBrace == -1 || closeBrace == -1 {
		// No valid brace group, fall back to standard glob
		matched, err := filepath.Match(pattern, name)
		return err == nil && matched
	}

	closeBrace += openBrace // Adjust for substring offset

	// Extract parts
	prefix := pattern[:openBrace]
	suffix := pattern[closeBrace+1:]
	options := pattern[openBrace+1 : closeBrace]

	// Handle empty braces
	if options == "" {
		return false
	}

	// Split options by comma
	optionList := strings.Split(options, ",")

	// Test each option
	for _, option := range optionList {
		expandedPattern := prefix + option + suffix
		// Use filepath.Match directly to avoid infinite recursion
		matched, err := filepath.Match(expandedPattern, name)
		if err == nil && matched {
			return true
		}
	}

	return false
}

// getCompiledRegex returns cached regex or compiles new one with thread safety
func (qe *QueryEngine) getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	// Strip regex delimiters if present
	cleanPattern := pattern
	if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") && len(pattern) > 2 {
		cleanPattern = pattern[1 : len(pattern)-1]
	}

	// Handle Go regex limitations - convert unsupported patterns to supported ones
	convertedPattern, err := qe.convertUnsupportedRegexFeaturesWithError(cleanPattern, false)
	if err != nil {
		return nil, fmt.Errorf("regex pattern contains unsupported features: %w", err)
	}
	cleanPattern = convertedPattern

	// Try to get from cache with read lock
	qe.regexMutex.RLock()
	if regex, exists := qe.regexCache[cleanPattern]; exists {
		qe.regexMutex.RUnlock()
		return regex, nil
	}
	qe.regexMutex.RUnlock()

	// Compile and cache with write lock
	qe.regexMutex.Lock()
	defer qe.regexMutex.Unlock()

	// Double-check after acquiring write lock (prevent race condition)
	if regex, exists := qe.regexCache[cleanPattern]; exists {
		return regex, nil
	}

	regex, err := regexp.Compile(cleanPattern)
	if err != nil {
		return nil, err
	}

	qe.regexCache[cleanPattern] = regex
	return regex, nil
}

// convertUnsupportedRegexFeatures converts unsupported regex features to supported alternatives.
// This function handles Go regexp package limitations by converting unsupported patterns
// to approximate alternatives. Warnings are logged when conversions occur.
//
// Unsupported features that are converted:
// - Negative lookbehind (?<!pattern): Removed entirely
// - Positive lookahead (?=pattern): Converted to .*pattern
// - Negative lookahead (?!pattern): Removed entirely
//
// See docs/regex_limitations.md for detailed documentation on limitations and alternatives.
func (qe *QueryEngine) convertUnsupportedRegexFeatures(pattern string) string {
	originalPattern := pattern
	conversionWarnings := []string{}

	// Convert negative lookbehind (?<!pattern) - simplified approach
	if strings.Contains(pattern, "(?<!") {
		// For patterns like "(?<!Process).*Data", we just match ".*Data"
		// This is a simplified fallback that may match more results than intended
		lookbehindPattern := regexp.MustCompile(`\(\?<![^)]+\)`)
		matches := lookbehindPattern.FindAllString(pattern, -1)
		pattern = lookbehindPattern.ReplaceAllString(pattern, "")

		for _, match := range matches {
			conversionWarnings = append(conversionWarnings,
				fmt.Sprintf("negative lookbehind '%s' removed", match))
		}
	}

	// Convert positive lookahead (?=pattern) - simplified approach
	if strings.Contains(pattern, "(?=") {
		// For patterns like "Handle(?=User)", we convert to "Handle.*User"
		// This is approximate and may have different matching behavior due to greedy matching
		lookaheadPattern := regexp.MustCompile(`\(\?=([^)]+)\)`)
		matches := lookaheadPattern.FindAllStringSubmatch(pattern, -1)
		pattern = lookaheadPattern.ReplaceAllString(pattern, ".*$1")

		for _, match := range matches {
			if len(match) >= LookAheadMatches {
				conversionWarnings = append(conversionWarnings,
					fmt.Sprintf("positive lookahead '(?=%s)' converted to '.*%s'", match[1], match[1]))
			}
		}
	}

	// Convert negative lookahead (?!pattern)
	if strings.Contains(pattern, "(?!") {
		// Remove the negative lookahead entirely (fallback behavior)
		// This may match more results than intended
		negLookaheadPattern := regexp.MustCompile(`\(\?![^)]+\)`)
		matches := negLookaheadPattern.FindAllString(pattern, -1)
		pattern = negLookaheadPattern.ReplaceAllString(pattern, "")

		for _, match := range matches {
			conversionWarnings = append(conversionWarnings,
				fmt.Sprintf("negative lookahead '%s' removed", match))
		}
	}

	// Log warnings if any conversions were made
	if len(conversionWarnings) > 0 {
		log.Printf("Regex pattern conversion warning: Original pattern '%s' contained unsupported features. Conversions: %s. "+
			"This may affect matching behavior. See docs/regex_limitations.md for details.",
			originalPattern, strings.Join(conversionWarnings, "; "))
	}

	return pattern
}

// convertUnsupportedRegexFeaturesWithError is like convertUnsupportedRegexFeatures but can return
// an error instead of silently converting patterns when strictMode is true.
//
// Parameters:
//   - pattern: The regex pattern to check and potentially convert
//   - strictMode: If true, returns an error when unsupported features are detected
//     If false, behaves like convertUnsupportedRegexFeatures with logging
//
// Returns the converted pattern and an error if strictMode is true and unsupported features were found.
func (qe *QueryEngine) convertUnsupportedRegexFeaturesWithError(pattern string, strictMode bool) (string, error) {
	originalPattern := pattern
	conversionWarnings := []string{}
	var unsupportedFeatures []string

	// Process negative lookbehind
	pattern = qe.processNegativeLookbehind(pattern, strictMode, &conversionWarnings, &unsupportedFeatures)

	// Process positive lookahead
	pattern = qe.processPositiveLookahead(pattern, strictMode, &conversionWarnings, &unsupportedFeatures)

	// Process negative lookahead
	pattern = qe.processNegativeLookahead(pattern, strictMode, &conversionWarnings, &unsupportedFeatures)

	// Handle results
	return qe.handleRegexConversionResults(originalPattern, pattern, strictMode, conversionWarnings, unsupportedFeatures)
}

// processLookaroundPattern is a generic helper for processing lookaround patterns
func (qe *QueryEngine) processLookaroundPattern(
	pattern,
	checkString,
	regexPattern,
	patternName,
	replacement string,
	strictMode bool,
	conversionWarnings, unsupportedFeatures *[]string,
) string {
	if !strings.Contains(pattern, checkString) {
		return pattern
	}

	compiledPattern := regexp.MustCompile(regexPattern)
	matches := compiledPattern.FindAllString(pattern, -1)

	if len(matches) == 0 {
		return pattern
	}

	if strictMode {
		for _, match := range matches {
			*unsupportedFeatures = append(*unsupportedFeatures,
				fmt.Sprintf("%s '%s'", patternName, match))
		}
		return pattern
	}

	// Convert pattern
	pattern = compiledPattern.ReplaceAllString(pattern, replacement)
	for _, match := range matches {
		if replacement == "" {
			*conversionWarnings = append(*conversionWarnings,
				fmt.Sprintf("%s '%s' removed", patternName, match))
		} else {
			*conversionWarnings = append(
				*conversionWarnings,
				fmt.Sprintf(
					"%s '%s' converted to '%s'",
					patternName,
					match,
					replacement,
				),
			)
		}
	}
	return pattern
}

// processNegativeLookbehind handles negative lookbehind patterns
func (qe *QueryEngine) processNegativeLookbehind(
	pattern string,
	strictMode bool,
	conversionWarnings,
	unsupportedFeatures *[]string,
) string {
	return qe.processLookaroundPattern(
		pattern,
		"(?<!",
		`\(\?<![^)]+\)`,
		"negative lookbehind",
		"",
		strictMode,
		conversionWarnings,
		unsupportedFeatures,
	)
}

// processPositiveLookahead handles positive lookahead patterns
func (qe *QueryEngine) processPositiveLookahead(pattern string, strictMode bool, conversionWarnings, unsupportedFeatures *[]string) string {
	if !strings.Contains(pattern, "(?=") {
		return pattern
	}

	lookaheadPattern := regexp.MustCompile(`\(\?=([^)]+)\)`)
	matches := lookaheadPattern.FindAllStringSubmatch(pattern, -1)

	if len(matches) == 0 {
		return pattern
	}

	if strictMode {
		for _, match := range matches {
			if len(match) >= LookAheadMatches {
				*unsupportedFeatures = append(*unsupportedFeatures,
					fmt.Sprintf("positive lookahead '(?=%s)'", match[1]))
			}
		}
		return pattern
	}

	// Convert pattern - this is more complex than the generic helper can handle
	// because we need to preserve the captured group content
	pattern = lookaheadPattern.ReplaceAllString(pattern, ".*$1")
	for _, match := range matches {
		if len(match) >= LookAheadMatches {
			*conversionWarnings = append(
				*conversionWarnings,
				fmt.Sprintf(
					"positive lookahead '(?=%s)' converted to '.*%s'",
					match[1],
					match[1],
				),
			)
		}
	}
	return pattern
}

// processNegativeLookahead handles negative lookahead patterns
func (qe *QueryEngine) processNegativeLookahead(
	pattern string,
	strictMode bool,
	conversionWarnings,
	unsupportedFeatures *[]string,
) string {
	return qe.processLookaroundPattern(
		pattern,
		"(?!",
		`\(\?![^)]+\)`,
		"negative lookahead",
		"",
		strictMode,
		conversionWarnings,
		unsupportedFeatures,
	)
}

// handleRegexConversionResults handles the final result processing
func (qe *QueryEngine) handleRegexConversionResults(
	originalPattern, pattern string, strictMode bool, conversionWarnings, unsupportedFeatures []string,
) (string, error) {
	// Handle strict mode - return error if unsupported features found
	if strictMode && len(unsupportedFeatures) > 0 {
		return "", fmt.Errorf("unsupported regex features detected: %s. "+
			"Use simpler patterns or see docs/regex_limitations.md for alternatives",
			strings.Join(unsupportedFeatures, ", "))
	}

	// Log warnings if any conversions were made (non-strict mode)
	if !strictMode && len(conversionWarnings) > 0 {
		log.Printf("Regex pattern conversion warning: Original pattern '%s' contained unsupported features. Conversions: %s. "+
			"This may affect matching behavior. See docs/regex_limitations.md for details.",
			originalPattern, strings.Join(conversionWarnings, "; "))
	}

	return pattern, nil
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

		// Show callers section if requested, even if empty
		if result.Options != nil && result.Options.IncludeCallers {
			output.WriteString("  Callers:\n")
			if len(result.CallGraph.Callers) > 0 {
				for _, caller := range result.CallGraph.Callers {
					output.WriteString(fmt.Sprintf("    - %s (%s:%d)\n", caller.Function, caller.File, caller.Line))
				}
			} else {
				output.WriteString("    (none)\n")
			}
		}

		// Show callees section if requested, even if empty
		if result.Options != nil && result.Options.IncludeCallees {
			output.WriteString("  Callees:\n")
			if len(result.CallGraph.Callees) > 0 {
				for _, callee := range result.CallGraph.Callees {
					output.WriteString(fmt.Sprintf("    - %s (%s:%d)\n", callee.Function, callee.File, callee.Line))
				}
			} else {
				output.WriteString("    (none)\n")
			}
		}
	}

	return []byte(output.String())
}
