package index

import (
	"repository-context-protocol/internal/models"
)

// GlobalCallGraph represents a repository-wide function call graph
type GlobalCallGraph struct {
	// callersMap maps function names to their callers
	callersMap map[string][]models.CallRelation
	// calleesMap maps function names to their callees
	calleesMap map[string][]models.CallRelation
	// allFunctions tracks all functions in the call graph
	allFunctions map[string]bool
}

// CallGraphStatistics provides metrics about the call graph
type CallGraphStatistics struct {
	TotalFunctions     int
	TotalCallRelations int
	MaxCallDepth       int
}

// NewGlobalCallGraph creates a new global call graph
func NewGlobalCallGraph() *GlobalCallGraph {
	return &GlobalCallGraph{
		callersMap:   make(map[string][]models.CallRelation),
		calleesMap:   make(map[string][]models.CallRelation),
		allFunctions: make(map[string]bool),
	}
}

// BuildFromFiles constructs the global call graph from multiple file contexts
func (gcg *GlobalCallGraph) BuildFromFiles(fileContexts []models.FileContext) error {
	// Clear existing data
	gcg.callersMap = make(map[string][]models.CallRelation)
	gcg.calleesMap = make(map[string][]models.CallRelation)
	gcg.allFunctions = make(map[string]bool)

	// First pass: collect all functions
	for i := range fileContexts {
		fileContext := &fileContexts[i]
		for j := range fileContext.Functions {
			function := &fileContext.Functions[j]
			gcg.allFunctions[function.Name] = true
		}
	}

	// Second pass: build call relationships using LocalCallsWithMetadata
	for i := range fileContexts {
		fileContext := &fileContexts[i]
		for j := range fileContext.Functions {
			function := &fileContext.Functions[j]

			// Process each call made by this function using metadata
			for _, callMeta := range function.LocalCallsWithMetadata {
				// Create call relation with actual call line number
				relation := models.CallRelation{
					Caller:     function.Name,
					Callee:     callMeta.FunctionName,
					File:       fileContext.Path,
					CallerFile: fileContext.Path,
					Line:       callMeta.Line, // Use actual call line number
				}

				// Add to callers map (callee -> list of callers)
				gcg.callersMap[callMeta.FunctionName] = append(gcg.callersMap[callMeta.FunctionName], relation)

				// Add to callees map (caller -> list of callees)
				gcg.calleesMap[function.Name] = append(gcg.calleesMap[function.Name], relation)

				// Track the callee as a function (even if it's external like fmt.Println)
				gcg.allFunctions[callMeta.FunctionName] = true
			}

			// Fallback to deprecated Calls field if LocalCallsWithMetadata is empty (backward compatibility)
			if len(function.LocalCallsWithMetadata) == 0 && len(function.Calls) > 0 {
				for _, callee := range function.Calls {
					// Create call relation with function start line (less accurate)
					relation := models.CallRelation{
						Caller:     function.Name,
						Callee:     callee,
						File:       fileContext.Path,
						CallerFile: fileContext.Path,
						Line:       function.StartLine,
					}

					// Add to callers map (callee -> list of callers)
					gcg.callersMap[callee] = append(gcg.callersMap[callee], relation)

					// Add to callees map (caller -> list of callees)
					gcg.calleesMap[function.Name] = append(gcg.calleesMap[function.Name], relation)

					// Track the callee as a function (even if it's external like fmt.Println)
					gcg.allFunctions[callee] = true
				}
			}
		}
	}

	return nil
}

// GetCallers returns all functions that call the specified function
func (gcg *GlobalCallGraph) GetCallers(functionName string) []models.CallRelation {
	if callers, exists := gcg.callersMap[functionName]; exists {
		// Return a copy to prevent external modification
		result := make([]models.CallRelation, len(callers))
		copy(result, callers)
		return result
	}
	return []models.CallRelation{}
}

// GetCallees returns all functions called by the specified function
func (gcg *GlobalCallGraph) GetCallees(functionName string) []models.CallRelation {
	if callees, exists := gcg.calleesMap[functionName]; exists {
		// Return a copy to prevent external modification
		result := make([]models.CallRelation, len(callees))
		copy(result, callees)
		return result
	}
	return []models.CallRelation{}
}

// GetAllFunctions returns a list of all functions in the call graph
func (gcg *GlobalCallGraph) GetAllFunctions() []string {
	functions := make([]string, 0, len(gcg.allFunctions))
	for functionName := range gcg.allFunctions {
		functions = append(functions, functionName)
	}
	return functions
}

// GetCallChainDepth returns the shortest call chain depth from caller to callee
// Returns -1 if no path exists
func (gcg *GlobalCallGraph) GetCallChainDepth(caller, callee string) int {
	if caller == callee {
		return 0
	}

	// Use BFS to find shortest path
	visited := make(map[string]bool)
	queue := []string{caller}
	depth := 0

	for len(queue) > 0 {
		levelSize := len(queue)
		depth++

		// Process all nodes at current depth level
		for i := 0; i < levelSize; i++ {
			current := queue[0]
			queue = queue[1:]

			if visited[current] {
				continue
			}
			visited[current] = true

			// Check all callees of current function
			callees := gcg.GetCallees(current)
			for _, relation := range callees {
				if relation.Callee == callee {
					return depth
				}

				if !visited[relation.Callee] {
					queue = append(queue, relation.Callee)
				}
			}
		}
	}

	return -1 // No path found
}

// GetCallPath returns the shortest call path from caller to callee
// Returns empty slice if no path exists
func (gcg *GlobalCallGraph) GetCallPath(caller, callee string) []string {
	if caller == callee {
		return []string{caller}
	}

	// Use BFS to find shortest path and track parents
	visited := make(map[string]bool)
	parent := make(map[string]string)
	queue := []string{caller}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		// Check all callees of current function
		callees := gcg.GetCallees(current)
		for _, relation := range callees {
			if relation.Callee == callee {
				// Found target, reconstruct path
				path := []string{callee}
				node := current
				for node != caller {
					path = append([]string{node}, path...)
					node = parent[node]
				}
				path = append([]string{caller}, path...)
				return path
			}

			if !visited[relation.Callee] {
				parent[relation.Callee] = current
				queue = append(queue, relation.Callee)
			}
		}
	}

	return []string{} // No path found
}

// GetStatistics returns statistics about the call graph
func (gcg *GlobalCallGraph) GetStatistics() CallGraphStatistics {
	totalCallRelations := 0
	for _, relations := range gcg.calleesMap {
		totalCallRelations += len(relations)
	}

	// Calculate max call depth by finding the longest path from any function
	maxDepth := 0
	for functionName := range gcg.allFunctions {
		depth := gcg.calculateMaxDepthFrom(functionName)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return CallGraphStatistics{
		TotalFunctions:     len(gcg.allFunctions),
		TotalCallRelations: totalCallRelations,
		MaxCallDepth:       maxDepth,
	}
}

// calculateMaxDepthFrom calculates the maximum call depth starting from a function
func (gcg *GlobalCallGraph) calculateMaxDepthFrom(functionName string) int {
	visited := make(map[string]bool)
	return gcg.dfsMaxDepth(functionName, visited)
}

// dfsMaxDepth performs DFS to find maximum depth, handling cycles
func (gcg *GlobalCallGraph) dfsMaxDepth(functionName string, visited map[string]bool) int {
	if visited[functionName] {
		return 0 // Cycle detected, stop recursion
	}

	visited[functionName] = true
	maxDepth := 0

	callees := gcg.GetCallees(functionName)
	if len(callees) == 0 {
		// Leaf node - no further calls
		visited[functionName] = false
		return 0
	}

	for _, relation := range callees {
		depth := gcg.dfsMaxDepth(relation.Callee, visited)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	visited[functionName] = false // Backtrack for other paths
	return maxDepth + 1
}
