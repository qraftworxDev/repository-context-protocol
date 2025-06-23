package python

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPythonParser_BasicIntegration(t *testing.T) {
	parser := NewPythonParser()

	// Read the test file from existing test data
	testFile := filepath.Join("..", "..", "..", "testdata", "python-simple", "main.py")
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Parse the file
	fileContext, err := parser.ParseFile(testFile, content)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Validate basic file information
	if fileContext.Path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, fileContext.Path)
	}
	if fileContext.Language != "python" {
		t.Errorf("Expected language 'python', got %s", fileContext.Language)
	}

	// Validate that we have reasonable content
	if len(fileContext.Functions) == 0 {
		t.Error("Expected to find functions in main.py")
	}

	// Check for main function (common in Python entry points)
	mainFunc := findFunction(fileContext.Functions, "main")
	if mainFunc != nil {
		t.Logf("Found main function at line %d", mainFunc.StartLine)
	}

	// Validate that imports are parsed
	if len(fileContext.Imports) == 0 {
		t.Log("No imports found - this may be expected for simple test files")
	}

	// Log summary for visibility
	t.Logf("Parsed Python file: %d functions, %d types, %d variables, %d imports",
		len(fileContext.Functions), len(fileContext.Types), len(fileContext.Variables), len(fileContext.Imports))
}

func TestPythonParser_ModelsIntegration(t *testing.T) {
	parser := NewPythonParser()

	// Read the models test file
	testFile := filepath.Join("..", "..", "..", "testdata", "python-simple", "models.py")
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read models file: %v", err)
	}

	// Parse the file
	fileContext, err := parser.ParseFile(testFile, content)
	if err != nil {
		t.Fatalf("Failed to parse models file: %v", err)
	}

	// Validate file metadata
	if fileContext.Path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, fileContext.Path)
	}
	if fileContext.Language != "python" {
		t.Errorf("Expected language 'python', got %s", fileContext.Language)
	}

	// Check for classes (models.py typically contains class definitions)
	if len(fileContext.Types) == 0 {
		t.Error("Expected to find classes in models.py")
	}

	// Validate that we have meaningful exports
	if len(fileContext.Exports) == 0 {
		t.Error("Expected to find exports in models.py")
	}

	// Check that classes have methods
	hasMethodsInClasses := false
	for _, class := range fileContext.Types {
		if len(class.Methods) > 0 {
			hasMethodsInClasses = true
			t.Logf("Class %s has %d methods", class.Name, len(class.Methods))
		}
	}

	if !hasMethodsInClasses {
		t.Error("Expected at least one class to have methods")
	}

	// Log detailed summary
	t.Logf("Parsed models.py: %d classes, %d functions, %d variables, %d exports",
		len(fileContext.Types), len(fileContext.Functions), len(fileContext.Variables), len(fileContext.Exports))
}

func TestPythonParser_MetadataValidation(t *testing.T) {
	parser := NewPythonParser()

	// Simple test code
	code := `#!/usr/bin/env python3
"""Test module for metadata validation."""

def test_function():
    """A simple test function."""
    return "test"

class TestClass:
    """A simple test class."""

    def __init__(self):
        self.value = 42

    def get_value(self):
        return self.value
`

	fileContext, err := parser.ParseFile("test_metadata.py", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse test code: %v", err)
	}

	// Validate metadata is set
	if fileContext.Checksum == "" {
		t.Error("Expected checksum to be set")
	}

	if fileContext.ModTime.IsZero() {
		t.Error("Expected modification time to be set")
	}

	// Validate line numbers are set correctly
	testFunc := findFunction(fileContext.Functions, "test_function")
	if testFunc == nil {
		t.Fatal("Expected to find test_function")
	}

	if testFunc.StartLine == 0 {
		t.Error("Expected function start line to be set")
	}

	if testFunc.EndLine == 0 {
		t.Error("Expected function end line to be set")
	}

	if testFunc.StartLine >= testFunc.EndLine {
		t.Errorf("Expected start line (%d) to be less than end line (%d)",
			testFunc.StartLine, testFunc.EndLine)
	}

	// Validate class line numbers
	testClass := findType(fileContext.Types, "TestClass")
	if testClass == nil {
		t.Fatal("Expected to find TestClass")
	}

	if testClass.StartLine == 0 || testClass.EndLine == 0 {
		t.Error("Expected class line numbers to be set")
	}

	// Validate method line numbers
	if len(testClass.Methods) < 2 {
		t.Fatalf("Expected TestClass to have at least 2 methods, got %d", len(testClass.Methods))
	}

	for _, method := range testClass.Methods {
		if method.StartLine == 0 || method.EndLine == 0 {
			t.Errorf("Method %s should have non-zero line numbers, got start=%d, end=%d",
				method.Name, method.StartLine, method.EndLine)
		}
	}
}

func TestPythonParser_IntegrationErrorHandling(t *testing.T) {
	parser := NewPythonParser()

	// Test with invalid Python syntax
	invalidCode := `def invalid_function(
		# Missing closing parenthesis and proper syntax
		return "invalid"
	`

	_, err := parser.ParseFile("invalid.py", []byte(invalidCode))
	if err == nil {
		t.Error("Expected error for invalid Python code")
	}

	// Test with empty file
	emptyContext, err := parser.ParseFile("empty.py", []byte(""))
	if err != nil {
		t.Errorf("Should handle empty files gracefully, got error: %v", err)
	}

	if emptyContext == nil {
		t.Error("Expected valid context for empty file")
	}

	// Test with only comments
	commentOnlyCode := `# This is just a comment
# Another comment
`
	commentContext, err := parser.ParseFile("comments.py", []byte(commentOnlyCode))
	if err != nil {
		t.Errorf("Should handle comment-only files gracefully, got error: %v", err)
	}

	if commentContext == nil {
		t.Error("Expected valid context for comment-only file")
	}
}
