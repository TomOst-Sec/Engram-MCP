package storage

// SchemaVersion is the current schema version.
const SchemaVersion = 3

// schemaV1 contains all DDL statements for schema version 1.
var schemaV1 = []string{
	// Core indexing
	`CREATE TABLE IF NOT EXISTS code_index (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_path TEXT NOT NULL,
		file_hash TEXT NOT NULL,
		language TEXT NOT NULL,
		symbol_name TEXT NOT NULL,
		symbol_type TEXT NOT NULL,
		signature TEXT,
		docstring TEXT,
		start_line INTEGER NOT NULL,
		end_line INTEGER NOT NULL,
		body_hash TEXT,
		embedding BLOB,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Memories from coding sessions
	`CREATE TABLE IF NOT EXISTS memories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		summary TEXT,
		type TEXT NOT NULL,
		tags TEXT,
		related_files TEXT,
		embedding BLOB,
		session_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME
	)`,

	// Inferred conventions
	`CREATE TABLE IF NOT EXISTS conventions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pattern TEXT NOT NULL,
		description TEXT NOT NULL,
		category TEXT NOT NULL,
		confidence REAL NOT NULL,
		examples TEXT,
		language TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Architecture module graph
	`CREATE TABLE IF NOT EXISTS architecture (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		module_name TEXT NOT NULL,
		module_path TEXT NOT NULL,
		description TEXT,
		dependencies TEXT,
		exports TEXT,
		complexity_score REAL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Git history context
	`CREATE TABLE IF NOT EXISTS git_context (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_path TEXT NOT NULL,
		symbol_name TEXT,
		last_author TEXT,
		last_commit_hash TEXT,
		last_commit_message TEXT,
		last_modified DATETIME,
		change_frequency INTEGER DEFAULT 0,
		co_changed_files TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	// Schema version tracking
	`CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
}

// ftsV1 contains FTS5 virtual table definitions.
var ftsV1 = []string{
	`CREATE VIRTUAL TABLE IF NOT EXISTS code_index_fts USING fts5(
		symbol_name, signature, docstring,
		content=code_index, content_rowid=id
	)`,

	`CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
		content, summary, tags,
		content=memories, content_rowid=id
	)`,
}

// triggersV1 keeps FTS5 indexes in sync with content tables.
var triggersV1 = []string{
	// code_index FTS sync triggers
	`CREATE TRIGGER IF NOT EXISTS code_index_ai AFTER INSERT ON code_index BEGIN
		INSERT INTO code_index_fts(rowid, symbol_name, signature, docstring)
		VALUES (new.id, new.symbol_name, new.signature, new.docstring);
	END`,
	`CREATE TRIGGER IF NOT EXISTS code_index_ad AFTER DELETE ON code_index BEGIN
		INSERT INTO code_index_fts(code_index_fts, rowid, symbol_name, signature, docstring)
		VALUES ('delete', old.id, old.symbol_name, old.signature, old.docstring);
	END`,
	`CREATE TRIGGER IF NOT EXISTS code_index_au AFTER UPDATE ON code_index BEGIN
		INSERT INTO code_index_fts(code_index_fts, rowid, symbol_name, signature, docstring)
		VALUES ('delete', old.id, old.symbol_name, old.signature, old.docstring);
		INSERT INTO code_index_fts(rowid, symbol_name, signature, docstring)
		VALUES (new.id, new.symbol_name, new.signature, new.docstring);
	END`,

	// memories FTS sync triggers
	`CREATE TRIGGER IF NOT EXISTS memories_ai AFTER INSERT ON memories BEGIN
		INSERT INTO memories_fts(rowid, content, summary, tags)
		VALUES (new.id, new.content, new.summary, new.tags);
	END`,
	`CREATE TRIGGER IF NOT EXISTS memories_ad AFTER DELETE ON memories BEGIN
		INSERT INTO memories_fts(memories_fts, rowid, content, summary, tags)
		VALUES ('delete', old.id, old.content, old.summary, old.tags);
	END`,
	`CREATE TRIGGER IF NOT EXISTS memories_au AFTER UPDATE ON memories BEGIN
		INSERT INTO memories_fts(memories_fts, rowid, content, summary, tags)
		VALUES ('delete', old.id, old.content, old.summary, old.tags);
		INSERT INTO memories_fts(rowid, content, summary, tags)
		VALUES (new.id, new.content, new.summary, new.tags);
	END`,
}

// indexesV1 contains index definitions for efficient queries.
var indexesV1 = []string{
	`CREATE INDEX IF NOT EXISTS idx_code_index_file ON code_index(file_path)`,
	`CREATE INDEX IF NOT EXISTS idx_code_index_file_hash ON code_index(file_hash)`,
	`CREATE INDEX IF NOT EXISTS idx_code_index_language ON code_index(language)`,
	`CREATE INDEX IF NOT EXISTS idx_code_index_symbol_type ON code_index(symbol_type)`,
	`CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type)`,
	`CREATE INDEX IF NOT EXISTS idx_memories_session ON memories(session_id)`,
	`CREATE INDEX IF NOT EXISTS idx_git_context_file ON git_context(file_path)`,
}

// schemaV2 adds the code_references table for tracking function calls,
// pointer assignments, callback arguments, and dispatch table entries.
var schemaV2 = []string{
	`CREATE TABLE IF NOT EXISTS code_references (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_path TEXT NOT NULL,
		to_name TEXT NOT NULL,
		kind TEXT NOT NULL,
		from_func TEXT NOT NULL,
		line INTEGER NOT NULL,
		context TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS idx_refs_to_name ON code_references(to_name)`,
	`CREATE INDEX IF NOT EXISTS idx_refs_file ON code_references(file_path)`,
	`CREATE INDEX IF NOT EXISTS idx_refs_kind ON code_references(kind)`,
	`CREATE VIRTUAL TABLE IF NOT EXISTS code_references_fts USING fts5(
		to_name, from_func, context,
		content=code_references, content_rowid=id
	)`,
	`CREATE TRIGGER IF NOT EXISTS code_refs_ai AFTER INSERT ON code_references BEGIN
		INSERT INTO code_references_fts(rowid, to_name, from_func, context)
		VALUES (new.id, new.to_name, new.from_func, new.context);
	END`,
	`CREATE TRIGGER IF NOT EXISTS code_refs_ad AFTER DELETE ON code_references BEGIN
		INSERT INTO code_references_fts(code_references_fts, rowid, to_name, from_func, context)
		VALUES ('delete', old.id, old.to_name, old.from_func, old.context);
	END`,
}

// schemaV3 adds param_name and table_name columns to code_references,
// and a table_membership table for tracking which functions live in which dispatch tables.
var schemaV3 = []string{
	`ALTER TABLE code_references ADD COLUMN param_name TEXT`,
	`ALTER TABLE code_references ADD COLUMN table_name TEXT`,
	`CREATE TABLE IF NOT EXISTS table_membership (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		table_var TEXT NOT NULL,
		func_name TEXT NOT NULL,
		file_path TEXT NOT NULL,
		slot_index INTEGER NOT NULL,
		line INTEGER NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_tm_func ON table_membership(func_name)`,
	`CREATE INDEX IF NOT EXISTS idx_tm_table ON table_membership(table_var)`,
}
