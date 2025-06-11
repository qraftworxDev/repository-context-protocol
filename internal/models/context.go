package models

import "time"

// Core data structures
type RepoContext struct {
	Path        string                  `json:"path"`
	Language    string                  `json:"language"`
	LastUpdate  time.Time               `json:"last_update"`
	Files       map[string]*FileContext `json:"files"`
	GlobalIndex *GlobalIndex            `json:"global_index"`
}

type FileContext struct {
	Path      string     `json:"path"`
	Language  string     `json:"language"`
	Checksum  string     `json:"checksum"`
	ModTime   time.Time  `json:"mod_time"`
	Functions []Function `json:"functions"`
	Types     []TypeDef  `json:"types"`
	Variables []Variable `json:"variables"`
	Imports   []Import   `json:"imports"`
	Exports   []Export   `json:"exports"`
}

type GlobalIndex struct {
	ByName    map[string][]Reference `json:"by_name"`
	ByFile    map[string][]Reference `json:"by_file"`
	ByType    map[string][]Reference `json:"by_type"`
	CallGraph map[string][]string    `json:"call_graph"`
	TypeGraph map[string][]string    `json:"type_graph"`
}
