package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndexBuilder_IntegratedGlobalEnrichment(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test Go files that demonstrate cross-file relationships
	testFiles := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	user := CreateUser("John")
	ProcessUser(user)
	fmt.Println("Done")
}`,
		"user.go": `package main

func CreateUser(name string) *User {
	return &User{Name: name}
}

func ProcessUser(user *User) {
	ValidateUser(user)
	SaveUser(user)
}

func ValidateUser(user *User) bool {
	return user.Name != ""
}

func SaveUser(user *User) error {
	return nil
}

type User struct {
	Name string
}`,
		"helpers.go": `package main

func Helper() {
	ProcessUser(&User{Name: "Helper"})
}`,
	}

	// Write test files to temp directory
	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		err = os.WriteFile(filePath, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create and initialize IndexBuilder
	builder := NewIndexBuilder(tempDir)
	err = builder.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize IndexBuilder: %v", err)
	}
	defer builder.Close()

	// Build the index with global enrichment
	stats, err := builder.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Verify statistics
	if stats.FilesProcessed != 3 {
		t.Errorf("Expected 3 files processed, got %d", stats.FilesProcessed)
	}

	expectedMinFunctions := 6 // main, CreateUser, ProcessUser, ValidateUser, SaveUser, Helper
	if stats.FunctionsIndexed < expectedMinFunctions {
		t.Errorf("Expected at least %d functions indexed, got %d", expectedMinFunctions, stats.FunctionsIndexed)
	}

	// Test that enriched data is stored and queryable
	storage := builder.storage

	// Query for ProcessUser function
	processUserResults, err := storage.QueryByName("ProcessUser")
	if err != nil {
		t.Fatalf("Failed to query ProcessUser: %v", err)
	}

	if len(processUserResults) != 1 {
		t.Errorf("Expected 1 ProcessUser result, got %d", len(processUserResults))
	}

	// Test call graph queries
	callsFromMain, err := storage.QueryCallsFrom("main")
	if err != nil {
		t.Fatalf("Failed to query calls from main: %v", err)
	}

	// main should call CreateUser, ProcessUser, and fmt.Println
	if len(callsFromMain) < 3 {
		t.Errorf("Expected at least 3 calls from main, got %d", len(callsFromMain))
	}

	// Verify specific calls
	callTargets := make(map[string]bool)
	for _, call := range callsFromMain {
		callTargets[call.Callee] = true
	}

	expectedCalls := []string{"CreateUser", "ProcessUser", "fmt.Println"}
	for _, expectedCall := range expectedCalls {
		if !callTargets[expectedCall] {
			t.Errorf("Expected main to call %s", expectedCall)
		}
	}

	// Test cross-file calls - ProcessUser should be called by both main and Helper
	callsToProcessUser, err := storage.QueryCallsTo("ProcessUser")
	if err != nil {
		t.Fatalf("Failed to query calls to ProcessUser: %v", err)
	}

	if len(callsToProcessUser) != 2 {
		t.Errorf("Expected 2 calls to ProcessUser, got %d", len(callsToProcessUser))
	}

	// Verify callers
	callerNames := make(map[string]bool)
	for _, call := range callsToProcessUser {
		callerNames[call.Caller] = true
	}

	if !callerNames["main"] {
		t.Error("Expected ProcessUser to be called by main")
	}
	if !callerNames["Helper"] {
		t.Error("Expected ProcessUser to be called by Helper")
	}
}

func TestIndexBuilder_GlobalEnrichmentWithRealParser(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_real_parser_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a more realistic Go project structure
	testFiles := map[string]string{
		"cmd/server/main.go": `package main

import (
	"fmt"
	"myapp/internal/service"
)

func main() {
	app := service.NewApp()
	app.Start()
	fmt.Println("Server started")
}`,
		"internal/service/app.go": `package service

import "myapp/internal/database"

type App struct {
	db *database.DB
}

func NewApp() *App {
	return &App{
		db: database.Connect(),
	}
}

func (a *App) Start() error {
	return a.db.Initialize()
}`,
		"internal/database/db.go": `package database

type DB struct{}

func Connect() *DB {
	return &DB{}
}

func (db *DB) Initialize() error {
	return nil
}`,
	}

	// Create directory structure and files
	for filePath, content := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)
		dir := filepath.Dir(fullPath)

		// Create directory if it doesn't exist
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		// Write file
		err = os.WriteFile(fullPath, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filePath, err)
		}
	}

	// Create and initialize IndexBuilder
	builder := NewIndexBuilder(tempDir)
	err = builder.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize IndexBuilder: %v", err)
	}
	defer builder.Close()

	// Build the index
	stats, err := builder.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Verify we processed the expected number of files
	if stats.FilesProcessed != 3 {
		t.Errorf("Expected 3 files processed, got %d", stats.FilesProcessed)
	}

	// Test that we can query the enriched data
	storage := builder.storage

	// Query for the main function
	mainResults, err := storage.QueryByName("main")
	if err != nil {
		t.Fatalf("Failed to query main function: %v", err)
	}

	if len(mainResults) != 1 {
		t.Errorf("Expected 1 main function result, got %d", len(mainResults))
	}

	// Test cross-package calls
	callsFromMain, err := storage.QueryCallsFrom("main")
	if err != nil {
		t.Fatalf("Failed to query calls from main: %v", err)
	}

	// main should call service.NewApp and other functions
	if len(callsFromMain) == 0 {
		t.Error("Expected main to have function calls")
	}

	// Print statistics for verification
	t.Logf("Files processed: %d", stats.FilesProcessed)
	t.Logf("Functions indexed: %d", stats.FunctionsIndexed)
	t.Logf("Types indexed: %d", stats.TypesIndexed)
	t.Logf("Calls indexed: %d", stats.CallsIndexed)
	t.Logf("Duration: %v", stats.Duration)
}
