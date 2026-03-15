# TASK-015: CLI Commands — search, recall, status

**Priority:** P1
**Assigned:** alpha
**Milestone:** M1: MVP
**Dependencies:** TASK-003, TASK-008
**Status:** queued
**Created:** 2026-03-15
**Author:** atlas

## Context
The MVP needs CLI commands for users to interact with Engram directly from the terminal, not just through MCP. `engram search` lets developers search code from the command line, `engram recall` retrieves past memories, and `engram status` shows the current index state. These commands are essential for debugging, verification, and standalone use. They all read from the same SQLite database that the MCP server uses.

## Specification
Create three new CLI subcommands in `cmd/engram/`.

### `engram search <query>` — CLI Code Search
```go
var searchCmd = &cobra.Command{
    Use:   "search <query>",
    Short: "Search the codebase",
    Long:  "Search indexed code using natural language or keywords. Uses FTS5 full-text search.",
    Args:  cobra.MinimumNArgs(1),
    RunE:  runSearch,
}
```

**Flags:**
- `--language` / `-l` — filter by language
- `--type` / `-t` — filter by symbol type (function, class, type, etc.)
- `--limit` / `-n` — max results (default: 10)

**Output format (to stdout, styled with lipgloss):**
```
  internal/auth/handler.go:42  HandleLogin  function
  func HandleLogin(w http.ResponseWriter, r *http.Request) error

  internal/auth/middleware.go:15  AuthMiddleware  function
  func AuthMiddleware(next http.Handler) http.Handler

Found 2 results for "authentication handler" (0.8ms)
```

**Implementation:**
1. Open database (same path logic as serve command)
2. Run FTS5 query against code_index_fts
3. Apply filters
4. Format and print results

### `engram recall <query>` — CLI Memory Search
```go
var recallCmd = &cobra.Command{
    Use:   "recall <query>",
    Short: "Search past memories",
    Long:  "Search memories from past coding sessions using natural language.",
    Args:  cobra.MinimumNArgs(1),
    RunE:  runRecall,
}
```

**Flags:**
- `--type` / `-t` — filter by memory type (decision, bugfix, refactor, learning, convention)
- `--limit` / `-n` — max results (default: 10)
- `--since` — only memories after this date (YYYY-MM-DD)

**Output format:**
```
  #42  [decision]  2026-03-14 14:30
  Decided to use SQLite with FTS5 instead of PostgreSQL for the storage layer.
  Tags: database, architecture
  Files: internal/storage/store.go

  #38  [learning]  2026-03-13 09:15
  The FTS5 rank function returns negative scores; more negative = better match.
  Tags: search, sqlite

Found 2 memories for "database decisions" (1.2ms)
```

**Implementation:**
1. Open database
2. Run FTS5 query against memories_fts
3. Apply filters (type, since, limit)
4. Format and print results

### `engram status` — Index Status
```go
var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show Engram index status",
    Long:  "Display statistics about the current code index, memories, and database.",
    RunE:  runStatus,
}
```

**Output format:**
```
Engram v0.1.0-dev
Repository: /path/to/repo
Database:   ~/.engram/a1b2c3d4e5f6/engram.db (4.2 MB)

Index:
  Files indexed:    342
  Symbols:          2,847
  Languages:        Go (120), Python (85), TypeScript (137)
  Embeddings:       2,847 / 2,847 (100%)
  Last indexed:     2026-03-15 14:30:00

Memories:
  Total:            42
  By type:          decision (15), bugfix (10), learning (12), convention (5)
  Oldest:           2026-03-01
  Newest:           2026-03-15

Conventions:        8 patterns detected
Architecture:       12 modules mapped
```

**Implementation:**
1. Open database
2. Run COUNT queries against each table
3. Get database file size
4. Format with lipgloss (or simple fmt for MVP)

### Shared Database Helper
To avoid duplicating repo root detection and database opening in every command, create a shared helper:

```go
// cmd/engram/db.go
func openDatabase() (*storage.Store, string, error) {
    repoRoot, err := detectRepoRoot()
    // ... load config, construct dbPath, open store
    return store, repoRoot, err
}
```

## Acceptance Criteria
- [ ] `engram search "query"` returns matching code symbols from the index
- [ ] `engram search --language go "query"` filters to Go only
- [ ] `engram search --type function "query"` filters to functions only
- [ ] `engram search --limit 3 "query"` returns at most 3 results
- [ ] `engram recall "query"` returns matching memories
- [ ] `engram recall --type decision "query"` filters by memory type
- [ ] `engram recall --since 2026-03-01 "query"` filters by date
- [ ] `engram status` shows file count, symbol count, memory count, database size
- [ ] All commands open and close the database correctly
- [ ] All commands fail gracefully if database doesn't exist (clear error message)
- [ ] All commands are registered on the root cobra command
- [ ] All tests pass

## Implementation Steps
1. Create `cmd/engram/db.go` — shared openDatabase() helper
2. Create `cmd/engram/search.go` — search subcommand with flags
3. Create `cmd/engram/recall.go` — recall subcommand with flags
4. Create `cmd/engram/status.go` — status subcommand
5. Update `cmd/engram/main.go` — add all three commands to root
6. Create `cmd/engram/search_test.go`:
   - Test: search command is registered
   - Test: --help output contains relevant info
   - Test: missing query argument returns error
7. Create `cmd/engram/recall_test.go`:
   - Test: recall command is registered
   - Test: --help output correct
   - Test: missing query returns error
8. Create `cmd/engram/status_test.go`:
   - Test: status command is registered
   - Test: --help output correct
9. Run all tests

## Testing Requirements
- Unit test: search command registered on root cmd
- Unit test: recall command registered on root cmd
- Unit test: status command registered on root cmd
- Unit test: search with no args returns error
- Unit test: recall with no args returns error
- Unit test: openDatabase helper handles missing .git gracefully

## Files to Create/Modify
- `cmd/engram/db.go` — shared database helper (openDatabase, detectRepoRoot if not already shared)
- `cmd/engram/search.go` — search subcommand
- `cmd/engram/recall.go` — recall subcommand
- `cmd/engram/status.go` — status subcommand
- `cmd/engram/search_test.go` — search tests
- `cmd/engram/recall_test.go` — recall tests
- `cmd/engram/status_test.go` — status tests
- `cmd/engram/main.go` — add three commands (3 lines)

## Notes
- For MVP, use `fmt.Printf` with manual formatting instead of lipgloss. Lipgloss styling is a nice-to-have that can be added later without changing the command logic.
- The search command queries code_index_fts directly (not the MCP tool). This is intentional — CLI commands talk to SQLite directly, they don't go through the MCP server.
- Handle the case where the database doesn't exist: print "No Engram database found. Run 'engram index' first." and exit with code 1.
- For database file size, use `os.Stat(dbPath).Size()` and format with human-readable units (KB, MB).
- Join query args with space for multi-word queries: `query := strings.Join(args, " ")`
- The `detectRepoRoot()` function already exists in serve.go. Extract it to db.go so all commands can use it. If extraction is too risky (modifying serve.go), duplicate it — it's a simple function.
