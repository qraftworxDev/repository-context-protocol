package golang

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
	"time"

	"repository-context-protocol/internal/models"
)

func TestGoParser_ChecksumGeneration(t *testing.T) {
	parser := NewGoParser()

	code1 := `package main
func hello() {
	fmt.Println("Hello")
}`

	code2 := `package main
func hello() {
	fmt.Println("World")
}`

	// Parse the same content twice
	result1a, err := parser.ParseFile("test1.go", []byte(code1))
	if err != nil {
		t.Fatalf("Failed to parse code1: %v", err)
	}

	result1b, err := parser.ParseFile("test1.go", []byte(code1))
	if err != nil {
		t.Fatalf("Failed to parse code1 again: %v", err)
	}

	// Parse different content
	result2, err := parser.ParseFile("test2.go", []byte(code2))
	if err != nil {
		t.Fatalf("Failed to parse code2: %v", err)
	}

	// Same content should have same checksum
	if result1a.Checksum != result1b.Checksum {
		t.Errorf("Same content should have same checksum: %s != %s", result1a.Checksum, result1b.Checksum)
	}

	// Different content should have different checksum
	if result1a.Checksum == result2.Checksum {
		t.Errorf("Different content should have different checksum: %s == %s", result1a.Checksum, result2.Checksum)
	}

	// Verify checksum is not empty
	if result1a.Checksum == "" {
		t.Error("Checksum should not be empty")
	}

	// Verify checksum format (SHA-256 should be 64 hex characters)
	if len(result1a.Checksum) != 64 {
		t.Errorf("SHA-256 checksum should be 64 characters, got %d", len(result1a.Checksum))
	}

	// Verify checksum matches manual calculation
	hash := sha256.Sum256([]byte(code1))
	expectedChecksum := fmt.Sprintf("%x", hash)
	if result1a.Checksum != expectedChecksum {
		t.Errorf("Checksum mismatch: expected %s, got %s", expectedChecksum, result1a.Checksum)
	}
}

func TestGoParser_ModificationTime(t *testing.T) {
	parser := NewGoParser()

	code := `package main
func test() {}`

	beforeParse := time.Now()

	// Parse in-memory content (should use current time)
	result, err := parser.ParseFile("memory.go", []byte(code))
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	afterParse := time.Now()

	// ModTime should not be zero
	if result.ModTime.IsZero() {
		t.Error("ModTime should not be zero")
	}

	// ModTime should be between before and after parse time (for in-memory parsing)
	if result.ModTime.Before(beforeParse) || result.ModTime.After(afterParse) {
		t.Errorf("ModTime %v should be between %v and %v", result.ModTime, beforeParse, afterParse)
	}
}

func TestGoParser_RealFileModificationTime(t *testing.T) {
	parser := NewGoParser()

	// Test with the actual test file
	testFile := "../../../testdata/simple-go/main.go"

	// Read the file content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Skipf("Skipping real file test, file not found: %v", err)
	}

	// Parse the real file
	result, err := parser.ParseFile(testFile, content)
	if err != nil {
		t.Fatalf("Failed to parse real file: %v", err)
	}

	// Get the actual file modification time
	fileInfo, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	// ModTime should match the file's actual modification time
	if !result.ModTime.Equal(fileInfo.ModTime()) {
		t.Errorf("ModTime mismatch: expected %v, got %v", fileInfo.ModTime(), result.ModTime)
	}

	// Checksum should be consistent
	if result.Checksum == "" {
		t.Error("Checksum should not be empty for real file")
	}
}

func TestGoParser_MetadataConsistency(t *testing.T) {
	parser := NewGoParser()

	code := `package test
import "fmt"
func greet(name string) {
	fmt.Printf("Hello, %s!", name)
}`

	// Parse multiple times
	results := make([]*models.FileContext, 3)
	for i := 0; i < 3; i++ {
		result, err := parser.ParseFile("consistency.go", []byte(code))
		if err != nil {
			t.Fatalf("Failed to parse code iteration %d: %v", i, err)
		}
		results[i] = result
	}

	// All checksums should be identical
	for i := 1; i < len(results); i++ {
		if results[0].Checksum != results[i].Checksum {
			t.Errorf("Checksum inconsistency: %s != %s", results[0].Checksum, results[i].Checksum)
		}
	}

	// All should have valid modification times
	for i, result := range results {
		if result.ModTime.IsZero() {
			t.Errorf("ModTime is zero for iteration %d", i)
		}
	}
}
