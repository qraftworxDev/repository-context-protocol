package golang

import (
	"testing"

	"repository-context-protocol/internal/models"
)

func TestGoParser_InterfaceMethodLineNumbers(t *testing.T) {
	parser := NewGoParser()

	code := `package test

// Writer interface with methods
type Writer interface {
	Write(data []byte) (int, error)
	Close() error
}

// Reader interface with single method
type Reader interface {
	Read(buffer []byte) (int, error)
}`

	fileContext, err := parser.ParseFile("interface_lines.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Find Writer interface
	var writerType *models.TypeDef
	for i := range fileContext.Types {
		if fileContext.Types[i].Name == "Writer" {
			writerType = &fileContext.Types[i]
			break
		}
	}

	if writerType == nil {
		t.Fatal("Expected to find Writer interface")
	}

	// Check Writer interface line numbers
	if writerType.StartLine == 0 || writerType.EndLine == 0 {
		t.Errorf("Writer interface should have non-zero line numbers, got start=%d, end=%d",
			writerType.StartLine, writerType.EndLine)
	}

	// Check Writer methods
	if len(writerType.Methods) != 2 {
		t.Fatalf("Expected Writer to have 2 methods, got %d", len(writerType.Methods))
	}

	// Check Write method line numbers
	writeMethod := writerType.Methods[0]
	if writeMethod.Name != "Write" {
		t.Errorf("Expected first method to be Write, got %s", writeMethod.Name)
	}
	if writeMethod.StartLine == 0 || writeMethod.EndLine == 0 {
		t.Errorf("Write method should have non-zero line numbers, got start=%d, end=%d",
			writeMethod.StartLine, writeMethod.EndLine)
	}

	// Check Close method line numbers
	closeMethod := writerType.Methods[1]
	if closeMethod.Name != "Close" {
		t.Errorf("Expected second method to be Close, got %s", closeMethod.Name)
	}
	if closeMethod.StartLine == 0 || closeMethod.EndLine == 0 {
		t.Errorf("Close method should have non-zero line numbers, got start=%d, end=%d",
			closeMethod.StartLine, closeMethod.EndLine)
	}

	// Methods should be on different lines
	if writeMethod.StartLine >= closeMethod.StartLine {
		t.Errorf("Write method (line %d) should be before Close method (line %d)",
			writeMethod.StartLine, closeMethod.StartLine)
	}

	// Find Reader interface
	var readerType *models.TypeDef
	for i := range fileContext.Types {
		if fileContext.Types[i].Name == "Reader" {
			readerType = &fileContext.Types[i]
			break
		}
	}

	if readerType == nil {
		t.Fatal("Expected to find Reader interface")
	}

	// Check Reader method
	if len(readerType.Methods) != 1 {
		t.Fatalf("Expected Reader to have 1 method, got %d", len(readerType.Methods))
	}

	readMethod := readerType.Methods[0]
	if readMethod.StartLine == 0 || readMethod.EndLine == 0 {
		t.Errorf("Read method should have non-zero line numbers, got start=%d, end=%d",
			readMethod.StartLine, readMethod.EndLine)
	}
}

func TestGoParser_ComplexInterfaceLineNumbers(t *testing.T) {
	parser := NewGoParser()

	code := `package complex

import "context"

// Service interface with complex methods
type Service interface {
	// Get retrieves an item
	Get(ctx context.Context, id string) (*Item, error)

	// List returns multiple items
	List(ctx context.Context, filter map[string]interface{}) ([]*Item, error)

	// Create adds a new item
	Create(ctx context.Context, item *Item) error

	// Update modifies an existing item
	Update(ctx context.Context, id string, updates map[string]interface{}) error

	// Delete removes an item
	Delete(ctx context.Context, id string) error
}`

	fileContext, err := parser.ParseFile("complex_interface.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Find Service interface
	var serviceType *models.TypeDef
	for i := range fileContext.Types {
		if fileContext.Types[i].Name == "Service" {
			serviceType = &fileContext.Types[i]
			break
		}
	}

	if serviceType == nil {
		t.Fatal("Expected to find Service interface")
	}

	// Should have 5 methods
	if len(serviceType.Methods) != 5 {
		t.Fatalf("Expected Service to have 5 methods, got %d", len(serviceType.Methods))
	}

	expectedMethods := []string{"Get", "List", "Create", "Update", "Delete"}

	for i, expectedName := range expectedMethods {
		method := serviceType.Methods[i]

		if method.Name != expectedName {
			t.Errorf("Expected method %d to be %s, got %s", i, expectedName, method.Name)
		}

		if method.StartLine == 0 || method.EndLine == 0 {
			t.Errorf("Method %s should have non-zero line numbers, got start=%d, end=%d",
				method.Name, method.StartLine, method.EndLine)
		}

		// Each method should be on a different line (they're spaced out in the code)
		if i > 0 {
			prevMethod := serviceType.Methods[i-1]
			if method.StartLine <= prevMethod.StartLine {
				t.Errorf("Method %s (line %d) should be after method %s (line %d)",
					method.Name, method.StartLine, prevMethod.Name, prevMethod.StartLine)
			}
		}
	}
}

func TestGoParser_EmbeddedInterfaceLineNumbers(t *testing.T) {
	parser := NewGoParser()

	code := `package embedded

// Base interface
type Base interface {
	BaseMethod() string
}

// Extended interface with embedding
type Extended interface {
	Base  // Embedded interface
	ExtendedMethod() int
}`

	fileContext, err := parser.ParseFile("embedded.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	// Find Extended interface
	var extendedType *models.TypeDef
	for i := range fileContext.Types {
		if fileContext.Types[i].Name == "Extended" {
			extendedType = &fileContext.Types[i]
			break
		}
	}

	if extendedType == nil {
		t.Fatal("Expected to find Extended interface")
	}

	// Should have 1 method (ExtendedMethod) and 1 embedded interface
	if len(extendedType.Methods) != 1 {
		t.Errorf("Expected Extended to have 1 method, got %d", len(extendedType.Methods))
	}

	if len(extendedType.Embedded) != 1 {
		t.Errorf("Expected Extended to have 1 embedded interface, got %d", len(extendedType.Embedded))
	}

	// Check method line numbers
	method := extendedType.Methods[0]
	if method.StartLine == 0 || method.EndLine == 0 {
		t.Errorf("ExtendedMethod should have non-zero line numbers, got start=%d, end=%d",
			method.StartLine, method.EndLine)
	}
}
