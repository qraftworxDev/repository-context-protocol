package golang

import (
	"testing"
)

func TestGoParser_ConstantExtraction(t *testing.T) {
	parser := NewGoParser()

	code := `package test

// Various constant declarations
const (
	MaxRetries   = 3
	DefaultName  = "default"
	Pi           = 3.14159
	IsEnabled    = true
	BufferSize   int = 1024
)

// Single constant declarations
const SingleConst string = "single"
const InferredConst = 42

// Typed constants
const TypedFloat float64 = 2.71828
const TypedBool bool = false`

	fileContext, err := parser.ParseFile("constants.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Should have 9 constants total
	expectedConstCount := 9
	if len(fileContext.Constants) != expectedConstCount {
		t.Errorf("Expected %d constants, got %d", expectedConstCount, len(fileContext.Constants))
	}

	// Should have 0 variables (all are constants)
	if len(fileContext.Variables) != 0 {
		t.Errorf("Expected 0 variables, got %d", len(fileContext.Variables))
	}

	// Check specific constants and their types
	expectedConsts := map[string]string{
		"MaxRetries":    "int",
		"DefaultName":   "string",
		"Pi":            "float64",
		"IsEnabled":     "bool",
		"BufferSize":    "int",
		"SingleConst":   "string",
		"InferredConst": "int",
		"TypedFloat":    "float64",
		"TypedBool":     "bool",
	}

	// Create a map of actual constants for easy lookup
	actualConsts := make(map[string]string)
	for _, constant := range fileContext.Constants {
		actualConsts[constant.Name] = constant.Type
	}

	// Check each expected constant
	for name, expectedType := range expectedConsts {
		actualType, found := actualConsts[name]
		if !found {
			t.Errorf("Expected to find constant '%s', but it was not extracted", name)
			continue
		}
		if actualType != expectedType {
			t.Errorf("Constant '%s' expected type '%s', got '%s'", name, expectedType, actualType)
		}
	}

	// Check line numbers are set
	for _, constant := range fileContext.Constants {
		if constant.StartLine == 0 || constant.EndLine == 0 {
			t.Errorf("Constant '%s' should have non-zero line numbers, got start=%d, end=%d",
				constant.Name, constant.StartLine, constant.EndLine)
		}
	}
}

func TestGoParser_ConstantValues(t *testing.T) {
	parser := NewGoParser()

	code := `package test

const (
	StringConst = "hello world"
	IntConst    = 42
	FloatConst  = 3.14
	BoolConst   = true
	RuneConst   = 'A'
)

const ImagConst = 2i`

	fileContext, err := parser.ParseFile("const_values.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Check that constants have values (if we implement value extraction)
	for _, constant := range fileContext.Constants {
		// For now, just check that the constant was extracted
		if constant.Name == "" {
			t.Errorf("Constant should have a name")
		}
		// Note: Some complex expressions might not have inferred types
		if constant.Type == "" && constant.Name != "ImagConst" {
			t.Errorf("Constant '%s' should have a type", constant.Name)
		}
	}
}

func TestGoParser_ConstantLineNumbers(t *testing.T) {
	parser := NewGoParser()

	code := `package test

const FirstConst = "first"

const SecondConst = 42

const (
	ThirdConst  = "third"
	FourthConst = 4
)

const (
	FifthConst = "fifth"
	SixthConst = 6
)`

	fileContext, err := parser.ParseFile("const_lines.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Check that constants are on different lines and in order
	constLines := make(map[string]int)
	for _, constant := range fileContext.Constants {
		constLines[constant.Name] = constant.StartLine
	}

	// FirstConst should be before SecondConst, etc.
	expectedOrder := []string{"FirstConst", "SecondConst", "ThirdConst", "FourthConst", "FifthConst", "SixthConst"}

	for i := 1; i < len(expectedOrder); i++ {
		prevConst := expectedOrder[i-1]
		currConst := expectedOrder[i]

		prevLine, prevExists := constLines[prevConst]
		currLine, currExists := constLines[currConst]

		if !prevExists {
			t.Errorf("Expected to find constant '%s'", prevConst)
			continue
		}
		if !currExists {
			t.Errorf("Expected to find constant '%s'", currConst)
			continue
		}

		if prevLine >= currLine {
			t.Errorf("Constant '%s' (line %d) should be before '%s' (line %d)",
				prevConst, prevLine, currConst, currLine)
		}
	}
}

func TestGoParser_MixedVariablesAndConstants(t *testing.T) {
	parser := NewGoParser()

	code := `package test

var GlobalVar = "variable"
const GlobalConst = "constant"

var (
	VarOne = 1
	VarTwo = 2
)

const (
	ConstOne = 10
	ConstTwo = 20
)`

	fileContext, err := parser.ParseFile("mixed.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Should have 3 variables
	expectedVarCount := 3
	if len(fileContext.Variables) != expectedVarCount {
		t.Errorf("Expected %d variables, got %d", expectedVarCount, len(fileContext.Variables))
	}

	// Should have 3 constants
	expectedConstCount := 3
	if len(fileContext.Constants) != expectedConstCount {
		t.Errorf("Expected %d constants, got %d", expectedConstCount, len(fileContext.Constants))
	}

	// Check variable names
	varNames := make(map[string]bool)
	for _, variable := range fileContext.Variables {
		varNames[variable.Name] = true
	}

	expectedVars := []string{"GlobalVar", "VarOne", "VarTwo"}
	for _, expectedVar := range expectedVars {
		if !varNames[expectedVar] {
			t.Errorf("Expected to find variable '%s'", expectedVar)
		}
	}

	// Check constant names
	constNames := make(map[string]bool)
	for _, constant := range fileContext.Constants {
		constNames[constant.Name] = true
	}

	expectedConsts := []string{"GlobalConst", "ConstOne", "ConstTwo"}
	for _, expectedConst := range expectedConsts {
		if !constNames[expectedConst] {
			t.Errorf("Expected to find constant '%s'", expectedConst)
		}
	}
}

func TestGoParser_ConstantTypeInference(t *testing.T) {
	parser := NewGoParser()

	code := `package test

const (
	// Type inference from literals
	IntConst    = 42
	FloatConst  = 3.14
	StringConst = "hello"
	BoolConst   = true
	RuneConst   = 'a'

	// Explicit types
	ExplicitInt    int     = 100
	ExplicitFloat  float32 = 2.71
	ExplicitString string  = "world"
	ExplicitBool   bool    = false
)`

	fileContext, err := parser.ParseFile("const_types.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	expectedTypes := map[string]string{
		"IntConst":       "int",
		"FloatConst":     "float64",
		"StringConst":    "string",
		"BoolConst":      "bool",
		"RuneConst":      "rune",
		"ExplicitInt":    "int",
		"ExplicitFloat":  "float32",
		"ExplicitString": "string",
		"ExplicitBool":   "bool",
	}

	actualConsts := make(map[string]string)
	for _, constant := range fileContext.Constants {
		actualConsts[constant.Name] = constant.Type
	}

	for name, expectedType := range expectedTypes {
		actualType, found := actualConsts[name]
		if !found {
			t.Errorf("Expected to find constant '%s', but it was not extracted", name)
			continue
		}
		if actualType != expectedType {
			t.Errorf("Constant '%s' expected type '%s', got '%s'", name, expectedType, actualType)
		}
	}
}
