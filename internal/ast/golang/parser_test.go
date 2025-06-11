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
