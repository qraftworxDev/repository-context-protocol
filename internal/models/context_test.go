package models

import (
	"testing"
	"time"
)

func TestRepoContext_Creation(t *testing.T) {
	repo := &RepoContext{
		Path:       "/path/to/repo",
		Language:   "go",
		LastUpdate: time.Now(),
		Files:      make(map[string]*FileContext),
		GlobalIndex: &GlobalIndex{
			ByName:    make(map[string][]Reference),
			ByFile:    make(map[string][]Reference),
			ByType:    make(map[string][]Reference),
			CallGraph: make(map[string][]string),
			TypeGraph: make(map[string][]string),
		},
	}

	if repo.Path != "/path/to/repo" {
		t.Errorf("Expected path '/path/to/repo', got %s", repo.Path)
	}
	if repo.Language != "go" {
		t.Errorf("Expected language 'go', got %s", repo.Language)
	}
	if repo.Files == nil {
		t.Error("Files map should be initialized")
	}
	if repo.GlobalIndex == nil {
		t.Error("GlobalIndex should be initialized")
	}
}

func TestFileContext_Creation(t *testing.T) {
	modTime := time.Now()
	file := &FileContext{
		Path:      "main.go",
		Language:  "go",
		Checksum:  "abc123",
		ModTime:   modTime,
		Functions: []Function{},
		Types:     []TypeDef{},
		Variables: []Variable{},
		Imports:   []Import{},
		Exports:   []Export{},
	}

	if file.Path != "main.go" {
		t.Errorf("Expected path 'main.go', got %s", file.Path)
	}
	if file.Language != "go" {
		t.Errorf("Expected language 'go', got %s", file.Language)
	}
	if file.Checksum != "abc123" {
		t.Errorf("Expected checksum 'abc123', got %s", file.Checksum)
	}
	if !file.ModTime.Equal(modTime) {
		t.Errorf("Expected modTime %v, got %v", modTime, file.ModTime)
	}
}

func TestGlobalIndex_Creation(t *testing.T) {
	index := &GlobalIndex{
		ByName:    make(map[string][]Reference),
		ByFile:    make(map[string][]Reference),
		ByType:    make(map[string][]Reference),
		CallGraph: make(map[string][]string),
		TypeGraph: make(map[string][]string),
	}

	if index.ByName == nil {
		t.Error("ByName map should be initialized")
	}
	if index.ByFile == nil {
		t.Error("ByFile map should be initialized")
	}
	if index.ByType == nil {
		t.Error("ByType map should be initialized")
	}
	if index.CallGraph == nil {
		t.Error("CallGraph map should be initialized")
	}
	if index.TypeGraph == nil {
		t.Error("TypeGraph map should be initialized")
	}
}
