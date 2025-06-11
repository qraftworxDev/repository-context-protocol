package golang

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGoParser_Demo(t *testing.T) {
	parser := NewGoParser()

	// Demo code with various Go constructs including variables and constants
	demoCode := `package demo

import (
	"fmt"
	"context"
)

// Constants
const (
	MaxUsers = 1000
	DefaultTimeout = 30
)

// Variables
var (
	GlobalCounter int
	AppName = "UserService"
)

// Single variable declaration
var DatabaseURL string = "localhost:5432"

// User represents a user entity
type User struct {
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}

// String returns string representation
func (u User) String() string {
	return fmt.Sprintf("User{ID: %d, Name: %s}", u.ID, u.Name)
}

// Repository defines data access interface
type Repository interface {
	Get(ctx context.Context, id int) (*User, error)
	Save(ctx context.Context, user *User) error
}

// CreateUser creates a new user
func CreateUser(name string) *User {
	return &User{Name: name}
}

func main() {
	user := CreateUser("Alice")
	fmt.Println(user.String())
}`

	// Parse the demo code
	fileContext, err := parser.ParseFile("demo.go", []byte(demoCode))
	if err != nil {
		t.Fatalf("Failed to parse demo code: %v", err)
	}

	// Convert to JSON for readable output
	jsonData, err := json.MarshalIndent(fileContext, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Print the parsed structure (this will show in test output with -v)
	fmt.Printf("\n=== PARSED GO AST STRUCTURE ===\n%s\n", string(jsonData))

	// Verify key elements are extracted
	if len(fileContext.Imports) != 2 {
		t.Errorf("Expected 2 imports, got %d", len(fileContext.Imports))
	}

	if len(fileContext.Types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(fileContext.Types))
	}

	if len(fileContext.Functions) < 3 {
		t.Errorf("Expected at least 3 functions, got %d", len(fileContext.Functions))
	}

	// Verify variables are extracted (should be 3: GlobalCounter, AppName, DatabaseURL)
	expectedVarCount := 3
	if len(fileContext.Variables) != expectedVarCount {
		t.Errorf("Expected %d variables, got %d", expectedVarCount, len(fileContext.Variables))
	}

	// Verify constants are extracted (should be 2: MaxUsers, DefaultTimeout)
	expectedConstCount := 2
	if len(fileContext.Constants) != expectedConstCount {
		t.Errorf("Expected %d constants, got %d", expectedConstCount, len(fileContext.Constants))
	}

	// Check for specific variables
	expectedVars := map[string]bool{
		"GlobalCounter": false,
		"AppName":       false,
		"DatabaseURL":   false,
	}

	for _, variable := range fileContext.Variables {
		if _, exists := expectedVars[variable.Name]; exists {
			expectedVars[variable.Name] = true
		}
	}

	for varName, found := range expectedVars {
		if !found {
			t.Errorf("Expected to find variable '%s', but it was not extracted", varName)
		}
	}

	// Check for specific constants
	expectedConsts := map[string]bool{
		"MaxUsers":       false,
		"DefaultTimeout": false,
	}

	for _, constant := range fileContext.Constants {
		if _, exists := expectedConsts[constant.Name]; exists {
			expectedConsts[constant.Name] = true
		}
	}

	for constName, found := range expectedConsts {
		if !found {
			t.Errorf("Expected to find constant '%s', but it was not extracted", constName)
		}
	}

	// Verify User struct has methods
	userType := findType(fileContext.Types, "User")
	if userType == nil {
		t.Error("Expected to find User type")
	} else if len(userType.Methods) != 1 {
		t.Errorf("Expected User to have 1 method, got %d", len(userType.Methods))
	}

	// Verify Repository interface has methods
	repoType := findType(fileContext.Types, "Repository")
	if repoType == nil {
		t.Error("Expected to find Repository type")
	} else if len(repoType.Methods) != 2 {
		t.Errorf("Expected Repository to have 2 methods, got %d", len(repoType.Methods))
	}
}
