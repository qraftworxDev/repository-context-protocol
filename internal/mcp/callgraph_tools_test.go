package mcp

import (
	"testing"
)

// TestEnhancedCallGraphTools_Registration tests tool registration
func TestEnhancedCallGraphTools_Registration(t *testing.T) {
	server := NewRepoContextMCPServer()

	tools := server.RegisterCallGraphTools()

	expectedTools := []string{
		"get_call_graph_enhanced",
		"find_dependencies",
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

// TestEnhancedCallGraphDepthValidation tests depth control validation
func TestEnhancedCallGraphDepthValidation(t *testing.T) {
	tests := []struct {
		name          string
		inputDepth    int
		expectedDepth int
	}{
		{"Valid depth 1", 1, 1},
		{"Valid depth 5", 5, 5},
		{"Max depth 10", 10, 10},
		{"Exceed max depth - should cap at 10", 15, 10},
		{"Zero depth - should default to 2", 0, 2},
		{"Negative depth - should default to 2", -1, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateEnhancedCallGraphDepth(tt.inputDepth)
			if result != tt.expectedDepth {
				t.Errorf("Expected depth %d, got %d", tt.expectedDepth, result)
			}
		})
	}
}

// TestDependencyTypeValidation tests dependency type validation
func TestDependencyTypeValidation(t *testing.T) {
	tests := []struct {
		name           string
		dependencyType string
		expectError    bool
	}{
		{"Valid type: callers", "callers", false},
		{"Valid type: callees", "callees", false},
		{"Valid type: both", "both", false},
		{"Valid type: empty (defaults to both)", "", false},
		{"Valid type: mixed case", "CALLERS", false},
		{"Invalid type", "invalid", true},
		{"Invalid type: random string", "random", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDependencyType(tt.dependencyType)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestExternalCallDetection tests external call detection logic
func TestExternalCallDetection(t *testing.T) {
	server := NewRepoContextMCPServer()

	tests := []struct {
		name         string
		functionName string
		isExternal   bool
	}{
		{"Standard library - fmt", "fmt.Println", true},
		{"Standard library - os", "os.Open", true},
		{"External package - github", "github.com/user/pkg.Function", true},
		{"External prefix", "external:SomeFunction", true},
		{"Internal function", "MyFunction", false},
		{"Internal package function", "mypackage.MyFunction", false},
		{"Golang.org package", "golang.org/x/tools.Parse", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.isExternalCall(tt.functionName)
			if result != tt.isExternal {
				t.Errorf("Expected isExternal=%v for '%s', got %v", tt.isExternal, tt.functionName, result)
			}
		})
	}
}

// TestTokenMeasurement tests empirical token measurement functionality
func TestTokenMeasurement(t *testing.T) {
	server := NewRepoContextMCPServer()

	tests := []struct {
		name            string
		input           interface{}
		expectMinTokens int
		expectMaxTokens int
	}{
		{
			name: "Simple struct",
			input: struct {
				Name string `json:"name"`
				ID   int    `json:"id"`
			}{Name: "test", ID: 123},
			expectMinTokens: 5,  // Minimum reasonable token count
			expectMaxTokens: 20, // Maximum reasonable token count
		},
		{
			name:            "Empty struct",
			input:           struct{}{},
			expectMinTokens: 1, // At least {}
			expectMaxTokens: 5, // JSON overhead
		},
		{
			name: "Complex nested structure",
			input: map[string]interface{}{
				"entity_name": "TestFunction",
				"callers": []map[string]interface{}{
					{"function": "Caller1", "file": "test.go", "line": 10},
					{"function": "Caller2", "file": "test.go", "line": 20},
				},
				"callees": []map[string]interface{}{
					{"function": "Callee1", "file": "helper.go", "line": 5},
				},
				"truncated": false,
			},
			expectMinTokens: 20, // Reasonable minimum for this structure
			expectMaxTokens: 50, // Reasonable maximum
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualTokens := server.measureActualTokens(tt.input)

			if actualTokens < tt.expectMinTokens {
				t.Errorf("Token count %d is below expected minimum %d", actualTokens, tt.expectMinTokens)
			}

			if actualTokens > tt.expectMaxTokens {
				t.Errorf("Token count %d exceeds expected maximum %d", actualTokens, tt.expectMaxTokens)
			}
		})
	}
}

// TestTokenEstimateValidation tests token estimation accuracy validation
func TestTokenEstimateValidation(t *testing.T) {
	server := NewRepoContextMCPServer()

	testData := struct {
		Name      string   `json:"name"`
		Functions []string `json:"functions"`
	}{
		Name:      "TestData",
		Functions: []string{"func1", "func2", "func3"},
	}

	// Test estimate validation
	estimated := 15 // Rough estimate
	actualTokens, accuracy := server.validateTokenEstimate(estimated, testData)

	if actualTokens <= 0 {
		t.Error("Actual token count should be positive")
	}

	if accuracy <= 0 {
		t.Error("Accuracy should be positive")
	}

	// Test optimization with actual measurement
	canOptimize := server.optimizeWithActualMeasurement(testData, 100)
	if !canOptimize {
		t.Error("Should be able to optimize with 100 token limit for simple test data")
	}

	canOptimizeTight := server.optimizeWithActualMeasurement(testData, 1)
	if canOptimizeTight {
		t.Error("Should not be able to optimize with 1 token limit")
	}
}

// TestTokenConstants tests that our token constants are reasonable
func TestTokenConstants(t *testing.T) {
	// Test that constants are within reasonable ranges
	tests := []struct {
		name     string
		value    int
		minValue int
		maxValue int
	}{
		{"JSONMetadataReserveTokens", JSONMetadataReserveTokens, 50, 200},
		{"FunctionEntryOverheadTokens", FunctionEntryOverheadTokens, 10, 50},
		{"SearchEntryOverheadTokens", SearchEntryOverheadTokens, 15, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value < tt.minValue {
				t.Errorf("%s value %d is below reasonable minimum %d", tt.name, tt.value, tt.minValue)
			}
			if tt.value > tt.maxValue {
				t.Errorf("%s value %d exceeds reasonable maximum %d", tt.name, tt.value, tt.maxValue)
			}
		})
	}

	// Test token ratios
	ratioTests := []struct {
		name     string
		value    float64
		minValue float64
		maxValue float64
	}{
		{"CallGraphCalleeTokenRatio", CallGraphCalleeTokenRatio, 0.3, 0.7},
		{"DependencyCalleeTokenRatio", DependencyCalleeTokenRatio, 0.2, 0.5},
		{"RemainingTokenSplitRatio", RemainingTokenSplitRatio, 0.3, 0.7},
	}

	for _, tt := range ratioTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value < tt.minValue {
				t.Errorf("%s ratio %f is below reasonable minimum %f", tt.name, tt.value, tt.minValue)
			}
			if tt.value > tt.maxValue {
				t.Errorf("%s ratio %f exceeds reasonable maximum %f", tt.name, tt.value, tt.maxValue)
			}
		})
	}
}
