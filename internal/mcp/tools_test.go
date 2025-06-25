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

func TestHandleQueryByName_Validation(t *testing.T) {
	t.Run("query engine not initialized", func(t *testing.T) {
		server := NewRepoContextMCPServer()
		// queryEngine is nil by default

		request := createMockCallToolRequest("query_by_name", map[string]any{"name": "testFunction"})

		result, err := server.HandleQueryByName(context.Background(), request)

		// Should return system error for nil query engine
		if err == nil {
			t.Error("Expected system error for nil query engine")
		}
		if result != nil {
			t.Error("Expected nil result when system error occurs")
		}
	})

	t.Run("repository validation with mock request", func(t *testing.T) {
		server := NewRepoContextMCPServer()
		server.QueryEngine = &index.QueryEngine{} // Set dummy query engine to pass nil check
		server.RepoPath = "/tmp/test"             // Set path that doesn't exist

		// Our mock request always returns empty string for name parameter
		request := createMockCallToolRequest("query_by_name", map[string]any{})

		result, err := server.HandleQueryByName(context.Background(), request)

		if err != nil {
			t.Fatalf("HandleQueryByName returned unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("HandleQueryByName returned nil result")
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
	})
}

func TestHandleQueryByName_ParameterParsing(t *testing.T) {
	t.Run("handler validation flow", func(t *testing.T) {
		server := NewRepoContextMCPServer()
		server.QueryEngine = &index.QueryEngine{} // Set dummy query engine to pass nil check
		server.RepoPath = "/tmp/test-repo"        // Set a test path that doesn't exist

		// Since our mock request doesn't provide real parameters,
		// we can only test that the handler follows the expected flow
		request := createMockCallToolRequest("query_by_name", map[string]any{})

		result, err := server.HandleQueryByName(context.Background(), request)

		if err != nil {
			t.Fatalf("HandleQueryByName returned unexpected error: %v", err)
		}

		// Should get error for repository validation (happens before parameter parsing)
		if result != nil && result.IsError {
			textContent, ok := result.Content[0].(mcp.TextContent)
			if ok && textContent.Text != "" {
				// The handler should detect repository validation error first
				if !strings.Contains(textContent.Text, "Repository validation failed") {
					t.Errorf("Expected 'Repository validation failed' error, got: %s", textContent.Text)
				}
			}
		} else {
			t.Error("Expected error result for repository validation")
		}
	})

	// Note: Parameter parsing testing is limited by our simplified mock.
	// In a real scenario, the MCP library handles complex request parsing.
	// Integration tests with actual MCP requests would validate full parameter flow.
}

func TestHandleQueryByName_ResponseFormat(t *testing.T) {
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
