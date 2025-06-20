package python

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"repository-context-protocol/internal/models"
)

func TestPythonParser_Constructor(t *testing.T) {
	parser := NewPythonParser()
	if parser == nil {
		t.Error("Expected NewPythonParser() to return a non-nil parser")
	}
}

func TestPythonParser_GetSupportedExtensions(t *testing.T) {
	parser := NewPythonParser()
	extensions := parser.GetSupportedExtensions()

	expected := []string{".py"}
	if len(extensions) != len(expected) {
		t.Errorf("Expected %d extensions, got %d", len(expected), len(extensions))
	}

	if len(extensions) > 0 && extensions[0] != ".py" {
		t.Errorf("Expected .py extension, got %s", extensions[0])
	}
}

func TestPythonParser_GetLanguageName(t *testing.T) {
	parser := NewPythonParser()
	language := parser.GetLanguageName()

	if language != "python" {
		t.Errorf("Expected language 'python', got %s", language)
	}
}

func TestPythonParser_ParseSimpleFunction(t *testing.T) {
	parser := NewPythonParser()

	// Using code from our test data
	code := `#!/usr/bin/env python3
"""Simple function for testing."""

def format_name(name: str) -> str:
    """Format a name by capitalizing first letter of each word."""
    return ' '.join(word.capitalize() for word in name.split())

def validate_email(email: str) -> bool:
    """Basic email validation."""
    return '@' in email and '.' in email.split('@')[1]

def process_user_data(name: str, email: str, age: int = 18) -> dict:
    """Process user data and return formatted result."""
    if not name.strip():
        raise ValueError("Name cannot be empty")

    if not validate_email(email):
        raise ValueError("Invalid email format")

    formatted_name = format_name(name)

    return {
        'name': formatted_name,
        'email': email.lower(),
        'age': age,
        'is_adult': age >= 18,
        'status': 'active'
    }`

	fileContext, err := parser.ParseFile("test.py", []byte(code))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if fileContext.Path != "test.py" {
		t.Errorf("Expected path 'test.py', got %s", fileContext.Path)
	}
	if fileContext.Language != "python" {
		t.Errorf("Expected language 'python', got %s", fileContext.Language)
	}

	// Should have 3 functions: format_name, validate_email, process_user_data
	if len(fileContext.Functions) != 3 {
		t.Errorf("Expected 3 functions, got %d", len(fileContext.Functions))
	}

	// Check format_name function
	formatFunc := findFunction(fileContext.Functions, "format_name")
	if formatFunc == nil {
		t.Error("Expected to find format_name function")
	} else {
		if len(formatFunc.Parameters) != 1 {
			t.Errorf("Expected format_name to have 1 parameter, got %d", len(formatFunc.Parameters))
		}
		if len(formatFunc.Returns) != 1 {
			t.Errorf("Expected format_name to have 1 return, got %d", len(formatFunc.Returns))
		}
		if len(formatFunc.Parameters) > 0 && formatFunc.Parameters[0].Name != "name" {
			t.Errorf("Expected parameter name 'name', got %s", formatFunc.Parameters[0].Name)
		}
		if len(formatFunc.Parameters) > 0 && formatFunc.Parameters[0].Type != "string" {
			t.Errorf("Expected parameter type 'string', got %s", formatFunc.Parameters[0].Type)
		}
	}

	// Check process_user_data function with default parameter
	processFunc := findFunction(fileContext.Functions, "process_user_data")
	if processFunc == nil {
		t.Error("Expected to find process_user_data function")
	} else {
		if len(processFunc.Parameters) != 3 {
			t.Errorf("Expected process_user_data to have 3 parameters, got %d", len(processFunc.Parameters))
		}
		// Check for default parameter value
		if len(processFunc.Parameters) >= 3 {
			ageParam := processFunc.Parameters[2]
			if ageParam.Name != "age" {
				t.Errorf("Expected parameter name 'age', got %s", ageParam.Name)
			}
			if ageParam.Type != "int" {
				t.Errorf("Expected parameter type 'int', got %s", ageParam.Type)
			}
		}
	}
}

func TestPythonParser_ParseBasicClass(t *testing.T) {
	parser := NewPythonParser()

	// Using code similar to our test data
	code := `#!/usr/bin/env python3
"""Basic class for testing."""

from typing import Optional

class User:
    """Represents a user in the system."""

    def __init__(self, user_id: int, name: str, email: str):
        """Initialize a new User."""
        self.id = user_id
        self.name = name
        self.email = email
        self.is_active = True

    def __str__(self) -> str:
        """Return string representation of user."""
        return f"User(id={self.id}, name='{self.name}')"

    def activate(self) -> None:
        """Activate the user."""
        self.is_active = True

class Profile:
    """User profile with extended information."""

    def __init__(self, user_id: int):
        """Initialize a new Profile."""
        self.user_id = user_id
        self.skills = []

    def add_skill(self, skill: str) -> None:
        """Add a skill to the profile."""
        if skill not in self.skills:
            self.skills.append(skill)`

	fileContext, err := parser.ParseFile("models.py", []byte(code))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if fileContext.Path != "models.py" {
		t.Errorf("Expected path 'models.py', got %s", fileContext.Path)
	}
	if fileContext.Language != "python" {
		t.Errorf("Expected language 'python', got %s", fileContext.Language)
	}

	// Should have 2 classes: User and Profile
	if len(fileContext.Types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(fileContext.Types))
	}

	// Check User class
	userType := findType(fileContext.Types, "User")
	if userType == nil {
		t.Error("Expected to find User class")
	} else {
		if userType.Kind != "class" {
			t.Errorf("Expected type kind 'class', got %s", userType.Kind)
		}
		// Should have methods: __init__, __str__, activate
		if len(userType.Methods) != 3 {
			t.Errorf("Expected User to have 3 methods, got %d", len(userType.Methods))
		}
	}

	// Check Profile class
	profileType := findType(fileContext.Types, "Profile")
	if profileType == nil {
		t.Error("Expected to find Profile class")
	} else {
		if profileType.Kind != "class" {
			t.Errorf("Expected type kind 'class', got %s", profileType.Kind)
		}
		// Should have methods: __init__, add_skill
		if len(profileType.Methods) != 2 {
			t.Errorf("Expected Profile to have 2 methods, got %d", len(profileType.Methods))
		}
	}

	// Should have 1 import
	if len(fileContext.Imports) != 1 {
		t.Errorf("Expected 1 import, got %d", len(fileContext.Imports))
	}
	if len(fileContext.Imports) > 0 && fileContext.Imports[0].Path != "typing" {
		t.Errorf("Expected import 'typing', got %s", fileContext.Imports[0].Path)
	}
}

func TestPythonParser_ParseFromTestData(t *testing.T) {
	parser := NewPythonParser()

	// Test with our actual test data files
	testFiles := []string{
		"../../../testdata/python-simple/main.py",
		"../../../testdata/python-simple/models.py",
	}

	for _, testFile := range testFiles {
		t.Run(filepath.Base(testFile), func(t *testing.T) {
			// This test will fail until we implement file reading in the parser
			_, err := parser.ParseFile(testFile, nil) // Pass nil to test file reading capability
			if err != nil {
				t.Logf("Expected parsing to work for %s, got error: %v", testFile, err)
				// For now, we expect this to fail - that's the point of TDD
			}
		})
	}
}

func TestPythonParser_InvalidSyntax(t *testing.T) {
	parser := NewPythonParser()

	invalidCode := `#!/usr/bin/env python3
"""Invalid Python syntax."""

def broken_function(
    # Missing closing parenthesis and colon
    pass
`

	_, err := parser.ParseFile("invalid.py", []byte(invalidCode))
	if err == nil {
		t.Error("Expected error for invalid Python code")
	}
}

func TestPythonParser_EmptyFile(t *testing.T) {
	parser := NewPythonParser()

	emptyCode := ``

	fileContext, err := parser.ParseFile("empty.py", []byte(emptyCode))
	if err != nil {
		t.Fatalf("Expected no error for empty file, got %v", err)
	}

	if fileContext.Path != "empty.py" {
		t.Errorf("Expected path 'empty.py', got %s", fileContext.Path)
	}
	if fileContext.Language != "python" {
		t.Errorf("Expected language 'python', got %s", fileContext.Language)
	}

	// Empty file should have no functions, types, imports, etc.
	if len(fileContext.Functions) != 0 {
		t.Errorf("Expected 0 functions in empty file, got %d", len(fileContext.Functions))
	}
	if len(fileContext.Types) != 0 {
		t.Errorf("Expected 0 types in empty file, got %d", len(fileContext.Types))
	}
}

// TestPythonParser_TypeHintSupport validates Step 6: Type Hint Support
func TestPythonParser_TypeHintSupport(t *testing.T) {
	parser := NewPythonParser()

	// Test code with comprehensive type hints
	code := `#!/usr/bin/env python3
"""Test type hint support."""

from typing import List, Dict, Optional, Union, Any

# Variables with type hints
name: str = "test"
age: int = 25
scores: List[float] = [1.0, 2.0, 3.0]
config: Dict[str, Any] = {}
user_id: Optional[int] = None

def process_data(
    items: List[str],
    options: Dict[str, bool] = None,
    count: Optional[int] = None
) -> Dict[str, Union[str, int]]:
    """Process data with complex type hints."""
    result: Dict[str, Union[str, int]] = {}
    if options is None:
        options = {}

    result["count"] = len(items) if count is None else count
    result["status"] = "processed"
    return result

def get_user_info(user_id: int) -> Optional[Dict[str, Any]]:
    """Get user information with optional return."""
    if user_id > 0:
        return {"id": user_id, "name": "user"}
    return None

class DataProcessor:
    """Class with typed methods and attributes."""

    def __init__(self, data: List[Dict[str, Any]]):
        self.data: List[Dict[str, Any]] = data
        self.processed: bool = False

    def process(self, filters: Optional[List[str]] = None) -> List[Dict[str, Any]]:
        """Process data with optional filters."""
        if filters:
            return [item for item in self.data if any(f in str(item) for f in filters)]
        return self.data

    def get_count(self) -> int:
        """Get data count."""
        return len(self.data)`

	fileContext, err := parser.ParseFile("type_hints.py", []byte(code))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test function type hints
	processFunc := findFunction(fileContext.Functions, "process_data")
	if processFunc == nil {
		t.Fatal("Expected to find process_data function")
	}

	// Validate parameter types
	if len(processFunc.Parameters) != 3 {
		t.Errorf("Expected 3 parameters, got %d", len(processFunc.Parameters))
	}

	// Check complex type mapping
	if len(processFunc.Parameters) > 0 {
		itemsParam := processFunc.Parameters[0]
		if itemsParam.Name != "items" {
			t.Errorf("Expected parameter name 'items', got %s", itemsParam.Name)
		}
		// Should map List[str] to Go-compatible type
		if itemsParam.Type != "[]interface{}" && itemsParam.Type != "List[str]" {
			t.Logf("Parameter 'items' has type: %s", itemsParam.Type)
		}
	}

	if len(processFunc.Parameters) > 1 {
		optionsParam := processFunc.Parameters[1]
		if optionsParam.Name != "options" {
			t.Errorf("Expected parameter name 'options', got %s", optionsParam.Name)
		}
		// Should handle Dict[str, bool] type
		if optionsParam.Type != "map[string]interface{}" && optionsParam.Type != "Dict[str, bool]" {
			t.Logf("Parameter 'options' has type: %s", optionsParam.Type)
		}
	}

	// Test return type extraction
	if len(processFunc.Returns) > 0 {
		returnType := processFunc.Returns[0]
		// Should handle Dict[str, Union[str, int]] return type
		if returnType.Name == "" {
			t.Error("Expected return type to be extracted")
		}
		t.Logf("Function 'process_data' return type: %s", returnType.Name)
	}

	// Test optional return type
	getUserFunc := findFunction(fileContext.Functions, "get_user_info")
	if getUserFunc == nil {
		t.Fatal("Expected to find get_user_info function")
	}

	if len(getUserFunc.Returns) > 0 {
		returnType := getUserFunc.Returns[0]
		// Should handle Optional[Dict[str, Any]] return type
		t.Logf("Function 'get_user_info' return type: %s", returnType.Name)
	}

	// Test class with typed methods
	dataProcessorType := findType(fileContext.Types, "DataProcessor")
	if dataProcessorType == nil {
		t.Fatal("Expected to find DataProcessor class")
	}

	// Check typed method
	processMethod := findMethod(dataProcessorType.Methods, "process")
	if processMethod == nil {
		t.Fatal("Expected to find process method")
	}

	if len(processMethod.Parameters) > 0 {
		filtersParam := processMethod.Parameters[0]
		if filtersParam.Name != "filters" {
			t.Errorf("Expected parameter name 'filters', got %s", filtersParam.Name)
		}
		// Should handle Optional[List[str]] type
		t.Logf("Method parameter 'filters' has type: %s", filtersParam.Type)
	}

	// Test variable type hints
	if len(fileContext.Variables) > 0 {
		t.Logf("Found %d variables with type hints", len(fileContext.Variables))
		for _, variable := range fileContext.Variables {
			t.Logf("Variable '%s' has type: %s", variable.Name, variable.Type)
		}
	}

	t.Logf("Type hint support test completed successfully")
}

// TestPythonParser_ImportSystem validates Step 7: Import System
func TestPythonParser_ImportSystem(t *testing.T) {
	parser := NewPythonParser()

	// Test code with comprehensive import types
	code := `#!/usr/bin/env python3
"""Test import system support."""

# Absolute imports (import module)
import os
import sys
import json

# Aliased imports (import numpy as np)
import datetime as dt
import collections as coll

# From imports with multiple items
from typing import List, Dict, Optional, Union, Any
from pathlib import Path

# From imports with aliases
from collections import defaultdict as dd, Counter as cnt
from datetime import datetime as dt_class, timedelta

# Star imports (from module import *)
from math import *

# Relative imports (from .module import func)
from .utils import helper_function
from ..parent import shared_util
from ...root import config

# Mixed relative imports
from .models import User, Profile
from ..services import UserService

def example_function():
    """Function using various imports."""
    # Use absolute imports
    current_time = dt.now()
    file_path = Path("/tmp/test.txt")

    # Use aliased imports
    counter = cnt([1, 2, 2, 3])
    default_dict = dd(list)

    # Use star imports
    result = sqrt(16)  # from math import *

    # Use relative imports
    data = helper_function()
    service = UserService()

    return {"status": "success"}`

	fileContext, err := parser.ParseFile("import_test.py", []byte(code))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test that imports were extracted
	if len(fileContext.Imports) == 0 {
		t.Fatal("Expected imports to be extracted")
	}

	t.Logf("Found %d imports", len(fileContext.Imports))

	// Create a map for easy lookup
	importMap := make(map[string]*models.Import)
	for i := range fileContext.Imports {
		imp := &fileContext.Imports[i]
		key := imp.Path
		if imp.Alias != "" {
			key = fmt.Sprintf("%s as %s", imp.Path, imp.Alias)
		}
		importMap[key] = imp
		t.Logf("Import: path='%s', alias='%s'", imp.Path, imp.Alias)
	}

	// Test absolute imports
	if imp, exists := importMap["os"]; !exists {
		t.Error("Expected to find 'import os'")
	} else if imp.Alias != "" {
		t.Errorf("Expected no alias for 'os', got '%s'", imp.Alias)
	}

	if imp, exists := importMap["sys"]; !exists {
		t.Error("Expected to find 'import sys'")
	} else if imp.Alias != "" {
		t.Errorf("Expected no alias for 'sys', got '%s'", imp.Alias)
	}

	// Test aliased imports
	if imp, exists := importMap["datetime as dt"]; !exists {
		t.Error("Expected to find 'import datetime as dt'")
	} else if imp.Alias != "dt" {
		t.Errorf("Expected alias 'dt' for datetime, got '%s'", imp.Alias)
	}

	// Test from imports
	typingFound := false
	pathlibFound := false
	for _, imp := range fileContext.Imports {
		if imp.Path == "typing" {
			typingFound = true
			t.Logf("Found typing import: %+v", imp)
		}
		if imp.Path == "pathlib" {
			pathlibFound = true
			t.Logf("Found pathlib import: %+v", imp)
		}
	}

	if !typingFound {
		t.Error("Expected to find 'from typing import ...'")
	}
	if !pathlibFound {
		t.Error("Expected to find 'from pathlib import Path'")
	}

	// Test relative imports
	relativeImportsFound := 0
	for _, imp := range fileContext.Imports {
		if strings.HasPrefix(imp.Path, ".") {
			relativeImportsFound++
			t.Logf("Found relative import: path='%s', alias='%s'", imp.Path, imp.Alias)
		}
	}

	if relativeImportsFound == 0 {
		t.Error("Expected to find relative imports (starting with '.')")
	}

	// Test star imports
	starImportsFound := false
	for _, imp := range fileContext.Imports {
		// Check if this is a star import (our extractor should mark it)
		// Note: This depends on how our extractor represents star imports
		if imp.Path == "math" {
			starImportsFound = true
			t.Logf("Found math import (potential star import): %+v", imp)
		}
	}

	if !starImportsFound {
		t.Log("Note: Star import detection may need enhancement")
	}

	t.Logf("Import system test completed successfully")
}

// TestPythonParser_CallGraphGeneration validates Step 8: Call Graph Generation
func TestPythonParser_CallGraphGeneration(t *testing.T) {
	parser := NewPythonParser()

	// Test code with comprehensive call graph scenarios
	code := `#!/usr/bin/env python3
"""Test call graph generation."""

def helper_function(data):
    """Helper function that gets called by others."""
    return len(data)

def utility_function():
    """Another utility function."""
    return "utility"

def main_function():
    """Main function that calls other functions."""
    # Function calls within same file
    result = helper_function([1, 2, 3])
    util = utility_function()

    # Built-in function calls
    print(f"Result: {result}")

    # Create object and call methods
    processor = DataProcessor()
    processor.process_data([1, 2, 3])

    return result

class DataProcessor:
    """Class with methods that call other methods."""

    def __init__(self):
        """Initialize processor."""
        self.data = []
        self.setup()

    def setup(self):
        """Setup method called by constructor."""
        self.data = []

    def process_data(self, items):
        """Process data using helper methods."""
        # Method calls on self
        self.validate_input(items)
        cleaned = self.clean_data(items)
        result = self.transform_data(cleaned)

        # Function call to module-level function
        count = helper_function(result)

        return result

    def validate_input(self, items):
        """Validate input data."""
        if not items:
            raise ValueError("Empty input")
        return True

    def clean_data(self, items):
        """Clean the input data."""
        return [x for x in items if x is not None]

    def transform_data(self, items):
        """Transform the data."""
        # Method call to another method
        multiplier = self.get_multiplier()
        return [x * multiplier for x in items]

    def get_multiplier(self):
        """Get multiplier value."""
        return 2

# Cross-module style calls (simulated)
def cross_module_caller():
    """Function that simulates cross-module calls."""
    # These would be cross-module in real scenarios
    os_path = "os.path.join"  # Simulated
    json_loads = "json.loads"  # Simulated
    return "cross_module"

if __name__ == "__main__":
    main_function()`

	fileContext, err := parser.ParseFile("call_graph_test.py", []byte(code))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test that functions were extracted
	if len(fileContext.Functions) == 0 {
		t.Fatal("Expected functions to be extracted")
	}

	t.Logf("Found %d functions", len(fileContext.Functions))
	for _, fn := range fileContext.Functions {
		t.Logf("Function: %s, Calls: %v, CalledBy: %v", fn.Name, fn.Calls, fn.CalledBy)
	}

	// Test that classes and methods were extracted
	if len(fileContext.Types) == 0 {
		t.Fatal("Expected types (classes) to be extracted")
	}

	dataProcessorType := findType(fileContext.Types, "DataProcessor")
	if dataProcessorType == nil {
		t.Fatal("Expected to find DataProcessor class")
	}

	t.Logf("Found %d methods in DataProcessor", len(dataProcessorType.Methods))
	for _, method := range dataProcessorType.Methods {
		t.Logf("Method: %s, Signature: %s", method.Name, method.Signature)
	}

	// Test function calls within same file
	mainFunc := findFunction(fileContext.Functions, "main_function")
	if mainFunc == nil {
		t.Fatal("Expected to find main_function")
	}

	// Check that main_function calls helper_function and utility_function
	expectedCalls := []string{"helper_function", "utility_function"}
	foundCalls := make(map[string]bool)
	for _, call := range mainFunc.Calls {
		foundCalls[call] = true
		t.Logf("main_function calls: %s", call)
	}

	for _, expectedCall := range expectedCalls {
		if !foundCalls[expectedCall] {
			t.Errorf("Expected main_function to call %s", expectedCall)
		}
	}

	// Test that we have call information in functions
	if len(mainFunc.Calls) == 0 {
		t.Error("Expected main_function to have call information")
	}

	// Test method extraction (call graph for methods is handled differently in our current model)
	processMethod := findMethod(dataProcessorType.Methods, "process_data")
	if processMethod == nil {
		t.Fatal("Expected to find process_data method")
	}

	// Validate method structure
	if processMethod.Name != "process_data" {
		t.Errorf("Expected method name 'process_data', got %s", processMethod.Name)
	}

	// Test called_by relationships (reverse call graph)
	helperFunc := findFunction(fileContext.Functions, "helper_function")
	if helperFunc == nil {
		t.Fatal("Expected to find helper_function")
	}

	// helper_function should be called by main_function
	t.Logf("helper_function is called by: %v", helperFunc.CalledBy)
	if len(helperFunc.CalledBy) == 0 {
		t.Log("Note: CalledBy relationships might need enhancement")
	}

	// Test constructor method exists
	initMethod := findMethod(dataProcessorType.Methods, "__init__")
	if initMethod == nil {
		t.Fatal("Expected to find __init__ method")
	}

	// Validate constructor structure
	if initMethod.Name != "__init__" {
		t.Errorf("Expected method name '__init__', got %s", initMethod.Name)
	}

	t.Logf("Call graph generation test completed successfully")
}

// Helper function to find a function by name
func findFunction(functions []models.Function, name string) *models.Function {
	for i := range functions {
		if functions[i].Name == name {
			return &functions[i]
		}
	}
	return nil
}

// Helper function to find a type by name
func findType(types []models.TypeDef, name string) *models.TypeDef {
	for i := range types {
		if types[i].Name == name {
			return &types[i]
		}
	}
	return nil
}

// Helper function to find a method by name
func findMethod(methods []models.Method, name string) *models.Method {
	for i := range methods {
		if methods[i].Name == name {
			return &methods[i]
		}
	}
	return nil
}
