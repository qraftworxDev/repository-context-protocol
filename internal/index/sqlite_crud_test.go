package index

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"repository-context-protocol/internal/models"
)

func TestSQLiteIndex_InsertIndexEntry(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sqlite_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	index := NewSQLiteIndex(dbPath)

	// Initialize database
	err = index.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer index.Close()

	// Create test index entry
	entry := models.IndexEntry{
		Name:      "TestFunction",
		Type:      "function",
		File:      "main.go",
		StartLine: 10,
		EndLine:   20,
		ChunkID:   "chunk_001",
		Signature: "func TestFunction()",
	}

	// Insert index entry
	err = index.InsertIndexEntry(&entry)
	if err != nil {
		t.Fatalf("Failed to insert index entry: %v", err)
	}

	// Verify entry was inserted
	var count int
	query := "SELECT COUNT(*) FROM index_entries WHERE name = ? AND type = ?"
	err = index.db.QueryRow(query, entry.Name, entry.Type).Scan(&count)
	if err != nil {
		t.Errorf("Failed to query index entry: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 index entry, got %d", count)
	}
}

func TestSQLiteIndex_QueryIndexEntries(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sqlite_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	index := NewSQLiteIndex(dbPath)

	// Initialize database
	err = index.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer index.Close()

	// Insert test entries
	entries := []models.IndexEntry{
		{Name: "Function1", Type: "function", File: "main.go", StartLine: 10, EndLine: 20, ChunkID: "chunk_001", Signature: "func Function1()"},
		{Name: "Function2", Type: "function", File: "utils.go", StartLine: 5, EndLine: 15, ChunkID: "chunk_002", Signature: "func Function2()"},
		{Name: "MyStruct", Type: "struct", File: "types.go", StartLine: 1, EndLine: 10, ChunkID: "chunk_003", Signature: "type MyStruct struct"},
	}

	for i := range entries {
		err = index.InsertIndexEntry(&entries[i])
		if err != nil {
			t.Fatalf("Failed to insert index entry: %v", err)
		}
	}

	// Query by name
	results, err := index.QueryIndexEntries("Function1")
	if err != nil {
		t.Fatalf("Failed to query index entries: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if results[0].Name != "Function1" {
		t.Errorf("Expected name Function1, got %s", results[0].Name)
	}

	// Query by type
	functionResults, err := index.QueryIndexEntriesByType("function")
	if err != nil {
		t.Fatalf("Failed to query index entries by type: %v", err)
	}
	if len(functionResults) != 2 {
		t.Errorf("Expected 2 function results, got %d", len(functionResults))
	}
}

func TestSQLiteIndex_InsertCallRelation(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sqlite_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	index := NewSQLiteIndex(dbPath)

	// Initialize database
	err = index.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer index.Close()

	// Create test call relation
	relation := models.CallRelation{
		Caller:     "main",
		CallerFile: "main.go",
		Callee:     "helper",
		File:       "utils.go",
		Line:       15,
	}

	// Insert call relation
	err = index.InsertCallRelation(relation)
	if err != nil {
		t.Fatalf("Failed to insert call relation: %v", err)
	}

	// Verify relation was inserted
	var count int
	query := "SELECT COUNT(*) FROM call_relations WHERE caller = ? AND callee = ?"
	err = index.db.QueryRow(query, relation.Caller, relation.Callee).Scan(&count)
	if err != nil {
		t.Errorf("Failed to query call relation: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 call relation, got %d", count)
	}
}

func TestSQLiteIndex_QueryCallRelations(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sqlite_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	index := NewSQLiteIndex(dbPath)

	// Initialize database
	err = index.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer index.Close()

	// Insert test call relations
	relations := []models.CallRelation{
		{Caller: "main", CallerFile: "main.go", Callee: "helper1", File: "utils.go", Line: 10},
		{Caller: "main", CallerFile: "main.go", Callee: "helper2", File: "utils.go", Line: 15},
		{Caller: "helper1", CallerFile: "utils.go", Callee: "dbQuery", File: "db.go", Line: 20},
	}

	for _, relation := range relations {
		err = index.InsertCallRelation(relation)
		if err != nil {
			t.Fatalf("Failed to insert call relation: %v", err)
		}
	}

	// Query calls from main
	results, err := index.QueryCallsFrom("main")
	if err != nil {
		t.Fatalf("Failed to query calls from main: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 calls from main, got %d", len(results))
	}

	// Query calls to helper1
	callers, err := index.QueryCallsTo("helper1")
	if err != nil {
		t.Fatalf("Failed to query calls to helper1: %v", err)
	}
	if len(callers) != 1 {
		t.Errorf("Expected 1 caller to helper1, got %d", len(callers))
	}
	if callers[0].Caller != "main" {
		t.Errorf("Expected caller main, got %s", callers[0].Caller)
	}
}

func TestSQLiteIndex_RegisterChunk(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sqlite_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	index := NewSQLiteIndex(dbPath)

	// Initialize database
	err = index.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer index.Close()

	// Register chunk
	chunkID := "chunk_001"
	files := []string{"main.go", "utils.go"}
	tokenCount := 150
	createdAt := time.Now()

	err = index.RegisterChunk(chunkID, files, tokenCount, createdAt)
	if err != nil {
		t.Fatalf("Failed to register chunk: %v", err)
	}

	// Verify chunk was registered
	var count int
	var storedTokenCount int
	query := "SELECT COUNT(*), token_count FROM chunks WHERE chunk_id = ?"
	err = index.db.QueryRow(query, chunkID).Scan(&count, &storedTokenCount)
	if err != nil {
		t.Errorf("Failed to query chunk: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 chunk, got %d", count)
	}
	if storedTokenCount != tokenCount {
		t.Errorf("Expected token count %d, got %d", tokenCount, storedTokenCount)
	}
}

func TestSQLiteIndex_DeleteChunk(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sqlite_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	index := NewSQLiteIndex(dbPath)

	// Initialize database
	err = index.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer index.Close()

	// Register chunk first
	chunkID := "chunk_001"
	files := []string{"main.go"}
	err = index.RegisterChunk(chunkID, files, 100, time.Now())
	if err != nil {
		t.Fatalf("Failed to register chunk: %v", err)
	}

	// Insert related index entries
	entry := models.IndexEntry{
		Name:      "TestFunction",
		Type:      "function",
		File:      "main.go",
		StartLine: 10,
		EndLine:   20,
		ChunkID:   chunkID,
		Signature: "func TestFunction()",
	}
	err = index.InsertIndexEntry(&entry)
	if err != nil {
		t.Fatalf("Failed to insert index entry: %v", err)
	}

	// Delete chunk (should cascade delete related entries)
	err = index.DeleteChunk(chunkID)
	if err != nil {
		t.Fatalf("Failed to delete chunk: %v", err)
	}

	// Verify chunk was deleted
	var chunkCount int
	err = index.db.QueryRow("SELECT COUNT(*) FROM chunks WHERE chunk_id = ?", chunkID).Scan(&chunkCount)
	if err != nil {
		t.Errorf("Failed to query chunks: %v", err)
	}
	if chunkCount != 0 {
		t.Errorf("Expected 0 chunks after deletion, got %d", chunkCount)
	}

	// Verify related index entries were deleted
	var entryCount int
	err = index.db.QueryRow("SELECT COUNT(*) FROM index_entries WHERE chunk_id = ?", chunkID).Scan(&entryCount)
	if err != nil {
		t.Errorf("Failed to query index entries: %v", err)
	}
	if entryCount != 0 {
		t.Errorf("Expected 0 index entries after chunk deletion, got %d", entryCount)
	}
}
