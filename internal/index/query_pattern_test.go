package index

import (
	"os"
	"testing"
	"time"

	"repository-context-protocol/internal/models"
)

// TestQueryEngine_SimplePatternMatching tests basic pattern matching functionality
func TestQueryEngine_SimplePatternMatching(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataForPatterns(t, storage)

	engine := NewQueryEngine(storage)

	tests := []PatternTest{
		{
			name:          "exact match",
			pattern:       "TestFunction",
			expectedNames: []string{"TestFunction"},
			description:   "Should match exact function name",
		},
		{
			name:          "prefix wildcard",
			pattern:       "Test*",
			expectedNames: []string{"TestFunction", "TestStruct", "TestVar"},
			description:   "Should match all entities starting with Test",
		},
		{
			name:          "suffix wildcard",
			pattern:       "*Function",
			expectedNames: []string{"TestFunction", "AnotherFunction", "HelperFunction"},
			description:   "Should match all entities ending with Function",
		},
		{
			name:          "middle wildcard",
			pattern:       "Test*Function",
			expectedNames: []string{"TestFunction"},
			description:   "Should match Test followed by anything then Function",
		},
		{
			name:          "single character wildcard",
			pattern:       "Test?ar",
			expectedNames: []string{"TestVar"},
			description:   "Should match Test + single char + ar",
		},
	}

	// Use the common verification helper
	verifyPatternMatchesHelper(t, engine, tests)
}

// TestQueryEngine_GlobPatternMatching tests shell-style glob patterns
func TestQueryEngine_GlobPatternMatching(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataForPatterns(t, storage)

	engine := NewQueryEngine(storage)

	tests := []PatternTest{
		{
			name:          "multiple wildcards",
			pattern:       "*Test*",
			expectedNames: []string{"TestFunction", "TestStruct", "TestVar"},
			description:   "Should match anything containing Test",
		},
		{
			name:          "question mark wildcard",
			pattern:       "?est*",
			expectedNames: []string{"TestFunction", "TestStruct", "TestVar"},
			description:   "Should match single char + est + anything",
		},
		{
			name:          "character class",
			pattern:       "Test[SV]*",
			expectedNames: []string{"TestStruct", "TestVar"},
			description:   "Should match character class patterns",
		},
	}

	// Use the common verification helper
	verifyPatternMatchesHelper(t, engine, tests)
}

// TestQueryEngine_RegexPatternMatching tests specific regex pattern functionality
func TestQueryEngine_RegexPatternMatching(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataForPatterns(t, storage)

	engine := NewQueryEngine(storage)

	tests := []PatternTest{
		{
			name:          "regex anchored start",
			pattern:       "/^Test/",
			expectedNames: []string{"TestFunction", "TestStruct", "TestVar"},
			description:   "Should match entities starting with 'Test'",
		},
		{
			name:          "regex anchored end",
			pattern:       "/Function$/",
			expectedNames: []string{"TestFunction", "AnotherFunction", "HelperFunction"},
			description:   "Should match entities ending with 'Function'",
		},
		{
			name:          "regex alternation",
			pattern:       "/(Test|Another)/",
			expectedNames: []string{"TestFunction", "TestStruct", "TestVar", "AnotherFunction"},
			description:   "Should match regex alternation",
		},
		{
			name:          "regex character class",
			pattern:       "/Test[SV]/",
			expectedNames: []string{"TestStruct", "TestVar"},
			description:   "Should match regex character class",
		},
		{
			name:          "regex quantifiers",
			pattern:       "/Test.+/",
			expectedNames: []string{"TestFunction", "TestStruct", "TestVar"},
			description:   "Should match regex quantifiers",
		},
	}

	// Use the common verification helper
	verifyPatternMatchesHelper(t, engine, tests)
}

// TestQueryEngine_PatternDetection tests the pattern type detection logic
func TestQueryEngine_PatternDetection(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	engine := NewQueryEngine(storage)

	tests := []struct {
		pattern     string
		isRegex     bool
		description string
	}{
		// Glob patterns
		{"*", false, "Simple wildcard"},
		{"Test*", false, "Prefix wildcard"},
		{"*Function", false, "Suffix wildcard"},
		{"Test?ar", false, "Single character wildcard"},
		{"Test[SV]*", false, "Character class glob"},

		// Regex patterns (delimited)
		{"/^Test/", true, "Regex with delimiters"},
		{"/Test$/", true, "Regex with delimiters"},
		{"/Test.+/", true, "Regex with quantifiers"},

		// Regex patterns (by metacharacters)
		{"^Test", true, "Regex anchor start"},
		{"Test$", true, "Regex anchor end"},
		{"Test+", true, "Regex quantifier"},
		{"Test|Another", true, "Regex alternation"},
		{"Test(ing)?", true, "Regex grouping"},
		{"Test{1,3}", true, "Regex quantifier braces"},
		{"Test\\w+", true, "Regex escape"},

		// Edge cases
		{"", false, "Empty pattern"},
		{"Test", false, "Simple string"},
		{"Test.go", false, "Filename with dot"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			isRegex := engine.isRegexPattern(tt.pattern)
			if isRegex != tt.isRegex {
				t.Errorf("Pattern '%s': expected isRegex=%v, got isRegex=%v",
					tt.pattern, tt.isRegex, isRegex)
			}
		})
	}
}

// TestQueryEngine_PatternMatchingWithOptions tests pattern matching with query options
func TestQueryEngine_PatternMatchingWithOptions(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataWithCallGraph(t, storage)

	engine := NewQueryEngine(storage)

	// Test with call graph options
	options := QueryOptions{
		IncludeCallers: true,
		IncludeCallees: true,
		MaxDepth:       2,
		MaxTokens:      1000,
	}

	results, err := engine.SearchByPatternWithOptions("*Function", options)
	if err != nil {
		t.Fatalf("Failed to search by pattern with options: %v", err)
	}

	if len(results.Entries) == 0 {
		t.Error("Expected to find functions matching *Function pattern")
	}

	// Should include call graph for function results
	if results.CallGraph == nil {
		t.Error("Expected call graph to be included in results")
	}

	// Verify options are preserved
	if results.Options.IncludeCallers != options.IncludeCallers {
		t.Error("Expected IncludeCallers option to be preserved")
	}
	if results.Options.IncludeCallees != options.IncludeCallees {
		t.Error("Expected IncludeCallees option to be preserved")
	}
}

// TestQueryEngine_PatternMatchingWithTokenLimit tests pattern matching with token limits
func TestQueryEngine_PatternMatchingWithTokenLimit(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	setupTestDataForPatterns(t, storage)

	engine := NewQueryEngine(storage)

	// Set a low token limit to test truncation
	options := QueryOptions{
		MaxTokens: 50,
	}

	results, err := engine.SearchByPatternWithOptions("*", options)
	if err != nil {
		t.Fatalf("Failed to search with token limit: %v", err)
	}

	// Should have limited results due to token constraint
	if results.TokenCount > 50 {
		t.Errorf("Expected token count <= 50, got %d", results.TokenCount)
	}

	// The test may result in 0 entries with a very low token limit, which is acceptable
	// The important thing is that the token counting is working correctly
	t.Logf("Token limit test: %d entries returned with token count %d (truncated: %v)",
		len(results.Entries), results.TokenCount, results.Truncated)
}

// PatternTest represents a test case for pattern matching
type PatternTest struct {
	name          string
	pattern       string
	expectedNames []string
	description   string
}

// verifyPatternMatchesHelper is a helper function to verify pattern matching results
func verifyPatternMatchesHelper(t *testing.T, engine *QueryEngine, tests []PatternTest) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Enable types in search since many tests expect to find type names
			options := QueryOptions{
				IncludeTypes: true,
			}
			results, err := engine.SearchByPatternWithOptions(tt.pattern, options)
			if err != nil {
				t.Fatalf("Failed to search by pattern '%s': %v", tt.pattern, err)
			}

			// Collect actual names
			actualNames := make([]string, len(results.Entries))
			for i, entry := range results.Entries {
				actualNames[i] = entry.IndexEntry.Name
			}

			// Verify all expected names are found
			for _, expectedName := range tt.expectedNames {
				found := false
				for _, actualName := range actualNames {
					if actualName == expectedName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Pattern '%s': expected to find '%s', but it was not in results %v",
						tt.pattern, expectedName, actualNames)
				}
			}
		})
	}
}

// setupTestDataForPatterns creates test data with diverse naming patterns
func setupTestDataForPatterns(t *testing.T, storage *HybridStorage) {
	fileContext := &models.FileContext{
		Path:     "patterns.go",
		Language: "go",
		Checksum: "patterns123",
		ModTime:  time.Now(),
		Functions: []models.Function{
			{
				Name:      "TestFunction",
				Signature: "func TestFunction()",
				StartLine: 10,
				EndLine:   15,
			},
			{
				Name:      "AnotherFunction",
				Signature: "func AnotherFunction(s string) int",
				StartLine: 20,
				EndLine:   25,
			},
			{
				Name:      "HelperFunction",
				Signature: "func HelperFunction() string",
				StartLine: 30,
				EndLine:   35,
			},
			{
				Name:      "ProcessData",
				Signature: "func ProcessData(data []byte) error",
				StartLine: 40,
				EndLine:   45,
			},
			{
				Name:      "validateInput",
				Signature: "func validateInput(input string) bool",
				StartLine: 50,
				EndLine:   55,
			},
		},
		Types: []models.TypeDef{
			{
				Name:      "TestStruct",
				Kind:      "struct",
				StartLine: 5,
				EndLine:   8,
			},
			{
				Name:      "DataProcessor",
				Kind:      "struct",
				StartLine: 60,
				EndLine:   65,
			},
			{
				Name:      "Validator",
				Kind:      "interface",
				StartLine: 70,
				EndLine:   75,
			},
		},
		Variables: []models.Variable{
			{
				Name:      "TestVar",
				Type:      "string",
				StartLine: 3,
				EndLine:   3,
			},
			{
				Name:      "DefaultConfig",
				Type:      "Config",
				StartLine: 80,
				EndLine:   80,
			},
		},
		Constants: []models.Constant{
			{
				Name:      "MaxRetries",
				Type:      "int",
				Value:     "3",
				StartLine: 85,
				EndLine:   85,
			},
			{
				Name:      "TimeoutSeconds",
				Type:      "int",
				Value:     "30",
				StartLine: 86,
				EndLine:   86,
			},
		},
	}

	err := storage.StoreFileContext(fileContext)
	if err != nil {
		t.Fatalf("Failed to store pattern test data: %v", err)
	}
}
