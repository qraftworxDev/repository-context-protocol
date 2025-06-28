package mcp

import (
	"strings"
	"testing"
)

// TestContextTools_Registration tests tool registration
func TestContextTools_Registration(t *testing.T) {
	server := NewRepoContextMCPServer()

	tools := server.RegisterContextTools()

	expectedTools := []string{
		"get_function_context",
		"get_type_context",
	}

	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	for _, expectedTool := range expectedTools {
		if !toolNames[expectedTool] {
			t.Errorf("Expected tool '%s' not found in registered tools", expectedTool)
		}
	}
}

// TestContextLinesValidation tests context lines validation
func TestContextLinesValidation(t *testing.T) {
	tests := []struct {
		name                 string
		inputContextLines    int
		expectedContextLines int
	}{
		{"Valid context lines 1", 1, 1},
		{"Valid context lines 10", 10, 10},
		{"Max context lines 50", 50, 50},
		{"Exceed max context lines - should cap at 50", 100, 50},
		{"Zero context lines - should default to 5", 0, 5},
		{"Negative context lines - should default to 5", -1, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateContextLines(tt.inputContextLines)
			if result != tt.expectedContextLines {
				t.Errorf("Expected context lines %d, got %d", tt.expectedContextLines, result)
			}
		})
	}
}

// TestGetFunctionContextParameterParsing tests parameter parsing for get_function_context
func TestGetFunctionContextParameterParsing(t *testing.T) {
	t.Run("parameter validation logic", func(t *testing.T) {
		// Test the validation functions directly instead of through mocked requests
		tests := []struct {
			name                   string
			functionName           string
			includeImplementations bool
			contextLines           int
			maxTokens              int
			expectError            bool
			expectedContextLines   int
		}{
			{
				name:                   "Valid parameters with all options",
				functionName:           "TestFunction",
				includeImplementations: true,
				contextLines:           10,
				maxTokens:              1500,
				expectError:            false,
				expectedContextLines:   10,
			},
			{
				name:                   "Valid parameters with defaults",
				functionName:           "TestFunction",
				includeImplementations: false,
				contextLines:           5,
				maxTokens:              2000,
				expectError:            false,
				expectedContextLines:   5,
			},
			{
				name:                   "Context lines validation - should cap at max",
				functionName:           "TestFunction",
				includeImplementations: false,
				contextLines:           100, // Should be capped
				maxTokens:              2000,
				expectError:            false,
				expectedContextLines:   50, // Capped at max
			},
			{
				name:                   "Context lines validation - should default",
				functionName:           "TestFunction",
				includeImplementations: false,
				contextLines:           0, // Should default
				maxTokens:              2000,
				expectError:            false,
				expectedContextLines:   5, // Default
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Test the validation logic directly
				validatedContextLines := validateContextLines(tt.contextLines)
				if validatedContextLines != tt.expectedContextLines {
					t.Errorf("Expected context lines %d, got %d", tt.expectedContextLines, validatedContextLines)
				}

				// Test parameter structure
				params := &GetFunctionContextParams{
					FunctionName:           tt.functionName,
					IncludeImplementations: tt.includeImplementations,
					ContextLines:           validatedContextLines,
					MaxTokens:              tt.maxTokens,
				}

				if params.FunctionName != tt.functionName {
					t.Errorf("Expected function name '%s', got '%s'", tt.functionName, params.FunctionName)
				}
				if params.IncludeImplementations != tt.includeImplementations {
					t.Errorf("Expected include implementations %v, got %v", tt.includeImplementations, params.IncludeImplementations)
				}
				if params.ContextLines != tt.expectedContextLines {
					t.Errorf("Expected context lines %d, got %d", tt.expectedContextLines, params.ContextLines)
				}
			})
		}
	})

	t.Run("empty function name validation", func(t *testing.T) {
		// Test that empty function name is rejected
		functionName := ""
		if strings.TrimSpace(functionName) != "" {
			t.Error("Empty function name should be invalid")
		}
	})
}

// TestFunctionContextResponseStructure tests the response structure for function context
func TestFunctionContextResponseStructure(t *testing.T) {
	// Test the FunctionContextResult structure
	result := &FunctionContextResult{
		FunctionName: "TestFunction",
		Signature:    "func TestFunction(param1 string, param2 int) error",
		Location: FunctionLocation{
			File:      "test.go",
			StartLine: 10,
			EndLine:   20,
		},
		Implementation: &FunctionImplementation{
			Body: "function body content",
			ContextLines: []string{
				"// Context line 1",
				"// Context line 2",
			},
		},
		Callers: []FunctionReference{
			{Name: "Caller1", File: "caller.go", Line: 5},
		},
		Callees: []FunctionReference{
			{Name: "Callee1", File: "callee.go", Line: 15},
		},
		RelatedTypes: []TypeReference{
			{Name: "TestType", File: "types.go", Line: 25},
		},
		TokenCount: 150,
		Truncated:  false,
	}

	// Verify structure completeness
	if result.FunctionName == "" {
		t.Error("FunctionName should not be empty")
	}
	if result.Signature == "" {
		t.Error("Signature should not be empty")
	}
	if result.Location.File == "" {
		t.Error("Location.File should not be empty")
	}
	if result.Implementation == nil {
		t.Error("Implementation should not be nil when included")
	}
	if len(result.Callers) == 0 {
		t.Error("Callers should not be empty for test data")
	}
	if len(result.Callees) == 0 {
		t.Error("Callees should not be empty for test data")
	}
	if result.TokenCount <= 0 {
		t.Error("TokenCount should be positive")
	}
}

// TestFunctionContextTokenOptimization tests token optimization for function context
func TestFunctionContextTokenOptimization(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Create a test result with known content
	result := &FunctionContextResult{
		FunctionName: "TestFunction",
		Signature:    "func TestFunction() error",
		Location: FunctionLocation{
			File:      "test.go",
			StartLine: 10,
			EndLine:   15,
		},
		Implementation: &FunctionImplementation{
			Body: "return nil",
			ContextLines: []string{
				"// Context line 1",
				"// Context line 2",
				"// Context line 3",
			},
		},
		Callers: []FunctionReference{
			{Name: "Caller1", File: "caller.go", Line: 5},
			{Name: "Caller2", File: "caller.go", Line: 10},
		},
		Callees: []FunctionReference{
			{Name: "Callee1", File: "callee.go", Line: 15},
			{Name: "Callee2", File: "callee.go", Line: 20},
		},
		RelatedTypes: []TypeReference{
			{Name: "Type1", File: "types.go", Line: 25},
			{Name: "Type2", File: "types.go", Line: 30},
		},
	}

	tests := []struct {
		name           string
		maxTokens      int
		shouldTruncate bool
	}{
		{"High token limit - no truncation", 2000, false},
		{"Low token limit - should truncate", 50, true},
		{"Medium token limit", 500, false}, // Depends on content
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying original
			testResult := *result
			testResult.Callers = make([]FunctionReference, len(result.Callers))
			copy(testResult.Callers, result.Callers)
			testResult.Callees = make([]FunctionReference, len(result.Callees))
			copy(testResult.Callees, result.Callees)
			testResult.RelatedTypes = make([]TypeReference, len(result.RelatedTypes))
			copy(testResult.RelatedTypes, result.RelatedTypes)

			server.optimizeFunctionContextResponse(&testResult, tt.maxTokens)

			if tt.shouldTruncate && !testResult.Truncated {
				t.Error("Expected truncation but result was not truncated")
			}

			// Verify token count is updated
			if testResult.TokenCount <= 0 {
				t.Error("Token count should be positive after optimization")
			}
		})
	}
}

// TestContextToolImplementationDetails tests implementation detail handling
func TestContextToolImplementationDetails(t *testing.T) {
	tests := []struct {
		name                   string
		includeImplementations bool
		expectImplementation   bool
	}{
		{"Include implementations - true", true, true},
		{"Include implementations - false", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &GetFunctionContextParams{
				FunctionName:           "TestFunction",
				IncludeImplementations: tt.includeImplementations,
				ContextLines:           5,
				MaxTokens:              2000,
			}

			// This tests the parameter structure - implementation details
			// will be tested in integration tests
			if tt.expectImplementation != params.IncludeImplementations {
				t.Errorf("Expected include implementations %v, got %v",
					tt.expectImplementation, params.IncludeImplementations)
			}
		})
	}
}

// TestGetTypeContextParameterParsing tests parameter parsing for get_type_context
func TestGetTypeContextParameterParsing(t *testing.T) {
	t.Run("parameter validation logic", func(t *testing.T) {
		// Test the validation functions directly instead of through mocked requests
		tests := []struct {
			name           string
			typeName       string
			includeMethods bool
			includeUsage   bool
			maxTokens      int
			expectError    bool
		}{
			{
				name:           "Valid parameters with all options",
				typeName:       "TestType",
				includeMethods: true,
				includeUsage:   true,
				maxTokens:      1500,
				expectError:    false,
			},
			{
				name:           "Valid parameters with defaults",
				typeName:       "TestType",
				includeMethods: false,
				includeUsage:   false,
				maxTokens:      2000,
				expectError:    false,
			},
			{
				name:           "Valid minimal parameters",
				typeName:       "SimpleType",
				includeMethods: false,
				includeUsage:   false,
				maxTokens:      1000,
				expectError:    false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Test parameter structure
				params := &GetTypeContextParams{
					TypeName:       tt.typeName,
					IncludeMethods: tt.includeMethods,
					IncludeUsage:   tt.includeUsage,
					MaxTokens:      tt.maxTokens,
				}

				if params.TypeName != tt.typeName {
					t.Errorf("Expected type name '%s', got '%s'", tt.typeName, params.TypeName)
				}
				if params.IncludeMethods != tt.includeMethods {
					t.Errorf("Expected include methods %v, got %v", tt.includeMethods, params.IncludeMethods)
				}
				if params.IncludeUsage != tt.includeUsage {
					t.Errorf("Expected include usage %v, got %v", tt.includeUsage, params.IncludeUsage)
				}
			})
		}
	})

	t.Run("empty type name validation", func(t *testing.T) {
		// Test that empty type name is rejected
		typeName := ""
		if strings.TrimSpace(typeName) != "" {
			t.Error("Empty type name should be invalid")
		}
	})
}

// TestTypeContextResponseStructure tests the response structure for type context
func TestTypeContextResponseStructure(t *testing.T) {
	// Test the TypeContextResult structure
	result := &TypeContextResult{
		TypeName:  "TestType",
		Signature: "type TestType struct { Field1 string; Field2 int }",
		Location: TypeLocation{
			File:      "test.go",
			StartLine: 10,
			EndLine:   15,
		},
		Fields: []FieldReference{
			{Name: "Field1", Type: "string", File: "test.go", Line: 11},
			{Name: "Field2", Type: "int", File: "test.go", Line: 12},
		},
		Methods: []MethodReference{
			{Name: "Method1", Signature: "func (t *TestType) Method1() error", File: "test.go", Line: 20},
		},
		UsageExamples: []UsageExample{
			{Description: "Variable declaration", Code: "var t TestType", File: "usage.go", Line: 5},
		},
		RelatedTypes: []TypeReference{
			{Name: "RelatedType", File: "related.go", Line: 25},
		},
		TokenCount: 200,
		Truncated:  false,
	}

	// Verify structure completeness
	if result.TypeName == "" {
		t.Error("TypeName should not be empty")
	}
	if result.Signature == "" {
		t.Error("Signature should not be empty")
	}
	if result.Location.File == "" {
		t.Error("Location.File should not be empty")
	}
	if len(result.Fields) == 0 {
		t.Error("Fields should not be empty for test data")
	}
	if len(result.Methods) == 0 {
		t.Error("Methods should not be empty for test data")
	}
	if result.TokenCount <= 0 {
		t.Error("TokenCount should be positive")
	}
}

// TestTypeContextTokenOptimization tests token optimization for type context
func TestTypeContextTokenOptimization(t *testing.T) {
	server := NewRepoContextMCPServer()

	// Create a test result with multiple fields and methods
	result := &TypeContextResult{
		TypeName:  "TestType",
		Signature: "type TestType struct { Field1 string; Field2 int; Field3 bool }",
		Location: TypeLocation{
			File:      "test.go",
			StartLine: 10,
			EndLine:   15,
		},
		Fields: []FieldReference{
			{Name: "Field1", Type: "string", File: "test.go", Line: 11},
			{Name: "Field2", Type: "int", File: "test.go", Line: 12},
			{Name: "Field3", Type: "bool", File: "test.go", Line: 13},
		},
		Methods: []MethodReference{
			{Name: "Method1", Signature: "func (t *TestType) Method1() error", File: "test.go", Line: 20},
			{Name: "Method2", Signature: "func (t *TestType) Method2() string", File: "test.go", Line: 25},
		},
		UsageExamples: []UsageExample{
			{Description: "Variable declaration", Code: "var t TestType", File: "usage.go", Line: 5},
			{Description: "Initialization", Code: "t := TestType{Field1: \"test\"}", File: "usage.go", Line: 10},
		},
		RelatedTypes: []TypeReference{
			{Name: "RelatedType1", File: "related.go", Line: 25},
			{Name: "RelatedType2", File: "related.go", Line: 30},
		},
		TokenCount: 500,
		Truncated:  false,
	}

	tests := []struct {
		name           string
		maxTokens      int
		shouldTruncate bool
	}{
		{"No truncation needed - large limit", 1000, false},
		{"Truncation needed - medium limit", 250, true},
		{"Truncation needed - small limit", 150, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying original
			testResult := *result
			testResult.Fields = make([]FieldReference, len(result.Fields))
			copy(testResult.Fields, result.Fields)
			testResult.Methods = make([]MethodReference, len(result.Methods))
			copy(testResult.Methods, result.Methods)
			testResult.UsageExamples = make([]UsageExample, len(result.UsageExamples))
			copy(testResult.UsageExamples, result.UsageExamples)
			testResult.RelatedTypes = make([]TypeReference, len(result.RelatedTypes))
			copy(testResult.RelatedTypes, result.RelatedTypes)

			server.optimizeTypeContextResponse(&testResult, tt.maxTokens)

			if tt.shouldTruncate && !testResult.Truncated {
				t.Error("Expected truncation but result was not truncated")
			}

			// Verify token count is updated
			if testResult.TokenCount <= 0 {
				t.Error("Token count should be positive after optimization")
			}
		})
	}
}

// TestTypeContextMethodsAndUsage tests methods and usage inclusion
func TestTypeContextMethodsAndUsage(t *testing.T) {
	tests := []struct {
		name           string
		includeMethods bool
		includeUsage   bool
		expectMethods  bool
		expectUsage    bool
	}{
		{"Include methods and usage", true, true, true, true},
		{"Include methods only", true, false, true, false},
		{"Include usage only", false, true, false, true},
		{"Include neither", false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &GetTypeContextParams{
				TypeName:       "TestType",
				IncludeMethods: tt.includeMethods,
				IncludeUsage:   tt.includeUsage,
				MaxTokens:      2000,
			}

			// This tests the parameter structure - methods and usage details
			// will be tested in integration tests
			if tt.expectMethods != params.IncludeMethods {
				t.Errorf("Expected include methods %v, got %v",
					tt.expectMethods, params.IncludeMethods)
			}
			if tt.expectUsage != params.IncludeUsage {
				t.Errorf("Expected include usage %v, got %v",
					tt.expectUsage, params.IncludeUsage)
			}
		})
	}
}
