package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenCreatesNewDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Verify all 6 tables exist
	tables := []string{
		"code_index", "memories", "conventions",
		"architecture", "git_context", "schema_version",
	}
	for _, table := range tables {
		var name string
		err := store.DB().QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&name)
		assert.NoError(t, err, "table %s should exist", table)
		assert.Equal(t, table, name)
	}

	// Verify 2 FTS tables exist
	ftsTables := []string{"code_index_fts", "memories_fts"}
	for _, table := range ftsTables {
		var name string
		err := store.DB().QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&name)
		assert.NoError(t, err, "FTS table %s should exist", table)
		assert.Equal(t, table, name)
	}

	// Verify indexes exist
	indexes := []string{
		"idx_code_index_file", "idx_code_index_file_hash",
		"idx_code_index_language", "idx_code_index_symbol_type",
		"idx_memories_type", "idx_memories_session",
		"idx_git_context_file",
	}
	for _, idx := range indexes {
		var name string
		err := store.DB().QueryRow(
			`SELECT name FROM sqlite_master WHERE type='index' AND name=?`, idx,
		).Scan(&name)
		assert.NoError(t, err, "index %s should exist", idx)
	}
}

func TestOpenIsIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Open and close twice
	store1, err := Open(dbPath)
	require.NoError(t, err)
	require.NoError(t, store1.Close())

	store2, err := Open(dbPath)
	require.NoError(t, err)
	defer store2.Close()

	// Verify schema version is still 1 (not duplicated)
	var count int
	err = store2.DB().QueryRow(`SELECT COUNT(*) FROM schema_version`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestWALModeEnabled(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	var journalMode string
	err = store.DB().QueryRow(`PRAGMA journal_mode`).Scan(&journalMode)
	require.NoError(t, err)
	assert.Equal(t, "wal", journalMode)
}

func TestSchemaVersionIsOne(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	var version int
	err = store.DB().QueryRow(`SELECT MAX(version) FROM schema_version`).Scan(&version)
	require.NoError(t, err)
	assert.Equal(t, 1, version)
}

func TestFTS5CodeIndex(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Insert a row into code_index
	_, err = store.DB().Exec(
		`INSERT INTO code_index (file_path, file_hash, language, symbol_name, symbol_type, signature, docstring, start_line, end_line)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"internal/auth/handler.go", "abc123", "go",
		"HandleLogin", "function",
		"func HandleLogin(w http.ResponseWriter, r *http.Request) error",
		"HandleLogin processes user authentication requests",
		10, 45,
	)
	require.NoError(t, err)

	// Search via FTS5
	var symbolName string
	err = store.DB().QueryRow(
		`SELECT symbol_name FROM code_index_fts WHERE code_index_fts MATCH ?`, "HandleLogin",
	).Scan(&symbolName)
	require.NoError(t, err)
	assert.Equal(t, "HandleLogin", symbolName)

	// Search via docstring
	err = store.DB().QueryRow(
		`SELECT symbol_name FROM code_index_fts WHERE code_index_fts MATCH ?`, "authentication",
	).Scan(&symbolName)
	require.NoError(t, err)
	assert.Equal(t, "HandleLogin", symbolName)
}

func TestFTS5Memories(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Insert a memory
	_, err = store.DB().Exec(
		`INSERT INTO memories (content, type, tags)
		 VALUES (?, ?, ?)`,
		"Decided to use JWT tokens for API authentication instead of sessions",
		"decision",
		`["auth","api"]`,
	)
	require.NoError(t, err)

	// Search via FTS5
	var content string
	err = store.DB().QueryRow(
		`SELECT content FROM memories_fts WHERE memories_fts MATCH ?`, "JWT authentication",
	).Scan(&content)
	require.NoError(t, err)
	assert.Contains(t, content, "JWT tokens")
}

func TestParentDirectoryCreation(t *testing.T) {
	baseDir := t.TempDir()
	dbPath := filepath.Join(baseDir, "deeply", "nested", "dir", "test.db")
	store, err := Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Verify file exists
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)
}

func TestCloseIsClean(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(dbPath)
	require.NoError(t, err)

	err = store.Close()
	assert.NoError(t, err)

	// Verify we can't query after close
	var n int
	err = store.DB().QueryRow(`SELECT 1`).Scan(&n)
	assert.Error(t, err) // should fail, connection closed
}

func TestDBReturnsUnderlyingConnection(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	db := store.DB()
	assert.NotNil(t, db)
	assert.IsType(t, &sql.DB{}, db)
}

func TestCloseNilDB(t *testing.T) {
	store := &Store{db: nil}
	err := store.Close()
	assert.NoError(t, err)
}
