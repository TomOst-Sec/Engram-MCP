# TASK-003: SQLite Storage Layer — Schema v1, WAL Mode, Migrations, Connection Management

**Priority:** P0
**Assigned:** alpha
**Milestone:** M1: MVP
**Dependencies:** TASK-001
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
Every feature in Engram stores data in SQLite — the AST index, embeddings, memories, conventions, git context, and architecture data. This task creates the foundational storage layer: database creation, schema definition, migration system, connection management with WAL mode, and basic CRUD helpers. This is the most critical dependency for Milestone 1 — almost every subsequent task builds on this.

## Specification
Create the `internal/storage` package implementing a SQLite-backed storage layer.

### Schema v1
Define these tables (matching the tech stack spec):

```sql
-- Core indexing
CREATE TABLE IF NOT EXISTS code_index (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL,
    file_hash TEXT NOT NULL,           -- SHA256 of file contents (for incremental re-indexing)
    language TEXT NOT NULL,
    symbol_name TEXT NOT NULL,
    symbol_type TEXT NOT NULL,          -- function, type, class, interface, import, export, test
    signature TEXT,                     -- function signature string
    docstring TEXT,
    start_line INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    body_hash TEXT,                     -- hash of the symbol body for change detection
    embedding BLOB,                    -- 384-dim float32 vector (1536 bytes) or NULL
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Memories from coding sessions
CREATE TABLE IF NOT EXISTS memories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
    summary TEXT,                       -- compressed/summarized version
    type TEXT NOT NULL,                 -- decision, bugfix, refactor, learning, convention
    tags TEXT,                          -- JSON array of tags
    related_files TEXT,                 -- JSON array of file paths
    embedding BLOB,                    -- 384-dim float32 vector or NULL
    session_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME                -- soft delete
);

-- Inferred conventions
CREATE TABLE IF NOT EXISTS conventions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pattern TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL,            -- naming, error_handling, testing, imports, structure, comments
    confidence REAL NOT NULL,          -- 0.0 to 1.0
    examples TEXT,                     -- JSON array of example snippets
    language TEXT,                     -- NULL = applies to all languages
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Architecture module graph
CREATE TABLE IF NOT EXISTS architecture (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    module_name TEXT NOT NULL,
    module_path TEXT NOT NULL,
    description TEXT,
    dependencies TEXT,                 -- JSON array of module names
    exports TEXT,                      -- JSON array of exported symbols
    complexity_score REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Git history context
CREATE TABLE IF NOT EXISTS git_context (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL,
    symbol_name TEXT,
    last_author TEXT,
    last_commit_hash TEXT,
    last_commit_message TEXT,
    last_modified DATETIME,
    change_frequency INTEGER DEFAULT 0, -- number of commits touching this
    co_changed_files TEXT,              -- JSON array of files frequently changed together
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Full-text search indexes
CREATE VIRTUAL TABLE IF NOT EXISTS code_index_fts USING fts5(
    symbol_name, signature, docstring,
    content=code_index, content_rowid=id
);

CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
    content, summary, tags,
    content=memories, content_rowid=id
);
```

Also create indexes:
```sql
CREATE INDEX IF NOT EXISTS idx_code_index_file ON code_index(file_path);
CREATE INDEX IF NOT EXISTS idx_code_index_file_hash ON code_index(file_hash);
CREATE INDEX IF NOT EXISTS idx_code_index_language ON code_index(language);
CREATE INDEX IF NOT EXISTS idx_code_index_symbol_type ON code_index(symbol_type);
CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type);
CREATE INDEX IF NOT EXISTS idx_memories_session ON memories(session_id);
CREATE INDEX IF NOT EXISTS idx_git_context_file ON git_context(file_path);
```

### Connection Management
```go
type Store struct {
    db *sql.DB
}

func Open(dbPath string) (*Store, error)   // open DB, apply migrations, enable WAL
func (s *Store) Close() error
func (s *Store) DB() *sql.DB               // raw access for advanced queries
```

### Key Requirements
- Enable WAL mode on open: `PRAGMA journal_mode=WAL`
- Enable foreign keys: `PRAGMA foreign_keys=ON`
- Set busy timeout: `PRAGMA busy_timeout=5000`
- Auto-vacuum: `PRAGMA auto_vacuum=INCREMENTAL`
- Run `PRAGMA incremental_vacuum` on close
- Create parent directories for dbPath if they don't exist
- Schema migrations: check schema_version table, apply missing migrations
- Thread-safe: multiple goroutines can call Store methods concurrently

## Acceptance Criteria
- [ ] `storage.Open(path)` creates a new database with all tables if it doesn't exist
- [ ] `storage.Open(path)` on an existing v1 database succeeds without recreating tables
- [ ] WAL mode is enabled (query `PRAGMA journal_mode` returns "wal")
- [ ] FTS5 tables are created and functional (insert to code_index, search via code_index_fts)
- [ ] Schema version is recorded as 1 in schema_version table
- [ ] Database file is created in the specified path, parent directories auto-created
- [ ] `Store.Close()` runs incremental vacuum and closes cleanly
- [ ] All tests pass with `go test ./internal/storage/`

## Implementation Steps
1. `go get github.com/mattn/go-sqlite3` — add CGo SQLite dependency
2. Create `internal/storage/schema.go` — SQL constants for table creation and indexes
3. Create `internal/storage/migrations.go` — migration system (check version, apply in order)
4. Create `internal/storage/store.go` — Store struct, Open(), Close(), DB(), pragma setup
5. Create `internal/storage/store_test.go`:
   - Test Open creates new DB with all tables
   - Test Open on existing DB is idempotent
   - Test WAL mode is enabled
   - Test FTS5 works (insert + search)
   - Test schema version is set
   - Test parent dir creation
   - Test Close runs cleanly
6. Run tests, ensure all pass

## Testing Requirements
- Unit test: Open() creates all 6 tables + 2 FTS tables + indexes
- Unit test: Open() is idempotent (call twice on same path, no error)
- Unit test: WAL pragma returns "wal"
- Unit test: Insert a row into code_index, search via code_index_fts, get result back
- Unit test: Insert a row into memories, search via memories_fts, get result back
- Unit test: schema_version contains version 1
- Unit test: Open with nested non-existent directories creates them

## Files to Create/Modify
- `internal/storage/store.go` — Store struct, Open, Close, DB methods, pragma setup
- `internal/storage/schema.go` — SQL DDL constants for all tables, indexes, FTS
- `internal/storage/migrations.go` — migration runner, version checking
- `internal/storage/store_test.go` — comprehensive tests

## Notes
- Use `github.com/mattn/go-sqlite3` which requires CGo. Build tags may be needed for cross-compilation, but don't worry about that now.
- Use `os.MkdirAll` for parent directory creation.
- For FTS5 content sync, you'll need triggers to keep FTS in sync with the content tables. Add INSERT/DELETE/UPDATE triggers that mirror changes to the FTS tables.
- Use `t.TempDir()` in tests for database paths — auto-cleaned up.
- Do NOT add any CRUD helper methods beyond Open/Close/DB in this task. Those will come with the tools that need them (remember, recall, search, etc.).
