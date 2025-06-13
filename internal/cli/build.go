package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"repository-context-protocol/internal/index"

	"github.com/spf13/cobra"
)

// NewBuildCommand creates the build command for building the semantic index
func NewBuildCommand() *cobra.Command {
	var path string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build semantic index for the repository",
		Long: `Build a semantic index by parsing source code files and creating searchable chunks.

This command:
- Parses all supported source files in the repository
- Extracts functions, types, variables, constants, and call relationships
- Creates semantic chunks for efficient querying
- Stores the index in .repocontext/index.db and .repocontext/chunks/

The repository must be initialized with 'repocontext init' before building.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(path, verbose)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&path, "path", "p", "", "Path to repository root (default: current directory)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}

// runBuild executes the build command logic
func runBuild(path string, verbose bool) error {
	// Determine the target path
	targetPath, err := determineTargetPath(path)
	if err != nil {
		return fmt.Errorf("failed to determine target path: %w", err)
	}

	if verbose {
		fmt.Printf("Building index for repository at: %s\n", targetPath)
	}

	// Validate that the repository is initialized
	if validateErr := validateRepositoryInitialized(targetPath); validateErr != nil {
		return fmt.Errorf("repository not initialized: %w", validateErr)
	}

	// Create and initialize the IndexBuilder
	builder := index.NewIndexBuilder(targetPath)
	if initErr := builder.Initialize(); initErr != nil {
		return fmt.Errorf("failed to initialize index builder: %w", initErr)
	}
	defer func() {
		if closeErr := builder.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close index builder: %v\n", closeErr)
		}
	}()

	if verbose {
		fmt.Println("Starting index build...")
	}

	// Build the index
	stats, err := builder.BuildIndex()
	if err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}

	// Display results
	fmt.Printf("Index built successfully!\n")
	fmt.Printf("Files processed: %d\n", stats.FilesProcessed)
	fmt.Printf("Functions indexed: %d\n", stats.FunctionsIndexed)
	fmt.Printf("Types indexed: %d\n", stats.TypesIndexed)
	fmt.Printf("Variables indexed: %d\n", stats.VariablesIndexed)
	fmt.Printf("Constants indexed: %d\n", stats.ConstantsIndexed)
	fmt.Printf("Call relationships: %d\n", stats.CallsIndexed)
	fmt.Printf("Build duration: %v\n", stats.Duration)

	if verbose {
		fmt.Printf("Index stored in: %s\n", filepath.Join(targetPath, ".repocontext"))
	}

	return nil
}

// determineTargetPath determines the target path for the build operation
func determineTargetPath(path string) (string, error) {
	var targetPath string
	var err error

	if path != "" {
		// Use provided path
		targetPath, err = filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path: %w", err)
		}
	} else {
		// Use current directory
		targetPath, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Verify the path exists and is a directory
	info, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path does not exist: %s", targetPath)
		}
		return "", fmt.Errorf("failed to access path: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", targetPath)
	}

	return targetPath, nil
}

// validateRepositoryInitialized checks if the repository has been initialized
func validateRepositoryInitialized(path string) error {
	repoContextDir := filepath.Join(path, ".repocontext")

	// Check if .repocontext directory exists
	info, err := os.Stat(repoContextDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("repository not initialized - run 'repocontext init' first")
		}
		return fmt.Errorf("failed to access .repocontext directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf(".repocontext exists but is not a directory")
	}

	// Check if manifest.json exists
	manifestPath := filepath.Join(repoContextDir, "manifest.json")
	if _, err := os.Stat(manifestPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("manifest.json not found - repository may be corrupted, try 'repocontext init'")
		}
		return fmt.Errorf("failed to access manifest.json: %w", err)
	}

	return nil
}
