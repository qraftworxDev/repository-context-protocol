package golang

import (
	"testing"
)

func TestGoParser_VariableExtraction(t *testing.T) {
	parser := NewGoParser()

	code := `package test

// Various variable declarations
var (
	GlobalInt    int = 42
	GlobalString     = "hello"
	GlobalBool       = true
	GlobalFloat      = 3.14
)

// Single variable declarations
var SingleVar string = "single"
var InferredVar = 100

// Constants
const (
	MaxSize     = 1000
	DefaultName = "default"
	Pi          = 3.14159
	IsEnabled   = true
)

// Single constant
const SingleConst int = 999`

	fileContext, err := parser.ParseFile("variables.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Should have 6 variables total
	expectedVarCount := 6
	if len(fileContext.Variables) != expectedVarCount {
		t.Errorf("Expected %d variables, got %d", expectedVarCount, len(fileContext.Variables))
	}

	// Should have 5 constants total
	expectedConstCount := 5
	if len(fileContext.Constants) != expectedConstCount {
		t.Errorf("Expected %d constants, got %d", expectedConstCount, len(fileContext.Constants))
	}

	// Check specific variables and their types
	expectedVars := map[string]string{
		"GlobalInt":    "int",
		"GlobalString": "string",
		"GlobalBool":   "bool",
		"GlobalFloat":  "float64",
		"SingleVar":    "string",
		"InferredVar":  "int",
	}

	// Check specific constants and their types
	expectedConsts := map[string]string{
		"MaxSize":     "int",
		"DefaultName": "string",
		"Pi":          "float64",
		"IsEnabled":   "bool",
		"SingleConst": "int",
	}

	// Create a map of actual variables for easy lookup
	actualVars := make(map[string]string)
	for _, variable := range fileContext.Variables {
		actualVars[variable.Name] = variable.Type
	}

	// Check each expected variable
	for name, expectedType := range expectedVars {
		actualType, found := actualVars[name]
		if !found {
			t.Errorf("Expected to find variable '%s', but it was not extracted", name)
			continue
		}
		if actualType != expectedType {
			t.Errorf("Variable '%s' expected type '%s', got '%s'", name, expectedType, actualType)
		}
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

	// Check line numbers are set for variables
	for _, variable := range fileContext.Variables {
		if variable.StartLine == 0 || variable.EndLine == 0 {
			t.Errorf("Variable '%s' should have non-zero line numbers, got start=%d, end=%d",
				variable.Name, variable.StartLine, variable.EndLine)
		}
	}

	// Check line numbers are set for constants
	for _, constant := range fileContext.Constants {
		if constant.StartLine == 0 || constant.EndLine == 0 {
			t.Errorf("Constant '%s' should have non-zero line numbers, got start=%d, end=%d",
				constant.Name, constant.StartLine, constant.EndLine)
		}
	}
}

func TestGoParser_ComplexVariableTypes(t *testing.T) {
	parser := NewGoParser()

	code := `package test

import "time"

var (
	// Complex types
	SliceVar    []string = []string{"a", "b"}
	MapVar      map[string]int = make(map[string]int)
	ChanVar     chan int = make(chan int)
	FuncVar     func(int) string
	PointerVar  *int
	InterfaceVar interface{}

	// Package types
	TimeVar time.Time
)`

	fileContext, err := parser.ParseFile("complex_vars.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	expectedTypes := map[string]string{
		"SliceVar":     "[]string",
		"MapVar":       "map[string]int",
		"ChanVar":      "chan int",
		"FuncVar":      "func(int) string",
		"PointerVar":   "*int",
		"InterfaceVar": "interface{}",
		"TimeVar":      "time.Time",
	}

	actualVars := make(map[string]string)
	for _, variable := range fileContext.Variables {
		actualVars[variable.Name] = variable.Type
	}

	for name, expectedType := range expectedTypes {
		actualType, found := actualVars[name]
		if !found {
			t.Errorf("Expected to find variable '%s', but it was not extracted", name)
			continue
		}
		if actualType != expectedType {
			t.Errorf("Variable '%s' expected type '%s', got '%s'", name, expectedType, actualType)
		}
	}
}

func TestGoParser_VariableLineNumbers(t *testing.T) {
	parser := NewGoParser()

	code := `package test

var FirstVar = "first"

const SecondVar = 42

var (
	ThirdVar  = "third"
	FourthVar = 4
)

const (
	FifthVar = "fifth"
	SixthVar = 6
)`

	fileContext, err := parser.ParseFile("line_numbers.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Check that variables are on different lines and in order
	varLines := make(map[string]int)
	for _, variable := range fileContext.Variables {
		varLines[variable.Name] = variable.StartLine
	}

	// Only check variables (SecondVar, FifthVar, SixthVar are constants)
	expectedOrder := []string{"FirstVar", "ThirdVar", "FourthVar"}

	for i := 1; i < len(expectedOrder); i++ {
		prevVar := expectedOrder[i-1]
		currVar := expectedOrder[i]

		prevLine, prevExists := varLines[prevVar]
		currLine, currExists := varLines[currVar]

		if !prevExists {
			t.Errorf("Expected to find variable '%s'", prevVar)
			continue
		}
		if !currExists {
			t.Errorf("Expected to find variable '%s'", currVar)
			continue
		}

		if prevLine >= currLine {
			t.Errorf("Variable '%s' (line %d) should be before '%s' (line %d)",
				prevVar, prevLine, currVar, currLine)
		}
	}
}

func TestGoParser_EmptyVariables(t *testing.T) {
	parser := NewGoParser()

	code := `package test

// No variables or constants
func main() {
	// Local variables should not be extracted
	localVar := "local"
	_ = localVar
}`

	fileContext, err := parser.ParseFile("no_vars.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	if len(fileContext.Variables) != 0 {
		t.Errorf("Expected no variables, got %d", len(fileContext.Variables))
	}
}

func TestGoParser_VariableTypeInference(t *testing.T) {
	parser := NewGoParser()

	code := `package test

var (
	// Type inference from literals
	IntVar    = 42
	FloatVar  = 3.14
	StringVar = "hello"
	BoolVar   = true
	RuneVar   = 'a'
)

const (
	// Constants with type inference
	ConstInt    = 100
	ConstFloat  = 2.71
	ConstString = "world"
	ConstBool   = false
)`

	fileContext, err := parser.ParseFile("type_inference.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Check variables only
	expectedVarTypes := map[string]string{
		"IntVar":    "int",
		"FloatVar":  "float64",
		"StringVar": "string",
		"BoolVar":   "bool",
		"RuneVar":   "rune",
	}

	actualVars := make(map[string]string)
	for _, variable := range fileContext.Variables {
		actualVars[variable.Name] = variable.Type
	}

	for name, expectedType := range expectedVarTypes {
		actualType, found := actualVars[name]
		if !found {
			t.Errorf("Expected to find variable '%s', but it was not extracted", name)
			continue
		}
		if actualType != expectedType {
			t.Errorf("Variable '%s' expected type '%s', got '%s'", name, expectedType, actualType)
		}
	}

	// Check constants separately
	expectedConstTypes := map[string]string{
		"ConstInt":    "int",
		"ConstFloat":  "float64",
		"ConstString": "string",
		"ConstBool":   "bool",
	}

	actualConsts := make(map[string]string)
	for _, constant := range fileContext.Constants {
		actualConsts[constant.Name] = constant.Type
	}

	for name, expectedType := range expectedConstTypes {
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
