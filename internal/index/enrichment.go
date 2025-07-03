package index

import (
	"repository-context-protocol/internal/models"
)

// GlobalEnrichment enhances FileContext objects with global cross-file analysis
type GlobalEnrichment struct {
	globalCallGraph *GlobalCallGraph
	// Keep a mapping of function name to file path for efficient lookups
	functionToFile map[string]string
}

// NewGlobalEnrichment creates a new global enrichment processor
func NewGlobalEnrichment() *GlobalEnrichment {
	return &GlobalEnrichment{
		globalCallGraph: NewGlobalCallGraph(),
		functionToFile:  make(map[string]string),
	}
}

// EnrichFileContexts enhances FileContext objects with global call graph information
func (ge *GlobalEnrichment) EnrichFileContexts(fileContexts []models.FileContext) ([]models.FileContext, error) {
	// Build mapping of function names to their defining files
	ge.buildFunctionToFileMapping(fileContexts)

	// Build global call graph from all file contexts
	if err := ge.globalCallGraph.BuildFromFiles(fileContexts); err != nil {
		return nil, err
	}

	// Create enriched copies of all file contexts
	enrichedContexts := make([]models.FileContext, len(fileContexts))
	for i := range fileContexts {
		enrichedContexts[i] = ge.enrichSingleFileContext(&fileContexts[i])
	}

	return enrichedContexts, nil
}

// buildFunctionToFileMapping creates a mapping of function names to their defining files
func (ge *GlobalEnrichment) buildFunctionToFileMapping(fileContexts []models.FileContext) {
	ge.functionToFile = make(map[string]string)

	for i := range fileContexts {
		fileContext := &fileContexts[i]
		for j := range fileContext.Functions {
			function := &fileContext.Functions[j]
			ge.functionToFile[function.Name] = fileContext.Path
		}
	}
}

// enrichSingleFileContext enriches a single FileContext with global information
func (ge *GlobalEnrichment) enrichSingleFileContext(fileContext *models.FileContext) models.FileContext {
	// Create a copy to avoid modifying the original
	enriched := *fileContext
	enriched.Functions = make([]models.Function, len(fileContext.Functions))

	// Enrich each function with global call information
	for i := range fileContext.Functions {
		enriched.Functions[i] = ge.enrichFunction(&fileContext.Functions[i], fileContext.Path)
	}

	return enriched
}

// enrichFunction enriches a single function with global call graph information
func (ge *GlobalEnrichment) enrichFunction(function *models.Function, currentFile string) models.Function {
	enriched := *function

	// Initialize all the enhanced fields
	enriched.LocalCalls = []string{}
	enriched.CrossFileCalls = []models.CallReference{}
	enriched.LocalCallers = []string{}
	enriched.CrossFileCallers = []models.CallReference{}

	// Process function calls - separate local from cross-file
	ge.categorizeCallees(&enriched, currentFile)

	// Process callers - separate local from cross-file
	ge.categorizeCallers(&enriched, currentFile)

	return enriched
}

// categorizeCallees separates function calls into local and cross-file categories
func (ge *GlobalEnrichment) categorizeCallees(function *models.Function, currentFile string) {
	// Clear existing LocalCalls and CrossFileCalls to rebuild from metadata
	function.LocalCalls = []string{}
	function.CrossFileCalls = []models.CallReference{}

	// Process each call with metadata to categorize correctly (preferred method)
	if len(function.LocalCallsWithMetadata) > 0 {
		for _, callMeta := range function.LocalCallsWithMetadata {
			calleeName := callMeta.FunctionName

			// Find the file where this function is defined
			calleeFile := ge.findFunctionFile(calleeName)

			// Check if this is a local call (within same file) or cross-file
			if calleeFile == currentFile {
				function.LocalCalls = append(function.LocalCalls, calleeName)
			} else {
				crossFileCall := models.CallReference{
					FunctionName: calleeName,
					File:         calleeFile,
					Line:         callMeta.Line, // Use actual call line number
					CallType:     callMeta.CallType,
				}
				function.CrossFileCalls = append(function.CrossFileCalls, crossFileCall)
			}
		}
	} else if len(function.Calls) > 0 {
		// Fallback to deprecated Calls field for backward compatibility
		for _, calleeName := range function.Calls {
			// Find the file where this function is defined
			calleeFile := ge.findFunctionFile(calleeName)

			// Check if this is a local call (within same file) or cross-file
			if calleeFile == currentFile {
				function.LocalCalls = append(function.LocalCalls, calleeName)
			} else {
				crossFileCall := models.CallReference{
					FunctionName: calleeName,
					File:         calleeFile,
					Line:         function.StartLine, // Use function start line (less accurate)
				}
				function.CrossFileCalls = append(function.CrossFileCalls, crossFileCall)
			}
		}
	}
}

// categorizeCallers separates function callers into local and cross-file categories
func (ge *GlobalEnrichment) categorizeCallers(function *models.Function, currentFile string) {
	// Get all callers from global call graph
	callerRelations := ge.globalCallGraph.GetCallers(function.Name)

	for _, relation := range callerRelations {
		callerName := relation.Caller

		// Check if this is a local caller (within same file) or cross-file
		if relation.CallerFile == currentFile {
			function.LocalCallers = append(function.LocalCallers, callerName)
		} else {
			crossFileCaller := models.CallReference{
				FunctionName: callerName,
				File:         relation.CallerFile,
				Line:         relation.Line, // This now has the actual call line number
			}
			function.CrossFileCallers = append(function.CrossFileCallers, crossFileCaller)
		}
	}
}

// findFunctionFile finds the file where a function is defined using the prebuilt mapping
func (ge *GlobalEnrichment) findFunctionFile(functionName string) string {
	if file, exists := ge.functionToFile[functionName]; exists {
		return file
	}

	// If not found in our mapping, it might be an external function (like fmt.Println)
	return "external"
}

// GetGlobalCallGraph returns the underlying global call graph for advanced queries
func (ge *GlobalEnrichment) GetGlobalCallGraph() *GlobalCallGraph {
	return ge.globalCallGraph
}
