package models

import "testing"

func TestTypeDef_Struct(t *testing.T) {
	typeDef := &TypeDef{
		Name: "User",
		Kind: "struct",
		Fields: []Field{
			{Name: "ID", Type: "int", Tag: `json:"id"`},
			{Name: "Name", Type: "string", Tag: `json:"name"`},
		},
		Methods: []Method{
			{
				Name:      "String",
				Signature: "func (u User) String() string",
				Returns:   []Type{{Name: "string", Kind: "basic"}},
				StartLine: 15,
				EndLine:   17,
			},
		},
		StartLine: 5,
		EndLine:   10,
	}

	if typeDef.Name != "User" {
		t.Errorf("Expected name 'User', got %s", typeDef.Name)
	}
	if typeDef.Kind != "struct" {
		t.Errorf("Expected kind 'struct', got %s", typeDef.Kind)
	}
	if len(typeDef.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(typeDef.Fields))
	}
	if len(typeDef.Methods) != 1 {
		t.Errorf("Expected 1 method, got %d", len(typeDef.Methods))
	}
}

func TestTypeDef_Interface(t *testing.T) {
	typeDef := &TypeDef{
		Name: "Writer",
		Kind: "interface",
		Methods: []Method{
			{
				Name:      "Write",
				Signature: "Write([]byte) (int, error)",
				Parameters: []Parameter{
					{Name: "p", Type: "[]byte"},
				},
				Returns: []Type{
					{Name: "int", Kind: "basic"},
					{Name: "error", Kind: "interface"},
				},
			},
		},
		StartLine: 20,
		EndLine:   22,
	}

	if typeDef.Kind != "interface" {
		t.Errorf("Expected kind 'interface', got %s", typeDef.Kind)
	}
	if len(typeDef.Fields) != 0 {
		t.Errorf("Expected 0 fields for interface, got %d", len(typeDef.Fields))
	}
}

func TestField_Creation(t *testing.T) {
	field := Field{
		Name: "Email",
		Type: "string",
		Tag:  `json:"email" validate:"email"`,
	}

	if field.Name != "Email" {
		t.Errorf("Expected name 'Email', got %s", field.Name)
	}
	if field.Type != "string" {
		t.Errorf("Expected type 'string', got %s", field.Type)
	}
	if field.Tag != `json:"email" validate:"email"` {
		t.Errorf("Expected tag with validation, got %s", field.Tag)
	}
}

func TestMethod_Creation(t *testing.T) {
	method := Method{
		Name:      "Process",
		Signature: "func (p *Processor) Process(data []byte) error",
		Parameters: []Parameter{
			{Name: "data", Type: "[]byte"},
		},
		Returns: []Type{
			{Name: "error", Kind: "interface"},
		},
		StartLine: 25,
		EndLine:   30,
	}

	if method.Name != "Process" {
		t.Errorf("Expected name 'Process', got %s", method.Name)
	}
	if len(method.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(method.Parameters))
	}
	if len(method.Returns) != 1 {
		t.Errorf("Expected 1 return type, got %d", len(method.Returns))
	}
}
