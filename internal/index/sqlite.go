package index

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"repository-context-protocol/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteIndex handles SQLite database operations for fast lookups
type SQLiteIndex struct {
	dbPath string
	db     *sql.DB
}

// NewSQLiteIndex creates a new SQLite index with the specified database path
func NewSQLiteIndex(dbPath string) *SQLiteIndex {
	return &SQLiteIndex{
		dbPath: dbPath,
	}
}

// Initialize opens the database connection and creates tables
func (si *SQLiteIndex) Initialize() error {
	var err error
	si.db, err = sql.Open("sqlite3", si.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := si.db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables
	if err := si.createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// createTables creates the necessary tables for the index
func (si *SQLiteIndex) createTables() error {
	// Create index_entries table
	indexEntriesSQL := `
	CREATE TABLE IF NOT EXISTS index_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		file_path TEXT NOT NULL,
		start_line INTEGER NOT NULL,
		end_line INTEGER NOT NULL,
		chunk_id TEXT NOT NULL,
		signature TEXT,
		FOREIGN KEY (chunk_id) REFERENCES chunks(chunk_id) ON DELETE CASCADE
	);`

	if _, err := si.db.Exec(indexEntriesSQL); err != nil {
		return fmt.Errorf("failed to create index_entries table: %w", err)
	}

	// Create call_relations table
	callRelationsSQL := `
	CREATE TABLE IF NOT EXISTS call_relations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		caller TEXT NOT NULL,
		callee TEXT NOT NULL,
		file TEXT NOT NULL,
		line INTEGER NOT NULL,
		caller_file TEXT NOT NULL
	);`

	if _, err := si.db.Exec(callRelationsSQL); err != nil {
		return fmt.Errorf("failed to create call_relations table: %w", err)
	}

	// Create chunks table
	chunksSQL := `
	CREATE TABLE IF NOT EXISTS chunks (
		chunk_id TEXT PRIMARY KEY,
		files TEXT NOT NULL,
		token_count INTEGER NOT NULL,
		created_at DATETIME NOT NULL
	);`

	if _, err := si.db.Exec(chunksSQL); err != nil {
		return fmt.Errorf("failed to create chunks table: %w", err)
	}

	// Create indexes for better performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_index_entries_name ON index_entries(name);",
		"CREATE INDEX IF NOT EXISTS idx_index_entries_type ON index_entries(type);",
		"CREATE INDEX IF NOT EXISTS idx_index_entries_file ON index_entries(file_path);",
		"CREATE INDEX IF NOT EXISTS idx_index_entries_chunk ON index_entries(chunk_id);",
		"CREATE INDEX IF NOT EXISTS idx_call_relations_caller ON call_relations(caller);",
		"CREATE INDEX IF NOT EXISTS idx_call_relations_callee ON call_relations(callee);",
		"CREATE INDEX IF NOT EXISTS idx_chunks_created_at ON chunks(created_at);",
	}

	for _, indexSQL := range indexes {
		if _, err := si.db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// InsertIndexEntry inserts a new index entry into the database
func (si *SQLiteIndex) InsertIndexEntry(entry *models.IndexEntry) error {
	query := `
	INSERT INTO index_entries (name, type, file_path, start_line, end_line, chunk_id, signature)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := si.db.Exec(query, entry.Name, entry.Type, entry.File, entry.StartLine, entry.EndLine, entry.ChunkID, entry.Signature)
	if err != nil {
		return fmt.Errorf("failed to insert index entry: %w", err)
	}

	return nil
}

// scanIndexEntries is a helper function to scan index entry rows
func (si *SQLiteIndex) scanIndexEntries(rows *sql.Rows) ([]models.IndexEntry, error) {
	var entries []models.IndexEntry
	for rows.Next() {
		var entry models.IndexEntry
		err := rows.Scan(&entry.Name, &entry.Type, &entry.File, &entry.StartLine, &entry.EndLine, &entry.ChunkID, &entry.Signature)
		if err != nil {
			return nil, fmt.Errorf("failed to scan index entry: %w", err)
		}
		entries = append(entries, entry)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return entries, nil
}

// QueryIndexEntries queries index entries by name
func (si *SQLiteIndex) QueryIndexEntries(name string) ([]models.IndexEntry, error) {
	query := `
	SELECT name, type, file_path, start_line, end_line, chunk_id, signature
	FROM index_entries
	WHERE name = ?`

	rows, err := si.db.Query(query, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query index entries: %w", err)
	}
	defer rows.Close()

	return si.scanIndexEntries(rows)
}

// QueryIndexEntriesByType queries index entries by type
func (si *SQLiteIndex) QueryIndexEntriesByType(entryType string) ([]models.IndexEntry, error) {
	query := `
	SELECT name, type, file_path, start_line, end_line, chunk_id, signature
	FROM index_entries
	WHERE type = ?`

	rows, err := si.db.Query(query, entryType)
	if err != nil {
		return nil, fmt.Errorf("failed to query index entries by type: %w", err)
	}
	defer rows.Close()

	return si.scanIndexEntries(rows)
}

// InsertCallRelation inserts a new call relation into the database
func (si *SQLiteIndex) InsertCallRelation(relation models.CallRelation) error {
	query := `
	INSERT INTO call_relations (caller, callee, file, line, caller_file)
	VALUES (?, ?, ?, ?, ?)`

	_, err := si.db.Exec(query, relation.Caller, relation.Callee, relation.File, relation.Line, relation.CallerFile)
	if err != nil {
		return fmt.Errorf("failed to insert call relation: %w", err)
	}

	return nil
}

// scanCallRelations is a helper function to scan call relation rows
func (si *SQLiteIndex) scanCallRelations(rows *sql.Rows) ([]models.CallRelation, error) {
	var relations []models.CallRelation
	for rows.Next() {
		var relation models.CallRelation
		err := rows.Scan(&relation.Caller, &relation.Callee, &relation.File, &relation.Line, &relation.CallerFile)
		if err != nil {
			return nil, fmt.Errorf("failed to scan call relation: %w", err)
		}
		relations = append(relations, relation)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return relations, nil
}

// QueryCallsFrom queries all functions called by the specified function
func (si *SQLiteIndex) QueryCallsFrom(caller string) ([]models.CallRelation, error) {
	query := `
	SELECT caller, callee, file, line, caller_file
	FROM call_relations
	WHERE caller = ?`

	rows, err := si.db.Query(query, caller)
	if err != nil {
		return nil, fmt.Errorf("failed to query calls from: %w", err)
	}
	defer rows.Close()

	return si.scanCallRelations(rows)
}

// QueryCallsTo queries all functions that call the specified function
func (si *SQLiteIndex) QueryCallsTo(callee string) ([]models.CallRelation, error) {
	query := `
	SELECT caller, callee, file, line, caller_file
	FROM call_relations
	WHERE callee = ?`

	rows, err := si.db.Query(query, callee)
	if err != nil {
		return nil, fmt.Errorf("failed to query calls to: %w", err)
	}
	defer rows.Close()

	return si.scanCallRelations(rows)
}

// RegisterChunk registers a new chunk in the database
func (si *SQLiteIndex) RegisterChunk(chunkID string, files []string, tokenCount int, createdAt time.Time) error {
	filesStr := strings.Join(files, ",")
	query := `
	INSERT INTO chunks (chunk_id, files, token_count, created_at)
	VALUES (?, ?, ?, ?)`

	_, err := si.db.Exec(query, chunkID, filesStr, tokenCount, createdAt)
	if err != nil {
		return fmt.Errorf("failed to register chunk: %w", err)
	}

	return nil
}

// DeleteChunk deletes a chunk and all related entries from the database
func (si *SQLiteIndex) DeleteChunk(chunkID string) error {
	// Start a transaction for atomic deletion
	tx, err := si.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		// Ignore rollback errors as they're expected after successful commit
		_ = tx.Rollback()
	}()

	// Delete related index entries first (due to foreign key constraint)
	_, err = tx.Exec("DELETE FROM index_entries WHERE chunk_id = ?", chunkID)
	if err != nil {
		return fmt.Errorf("failed to delete index entries: %w", err)
	}

	// Delete the chunk
	_, err = tx.Exec("DELETE FROM chunks WHERE chunk_id = ?", chunkID)
	if err != nil {
		return fmt.Errorf("failed to delete chunk: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close closes the database connection
func (si *SQLiteIndex) Close() error {
	if si.db != nil {
		return si.db.Close()
	}
	return nil
}
