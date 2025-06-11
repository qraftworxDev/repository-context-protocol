package models

import "testing"

func TestReference_Creation(t *testing.T) {
	ref := &Reference{
		File:      "main.go",
		Entity:    "greet",
		Type:      "function",
		StartLine: 10,
		EndLine:   12,
		Signature: "func greet(name string) string",
	}

	if ref.File != "main.go" {
		t.Errorf("Expected file 'main.go', got %s", ref.File)
	}
	if ref.Entity != "greet" {
		t.Errorf("Expected entity 'greet', got %s", ref.Entity)
	}
	if ref.Type != "function" {
		t.Errorf("Expected type 'function', got %s", ref.Type)
	}
	if ref.StartLine != 10 {
		t.Errorf("Expected start line 10, got %d", ref.StartLine)
	}
	if ref.EndLine != 12 {
		t.Errorf("Expected end line 12, got %d", ref.EndLine)
	}
}

func TestReference_TypeReference(t *testing.T) {
	ref := &Reference{
		File:      "user.go",
		Entity:    "User",
		Type:      "type",
		StartLine: 5,
		EndLine:   15,
		Signature: "type User struct",
	}

	if ref.Type != "type" {
		t.Errorf("Expected type 'type', got %s", ref.Type)
	}
	if ref.Entity != "User" {
		t.Errorf("Expected entity 'User', got %s", ref.Entity)
	}
}

func TestReference_VariableReference(t *testing.T) {
	ref := &Reference{
		File:      "config.go",
		Entity:    "DefaultTimeout",
		Type:      "variable",
		StartLine: 8,
		EndLine:   8,
		Signature: "var DefaultTimeout time.Duration",
	}

	if ref.Type != "variable" {
		t.Errorf("Expected type 'variable', got %s", ref.Type)
	}
	if ref.Entity != "DefaultTimeout" {
		t.Errorf("Expected entity 'DefaultTimeout', got %s", ref.Entity)
	}
}

func TestReference_ImportReference(t *testing.T) {
	ref := &Reference{
		File:      "main.go",
		Entity:    "fmt",
		Type:      "import",
		StartLine: 3,
		EndLine:   3,
		Signature: `import "fmt"`,
	}

	if ref.Type != "import" {
		t.Errorf("Expected type 'import', got %s", ref.Type)
	}
	if ref.Entity != "fmt" {
		t.Errorf("Expected entity 'fmt', got %s", ref.Entity)
	}
}
