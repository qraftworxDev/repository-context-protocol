package models

// Cross-references between code entities
type Reference struct {
	File      string `json:"file"`
	Entity    string `json:"entity"`
	Type      string `json:"type"` // "function", "type", "variable", "import"
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Signature string `json:"signature,omitempty"`
}
