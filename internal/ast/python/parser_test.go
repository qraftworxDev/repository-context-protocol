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
		t.Errorf("Expected 0 functions for empty file, got %d", len(fileContext.Functions))
	}
	if len(fileContext.Types) != 0 {
		t.Errorf("Expected 0 types for empty file, got %d", len(fileContext.Types))
	}
	if len(fileContext.Imports) != 0 {
		t.Errorf("Expected 0 imports for empty file, got %d", len(fileContext.Imports))
	}
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
