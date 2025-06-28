package index

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"repository-context-protocol/internal/models"
)

const (
	constFileWritePermissionMode = 0600
)

func TestHybridStorage_GetFunctionImplementation(t *testing.T) {
	tempDir := t.TempDir()

	storage := NewHybridStorage(tempDir)
	err := storage.Initialize()
	require.NoError(t, err, "Failed to initialize storage")
	defer storage.Close()

	// Create a test source file
	testFileContent := `package test

import "fmt"

// TestFunction is a sample function for testing
func TestFunction(name string) string {
	if name == "" {
		return "empty"
	}
	return fmt.Sprintf("Hello, %s!", name)
}

// AnotherFunction is another test function
func AnotherFunction() int {
	x := 42
	return x * 2
}
`

	testFilePath := filepath.Join(tempDir, "test.go")
	err = os.WriteFile(testFilePath, []byte(testFileContent), constFileWritePermissionMode)
	require.NoError(t, err, "Failed to write test file")

	// Create file context with function data
	fileContext := &models.FileContext{
		Path: testFilePath,
		Functions: []models.Function{
			{
				Name:      "TestFunction",
				Signature: "func TestFunction(name string) string",
				StartLine: 6,
				EndLine:   10,
				Parameters: []models.Parameter{
					{Name: "name", Type: "string"},
				},
				Returns: []models.Type{
					{Name: "", Kind: "string"},
				},
			},
			{
				Name:      "AnotherFunction",
				Signature: "func AnotherFunction() int",
				StartLine: 13,
				EndLine:   16,
				Returns: []models.Type{
					{Name: "", Kind: "int"},
				},
			},
		},
	}

	// Store the file context
	err = storage.StoreFileContext(fileContext)
	require.NoError(t, err, "Failed to store file context")

	// Test extracting TestFunction implementation
	impl, err := storage.GetFunctionImplementation("TestFunction", 2)
	require.NoError(t, err, "Failed to get function implementation")
	assert.NotNil(t, impl, "Implementation should not be nil")

	// Verify the function body contains the expected content
	assert.Contains(t, impl.Body, "if name == \"\"", "Function body should contain the condition")
	assert.Contains(t, impl.Body, "return \"empty\"", "Function body should contain the return statement")
	assert.Contains(t, impl.Body, "fmt.Sprintf", "Function body should contain the fmt.Sprintf call")

	// Verify context lines are provided
	assert.NotEmpty(t, impl.ContextLines, "Context lines should not be empty")

	// Test extracting AnotherFunction implementation
	impl2, err := storage.GetFunctionImplementation("AnotherFunction", 1)
	require.NoError(t, err, "Failed to get AnotherFunction implementation")
	assert.NotNil(t, impl2, "Implementation should not be nil")

	// Verify the function body
	assert.Contains(t, impl2.Body, "x := 42", "Function body should contain variable assignment")
	assert.Contains(t, impl2.Body, "return x * 2", "Function body should contain the return statement")

	// Test non-existent function
	_, err = storage.GetFunctionImplementation("NonExistentFunction", 2)
	assert.Error(t, err, "Should return error for non-existent function")
	assert.Contains(t, err.Error(), "not found", "Error should indicate function not found")
}
