package index

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"repository-context-protocol/internal/ast"
	"repository-context-protocol/internal/ast/golang"
	"repository-context-protocol/internal/ast/python"
	"repository-context-protocol/internal/models"
)

// Index builder orchestrates the indexing process

// IndexBuilder orchestrates the parsing and indexing of repository files
type IndexBuilder struct {
	rootPath       string
	storage        *HybridStorage
	parserRegistry *ast.ParserRegistry
	stats          IndexStatistics
}

// IndexStatistics tracks indexing progress and results
type IndexStatistics struct {
	FilesProcessed   int
	FunctionsIndexed int
	TypesIndexed     int
	VariablesIndexed int
	ConstantsIndexed int
	CallsIndexed     int
	StartTime        time.Time
	EndTime          time.Time
	Duration         time.Duration
}

// NewIndexBuilder creates a new index builder for the given root path
func NewIndexBuilder(rootPath string) *IndexBuilder {
	return &IndexBuilder{
		rootPath: rootPath,
		stats:    IndexStatistics{},
	}
}

// Initialize sets up the index builder and its components
func (ib *IndexBuilder) Initialize() error {
	// Initialize hybrid storage with .repocontext subdirectory
	repoContextDir := filepath.Join(ib.rootPath, ".repocontext")
	ib.storage = NewHybridStorage(repoContextDir)
	if err := ib.storage.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize parser registry
	ib.initializeParsers()

	return nil
}

// initializeParsers sets up the parser registry with available language parsers
func (ib *IndexBuilder) initializeParsers() {
	// Create parser registry
	ib.parserRegistry = ast.NewParserRegistry()

	// Register Go parser
	goParser := golang.NewGoParser()
	ib.parserRegistry.Register(goParser)

	// Register Python parser
	pythonParser := python.NewPythonParser()
	ib.parserRegistry.Register(pythonParser)

	// Future: Register additional parsers
	// typescriptParser := typescript.NewTypeScriptParser()
	// ib.parserRegistry.Register(typescriptParser)
}

// ProcessFile processes a single file and adds it to the index
func (ib *IndexBuilder) ProcessFile(filePath string) error {
	if ib.storage == nil {
		return fmt.Errorf("index builder not initialized")
	}

	// Validate and clean the file path
	cleanPath, err := ib.validateAndCleanPath(filePath)
	if err != nil {
		return fmt.Errorf("invalid file path %s: %w", filePath, err)
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(cleanPath))

	// Find appropriate parser using registry
	parser, exists := ib.parserRegistry.GetParser(ext)
	if !exists {
		// Skip unsupported file types
		return nil
	}

	// Read file content
	content, err := os.ReadFile(cleanPath) // #nosec G304 - Path validated above
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", cleanPath, err)
	}

	// Parse the file using the registry parser
	fileContext, err := parser.ParseFile(cleanPath, content)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", cleanPath, err)
	}

	// Store in hybrid storage
	if err := ib.storage.StoreFileContext(fileContext); err != nil {
		return fmt.Errorf("failed to store file context: %w", err)
	}

	// Update statistics
	ib.updateStatistics(fileContext)

	return nil
}

// ProcessDirectory processes all supported files in a directory recursively
func (ib *IndexBuilder) ProcessDirectory(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Process the file
		return ib.ProcessFile(path)
	})
}

// BuildIndex builds a complete index of the repository
func (ib *IndexBuilder) BuildIndex() (*IndexStatistics, error) {
	ib.stats.StartTime = time.Now()

	// Phase 1: Parse all files individually
	var fileContexts []models.FileContext
	err := filepath.Walk(ib.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Validate and clean the file path
		cleanPath, validateErr := ib.validateAndCleanPath(path)
		if validateErr != nil {
			// Log but continue with other files
			return nil
		}

		// Get file extension
		ext := strings.ToLower(filepath.Ext(cleanPath))

		// Find appropriate parser using registry
		parser, exists := ib.parserRegistry.GetParser(ext)
		if !exists {
			// Skip unsupported file types
			return nil
		}

		// Read file content
		content, err := os.ReadFile(cleanPath) // #nosec G304 - Path validated above
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", cleanPath, err)
		}

		// Parse the file using the registry parser
		fileContext, err := parser.ParseFile(cleanPath, content)
		if err != nil {
			return fmt.Errorf("failed to parse file %s: %w", cleanPath, err)
		}

		// Add to collection for global analysis
		fileContexts = append(fileContexts, *fileContext)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to process directory: %w", err)
	}

	// Phase 2: Global enrichment - enhance file contexts with cross-file analysis
	enrichment := NewGlobalEnrichment()
	enrichedContexts, err := enrichment.EnrichFileContexts(fileContexts)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich file contexts: %w", err)
	}

	// Phase 3: Store enriched contexts
	for i := range enrichedContexts {
		if err := ib.storage.StoreFileContext(&enrichedContexts[i]); err != nil {
			return nil, fmt.Errorf("failed to store file context: %w", err)
		}

		// Update statistics
		ib.updateStatistics(&enrichedContexts[i])
	}

	ib.stats.EndTime = time.Now()
	ib.stats.Duration = ib.stats.EndTime.Sub(ib.stats.StartTime)

	return &ib.stats, nil
}

// GetStatistics returns current indexing statistics
func (ib *IndexBuilder) GetStatistics() IndexStatistics {
	return ib.stats
}

// updateStatistics updates the internal statistics based on processed file
func (ib *IndexBuilder) updateStatistics(fileContext *models.FileContext) {
	ib.stats.FilesProcessed++
	ib.stats.FunctionsIndexed += len(fileContext.Functions)
	ib.stats.TypesIndexed += len(fileContext.Types)
	ib.stats.VariablesIndexed += len(fileContext.Variables)
	ib.stats.ConstantsIndexed += len(fileContext.Constants)

	// Count calls from all functions
	for i := range fileContext.Functions {
		ib.stats.CallsIndexed += len(fileContext.Functions[i].Calls)
	}
}

// Close properly shuts down the index builder and its components
func (ib *IndexBuilder) Close() error {
	if ib.storage != nil {
		if err := ib.storage.Close(); err != nil {
			return fmt.Errorf("failed to close storage: %w", err)
		}
		ib.storage = nil
	}

	// Clear parser registry
	ib.parserRegistry = nil

	return nil
}

// validateAndCleanPath validates and cleans a file path for security
func (ib *IndexBuilder) validateAndCleanPath(path string) (string, error) {
	// Clean the path to prevent path traversal attacks
	cleanPath := filepath.Clean(path)

	// Get absolute paths for comparison
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	absRootPath, err := filepath.Abs(ib.rootPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute root path: %w", err)
	}

	// Check if the path is within the root directory
	if !strings.HasPrefix(absPath, absRootPath) {
		return "", fmt.Errorf("path is outside root directory")
	}

	return cleanPath, nil
}
