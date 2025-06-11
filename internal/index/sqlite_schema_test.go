package index

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteIndex_InitializeDatabase(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sqlite_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	index := &SQLiteIndex{
		dbPath: dbPath,
	}

	// Initialize database
	err = index.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer index.Close()

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Expected database file to be created at %s", dbPath)
	}

	// Verify database connection is working
	if index.db == nil {
		t.Error("Expected database connection to be established")
	}
}

func TestSQLiteIndex_CreateTables(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sqlite_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	index := &SQLiteIndex{
		dbPath: dbPath,
	}

	// Initialize database
	err = index.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer index.Close()

	// Verify tables were created
	tables := []string{"index_entries", "call_relations", "chunks"}
	for _, table := range tables {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		err := index.db.QueryRow(query, table).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("Expected table %s to exist, but it doesn't", table)
		}
	}
}

func TestSQLiteIndex_TableSchema(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sqlite_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	index := &SQLiteIndex{
		dbPath: dbPath,
	}

	// Initialize database
	err = index.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer index.Close()

	// Test index_entries table schema
	rows, err := index.db.Query("PRAGMA table_info(index_entries)")
	if err != nil {
		t.Fatalf("Failed to get table info: %v", err)
	}
	defer rows.Close()

	expectedColumns := map[string]bool{
		"id":         false,
		"name":       false,
		"type":       false,
		"file_path":  false,
		"start_line": false,
		"end_line":   false,
		"chunk_id":   false,
	}

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			t.Errorf("Failed to scan column info: %v", err)
			continue
		}

		if _, exists := expectedColumns[name]; exists {
			expectedColumns[name] = true
		}
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		t.Errorf("Error iterating over rows: %v", err)
	}

	// Verify all expected columns exist
	for column, found := range expectedColumns {
		if !found {
			t.Errorf("Expected column %s not found in index_entries table", column)
		}
	}
}

func TestSQLiteIndex_Close(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sqlite_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	index := &SQLiteIndex{
		dbPath: dbPath,
	}

	// Initialize database
	err = index.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Close database
	err = index.Close()
	if err != nil {
		t.Errorf("Failed to close database: %v", err)
	}

	// Verify connection is closed (attempting to use it should fail)
	err = index.db.Ping()
	if err == nil {
		t.Error("Expected error when using closed database connection")
	}
}

func TestSQLiteIndex_InvalidPath(t *testing.T) {
	// Use an invalid database path
	index := &SQLiteIndex{
		dbPath: "/invalid/readonly/path/test.db",
	}

	// Should fail to initialize
	err := index.Initialize()
	if err == nil {
		t.Error("Expected error when initializing database with invalid path")
	}
}
