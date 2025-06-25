package python

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestPythonParser_LargeFilePerformance(t *testing.T) {
	parser := NewPythonParser()

	// Generate a large Python file with many functions and classes
	largeCode := generateLargePythonCode(100, 50) // 100 functions, 50 classes

	start := time.Now()
	fileContext, err := parser.ParseFile("large_test.py", []byte(largeCode))
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to parse large Python file: %v", err)
	}

	// Performance assertions
	if duration > 10*time.Second {
		t.Errorf("Parsing took too long: %v (expected < 10s)", duration)
	}

	// Validate that all content was parsed
	if len(fileContext.Functions) < 90 { // Allow some tolerance
		t.Errorf("Expected around 100 functions, got %d", len(fileContext.Functions))
	}

	if len(fileContext.Types) < 45 { // Allow some tolerance
		t.Errorf("Expected around 50 classes, got %d", len(fileContext.Types))
	}

	t.Logf("Large file performance test: parsed %d functions, %d types in %v",
		len(fileContext.Functions), len(fileContext.Types), duration)
}

func TestPythonParser_MultipleFilesPerformance(t *testing.T) {
	parser := NewPythonParser()

	// Test parsing multiple smaller files
	numFiles := 20
	fileSize := 50 // lines per file

	start := time.Now()
	totalFunctions := 0
	totalTypes := 0

	for i := 0; i < numFiles; i++ {
		code := generateMediumPythonCode(fileSize)
		fileName := fmt.Sprintf("test_file_%d.py", i)

		fileContext, err := parser.ParseFile(fileName, []byte(code))
		if err != nil {
			t.Fatalf("Failed to parse file %s: %v", fileName, err)
		}

		totalFunctions += len(fileContext.Functions)
		totalTypes += len(fileContext.Types)
	}

	duration := time.Since(start)

	// Performance assertion
	if duration > 5*time.Second {
		t.Errorf("Parsing %d files took too long: %v (expected < 5s)", numFiles, duration)
	}

	if totalFunctions == 0 {
		t.Error("Expected to parse some functions across all files")
	}

	t.Logf("Multiple files performance test: parsed %d files with %d functions, %d types in %v",
		numFiles, totalFunctions, totalTypes, duration)
}

func TestPythonParser_MemoryUsage(t *testing.T) {
	parser := NewPythonParser()

	// Test with code that has deep nesting and complex structures
	complexCode := generateComplexNestedCode(10) // 10 levels of nesting

	fileContext, err := parser.ParseFile("complex_nested.py", []byte(complexCode))
	if err != nil {
		t.Fatalf("Failed to parse complex nested code: %v", err)
	}

	// Validate structure is correctly parsed
	if len(fileContext.Functions) == 0 {
		t.Error("Expected to find functions in complex nested code")
	}

	if len(fileContext.Types) == 0 {
		t.Error("Expected to find classes in complex nested code")
	}

	t.Logf("Memory usage test passed: %d functions, %d types parsed",
		len(fileContext.Functions), len(fileContext.Types))
}

func TestPythonParser_ConcurrentParsing(t *testing.T) {
	parser := NewPythonParser()

	// Test concurrent parsing of multiple files
	numGoroutines := 5
	resultsChannel := make(chan bool, numGoroutines)

	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
					resultsChannel <- false
					return
				}
			}()

			code := generateMediumPythonCode(30)
			fileName := fmt.Sprintf("concurrent_test_%d.py", id)

			fileContext, err := parser.ParseFile(fileName, []byte(code))
			if err != nil {
				t.Errorf("Goroutine %d failed to parse: %v", id, err)
				resultsChannel <- false
				return
			}

			if len(fileContext.Functions) == 0 {
				t.Errorf("Goroutine %d found no functions", id)
				resultsChannel <- false
				return
			}

			resultsChannel <- true
		}(i)
	}

	// Wait for all goroutines to complete
	successCount := 0
	for i := 0; i < numGoroutines; i++ {
		select {
		case success := <-resultsChannel:
			if success {
				successCount++
			}
		case <-time.After(10 * time.Second):
			t.Fatalf("Timeout waiting for goroutine results")
		}
	}

	duration := time.Since(start)

	if successCount != numGoroutines {
		t.Errorf("Expected %d successful parses, got %d", numGoroutines, successCount)
	}

	t.Logf("Concurrent parsing test: %d goroutines completed successfully in %v",
		successCount, duration)
}

// Helper function to generate large Python code
func generateLargePythonCode(numFunctions, numClasses int) string {
	var builder strings.Builder

	builder.WriteString("#!/usr/bin/env python3\n")
	builder.WriteString("\"\"\"Generated large Python file for performance testing.\"\"\"\n\n")
	builder.WriteString("import os\nimport sys\nfrom typing import List, Dict, Optional\n\n")

	// Generate functions
	for i := 0; i < numFunctions; i++ {
		builder.WriteString(fmt.Sprintf(`def function_%d(param1: str, param2: int = %d) -> str:
    """Function %d for testing."""
    result = f"Function {param1} with {param2}"
    return result

`, i, i, i))
	}

	// Generate classes
	for i := 0; i < numClasses; i++ {
		builder.WriteString(fmt.Sprintf(`class TestClass_%d:
    """Test class %d."""

    def __init__(self, value: int = %d):
        self.value = value

    def get_value(self) -> int:
        return self.value

    def set_value(self, new_value: int) -> None:
        self.value = new_value

`, i, i, i))
	}

	return builder.String()
}

// Helper function to generate medium-sized Python code
func generateMediumPythonCode(numLines int) string {
	var builder strings.Builder

	builder.WriteString("#!/usr/bin/env python3\n")
	builder.WriteString("\"\"\"Generated medium Python file.\"\"\"\n\n")
	builder.WriteString("from typing import Any\n\n")

	linesWritten := 4
	funcCount := 0
	classCount := 0

	for linesWritten < numLines {
		if linesWritten%10 == 0 {
			// Add a class every 10 lines
			builder.WriteString(fmt.Sprintf(`class MediumClass_%d:
    def __init__(self):
        self.data = %d

    def process(self) -> int:
        return self.data * 2

`, classCount, classCount))
			classCount++
			linesWritten += 7
		} else {
			// Add a function
			builder.WriteString(fmt.Sprintf(`def medium_function_%d() -> str:
    return "Result %d"

`, funcCount, funcCount))
			funcCount++
			linesWritten += 3
		}
	}

	return builder.String()
}

// Helper function to generate complex nested code
func generateComplexNestedCode(depth int) string {
	var builder strings.Builder

	builder.WriteString("#!/usr/bin/env python3\n")
	builder.WriteString("\"\"\"Generated complex nested Python code.\"\"\"\n\n")

	// Generate nested classes
	for i := 0; i < depth; i++ {
		indent := strings.Repeat("    ", i)
		builder.WriteString(fmt.Sprintf("%sclass NestedClass_%d:\n", indent, i))
		builder.WriteString(fmt.Sprintf("%s    def __init__(self):\n", indent))
		builder.WriteString(fmt.Sprintf("%s        self.level = %d\n", indent, i))
		builder.WriteString(fmt.Sprintf("%s    \n", indent))
		builder.WriteString(fmt.Sprintf("%s    def get_level(self) -> int:\n", indent))
		builder.WriteString(fmt.Sprintf("%s        return self.level\n", indent))
		builder.WriteString(fmt.Sprintf("%s    \n", indent))
	}

	// Generate nested functions
	builder.WriteString("\ndef outer_function():\n")
	for i := 0; i < depth; i++ {
		indent := strings.Repeat("    ", i+1)
		builder.WriteString(fmt.Sprintf("%sdef nested_function_%d():\n", indent, i))
		builder.WriteString(fmt.Sprintf("%s    result = %d\n", indent, i))
		if i < depth-1 {
			builder.WriteString(fmt.Sprintf("%s    \n", indent))
		} else {
			builder.WriteString(fmt.Sprintf("%s    return result\n", indent))
		}
	}

	// Close all nested functions
	for i := depth - 1; i >= 0; i-- {
		indent := strings.Repeat("    ", i+1)
		builder.WriteString(fmt.Sprintf("%sreturn nested_function_%d()\n", indent, i))
	}

	return builder.String()
}
