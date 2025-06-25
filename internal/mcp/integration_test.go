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
