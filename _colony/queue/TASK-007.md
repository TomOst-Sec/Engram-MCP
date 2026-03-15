# TASK-007: Remember and Recall MCP Tools — Memory Storage and Retrieval

**Priority:** P1
**Assigned:** bravo
**Milestone:** M1: MVP
**Dependencies:** TASK-003, TASK-004
**Status:** queued
**Created:** 2026-03-15
**Author:** atlas

## Context
The `remember` and `recall` tools are what make Engram "remember yesterday." They allow AI coding agents to store important context from coding sessions (decisions, bug fixes, refactors, learnings) and retrieve them later using natural language queries. This is the core persistence feature that differentiates Engram from a stateless MCP tool — it gives AI agents long-term memory across sessions.

## Specification
Create the `internal/tools/memory` package implementing the `remember` and `recall` MCP tools.

### remember Tool
**MCP Tool Name:** `remember`
**Description:** "Store a memory from the current coding session for future reference"

**Input Schema:**
```json
{
    "type": "object",
    "properties": {
        "content": {
            "type": "string",
            "description": "What to remember (decision, learning, bug fix, etc.)"
        },
        "type": {
            "type": "string",
            "enum": ["decision", "bugfix", "refactor", "learning", "convention"],
            "description": "Category of this memory"
        },
        "tags": {
            "type": "array",
            "items": {"type": "string"},
            "description": "Optional tags for categorization"
        },
        "related_files": {
            "type": "array",
            "items": {"type": "string"},
            "description": "Optional file paths related to this memory"
        }
    },
    "required": ["content", "type"]
}
```

**Behavior:**
1. Validate input (content not empty, type is valid enum value)
2. Generate a session ID if not already set (UUID, persisted for server lifetime)
3. Store in the `memories` table with all fields
4. Update the FTS5 index (via trigger — set up in TASK-003)
5. Return confirmation with the memory ID and timestamp

**Response:**
```json
{
    "id": 42,
    "status": "stored",
    "created_at": "2026-03-15T10:30:00Z",
    "content_preview": "First 100 chars of content..."
}
```

### recall Tool
**MCP Tool Name:** `recall`
**Description:** "Search memories from past coding sessions"

**Input Schema:**
```json
{
    "type": "object",
    "properties": {
        "query": {
            "type": "string",
            "description": "Natural language search query"
        },
        "type": {
            "type": "string",
            "enum": ["decision", "bugfix", "refactor", "learning", "convention"],
            "description": "Filter by memory type"
        },
        "tags": {
            "type": "array",
            "items": {"type": "string"},
            "description": "Filter by tags (AND logic)"
        },
        "limit": {
            "type": "integer",
            "description": "Maximum results to return (default: 10)"
        },
        "since": {
            "type": "string",
            "description": "Only memories after this ISO date"
        }
    },
    "required": ["query"]
}
```

**Behavior:**
1. Search memories using FTS5 full-text search on content and summary fields
2. Apply optional filters (type, tags, date range)
3. Rank by FTS5 relevance score
4. Exclude soft-deleted memories (deleted_at IS NULL)
5. Return top results with metadata

**Response:**
```json
{
    "memories": [
        {
            "id": 42,
            "content": "Full content...",
            "type": "decision",
            "tags": ["auth", "security"],
            "related_files": ["internal/auth/handler.go"],
            "created_at": "2026-03-15T10:30:00Z",
            "relevance_score": 0.95
        }
    ],
    "total_matches": 5,
    "query": "authentication decisions"
}
```

### Package Structure
```go
// internal/tools/memory/remember.go
type RememberTool struct {
    store     *storage.Store
    sessionID string
}

func NewRememberTool(store *storage.Store) *RememberTool
func (t *RememberTool) Definition() mcp.Tool        // MCP tool definition with schema
func (t *RememberTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

// internal/tools/memory/recall.go
type RecallTool struct {
    store *storage.Store
}

func NewRecallTool(store *storage.Store) *RecallTool
func (t *RecallTool) Definition() mcp.Tool
func (t *RecallTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

// internal/tools/memory/register.go
func RegisterTools(server *mcp.Server, store *storage.Store)
```

## Acceptance Criteria
- [ ] `remember` tool stores a memory with content, type, tags, and related_files
- [ ] `remember` returns the memory ID and timestamp in response
- [ ] `remember` rejects empty content (returns error)
- [ ] `remember` rejects invalid type values (returns error)
- [ ] `recall` finds memories by keyword search (FTS5)
- [ ] `recall` filters by type correctly
- [ ] `recall` filters by date range (since parameter)
- [ ] `recall` excludes soft-deleted memories
- [ ] `recall` returns results sorted by relevance
- [ ] `recall` respects the limit parameter
- [ ] Both tools integrate with the MCP server via RegisterTools()
- [ ] All tests pass

## Implementation Steps
1. Create `internal/tools/memory/remember.go` — RememberTool struct, Definition, Handle
2. Create `internal/tools/memory/recall.go` — RecallTool struct, Definition, Handle
3. Create `internal/tools/memory/register.go` — RegisterTools function
4. Create `internal/tools/memory/remember_test.go`:
   - Test: store a memory, verify it's in the database
   - Test: reject empty content
   - Test: reject invalid type
   - Test: session ID is consistent across calls
   - Test: tags and related_files are stored as JSON
5. Create `internal/tools/memory/recall_test.go`:
   - Test: store 5 memories, recall by keyword, get relevant results
   - Test: filter by type
   - Test: filter by date
   - Test: limit parameter works
   - Test: soft-deleted memories excluded
   - Test: empty query returns error
6. Run all tests

## Testing Requirements
- Unit test: RememberTool.Handle stores memory correctly in SQLite
- Unit test: RememberTool.Handle returns error for empty content
- Unit test: RememberTool.Handle returns error for invalid type "foo"
- Unit test: RecallTool.Handle finds memories by keyword (insert "authentication decision", search "auth")
- Unit test: RecallTool.Handle with type filter only returns matching type
- Unit test: RecallTool.Handle respects limit (insert 10, limit 3, get 3)
- Unit test: RecallTool.Handle excludes soft-deleted memories
- Integration test: RegisterTools + call tool through MCP server handler

## Files to Create/Modify
- `internal/tools/memory/remember.go` — remember tool implementation
- `internal/tools/memory/recall.go` — recall tool implementation
- `internal/tools/memory/register.go` — MCP registration helper
- `internal/tools/memory/remember_test.go` — remember tests
- `internal/tools/memory/recall_test.go` — recall tests

## Notes
- Use `encoding/json` to marshal tags and related_files as JSON strings for the TEXT columns in SQLite.
- For FTS5 search, the query syntax is: `SELECT * FROM memories_fts WHERE memories_fts MATCH ?` with ranked results via `rank`.
- The session ID should be a UUID generated once when the server starts. Use `crypto/rand` to generate it (avoid external UUID libraries).
- For date filtering, use SQLite's `datetime()` function: `created_at > datetime(?)`.
- The `mcp.CallToolRequest` contains the arguments as a map. Parse them according to the input schema.
- Study the mcp-go SDK's `mcp.Tool` and `mcp.CallToolResult` types to match their expected formats.
- Do NOT implement vector similarity search in recall yet — that comes when the embedding pipeline is integrated. For now, FTS5-only search is sufficient.
