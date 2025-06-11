package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndexBuilder_ComprehensiveModelPopulation(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_models_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a comprehensive Go file that exercises all model fields
	comprehensiveCode := `package demo

import (
	"fmt"
	"context"
	"encoding/json"
)

// Constants - various declaration styles
const (
	MaxUsers = 1000
	DefaultTimeout = 30 * time.Second
	AppVersion = "1.0.0"
)

// Single constant
const DatabaseRetries = 3

// Variables - various declaration styles
var (
	GlobalCounter int
	AppName = "UserService"
	ConfigPath string
)

// Single variable declarations
var DatabaseURL string = "localhost:5432"
var IsProduction bool

// User represents a user entity with JSON tags
type User struct {
	ID       int64     ` + "`json:\"id\"`" + `
	Name     string    ` + "`json:\"name\"`" + `
	Email    string    ` + "`json:\"email\"`" + `
	Active   bool      ` + "`json:\"active\"`" + `
	Metadata map[string]interface{} ` + "`json:\"metadata,omitempty\"`" + `
}

// String method for User - should be detected as a method
func (u User) String() string {
	return fmt.Sprintf("User{ID: %d, Name: %s, Email: %s}", u.ID, u.Name, u.Email)
}

// IsValid method for User
func (u *User) IsValid() bool {
	return u.Name != "" && u.Email != ""
}

// Activate method for User
func (u *User) Activate() {
	u.Active = true
	LogUserAction(u.ID, "activated")
}

// Repository interface with multiple methods
type Repository interface {
	Get(ctx context.Context, id int64) (*User, error)
	Save(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit int) ([]*User, error)
}

// Service interface extending Repository
type Service interface {
	Repository
	Validate(user *User) error
	SendNotification(userID int64, message string) error
}

// UserService struct implementing Service interface
type UserService struct {
	repo Repository
	logger Logger
}

// Logger interface for dependency injection
type Logger interface {
	Info(message string)
	Error(message string)
	Debug(message string)
}

// NewUserService constructor function
func NewUserService(repo Repository, logger Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

// CreateUser creates a new user with validation
func CreateUser(name, email string) (*User, error) {
	if name == "" {
		return nil, ValidateError("name cannot be empty")
	}

	user := &User{
		Name:   name,
		Email:  email,
		Active: false,
	}

	if !user.IsValid() {
		return nil, ValidateError("user validation failed")
	}

	LogUserAction(0, "created")
	return user, nil
}

// ProcessUser processes a user through the service layer
func ProcessUser(user *User) error {
	if user == nil {
		return ValidateError("user cannot be nil")
	}

	// Call local helper
	if err := ValidateUserData(user); err != nil {
		return err
	}

	// Call method on user
	user.Activate()

	// Call external function
	return SaveToDatabase(user)
}

// ValidateUserData validates user data locally
func ValidateUserData(user *User) error {
	if !user.IsValid() {
		return ValidateError("invalid user data")
	}
	return nil
}

// ValidateError creates a validation error
func ValidateError(message string) error {
	return fmt.Errorf("validation error: %s", message)
}

// LogUserAction logs user actions (external dependency)
func LogUserAction(userID int64, action string) {
	fmt.Printf("User %d: %s\n", userID, action)
}

// SaveToDatabase saves user to database (external dependency)
func SaveToDatabase(user *User) error {
	// Simulate database save
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	fmt.Printf("Saving to database: %s\n", string(data))
	return nil
}

// Main function demonstrating usage
func main() {
	// Create user
	user, err := CreateUser("Alice", "alice@example.com")
	if err != nil {
		panic(err)
	}

	// Process user
	err = ProcessUser(user)
	if err != nil {
		panic(err)
	}

	// Print user
	fmt.Println(user.String())
}

// Helper function for main
func init() {
	GlobalCounter = 0
	ConfigPath = "/etc/app/config.json"
}`

	// Write the comprehensive test file
	testFile := filepath.Join(tempDir, "comprehensive.go")
	err = os.WriteFile(testFile, []byte(comprehensiveCode), 0600)
	if err != nil {
		t.Fatalf("Failed to create comprehensive test file: %v", err)
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

	// Print statistics for verification
	t.Logf("=== INDEX BUILDER STATISTICS ===")
	t.Logf("Files processed: %d", stats.FilesProcessed)
	t.Logf("Functions indexed: %d", stats.FunctionsIndexed)
	t.Logf("Types indexed: %d", stats.TypesIndexed)
	t.Logf("Variables indexed: %d", stats.VariablesIndexed)
	t.Logf("Constants indexed: %d", stats.ConstantsIndexed)
	t.Logf("Calls indexed: %d", stats.CallsIndexed)
	t.Logf("Duration: %v", stats.Duration)

	// Verify basic statistics
	if stats.FilesProcessed != 1 {
		t.Errorf("Expected 1 file processed, got %d", stats.FilesProcessed)
	}

	// Test comprehensive model population
	testFunctionModels(t, builder.storage)
	testTypeModels(t, builder.storage)
	testVariableModels(t, builder.storage)
	testConstantModels(t, builder.storage)
	testImportModels(t, builder.storage)
	testCallGraphModels(t, builder.storage)
}

func testFunctionModels(t *testing.T, storage *HybridStorage) {
	t.Run("Functions", func(t *testing.T) {
		// Test that all expected functions are indexed
		expectedFunctions := []string{
			"NewUserService", "CreateUser", "ProcessUser", "ValidateUserData",
			"ValidateError", "LogUserAction", "SaveToDatabase", "main", "init",
		}

		for _, funcName := range expectedFunctions {
			results, err := storage.QueryByName(funcName)
			if err != nil {
				t.Fatalf("Failed to query for function %s: %v", funcName, err)
			}
			if len(results) != 1 {
				t.Errorf("Expected 1 result for function %s, got %d", funcName, len(results))
				continue
			}

			// Verify function has proper metadata
			if results[0].IndexEntry.Type != "function" {
				t.Errorf("Expected function %s to have type 'function', got %s", funcName, results[0].IndexEntry.Type)
			}
			if results[0].IndexEntry.Name != funcName {
				t.Errorf("Expected function name %s, got %s", funcName, results[0].IndexEntry.Name)
			}
		}

		// Test method detection on types
		expectedMethods := []string{"String", "IsValid", "Activate"}
		for _, methodName := range expectedMethods {
			results, err := storage.QueryByName(methodName)
			if err != nil {
				t.Fatalf("Failed to query for method %s: %v", methodName, err)
			}
			if len(results) != 1 {
				t.Errorf("Expected 1 result for method %s, got %d", methodName, len(results))
			}
		}
	})
}

func testTypeModels(t *testing.T, storage *HybridStorage) {
	t.Run("Types", func(t *testing.T) {
		// Test struct types
		userResults, err := storage.QueryByName("User")
		if err != nil {
			t.Fatalf("Failed to query for User type: %v", err)
		}
		if len(userResults) != 1 {
			t.Fatalf("Expected 1 User type result, got %d", len(userResults))
		}

		// Test interface types
		expectedInterfaces := []string{"Repository", "Service", "Logger"}
		for _, interfaceName := range expectedInterfaces {
			var results []QueryResult
			results, err = storage.QueryByName(interfaceName)
			if err != nil {
				t.Fatalf("Failed to query for interface %s: %v", interfaceName, err)
			}
			if len(results) != 1 {
				t.Errorf("Expected 1 result for interface %s, got %d", interfaceName, len(results))
			}
			if results[0].IndexEntry.Type != "interface" {
				t.Errorf("Expected interface %s to have type 'interface', got %s", interfaceName, results[0].IndexEntry.Type)
			}
		}

		// Test struct with methods
		var serviceResults []QueryResult
		serviceResults, err = storage.QueryByName("UserService")
		if err != nil {
			t.Fatalf("Failed to query for UserService type: %v", err)
		}
		if len(serviceResults) != 1 {
			t.Errorf("Expected 1 UserService type result, got %d", len(serviceResults))
		}
	})
}

func testVariableModels(t *testing.T, storage *HybridStorage) {
	t.Run("Variables", func(t *testing.T) {
		expectedVariables := []string{"GlobalCounter", "AppName", "ConfigPath", "DatabaseURL", "IsProduction"}

		for _, varName := range expectedVariables {
			results, err := storage.QueryByName(varName)
			if err != nil {
				t.Fatalf("Failed to query for variable %s: %v", varName, err)
			}
			if len(results) != 1 {
				t.Errorf("Expected 1 result for variable %s, got %d", varName, len(results))
				continue
			}

			if results[0].IndexEntry.Type != "variable" {
				t.Errorf("Expected variable %s to have type 'variable', got %s", varName, results[0].IndexEntry.Type)
			}
		}
	})
}

func testConstantModels(t *testing.T, storage *HybridStorage) {
	t.Run("Constants", func(t *testing.T) {
		expectedConstants := []string{"MaxUsers", "DefaultTimeout", "AppVersion", "DatabaseRetries"}

		for _, constName := range expectedConstants {
			results, err := storage.QueryByName(constName)
			if err != nil {
				t.Fatalf("Failed to query for constant %s: %v", constName, err)
			}
			if len(results) != 1 {
				t.Errorf("Expected 1 result for constant %s, got %d", constName, len(results))
				continue
			}

			if results[0].IndexEntry.Type != "constant" {
				t.Errorf("Expected constant %s to have type 'constant', got %s", constName, results[0].IndexEntry.Type)
			}
		}
	})
}

func testImportModels(t *testing.T, storage *HybridStorage) {
	t.Run("Imports", func(t *testing.T) {
		// Test that imports are tracked in call relationships
		// Since imports create dependencies, they should show up in call graph

		// Query for functions that use external packages
		mainResults, err := storage.QueryByName("main")
		if err != nil {
			t.Fatalf("Failed to query for main function: %v", err)
		}
		if len(mainResults) != 1 {
			t.Fatalf("Expected 1 main function result, got %d", len(mainResults))
		}

		// Verify that external package calls are tracked
		callsFromMain, err := storage.QueryCallsFrom("main")
		if err != nil {
			t.Fatalf("Failed to query calls from main: %v", err)
		}

		// main should call CreateUser, ProcessUser, fmt.Println
		expectedMinCalls := 3
		if len(callsFromMain) < expectedMinCalls {
			t.Errorf("Expected at least %d calls from main, got %d", expectedMinCalls, len(callsFromMain))
		}
	})
}

func testCallGraphModels(t *testing.T, storage *HybridStorage) {
	t.Run("CallGraph", func(t *testing.T) {
		// Test that call relationships are properly established

		// CreateUser should call ValidateError and LogUserAction
		callsFromCreateUser, err := storage.QueryCallsFrom("CreateUser")
		if err != nil {
			t.Fatalf("Failed to query calls from CreateUser: %v", err)
		}

		expectedCallTargets := map[string]bool{
			"ValidateError": false,
			"LogUserAction": false,
		}

		for _, call := range callsFromCreateUser {
			if _, exists := expectedCallTargets[call.Callee]; exists {
				expectedCallTargets[call.Callee] = true
			}
		}

		for target, found := range expectedCallTargets {
			if !found {
				t.Errorf("Expected CreateUser to call %s", target)
			}
		}

		// ProcessUser should call ValidateUserData, user.Activate, and SaveToDatabase
		callsFromProcessUser, err := storage.QueryCallsFrom("ProcessUser")
		if err != nil {
			t.Fatalf("Failed to query calls from ProcessUser: %v", err)
		}

		processUserTargets := map[string]bool{
			"ValidateUserData": false,
			"SaveToDatabase":   false,
		}

		for _, call := range callsFromProcessUser {
			if _, exists := processUserTargets[call.Callee]; exists {
				processUserTargets[call.Callee] = true
			}
		}

		for target, found := range processUserTargets {
			if !found {
				t.Errorf("Expected ProcessUser to call %s", target)
			}
		}

		// Test reverse relationships - ValidateError should be called by CreateUser and ValidateUserData
		callsToValidateError, err := storage.QueryCallsTo("ValidateError")
		if err != nil {
			t.Fatalf("Failed to query calls to ValidateError: %v", err)
		}

		if len(callsToValidateError) < 2 {
			t.Errorf("Expected at least 2 calls to ValidateError, got %d", len(callsToValidateError))
		}

		callers := make(map[string]bool)
		for _, call := range callsToValidateError {
			callers[call.Caller] = true
		}

		expectedCallers := []string{"CreateUser", "ValidateUserData"}
		for _, caller := range expectedCallers {
			if !callers[caller] {
				t.Errorf("Expected ValidateError to be called by %s", caller)
			}
		}
	})
}

func TestIndexBuilder_ModelFieldCompleteness(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_completeness_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a file that exercises model field completeness
	testCode := `package completeness

import "fmt"

const TestConst = "test value"
var TestVar int = 42

type TestStruct struct {
	Field string
}

func (t TestStruct) Method() string {
	return t.Field
}

func TestFunction(param string) (string, error) {
	fmt.Println(param)
	return param, nil
}`

	// Write the test file
	testFile := filepath.Join(tempDir, "completeness.go")
	err = os.WriteFile(testFile, []byte(testCode), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create and initialize IndexBuilder
	builder := NewIndexBuilder(tempDir)
	err = builder.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize IndexBuilder: %v", err)
	}
	defer builder.Close()

	// Build the index
	_, err = builder.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Query the indexed file context to verify completeness
	// Note: We would need to add a method to query the full FileContext
	// For now, we test that all entities are queryable

	entities := []struct {
		name     string
		expected string
	}{
		{"TestConst", "constant"},
		{"TestVar", "variable"},
		{"TestStruct", "struct"},
		{"TestFunction", "function"},
		{"Method", "function"},
	}

	for _, entity := range entities {
		results, err := builder.storage.QueryByName(entity.name)
		if err != nil {
			t.Fatalf("Failed to query for %s: %v", entity.name, err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result for %s, got %d", entity.name, len(results))
			continue
		}
		if results[0].IndexEntry.Type != entity.expected {
			t.Errorf("Expected %s to have type %s, got %s", entity.name, entity.expected, results[0].IndexEntry.Type)
		}
	}
}
