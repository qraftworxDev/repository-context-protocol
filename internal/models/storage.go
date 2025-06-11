package models

// IndexEntry represents a lightweight index entry stored in SQLite for fast lookups
type IndexEntry struct {
	Name      string `json:"name"`       // Name of the entity (function, type, variable, etc.)
	Type      string `json:"type"`       // "function", "type", "variable", "constant"
	File      string `json:"file"`       // File path where entity is defined
	StartLine int    `json:"start_line"` // Starting line number
	EndLine   int    `json:"end_line"`   // Ending line number
	ChunkID   string `json:"chunk_id"`   // ID of the MessagePack chunk containing detailed data
	Signature string `json:"signature"`  // Function signature, type definition, etc.
}

// CallRelation represents a function call relationship stored in SQLite
type CallRelation struct {
	Caller     string `json:"caller"`      // Name of the calling function
	Callee     string `json:"callee"`      // Name of the called function
	File       string `json:"file"`        // File where the call occurs
	Line       int    `json:"line"`        // Line number of the call
	CallerFile string `json:"caller_file"` // File where the caller function is defined
}
