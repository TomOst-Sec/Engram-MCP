# TASK-038: `engram export` + `engram import` CLI Commands

**Priority:** P2
**Assigned:** bravo
**Milestone:** M2: Core Features
**Dependencies:** TASK-003
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
Part of Feature 13 (Full CLI) and CEO directive request. These commands dump the SQLite database to JSON and load from JSON. Useful for sharing Engram data between machines, backing up before upgrades, and debugging. `engram export` writes to stdout or a file. `engram import` reads from stdin or a file.

## Specification

### `engram export` Command

```go
var exportCmd = &cobra.Command{
    Use:   "export [file]",
    Short: "Export database to JSON",
    Long:  "Export all Engram data (memories, code index, conventions, git context) to JSON format.",
    RunE:  runExport,
}
```

**Flags:**
- `--tables` / `-t` — comma-separated tables to export (default: all). Options: memories, code_index, conventions, git_context, architecture
- `--pretty` — pretty-print JSON (default: compact)

**Output format:**
```json
{
  "version": 1,
  "exported_at": "2026-03-15T17:00:00Z",
  "repo_root": "/path/to/repo",
  "tables": {
    "memories": [...],
    "code_index": [...],
    "conventions": [...],
    "git_context": [...],
    "architecture": [...]
  }
}
```

**Behavior:**
1. Open database (reuse openDatabase helper from db.go)
2. Query each table: `SELECT * FROM <table>`
3. Marshal rows to JSON
4. Write to file (if arg given) or stdout

### `engram import` Command

```go
var importCmd = &cobra.Command{
    Use:   "import [file]",
    Short: "Import database from JSON",
    Long:  "Import Engram data from a JSON export file. Merges with existing data.",
    RunE:  runImport,
}
```

**Flags:**
- `--replace` — clear existing data before importing (default: merge/upsert)

**Behavior:**
1. Read JSON from file (if arg given) or stdin
2. Validate version field
3. For each table: upsert rows (INSERT OR REPLACE)
4. Print summary: "Imported: 42 memories, 2847 symbols, 8 conventions"

## Acceptance Criteria
- [ ] `engram export` outputs valid JSON with all tables
- [ ] `engram export > file.json` writes to file
- [ ] `engram export --tables memories` exports only memories
- [ ] `engram export --pretty` produces indented JSON
- [ ] `engram import file.json` imports data into the database
- [ ] `engram import --replace` clears existing data first
- [ ] Import validates version field
- [ ] Round-trip: export then import produces identical data
- [ ] Both commands registered on root cobra command
- [ ] All tests pass

## Implementation Steps
1. Create `cmd/engram/export.go` — export subcommand
2. Create `cmd/engram/import_cmd.go` — import subcommand (note: `import.go` conflicts with Go keyword, use `import_cmd.go`)
3. Register both in `cmd/engram/main.go`
4. Create `cmd/engram/export_test.go`:
   - Test: export command registered
   - Test: export produces valid JSON structure
5. Create `cmd/engram/import_test.go`:
   - Test: import command registered
   - Test: import with --replace clears data
6. Run all tests

## Files to Create/Modify
- `cmd/engram/export.go` — export command
- `cmd/engram/import_cmd.go` — import command
- `cmd/engram/export_test.go` — tests
- `cmd/engram/import_test.go` — tests
- `cmd/engram/main.go` — register both commands

## Notes
- File name `import_cmd.go` avoids collision with `import` Go keyword.
- For the code_index table, exclude the `embedding` BLOB column from JSON export (it's binary data). Include a `has_embedding: true/false` field instead.
- Export should handle large databases without loading everything into memory. Use `sql.Rows` iteration and `json.Encoder` streaming.
- Import should use transactions — wrap all inserts in a single transaction for performance.

---
## Completion Notes
- **Completed by:** bravo-1
- **Date:** 2026-03-15 18:02:56
- **Branch:** task/038
