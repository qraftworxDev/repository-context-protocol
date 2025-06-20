package index

import (
	"os"
	"path/filepath"
	"testing"

	"repository-context-protocol/internal/models"
)

func TestIndexBuilder_ParserRegistry(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_registry_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	builder := NewIndexBuilder(tempDir)

	// Initialize index builder
	err = builder.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize index builder: %v", err)
	}
	defer builder.Close()

	// Verify parser registry is initialized
	if builder.parserRegistry == nil {
		t.Error("Expected parser registry to be initialized")
	}

	// Verify Go parser is registered
	goParser, exists := builder.parserRegistry.GetParser(".go")
	if !exists {
		t.Error("Expected Go parser to be registered for .go extension")
	}
	if goParser == nil {
		t.Error("Expected Go parser to not be nil")
	}

	// Verify parser implements the interface correctly
	if goParser.GetLanguageName() != "go" {
		t.Errorf("Expected language name 'go', got '%s'", goParser.GetLanguageName())
	}

	extensions := goParser.GetSupportedExtensions()
	if len(extensions) != 1 || extensions[0] != ".go" {
		t.Errorf("Expected extensions ['.go'], got %v", extensions)
	}
}

func TestIndexBuilder_MultipleParserSupport(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_multiple_parser_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	builder := NewIndexBuilder(tempDir)

	// Initialize index builder
	err = builder.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize index builder: %v", err)
	}
	defer builder.Close()

	// Test that we can register additional parsers
	// (This tests the extensibility even though we don't have other parsers yet)

	// Verify Python parser is now registered (we added Python support)
	pythonParser, exists := builder.parserRegistry.GetParser(".py")
	if !exists {
		t.Error("Expected Python parser to be registered")
	}
	if pythonParser == nil {
		t.Error("Expected Python parser to not be nil")
	}
	if pythonParser.GetLanguageName() != "python" {
		t.Errorf("Expected Python parser language name 'python', got '%s'", pythonParser.GetLanguageName())
	}

	// TypeScript is still not supported
	_, exists = builder.parserRegistry.GetParser(".ts")
	if exists {
		t.Error("Expected TypeScript parser to not be registered")
	}

	// Verify we can still get the Go parser
	goParser, exists := builder.parserRegistry.GetParser(".go")
	if !exists {
		t.Error("Expected Go parser to be registered")
	}
	if goParser == nil {
		t.Error("Expected Go parser to not be nil")
	}
}

func TestIndexBuilder_ProcessFileWithRegistry(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_registry_process_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	builder := NewIndexBuilder(tempDir)
	err = builder.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize index builder: %v", err)
	}
	defer builder.Close()

	// Test that the registry correctly identifies and processes Go files
	testFile := filepath.Join(tempDir, "registry_test.go")
	testContent := `package main

func registryTestFunction() {
	println("registry test")
}
`
	err = os.WriteFile(testFile, []byte(testContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Verify the parser registry is used for processing
	parser, exists := builder.parserRegistry.GetParser(".go")
	if !exists {
		t.Fatal("Expected Go parser to be registered")
	}
	if parser.GetLanguageName() != "go" {
		t.Errorf("Expected Go parser, got %s", parser.GetLanguageName())
	}

	// Process the file using the registry
	err = builder.ProcessFile(testFile)
	if err != nil {
		t.Fatalf("Failed to process file: %v", err)
	}

	// Verify the file was parsed and indexed correctly
	results, err := builder.storage.QueryByName("registryTestFunction")
	if err != nil {
		t.Fatalf("Failed to query for registryTestFunction: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for registryTestFunction, got %d", len(results))
	}
}

func TestIndexBuilder_UnsupportedFileTypes(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_unsupported_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	builder := NewIndexBuilder(tempDir)
	err = builder.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize index builder: %v", err)
	}
	defer builder.Close()

	// Create unsupported file types (removed test.py since Python is now supported)
	unsupportedFiles := []string{
		"test.ts",   // TypeScript
		"test.js",   // JavaScript
		"test.txt",  // Text
		"README.md", // Markdown
	}

	for _, filename := range unsupportedFiles {
		filePath := filepath.Join(tempDir, filename)
		err = os.WriteFile(filePath, []byte("# test content"), 0600)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}

		// Process the file - should not error, just skip
		err = builder.ProcessFile(filePath)
		if err != nil {
			t.Errorf("Expected ProcessFile to skip unsupported file %s without error, got: %v", filename, err)
		}
	}

	// Verify no files were processed (statistics should be 0)
	stats := builder.GetStatistics()
	if stats.FilesProcessed != 0 {
		t.Errorf("Expected 0 files processed for unsupported types, got %d", stats.FilesProcessed)
	}
}

func TestIndexBuilder_RegisterAdditionalParser(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_additional_parser_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	builder := NewIndexBuilder(tempDir)
	err = builder.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize index builder: %v", err)
	}
	defer builder.Close()

	// Test that we can register additional parsers programmatically
	// (This demonstrates the extensibility)

	// Create a mock parser for testing
	mockParser := &mockLanguageParser{
		language:   "mock",
		extensions: []string{".mock"},
	}

	// Register the mock parser
	builder.parserRegistry.Register(mockParser)

	// Verify the mock parser is registered
	registeredParser, exists := builder.parserRegistry.GetParser(".mock")
	if !exists {
		t.Error("Expected mock parser to be registered for .mock extension")
	}
	if registeredParser != mockParser {
		t.Error("Expected registered parser to be the same instance as the mock parser")
	}

	// Verify the mock parser properties
	if registeredParser.GetLanguageName() != "mock" {
		t.Errorf("Expected language name 'mock', got '%s'", registeredParser.GetLanguageName())
	}
}

// mockLanguageParser is a simple mock implementation for testing
type mockLanguageParser struct {
	language   string
	extensions []string
}

func (m *mockLanguageParser) ParseFile(path string, content []byte) (*models.FileContext, error) {
	// Return a minimal file context for testing
	return &models.FileContext{
		Path:     path,
		Language: m.language,
	}, nil
}

func (m *mockLanguageParser) GetSupportedExtensions() []string {
	return m.extensions
}

func (m *mockLanguageParser) GetLanguageName() string {
	return m.language
}
