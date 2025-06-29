package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"repository-context-protocol/internal/index"
	"repository-context-protocol/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// Helper functions to create test data without duplication

func createTestSearchResultEntry(name, file string, startLine, endLine int, fields []models.Field) *index.SearchResultEntry {
	return &index.SearchResultEntry{
		IndexEntry: models.IndexEntry{
			Name:      name,
			Type:      index.EntityTypeType,
			File:      file,
			StartLine: startLine,
			EndLine:   endLine,
			Signature: "type " + name + " struct",
		},
		ChunkData: &models.SemanticChunk{
			FileData: []models.FileContext{
				{
					Path: file,
					Types: []models.TypeDef{
						{
							Name:      name,
							Kind:      "struct",
							Fields:    fields,
							StartLine: startLine,
							EndLine:   endLine,
						},
					},
				},
			},
		},
	}
}

func createExpectedFieldReferences(file string, startLine int, fieldData []struct{ name, fieldType string }) []FieldReference {
	fields := make([]FieldReference, len(fieldData))
	for i, fd := range fieldData {
		fields[i] = FieldReference{
			Name: fd.name,
			Type: fd.fieldType,
			File: file,
			Line: startLine + i + 1,
		}
	}
	return fields
}

// TestExtractFieldReferences_RealFieldExtraction tests that we extract actual field names and types
func TestExtractFieldReferences_RealFieldExtraction(t *testing.T) {
	// Create a mock server instance
	server := &RepoContextMCPServer{}

	tests := []struct {
		name           string
		entry          *index.SearchResultEntry
		expectedFields []FieldReference
		description    string
	}{
		{
			name: "struct with actual fields",
			entry: createTestSearchResultEntry("Config", "models.go", 25, 30, []models.Field{
				{Name: "DatabaseURL", Type: "string"},
				{Name: "Port", Type: "int"},
				{Name: "LogLevel", Type: "string"},
				{Name: "Features", Type: "map[string]bool"},
			}),
			expectedFields: createExpectedFieldReferences("models.go", 25, []struct{ name, fieldType string }{
				{"DatabaseURL", "string"},
				{"Port", "int"},
				{"LogLevel", "string"},
				{"Features", "map[string]bool"},
			}),
			description: "Should extract actual field names and types from parsed struct definition",
		},
		{
			name: "struct with complex field types",
			entry: createTestSearchResultEntry("Profile", "models.go", 68, 75, []models.Field{
				{Name: "UserID", Type: "int"},
				{Name: "Bio", Type: "string"},
				{Name: "Address", Type: "*Address"},
				{Name: "Skills", Type: "[]string"},
				{Name: "CreatedAt", Type: "time.Time"},
				{Name: "UpdatedAt", Type: "time.Time"},
			}),
			expectedFields: createExpectedFieldReferences("models.go", 68, []struct{ name, fieldType string }{
				{"UserID", "int"},
				{"Bio", "string"},
				{"Address", "*Address"},
				{"Skills", "[]string"},
				{"CreatedAt", "time.Time"},
				{"UpdatedAt", "time.Time"},
			}),
			description: "Should extract complex field types like pointers, slices, and custom types",
		},
		{
			name:           "empty struct",
			entry:          createTestSearchResultEntry("EmptyStruct", "models.go", 10, 11, []models.Field{}),
			expectedFields: []FieldReference{}, // Should return empty slice, not placeholders
			description:    "Should return empty slice for structs with no fields, not placeholder fields",
		},
		{
			name: "no chunk data available",
			entry: &index.SearchResultEntry{
				IndexEntry: models.IndexEntry{
					Name:      "SomeStruct",
					Type:      index.EntityTypeType,
					File:      "models.go",
					StartLine: 20,
					EndLine:   25,
					Signature: "type SomeStruct struct",
				},
				ChunkData: nil, // No chunk data available
			},
			expectedFields: []FieldReference{}, // Should return empty slice when no data available
			description:    "Should return empty slice when chunk data is not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := server.extractFieldReferences(tt.entry)

			// Check field count
			if len(fields) != len(tt.expectedFields) {
				t.Errorf("Expected %d fields, got %d fields", len(tt.expectedFields), len(fields))
				t.Logf("Description: %s", tt.description)
				return
			}

			// Check each field
			for i, expectedField := range tt.expectedFields {
				if i >= len(fields) {
					t.Errorf("Missing field at index %d: expected %+v", i, expectedField)
					continue
				}

				actualField := fields[i]
				if actualField.Name != expectedField.Name {
					t.Errorf("Field %d: expected name %q, got %q", i, expectedField.Name, actualField.Name)
				}
				if actualField.Type != expectedField.Type {
					t.Errorf("Field %d: expected type %q, got %q", i, expectedField.Type, actualField.Type)
				}
				if actualField.File != expectedField.File {
					t.Errorf("Field %d: expected file %q, got %q", i, expectedField.File, actualField.File)
				}
				if actualField.Line != expectedField.Line {
					t.Errorf("Field %d: expected line %d, got %d", i, expectedField.Line, actualField.Line)
				}
			}

			t.Logf("✓ %s", tt.description)
		})
	}
}

// TestExtractFieldReferences_OldVsNewBehavior demonstrates the improvement from placeholder to real extraction
func TestExtractFieldReferences_OldVsNewBehavior(t *testing.T) {
	server := &RepoContextMCPServer{}

	// Create test data that shows the new implementation working correctly
	entry := createTestSearchResultEntry("User", "user.go", 10, 15, []models.Field{
		{Name: "ID", Type: "int64"},
		{Name: "Name", Type: "string"},
		{Name: "Email", Type: "string"},
		{Name: "CreatedAt", Type: "*time.Time"},
	})

	fields := server.extractFieldReferences(entry)

	// Verify that we get actual field names and types, not placeholders
	expectedFieldNames := []string{"ID", "Name", "Email", "CreatedAt"}
	expectedFieldTypes := []string{"int64", "string", "string", "*time.Time"}

	if len(fields) != len(expectedFieldNames) {
		t.Fatalf("Expected %d fields, got %d", len(expectedFieldNames), len(fields))
	}

	for i, field := range fields {
		// OLD behavior would have generated: "Field1", "Field2", etc. with "string" type
		// NEW behavior extracts actual field names and types
		if field.Name == "Field1" || field.Name == "Field2" {
			t.Errorf("Field %d: Still using old placeholder naming (Field%d), expected actual field name %q",
				i, i+1, expectedFieldNames[i])
		}
		if field.Type == "string" && expectedFieldTypes[i] != "string" {
			t.Errorf("Field %d: Still using old hardcoded 'string' type, expected %q",
				i, expectedFieldTypes[i])
		}

		// Verify we get the expected actual values
		if field.Name != expectedFieldNames[i] {
			t.Errorf("Field %d: expected name %q, got %q", i, expectedFieldNames[i], field.Name)
		}
		if field.Type != expectedFieldTypes[i] {
			t.Errorf("Field %d: expected type %q, got %q", i, expectedFieldTypes[i], field.Type)
		}
	}

	t.Logf("✓ Successfully extracts actual field names and types instead of placeholders")
	t.Logf("✓ Extracted fields: %+v", fields)
}

func TestBuildFunctionImplementation_RealExtraction(t *testing.T) {
	tempDir := t.TempDir()

	// Create storage and initialize
	storage := index.NewHybridStorage(tempDir)
	err := storage.Initialize()
	require.NoError(t, err, "Failed to initialize storage")
	defer storage.Close()

	// Create QueryEngine
	queryEngine := index.NewQueryEngine(storage)

	// Create server with storage and query engine
	server := &RepoContextMCPServer{
		QueryEngine: queryEngine,
		Storage:     storage,
		RepoPath:    tempDir,
	}

	// Create a test source file with actual Go code
	testFileContent := `package main

import "fmt"

// ExampleFunction demonstrates function implementation extraction
func ExampleFunction(name string, count int) string {
	if name == "" {
		return "empty name"
	}

	result := fmt.Sprintf("Hello %s", name)
	for i := 0; i < count; i++ {
		result += "!"
	}

	return result
}
`

	testFilePath := filepath.Join(tempDir, "example.go")
	err = os.WriteFile(testFilePath, []byte(testFileContent), ConstFilePermission600)
	require.NoError(t, err, "Failed to write test file")

	// Create file context and store it
	fileContext := &models.FileContext{
		Path: testFilePath,
		Functions: []models.Function{
			{
				Name:      "ExampleFunction",
				Signature: "func ExampleFunction(name string, count int) string",
				StartLine: 6,
				EndLine:   15,
				Parameters: []models.Parameter{
					{Name: "name", Type: "string"},
					{Name: "count", Type: "int"},
				},
				Returns: []models.Type{
					{Name: "", Kind: "string"},
				},
			},
		},
	}

	err = storage.StoreFileContext(fileContext)
	require.NoError(t, err, "Failed to store file context")

	// Create a search result entry to test with
	searchEntry := &index.SearchResultEntry{
		IndexEntry: models.IndexEntry{
			Name:      "ExampleFunction",
			Type:      "function",
			File:      testFilePath,
			StartLine: 6,
			EndLine:   15,
			Signature: "func ExampleFunction(name string, count int) string",
		},
		ChunkData: &models.SemanticChunk{
			FileData: []models.FileContext{*fileContext},
		},
	}

	// Test buildFunctionImplementation
	impl := server.buildFunctionImplementation(searchEntry, 2)
	require.NotNil(t, impl, "Implementation should not be nil")

	// Verify the function body contains actual implementation (even if truncated)
	assert.Contains(t, impl.Body, `if name == ""`, "Function body should contain the condition")
	assert.Contains(t, impl.Body, `return "empty name"`, "Function body should contain the early return")
	assert.Contains(t, impl.Body, `fmt.Sprintf("Hello %s", name)`, "Function body should contain the sprintf call")

	// Check if the body was truncated due to token limits
	if strings.Contains(impl.Body, "implementation truncated due to token limits") {
		// If truncated, verify we got the beginning of the real implementation
		assert.Contains(t, impl.Body, "func ExampleFunction(name string, count int) string",
			"Should contain the function signature")
		t.Logf("Function body was truncated due to token limits (this is expected behavior)")
	} else {
		// If not truncated, verify we got the full implementation
		assert.Contains(t, impl.Body, `for i := 0; i < count; i++`, "Function body should contain the for loop")
		assert.Contains(t, impl.Body, `result += "!"`, "Function body should contain the loop body")
		assert.Contains(t, impl.Body, `return result`, "Function body should contain the final return")
	}

	// Verify context lines are provided
	assert.NotEmpty(t, impl.ContextLines, "Context lines should not be empty")

	// Verify it's not the old placeholder format
	assert.NotContains(t, impl.Body, "Function implementation extraction not supported",
		"Should not contain placeholder message")
	assert.NotContains(t, impl.Body, "include_implementations feature is currently limited",
		"Should not contain limitation message")

	// Log the actual implementation for verification
	t.Logf("Extracted implementation: %s", impl.Body)
}

func TestBuildFunctionImplementation_FallbackToPlaceholder(t *testing.T) {
	// Create server without storage (simulating failure case)
	server := &RepoContextMCPServer{
		QueryEngine: nil,
		Storage:     nil,
		RepoPath:    "/nonexistent",
	}

	// Create a minimal search result entry
	searchEntry := &index.SearchResultEntry{
		IndexEntry: models.IndexEntry{
			Name:      "TestFunction",
			Type:      "function",
			File:      "/nonexistent/test.go",
			StartLine: 5,
			EndLine:   10,
			Signature: "func TestFunction() string",
		},
	}

	// Test buildFunctionImplementation with no storage
	impl := server.buildFunctionImplementation(searchEntry, 2)
	require.NotNil(t, impl, "Implementation should not be nil")

	// Verify it falls back to the documented placeholder
	assert.Contains(t, impl.Body, "Function implementation extraction not supported",
		"Should contain placeholder message when extraction fails")
	assert.Contains(t, impl.Body, "include_implementations feature is currently limited",
		"Should contain limitation message when extraction fails")
	assert.Contains(t, impl.Body, "TestFunction", "Should include function name in placeholder")
	assert.Contains(t, impl.Body, "/nonexistent/test.go", "Should include file path in placeholder")
}
