package index

import (
	"regexp"
	"strings"
	"testing"
)

func TestConvertUnsupportedRegexFeatures(t *testing.T) {
	qe := &QueryEngine{
		regexCache: make(map[string]*regexp.Regexp),
	}

	tests := []struct {
		name       string
		pattern    string
		expected   string
		hasWarning bool
	}{
		{
			name:       "No unsupported features",
			pattern:    "simple.*pattern",
			expected:   "simple.*pattern",
			hasWarning: false,
		},
		{
			name:       "Negative lookbehind removed",
			pattern:    "(?<!Test).*Handler",
			expected:   ".*Handler",
			hasWarning: true,
		},
		{
			name:       "Positive lookahead converted",
			pattern:    "Handle(?=User)",
			expected:   "Handle.*User",
			hasWarning: true,
		},
		{
			name:       "Negative lookahead removed",
			pattern:    "Handle(?!Error)",
			expected:   "Handle",
			hasWarning: true,
		},
		{
			name:       "Multiple unsupported features",
			pattern:    "(?<!Test)Handle(?=User)(?!Error)",
			expected:   "Handle.*User",
			hasWarning: true,
		},
		{
			name:       "Complex pattern with lookbehind",
			pattern:    "func.*(?<!_test).*Handler",
			expected:   "func.*.*Handler",
			hasWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := qe.convertUnsupportedRegexFeatures(tt.pattern)
			if result != tt.expected {
				t.Errorf("convertUnsupportedRegexFeatures() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertUnsupportedRegexFeaturesWithError(t *testing.T) {
	qe := &QueryEngine{
		regexCache: make(map[string]*regexp.Regexp),
	}

	tests := []struct {
		name          string
		pattern       string
		strictMode    bool
		expected      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "No unsupported features - strict mode",
			pattern:     "simple.*pattern",
			strictMode:  true,
			expected:    "simple.*pattern",
			expectError: false,
		},
		{
			name:        "No unsupported features - non-strict mode",
			pattern:     "simple.*pattern",
			strictMode:  false,
			expected:    "simple.*pattern",
			expectError: false,
		},
		{
			name:          "Negative lookbehind - strict mode",
			pattern:       "(?<!Test).*Handler",
			strictMode:    true,
			expected:      "",
			expectError:   true,
			errorContains: "negative lookbehind",
		},
		{
			name:        "Negative lookbehind - non-strict mode",
			pattern:     "(?<!Test).*Handler",
			strictMode:  false,
			expected:    ".*Handler",
			expectError: false,
		},
		{
			name:          "Positive lookahead - strict mode",
			pattern:       "Handle(?=User)",
			strictMode:    true,
			expected:      "",
			expectError:   true,
			errorContains: "positive lookahead",
		},
		{
			name:        "Positive lookahead - non-strict mode",
			pattern:     "Handle(?=User)",
			strictMode:  false,
			expected:    "Handle.*User",
			expectError: false,
		},
		{
			name:          "Negative lookahead - strict mode",
			pattern:       "Handle(?!Error)",
			strictMode:    true,
			expected:      "",
			expectError:   true,
			errorContains: "negative lookahead",
		},
		{
			name:        "Negative lookahead - non-strict mode",
			pattern:     "Handle(?!Error)",
			strictMode:  false,
			expected:    "Handle",
			expectError: false,
		},
		{
			name:          "Multiple features - strict mode",
			pattern:       "(?<!Test)Handle(?=User)(?!Error)",
			strictMode:    true,
			expected:      "",
			expectError:   true,
			errorContains: "unsupported regex features",
		},
		{
			name:        "Multiple features - non-strict mode",
			pattern:     "(?<!Test)Handle(?=User)(?!Error)",
			strictMode:  false,
			expected:    "Handle.*User",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := qe.convertUnsupportedRegexFeaturesWithError(tt.pattern, tt.strictMode)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Error should contain '%s', got: %s", tt.errorContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("convertUnsupportedRegexFeaturesWithError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetCompiledRegexWithUnsupportedFeatures(t *testing.T) {
	qe := &QueryEngine{
		regexCache: make(map[string]*regexp.Regexp),
	}

	// Test that the integration works with actual regex compilation
	tests := []struct {
		name        string
		pattern     string
		expectError bool
	}{
		{
			name:        "Simple pattern works",
			pattern:     "simple.*pattern",
			expectError: false,
		},
		{
			name:        "Converted lookbehind pattern compiles",
			pattern:     "(?<!Test).*Handler",
			expectError: false,
		},
		{
			name:        "Converted lookahead pattern compiles",
			pattern:     "Handle(?=User)",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := qe.getCompiledRegex(tt.pattern)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if regex == nil {
				t.Errorf("Expected compiled regex but got nil")
			}
		})
	}
}
