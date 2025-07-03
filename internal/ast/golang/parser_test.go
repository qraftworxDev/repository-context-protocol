package golang

import (
	"testing"

	"repository-context-protocol/internal/models"
)

func TestGoParser_GetSupportedExtensions(t *testing.T) {
	parser := NewGoParser()
	extensions := parser.GetSupportedExtensions()

	expected := []string{".go"}
	if len(extensions) != len(expected) {
		t.Errorf("Expected %d extensions, got %d", len(expected), len(extensions))
	}

	if extensions[0] != ".go" {
		t.Errorf("Expected .go extension, got %s", extensions[0])
	}
}

func TestGoParser_GetLanguageName(t *testing.T) {
	parser := NewGoParser()
	language := parser.GetLanguageName()

	if language != "go" {
		t.Errorf("Expected language 'go', got %s", language)
	}
}

func TestGoParser_ParseSimpleFunction(t *testing.T) {
	parser := NewGoParser()

	code := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}

func greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}`

	fileContext, err := parser.ParseFile("main.go", []byte(code))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if fileContext.Path != "main.go" {
		t.Errorf("Expected path 'main.go', got %s", fileContext.Path)
	}
	if fileContext.Language != "go" {
		t.Errorf("Expected language 'go', got %s", fileContext.Language)
	}

	// Should have 2 functions: main and greet
	if len(fileContext.Functions) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(fileContext.Functions))
	}

	// Check main function
	mainFunc := findFunction(fileContext.Functions, "main")
	if mainFunc == nil {
		t.Error("Expected to find main function")
	} else {
		if len(mainFunc.Parameters) != 0 {
			t.Errorf("Expected main to have 0 parameters, got %d", len(mainFunc.Parameters))
		}
		if len(mainFunc.Returns) != 0 {
			t.Errorf("Expected main to have 0 returns, got %d", len(mainFunc.Returns))
		}
	}

	// Check greet function
	greetFunc := findFunction(fileContext.Functions, "greet")
	if greetFunc == nil {
		t.Error("Expected to find greet function")
	} else {
		if len(greetFunc.Parameters) != 1 {
			t.Errorf("Expected greet to have 1 parameter, got %d", len(greetFunc.Parameters))
		}
		if len(greetFunc.Returns) != 1 {
			t.Errorf("Expected greet to have 1 return, got %d", len(greetFunc.Returns))
		}
		if greetFunc.Parameters[0].Name != "name" {
			t.Errorf("Expected parameter name 'name', got %s", greetFunc.Parameters[0].Name)
		}
		if greetFunc.Parameters[0].Type != "string" {
			t.Errorf("Expected parameter type 'string', got %s", greetFunc.Parameters[0].Type)
		}
	}

	// Should have 1 import
	if len(fileContext.Imports) != 1 {
		t.Errorf("Expected 1 import, got %d", len(fileContext.Imports))
	}
	if fileContext.Imports[0].Path != "fmt" {
		t.Errorf("Expected import 'fmt', got %s", fileContext.Imports[0].Path)
	}
}

func TestGoParser_ParseStruct(t *testing.T) {
	parser := NewGoParser()

	code := `package models

type User struct {
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}

func (u User) String() string {
	return u.Name
}`

	fileContext, err := parser.ParseFile("user.go", []byte(code))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should have 1 type
	if len(fileContext.Types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(fileContext.Types))
	}

	userType := fileContext.Types[0]
	if userType.Name != "User" {
		t.Errorf("Expected type name 'User', got %s", userType.Name)
	}
	if userType.Kind != "struct" {
		t.Errorf("Expected type kind 'struct', got %s", userType.Kind)
	}

	// Should have 2 fields
	if len(userType.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(userType.Fields))
	}

	// Should have 1 method (String)
	if len(userType.Methods) != 1 {
		t.Errorf("Expected 1 method, got %d", len(userType.Methods))
	} else if userType.Methods[0].Name != "String" {
		t.Errorf("Expected method name 'String', got %s", userType.Methods[0].Name)
	}
}

func TestGoParser_ParseInterface(t *testing.T) {
	parser := NewGoParser()

	code := `package io

type Writer interface {
	Write([]byte) (int, error)
}`

	fileContext, err := parser.ParseFile("writer.go", []byte(code))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should have 1 type
	if len(fileContext.Types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(fileContext.Types))
	}

	writerType := fileContext.Types[0]
	if writerType.Name != "Writer" {
		t.Errorf("Expected type name 'Writer', got %s", writerType.Name)
	}
	if writerType.Kind != "interface" {
		t.Errorf("Expected type kind 'interface', got %s", writerType.Kind)
	}

	// Should have 0 fields but 1 method
	if len(writerType.Fields) != 0 {
		t.Errorf("Expected 0 fields for interface, got %d", len(writerType.Fields))
	}
	if len(writerType.Methods) != 1 {
		t.Errorf("Expected 1 method, got %d", len(writerType.Methods))
	}
}

func TestGoParser_InvalidCode(t *testing.T) {
	parser := NewGoParser()

	invalidCode := `package main

	func main( {
		// Invalid syntax - missing closing parenthesis
	}`

	_, err := parser.ParseFile("invalid.go", []byte(invalidCode))
	if err == nil {
		t.Error("Expected error for invalid Go code")
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

func TestGoParser_isPackageCall(t *testing.T) {
	parser := NewGoParser()

	tests := []struct {
		name     string
		imports  []models.Import
		callName string
		expected bool
	}{
		// Standard library without alias
		{
			name: "Standard library fmt",
			imports: []models.Import{
				{Path: "fmt", Alias: ""},
			},
			callName: "fmt",
			expected: true,
		},
		{
			name: "Standard library os",
			imports: []models.Import{
				{Path: "os", Alias: ""},
			},
			callName: "os",
			expected: true,
		},
		// Import with alias
		{
			name: "Standard library with alias",
			imports: []models.Import{
				{Path: "fmt", Alias: "f"},
			},
			callName: "f",
			expected: true,
		},
		{
			name: "Third-party package with alias",
			imports: []models.Import{
				{Path: "github.com/pkg/errors", Alias: "errs"},
			},
			callName: "errs",
			expected: true,
		},
		// Third-party packages
		{
			name: "Third-party package github.com/pkg/errors",
			imports: []models.Import{
				{Path: "github.com/pkg/errors", Alias: ""},
			},
			callName: "errors",
			expected: true,
		},
		{
			name: "Third-party package golang.org/x/tools",
			imports: []models.Import{
				{Path: "golang.org/x/tools/go/ast", Alias: ""},
			},
			callName: "ast",
			expected: true,
		},
		{
			name: "Deep path package",
			imports: []models.Import{
				{Path: "path/to/deep/package", Alias: ""},
			},
			callName: "package",
			expected: true,
		},
		// Non-package calls
		{
			name: "Non-package call",
			imports: []models.Import{
				{Path: "fmt", Alias: ""},
			},
			callName: "myFunction",
			expected: false,
		},
		{
			name:     "Empty imports",
			imports:  []models.Import{},
			callName: "fmt",
			expected: false,
		},
		{
			name: "Call name doesn't match any import",
			imports: []models.Import{
				{Path: "fmt", Alias: ""},
				{Path: "os", Alias: ""},
			},
			callName: "strings",
			expected: false,
		},
		// Edge cases
		{
			name: "Alias doesn't match original package name",
			imports: []models.Import{
				{Path: "fmt", Alias: "f"},
			},
			callName: "fmt", // Should be false because the alias is "f"
			expected: false,
		},
		{
			name: "Multiple imports with same package name",
			imports: []models.Import{
				{Path: "fmt", Alias: ""},
				{Path: "custom/fmt", Alias: "customfmt"},
			},
			callName: "fmt",
			expected: true,
		},
		{
			name: "Dot import (alias is '.')",
			imports: []models.Import{
				{Path: "math", Alias: "."},
			},
			callName: "math", // Even with dot import, the package name should still work
			expected: false,  // Dot imports use "." as alias, so "math" won't match
		},
		{
			name: "Blank import (alias is '_')",
			imports: []models.Import{
				{Path: "database/sql/driver", Alias: "_"},
			},
			callName: "driver", // Blank imports shouldn't be callable
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isPackageCall(tt.callName, tt.imports)
			if result != tt.expected {
				t.Errorf("isPackageCall(%q, %+v) = %v, want %v", tt.callName, tt.imports, result, tt.expected)
			}
		})
	}
}

func TestGoParser_PackageCallIntegration(t *testing.T) {
	parser := NewGoParser()

	// Test complete integration with real Go code
	code := `package main

import (
	"fmt"
	f "fmt"
	"os"
	"github.com/pkg/errors"
	errs "github.com/pkg/errors"
	"golang.org/x/tools/go/ast"
	_ "database/sql/driver"
)

func main() {
	fmt.Println("standard fmt")
	f.Printf("aliased fmt")
	os.Open("file")
	errors.New("pkg errors")
	errs.Wrap(nil, "aliased errors")
	ast.Print(nil)
	localFunction()
	obj.method()
}

func localFunction() {}
`

	fileContext, err := parser.ParseFile("test.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Find the main function
	mainFunc := findFunction(fileContext.Functions, "main")
	if mainFunc == nil {
		t.Fatal("Main function not found")
	}

	// Check that package calls are correctly identified as external
	externalCalls := []string{"fmt.Println", "f.Printf", "os.Open", "errors.New", "errs.Wrap", "ast.Print"}
	localCalls := []string{"localFunction", "obj.method"}

	for _, call := range mainFunc.LocalCallsWithMetadata {
		callName := call.FunctionName
		callType := call.CallType

		// Check if it should be external
		shouldBeExternal := false
		for _, expectedExternal := range externalCalls {
			if callName == expectedExternal {
				shouldBeExternal = true
				break
			}
		}

		// Check if it should be local/method
		shouldBeLocal := false
		for _, expectedLocal := range localCalls {
			if callName == expectedLocal {
				shouldBeLocal = true
				break
			}
		}

		if shouldBeExternal {
			if callType != models.CallTypeExternal {
				t.Errorf("Expected %s to be CallTypeExternal, got %s", callName, callType)
			}
		} else if shouldBeLocal {
			if callType != models.CallTypeFunction && callType != models.CallTypeMethod {
				t.Errorf("Expected %s to be CallTypeFunction or CallTypeMethod, got %s", callName, callType)
			}
		}
	}

	// Verify that we correctly identified some external calls
	hasExternalCalls := false
	for _, call := range mainFunc.LocalCallsWithMetadata {
		if call.CallType == models.CallTypeExternal {
			hasExternalCalls = true
			break
		}
	}
	if !hasExternalCalls {
		t.Error("Expected to find at least one external call, but found none")
	}
}
