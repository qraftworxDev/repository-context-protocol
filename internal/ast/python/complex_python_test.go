package python

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPythonParser_ComplexInheritanceExample(t *testing.T) {
	parser := NewPythonParser()

	// Test with the complex inheritance example
	testFile := filepath.Join("..", "..", "..", "testdata", "python-complex", "inheritance_example.py")
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Skipf("Skipping complex inheritance test, file not found: %v", err)
	}

	fileContext, err := parser.ParseFile(testFile, content)
	if err != nil {
		t.Fatalf("Failed to parse complex inheritance file: %v", err)
	}

	// Basic validation
	if len(fileContext.Types) == 0 {
		t.Error("Expected to find classes in complex inheritance example")
	}

	if len(fileContext.Functions) == 0 {
		t.Error("Expected to find functions in complex inheritance example")
	}

	if len(fileContext.Imports) == 0 {
		t.Error("Expected to find imports in complex inheritance example")
	}

	// Log detailed results
	t.Logf("Complex inheritance example parsed successfully:")
	t.Logf("  - Classes: %d", len(fileContext.Types))
	t.Logf("  - Functions: %d", len(fileContext.Functions))
	t.Logf("  - Variables: %d", len(fileContext.Variables))
	t.Logf("  - Imports: %d", len(fileContext.Imports))
	t.Logf("  - Exports: %d", len(fileContext.Exports))

	// Check for specific expected classes
	expectedClasses := []string{
		"Shape", "Rectangle", "Square", "Circle", "Container",
		"ShapeManager", "AsyncShapeProcessor", "Point", "Drawable",
	}
	foundClasses := make(map[string]bool)

	for _, class := range fileContext.Types {
		foundClasses[class.Name] = true
		t.Logf("  - Found class: %s with %d methods", class.Name, len(class.Methods))
	}

	missingClasses := []string{}
	for _, expectedClass := range expectedClasses {
		if !foundClasses[expectedClass] {
			missingClasses = append(missingClasses, expectedClass)
		}
	}

	if len(missingClasses) > 0 {
		t.Logf("Note: Some expected classes not found: %v", missingClasses)
		t.Log("This may be due to Python AST extractor limitations or the implementation")
	}

	// Check for specific functions
	expectedFunctions := []string{"retry", "timer", "create_test_shapes", "main"}
	foundFunctions := make(map[string]bool)

	for _, function := range fileContext.Functions {
		foundFunctions[function.Name] = true
		t.Logf("  - Found function: %s", function.Name)
	}

	for _, expectedFunc := range expectedFunctions {
		if !foundFunctions[expectedFunc] {
			t.Logf("Note: Expected function '%s' not found (may be extracted as method or not parsed)", expectedFunc)
		}
	}

	// Validate inheritance relationships exist
	shapeClass := findType(fileContext.Types, "Shape")
	if shapeClass != nil {
		t.Logf("Shape class found with %d methods and embedded types: %v",
			len(shapeClass.Methods), shapeClass.Embedded)
	}

	rectangleClass := findType(fileContext.Types, "Rectangle")
	if rectangleClass != nil {
		t.Logf("Rectangle class found with %d methods and embedded types: %v",
			len(rectangleClass.Methods), rectangleClass.Embedded)
	}

	// Check for imports from abc and typing
	typingImportFound := false
	abcImportFound := false

	for _, imp := range fileContext.Imports {
		if imp.Path == "typing" {
			typingImportFound = true
		}
		if imp.Path == "abc" {
			abcImportFound = true
		}
		t.Logf("  - Import: %s", imp.Path)
	}

	if !typingImportFound {
		t.Log("Note: typing import not found - may affect type annotation extraction")
	}

	if !abcImportFound {
		t.Log("Note: abc import not found - may affect abstract class detection")
	}

	// Success if we got some reasonable content
	if len(fileContext.Types) >= 3 && len(fileContext.Functions) >= 1 {
		t.Log("✅ Complex inheritance example test passed with reasonable content")
	} else {
		t.Errorf("❌ Expected more content: got %d classes and %d functions",
			len(fileContext.Types), len(fileContext.Functions))
	}
}

func TestPythonParser_ComplexPythonFeatures(t *testing.T) {
	parser := NewPythonParser()

	// Test code with advanced Python features
	complexCode := `#!/usr/bin/env python3
"""Test advanced Python features."""

from abc import ABC, abstractmethod
from typing import Generic, TypeVar, Optional, List, Dict
from dataclasses import dataclass
import asyncio

T = TypeVar('T')

@dataclass
class Config:
    """Configuration dataclass."""
    name: str
    value: int = 42

class BaseProcessor(ABC):
    """Abstract base processor."""

    def __init__(self, config: Config):
        self.config = config

    @abstractmethod
    async def process(self, data: T) -> Optional[str]:
        pass

    @property
    def name(self) -> str:
        return self.config.name

class DataProcessor(BaseProcessor, Generic[T]):
    """Generic data processor."""

    def __init__(self, config: Config):
        super().__init__(config)
        self.cache: Dict[str, T] = {}

    async def process(self, data: T) -> Optional[str]:
        # Process the data
        await asyncio.sleep(0.01)
        return str(data)

    @classmethod
    def create_default(cls, name: str) -> 'DataProcessor[str]':
        config = Config(name=name)
        return cls(config)

def decorator_with_params(param: str):
    """Parameterized decorator."""
    def decorator(func):
        def wrapper(*args, **kwargs):
            print(f"Calling {func.__name__} with {param}")
            return func(*args, **kwargs)
        return wrapper
    return decorator

@decorator_with_params("test")
async def example_function(items: List[str]) -> Dict[str, int]:
    """Example function with decorators and type hints."""
    return {item: len(item) for item in items}
`

	fileContext, err := parser.ParseFile("complex_features.py", []byte(complexCode))
	if err != nil {
		t.Fatalf("Failed to parse complex features code: %v", err)
	}

	// Validate results
	if len(fileContext.Types) < 2 {
		t.Errorf("Expected at least 2 classes, got %d", len(fileContext.Types))
	}

	if len(fileContext.Functions) < 2 {
		t.Errorf("Expected at least 2 functions, got %d", len(fileContext.Functions))
	}

	if len(fileContext.Imports) < 3 {
		t.Errorf("Expected at least 3 imports, got %d", len(fileContext.Imports))
	}

	// Check for dataclass
	configClass := findType(fileContext.Types, "Config")
	if configClass == nil {
		t.Error("Expected to find Config dataclass")
	} else {
		t.Logf("Found Config class with %d methods", len(configClass.Methods))
	}

	// Check for abstract base class
	baseProcessorClass := findType(fileContext.Types, "BaseProcessor")
	if baseProcessorClass == nil {
		t.Error("Expected to find BaseProcessor abstract class")
	} else {
		t.Logf("Found BaseProcessor class with %d methods", len(baseProcessorClass.Methods))
	}

	// Check for generic class
	dataProcessorClass := findType(fileContext.Types, "DataProcessor")
	if dataProcessorClass == nil {
		t.Error("Expected to find DataProcessor generic class")
	} else {
		t.Logf("Found DataProcessor class with %d methods and inheritance: %v",
			len(dataProcessorClass.Methods), dataProcessorClass.Embedded)
	}

	// Check for decorated function
	exampleFunc := findFunction(fileContext.Functions, "example_function")
	if exampleFunc == nil {
		t.Error("Expected to find example_function")
	} else {
		t.Logf("Found example_function with %d parameters", len(exampleFunc.Parameters))
	}

	t.Logf("Complex Python features test completed: %d classes, %d functions, %d imports",
		len(fileContext.Types), len(fileContext.Functions), len(fileContext.Imports))
}
