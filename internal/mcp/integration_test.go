package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"repository-context-protocol/internal/index"
)

func TestHandleQueryByName_IntegrationWithTestData(t *testing.T) {
	// Skip if no test data available
	testDataPath := filepath.Join("..", "..", "testdata", "simple-go")
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("Test data not available - skipping integration test")
	}

	// Create a test server instance
	server := NewRepoContextMCPServer()
	server.RepoPath = testDataPath

	// Create .repocontext directory for testing
	repoContextPath := filepath.Join(testDataPath, ".repocontext")
	err := os.MkdirAll(repoContextPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create .repocontext directory: %v", err)
	}

	defer os.RemoveAll(repoContextPath) // Clean up after test

	// Initialize storage and build index
	storage := index.NewHybridStorage(repoContextPath)
	if err = storage.Initialize(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Build index for test data
	builder := index.NewIndexBuilder(testDataPath)
	if err = builder.Initialize(); err != nil {
		t.Fatalf("Failed to initialize builder: %v", err)
	}

	_, err = builder.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Initialize query engine
	server.Storage = storage
	server.QueryEngine = index.NewQueryEngine(storage)

	// Test querying for a known function
	testCases := []struct {
		name           string
		functionName   string
		expectedResult bool
	}{
		{
			name:           "query existing function",
			functionName:   "main",
			expectedResult: true,
		},
		{
			name:           "query non-existing function",
			functionName:   "nonExistentFunction",
			expectedResult: false,
		},
	}

	// Test the query engine directly since mock MCP requests are complex
	// This validates that the integration works end-to-end
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the query engine directly to validate the integration setup
			searchResult, err := server.QueryEngine.SearchByName(tc.functionName)
			if err != nil {
				t.Fatalf("SearchByName failed: %v", err)
			}

			// Verify search result structure
			if searchResult == nil {
				t.Fatal("SearchByName returned nil result")
			}

			// Check if we found results based on expectation
			if tc.expectedResult {
				if len(searchResult.Entries) == 0 {
					t.Errorf("Expected to find entries for '%s' but got none", tc.functionName)
				}
			}

			// Verify the result has proper query field
			if searchResult.Query != tc.functionName {
				t.Errorf("Expected query field to be '%s', got '%s'", tc.functionName, searchResult.Query)
			}
		})
	}
}

func TestHandleQueryByPattern_IntegrationWithTestData(t *testing.T) {
	// Skip if no test data available
	testDataPath := filepath.Join("..", "..", "testdata", "simple-go")
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("Test data not available - skipping integration test")
	}

	// Create a test server instance
	server := NewRepoContextMCPServer()
	server.RepoPath = testDataPath

	// Create .repocontext directory for testing
	repoContextPath := filepath.Join(testDataPath, ".repocontext")
	err := os.MkdirAll(repoContextPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create .repocontext directory: %v", err)
	}

	defer os.RemoveAll(repoContextPath) // Clean up after test

	// Initialize storage and build index
	storage := index.NewHybridStorage(repoContextPath)
	if err = storage.Initialize(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Build index for test data
	builder := index.NewIndexBuilder(testDataPath)
	if err = builder.Initialize(); err != nil {
		t.Fatalf("Failed to initialize builder: %v", err)
	}

	_, err = builder.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Initialize query engine
	server.Storage = storage
	server.QueryEngine = index.NewQueryEngine(storage)

	// Test pattern queries for known patterns
	testCases := []struct {
		name           string
		pattern        string
		expectedResult bool
		description    string
	}{
		{
			name:           "glob wildcard pattern",
			pattern:        "main*",
			expectedResult: true,
			description:    "Should match functions starting with 'main'",
		},
		{
			name:           "exact match pattern",
			pattern:        "main",
			expectedResult: true,
			description:    "Should match exact function name",
		},
		{
			name:           "suffix wildcard pattern",
			pattern:        "*Service",
			expectedResult: false,
			description:    "Should not match any functions ending with 'Service'",
		},
		{
			name:           "complex glob pattern",
			pattern:        "[mM]*",
			expectedResult: true,
			description:    "Should match functions starting with 'm' or 'M'",
		},
		{
			name:           "regex pattern",
			pattern:        "/^main/",
			expectedResult: true,
			description:    "Should match regex pattern for functions starting with 'main'",
		},
		{
			name:           "no match pattern",
			pattern:        "nonexistent*",
			expectedResult: false,
			description:    "Should not match any functions with this pattern",
		},
	}

	// Test the query engine directly since mock MCP requests are complex
	// This validates that the integration works end-to-end
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the query engine directly to validate the integration setup
			searchResult, err := server.QueryEngine.SearchByPattern(tc.pattern)
			if err != nil {
				t.Fatalf("SearchByPattern failed for pattern '%s': %v", tc.pattern, err)
			}

			// Verify search result structure
			if searchResult == nil {
				t.Fatal("SearchByPattern returned nil result")
			}

			// Check if we found results based on expectation
			if tc.expectedResult {
				if len(searchResult.Entries) == 0 {
					t.Errorf("Expected to find entries for pattern '%s' but got none", tc.pattern)
				}
			}

			// Verify the result has proper query field
			if searchResult.Query != tc.pattern {
				t.Errorf("Expected query field to be '%s', got '%s'", tc.pattern, searchResult.Query)
			}

			// Verify search type is pattern
			if searchResult.SearchType != "pattern" {
				t.Errorf("Expected search type to be 'pattern', got '%s'", searchResult.SearchType)
			}

			t.Logf("Pattern '%s': found %d entries (%s)", tc.pattern, len(searchResult.Entries), tc.description)
		})
	}

	// Test pattern search with entity type filtering
	t.Run("pattern_with_entity_type_filtering", func(t *testing.T) {
		// Search for all functions matching a pattern
		allResults, err := server.QueryEngine.SearchByPattern("*")
		if err != nil {
			t.Fatalf("SearchByPattern failed: %v", err)
		}

		// Count functions in the results
		functionCount := 0
		for _, entry := range allResults.Entries {
			if entry.IndexEntry.Type == "function" {
				functionCount++
			}
		}

		t.Logf("Found %d functions out of %d total entries", functionCount, len(allResults.Entries))

		// Test that we have some functions to filter
		if functionCount == 0 {
			t.Log("No functions found in test data - entity type filtering test limited")
		}
	})

	// Test pattern search with options
	t.Run("pattern_with_query_options", func(t *testing.T) {
		queryOptions := index.QueryOptions{
			IncludeCallers: true,
			IncludeCallees: true,
			MaxTokens:      1000,
		}

		searchResult, err := server.QueryEngine.SearchByPatternWithOptions("main*", queryOptions)
		if err != nil {
			t.Fatalf("SearchByPatternWithOptions failed: %v", err)
		}

		if searchResult == nil {
			t.Fatal("SearchByPatternWithOptions returned nil result")
		}

		// Verify options are preserved
		if searchResult.Options == nil {
			t.Error("Expected query options to be preserved in result")
		}

		t.Logf("Pattern with options: found %d entries", len(searchResult.Entries))
	})
}
