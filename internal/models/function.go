package models

import "slices"

// Function representation and metadata
type Function struct {
	Name       string      `json:"name"`
	Signature  string      `json:"signature"`
	Parameters []Parameter `json:"parameters"`
	Returns    []Type      `json:"returns"`
	StartLine  int         `json:"start_line"`
	EndLine    int         `json:"end_line"`

	// Deprecated: Will be removed in v2.0 - Use LocalCalls + CrossFileCalls instead
	Calls    []string `json:"calls,omitempty"`     // All function calls (local + cross-file)
	CalledBy []string `json:"called_by,omitempty"` // All callers (local + cross-file)

	// PRIMARY FIELDS: Enhanced with consistent metadata
	LocalCalls       []string        `json:"local_calls"`        // Same-file calls only
	CrossFileCalls   []CallReference `json:"cross_file_calls"`   // Cross-file calls with metadata
	LocalCallers     []string        `json:"local_callers"`      // Same-file callers only
	CrossFileCallers []CallReference `json:"cross_file_callers"` // Cross-file callers with metadata

	// INTERNAL FIELD: Used during parsing to preserve call metadata before enrichment
	LocalCallsWithMetadata []CallReference `json:"-"` // Local calls with metadata (not serialized)
}

// CallReference represents a call relationship with additional metadata
type CallReference struct {
	FunctionName string `json:"function_name"`       // Name of the called/calling function
	File         string `json:"file"`                // File where the function is defined
	Line         int    `json:"line,omitempty"`      // Line number where the call occurs
	CallType     string `json:"call_type,omitempty"` // "function", "method", "external"
}

// CallType constants for consistent classification
const (
	CallTypeFunction = "function" // Regular function call
	CallTypeMethod   = "method"   // Method call on an object
	CallTypeExternal = "external" // External library/package call
	CallTypeComplex  = "complex"  // Complex expression call (e.g., function pointer)
)

// IsValid validates the CallReference fields
func (cr CallReference) IsValid() bool {
	if cr.FunctionName == "" {
		return false
	}
	if cr.File == "" {
		return false
	}
	// CallType is optional but if provided, should be valid
	if cr.CallType != "" {
		return cr.CallType == CallTypeFunction ||
			cr.CallType == CallTypeMethod ||
			cr.CallType == CallTypeExternal ||
			cr.CallType == CallTypeComplex
	}
	return true
}

type Parameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Type struct {
	Name string `json:"name"`
	Kind string `json:"kind"` // "basic", "struct", "interface", etc.
}

type Variable struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type Constant struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Value     string `json:"value,omitempty"` // Optional: the constant value
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type Import struct {
	Path  string `json:"path"`
	Alias string `json:"alias,omitempty"`
}

type Export struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Kind string `json:"kind"` // "function", "type", "variable", "constant"
}

// Backward compatibility helpers for Function

// GetAllCalls returns all function calls (local + cross-file) for backward compatibility
func (f *Function) GetAllCalls() []string {
	// If deprecated field is populated, use it for backward compatibility
	if len(f.Calls) > 0 {
		return f.Calls
	}

	// Build from enhanced fields
	var allCalls []string
	allCalls = append(allCalls, f.LocalCalls...)

	for _, call := range f.CrossFileCalls {
		allCalls = append(allCalls, call.FunctionName)
	}

	return allCalls
}

// GetAllCallers returns all function callers (local + cross-file) for backward compatibility
func (f *Function) GetAllCallers() []string {
	// If deprecated field is populated, use it for backward compatibility
	if len(f.CalledBy) > 0 {
		return f.CalledBy
	}

	// Build from enhanced fields
	var allCallers []string
	allCallers = append(allCallers, f.LocalCallers...)

	for _, caller := range f.CrossFileCallers {
		allCallers = append(allCallers, caller.FunctionName)
	}

	return allCallers
}

// GetCallsInFile returns all calls to functions in a specific file
func (f *Function) GetCallsInFile(filePath string) []CallReference {
	var calls []CallReference

	for _, call := range f.CrossFileCalls {
		if call.File == filePath {
			calls = append(calls, call)
		}
	}

	return calls
}

// GetCallersFromFile returns all callers from a specific file
func (f *Function) GetCallersFromFile(filePath string) []CallReference {
	var callers []CallReference

	for _, caller := range f.CrossFileCallers {
		if caller.File == filePath {
			callers = append(callers, caller)
		}
	}

	return callers
}

// HasCall checks if the function calls a specific function by name
func (f *Function) HasCall(functionName string) bool {
	// Check local calls
	if slices.Contains(f.LocalCalls, functionName) {
		return true
	}

	// Check cross-file calls
	for _, call := range f.CrossFileCalls {
		if call.FunctionName == functionName {
			return true
		}
	}

	return false
}

// HasCaller checks if the function is called by a specific function
func (f *Function) HasCaller(functionName string) bool {
	// Check local callers
	if slices.Contains(f.LocalCallers, functionName) {
		return true
	}

	// Check cross-file callers
	for _, caller := range f.CrossFileCallers {
		if caller.FunctionName == functionName {
			return true
		}
	}

	return false
}
