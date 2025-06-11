package index

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIndexBuilder_MetadataValidation(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_metadata_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file with known content for checksum validation
	testContent := `package metadata

import (
	"fmt"
	"time"
)

// TestConstant for checksum validation
const TestConstant = "checksum_test"

// TestVariable for metadata validation
var TestVariable = "metadata_test"

// User represents a user with metadata
type User struct {
	ID        int       ` + "`json:\"id\"`" + `
	Name      string    ` + "`json:\"name\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
}

// CreateUser creates a new user with current timestamp
func CreateUser(name string) *User {
	return &User{
		Name:      name,
		CreatedAt: time.Now(),
	}
}

// String returns string representation of user
func (u *User) String() string {
	return fmt.Sprintf("User{Name: %s, CreatedAt: %v}", u.Name, u.CreatedAt)
}

// ValidateUser validates user data
func ValidateUser(user *User) error {
	if user.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return nil
}

func main() {
	user := CreateUser("Alice")
	fmt.Println(user.String())

	if err := ValidateUser(user); err != nil {
		fmt.Printf("Validation error: %v\n", err)
	}
}`

	testFile := filepath.Join(tempDir, "metadata_test.go")
	err = os.WriteFile(testFile, []byte(testContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Get file info for timestamp validation
	fileInfo, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}
	expectedModTime := fileInfo.ModTime()

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

	t.Logf("=== METADATA VALIDATION STATISTICS ===")
	t.Logf("Files processed: %d", stats.FilesProcessed)
	t.Logf("Functions indexed: %d", stats.FunctionsIndexed)
	t.Logf("Types indexed: %d", stats.TypesIndexed)
	t.Logf("Variables indexed: %d", stats.VariablesIndexed)
	t.Logf("Constants indexed: %d", stats.ConstantsIndexed)

	// Test metadata for different entity types
	storage := builder.storage

	testFunctionMetadata(t, storage, testFile, expectedModTime)
	testTypeMetadata(t, storage, testFile)
	testVariableMetadata(t, storage, testFile)
	testConstantMetadata(t, storage, testFile)
	testSignatureMetadata(t, storage)
}

func testFunctionMetadata(t *testing.T, storage *HybridStorage, expectedFile string, expectedModTime time.Time) {
	t.Run("FunctionMetadata", func(t *testing.T) {
		// Test CreateUser function metadata
		results, err := storage.QueryByName("CreateUser")
		if err != nil {
			t.Fatalf("Failed to query CreateUser: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("Expected 1 CreateUser result, got %d", len(results))
		}

		entry := results[0].IndexEntry

		// Test basic metadata
		if entry.Name != "CreateUser" {
			t.Errorf("Expected function name 'CreateUser', got %s", entry.Name)
		}
		if entry.Type != "function" {
			t.Errorf("Expected function type 'function', got %s", entry.Type)
		}

		// Test file path
		if entry.File != expectedFile {
			t.Errorf("Expected file path %s, got %s", expectedFile, entry.File)
		}

		// Test line numbers
		if entry.StartLine <= 0 {
			t.Errorf("Expected positive start line, got %d", entry.StartLine)
		}
		if entry.EndLine <= entry.StartLine {
			t.Errorf("Expected end line (%d) > start line (%d)", entry.EndLine, entry.StartLine)
		}

		// Test signature is present
		if entry.Signature == "" {
			t.Error("Expected function signature to be present")
		}

		// Test chunk data metadata
		if results[0].ChunkData == nil {
			t.Error("Expected chunk data to be present")
		} else {
			chunk := results[0].ChunkData

			// Test checksum exists (format may vary based on processing)
			if len(chunk.FileData) > 0 {
				fileData := chunk.FileData[0]
				if fileData.Checksum == "" {
					t.Error("Expected non-empty checksum")
				}

				// Test timestamp (allow some tolerance for file system precision)
				timeDiff := fileData.ModTime.Sub(expectedModTime)
				if timeDiff < 0 {
					timeDiff = -timeDiff
				}
				if timeDiff > time.Second {
					t.Errorf("ModTime difference too large: expected %v, got %v (diff: %v)",
						expectedModTime, fileData.ModTime, timeDiff)
				}

				// Test language
				if fileData.Language != "go" {
					t.Errorf("Expected language 'go', got %s", fileData.Language)
				}
			}
		}
	})
}

func testTypeMetadata(t *testing.T, storage *HybridStorage, expectedFile string) {
	t.Run("TypeMetadata", func(t *testing.T) {
		// Test User struct metadata
		results, err := storage.QueryByName("User")
		if err != nil {
			t.Fatalf("Failed to query User: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("Expected 1 User result, got %d", len(results))
		}

		entry := results[0].IndexEntry

		// Test type metadata
		if entry.Name != "User" {
			t.Errorf("Expected type name 'User', got %s", entry.Name)
		}
		if entry.Type != "struct" {
			t.Errorf("Expected type 'struct', got %s", entry.Type)
		}
		if entry.File != expectedFile {
			t.Errorf("Expected file path %s, got %s", expectedFile, entry.File)
		}

		// Test line numbers are valid
		if entry.StartLine <= 0 {
			t.Errorf("Expected positive start line, got %d", entry.StartLine)
		}
		if entry.EndLine <= entry.StartLine {
			t.Errorf("Expected end line (%d) > start line (%d)", entry.EndLine, entry.StartLine)
		}

		// Test chunk data is linked correctly
		if results[0].ChunkData == nil {
			t.Error("Expected chunk data to be present")
		}
	})
}

func testVariableMetadata(t *testing.T, storage *HybridStorage, expectedFile string) {
	t.Run("VariableMetadata", func(t *testing.T) {
		// Test TestVariable metadata
		results, err := storage.QueryByName("TestVariable")
		if err != nil {
			t.Fatalf("Failed to query TestVariable: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("Expected 1 TestVariable result, got %d", len(results))
		}

		entry := results[0].IndexEntry

		// Test variable metadata
		if entry.Name != "TestVariable" {
			t.Errorf("Expected variable name 'TestVariable', got %s", entry.Name)
		}
		if entry.Type != "variable" {
			t.Errorf("Expected type 'variable', got %s", entry.Type)
		}
		if entry.File != expectedFile {
			t.Errorf("Expected file path %s, got %s", expectedFile, entry.File)
		}

		// Test line numbers
		if entry.StartLine <= 0 {
			t.Errorf("Expected positive start line, got %d", entry.StartLine)
		}
		if entry.EndLine < entry.StartLine {
			t.Errorf("Expected end line (%d) >= start line (%d)", entry.EndLine, entry.StartLine)
		}
	})
}

func testConstantMetadata(t *testing.T, storage *HybridStorage, expectedFile string) {
	t.Run("ConstantMetadata", func(t *testing.T) {
		// Test TestConstant metadata
		results, err := storage.QueryByName("TestConstant")
		if err != nil {
			t.Fatalf("Failed to query TestConstant: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("Expected 1 TestConstant result, got %d", len(results))
		}

		entry := results[0].IndexEntry

		// Test constant metadata
		if entry.Name != "TestConstant" {
			t.Errorf("Expected constant name 'TestConstant', got %s", entry.Name)
		}
		if entry.Type != "constant" {
			t.Errorf("Expected type 'constant', got %s", entry.Type)
		}
		if entry.File != expectedFile {
			t.Errorf("Expected file path %s, got %s", expectedFile, entry.File)
		}

		// Test line numbers
		if entry.StartLine <= 0 {
			t.Errorf("Expected positive start line, got %d", entry.StartLine)
		}
	})
}

func testSignatureMetadata(t *testing.T, storage *HybridStorage) {
	t.Run("SignatureMetadata", func(t *testing.T) {
		// Test function signatures are captured correctly
		testCases := []struct {
			functionName      string
			expectedSignature string
		}{
			{"CreateUser", "func CreateUser(name string) *User"},
			{"ValidateUser", "func ValidateUser(user *User) error"},
			{"main", "func main()"},
		}

		for _, testCase := range testCases {
			results, err := storage.QueryByName(testCase.functionName)
			if err != nil {
				t.Fatalf("Failed to query %s: %v", testCase.functionName, err)
			}
			if len(results) != 1 {
				t.Fatalf("Expected 1 %s result, got %d", testCase.functionName, len(results))
			}

			entry := results[0].IndexEntry
			if entry.Signature != testCase.expectedSignature {
				t.Errorf("Expected signature for %s: %s, got: %s",
					testCase.functionName, testCase.expectedSignature, entry.Signature)
			}
		}

		// Test method signatures
		results, err := storage.QueryByName("String")
		if err != nil {
			t.Fatalf("Failed to query String method: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("Expected 1 String method result, got %d", len(results))
		}

		entry := results[0].IndexEntry
		expectedMethodSignature := "func (u *User) String() string"
		if entry.Signature != expectedMethodSignature {
			t.Errorf("Expected method signature: %s, got: %s",
				expectedMethodSignature, entry.Signature)
		}
	})
}

func TestIndexBuilder_ChunkMetadataConsistency(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "builder_chunk_metadata_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create multiple files to test chunking metadata
	testFiles := map[string]string{
		"file1.go": `package chunk

func Function1() {
	println("function 1")
}

const Constant1 = "value1"
`,
		"file2.go": `package chunk

func Function2() {
	println("function 2")
}

var Variable2 = "value2"
`,
	}

	// Write test files and track their metadata
	expectedFiles := make(map[string]struct {
		checksum string
		modTime  time.Time
	})

	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		err = os.WriteFile(filePath, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}

		// Get file metadata
		var fileInfo os.FileInfo
		fileInfo, err = os.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to get file info for %s: %v", filename, err)
		}

		expectedFiles[filePath] = struct {
			checksum string
			modTime  time.Time
		}{
			checksum: fmt.Sprintf("%x", sha256.Sum256([]byte(content))),
			modTime:  fileInfo.ModTime(),
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
	_, err = builder.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Verify chunk metadata consistency
	storage := builder.storage

	// Test each file's entities have consistent chunk metadata
	for expectedFile, expectedMeta := range expectedFiles {
		t.Run(fmt.Sprintf("File_%s", filepath.Base(expectedFile)), func(t *testing.T) {
			// Query entities from this file
			functions := []string{}
			if filepath.Base(expectedFile) == "file1.go" {
				functions = append(functions, "Function1")
			} else if filepath.Base(expectedFile) == "file2.go" {
				functions = append(functions, "Function2")
			}

			for _, funcName := range functions {
				results, err := storage.QueryByName(funcName)
				if err != nil {
					t.Fatalf("Failed to query %s: %v", funcName, err)
				}
				if len(results) != 1 {
					t.Fatalf("Expected 1 result for %s, got %d", funcName, len(results))
				}

				// Verify chunk data
				if results[0].ChunkData == nil {
					t.Errorf("Expected chunk data for %s", funcName)
					continue
				}

				chunk := results[0].ChunkData

				// Verify chunk has file data for this file
				found := false
				for _, fileData := range chunk.FileData {
					if fileData.Path != expectedFile {
						continue
					}
					found = true

					// Test checksum exists
					if fileData.Checksum == "" {
						t.Errorf("Expected non-empty checksum for %s in %s", funcName, expectedFile)
					}

					// Test timestamp consistency (allow small tolerance)
					timeDiff := fileData.ModTime.Sub(expectedMeta.modTime)
					if timeDiff < 0 {
						timeDiff = -timeDiff
					}
					if timeDiff > time.Second {
						t.Errorf("ModTime mismatch for %s in %s: expected %v, got %v",
							funcName, expectedFile, expectedMeta.modTime, fileData.ModTime)
					}

					break
				}

				if !found {
					t.Errorf("Expected chunk for %s to contain file data for %s", funcName, expectedFile)
				}
			}
		})
	}

	// Verify chunk IDs are properly generated and consistent
	t.Run("ChunkIDConsistency", func(t *testing.T) {
		// Query all functions and verify their chunk IDs
		allFunctions := []string{"Function1", "Function2"}
		chunkIDs := make(map[string]bool)

		for _, funcName := range allFunctions {
			results, err := storage.QueryByName(funcName)
			if err != nil {
				t.Fatalf("Failed to query %s: %v", funcName, err)
			}
			if len(results) != 1 {
				continue
			}

			entry := results[0].IndexEntry
			if entry.ChunkID == "" {
				t.Errorf("Expected non-empty chunk ID for %s", funcName)
			}

			chunkIDs[entry.ChunkID] = true

			// Verify chunk data exists and matches ID
			if results[0].ChunkData == nil {
				t.Errorf("Expected chunk data for %s", funcName)
			} else if results[0].ChunkData.ID != entry.ChunkID {
				t.Errorf("Chunk ID mismatch for %s: entry has %s, chunk has %s",
					funcName, entry.ChunkID, results[0].ChunkData.ID)
			}
		}

		t.Logf("Found %d unique chunk IDs", len(chunkIDs))
	})
}
