package golang

import (
	"testing"
)

func TestGoParser_CallGraphAnalysis(t *testing.T) {
	parser := NewGoParser()

	code := `package main

import "fmt"

func helper() string {
	return "helper"
}

func process(data string) string {
	result := helper()
	return fmt.Sprintf("processed: %s", result)
}

func main() {
	data := "test"
	result := process(data)
	fmt.Println(result)
}`

	fileContext, err := parser.ParseFile("callgraph.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Find functions
	helperFunc := findFunction(fileContext.Functions, "helper")
	processFunc := findFunction(fileContext.Functions, "process")
	mainFunc := findFunction(fileContext.Functions, "main")

	if helperFunc == nil || processFunc == nil || mainFunc == nil {
		t.Fatal("Expected to find all three functions")
	}

	// Test helper function
	if len(helperFunc.Calls) != 0 {
		t.Errorf("Expected helper to have 0 calls, got %d: %v", len(helperFunc.Calls), helperFunc.Calls)
	}
	if len(helperFunc.CalledBy) != 1 || helperFunc.CalledBy[0] != "process" {
		t.Errorf("Expected helper to be called by [process], got %v", helperFunc.CalledBy)
	}

	// Test process function
	expectedProcessCalls := []string{"helper", "fmt.Sprintf"}
	if len(processFunc.Calls) != len(expectedProcessCalls) {
		t.Errorf("Expected process to have %d calls, got %d: %v", len(expectedProcessCalls), len(processFunc.Calls), processFunc.Calls)
	}
	for _, expectedCall := range expectedProcessCalls {
		found := false
		for _, actualCall := range processFunc.Calls {
			if actualCall == expectedCall {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected process to call %s, but it wasn't found in %v", expectedCall, processFunc.Calls)
		}
	}
	if len(processFunc.CalledBy) != 1 || processFunc.CalledBy[0] != "main" {
		t.Errorf("Expected process to be called by [main], got %v", processFunc.CalledBy)
	}

	// Test main function
	expectedMainCalls := []string{"process", "fmt.Println"}
	if len(mainFunc.Calls) != len(expectedMainCalls) {
		t.Errorf("Expected main to have %d calls, got %d: %v", len(expectedMainCalls), len(mainFunc.Calls), mainFunc.Calls)
	}
	for _, expectedCall := range expectedMainCalls {
		found := false
		for _, actualCall := range mainFunc.Calls {
			if actualCall == expectedCall {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected main to call %s, but it wasn't found in %v", expectedCall, mainFunc.Calls)
		}
	}
	if len(mainFunc.CalledBy) != 0 {
		t.Errorf("Expected main to have 0 callers, got %v", mainFunc.CalledBy)
	}
}

func TestGoParser_MethodCallAnalysis(t *testing.T) {
	parser := NewGoParser()

	code := `package main

type Calculator struct {
	value int
}

func (c *Calculator) Add(n int) {
	c.value += n
}

func (c Calculator) GetValue() int {
	return c.value
}

func NewCalculator() *Calculator {
	return &Calculator{value: 0}
}

func main() {
	calc := NewCalculator()
	calc.Add(5)
	result := calc.GetValue()
	_ = result
}`

	fileContext, err := parser.ParseFile("methods.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Find main function
	mainFunc := findFunction(fileContext.Functions, "main")
	if mainFunc == nil {
		t.Fatal("Expected to find main function")
	}

	// Check that main calls the constructor and methods
	if len(mainFunc.Calls) < 3 {
		t.Errorf("Expected main to have at least 3 calls, got %d: %v", len(mainFunc.Calls), mainFunc.Calls)
	}

	// Check for specific method calls
	hasNewCalculator := false
	hasMethodCall := false
	for _, call := range mainFunc.Calls {
		if call == "NewCalculator" {
			hasNewCalculator = true
		}
		if call == "calc.Add" || call == "calc.GetValue" {
			hasMethodCall = true
		}
	}

	if !hasNewCalculator {
		t.Error("Expected main to call NewCalculator")
	}
	if !hasMethodCall {
		t.Error("Expected main to call at least one method")
	}
}

func TestGoParser_EmptyFunctionCalls(t *testing.T) {
	parser := NewGoParser()

	code := `package main

func empty() {
	// No function calls
}

func main() {
	empty()
}`

	fileContext, err := parser.ParseFile("empty.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	emptyFunc := findFunction(fileContext.Functions, "empty")
	if emptyFunc == nil {
		t.Fatal("Expected to find empty function")
	}

	// Test that empty function has empty calls array (not nil)
	if emptyFunc.Calls == nil {
		t.Error("Expected Calls to be empty array, not nil")
	}
	if len(emptyFunc.Calls) != 0 {
		t.Errorf("Expected empty function to have 0 calls, got %d", len(emptyFunc.Calls))
	}

	// Test that empty function is called by main
	if len(emptyFunc.CalledBy) != 1 || emptyFunc.CalledBy[0] != "main" {
		t.Errorf("Expected empty to be called by [main], got %v", emptyFunc.CalledBy)
	}
}
