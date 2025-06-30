package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

	// Expected tools in the correct order
	expectedTools := []struct {
		name        string
		description string
	}{
		{
			name:        "query_by_name",
			description: "Search for functions, types, or variables by exact name with advanced options",
		},
		{
			name: "query_by_pattern",
			description: "Search for entities using glob or regex patterns with advanced filtering " +
				"(supports wildcards *, ?, character classes [abc], brace expansion {a,b}, and regex /pattern/). " +
				"This tool is useful for searching for entities in the repository.",
		},
		{
			name:        "get_call_graph",
			description: "Get detailed call graph for a function with configurable depth and selective inclusion",
		},
		{
			name:        "list_functions",
			description: "List all functions in the repository with pagination and signature control",
		},
		{
			name:        "list_types",
			description: "List all types in the repository with pagination and signature control",
		},
	}

	// Verify exact number of tools returned
	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}

	// Create map for efficient lookup
	toolsByName := make(map[string]mcp.Tool)
	for _, tool := range tools {
		toolsByName[tool.Name] = tool
	}

	// Verify each expected tool is present with correct properties
	for _, expected := range expectedTools {
		tool, exists := toolsByName[expected.name]
		if !exists {
			t.Errorf("Expected tool '%s' not found in registered tools", expected.name)
			continue
		}

		// Validate tool name
		if tool.Name != expected.name {
			t.Errorf("Tool name mismatch: expected '%s', got '%s'", expected.name, tool.Name)
		}

		// Validate tool description
		if tool.Description != expected.description {
			t.Errorf("Tool '%s' description mismatch:\nExpected: %s\nGot: %s", expected.name, expected.description, tool.Description)
		}

		// Validate that tool has a non-empty description
		if tool.Description == "" {
			t.Errorf("Tool '%s' should have a description", tool.Name)
		}

		// Validate that tool has input schema (all tools should have parameters)
		if tool.InputSchema.Type == "" {
			t.Errorf("Tool '%s' should have input schema with type defined", tool.Name)
		}
	}

	// Verify no unexpected tools are present
	for _, tool := range tools {
		found := false
		for _, expected := range expectedTools {
			if tool.Name == expected.name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Unexpected tool '%s' found in registered tools", tool.Name)
		}
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

// TestInitializeRepository tests the initialize_repository tool functionality
func TestInitializeRepository(t *testing.T) {
	t.Run("successful initialization in current directory", func(t *testing.T) {
		// Create a temporary directory to simulate current directory
		tempDir, err := os.MkdirTemp("", "init_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create server and change to temp directory
		server := NewRepoContextMCPServer()
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer func() {
			if err = os.Chdir(originalDir); err != nil {
				t.Fatalf("Failed to change back to original directory: %v", err)
			}
		}()

		if err = os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		// Create a simple mock request without path parameter
		request := mcp.CallToolRequest{}

		result, err := server.HandleInitializeRepository(context.Background(), request)

		// Verify success
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.IsError {
			t.Errorf("Expected success result, got error: %v", result.Content)
		}

		// Verify .repocontext directory was created
		repoContextPath := filepath.Join(tempDir, ".repocontext")
		if _, err := os.Stat(repoContextPath); os.IsNotExist(err) {
			t.Error("Expected .repocontext directory to be created")
		}

		// Verify chunks directory was created
		chunksPath := filepath.Join(repoContextPath, "chunks")
		if _, err := os.Stat(chunksPath); os.IsNotExist(err) {
			t.Error("Expected chunks directory to be created")
		}

		// Verify manifest.json was created
		manifestPath := filepath.Join(repoContextPath, "manifest.json")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			t.Error("Expected manifest.json to be created")
		}
	})

	t.Run("successful initialization with custom path", func(t *testing.T) {
		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "init_custom_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		server := NewRepoContextMCPServer()

		// Create mock request with custom path
		// Since we can't easily mock the MCP request structure, we'll test the underlying logic
		params := &InitializeRepositoryParams{
			Path: tempDir,
		}

		// Test parameter parsing logic
		if params.Path != tempDir {
			t.Errorf("Expected path %s, got %s", tempDir, params.Path)
		}

		// Test path determination
		determinedPath, err := server.determineInitializationPath(params.Path)
		if err != nil {
			t.Fatalf("Failed to determine path: %v", err)
		}

		// Test path validation
		err = server.validateInitializationPath(determinedPath)
		if err != nil {
			t.Fatalf("Path validation failed: %v", err)
		}

		// Test initialization
		result, err := server.initializeRepositoryStructure(determinedPath)
		if err != nil {
			t.Fatalf("Initialization failed: %v", err)
		}

		if result.AlreadyInitialized {
			t.Error("Expected new initialization, got already initialized")
		}
		if len(result.CreatedDirectories) == 0 {
			t.Error("Expected directories to be created")
		}
		if len(result.CreatedFiles) == 0 {
			t.Error("Expected files to be created")
		}
	})

	t.Run("already initialized repository", func(t *testing.T) {
		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "init_existing_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Pre-create .repocontext directory
		repoContextPath := filepath.Join(tempDir, ".repocontext")
		err = os.MkdirAll(repoContextPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create existing .repocontext: %v", err)
		}

		server := NewRepoContextMCPServer()

		// Test initialization on already initialized repository
		result, err := server.initializeRepositoryStructure(tempDir)
		if err != nil {
			t.Fatalf("Initialization failed: %v", err)
		}

		if !result.AlreadyInitialized {
			t.Error("Expected already initialized status")
		}
		if len(result.CreatedDirectories) != 0 {
			t.Error("Expected no new directories to be created")
		}
		if result.Message != "Repository already initialized" {
			t.Errorf("Expected 'Repository already initialized' message, got: %s", result.Message)
		}
	})

	t.Run("invalid path validation", func(t *testing.T) {
		server := NewRepoContextMCPServer()

		// Test with non-existent path
		err := server.validateInitializationPath("/nonexistent/invalid/path")
		if err == nil {
			t.Error("Expected error for non-existent path")
		}

		// Test with file instead of directory
		tempFile, err := os.CreateTemp("", "not_a_dir")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		err = server.validateInitializationPath(tempFile.Name())
		if err == nil {
			t.Error("Expected error for file path instead of directory")
		}
	})

	t.Run("path determination logic", func(t *testing.T) {
		server := NewRepoContextMCPServer()

		// Test empty path (should use current directory)
		path, err := server.determineInitializationPath("")
		if err != nil {
			t.Errorf("Failed to determine path for empty string: %v", err)
		}
		if path == "" {
			t.Error("Expected non-empty path for empty input")
		}

		// Test custom path
		customPath := "/tmp/custom"
		path, err = server.determineInitializationPath(customPath)
		if err != nil {
			t.Errorf("Failed to determine path for custom path: %v", err)
		}
		if !filepath.IsAbs(path) {
			t.Error("Expected absolute path to be returned")
		}
	})

	t.Run("manifest creation", func(t *testing.T) {
		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "manifest_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		server := NewRepoContextMCPServer()
		manifestPath := filepath.Join(tempDir, "manifest.json")

		// Test manifest creation
		err = server.createInitialManifest(manifestPath)
		if err != nil {
			t.Fatalf("Failed to create manifest: %v", err)
		}

		// Verify manifest file exists
		if _, err = os.Stat(manifestPath); os.IsNotExist(err) {
			t.Error("Expected manifest file to be created")
		}

		// Verify manifest content
		manifestContent, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatalf("Failed to read manifest: %v", err)
		}

		var manifest map[string]interface{}
		err = json.Unmarshal(manifestContent, &manifest)
		if err != nil {
			t.Fatalf("Failed to parse manifest JSON: %v", err)
		}

		if manifest["version"] != "1.0" {
			t.Errorf("Expected version '1.0', got: %v", manifest["version"])
		}
		if manifest["description"] == nil {
			t.Error("Expected description field in manifest")
		}
		if manifest["created_at"] == nil {
			t.Error("Expected created_at field in manifest")
		}
	})
}

// TestBuildIndex tests the build_index tool functionality
func TestBuildIndex(t *testing.T) {
	t.Run("successful index build in current directory", func(t *testing.T) {
		// Create a temporary directory to simulate repository
		tempDir, err := os.MkdirTemp("", "build_index_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Initialize repository first
		server := NewRepoContextMCPServer()
		result, err := server.initializeRepositoryStructure(tempDir)
		if err != nil {
			t.Fatalf("Failed to initialize repository: %v", err)
		}
		if result.AlreadyInitialized {
			t.Fatal("Expected new initialization")
		}

		// Create a simple Go file for testing
		goFile := filepath.Join(tempDir, "main.go")
		goContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}

func helper() string {
	return "helper"
}
`
		err = os.WriteFile(goFile, []byte(goContent), ConstFilePermission600)
		if err != nil {
			t.Fatalf("Failed to create test Go file: %v", err)
		}

		// Change to temp directory
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer func() {
			if err = os.Chdir(originalDir); err != nil {
				t.Fatalf("Failed to change back to original directory: %v", err)
			}
		}()

		if err = os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		// Create a simple mock request without path parameter
		request := mcp.CallToolRequest{}

		buildResult, err := server.HandleBuildIndex(context.Background(), request)

		// Verify success
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if buildResult == nil {
			t.Fatal("Expected result, got nil")
		}
		if buildResult.IsError {
			t.Errorf("Expected success result, got error: %v", buildResult.Content)
		}

		// Verify index files were created
		indexPath := filepath.Join(tempDir, ".repocontext", "index.db")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			t.Error("Expected index.db to be created")
		}
	})

	t.Run("successful index build with custom path", func(t *testing.T) {
		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "build_custom_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		server := NewRepoContextMCPServer()

		// Initialize repository first
		_, err = server.initializeRepositoryStructure(tempDir)
		if err != nil {
			t.Fatalf("Failed to initialize repository: %v", err)
		}

		// Create a simple Go file for testing
		goFile := filepath.Join(tempDir, "service.go")
		goContent := `package main

type Service struct {
	name string
}

func (s *Service) GetName() string {
	return s.name
}

func NewService(name string) *Service {
	return &Service{name: name}
}
`
		err = os.WriteFile(goFile, []byte(goContent), ConstFilePermission600)
		if err != nil {
			t.Fatalf("Failed to create test Go file: %v", err)
		}

		// Test parameter parsing logic
		params := &BuildIndexParams{
			Path:    tempDir,
			Verbose: true,
		}

		if params.Path != tempDir {
			t.Errorf("Expected path %s, got %s", tempDir, params.Path)
		}
		if !params.Verbose {
			t.Error("Expected verbose to be true")
		}

		// Test path determination
		determinedPath, err := server.determineBuildPath(params.Path)
		if err != nil {
			t.Fatalf("Failed to determine path: %v", err)
		}

		// Test repository validation
		err = server.validateRepositoryForBuild(determinedPath)
		if err != nil {
			t.Fatalf("Repository validation failed: %v", err)
		}

		// Test index building
		result, err := server.buildRepositoryIndex(determinedPath, params.Verbose)
		if err != nil {
			t.Fatalf("Index build failed: %v", err)
		}

		if result.FilesProcessed == 0 {
			t.Error("Expected files to be processed")
		}
		if result.Success != true {
			t.Error("Expected successful build")
		}
	})

	t.Run("build index on uninitialized repository", func(t *testing.T) {
		// Create a temporary directory without initializing
		tempDir, err := os.MkdirTemp("", "build_uninit_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		server := NewRepoContextMCPServer()

		// Test repository validation on uninitialized repository
		err = server.validateRepositoryForBuild(tempDir)
		if err == nil {
			t.Error("Expected error for uninitialized repository")
		}
		if !strings.Contains(err.Error(), "not initialized") {
			t.Errorf("Expected 'not initialized' error, got: %v", err)
		}
	})

	t.Run("invalid path validation", func(t *testing.T) {
		server := NewRepoContextMCPServer()

		// Test with non-existent path
		err := server.validateRepositoryForBuild("/nonexistent/invalid/path")
		if err == nil {
			t.Error("Expected error for non-existent path")
		}

		// Test with file instead of directory
		tempFile, err := os.CreateTemp("", "not_a_dir")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		err = server.validateRepositoryForBuild(tempFile.Name())
		if err == nil {
			t.Error("Expected error for file path instead of directory")
		}
	})

	t.Run("path determination logic", func(t *testing.T) {
		server := NewRepoContextMCPServer()

		// Test empty path (should use current directory)
		path, err := server.determineBuildPath("")
		if err != nil {
			t.Errorf("Failed to determine path for empty string: %v", err)
		}
		if path == "" {
			t.Error("Expected non-empty path for empty input")
		}

		// Test custom path
		customPath := "/tmp/custom"
		path, err = server.determineBuildPath(customPath)
		if err != nil {
			t.Errorf("Failed to determine path for custom path: %v", err)
		}
		if !filepath.IsAbs(path) {
			t.Error("Expected absolute path to be returned")
		}
	})

	t.Run("verbose mode statistics", func(t *testing.T) {
		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "build_verbose_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		server := NewRepoContextMCPServer()

		// Initialize repository
		_, err = server.initializeRepositoryStructure(tempDir)
		if err != nil {
			t.Fatalf("Failed to initialize repository: %v", err)
		}

		// Create multiple test files
		goFile1 := filepath.Join(tempDir, "file1.go")
		goContent1 := `package main

func function1() {
	// Function 1
}

type Type1 struct {
	Field1 string
}
`
		err = os.WriteFile(goFile1, []byte(goContent1), ConstFilePermission600)
		if err != nil {
			t.Fatalf("Failed to create test Go file 1: %v", err)
		}

		goFile2 := filepath.Join(tempDir, "file2.go")
		goContent2 := `package main

func function2() {
	// Function 2
}

const Constant1 = "value"
var Variable1 string
`
		err = os.WriteFile(goFile2, []byte(goContent2), ConstFilePermission600)
		if err != nil {
			t.Fatalf("Failed to create test Go file 2: %v", err)
		}

		// Test index building with verbose mode
		result, err := server.buildRepositoryIndex(tempDir, true)
		if err != nil {
			t.Fatalf("Index build failed: %v", err)
		}

		// Verify verbose statistics
		if result.FilesProcessed < 2 {
			t.Errorf("Expected at least 2 files processed, got %d", result.FilesProcessed)
		}
		if result.FunctionsIndexed < 2 {
			t.Errorf("Expected at least 2 functions indexed, got %d", result.FunctionsIndexed)
		}
		if result.TypesIndexed < 1 {
			t.Errorf("Expected at least 1 type indexed, got %d", result.TypesIndexed)
		}
		if result.Duration.Seconds() <= 0 {
			t.Error("Expected positive build duration")
		}
	})
}

// TestGetRepositoryStatus tests the get_repository_status tool functionality
func TestGetRepositoryStatus(t *testing.T) {
	t.Run("successful status check with initialized and indexed repository", func(t *testing.T) {
		// Create a temporary directory to simulate repository
		tempDir, err := os.MkdirTemp("", "status_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		server := NewRepoContextMCPServer()

		// Initialize repository first
		initResult, err := server.initializeRepositoryStructure(tempDir)
		if err != nil {
			t.Fatalf("Failed to initialize repository: %v", err)
		}
		if initResult.AlreadyInitialized {
			t.Fatal("Expected new initialization")
		}

		// Create test files for indexing
		goFile := filepath.Join(tempDir, "main.go")
		goContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}

func helper() string {
	return "helper"
}

type Service struct {
	name string
}

const AppName = "test"
var GlobalVar = "global"
`
		err = os.WriteFile(goFile, []byte(goContent), ConstFilePermission600)
		if err != nil {
			t.Fatalf("Failed to create test Go file: %v", err)
		}

		// Build index
		buildResult, err := server.buildRepositoryIndex(tempDir, false)
		if err != nil {
			t.Fatalf("Failed to build index: %v", err)
		}
		if !buildResult.Success {
			t.Fatal("Expected successful index build")
		}

		// Test parameter parsing
		params := server.parseGetRepositoryStatusParameters(mcp.CallToolRequest{})
		if params.Path != "" {
			t.Errorf("Expected empty path for default, got: %s", params.Path)
		}

		// Test path determination
		determinedPath, err := server.determineStatusPath(tempDir)
		if err != nil {
			t.Fatalf("Failed to determine path: %v", err)
		}
		if determinedPath != tempDir {
			t.Errorf("Expected path %s, got %s", tempDir, determinedPath)
		}

		// Test repository status collection
		status, err := server.collectRepositoryStatus(tempDir)
		if err != nil {
			t.Fatalf("Failed to collect repository status: %v", err)
		}

		// Verify status information
		if status.Path != tempDir {
			t.Errorf("Expected path %s, got %s", tempDir, status.Path)
		}
		if !status.IsInitialized {
			t.Error("Expected repository to be initialized")
		}
		if !status.IsIndexed {
			t.Error("Expected repository to be indexed")
		}
		if status.Statistics.FilesProcessed != buildResult.FilesProcessed {
			t.Errorf("Expected %d files processed, got %d", buildResult.FilesProcessed, status.Statistics.FilesProcessed)
		}
		if status.Statistics.FunctionsIndexed != buildResult.FunctionsIndexed {
			t.Errorf("Expected %d functions indexed, got %d", buildResult.FunctionsIndexed, status.Statistics.FunctionsIndexed)
		}
		if status.Statistics.TypesIndexed != buildResult.TypesIndexed {
			t.Errorf("Expected %d types indexed, got %d", buildResult.TypesIndexed, status.Statistics.TypesIndexed)
		}
	})

	t.Run("status check with initialized but not indexed repository", func(t *testing.T) {
		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "status_uninit_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		server := NewRepoContextMCPServer()

		// Initialize repository only (don't build index)
		_, err = server.initializeRepositoryStructure(tempDir)
		if err != nil {
			t.Fatalf("Failed to initialize repository: %v", err)
		}

		// Collect status
		status, err := server.collectRepositoryStatus(tempDir)
		if err != nil {
			t.Fatalf("Failed to collect repository status: %v", err)
		}

		// Verify status information
		if !status.IsInitialized {
			t.Error("Expected repository to be initialized")
		}
		if status.IsIndexed {
			t.Error("Expected repository to not be indexed")
		}
		if status.Statistics.IndexSize > 0 {
			t.Error("Expected zero index size for unindexed repository")
		}
	})

	t.Run("status check with uninitialized repository", func(t *testing.T) {
		// Create a temporary directory without initializing
		tempDir, err := os.MkdirTemp("", "status_uninit_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		server := NewRepoContextMCPServer()

		// Collect status for uninitialized repository
		status, err := server.collectRepositoryStatus(tempDir)
		if err != nil {
			t.Fatalf("Failed to collect repository status: %v", err)
		}

		// Verify status information
		if status.IsInitialized {
			t.Error("Expected repository to not be initialized")
		}
		if status.IsIndexed {
			t.Error("Expected repository to not be indexed")
		}
		if status.Message == "" {
			t.Error("Expected status message for uninitialized repository")
		}
	})

	t.Run("invalid path validation", func(t *testing.T) {
		server := NewRepoContextMCPServer()

		// Test with non-existent path
		_, err := server.collectRepositoryStatus("/nonexistent/invalid/path")
		if err == nil {
			t.Error("Expected error for non-existent path")
		}

		// Test with file instead of directory
		tempFile, err := os.CreateTemp("", "not_a_dir")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		_, err = server.collectRepositoryStatus(tempFile.Name())
		if err == nil {
			t.Error("Expected error for file path instead of directory")
		}
	})

	t.Run("path determination logic", func(t *testing.T) {
		server := NewRepoContextMCPServer()

		// Test empty path (should use current directory)
		path, err := server.determineStatusPath("")
		if err != nil {
			t.Errorf("Failed to determine path for empty string: %v", err)
		}
		if path == "" {
			t.Error("Expected non-empty path for empty input")
		}

		// Test custom path
		customPath := "/tmp/custom"
		path, err = server.determineStatusPath(customPath)
		if err != nil {
			t.Errorf("Failed to determine path for custom path: %v", err)
		}
		if !filepath.IsAbs(path) {
			t.Error("Expected absolute path to be returned")
		}
	})

	t.Run("repository status with detailed statistics", func(t *testing.T) {
		// Create a temporary directory
		tempDir, err := os.MkdirTemp("", "status_detailed_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		server := NewRepoContextMCPServer()

		// Initialize repository
		_, err = server.initializeRepositoryStructure(tempDir)
		if err != nil {
			t.Fatalf("Failed to initialize repository: %v", err)
		}

		// Create multiple test files with varied content
		goFile1 := filepath.Join(tempDir, "service.go")
		goContent1 := `package main

type UserService struct {
	users []User
}

func (s *UserService) GetUser(id int) *User {
	for _, user := range s.users {
		if user.ID == id {
			return &user
		}
	}
	return nil
}

func (s *UserService) AddUser(user User) {
	s.users = append(s.users, user)
}

type User struct {
	ID   int
	Name string
}

const MaxUsers = 1000
var DefaultUser = User{ID: 1, Name: "Default"}
`
		err = os.WriteFile(goFile1, []byte(goContent1), ConstFilePermission600)
		if err != nil {
			t.Fatalf("Failed to create test Go file 1: %v", err)
		}

		goFile2 := filepath.Join(tempDir, "utils.go")
		goContent2 := `package main

func FormatName(name string) string {
	return strings.Title(name)
}

func ValidateEmail(email string) bool {
	return strings.Contains(email, "@")
}

type Config struct {
	Port int
	Host string
}

const Version = "1.0.0"
var Config GlobalConfig = Config{Port: 8080, Host: "localhost"}
`
		err = os.WriteFile(goFile2, []byte(goContent2), ConstFilePermission600)
		if err != nil {
			t.Fatalf("Failed to create test Go file 2: %v", err)
		}

		// Build index
		buildResult, err := server.buildRepositoryIndex(tempDir, true)
		if err != nil {
			t.Fatalf("Failed to build index: %v", err)
		}

		// Collect detailed status
		status, err := server.collectRepositoryStatus(tempDir)
		if err != nil {
			t.Fatalf("Failed to collect repository status: %v", err)
		}

		// Verify detailed statistics
		if status.Statistics.FilesProcessed < 2 {
			t.Errorf("Expected at least 2 files processed, got %d", status.Statistics.FilesProcessed)
		}
		if status.Statistics.FunctionsIndexed < 4 {
			t.Errorf("Expected at least 4 functions indexed, got %d", status.Statistics.FunctionsIndexed)
		}
		if status.Statistics.TypesIndexed < 3 {
			t.Errorf("Expected at least 3 types indexed, got %d", status.Statistics.TypesIndexed)
		}
		if status.Statistics.LastBuildDuration.Seconds() <= 0 {
			t.Error("Expected positive last build duration")
		}
		if status.Statistics.IndexSize <= 0 {
			t.Error("Expected positive index size")
		}

		// Verify that build result matches status statistics
		if status.Statistics.FilesProcessed != buildResult.FilesProcessed {
			t.Errorf("Status files processed (%d) doesn't match build result (%d)",
				status.Statistics.FilesProcessed, buildResult.FilesProcessed)
		}
	})
}

// ============================================================================
// Phase 4.2: Error Recovery Integration Tests for Tool Handlers
// ============================================================================

func TestAdvancedQueryByName_ErrorRecoveryIntegration(t *testing.T) {
	server := NewRepoContextMCPServer()
	ctx := context.Background()

	// Test that the handler correctly integrates with error recovery
	request := mcp.CallToolRequest{}

	// This should trigger the error recovery mechanism because QueryEngine is nil
	_, err := server.HandleAdvancedQueryByName(ctx, request)

	if err == nil {
		t.Error("Expected error from HandleAdvancedQueryByName with nil QueryEngine")
	}

	// Verify circuit breaker was used (should have a failure recorded)
	cb := server.errorRecoveryMgr.GetCircuitBreaker("query_by_name")
	if cb.GetFailureCount() == 0 {
		t.Error("Circuit breaker should have recorded failure")
	}

	// Verify error recovery stats include this operation
	stats := server.GetErrorRecoveryStats()
	circuitBreakers := stats["circuit_breakers"].(map[string]interface{})
	if _, exists := circuitBreakers["query_by_name"]; !exists {
		t.Error("Error recovery stats should include query_by_name circuit breaker")
	}
}
