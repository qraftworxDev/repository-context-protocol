package index

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"repository-context-protocol/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// Prepared statement cache for better performance (Phase 4A.2.2)
type PreparedStatementCache struct {
	statements map[string]*sql.Stmt
	mu         sync.RWMutex
}

// Common query constants for prepared statement caching
const (
	QueryByName    = "SELECT name, type, file_path, start_line, end_line, chunk_id, signature FROM index_entries WHERE name = ?"
	QueryByType    = "SELECT name, type, file_path, start_line, end_line, chunk_id, signature FROM index_entries WHERE type = ?"
	QueryCallsFrom = "SELECT caller, callee, file, line, caller_file FROM call_relations WHERE caller = ?"
	QueryCallsTo   = "SELECT caller, callee, file, line, caller_file FROM call_relations WHERE callee = ?"
)

// SQLiteIndex handles SQLite database operations for fast lookups
type SQLiteIndex struct {
	dbPath    string
	db        *sql.DB
	stmtCache *PreparedStatementCache
}

// NewSQLiteIndex creates a new SQLite index with the specified database path
func NewSQLiteIndex(dbPath string) *SQLiteIndex {
	return &SQLiteIndex{
		dbPath: dbPath,
		stmtCache: &PreparedStatementCache{
			statements: make(map[string]*sql.Stmt),
		},
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
	if err := si.createCoreTable(); err != nil {
		return err
	}
	if err := si.createRelationTables(); err != nil {
		return err
	}
	return si.createIndexes()
}

// createCoreTable creates the main index_entries table
func (si *SQLiteIndex) createCoreTable() error {
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
	return nil
}

// createRelationTables creates call relations and chunks tables
func (si *SQLiteIndex) createRelationTables() error {
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
	return nil
}

// createIndexes creates all database indexes for performance
func (si *SQLiteIndex) createIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_index_entries_name ON index_entries(name);",
		"CREATE INDEX IF NOT EXISTS idx_index_entries_type ON index_entries(type);",
		"CREATE INDEX IF NOT EXISTS idx_index_entries_file ON index_entries(file_path);",
		"CREATE INDEX IF NOT EXISTS idx_index_entries_chunk ON index_entries(chunk_id);",
		"CREATE INDEX IF NOT EXISTS idx_call_relations_caller ON call_relations(caller);",
		"CREATE INDEX IF NOT EXISTS idx_call_relations_callee ON call_relations(callee);",
		"CREATE INDEX IF NOT EXISTS idx_chunks_created_at ON chunks(created_at);",

		// Composite indexes for common query patterns (Phase 4A.1.1)
		"CREATE INDEX IF NOT EXISTS idx_type_name ON index_entries(type, name);",
		"CREATE INDEX IF NOT EXISTS idx_file_type ON index_entries(file_path, type);",
		"CREATE INDEX IF NOT EXISTS idx_name_file ON index_entries(name, file_path);",

		// Covering indexes to avoid chunk loading for index-only queries
		"CREATE INDEX IF NOT EXISTS idx_covering_basic ON index_entries(type, name, file_path, chunk_id);",
	}

	for _, indexSQL := range indexes {
		if _, err := si.db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	return nil
}

// getOrPrepareStatement returns cached statement or prepares new one with thread safety (Phase 4A.2.2)
func (si *SQLiteIndex) getOrPrepareStatement(queryKey, sqlQuery string) (*sql.Stmt, error) {
	// Try to get from cache with read lock
	si.stmtCache.mu.RLock()
	if stmt, exists := si.stmtCache.statements[queryKey]; exists {
		si.stmtCache.mu.RUnlock()
		return stmt, nil
	}
	si.stmtCache.mu.RUnlock()

	// Prepare statement with write lock
	si.stmtCache.mu.Lock()
	defer si.stmtCache.mu.Unlock()

	// Double-check pattern
	if stmt, exists := si.stmtCache.statements[queryKey]; exists {
		return stmt, nil
	}

	stmt, err := si.db.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}

	si.stmtCache.statements[queryKey] = stmt
	return stmt, nil
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
	stmt, err := si.getOrPrepareStatement("QueryByName", QueryByName)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	rows, err := stmt.Query(name)
	if err != nil {
		return nil, fmt.Errorf("failed to query index entries: %w", err)
	}
	defer rows.Close()

	return si.scanIndexEntries(rows)
}

// QueryIndexEntriesByType queries index entries by type
func (si *SQLiteIndex) QueryIndexEntriesByType(entryType string) ([]models.IndexEntry, error) {
	stmt, err := si.getOrPrepareStatement("QueryByType", QueryByType)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	rows, err := stmt.Query(entryType)
	if err != nil {
		return nil, fmt.Errorf("failed to query index entries by type: %w", err)
	}
	defer rows.Close()

	return si.scanIndexEntries(rows)
}

// queryIndexEntriesByField is a helper function to eliminate code duplication (Phase 4A.1.2)
func (si *SQLiteIndex) queryIndexEntriesByField(fieldName string, values []string, orderBy string) ([]models.IndexEntry, error) {
	if len(values) == 0 {
		return nil, nil
	}

	// Validate field names to prevent SQL injection
	baseQuery := "SELECT name, type, file_path, start_line, end_line, chunk_id, signature FROM index_entries"
	var query string
	switch fieldName {
	case "name":
		query = baseQuery + " WHERE name IN (%s) ORDER BY " + orderBy
	case EntityTypeType:
		query = baseQuery + " WHERE type IN (%s) ORDER BY " + orderBy
	default:
		return nil, fmt.Errorf("invalid field name: %s", fieldName)
	}

	// Build IN clause with placeholders
	placeholders := make([]string, len(values))
	args := make([]interface{}, len(values))
	for i, value := range values {
		placeholders[i] = "?"
		args[i] = value
	}

	finalQuery := fmt.Sprintf(query, strings.Join(placeholders, ","))
	rows, err := si.db.Query(finalQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query index entries by %s: %w", fieldName, err)
	}
	defer rows.Close()

	return si.scanIndexEntries(rows)
}

// QueryIndexEntriesByTypes queries index entries for multiple types in a single query (Phase 4A.1.2)
func (si *SQLiteIndex) QueryIndexEntriesByTypes(entryTypes []string) ([]models.IndexEntry, error) {
	return si.queryIndexEntriesByField(EntityTypeType, entryTypes, "type, name")
}

// QueryIndexEntriesByNames queries index entries for multiple names in a single query (Phase 4A.1.2)
func (si *SQLiteIndex) QueryIndexEntriesByNames(names []string) ([]models.IndexEntry, error) {
	return si.queryIndexEntriesByField("name", names, "name")
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
	stmt, err := si.getOrPrepareStatement("QueryCallsFrom", QueryCallsFrom)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	rows, err := stmt.Query(caller)
	if err != nil {
		return nil, fmt.Errorf("failed to query calls from: %w", err)
	}
	defer rows.Close()

	return si.scanCallRelations(rows)
}

// QueryCallsTo queries all functions that call the specified function
func (si *SQLiteIndex) QueryCallsTo(callee string) ([]models.CallRelation, error) {
	stmt, err := si.getOrPrepareStatement("QueryCallsTo", QueryCallsTo)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	rows, err := stmt.Query(callee)
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

	// First, get the files associated with this chunk so we can delete call relations
	var filesStr string
	err = tx.QueryRow("SELECT files FROM chunks WHERE chunk_id = ?", chunkID).Scan(&filesStr)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get chunk files: %w", err)
	}

	// Delete call relations for files in this chunk
	if filesStr != "" {
		files := strings.Split(filesStr, ",")
		for _, file := range files {
			_, err = tx.Exec("DELETE FROM call_relations WHERE file = ? OR caller_file = ?", file, file)
			if err != nil {
				return fmt.Errorf("failed to delete call relations for file %s: %w", file, err)
			}
		}
	}

	// Delete related index entries (due to foreign key constraint, this should cascade)
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
	// Close prepared statements first
	if si.stmtCache != nil {
		si.stmtCache.mu.Lock()
		for _, stmt := range si.stmtCache.statements {
			_ = stmt.Close()
		}
		si.stmtCache.statements = make(map[string]*sql.Stmt)
		si.stmtCache.mu.Unlock()
	}

	if si.db != nil {
		return si.db.Close()
	}
	return nil
}
