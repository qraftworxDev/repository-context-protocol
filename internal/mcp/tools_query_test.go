package mcp

import (
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"repository-context-protocol/internal/index"
	"repository-context-protocol/internal/models"
)

// TestAdvancedQueryTools_Registration tests the enhanced tool registration functionality
func TestAdvancedQueryTools_Registration(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Test that RegisterAdvancedQueryTools returns the expected tools
	tools := server.RegisterAdvancedQueryTools()

	expectedToolNames := []string{
		"query_by_name",
		"query_by_pattern",
		"get_call_graph",
		"list_functions",
		"list_types",
	}

	if len(tools) != len(expectedToolNames) {
		t.Errorf("Expected %d tools, got %d", len(expectedToolNames), len(tools))
	}

	// Verify all expected tool names are present
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	for _, expectedName := range expectedToolNames {
		if !toolNames[expectedName] {
			t.Errorf("Expected tool '%s' not found in registered tools", expectedName)
		}
	}

	// Verify tool descriptions are present (advanced tools should have detailed descriptions)
	for _, tool := range tools {
		if tool.Description == "" {
			t.Errorf("Tool '%s' should have a description", tool.Name)
		}
	}
}

// TestParameterParsing tests the enhanced parameter parsing methods
func TestParameterParsing(t *testing.T) {
	server := NewRepoContextMCPServer()

	t.Run("parseQueryByNameParameters", func(t *testing.T) {
		// Test valid parameters
		// Note: In real usage, this would be called with actual MCP request
		// For now, we test the interface exists and can be called
		request := mcp.CallToolRequest{}

		// This should fail because the mock request doesn't have the required name parameter
		params, err := server.parseQueryByNameParameters(request)
		if err == nil {
			t.Error("Expected error for missing name parameter")
		}
		if params != nil {
			t.Error("Expected nil params when error occurs")
		}
	})

	t.Run("parseQueryByPatternParameters", func(t *testing.T) {
		request := mcp.CallToolRequest{}

		// This should fail because the mock request doesn't have the required pattern parameter
		params, err := server.parseQueryByPatternParameters(request)
		if err == nil {
			t.Error("Expected error for missing pattern parameter")
		}
		if params != nil {
			t.Error("Expected nil params when error occurs")
		}
	})

	t.Run("parseGetCallGraphParameters", func(t *testing.T) {
		request := mcp.CallToolRequest{}

		// This should fail because the mock request doesn't have the required function_name parameter
		params, err := server.parseGetCallGraphParameters(request)
		if err == nil {
			t.Error("Expected error for missing function_name parameter")
		}
		if params != nil {
			t.Error("Expected nil params when error occurs")
		}
	})

	t.Run("parseListEntitiesParameters", func(t *testing.T) {
		request := mcp.CallToolRequest{}

		// This should succeed since all parameters are optional
		params := server.parseListEntitiesParameters(request)
		if params == nil {
			t.Error("Expected valid params for list entities")
			return
		}

		// Verify default values
		if params.MaxTokens != constMaxTokens {
			t.Errorf("Expected default MaxTokens %d, got %d", constMaxTokens, params.MaxTokens)
		}
		if !params.IncludeSignatures {
			t.Error("Expected default IncludeSignatures to be true")
		}
	})
}

// TestEntityTypeValidation tests the entity type validation logic
func TestEntityTypeValidation(t *testing.T) {
	server := NewRepoContextMCPServer()

	testCases := []struct {
		entityType  string
		expectError bool
	}{
		{"", false},         // Empty is valid (no filter)
		{"function", false}, // Valid type
		{"type", false},     // Valid type
		{"variable", false}, // Valid type
		{"constant", false}, // Valid type
		{"invalid", true},   // Invalid type
		{"class", true},     // Invalid type (not supported)
		{"method", true},    // Invalid type (not supported)
	}

	for _, tc := range testCases {
		t.Run("entity_type_"+tc.entityType, func(t *testing.T) {
			err := server.validateEntityType(tc.entityType)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for entity type '%s'", tc.entityType)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for entity type '%s': %v", tc.entityType, err)
			}
		})
	}
}

// TestQueryOptionsBuilder tests the interface for building query options
func TestQueryOptionsBuilder(t *testing.T) {
	server := NewRepoContextMCPServer()

	t.Run("QueryByNameParams", func(t *testing.T) {
		params := &QueryByNameParams{
			Name:           "testFunction",
			IncludeCallers: true,
			IncludeCallees: false,
			IncludeTypes:   true,
			MaxTokens:      1000,
		}

		options := server.buildQueryOptionsFromParams(params)

		if !options.IncludeCallers {
			t.Error("Expected IncludeCallers to be true")
		}
		if options.IncludeCallees {
			t.Error("Expected IncludeCallees to be false")
		}
		if !options.IncludeTypes {
			t.Error("Expected IncludeTypes to be true")
		}
		if options.MaxTokens != 1000 {
			t.Errorf("Expected MaxTokens 1000, got %d", options.MaxTokens)
		}
		if options.Format != "json" {
			t.Errorf("Expected Format 'json', got '%s'", options.Format)
		}
	})

	t.Run("GetCallGraphParams", func(t *testing.T) {
		params := &GetCallGraphParams{
			FunctionName:   "testFunction",
			MaxDepth:       3,
			IncludeCallers: false,
			IncludeCallees: true,
			MaxTokens:      2000,
		}

		options := server.buildQueryOptionsFromParams(params)

		if options.IncludeCallers {
			t.Error("Expected IncludeCallers to be false")
		}
		if !options.IncludeCallees {
			t.Error("Expected IncludeCallees to be true")
		}
		if options.IncludeTypes {
			t.Error("Expected IncludeTypes to be false for call graph")
		}
		if options.MaxTokens != 2000 {
			t.Errorf("Expected MaxTokens 2000, got %d", options.MaxTokens)
		}
	})

	t.Run("ListEntitiesParams", func(t *testing.T) {
		params := &ListEntitiesParams{
			MaxTokens:         1500,
			IncludeSignatures: false,
			Limit:             10,
			Offset:            5,
		}

		options := server.buildQueryOptionsFromParams(params)

		// List operations don't use caller/callee/types options
		if options.IncludeCallers {
			t.Error("Expected IncludeCallers to be false for list operations")
		}
		if options.IncludeCallees {
			t.Error("Expected IncludeCallees to be false for list operations")
		}
		if options.IncludeTypes {
			t.Error("Expected IncludeTypes to be false for list operations")
		}
		if options.MaxTokens != 1500 {
			t.Errorf("Expected MaxTokens 1500, got %d", options.MaxTokens)
		}
	})
}

// TestPaginationHelpers tests pagination functionality
func TestPaginationHelpers(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Create sample search result
	createTestResult := func() *index.SearchResult {
		entries := make([]index.SearchResultEntry, 10)
		for i := 0; i < 10; i++ {
			entries[i] = index.SearchResultEntry{
				IndexEntry: models.IndexEntry{
					Name: fmt.Sprintf("item%d", i),
					Type: "function",
				},
			}
		}
		return &index.SearchResult{Entries: entries}
	}

	t.Run("applyPagination_limit_only", func(t *testing.T) {
		result := createTestResult()
		server.applyPagination(result, 5, 0)

		if len(result.Entries) != 5 {
			t.Errorf("Expected 5 entries after limit, got %d", len(result.Entries))
		}
		if !result.Truncated {
			t.Error("Expected result to be marked as truncated")
		}
	})

	t.Run("applyPagination_offset_only", func(t *testing.T) {
		result := createTestResult()
		server.applyPagination(result, 0, 3)

		if len(result.Entries) != 7 {
			t.Errorf("Expected 7 entries after offset, got %d", len(result.Entries))
		}
		if result.Entries[0].IndexEntry.Name != "item3" {
			t.Errorf("Expected first entry 'item3', got '%s'", result.Entries[0].IndexEntry.Name)
		}
	})

	t.Run("applyPagination_limit_and_offset", func(t *testing.T) {
		result := createTestResult()
		server.applyPagination(result, 3, 2)

		if len(result.Entries) != 3 {
			t.Errorf("Expected 3 entries after limit and offset, got %d", len(result.Entries))
		}
		if result.Entries[0].IndexEntry.Name != "item2" {
			t.Errorf("Expected first entry 'item2', got '%s'", result.Entries[0].IndexEntry.Name)
		}
		if !result.Truncated {
			t.Error("Expected result to be marked as truncated")
		}
	})

	t.Run("removeSignatures", func(t *testing.T) {
		result := createTestResult()
		// Add signatures
		for i := range result.Entries {
			result.Entries[i].IndexEntry.Signature = fmt.Sprintf("func item%d()", i)
		}

		server.removeSignatures(result)

		for i, entry := range result.Entries {
			if entry.IndexEntry.Signature != "" {
				t.Errorf("Expected empty signature for entry %d, got '%s'", i, entry.IndexEntry.Signature)
			}
		}
	})
}

// TestAdvancedResponseFormatting tests response formatting for advanced features
func TestAdvancedResponseFormatting(t *testing.T) {
	server := NewRepoContextMCPServer()

	t.Run("complex_search_result", func(t *testing.T) {
		searchResult := &index.SearchResult{
			Query:      "testFunction",
			SearchType: "name",
			Entries: []index.SearchResultEntry{
				{
					IndexEntry: models.IndexEntry{
						Name:      "testFunction",
						Type:      "function",
						File:      "test.go",
						StartLine: 10,
						Signature: "func testFunction() error",
					},
				},
			},
			TokenCount: 150,
			Truncated:  false,
		}

		result := server.FormatSuccessResponse(searchResult)
		if result == nil || result.IsError {
			t.Error("Expected successful response formatting")
		}

		if len(result.Content) == 0 {
			t.Error("Expected response content")
		}
	})

	t.Run("error_with_operation_context", func(t *testing.T) {
		testErr := fmt.Errorf("parameter validation failed: name is required")

		result := server.FormatErrorResponse("query_by_name", testErr)
		if result == nil || !result.IsError {
			t.Error("Expected error response")
		}

		if len(result.Content) == 0 {
			t.Error("Expected error content")
		}
	})
}
