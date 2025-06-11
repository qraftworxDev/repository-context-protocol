package golang

import (
	"testing"

	"repository-context-protocol/internal/models"
)

func TestGoParser_ExportsExtraction(t *testing.T) {
	parser := NewGoParser()

	code := `package mypackage

import "fmt"

// Exported function
func PublicFunction() string {
	return "public"
}

// Non-exported function
func privateFunction() string {
	return "private"
}

// Exported type
type PublicStruct struct {
	PublicField   string
	privateField  int
}

// Non-exported type
type privateStruct struct {
	field string
}

// Exported interface
type PublicInterface interface {
	PublicMethod() string
}

// Non-exported interface
type privateInterface interface {
	method() string
}

// Exported variable
var PublicVar = "exported"

// Non-exported variable
var privateVar = "not exported"

// Exported constant
const PublicConst = 42

// Non-exported constant
const privateConst = 24

// Method on exported type
func (p PublicStruct) PublicMethod() string {
	return p.PublicField
}

// Method on exported type (non-exported method)
func (p PublicStruct) privateMethod() int {
	return p.privateField
}`

	fileContext, err := parser.ParseFile("exports.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Expected exports
	expectedExports := map[string]string{
		"PublicFunction":  "function",
		"PublicStruct":    "type",
		"PublicInterface": "type",
		"PublicVar":       "variable",
		"PublicMethod":    "function",
	}

	// Should NOT be exported
	notExported := []string{
		"privateFunction",
		"privateStruct",
		"privateInterface",
		"privateVar",
		"privateMethod",
	}

	// Check that expected exports are present
	exportMap := make(map[string]models.Export)
	for _, export := range fileContext.Exports {
		exportMap[export.Name] = export
	}

	for name, expectedKind := range expectedExports {
		export, found := exportMap[name]
		if !found {
			t.Errorf("Expected export %s not found", name)
			continue
		}
		if export.Kind != expectedKind {
			t.Errorf("Expected export %s to have kind %s, got %s", name, expectedKind, export.Kind)
		}
	}

	// Check that non-exported items are not in exports
	for _, name := range notExported {
		if _, found := exportMap[name]; found {
			t.Errorf("Non-exported item %s should not be in exports", name)
		}
	}

	// Verify export details
	if publicFunc, found := exportMap["PublicFunction"]; found {
		if publicFunc.Type != "func PublicFunction() string" {
			t.Errorf("Expected PublicFunction type to be 'func PublicFunction() string', got %s", publicFunc.Type)
		}
	}

	if publicStruct, found := exportMap["PublicStruct"]; found {
		if publicStruct.Type != "struct" {
			t.Errorf("Expected PublicStruct type to be 'struct', got %s", publicStruct.Type)
		}
	}
}

func TestGoParser_ExportsEdgeCases(t *testing.T) {
	parser := NewGoParser()

	code := `package test

// Single letter exports
func A() {}
func B() {}

// Single letter non-exports
func a() {}
func b() {}

// Mixed case
func CamelCase() {}
func camelCase() {}

// Numbers and underscores
func Public123() {}
func private123() {}
func Public_Func() {}
func private_func() {}

// Unicode (should work with Go's definition)
func Ñoño() {} // Starts with uppercase Unicode
func αβγ() {} // Starts with lowercase Unicode`

	fileContext, err := parser.ParseFile("edge_cases.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	exportMap := make(map[string]bool)
	for _, export := range fileContext.Exports {
		exportMap[export.Name] = true
	}

	// Should be exported
	exported := []string{"A", "B", "CamelCase", "Public123", "Public_Func", "Ñoño"}
	for _, name := range exported {
		if !exportMap[name] {
			t.Errorf("Expected %s to be exported", name)
		}
	}

	// Should NOT be exported
	notExported := []string{"a", "b", "camelCase", "private123", "private_func", "αβγ"}
	for _, name := range notExported {
		if exportMap[name] {
			t.Errorf("Expected %s to NOT be exported", name)
		}
	}
}

func TestGoParser_EmptyExports(t *testing.T) {
	parser := NewGoParser()

	code := `package internal

// All functions are non-exported
func helper() {}
func process() {}
func validate() {}

// All types are non-exported
type config struct {
	value string
}

type handler interface {
	handle() error
}`

	fileContext, err := parser.ParseFile("internal.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Should have no exports
	if len(fileContext.Exports) != 0 {
		t.Errorf("Expected no exports, got %d: %v", len(fileContext.Exports), fileContext.Exports)
	}

	// Ensure exports array is not nil
	if fileContext.Exports == nil {
		t.Error("Exports array should not be nil, should be empty slice")
	}
}
