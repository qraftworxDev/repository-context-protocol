package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
)

func TestNewInitCommand(t *testing.T) {
	cmd := NewInitCommand()

	if cmd == nil {
		t.Fatal("Expected NewInitCommand to return a command, got nil")
	}

	if cmd.Use != "init" {
		t.Errorf("Expected command use 'init', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command to have a short description")
	}

	if cmd.RunE == nil {
		t.Error("Expected command to have a RunE function")
	}
}

func TestInitCommand_CreateRepoContextDirectory(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "init_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

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

	// Run init command
	cmd := NewInitCommand()
	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Init command failed: %v", err)
	}

	// Verify .repocontext directory was created
	repoContextDir := filepath.Join(tempDir, ".repocontext")
	if _, statErr := os.Stat(repoContextDir); os.IsNotExist(statErr) {
		t.Errorf("Expected .repocontext directory to be created at %s", repoContextDir)
	}

	// Verify directory has correct permissions
	info, err := os.Stat(repoContextDir)
	if err != nil {
		t.Fatalf("Failed to stat .repocontext directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("Expected .repocontext to be a directory")
	}

	// Check permissions (should respect system umask)
	// Get current umask to calculate expected permissions
	currentUmask := syscall.Umask(0) // Get current umask
	syscall.Umask(currentUmask)      // Restore the umask immediately

	// Calculate expected permissions: intended permissions (0755) masked by umask
	// Ensure umask is within valid range for uint32 conversion
	if currentUmask < 0 || currentUmask > 0777 {
		t.Fatalf("Invalid umask value: %d", currentUmask)
	}

	intendedPerm := os.FileMode(0755)
	expectedPerm := intendedPerm &^ os.FileMode(uint32(currentUmask))

	actualPerm := info.Mode().Perm()
	if actualPerm != expectedPerm {
		t.Errorf("Expected .repocontext directory permissions %v (0755 masked by umask %03o), got %v",
			expectedPerm, currentUmask, actualPerm)
	}
}

func TestInitCommand_CreateSubdirectories(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "init_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

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

	// Run init command
	cmd := NewInitCommand()
	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Init command failed: %v", err)
	}

	// Verify chunks subdirectory was created
	chunksDir := filepath.Join(tempDir, ".repocontext", "chunks")
	if _, statErr := os.Stat(chunksDir); os.IsNotExist(statErr) {
		t.Errorf("Expected chunks directory to be created at %s", chunksDir)
	}

	// Verify chunks directory is actually a directory
	info, err := os.Stat(chunksDir)
	if err != nil {
		t.Fatalf("Failed to stat chunks directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("Expected chunks to be a directory")
	}
}

func TestInitCommand_AlreadyInitialized(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "init_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

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

	// Create .repocontext directory manually
	repoContextDir := filepath.Join(tempDir, ".repocontext")
	err = os.MkdirAll(repoContextDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .repocontext directory: %v", err)
	}

	// Run init command - should not fail but should detect existing directory
	cmd := NewInitCommand()
	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Init command should not fail when directory exists: %v", err)
	}

	// Directory should still exist
	if _, err := os.Stat(repoContextDir); os.IsNotExist(err) {
		t.Error("Expected .repocontext directory to still exist")
	}
}

func TestInitCommand_WithCustomPath(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "init_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory to init in
	projectDir := filepath.Join(tempDir, "myproject")
	err = os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Run init command with custom path
	cmd := NewInitCommand()

	// Set the path flag
	if setErr := cmd.Flags().Set("path", projectDir); setErr != nil {
		t.Fatalf("Failed to set path flag: %v", setErr)
	}

	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Init command failed with custom path: %v", err)
	}

	// Verify .repocontext directory was created in custom path
	repoContextDir := filepath.Join(projectDir, ".repocontext")
	if _, err := os.Stat(repoContextDir); os.IsNotExist(err) {
		t.Errorf("Expected .repocontext directory to be created at %s", repoContextDir)
	}
}

func TestInitCommand_InvalidPath(t *testing.T) {
	// Try to init in a non-existent directory
	cmd := NewInitCommand()

	// Set an invalid path
	invalidPath := "/nonexistent/invalid/path"
	if setErr := cmd.Flags().Set("path", invalidPath); setErr != nil {
		t.Fatalf("Failed to set path flag: %v", setErr)
	}

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected init command to fail with invalid path")
	}
}

func TestInitCommand_PermissionDenied(t *testing.T) {
	// Skip this test on Windows as permission handling is different
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "init_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a directory that will be made inaccessible
	restrictedDir := filepath.Join(tempDir, "restricted")
	err = os.MkdirAll(restrictedDir, 0755) // Create with normal permissions first
	if err != nil {
		t.Fatalf("Failed to create restricted directory: %v", err)
	}

	// Make the directory completely inaccessible (no permissions)
	err = os.Chmod(restrictedDir, 0000)
	if err != nil {
		t.Fatalf("Failed to set restrictive permissions: %v", err)
	}

	// Ensure we restore permissions for cleanup
	defer func() {
		// Restore permissions so the directory can be removed
		if chmodErr := os.Chmod(restrictedDir, 0755); chmodErr != nil {
			t.Logf("Warning: Failed to restore permissions for cleanup: %v", chmodErr)
		}
	}()

	// Try to init in the inaccessible directory
	cmd := NewInitCommand()
	if setErr := cmd.Flags().Set("path", restrictedDir); setErr != nil {
		t.Fatalf("Failed to set path flag: %v", setErr)
	}

	err = cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("Expected init command to fail with permission denied")
	}
}

func TestInitCommand_CreateManifest(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "init_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

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

	// Run init command
	cmd := NewInitCommand()
	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Init command failed: %v", err)
	}

	// Verify manifest.json was created
	manifestPath := filepath.Join(tempDir, ".repocontext", "manifest.json")
	if _, statErr := os.Stat(manifestPath); os.IsNotExist(statErr) {
		t.Errorf("Expected manifest.json to be created at %s", manifestPath)
	}

	// Verify manifest.json has valid content
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read manifest.json: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected manifest.json to have content")
	}

	// Should be valid JSON - test by attempting to unmarshal
	var jsonData interface{}
	if err := json.Unmarshal(content, &jsonData); err != nil {
		t.Errorf("Expected manifest.json to contain valid JSON, but got error: %v", err)
	}
}
