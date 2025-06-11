package models

// Function representation and metadata
type Function struct {
	Name       string      `json:"name"`
	Signature  string      `json:"signature"`
	Parameters []Parameter `json:"parameters"`
	Returns    []Type      `json:"returns"`
	StartLine  int         `json:"start_line"`
	EndLine    int         `json:"end_line"`
	Calls      []string    `json:"calls"`     // All function calls (local + cross-file)
	CalledBy   []string    `json:"called_by"` // All callers (local + cross-file)
	// Enhanced global relationship tracking
	LocalCalls       []string        `json:"local_calls"`        // Same-file calls only
	CrossFileCalls   []CallReference `json:"cross_file_calls"`   // Cross-file calls with metadata
	LocalCallers     []string        `json:"local_callers"`      // Same-file callers only
	CrossFileCallers []CallReference `json:"cross_file_callers"` // Cross-file callers with metadata
}

// CallReference represents a call relationship with additional metadata
type CallReference struct {
	FunctionName string `json:"function_name"`  // Name of the called/calling function
	File         string `json:"file"`           // File where the function is defined
	Line         int    `json:"line,omitempty"` // Line number where the call occurs
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
