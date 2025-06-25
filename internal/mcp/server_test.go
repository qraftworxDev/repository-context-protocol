package mcp

import (
	"context"
	"testing"
	"time"
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

	// This should detect the current repository root
	repoPath, err := server.detectRepositoryRoot()
	if err != nil {
		t.Fatalf("detectRepositoryRoot failed: %v", err)
	}

	if repoPath == "" {
		t.Error("detectRepositoryRoot should return non-empty path")
	}

	// Should contain .git directory or be a valid git repository
	// We'll implement this validation in the actual function
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

	// Create a context with timeout to test server startup
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This should timeout since we don't have stdin/stdout setup
	err := server.Run(ctx)
	if err == nil {
		t.Error("Run should fail or timeout without proper stdin/stdout setup")
	}
}
