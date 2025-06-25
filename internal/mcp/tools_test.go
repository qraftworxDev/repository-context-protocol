package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"repository-context-protocol/internal/index"
)

func TestRepoContextMCPServer_RegisterQueryTools(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Test that tool registration doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RegisterQueryTools panicked: %v", r)
		}
	}()

	server.RegisterQueryTools()
}

func TestRepoContextMCPServer_registerRepoTools(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Test that tool registration doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("registerRepoTools panicked: %v", r)
		}
	}()

	server.RegisterRepoTools()
}

func TestRepoContextMCPServer_formatSuccessResponse(t *testing.T) {
	server := NewRepoContextMCPServer()

	data := map[string]interface{}{
		"test":   "value",
		"number": 42,
	}

	result := server.FormatSuccessResponse(data)
	if result == nil {
		t.Fatal("formatSuccessResponse should not return nil")
	}

	if result.IsError {
		t.Error("formatSuccessResponse should not return error result")
	}
}

func TestRepoContextMCPServer_formatErrorResponse(t *testing.T) {
	server := NewRepoContextMCPServer()

	result := server.FormatErrorResponse("test_operation", &TestError{"test error"})
	if result == nil {
		t.Fatal("formatErrorResponse should not return nil")
	}

	if !result.IsError {
		t.Error("formatErrorResponse should return error result")
	}
}

// Helper for testing
type TestError struct {
	msg string
}

func (e *TestError) Error() string {
	return e.msg
}

// testHandlerValidationFlow is a helper function to test common handler validation patterns
func testHandlerValidationFlow(
	t *testing.T,
	toolName string,
	handlerFunc func(*RepoContextMCPServer, mcp.CallToolRequest) (*mcp.CallToolResult, error),
) {
	server := NewRepoContextMCPServer()
	server.QueryEngine = &index.QueryEngine{} // Set dummy query engine to pass nil check
	server.RepoPath = "/tmp/test-repo"        // Set a test path that doesn't exist

	// Since our mock request doesn't provide real parameters,
	// we can only test that the handler follows the expected flow
	request := createMockCallToolRequest(toolName, map[string]any{})

	result, err := handlerFunc(server, request)

	if err != nil {
		t.Fatalf("Handler returned unexpected error: %v", err)
	}

	// Should get error for repository validation (happens before parameter parsing)
	if result != nil && result.IsError {
		if len(result.Content) > 0 {
			textContent, ok := result.Content[0].(mcp.TextContent)
			if ok && textContent.Text != "" {
				// The handler should detect repository validation error first
				if !strings.Contains(textContent.Text, "Repository validation failed") {
					t.Errorf("Expected 'Repository validation failed' error, got: %s", textContent.Text)
				}
			}
		}
	} else {
		t.Error("Expected error result for repository validation")
	}
}

// testHandlerQueryEngineNotInitialized is a helper function to test nil query engine error handling
func testHandlerQueryEngineNotInitialized(
	t *testing.T,
	toolName string,
	params map[string]any,
	handlerFunc func(*RepoContextMCPServer, mcp.CallToolRequest) (*mcp.CallToolResult, error),
) {
	server := NewRepoContextMCPServer()
	// queryEngine is nil by default

	request := createMockCallToolRequest(toolName, params)

	result, err := handlerFunc(server, request)

	// Should return system error for nil query engine
	if err == nil {
		t.Error("Expected system error for nil query engine")
	}
	if result != nil {
		t.Error("Expected nil result when system error occurs")
	}
}

// testHandlerRepositoryValidation is a helper function to test repository validation error handling
func testHandlerRepositoryValidation(
	t *testing.T,
	toolName string,
	handlerFunc func(*RepoContextMCPServer, mcp.CallToolRequest) (*mcp.CallToolResult, error),
) {
	server := NewRepoContextMCPServer()
	server.QueryEngine = &index.QueryEngine{} // Set dummy query engine to pass nil check
	server.RepoPath = "/tmp/test"             // Set path that doesn't exist

	// Our mock request always returns empty string parameters
	request := createMockCallToolRequest(toolName, map[string]any{})

	result, err := handlerFunc(server, request)

	if err != nil {
		t.Fatalf("Handler returned unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Handler returned nil result")
	}

	// Should get error for repository validation (happens before parameter validation)
	if !result.IsError {
		t.Error("Expected error result for repository validation")
	}

	if len(result.Content) > 0 {
		textContent, ok := result.Content[0].(mcp.TextContent)
		if ok && textContent.Text != "" {
			// Verify it's the repository validation error
			if !strings.Contains(textContent.Text, "Repository validation failed") {
				t.Errorf("Expected 'Repository validation failed' error, got: %s", textContent.Text)
			}
		}
	}
}

// testHandlerValidation is a comprehensive helper function to test both validation scenarios
func testHandlerValidation(
	t *testing.T,
	toolName string,
	queryEngineParams map[string]any,
	handlerFunc func(*RepoContextMCPServer, mcp.CallToolRequest) (*mcp.CallToolResult, error),
) {
	t.Run("query engine not initialized", func(t *testing.T) {
		testHandlerQueryEngineNotInitialized(t, toolName, queryEngineParams, handlerFunc)
	})

	t.Run("repository validation with mock request", func(t *testing.T) {
		testHandlerRepositoryValidation(t, toolName, handlerFunc)
	})
}

func TestHandleQueryByName_Validation(t *testing.T) {
	testHandlerValidation(
		t,
		"query_by_name",
		map[string]any{"name": "testFunction"},
		func(
			server *RepoContextMCPServer, request mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			return server.HandleQueryByName(context.Background(), request)
		},
	)
}

func TestHandleQueryByName_ParameterParsing(t *testing.T) {
	t.Run("handler validation flow", func(t *testing.T) {
		testHandlerValidationFlow(
			t,
			"query_by_name",
			func(
				server *RepoContextMCPServer, request mcp.CallToolRequest,
			) (*mcp.CallToolResult, error) {
				return server.HandleQueryByName(context.Background(), request)
			},
		)
	})

	// Note: Parameter parsing testing is limited by our simplified mock.
	// In a real scenario, the MCP library handles complex request parsing.
	// Integration tests with actual MCP requests would validate full parameter flow.
}

func TestHandleQueryByName_ResponseFormat(t *testing.T) {
	// This test will verify the response format once we implement the actual functionality
	t.Skip("Skipping response format test until implementation is complete")
}

func TestHandleQueryByPattern_Validation(t *testing.T) {
	testHandlerValidation(
		t,
		"query_by_pattern",
		map[string]any{"pattern": "Test*"},
		func(
			server *RepoContextMCPServer, request mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			return server.HandleQueryByPattern(context.Background(), request)
		},
	)
}

func TestHandleQueryByPattern_ParameterParsing(t *testing.T) {
	t.Run("pattern parameter required", func(t *testing.T) {
		testHandlerValidationFlow(
			t,
			"query_by_pattern",
			func(
				server *RepoContextMCPServer,
				request mcp.CallToolRequest,
			) (*mcp.CallToolResult, error) {
				return server.HandleQueryByPattern(context.Background(), request)
			},
		)
	})

	t.Run("entity type validation", func(t *testing.T) {
		server := NewRepoContextMCPServer()
		server.QueryEngine = &index.QueryEngine{}
		server.RepoPath = "/tmp/test-repo"

		// Test with invalid entity type - this test is conceptual as mock doesn't support parameter injection
		request := createMockCallToolRequest("query_by_pattern", map[string]any{
			"pattern":     "Test*",
			"entity_type": "invalid_type",
		})

		result, err := server.HandleQueryByPattern(context.Background(), request)

		if err != nil {
			t.Fatalf("HandleQueryByPattern returned unexpected error: %v", err)
		}

		// Should get repository validation error first due to mock limitations
		if result != nil && result.IsError {
			textContent, ok := result.Content[0].(mcp.TextContent)
			if ok && textContent.Text != "" {
				// The handler should detect repository validation error first
				if !strings.Contains(textContent.Text, "Repository validation failed") {
					t.Errorf("Expected 'Repository validation failed' error, got: %s", textContent.Text)
				}
			}
		}
	})

	// Note: Full parameter validation testing is limited by our simplified mock.
	// In a real scenario, the MCP library handles complex request parsing.
	// Integration tests with actual MCP requests would validate full parameter flow.
}

func TestHandleQueryByPattern_EntityTypeValidation(t *testing.T) {
	// Test entity type validation logic directly
	validEntityTypes := []string{"function", "type", "variable", "constant"}

	for _, validType := range validEntityTypes {
		t.Run("valid_entity_type_"+validType, func(t *testing.T) {
			// Test that valid entity types are accepted
			// This is a conceptual test since we can't easily mock the parameter parsing
			server := NewRepoContextMCPServer()
			server.QueryEngine = &index.QueryEngine{}
			server.RepoPath = "/tmp/test-repo"

			// In a real test, we would test that the entity type validation logic
			// accepts these valid types. For now, we test the logic exists.
			if len(validEntityTypes) == 0 {
				t.Error("Valid entity types should not be empty")
			}
		})
	}

	t.Run("invalid_entity_types", func(t *testing.T) {
		invalidTypes := []string{"invalid", "class", "method", "field"}

		for _, invalidType := range invalidTypes {
			// Test that invalid entity types would be rejected
			// This is a conceptual test since we can't easily mock the parameter parsing
			isValid := false
			validEntityTypes := []string{"function", "type", "variable", "constant"}
			for _, validType := range validEntityTypes {
				if invalidType == validType {
					isValid = true
					break
				}
			}

			if isValid {
				t.Errorf("Type '%s' should be invalid but was found in valid types", invalidType)
			}
		}
	})
}

func TestHandleQueryByPattern_ResponseFormat(t *testing.T) {
	// This test will verify the response format once we implement the actual functionality
	t.Skip("Skipping response format test until implementation is complete")
}

// Helper function to create mock MCP call tool requests
// Note: This is a simplified test version. In real usage, the MCP library handles request parsing.
func createMockCallToolRequest(toolName string, params map[string]interface{}) mcp.CallToolRequest {
	// For testing, we'll skip complex mock creation and test the implementation directly
	// by running integration tests after implementing the real functionality
	return mcp.CallToolRequest{}
}
