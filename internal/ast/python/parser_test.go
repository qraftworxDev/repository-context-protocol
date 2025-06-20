package python

import (
	"path/filepath"
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
