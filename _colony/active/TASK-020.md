# TASK-020: `get_history` MCP Tool — Expose Git History via MCP

**Priority:** P1
**Assigned:** alpha
**Milestone:** M2: Core Features
**Dependencies:** TASK-018
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
Feature 10 from GOALS.md — the user-facing MCP tool for git history. TASK-018 built the `internal/git` package that parses git history and stores it in the `git_context` table. This task creates the MCP tool that AI agents call to ask "Why does this function exist? Who last changed it? What are the hottest files?" This completes the git history feature end-to-end.

## Specification
Create the `internal/tools/history` package with the `get_history` MCP tool.

### MCP Tool: `get_history`

**Tool Definition:**
```json
{
  "name": "get_history",
  "description": "Get git history context for files and symbols. Shows who last changed code, why (commit messages as decision context), change frequency (hotspots), and which files frequently change together.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "file_path": {
        "type": "string",
        "description": "File path to get history for (relative to repo root)"
      },
      "mode": {
        "type": "string",
        "enum": ["file", "hotspots", "cochanged"],
        "description": "Query mode: 'file' for single file history, 'hotspots' for most-changed files, 'cochanged' for files that change together"
      },
      "limit": {
        "type": "integer",
        "description": "Maximum number of results (default: 10)"
      }
    },
    "required": ["mode"]
  }
}
```

### Handler Logic

**Mode: `file`** (requires `file_path`)
- Query `git_context` table for the given file_path
- Return: last author, last commit hash, last commit message, last modified date, change frequency, co-changed files
- If file not found in git_context, return helpful message: "No git history found. Run 'engram index' to analyze git history."

**Mode: `hotspots`**
- Query `git_context` table ordered by `change_frequency DESC`
- Return top N files (default 10) with their change frequency and last author
- This shows the "hot" areas of the codebase that change frequently

**Mode: `cochanged`** (requires `file_path`)
- Query `git_context` table for the given file_path's `co_changed_files`
- Parse JSON array and return the list with context
- Useful for: "If I'm changing file X, what else might I need to change?"

### Response Format

Return a structured text response (not raw JSON) that's easy for AI agents to consume:

**File mode:**
```
File: internal/auth/handler.go

Last modified: 2026-03-14 by alice
Commit: a1b2c3d "Fix session token expiry check"
Change frequency: 15 commits (hotspot rank: #3)

Often changed with:
  - internal/auth/middleware.go (12 co-changes)
  - internal/auth/session.go (8 co-changes)
  - tests/auth_test.go (15 co-changes)
```

**Hotspots mode:**
```
Codebase Hotspots (most frequently changed files):

 1. internal/api/router.go          — 42 changes, last by bob
 2. internal/auth/handler.go        — 15 changes, last by alice
 3. internal/db/migrations.go       — 12 changes, last by charlie
 ...
```

**Cochanged mode:**
```
Files that frequently change with internal/auth/handler.go:

 1. tests/auth_test.go              — 15 co-changes
 2. internal/auth/middleware.go      — 12 co-changes
 3. internal/auth/session.go        — 8 co-changes
```

### Package Structure

```
internal/tools/history/
├── tools.go       — HistoryTool struct, Definition(), Handle()
└── register.go    — RegisterTools function
```

### Registration

```go
// register.go
func RegisterTools(server *engmcp.Server, store *storage.Store) {
    tool := NewHistoryTool(store)
    server.RegisterTool(tool.Definition(), tool.Handle)
}
```

## Acceptance Criteria
- [ ] `get_history` with mode `file` returns correct history for a given file path
- [ ] `get_history` with mode `hotspots` returns files sorted by change frequency
- [ ] `get_history` with mode `cochanged` returns co-changed files for a given file
- [ ] `get_history` with missing `file_path` for modes that require it returns clear error
- [ ] `get_history` returns helpful message when git_context is empty (no analysis run yet)
- [ ] `limit` parameter works correctly for all modes
- [ ] Default limit is 10
- [ ] Tool follows the same registration pattern as other MCP tools
- [ ] All tests pass

## Implementation Steps
1. Create `internal/tools/history/tools.go`:
   - HistoryTool struct with store field
   - NewHistoryTool constructor
   - Definition() returning mcpgo.Tool with schema
   - Handle() dispatching on mode parameter
   - handleFile(), handleHotspots(), handleCochanged() private methods
   - Format response text for each mode
2. Create `internal/tools/history/register.go`:
   - RegisterTools function following existing pattern
3. Create `internal/tools/history/tools_test.go`:
   - Test: Definition returns correct tool name and schema
   - Test: Handle with mode "file" and valid file_path queries git_context
   - Test: Handle with mode "hotspots" returns ordered results
   - Test: Handle with mode "cochanged" returns co-changed files
   - Test: Handle with mode "file" and missing file_path returns error
   - Test: Handle with empty git_context table returns helpful message
   - Test: limit parameter caps results
4. Run all tests

## Testing Requirements
- Unit test: Tool definition has name "get_history" with correct input schema
- Unit test: File mode returns formatted history for a known file (seed git_context with test data)
- Unit test: Hotspots mode returns files in descending frequency order (seed test data)
- Unit test: Cochanged mode parses JSON co_changed_files correctly
- Unit test: Missing file_path on file/cochanged mode returns error
- Unit test: Empty git_context returns "run engram index" message
- Unit test: Limit parameter is respected

## Files to Create/Modify
- `internal/tools/history/tools.go` — get_history MCP tool implementation
- `internal/tools/history/register.go` — tool registration
- `internal/tools/history/tools_test.go` — tool tests

## Notes
- Follow the exact pattern from `internal/tools/search/` and `internal/tools/memory/`. Study those packages for the constructor, Definition, Handle, and RegisterTools patterns.
- The git_context table queries are straightforward SELECT statements. Use `store.DB().Query()` or `store.DB().QueryRow()` directly, similar to how other tools query their tables.
- The `co_changed_files` column stores a JSON array of strings. Use `json.Unmarshal` to parse it.
- Do NOT modify `cmd/engram/serve.go` in this task. The tool will be wired into the serve command in a later wiring task when all M2 tools are ready.
- Do NOT import or depend on `internal/git/` directly. This tool reads from the `git_context` table only — the git package populates it, this tool reads it. They are decoupled by the database.
- Response format should be plain text with clear structure, not JSON. AI agents parse structured text better than nested JSON for contextual information.
