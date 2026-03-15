# TASK-010: `search_code` MCP Tool — Hybrid FTS5 + Vector Similarity Search

**Priority:** P1
**Assigned:** alpha
**Milestone:** M1: MVP
**Dependencies:** TASK-003, TASK-005, TASK-006
**Status:** queued
**Created:** 2026-03-15
**Author:** atlas

## Context
`search_code` is the flagship MCP tool — the primary way AI coding agents find relevant code in a repository. It combines SQLite FTS5 full-text search (fast keyword matching) with ONNX vector similarity search (semantic understanding) to return ranked code snippets. This is the tool that makes AI agents say "Engram helped me find the right code instantly." Target: <200ms response time for repos up to 100K LOC.

## Specification
Create the `internal/tools/search` package implementing the `search_code` MCP tool.

### MCP Tool Definition
**Name:** `search_code`
**Description:** "Search the codebase using natural language or keyword queries. Returns ranked code snippets with file paths, line numbers, and relevance scores."

**Input Schema:**
```json
{
    "type": "object",
    "properties": {
        "query": {
            "type": "string",
            "description": "Natural language or keyword search query"
        },
        "language": {
            "type": "string",
            "description": "Filter by programming language (e.g., 'go', 'python', 'typescript')"
        },
        "symbol_type": {
            "type": "string",
            "enum": ["function", "method", "type", "class", "interface", "import", "export", "test"],
            "description": "Filter by symbol type"
        },
        "directory": {
            "type": "string",
            "description": "Filter to symbols within this directory path"
        },
        "limit": {
            "type": "integer",
            "description": "Maximum results to return (default: 10, max: 50)"
        }
    },
    "required": ["query"]
}
```

### Search Algorithm

The search_code tool uses a **hybrid ranking** approach:

1. **FTS5 Keyword Search:**
   - Query `code_index_fts` with the user's query string
   - Get FTS5 rank scores for each matching row
   - Normalize scores to 0.0–1.0 range

2. **Vector Similarity Search (if embeddings available):**
   - Generate an embedding for the query text using the Embedder
   - Compute cosine similarity against all code_index embeddings (brute force)
   - Get top candidates with similarity scores

3. **Hybrid Ranking:**
   - If both FTS5 and vector results available: `final_score = 0.4 * fts5_score + 0.6 * vector_score`
   - If only FTS5 available (no embeddings): `final_score = fts5_score`
   - If only vector available (FTS5 returns empty): `final_score = vector_score`
   - Merge results by final_score, remove duplicates (same symbol ID)

4. **Apply Filters:**
   - Filter by language if specified
   - Filter by symbol_type if specified
   - Filter by directory prefix if specified

5. **Build Response:**
   - For each result, read the source file to get surrounding context (±3 lines)
   - Return ranked list with metadata

### Response Format
```json
{
    "results": [
        {
            "file_path": "internal/auth/handler.go",
            "symbol_name": "HandleLogin",
            "symbol_type": "function",
            "language": "go",
            "signature": "func HandleLogin(w http.ResponseWriter, r *http.Request) error",
            "start_line": 42,
            "end_line": 87,
            "score": 0.92,
            "context": "// HandleLogin authenticates a user...\nfunc HandleLogin(w http.ResponseWriter, r *http.Request) error {\n    ..."
        }
    ],
    "total_matches": 15,
    "query": "login authentication handler",
    "search_mode": "hybrid"
}
```

### Package Structure
```go
// internal/tools/search/search.go
type SearchTool struct {
    store    *storage.Store
    embedder *embeddings.Embedder  // may be nil (NoOpEmbedder)
    repoRoot string
}

func NewSearchTool(store *storage.Store, embedder *embeddings.Embedder, repoRoot string) *SearchTool
func (t *SearchTool) Definition() mcp.Tool
func (t *SearchTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

// internal/tools/search/ranking.go
func HybridRank(ftsResults []FTSResult, vectorResults []VectorResult) []RankedResult
func NormalizeFTSScores(results []FTSResult) []FTSResult

// internal/tools/search/register.go
func RegisterTools(server *mcp.Server, store *storage.Store, embedder *embeddings.Embedder, repoRoot string)
```

## Acceptance Criteria
- [ ] `search_code` tool is registered as an MCP tool with correct schema
- [ ] FTS5 keyword search returns relevant results for exact matches (search "HandleLogin" finds HandleLogin function)
- [ ] Vector similarity search returns semantically relevant results (search "authentication" finds login-related functions)
- [ ] Hybrid ranking combines FTS5 and vector scores correctly
- [ ] Language filter works (search with language="go" only returns Go symbols)
- [ ] Symbol type filter works (search with symbol_type="function" only returns functions)
- [ ] Directory filter works (search within "internal/auth/" only returns symbols in that path)
- [ ] Limit parameter caps the number of results
- [ ] Response includes file_path, symbol_name, signature, line numbers, and score
- [ ] Graceful degradation: works with FTS5-only when embedder is nil/NoOp
- [ ] All tests pass

## Implementation Steps
1. Create `internal/tools/search/search.go` — SearchTool struct, Definition, Handle
2. Create `internal/tools/search/ranking.go` — HybridRank, NormalizeFTSScores
3. Create `internal/tools/search/register.go` — RegisterTools function
4. Create `internal/tools/search/search_test.go`:
   - Set up test DB with sample code_index rows (manually insert)
   - Test: FTS5 search finds exact keyword match
   - Test: language filter works
   - Test: symbol_type filter works
   - Test: directory filter works
   - Test: limit caps results
   - Test: empty query returns error
   - Test: graceful degradation without embeddings
5. Create `internal/tools/search/ranking_test.go`:
   - Test: HybridRank merges and sorts correctly
   - Test: NormalizeFTSScores produces 0.0–1.0 range
   - Test: duplicate removal works
6. Run all tests

## Testing Requirements
- Unit test: SearchTool.Handle with FTS5 query finds matching symbols
- Unit test: SearchTool.Handle with language filter returns only matching language
- Unit test: SearchTool.Handle with symbol_type filter returns only matching types
- Unit test: SearchTool.Handle with directory filter returns only matching paths
- Unit test: SearchTool.Handle respects limit (insert 20 results, limit 5, get 5)
- Unit test: SearchTool.Handle returns error for empty query
- Unit test: HybridRank correctly weights FTS5 (0.4) and vector (0.6) scores
- Unit test: HybridRank deduplicates by symbol ID
- Unit test: NormalizeFTSScores maps to 0.0–1.0 range
- Unit test: Works correctly when embedder is nil (FTS5-only mode)

## Files to Create/Modify
- `internal/tools/search/search.go` — SearchTool implementation
- `internal/tools/search/ranking.go` — hybrid ranking logic
- `internal/tools/search/register.go` — MCP registration helper
- `internal/tools/search/search_test.go` — search tool tests
- `internal/tools/search/ranking_test.go` — ranking algorithm tests

## Notes
- For FTS5 queries, use `code_index_fts MATCH ?` with `rank` for scoring. The FTS5 rank function returns negative values (more negative = better match). Normalize by taking absolute values and scaling.
- For context snippets, read the source file from disk using the file_path from code_index. Read lines `start_line-3` to `end_line+3`. If the file doesn't exist or can't be read, return the symbol data without context (don't fail the whole search).
- The SearchTool takes an `*embeddings.Embedder` which may be a NoOpEmbedder. Check if Embed() returns nil — if so, skip vector search entirely and use FTS5-only.
- For MVP, brute-force vector search is acceptable. Scan all rows with non-null embeddings, compute cosine similarity, return top-k. This will be replaced with HNSW in Milestone 2.
- Do NOT import or depend on `internal/parser` directly. The search tool reads from the `code_index` table which was populated by the parser + indexer pipeline. The tool is decoupled from parsing.
- The `mcp.CallToolRequest` arguments need to be parsed from the request. Study how TASK-007's remember/recall tools parse their arguments.
