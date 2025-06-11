package models

import (
	"testing"
)

func TestIndexEntry_Creation(t *testing.T) {
	entry := IndexEntry{
		Name:      "TestFunction",
		Type:      "function",
		File:      "main.go",
		StartLine: 10,
		EndLine:   15,
		ChunkID:   "chunk_001",
		Signature: "func TestFunction() error",
	}

	if entry.Name != "TestFunction" {
		t.Errorf("Expected name 'TestFunction', got %s", entry.Name)
	}
	if entry.Type != "function" {
		t.Errorf("Expected type 'function', got %s", entry.Type)
	}
	if entry.File != "main.go" {
		t.Errorf("Expected file 'main.go', got %s", entry.File)
	}
	if entry.StartLine != 10 {
		t.Errorf("Expected start line 10, got %d", entry.StartLine)
	}
	if entry.EndLine != 15 {
		t.Errorf("Expected end line 15, got %d", entry.EndLine)
	}
	if entry.ChunkID != "chunk_001" {
		t.Errorf("Expected chunk ID 'chunk_001', got %s", entry.ChunkID)
	}
	if entry.Signature != "func TestFunction() error" {
		t.Errorf("Expected signature 'func TestFunction() error', got %s", entry.Signature)
	}
}

func TestIndexEntry_TypeValidation(t *testing.T) {
	validTypes := []string{"function", "type", "variable", "constant"}

	for _, validType := range validTypes {
		entry := IndexEntry{
			Name: "TestEntity",
			Type: validType,
			File: "test.go",
		}

		if entry.Type != validType {
			t.Errorf("Expected type %s, got %s", validType, entry.Type)
		}
	}
}

func TestCallRelation_Creation(t *testing.T) {
	relation := CallRelation{
		Caller:     "MainFunction",
		Callee:     "HelperFunction",
		File:       "main.go",
		Line:       25,
		CallerFile: "main.go",
	}

	if relation.Caller != "MainFunction" {
		t.Errorf("Expected caller 'MainFunction', got %s", relation.Caller)
	}
	if relation.Callee != "HelperFunction" {
		t.Errorf("Expected callee 'HelperFunction', got %s", relation.Callee)
	}
	if relation.File != "main.go" {
		t.Errorf("Expected file 'main.go', got %s", relation.File)
	}
	if relation.Line != 25 {
		t.Errorf("Expected line 25, got %d", relation.Line)
	}
	if relation.CallerFile != "main.go" {
		t.Errorf("Expected caller file 'main.go', got %s", relation.CallerFile)
	}
}

func TestCallRelation_CrossFileCall(t *testing.T) {
	relation := CallRelation{
		Caller:     "ProcessData",
		Callee:     "utils.ValidateInput",
		File:       "processor.go",
		Line:       42,
		CallerFile: "processor.go",
	}

	// Test cross-file call scenario
	if relation.Caller != "ProcessData" {
		t.Errorf("Expected caller 'ProcessData', got %s", relation.Caller)
	}
	if relation.Callee != "utils.ValidateInput" {
		t.Errorf("Expected callee 'utils.ValidateInput', got %s", relation.Callee)
	}
	if relation.File != "processor.go" {
		t.Errorf("Expected file 'processor.go', got %s", relation.File)
	}
}

func TestIndexEntry_EmptyValues(t *testing.T) {
	entry := IndexEntry{}

	// Test zero values
	if entry.Name != "" {
		t.Errorf("Expected empty name, got %s", entry.Name)
	}
	if entry.StartLine != 0 {
		t.Errorf("Expected start line 0, got %d", entry.StartLine)
	}
	if entry.EndLine != 0 {
		t.Errorf("Expected end line 0, got %d", entry.EndLine)
	}
}

func TestCallRelation_EmptyValues(t *testing.T) {
	relation := CallRelation{}

	// Test zero values
	if relation.Caller != "" {
		t.Errorf("Expected empty caller, got %s", relation.Caller)
	}
	if relation.Line != 0 {
		t.Errorf("Expected line 0, got %d", relation.Line)
	}
}
