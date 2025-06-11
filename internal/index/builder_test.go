package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndexBuilder_Initialize(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	builder := NewIndexBuilder(tempDir)

	// Initialize index builder
	err = builder.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize index builder: %v", err)
	}
	defer builder.Close()

	// Verify components are initialized
	if builder.storage == nil {
		t.Error("Expected storage to be initialized")
	}
	if builder.parserRegistry == nil {
		t.Error("Expected parser registry to be initialized")
	}

	// Verify directory structure was created
	indexPath := filepath.Join(tempDir, ".repocontext")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Errorf("Expected .repocontext directory to be created at %s", indexPath)
	}
}

func TestIndexBuilder_ProcessFile(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	builder := NewIndexBuilder(tempDir)
	err = builder.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize index builder: %v", err)
	}
	defer builder.Close()

	// Create a test Go file
	testFile := filepath.Join(tempDir, "test.go")
	testContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}

type User struct {
	Name string
	Age  int
}
`
	err = os.WriteFile(testFile, []byte(testContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Process the file
	err = builder.ProcessFile(testFile)
	if err != nil {
		t.Fatalf("Failed to process file: %v", err)
	}

	// Verify the file was indexed
	results, err := builder.storage.QueryByName("main")
	if err != nil {
		t.Fatalf("Failed to query for main function: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for main function, got %d", len(results))
	}

	// Verify type was indexed
	typeResults, err := builder.storage.QueryByName("User")
	if err != nil {
		t.Fatalf("Failed to query for User type: %v", err)
	}
	if len(typeResults) != 1 {
		t.Errorf("Expected 1 result for User type, got %d", len(typeResults))
	}
}

func TestIndexBuilder_ProcessDirectory(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	builder := NewIndexBuilder(tempDir)
	err = builder.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize index builder: %v", err)
	}
	defer builder.Close()

	// Create test directory structure
	srcDir := filepath.Join(tempDir, "src")
	err = os.MkdirAll(srcDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}

	// Create multiple Go files
	files := map[string]string{
		"main.go": `package main

func main() {
	helper()
}

func helper() {
	println("helper")
}`,
		"utils.go": `package main

func Process(data string) string {
	return validate(data)
}

func validate(s string) string {
	return s
}`,
		"types.go": `package main

type Config struct {
	Host string
	Port int
}

type Database interface {
	Connect() error
	Close() error
}`,
	}

	for filename, content := range files {
		filePath := filepath.Join(srcDir, filename)
		err = os.WriteFile(filePath, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Process the directory
	err = builder.ProcessDirectory(srcDir)
	if err != nil {
		t.Fatalf("Failed to process directory: %v", err)
	}

	// Verify all functions were indexed
	expectedFunctions := []string{"main", "helper", "Process", "validate"}
	for _, funcName := range expectedFunctions {
		results, err := builder.storage.QueryByName(funcName)
		if err != nil {
			t.Fatalf("Failed to query for function %s: %v", funcName, err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result for function %s, got %d", funcName, len(results))
		}
	}

	// Verify types were indexed
	expectedTypes := []string{"Config", "Database"}
	for _, typeName := range expectedTypes {
		results, err := builder.storage.QueryByName(typeName)
		if err != nil {
			t.Fatalf("Failed to query for type %s: %v", typeName, err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result for type %s, got %d", typeName, len(results))
		}
	}
}

func TestIndexBuilder_BuildIndex(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a project structure
	projectDir := filepath.Join(tempDir, "project")
	err = os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create go.mod file
	goModContent := `module testproject

go 1.21
`
	err = os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte(goModContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create source files
	mainContent := `package main

import "fmt"

func main() {
	user := CreateUser("John")
	fmt.Println(user.String())
}
`
	err = os.WriteFile(filepath.Join(projectDir, "main.go"), []byte(mainContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	userContent := `package main

import "fmt"

type User struct {
	Name string
}

func CreateUser(name string) *User {
	return &User{Name: name}
}

func (u *User) String() string {
	return fmt.Sprintf("User: %s", u.Name)
}
`
	err = os.WriteFile(filepath.Join(projectDir, "user.go"), []byte(userContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create user.go: %v", err)
	}

	// Initialize builder with project directory
	builder := NewIndexBuilder(projectDir)
	err = builder.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize index builder: %v", err)
	}
	defer builder.Close()

	// Build the index
	stats, err := builder.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Verify statistics
	if stats.FilesProcessed != 2 {
		t.Errorf("Expected 2 files processed, got %d", stats.FilesProcessed)
	}
	if stats.FunctionsIndexed < 3 { // main, CreateUser, String
		t.Errorf("Expected at least 3 functions indexed, got %d", stats.FunctionsIndexed)
	}
	if stats.TypesIndexed < 1 { // User
		t.Errorf("Expected at least 1 type indexed, got %d", stats.TypesIndexed)
	}

	// Verify call relationships were established
	callsFromMain, err := builder.storage.QueryCallsFrom("main")
	if err != nil {
		t.Fatalf("Failed to query calls from main: %v", err)
	}
	if len(callsFromMain) < 2 { // CreateUser, fmt.Println
		t.Errorf("Expected at least 2 calls from main, got %d", len(callsFromMain))
	}
}

func TestIndexBuilder_GetStatistics(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	builder := NewIndexBuilder(tempDir)
	err = builder.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize index builder: %v", err)
	}
	defer builder.Close()

	// Get initial statistics
	stats := builder.GetStatistics()
	if stats.FilesProcessed != 0 {
		t.Errorf("Expected 0 files processed initially, got %d", stats.FilesProcessed)
	}

	// Create and process a test file
	testFile := filepath.Join(tempDir, "test.go")
	testContent := `package main

func test() {}
func another() {}

type TestType struct {}
type AnotherType interface {}

var testVar string
const testConst = "test"
`
	err = os.WriteFile(testFile, []byte(testContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = builder.ProcessFile(testFile)
	if err != nil {
		t.Fatalf("Failed to process file: %v", err)
	}

	// Get updated statistics
	stats = builder.GetStatistics()
	if stats.FilesProcessed != 1 {
		t.Errorf("Expected 1 file processed, got %d", stats.FilesProcessed)
	}
	if stats.FunctionsIndexed != 2 {
		t.Errorf("Expected 2 functions indexed, got %d", stats.FunctionsIndexed)
	}
	if stats.TypesIndexed != 2 {
		t.Errorf("Expected 2 types indexed, got %d", stats.TypesIndexed)
	}
	if stats.VariablesIndexed != 1 {
		t.Errorf("Expected 1 variable indexed, got %d", stats.VariablesIndexed)
	}
	if stats.ConstantsIndexed != 1 {
		t.Errorf("Expected 1 constant indexed, got %d", stats.ConstantsIndexed)
	}
}

func TestIndexBuilder_Close(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	builder := NewIndexBuilder(tempDir)
	err = builder.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize index builder: %v", err)
	}

	// Close the builder
	err = builder.Close()
	if err != nil {
		t.Errorf("Failed to close index builder: %v", err)
	}

	// Verify components are properly closed
	// Attempting to use closed builder should fail gracefully
	err = builder.ProcessFile("nonexistent.go")
	if err == nil {
		t.Error("Expected error when using closed builder")
	}
}
