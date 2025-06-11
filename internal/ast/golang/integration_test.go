package golang

import (
	"os"
	"path/filepath"
	"testing"

	"repository-context-protocol/internal/models"
)

func TestGoParser_IntegrationTest(t *testing.T) {
	parser := NewGoParser()

	// Read the test file
	testFile := filepath.Join("..", "..", "..", "testdata", "simple-go", "main.go")
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Parse the file
	fileContext, err := parser.ParseFile(testFile, content)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Validate basic file information
	if fileContext.Path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, fileContext.Path)
	}
	if fileContext.Language != "go" {
		t.Errorf("Expected language 'go', got %s", fileContext.Language)
	}

	// Validate imports
	expectedImports := []string{"fmt", "strings"}
	if len(fileContext.Imports) != len(expectedImports) {
		t.Errorf("Expected %d imports, got %d", len(expectedImports), len(fileContext.Imports))
	}
	for i, expectedImport := range expectedImports {
		if i < len(fileContext.Imports) && fileContext.Imports[i].Path != expectedImport {
			t.Errorf("Expected import %s, got %s", expectedImport, fileContext.Imports[i].Path)
		}
	}

	// Validate types
	expectedTypes := []string{"User", "UserService", "InMemoryUserService"}
	if len(fileContext.Types) != len(expectedTypes) {
		t.Errorf("Expected %d types, got %d", len(expectedTypes), len(fileContext.Types))
	}

	// Find and validate User struct
	userType := findType(fileContext.Types, "User")
	if userType == nil {
		t.Error("Expected to find User type")
	} else {
		if userType.Kind != "struct" {
			t.Errorf("Expected User to be struct, got %s", userType.Kind)
		}
		if len(userType.Fields) != 4 {
			t.Errorf("Expected User to have 4 fields, got %d", len(userType.Fields))
		}
		if len(userType.Methods) != 2 {
			t.Errorf("Expected User to have 2 methods, got %d", len(userType.Methods))
		}
	}

	// Find and validate UserService interface
	serviceType := findType(fileContext.Types, "UserService")
	if serviceType == nil {
		t.Error("Expected to find UserService type")
	} else {
		if serviceType.Kind != "interface" {
			t.Errorf("Expected UserService to be interface, got %s", serviceType.Kind)
		}
		if len(serviceType.Methods) != 4 {
			t.Errorf("Expected UserService to have 4 methods, got %d", len(serviceType.Methods))
		}
	}

	// Find and validate InMemoryUserService struct
	implType := findType(fileContext.Types, "InMemoryUserService")
	if implType == nil {
		t.Error("Expected to find InMemoryUserService type")
	} else {
		if implType.Kind != "struct" {
			t.Errorf("Expected InMemoryUserService to be struct, got %s", implType.Kind)
		}
		if len(implType.Fields) != 2 {
			t.Errorf("Expected InMemoryUserService to have 2 fields, got %d", len(implType.Fields))
		}
		if len(implType.Methods) != 4 {
			t.Errorf("Expected InMemoryUserService to have 4 methods, got %d", len(implType.Methods))
		}
	}

	// Validate functions (should include main and NewInMemoryUserService)
	expectedFunctions := []string{"main", "NewInMemoryUserService"}
	foundFunctions := 0
	for _, expectedFunc := range expectedFunctions {
		if findFunction(fileContext.Functions, expectedFunc) != nil {
			foundFunctions++
		}
	}
	if foundFunctions != len(expectedFunctions) {
		t.Errorf("Expected to find %d specific functions, found %d", len(expectedFunctions), foundFunctions)
	}

	// Validate that we have a reasonable number of total functions
	// (including methods which are also parsed as functions)
	if len(fileContext.Functions) < 8 {
		t.Errorf("Expected at least 8 functions (including methods), got %d", len(fileContext.Functions))
	}
}

func TestGoParser_RealWorldComplexity(t *testing.T) {
	parser := NewGoParser()

	// Test with a more complex Go construct
	complexCode := `package complex

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
)

type Repository[T any] interface {
	Get(ctx context.Context, id string) (T, error)
	List(ctx context.Context, filters map[string]interface{}) ([]T, error)
	Create(ctx context.Context, entity T) error
	Update(ctx context.Context, id string, entity T) error
	Delete(ctx context.Context, id string) error
}

type HTTPHandler struct {
	repo Repository[User]
	db   *sql.DB
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HTTPHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	// Implementation here
}

func (h *HTTPHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	// Implementation here
}`

	fileContext, err := parser.ParseFile("complex.go", []byte(complexCode))
	if err != nil {
		t.Fatalf("Failed to parse complex code: %v", err)
	}

	// Should handle generics and complex types
	if len(fileContext.Types) < 2 {
		t.Errorf("Expected at least 2 types in complex code, got %d", len(fileContext.Types))
	}

	// Should handle multiple imports
	if len(fileContext.Imports) != 4 {
		t.Errorf("Expected 4 imports, got %d", len(fileContext.Imports))
	}

	// Should handle methods
	httpHandlerType := findType(fileContext.Types, "HTTPHandler")
	if httpHandlerType == nil {
		t.Error("Expected to find HTTPHandler type")
	} else if len(httpHandlerType.Methods) < 3 {
		t.Errorf("Expected HTTPHandler to have at least 3 methods, got %d", len(httpHandlerType.Methods))
	}
}

// Helper function to find a type by name
func findType(types []models.TypeDef, name string) *models.TypeDef {
	for i := range types {
		if types[i].Name == name {
			return &types[i]
		}
	}
	return nil
}
