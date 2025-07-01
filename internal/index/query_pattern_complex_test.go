package index

import (
	"os"
	"testing"
)

// TestQueryEngine_ComplexRegexPatternMatching tests advanced regex pattern matching scenarios
func TestQueryEngine_ComplexRegexPatternMatching(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	// Use the setup function from query_pattern_setup_test.go
	SetupComplexPatternTestData(t, storage)
	SetupRegexTestCases(t, storage)

	engine := NewQueryEngine(storage)

	tests := []PatternTest{
		// Advanced regex patterns
		{
			name:    "regex word boundary",
			pattern: `/\bUser\b/`,
			expectedNames: []string{
				"User", // Only exact word "User" should match word boundary
			},
			description: "Should match word boundaries with User (exact word only)",
		},
		{
			name:    "regex negative character class",
			pattern: `/^[^V].*Data$/`,
			expectedNames: []string{
				"ProcessUserData", "ProcessPaymentData", "UserData", "PaymentData",
				"parseJSONData",
			},
			description: "Should match entities not starting with V and ending with Data",
		},
		{
			name:    "regex lookahead simulation",
			pattern: `/Handle.*User/`,
			expectedNames: []string{
				"HandleUserLogin", "HandleUserLogout",
			},
			description: "Should match Handle followed by User",
		},
		{
			name:    "regex capture groups",
			pattern: `/(Handle|Process)(User|Payment)Data/`,
			expectedNames: []string{
				"ProcessUserData", "ProcessPaymentData",
			},
			description: "Should match grouped patterns",
		},
		{
			name:    "regex quantifiers complex",
			pattern: `/^[A-Z][a-z]+[A-Z][a-z]+$/`,
			expectedNames: []string{
				// Only matches exactly 2 CamelCase words: [A-Z][a-z]+[A-Z][a-z]+
				"ValidateCredentials", "ValidatePayment", "QueryUsers", "QueryPayments",
				"UserData", "PaymentData", "UserValidator", "PaymentProcessor",
				"ResponseWriter", "GoString",
			},
			description: "Should match exactly 2 CamelCase words pattern",
		},
		{
			name:    "regex anchors and alternation",
			pattern: `/^(Handle|Process|Validate)/`,
			expectedNames: []string{
				"HandleUserLogin", "HandleUserLogout", "HandleAPIRequest",
				"ProcessUserData", "ProcessPaymentData", "ValidateCredentials",
				"ValidateUserData", "ValidatePayment",
			},
			description: "Should match functions starting with Handle, Process, or Validate",
		},
		// Complex business logic patterns
		{
			name:    "workflow pattern matching",
			pattern: `/(Handle.*|Process.*|Validate.*)/`,
			expectedNames: []string{
				"HandleUserLogin", "HandleUserLogout", "HandleAPIRequest",
				"ProcessUserData", "ProcessPaymentData", "ValidateCredentials",
				"ValidateUserData", "ValidatePayment",
			},
			description: "Should match workflow patterns",
		},
		{
			name:    "data entity patterns",
			pattern: `/.*Data$/`,
			expectedNames: []string{
				"ProcessUserData", "ProcessPaymentData", "UserData", "PaymentData",
				"parseJSONData", "processV1Data", "processV2Data",
			},
			description: "Should match data-related entities",
		},
		{
			name:    "interface patterns",
			pattern: `/.*[A-Z][a-z]+[A-Z][a-z]*$/`,
			expectedNames: []string{
				// Must end with [A-Z][a-z]+[A-Z][a-z]* pattern
				"HandleUserLogin", "HandleUserLogout", "ProcessUserData",
				"ProcessPaymentData", "ValidateCredentials", "ValidateUserData",
				"ValidatePayment", "QueryUsers", "QueryPayments",
				"UserData", "PaymentData", "UserValidator",
				"PaymentProcessor", "ResponseWriter", "GoString",
			},
			description: "Should match interface-style naming ending with CamelCase",
		},
		{
			name:    "special character handling",
			pattern: `/.*_.*_.*$/`,
			expectedNames: []string{
				"handle_legacy_format", "parse_config_file",
			},
			description: "Should match snake_case with underscores",
		},
		// Performance-oriented patterns
		{
			name:    "efficient prefix matching",
			pattern: `/^Query/`,
			expectedNames: []string{
				"QueryUsers", "QueryPayments",
			},
			description: "Should efficiently match Query prefix",
		},
		{
			name:    "efficient suffix matching",
			pattern: `/JSON$/`,
			expectedNames: []string{
				"MarshalJSON", "UnmarshalJSON",
			},
			description: "Should efficiently match JSON suffix",
		},
		{
			name:    "mid-string patterns",
			pattern: `/.*API.*/`,
			expectedNames: []string{
				"HandleAPIRequest", "APIResponse",
			},
			description: "Should match patterns containing API",
		},
	}

	// Use the common test verification helper from query_pattern_test.go
	verifyPatternMatchesHelper(t, engine, tests)
}

// TestQueryEngine_RegexCacheEfficiency tests regex compilation caching
func TestQueryEngine_RegexCacheEfficiency(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	SetupComplexPatternTestData(t, storage)

	engine := NewQueryEngine(storage)

	pattern := `/^Handle.*User.*/`

	// Enable types in search for consistent results
	options := QueryOptions{
		IncludeTypes: true,
	}

	// First search should compile the regex
	results1, err := engine.SearchByPatternWithOptions(pattern, options)
	if err != nil {
		t.Fatalf("First search failed: %v", err)
	}

	// Second search should use cached regex
	results2, err := engine.SearchByPatternWithOptions(pattern, options)
	if err != nil {
		t.Fatalf("Second search failed: %v", err)
	}

	// Results should be identical
	if len(results1.Entries) != len(results2.Entries) {
		t.Errorf("Cache inconsistency: first search found %d entries, second found %d",
			len(results1.Entries), len(results2.Entries))
	}

	// Verify cache contains the pattern
	engine.regexMutex.RLock()
	_, exists := engine.regexCache["^Handle.*User.*"]
	engine.regexMutex.RUnlock()

	if !exists {
		t.Error("Expected compiled regex to be cached")
	}
}

// TestQueryEngine_RegexDelimiterHandling tests regex pattern delimiter handling
func TestQueryEngine_RegexDelimiterHandling(t *testing.T) {
	tempDir, storage := setupTestStorage(t)
	defer os.RemoveAll(tempDir)

	SetupComplexPatternTestData(t, storage)

	engine := NewQueryEngine(storage)

	tests := []struct {
		name             string
		delimitedPattern string
		rawPattern       string
		description      string
	}{
		{
			name:             "forward slash delimiters",
			delimitedPattern: `/^Handle/`,
			rawPattern:       `^Handle`,
			description:      "Should handle forward slash regex delimiters",
		},
		{
			name:             "complex delimited pattern",
			delimitedPattern: `/^(Handle|Process).*User/`,
			rawPattern:       `^(Handle|Process).*User`,
			description:      "Should handle complex delimited patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Enable types in search for consistent results
			options := QueryOptions{
				IncludeTypes: true,
			}

			// Search with delimited pattern
			delimitedResults, err := engine.SearchByPatternWithOptions(tt.delimitedPattern, options)
			if err != nil {
				t.Fatalf("Delimited pattern search failed: %v", err)
			}

			// Search with raw pattern
			rawResults, err := engine.SearchByPatternWithOptions(tt.rawPattern, options)
			if err != nil {
				t.Fatalf("Raw pattern search failed: %v", err)
			}

			// Results should be identical
			if len(delimitedResults.Entries) != len(rawResults.Entries) {
				t.Errorf("Delimiter handling inconsistency: delimited found %d, raw found %d",
					len(delimitedResults.Entries), len(rawResults.Entries))
			}

			// Verify same entries are found
			delimitedNames := make(map[string]bool)
			for _, entry := range delimitedResults.Entries {
				delimitedNames[entry.IndexEntry.Name] = true
			}

			for _, entry := range rawResults.Entries {
				if !delimitedNames[entry.IndexEntry.Name] {
					t.Errorf("Entry %s found in raw pattern but not in delimited pattern", entry.IndexEntry.Name)
				}
			}
		})
	}
}
