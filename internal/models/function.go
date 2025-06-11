package models

// Function representation and metadata
type Function struct {
	Name       string      `json:"name"`
	Signature  string      `json:"signature"`
	Parameters []Parameter `json:"parameters"`
	Returns    []Type      `json:"returns"`
	StartLine  int         `json:"start_line"`
	EndLine    int         `json:"end_line"`
	Calls      []string    `json:"calls"`     // Functions this calls
	CalledBy   []string    `json:"called_by"` // Functions that call this
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
