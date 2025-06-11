package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndexBuilder_GlobalEnrichmentValidation(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_enrichment_validation_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a multi-file project that will test enrichment functionality
	testFiles := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	// Cross-file calls
	user := createUser("Alice")
	processUser(user)

	// External package call
	fmt.Println("User processed")

	// Local helper call
	logAction("main_completed")
}

// Local helper function
func logAction(action string) {
	fmt.Printf("Action: %s\n", action)
}`,

		"user.go": `package main

type User struct {
	Name string
	ID   int
}

func createUser(name string) *User {
	user := &User{Name: name, ID: generateID()}
	validateUser(user) // Local call within user.go
	return user
}

func processUser(user *User) {
	validateUser(user) // Local call within user.go
	saveUser(user)     // Cross-file call to database.go
}

func validateUser(user *User) bool {
	return user.Name != ""
}`,

		"database.go": `package main

import "fmt"

func saveUser(user *User) error {
	// Local helper within database.go
	conn := getConnection()
	defer closeConnection(conn)

	// External package call
	fmt.Printf("Saving user: %s\n", user.Name)
	return nil
}

func getConnection() string {
	return "db_connection"
}

func closeConnection(conn string) {
	fmt.Printf("Closing connection: %s\n", conn)
}`,

		"utils.go": `package main

import "math/rand"

func generateID() int {
	return rand.Intn(10000)
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

	// Print comprehensive statistics
	t.Logf("=== ENRICHMENT VALIDATION STATISTICS ===")
	t.Logf("Files processed: %d", stats.FilesProcessed)
	t.Logf("Functions indexed: %d", stats.FunctionsIndexed)
	t.Logf("Types indexed: %d", stats.TypesIndexed)
	t.Logf("Variables indexed: %d", stats.VariablesIndexed)
	t.Logf("Constants indexed: %d", stats.ConstantsIndexed)
	t.Logf("Calls indexed: %d", stats.CallsIndexed)

	// Verify enrichment worked correctly
	storage := builder.storage

	// Test 1: main() function should have mixed local and cross-file calls
	testMainFunctionEnrichment(t, storage)

	// Test 2: createUser() should have local and cross-file calls
	testCreateUserEnrichment(t, storage)

	// Test 3: processUser() should have local and cross-file calls
	testProcessUserEnrichment(t, storage)

	// Test 4: saveUser() should have only local calls
	testSaveUserEnrichment(t, storage)

	// Test 5: validateUser() should be called from multiple functions
	testValidateUserEnrichment(t, storage)

	// Test 6: Cross-file call statistics
	testCrossFileCallStatistics(t, storage)
}

func testMainFunctionEnrichment(t *testing.T, storage *HybridStorage) {
	t.Run("MainFunctionEnrichment", func(t *testing.T) {
		// Query calls from main
		callsFromMain, err := storage.QueryCallsFrom("main")
		if err != nil {
			t.Fatalf("Failed to query calls from main: %v", err)
		}

		// main should call: createUser, processUser, fmt.Println, logAction
		expectedCalls := map[string]struct{}{
			"createUser":  {},
			"processUser": {},
			"fmt.Println": {},
			"logAction":   {},
		}

		actualCalls := make(map[string]struct{})
		for _, call := range callsFromMain {
			actualCalls[call.Callee] = struct{}{}
		}

		for expectedCall := range expectedCalls {
			if _, found := actualCalls[expectedCall]; !found {
				t.Errorf("Expected main to call %s", expectedCall)
			}
		}

		// Verify we have at least the expected number of calls
		if len(callsFromMain) < len(expectedCalls) {
			t.Errorf("Expected at least %d calls from main, got %d", len(expectedCalls), len(callsFromMain))
		}
	})
}

func testCreateUserEnrichment(t *testing.T, storage *HybridStorage) {
	t.Run("CreateUserEnrichment", func(t *testing.T) {
		// Query calls from createUser
		callsFromCreateUser, err := storage.QueryCallsFrom("createUser")
		if err != nil {
			t.Fatalf("Failed to query calls from createUser: %v", err)
		}

		// createUser should call: generateID (cross-file), validateUser (local)
		expectedCalls := map[string]struct{}{
			"generateID":   {},
			"validateUser": {},
		}

		actualCalls := make(map[string]struct{})
		for _, call := range callsFromCreateUser {
			actualCalls[call.Callee] = struct{}{}
		}

		for expectedCall := range expectedCalls {
			if _, found := actualCalls[expectedCall]; !found {
				t.Errorf("Expected createUser to call %s", expectedCall)
			}
		}

		// Verify createUser is called by main
		callsToCreateUser, err := storage.QueryCallsTo("createUser")
		if err != nil {
			t.Fatalf("Failed to query calls to createUser: %v", err)
		}

		mainCallsCreateUser := false
		for _, call := range callsToCreateUser {
			if call.Caller == "main" {
				mainCallsCreateUser = true
				break
			}
		}

		if !mainCallsCreateUser {
			t.Error("Expected createUser to be called by main")
		}
	})
}

func testProcessUserEnrichment(t *testing.T, storage *HybridStorage) {
	t.Run("ProcessUserEnrichment", func(t *testing.T) {
		// Query calls from processUser
		callsFromProcessUser, err := storage.QueryCallsFrom("processUser")
		if err != nil {
			t.Fatalf("Failed to query calls from processUser: %v", err)
		}

		// processUser should call: validateUser (local), saveUser (cross-file)
		expectedCalls := map[string]struct{}{
			"validateUser": {},
			"saveUser":     {},
		}

		actualCalls := make(map[string]struct{})
		for _, call := range callsFromProcessUser {
			actualCalls[call.Callee] = struct{}{}
		}

		for expectedCall := range expectedCalls {
			if _, found := actualCalls[expectedCall]; !found {
				t.Errorf("Expected processUser to call %s", expectedCall)
			}
		}
	})
}

func testSaveUserEnrichment(t *testing.T, storage *HybridStorage) {
	t.Run("SaveUserEnrichment", func(t *testing.T) {
		// Query calls from saveUser
		callsFromSaveUser, err := storage.QueryCallsFrom("saveUser")
		if err != nil {
			t.Fatalf("Failed to query calls from saveUser: %v", err)
		}

		// saveUser should call: getConnection, closeConnection (both local), fmt.Printf (external)
		expectedCalls := map[string]struct{}{
			"getConnection":   {},
			"closeConnection": {},
			"fmt.Printf":      {},
		}

		actualCalls := make(map[string]struct{})
		for _, call := range callsFromSaveUser {
			actualCalls[call.Callee] = struct{}{}
		}

		for expectedCall := range expectedCalls {
			if _, found := actualCalls[expectedCall]; !found {
				t.Errorf("Expected saveUser to call %s", expectedCall)
			}
		}

		// Verify saveUser is called by processUser (cross-file call)
		callsToSaveUser, err := storage.QueryCallsTo("saveUser")
		if err != nil {
			t.Fatalf("Failed to query calls to saveUser: %v", err)
		}

		processUserCallsSaveUser := false
		for _, call := range callsToSaveUser {
			if call.Caller == "processUser" {
				processUserCallsSaveUser = true
				break
			}
		}

		if !processUserCallsSaveUser {
			t.Error("Expected saveUser to be called by processUser")
		}
	})
}

func testValidateUserEnrichment(t *testing.T, storage *HybridStorage) {
	t.Run("ValidateUserEnrichment", func(t *testing.T) {
		// validateUser should be called by multiple functions
		callsToValidateUser, err := storage.QueryCallsTo("validateUser")
		if err != nil {
			t.Fatalf("Failed to query calls to validateUser: %v", err)
		}

		// validateUser should be called by: createUser, processUser
		expectedCallers := map[string]struct{}{
			"createUser":  {},
			"processUser": {},
		}

		actualCallers := make(map[string]struct{})
		for _, call := range callsToValidateUser {
			actualCallers[call.Caller] = struct{}{}
		}

		for expectedCaller := range expectedCallers {
			if _, found := actualCallers[expectedCaller]; !found {
				t.Errorf("Expected validateUser to be called by %s", expectedCaller)
			}
		}

		// Should have at least 2 callers
		if len(callsToValidateUser) < 2 {
			t.Errorf("Expected validateUser to have at least 2 callers, got %d", len(callsToValidateUser))
		}
	})
}

func testCrossFileCallStatistics(t *testing.T, storage *HybridStorage) {
	t.Run("CrossFileCallStatistics", func(t *testing.T) {
		// Test cross-file calls
		crossFileCalls := []struct {
			caller string
			callee string
		}{
			{"main", "createUser"},       // main.go -> user.go
			{"main", "processUser"},      // main.go -> user.go
			{"createUser", "generateID"}, // user.go -> utils.go
			{"processUser", "saveUser"},  // user.go -> database.go
		}

		for _, crossCall := range crossFileCalls {
			callsFromCaller, err := storage.QueryCallsFrom(crossCall.caller)
			if err != nil {
				t.Fatalf("Failed to query calls from %s: %v", crossCall.caller, err)
			}

			found := false
			for _, call := range callsFromCaller {
				if call.Callee == crossCall.callee {
					found = true
					// Verify it's actually a cross-file call by checking file paths
					if call.CallerFile == call.File {
						t.Logf("Note: Call from %s to %s appears to be in same file: %s",
							crossCall.caller, crossCall.callee, call.CallerFile)
					}
					break
				}
			}

			if !found {
				t.Errorf("Expected cross-file call from %s to %s", crossCall.caller, crossCall.callee)
			}
		}
	})
}

func TestIndexBuilder_EnrichmentWithFileStructureOutput(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_structure_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a more complex directory structure
	projectStructure := map[string]string{
		"cmd/api/main.go": `package main

import (
	"service/internal/handler"
	"service/internal/config"
)

func main() {
	cfg := config.Load()
	server := handler.NewServer(cfg)
	server.Start()
}`,

		"internal/config/config.go": `package config

type Config struct {
	Port string
	DB   string
}

func Load() *Config {
	return &Config{
		Port: "8080",
		DB:   "postgres://localhost",
	}
}`,

		"internal/handler/server.go": `package handler

import (
	"service/internal/config"
	"service/internal/database"
)

type Server struct {
	config *config.Config
	db     *database.DB
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
		db:     database.Connect(cfg.DB),
	}
}

func (s *Server) Start() error {
	return s.db.Initialize()
}`,

		"internal/database/db.go": `package database

import "service/internal/config"

type DB struct {
	connectionString string
}

func Connect(connStr string) *DB {
	return &DB{connectionString: connStr}
}

func (db *DB) Initialize() error {
	return validateConnection(db.connectionString)
}

func validateConnection(connStr string) error {
	// Connection validation logic
	return nil
}`,
	}

	// Create directory structure and files
	for filePath, content := range projectStructure {
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

	t.Logf("=== COMPLEX PROJECT STRUCTURE RESULTS ===")
	t.Logf("Files processed: %d", stats.FilesProcessed)
	t.Logf("Functions indexed: %d", stats.FunctionsIndexed)
	t.Logf("Types indexed: %d", stats.TypesIndexed)
	t.Logf("Duration: %v", stats.Duration)

	// Verify cross-package dependencies are tracked
	storage := builder.storage

	// Test that main calls functions from other packages
	callsFromMain, err := storage.QueryCallsFrom("main")
	if err != nil {
		t.Fatalf("Failed to query calls from main: %v", err)
	}

	if len(callsFromMain) == 0 {
		t.Error("Expected main to have function calls")
	}

	// Print call graph for inspection
	t.Logf("\n=== CALL GRAPH ANALYSIS ===")
	allFunctions := []string{"main", "Load", "NewServer", "Start", "Connect", "Initialize", "validateConnection"}

	for _, funcName := range allFunctions {
		calls, err := storage.QueryCallsFrom(funcName)
		if err != nil {
			continue // Skip if function not found
		}
		if len(calls) > 0 {
			t.Logf("%s calls:", funcName)
			for _, call := range calls {
				t.Logf("  -> %s (in %s)", call.Callee, call.File)
			}
		}
	}
}
