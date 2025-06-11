package models

import "testing"

func TestFunction_Creation(t *testing.T) {
	fn := &Function{
		Name:      "greet",
		Signature: "func greet(name string) string",
		Parameters: []Parameter{
			{Name: "name", Type: "string"},
		},
		Returns: []Type{
			{Name: "string", Kind: "basic"},
		},
		StartLine: 10,
		EndLine:   12,
		Calls:     []string{"fmt.Sprintf"},
		CalledBy:  []string{"main"},
	}

	if fn.Name != "greet" {
		t.Errorf("Expected name 'greet', got %s", fn.Name)
	}
	if len(fn.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(fn.Parameters))
	}
	if fn.Parameters[0].Name != "name" {
		t.Errorf("Expected parameter name 'name', got %s", fn.Parameters[0].Name)
	}
	if len(fn.Returns) != 1 {
		t.Errorf("Expected 1 return type, got %d", len(fn.Returns))
	}
	if fn.StartLine != 10 {
		t.Errorf("Expected start line 10, got %d", fn.StartLine)
	}
}

func TestParameter_Creation(t *testing.T) {
	param := Parameter{
		Name: "ctx",
		Type: "context.Context",
	}

	if param.Name != "ctx" {
		t.Errorf("Expected name 'ctx', got %s", param.Name)
	}
	if param.Type != "context.Context" {
		t.Errorf("Expected type 'context.Context', got %s", param.Type)
	}
}

func TestVariable_Creation(t *testing.T) {
	variable := Variable{
		Name:      "result",
		Type:      "string",
		StartLine: 5,
		EndLine:   5,
	}

	if variable.Name != "result" {
		t.Errorf("Expected name 'result', got %s", variable.Name)
	}
	if variable.Type != "string" {
		t.Errorf("Expected type 'string', got %s", variable.Type)
	}
}

func TestImport_Creation(t *testing.T) {
	imp := Import{
		Path:  "fmt",
		Alias: "",
	}

	if imp.Path != "fmt" {
		t.Errorf("Expected path 'fmt', got %s", imp.Path)
	}

	// Test with alias
	impWithAlias := Import{
		Path:  "github.com/stretchr/testify/assert",
		Alias: "assert",
	}

	if impWithAlias.Alias != "assert" {
		t.Errorf("Expected alias 'assert', got %s", impWithAlias.Alias)
	}
}

func TestExport_Creation(t *testing.T) {
	export := Export{
		Name: "ProcessUser",
		Type: "func(User) error",
		Kind: "function",
	}

	if export.Name != "ProcessUser" {
		t.Errorf("Expected name 'ProcessUser', got %s", export.Name)
	}
	if export.Kind != "function" {
		t.Errorf("Expected kind 'function', got %s", export.Kind)
	}
}
