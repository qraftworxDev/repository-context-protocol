package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"repository-context-protocol/internal/models"

	"github.com/spf13/cobra"
)

const (
	// Directory permissions for .repocontext and subdirectories
	repoContextDirPermissions = 0755
	// File permissions for manifest.json
	manifestFilePermissions = 0644
)

// NewInitCommand creates the init command for initializing a repository
func NewInitCommand() *cobra.Command {
	var path string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a repository for semantic indexing",
		Long: `Initialize a repository by creating the .repocontext directory structure.

This command creates:
- .repocontext/ directory for storing index data
- .repocontext/chunks/ directory for MessagePack chunk storage
- .repocontext/manifest.json file for chunk metadata

By default, initializes in the current directory. Use --path to specify a different location.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(path)
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", ".", "Path to initialize (defaults to current directory)")

	return cmd
}

// runInit performs the actual initialization logic
func runInit(targetPath string) error {
	// Resolve the target path to absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path %s: %w", targetPath, err)
	}

	// Verify the target directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("target directory does not exist: %s", absPath)
	}

	// Define the .repocontext directory path
	repoContextDir := filepath.Join(absPath, ".repocontext")

	// Check if already initialized
	if _, err := os.Stat(repoContextDir); err == nil {
		fmt.Printf("Repository already initialized at %s\n", repoContextDir)
		// Still ensure all subdirectories exist
		if err := createSubdirectories(repoContextDir); err != nil {
			return fmt.Errorf("failed to ensure subdirectories exist: %w", err)
		}
		// Ensure manifest exists
		if err := createManifest(repoContextDir); err != nil {
			return fmt.Errorf("failed to ensure manifest exists: %w", err)
		}
		return nil
	}

	// Create .repocontext directory
	if err := os.MkdirAll(repoContextDir, repoContextDirPermissions); err != nil {
		return fmt.Errorf("failed to create .repocontext directory: %w", err)
	}

	// Create subdirectories
	if err := createSubdirectories(repoContextDir); err != nil {
		return fmt.Errorf("failed to create subdirectories: %w", err)
	}

	// Create initial manifest.json
	if err := createManifest(repoContextDir); err != nil {
		return fmt.Errorf("failed to create manifest: %w", err)
	}

	fmt.Printf("Initialized repository at %s\n", repoContextDir)
	return nil
}

// createSubdirectories creates the necessary subdirectories within .repocontext
func createSubdirectories(repoContextDir string) error {
	subdirs := []string{
		"chunks", // For MessagePack chunk storage
	}

	for _, subdir := range subdirs {
		subdirPath := filepath.Join(repoContextDir, subdir)
		if err := os.MkdirAll(subdirPath, repoContextDirPermissions); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", subdir, err)
		}
	}

	return nil
}

// createManifest creates the initial manifest.json file
func createManifest(repoContextDir string) error {
	manifestPath := filepath.Join(repoContextDir, "manifest.json")

	// Check if manifest already exists
	if _, err := os.Stat(manifestPath); err == nil {
		// Manifest already exists, don't overwrite
		return nil
	}

	// Create initial manifest
	manifest := models.Manifest{
		Version:   "1.0.0",
		Chunks:    make(map[string]models.ChunkInfo),
		UpdatedAt: time.Now(),
	}

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Write to file
	if err := os.WriteFile(manifestPath, data, manifestFilePermissions); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}
