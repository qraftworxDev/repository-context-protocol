package models

// Type definitions and relationships
type TypeDef struct {
	Name      string   `json:"name"`
	Kind      string   `json:"kind"` // "struct", "interface", "alias", "basic"
	Fields    []Field  `json:"fields,omitempty"`
	Methods   []Method `json:"methods,omitempty"`
	StartLine int      `json:"start_line"`
	EndLine   int      `json:"end_line"`
	Embedded  []string `json:"embedded,omitempty"` // Embedded types
}

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Tag  string `json:"tag,omitempty"`
}

type Method struct {
	Name       string      `json:"name"`
	Signature  string      `json:"signature"`
	Parameters []Parameter `json:"parameters"`
	Returns    []Type      `json:"returns"`
	StartLine  int         `json:"start_line"`
	EndLine    int         `json:"end_line"`
}
