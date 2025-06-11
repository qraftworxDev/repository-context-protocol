package ast

import (
	"testing"

	"repository-context-protocol/internal/models"
)

func TestParserRegistry_Register(t *testing.T) {
	registry := NewParserRegistry()

	// Create a mock parser for testing
	mockParser := &mockLanguageParser{
		extensions: []string{".go"},
		language:   "go",
	}

	registry.Register(mockParser)

	parser, exists := registry.GetParser(".go")
	if !exists {
		t.Error("Expected parser to be registered for .go extension")
	}
	if parser != mockParser {
		t.Error("Expected registered parser to be returned")
	}
}

func TestParserRegistry_GetParser_NotFound(t *testing.T) {
	registry := NewParserRegistry()

	_, exists := registry.GetParser(".unknown")
	if exists {
		t.Error("Expected no parser for unknown extension")
	}
}

func TestParserRegistry_MultipleExtensions(t *testing.T) {
	registry := NewParserRegistry()

	mockParser := &mockLanguageParser{
		extensions: []string{".go", ".mod"},
		language:   "go",
	}

	registry.Register(mockParser)

	// Test both extensions
	parser1, exists1 := registry.GetParser(".go")
	parser2, exists2 := registry.GetParser(".mod")

	if !exists1 || !exists2 {
		t.Error("Expected parser to be registered for both extensions")
	}
	if parser1 != mockParser || parser2 != mockParser {
		t.Error("Expected same parser for both extensions")
	}
}

// Mock parser for testing
type mockLanguageParser struct {
	extensions []string
	language   string
}

func (m *mockLanguageParser) ParseFile(path string, content []byte) (*models.FileContext, error) {
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
