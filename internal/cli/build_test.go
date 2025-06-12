package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewBuildCommand(t *testing.T) {
	cmd := NewBuildCommand()

	if cmd == nil {
		t.Fatal("Expected NewBuildCommand to return a command, got nil")
	}

	if cmd.Use != "build" {
		t.Errorf("Expected command use 'build', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command to have a short description")
	}

	if cmd.RunE == nil {
		t.Error("Expected command to have a RunE function")
	}
}

func TestBuildCommand_BuildIndex(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "build_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize the repository first
	err = initializeRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Create test Go files
	testFiles := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	user := CreateUser("Alice")
	fmt.Println(user.String())
}`,
		"user.go": `package main

type User struct {
	Name string
}

func CreateUser(name string) *User {
	return &User{Name: name}
}

func (u *User) String() string {
	return "User: " + u.Name
}`,
	}

	// Write test files to temp directory
	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		err = os.WriteFile(filePath, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("Failed to restore original directory: %v", chErr)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Run build command
	cmd := NewBuildCommand()
	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Build command failed: %v", err)
	}

	// Verify index was created
	indexDir := filepath.Join(tempDir, ".repocontext")
	if _, err := os.Stat(indexDir); os.IsNotExist(err) {
		t.Error("Expected .repocontext directory to exist after build")
	}

	// Verify index database was created
	dbPath := filepath.Join(indexDir, "index.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Expected index.db to be created")
	}

	// Verify chunks directory was created
	chunksDir := filepath.Join(indexDir, "chunks")
	if _, err := os.Stat(chunksDir); os.IsNotExist(err) {
		t.Error("Expected chunks directory to be created")
	}
}

func TestBuildCommand_WithCustomPath(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "build_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory to build
	projectDir := filepath.Join(tempDir, "myproject")
	err = os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Initialize the repository
	err = initializeRepository(projectDir)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Create test file
	testFile := filepath.Join(projectDir, "main.go")
	testContent := `package main

func main() {
	println("Hello, World!")
}`
	err = os.WriteFile(testFile, []byte(testContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run build command with custom path
	cmd := NewBuildCommand()

	// Set the path flag
	if setErr := cmd.Flags().Set("path", projectDir); setErr != nil {
		t.Fatalf("Failed to set path flag: %v", setErr)
	}

	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Build command failed with custom path: %v", err)
	}

	// Verify index was created in custom path
	indexDir := filepath.Join(projectDir, ".repocontext")
	if _, err := os.Stat(indexDir); os.IsNotExist(err) {
		t.Errorf("Expected .repocontext directory to be created at %s", indexDir)
	}
}

func TestBuildCommand_NotInitialized(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "build_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Don't initialize the repository - should fail

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("Failed to restore original directory: %v", chErr)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Run build command - should fail
	cmd := NewBuildCommand()
	err = cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected build command to fail when repository not initialized")
	}
}

func TestBuildCommand_InvalidPath(t *testing.T) {
	// Try to build in a non-existent directory
	cmd := NewBuildCommand()

	// Set an invalid path
	invalidPath := "/nonexistent/invalid/path"
	if setErr := cmd.Flags().Set("path", invalidPath); setErr != nil {
		t.Fatalf("Failed to set path flag: %v", setErr)
	}

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected build command to fail with invalid path")
	}
}

func TestBuildCommand_EmptyDirectory(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "build_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize the repository
	err = initializeRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Change to temp directory (no source files)
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("Failed to restore original directory: %v", chErr)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Run build command - should succeed but with no files processed
	cmd := NewBuildCommand()
	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Build command should succeed with empty directory: %v", err)
	}

	// Verify index was still created
	indexDir := filepath.Join(tempDir, ".repocontext")
	if _, err := os.Stat(indexDir); os.IsNotExist(err) {
		t.Error("Expected .repocontext directory to exist even with empty directory")
	}
}

func TestBuildCommand_WithVerboseOutput(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "build_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize the repository
	err = initializeRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Create test file
	testFile := filepath.Join(tempDir, "test.go")
	testContent := `package main

func test() {
	println("test")
}`
	err = os.WriteFile(testFile, []byte(testContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("Failed to restore original directory: %v", chErr)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Run build command with verbose flag
	cmd := NewBuildCommand()

	// Set the verbose flag
	if setErr := cmd.Flags().Set("verbose", "true"); setErr != nil {
		t.Fatalf("Failed to set verbose flag: %v", setErr)
	}

	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Build command failed with verbose flag: %v", err)
	}

	// Verify index was created
	indexDir := filepath.Join(tempDir, ".repocontext")
	if _, err := os.Stat(indexDir); os.IsNotExist(err) {
		t.Error("Expected .repocontext directory to exist")
	}
}

func TestBuildCommand_RebuildExistingIndex(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "build_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize the repository
	err = initializeRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Create initial test file
	testFile := filepath.Join(tempDir, "initial.go")
	initialContent := `package main

func initial() {
	println("initial")
}`
	err = os.WriteFile(testFile, []byte(initialContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create initial test file: %v", err)
	}

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("Failed to restore original directory: %v", chErr)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Run build command first time
	cmd := NewBuildCommand()
	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("First build command failed: %v", err)
	}

	// Add another test file
	testFile2 := filepath.Join(tempDir, "updated.go")
	updatedContent := `package main

func updated() {
	println("updated")
}`
	err = os.WriteFile(testFile2, []byte(updatedContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create updated test file: %v", err)
	}

	// Run build command second time (rebuild)
	cmd2 := NewBuildCommand()
	err = cmd2.RunE(cmd2, []string{})
	if err != nil {
		t.Fatalf("Second build command failed: %v", err)
	}

	// Verify index still exists and was updated
	indexDir := filepath.Join(tempDir, ".repocontext")
	if _, err := os.Stat(indexDir); os.IsNotExist(err) {
		t.Error("Expected .repocontext directory to exist after rebuild")
	}
}

// Helper function to initialize a repository for testing
func initializeRepository(path string) error {
	// Create .repocontext directory
	repoContextDir := filepath.Join(path, ".repocontext")
	if err := os.MkdirAll(repoContextDir, 0755); err != nil {
		return err
	}

	// Create manifest.json
	manifestPath := filepath.Join(repoContextDir, "manifest.json")
	manifestContent := `{
  "version": "1.0.0",
  "created_at": "2024-01-01T00:00:00Z",
  "description": "Test repository context"
}`
	return os.WriteFile(manifestPath, []byte(manifestContent), 0600)
}
