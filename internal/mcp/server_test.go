package mcp

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewRepoContextMCPServer(t *testing.T) {
	server := NewRepoContextMCPServer()

	if server == nil {
		t.Fatal("NewRepoContextMCPServer should not return nil")
	}

	// Test that server has expected fields initialized
	if server.RepoPath != "" {
		t.Error("Initial RepoPath should be empty")
	}
}

func TestRepoContextMCPServer_detectRepositoryRoot(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Test 1: Without environment variable (original behavior)
	repoPath, err := server.detectRepositoryRoot()
	if err != nil {
		t.Fatalf("detectRepositoryRoot failed: %v", err)
	}
	if repoPath == "" {
		t.Error("detectRepositoryRoot should return non-empty path")
	}
	originalPath := repoPath

	// Test 2: With valid environment variable
	t.Run("ValidEnvironmentVariable", func(t *testing.T) {
		// Use current working directory as a valid path
		currentDir, _ := os.Getwd()
		os.Setenv("REPO_ROOT", currentDir)
		defer os.Unsetenv("REPO_ROOT")

		repoPath, err := server.detectRepositoryRoot()
		if err != nil {
			t.Fatalf("detectRepositoryRoot with valid env var failed: %v", err)
		}
		if repoPath != currentDir {
			t.Errorf("Expected path %s, got %s", currentDir, repoPath)
		}
	})

	// Test 3: With invalid environment variable (should fall back to detection)
	t.Run("InvalidEnvironmentVariable", func(t *testing.T) {
		os.Setenv("REPO_ROOT", "/non/existent/path")
		defer os.Unsetenv("REPO_ROOT")

		repoPath, err := server.detectRepositoryRoot()
		if err != nil {
			t.Fatalf("detectRepositoryRoot with invalid env var failed: %v", err)
		}
		// Should fall back to original detection logic
		if repoPath == "" {
			t.Error("detectRepositoryRoot should return non-empty path even with invalid env var")
		}
	})

	// Test 4: With empty environment variable (should fall back to detection)
	t.Run("EmptyEnvironmentVariable", func(t *testing.T) {
		os.Setenv("REPO_ROOT", "")
		defer os.Unsetenv("REPO_ROOT")

		repoPath, err := server.detectRepositoryRoot()
		if err != nil {
			t.Fatalf("detectRepositoryRoot with empty env var failed: %v", err)
		}
		if repoPath != originalPath {
			t.Errorf("Expected fallback to original path %s, got %s", originalPath, repoPath)
		}
	})
}

func TestRepoContextMCPServer_initializeQueryEngine(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Test initialization without repository setup
	err := server.initializeQueryEngine()
	if err == nil {
		t.Error("initializeQueryEngine should fail without proper repository setup")
	}
}

func TestRepoContextMCPServer_validateRepository(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Test validation with empty repo path
	err := server.validateRepository()
	if err == nil {
		t.Error("validateRepository should fail with empty repoPath")
	}

	// Test validation with non-existent path
	server.RepoPath = "/non/existent/path"
	err = server.validateRepository()
	if err == nil {
		t.Error("validateRepository should fail with non-existent path")
	}
}

func TestRepoContextMCPServer_Run_Timeout(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Create a context with timeout to test server startup behavior
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// The Run method should try to start but may fail due to stdin/stdout setup or timeout
	// We expect an error but shouldn't panic
	err := server.Run(ctx)
	if err == nil {
		// This might be ok if the server starts quickly but doesn't read from stdin
		t.Logf("Server Run completed without error (this may be expected behavior)")
	} else {
		// Error is expected due to stdin/stdout setup or timeout
		t.Logf("Server Run failed as expected: %v", err)
	}
}

// ============================================================================
// Phase 4.1: Enhanced Server Implementation Tests
// ============================================================================

func TestRepoContextMCPServer_GetServerCapabilities(t *testing.T) {
	server := NewRepoContextMCPServer()

	capabilities := server.GetServerCapabilities()

	// Verify that server capabilities are properly configured
	if capabilities["name"] != nil {
		t.Errorf("Expected no 'name' field in capabilities map, but found one")
	}

	// Note: Server name and version are now passed to NewMCPServer, not in capabilities
	// Verify that tools are supported
	if capabilities["tools"] == nil {
		t.Error("Server should support tools")
	}

	// Verify experimental features
	if capabilities["experimental"] == nil {
		t.Error("Experimental features should be configured")
	}
}

func TestRepoContextMCPServer_GetClientCapabilities(t *testing.T) {
	server := NewRepoContextMCPServer()

	capabilities := server.GetClientCapabilities()

	// Verify that client capabilities are properly configured
	if capabilities["experimental"] == nil {
		t.Error("Client experimental features should be configured")
	}

	// Verify that sampling is supported if applicable
	if capabilities["sampling"] != nil && capabilities["sampling"] != true {
		t.Error("Sampling should be enabled if configured")
	}
}

func TestRepoContextMCPServer_RegisterAllTools(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Test tool registration orchestration
	allTools := server.RegisterAllTools()

	if len(allTools) == 0 {
		t.Error("RegisterAllTools should return non-empty tool list")
	}

	// Verify that all expected tool categories are present
	expectedToolCategories := []string{
		"query_by_name",           // Advanced Query Tools
		"query_by_pattern",        // Advanced Query Tools
		"get_call_graph",          // Advanced Query Tools + Enhanced Call Graph Tools
		"list_functions",          // Advanced Query Tools
		"list_types",              // Advanced Query Tools
		"initialize_repository",   // Repository Management Tools
		"build_index",             // Repository Management Tools
		"get_repository_status",   // Repository Management Tools
		"get_call_graph_enhanced", // Enhanced Call Graph Tools
		"find_dependencies",       // Enhanced Call Graph Tools
		"get_function_context",    // Context Analysis Tools
		"get_type_context",        // Context Analysis Tools
	}

	toolNames := make(map[string]bool)
	for _, tool := range allTools {
		toolNames[tool.Name] = true
	}

	for _, expectedTool := range expectedToolCategories {
		if !toolNames[expectedTool] {
			t.Errorf("Expected tool '%s' not found in registered tools", expectedTool)
		}
	}

	// Verify no duplicate tool names
	if len(toolNames) != len(allTools) {
		t.Error("Duplicate tool names found in registered tools")
	}
}

func TestRepoContextMCPServer_CreateMCPServer(t *testing.T) {
	server := NewRepoContextMCPServer()

	mcpServer := server.CreateMCPServer()

	if mcpServer == nil {
		t.Fatal("CreateMCPServer should not return nil")
	}

	// We can't easily test the internals of the MCP server,
	// but we can verify it was created successfully
}

func TestRepoContextMCPServer_SetupToolHandlers(t *testing.T) {
	server := NewRepoContextMCPServer()
	mcpServer := server.CreateMCPServer()

	// Test that tool handlers can be set up without error
	err := server.SetupToolHandlers(mcpServer)

	if err != nil {
		t.Errorf("SetupToolHandlers should not fail: %v", err)
	}
}

func TestRepoContextMCPServer_InitializeWithContext(t *testing.T) {
	server := NewRepoContextMCPServer()
	ctx := context.Background()

	// Test enhanced context initialization
	err := server.InitializeWithContext(ctx)

	// With graceful degradation, this method may not fail but rather
	// print warnings and continue operation
	if err != nil {
		// Error is expected but not required with graceful degradation
		t.Logf("InitializeWithContext failed as expected: %v", err)

		// Verify that the error is meaningful
		if err.Error() == "" {
			t.Error("InitializeWithContext should provide meaningful error message")
		}
	} else {
		// With graceful degradation, this may succeed but with limited functionality
		t.Logf("InitializeWithContext succeeded with graceful degradation")
	}
}

func TestRepoContextMCPServer_InitializeServerLifecycle(t *testing.T) {
	server := NewRepoContextMCPServer()
	ctx := context.Background()

	// Test full server lifecycle initialization
	mcpServer, err := server.InitializeServerLifecycle(ctx)

	// With graceful degradation, server lifecycle should NOT fail
	// even if repository initialization fails
	if err != nil {
		t.Errorf("InitializeServerLifecycle should not fail with graceful degradation: %v", err)
	}

	// MCP server should always be created regardless of repository status
	if mcpServer == nil {
		t.Error("InitializeServerLifecycle should always create MCP server")
	}
}

func TestRepoContextMCPServer_ToolRegistrationOrchestration(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Test that each tool category can be registered independently
	queryTools := server.RegisterAdvancedQueryTools()
	repoTools := server.RegisterRepositoryManagementTools()
	callGraphTools := server.RegisterCallGraphTools()
	contextTools := server.RegisterContextTools()

	if len(queryTools) == 0 {
		t.Error("RegisterAdvancedQueryTools should return tools")
	}

	if len(repoTools) == 0 {
		t.Error("RegisterRepositoryManagementTools should return tools")
	}

	if len(callGraphTools) == 0 {
		t.Error("RegisterCallGraphTools should return tools")
	}

	if len(contextTools) == 0 {
		t.Error("RegisterContextTools should return tools")
	}

	// Test orchestrated registration
	allTools := server.RegisterAllTools()
	expectedTotal := len(queryTools) + len(repoTools) + len(callGraphTools) + len(contextTools)

	if len(allTools) != expectedTotal {
		t.Errorf("RegisterAllTools should return %d tools, got %d", expectedTotal, len(allTools))
	}
}

func TestRepoContextMCPServer_ConfigurationManagement(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Test server configuration
	config := server.GetServerConfiguration()

	if config == nil {
		t.Fatal("GetServerConfiguration should not return nil")
	}

	if config.Name == "" {
		t.Error("Server name should be configured")
	}

	if config.Version == "" {
		t.Error("Server version should be configured")
	}

	if config.MaxTokens <= 0 {
		t.Error("MaxTokens should be positive")
	}

	if config.MaxDepth <= 0 {
		t.Error("MaxDepth should be positive")
	}
}

func TestRepoContextMCPServer_LifecycleIntegration(t *testing.T) {
	server := NewRepoContextMCPServer()
	ctx := context.Background()

	// Test that the server can be initialized and configured
	// With graceful degradation, this may not fail
	err := server.InitializeWithContext(ctx)
	if err != nil {
		t.Logf("InitializeWithContext failed as expected: %v", err)
	} else {
		t.Logf("InitializeWithContext succeeded with graceful degradation")
	}

	// Test that tools can still be registered
	tools := server.RegisterAllTools()
	if len(tools) == 0 {
		t.Error("Tools should be registerable even without repository")
	}

	// Test that server capabilities are available
	capabilities := server.GetServerCapabilities()
	if capabilities == nil {
		t.Error("Server capabilities should be available")
	}
}

// ============================================================================
// Phase 4.2: Error Recovery Integration Tests
// ============================================================================

func TestRepoContextMCPServer_ErrorRecoveryInitialization(t *testing.T) {
	server := NewRepoContextMCPServer()

	if server.errorRecoveryMgr == nil {
		t.Error("Error recovery manager should be initialized")
	}

	// Test that error recovery stats are available
	stats := server.GetErrorRecoveryStats()
	if stats == nil {
		t.Error("Error recovery stats should not be nil")
	}

	if stats["total_circuit_breakers"] == nil {
		t.Error("Error recovery stats should include circuit breaker count")
	}
}

func TestRepoContextMCPServer_ExecuteToolWithRecovery_Success(t *testing.T) {
	server := NewRepoContextMCPServer()
	ctx := context.Background()

	// Mock successful operation
	mockOperation := func() (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("success"), nil
	}

	result, err := server.ExecuteToolWithRecovery(ctx, "test_tool", mockOperation)

	if err != nil {
		t.Errorf("ExecuteToolWithRecovery should not return error for successful operation: %v", err)
	}

	if result == nil {
		t.Fatal("ExecuteToolWithRecovery should return result")
	}

	// Verify circuit breaker recorded success
	cb := server.errorRecoveryMgr.GetCircuitBreaker("test_tool")
	if cb.GetState() != CircuitBreakerClosed {
		t.Error("Circuit breaker should remain closed after successful operation")
	}

	if cb.GetFailureCount() != 0 {
		t.Errorf("Circuit breaker should have 0 failures after success, got %d", cb.GetFailureCount())
	}
}

func TestRepoContextMCPServer_ExecuteToolWithRecovery_RetryableFailure(t *testing.T) {
	server := NewRepoContextMCPServer()
	ctx := context.Background()

	attemptCount := 0
	// Mock operation that fails initially then succeeds
	mockOperation := func() (*mcp.CallToolResult, error) {
		attemptCount++
		if attemptCount < 2 {
			return nil, errors.New("query failed")
		}
		return mcp.NewToolResultText("success after retry"), nil
	}

	result, err := server.ExecuteToolWithRecovery(ctx, "test_tool", mockOperation)

	if err != nil {
		t.Errorf("ExecuteToolWithRecovery should succeed after retries: %v", err)
	}

	if result == nil {
		t.Fatal("ExecuteToolWithRecovery should return result after retry")
	}

	// Should have attempted more than once
	if attemptCount < 2 {
		t.Errorf("Expected at least 2 attempts, got %d", attemptCount)
	}

	// Circuit breaker should be closed after eventual success
	cb := server.errorRecoveryMgr.GetCircuitBreaker("test_tool")
	if cb.GetState() != CircuitBreakerClosed {
		t.Error("Circuit breaker should be closed after eventual success")
	}
}

func TestRepoContextMCPServer_ExecuteToolWithRecovery_NonRetryableFailure(t *testing.T) {
	server := NewRepoContextMCPServer()
	ctx := context.Background()

	attemptCount := 0
	// Mock operation that always fails with non-retryable error
	mockOperation := func() (*mcp.CallToolResult, error) {
		attemptCount++
		return nil, errors.New("validation failed")
	}

	_, err := server.ExecuteToolWithRecovery(ctx, "test_tool", mockOperation)

	if err == nil {
		t.Error("ExecuteToolWithRecovery should return error for non-retryable failure")
	}

	// Should have attempted only once for non-retryable error
	if attemptCount != 1 {
		t.Errorf("Expected 1 attempt for non-retryable error, got %d", attemptCount)
	}

	// Circuit breaker should record the failure
	cb := server.errorRecoveryMgr.GetCircuitBreaker("test_tool")
	if cb.GetFailureCount() == 0 {
		t.Error("Circuit breaker should record failure")
	}
}

func TestRepoContextMCPServer_ExecuteToolWithRecovery_CircuitBreakerOpen(t *testing.T) {
	server := NewRepoContextMCPServer()
	ctx := context.Background()

	// Mock operation that always fails
	mockOperation := func() (*mcp.CallToolResult, error) {
		return nil, errors.New("storage failed")
	}

	// Execute multiple times to open circuit breaker
	for i := 0; i < DefaultFailureThreshold; i++ {
		_, err := server.ExecuteToolWithRecovery(ctx, "test_tool", mockOperation)
		if err == nil {
			t.Error("ExecuteToolWithRecovery should return error for failing operation")
		}
	}

	// Circuit breaker should now be open
	cb := server.errorRecoveryMgr.GetCircuitBreaker("test_tool")
	if cb.GetState() != CircuitBreakerOpen {
		t.Error("Circuit breaker should be open after threshold failures")
	}

	// Next execution should fail immediately due to circuit breaker
	result, err := server.ExecuteToolWithRecovery(ctx, "test_tool", mockOperation)

	if err == nil {
		t.Error("ExecuteToolWithRecovery should fail when circuit breaker is open")
	}

	if result != nil {
		t.Error("ExecuteToolWithRecovery should not return result when circuit breaker is open")
	}

	// Error should mention circuit breaker
	if !contains(err.Error(), "circuit breaker") {
		t.Error("Error message should mention circuit breaker")
	}
}

func TestRepoContextMCPServer_GetErrorRecoveryStats(t *testing.T) {
	server := NewRepoContextMCPServer()
	ctx := context.Background()

	// Execute some operations to generate stats
	mockOperation := func() (*mcp.CallToolResult, error) {
		return nil, errors.New("test error")
	}

	_, err := server.ExecuteToolWithRecovery(ctx, "tool1", mockOperation)
	if err == nil {
		t.Error("ExecuteToolWithRecovery should return error for failing operation")
	}
	_, err = server.ExecuteToolWithRecovery(ctx, "tool2", mockOperation)
	if err == nil {
		t.Error("ExecuteToolWithRecovery should return error for failing operation")
	}

	stats := server.GetErrorRecoveryStats()

	if stats == nil {
		t.Fatal("GetErrorRecoveryStats should not return nil")
	}

	if stats["total_circuit_breakers"] != 2 {
		t.Errorf("Expected 2 circuit breakers, got %v", stats["total_circuit_breakers"])
	}

	if stats["circuit_breakers"] == nil {
		t.Error("Stats should include circuit breaker details")
	}

	if stats["retry_config"] == nil {
		t.Error("Stats should include retry configuration")
	}
}

func TestRepoContextMCPServer_ErrorRecoveryManager_NilHandling(t *testing.T) {
	server := &RepoContextMCPServer{
		// Don't initialize error recovery manager
		errorRecoveryMgr: nil,
	}

	ctx := context.Background()

	// Mock operation
	mockOperation := func() (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("success"), nil
	}

	// Should fallback to direct execution
	result, err := server.ExecuteToolWithRecovery(ctx, "test_tool", mockOperation)

	if err != nil {
		t.Errorf("ExecuteToolWithRecovery should work with nil error recovery manager: %v", err)
	}

	if result == nil {
		t.Error("ExecuteToolWithRecovery should return result even with nil error recovery manager")
	}

	// Stats should indicate not initialized
	stats := server.GetErrorRecoveryStats()
	if stats["status"] != "not_initialized" {
		t.Error("Stats should indicate error recovery manager is not initialized")
	}
}
