package mcp

import (
	"testing"
)

func TestRepoContextMCPServer_registerQueryTools(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Test that tool registration doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("registerQueryTools panicked: %v", r)
		}
	}()

	server.registerQueryTools()
}

func TestRepoContextMCPServer_registerRepoTools(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Test that tool registration doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("registerRepoTools panicked: %v", r)
		}
	}()

	server.registerRepoTools()
}

func TestRepoContextMCPServer_formatSuccessResponse(t *testing.T) {
	server := NewRepoContextMCPServer()

	data := map[string]interface{}{
		"test":   "value",
		"number": 42,
	}

	result := server.formatSuccessResponse(data)
	if result == nil {
		t.Fatal("formatSuccessResponse should not return nil")
	}

	if result.IsError {
		t.Error("formatSuccessResponse should not return error result")
	}
}

func TestRepoContextMCPServer_formatErrorResponse(t *testing.T) {
	server := NewRepoContextMCPServer()

	result := server.formatErrorResponse("test_operation", &TestError{"test error"})
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
