package mcp

import (
	"context"
	"testing"

	"repository-context-protocol/internal/index"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestRepoContextMCPServer_RegisterQueryTools tests basic tool registration
func TestRepoContextMCPServer_RegisterQueryTools(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Test that tool registration doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RegisterQueryTools panicked: %v", r)
		}
	}()

	tools := server.RegisterQueryTools()

	// Verify we get tools back
	if len(tools) == 0 {
		t.Error("RegisterQueryTools should return tools")
	}
}

// TestRepoContextMCPServer_RegisterRepoTools tests repo tools registration
func TestRepoContextMCPServer_RegisterRepoTools(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Test that tool registration doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RegisterRepoTools panicked: %v", r)
		}
	}()

	server.RegisterRepoTools()
}

// TestAdvancedHandlers_ValidationFlow tests the advanced handlers follow proper validation flow
func TestAdvancedHandlers_ValidationFlow(t *testing.T) {
	testCases := []struct {
		name        string
		toolName    string
		handlerFunc func(*RepoContextMCPServer, context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	}{
		{
			name:     "HandleAdvancedQueryByName",
			toolName: "query_by_name",
			handlerFunc: func(server *RepoContextMCPServer, ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return server.HandleAdvancedQueryByName(ctx, request)
			},
		},
		{
			name:     "HandleAdvancedQueryByPattern",
			toolName: "query_by_pattern",
			handlerFunc: func(server *RepoContextMCPServer, ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return server.HandleAdvancedQueryByPattern(ctx, request)
			},
		},
		{
			name:     "HandleAdvancedGetCallGraph",
			toolName: "get_call_graph",
			handlerFunc: func(server *RepoContextMCPServer, ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return server.HandleAdvancedGetCallGraph(ctx, request)
			},
		},
		{
			name:     "HandleAdvancedListFunctions",
			toolName: "list_functions",
			handlerFunc: func(server *RepoContextMCPServer, ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return server.HandleAdvancedListFunctions(ctx, request)
			},
		},
		{
			name:     "HandleAdvancedListTypes",
			toolName: "list_types",
			handlerFunc: func(server *RepoContextMCPServer, ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return server.HandleAdvancedListTypes(ctx, request)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test system error for nil query engine
			t.Run("nil_query_engine", func(t *testing.T) {
				server := NewRepoContextMCPServer()
				// QueryEngine is nil by default

				request := mcp.CallToolRequest{}
				_, err := tc.handlerFunc(server, context.Background(), request)

				if err == nil {
					t.Error("Expected system error for nil query engine")
				}
			})

			// Test repository validation error
			t.Run("repository_validation", func(t *testing.T) {
				server := NewRepoContextMCPServer()
				server.QueryEngine = &index.QueryEngine{} // Set dummy query engine
				server.RepoPath = "/tmp/nonexistent"      // Path that doesn't exist

				request := mcp.CallToolRequest{}
				result, err := tc.handlerFunc(server, context.Background(), request)

				if err != nil {
					t.Fatalf("Handler should not return system error: %v", err)
				}

				if result == nil || !result.IsError {
					t.Error("Expected error result for repository validation")
				}
			})
		})
	}
}

// TestResponseFormatting tests the shared response formatting methods
func TestResponseFormatting(t *testing.T) {
	server := NewRepoContextMCPServer()

	t.Run("FormatSuccessResponse", func(t *testing.T) {
		data := map[string]interface{}{
			"test":   "value",
			"number": 42,
		}

		result := server.FormatSuccessResponse(data)
		if result == nil {
			t.Fatal("FormatSuccessResponse should not return nil")
		}

		if result.IsError {
			t.Error("FormatSuccessResponse should not return error result")
		}

		if len(result.Content) == 0 {
			t.Error("Expected content in response")
		}
	})

	t.Run("FormatErrorResponse", func(t *testing.T) {
		testErr := &TestError{"test error"}

		result := server.FormatErrorResponse("test_operation", testErr)
		if result == nil {
			t.Fatal("FormatErrorResponse should not return nil")
		}

		if !result.IsError {
			t.Error("FormatErrorResponse should return error result")
		}

		if len(result.Content) == 0 {
			t.Error("Expected content in error response")
		}
	})
}

// Helper for testing
type TestError struct {
	msg string
}

func (e *TestError) Error() string {
	return e.msg
}
