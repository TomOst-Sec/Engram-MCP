# TASK-012: Integration Wire-Up — Connect All MCP Tools to Serve Command

**Priority:** P0
**Assigned:** alpha
**Milestone:** M1: MVP
**Dependencies:** TASK-008, TASK-003, TASK-007, TASK-010, TASK-011
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
The `engram serve` command currently starts the MCP server but only registers the built-in `engram_status` tool. None of the real tools (search_code, remember, recall, get_architecture) are connected. This task wires everything together: open the database, create the embedder, detect the repo root, and register all tools. After this task, `engram serve` will expose the full MCP tool suite to AI coding agents.

## Specification
Modify `cmd/engram/serve.go` to initialize all dependencies and register all MCP tools.

### Current State (what exists)
```go
// cmd/engram/serve.go — runServe currently does:
server := mcpmcp.New("engram", version)
mcpmcp.RegisterBuiltinTools(server)
server.ServeStdio()
```

### Target State (what it should do)
```go
func runServe(cmd *cobra.Command, args []string) error {
    // 1. Detect repo root (walk up from cwd looking for .git/)
    repoRoot, err := detectRepoRoot()

    // 2. Load configuration
    cfg, err := config.Load(repoRoot)

    // 3. Open storage (create DB dir if needed)
    dbPath := cfg.DatabasePath
    if dbPath == "" {
        dbPath = filepath.Join(config.DatabaseDir(repoRoot), "engram.db")
    }
    store, err := storage.Open(dbPath)
    defer store.Close()

    // 4. Create embedder (NoOp if model unavailable)
    var emb *embeddings.Embedder
    emb, err = embeddings.New("")  // empty = use default/bundled model
    if err != nil {
        fmt.Fprintf(os.Stderr, "Warning: embeddings unavailable (%v), using FTS5-only search\n", err)
        emb = nil  // tools handle nil embedder gracefully
    }

    // 5. Create MCP server
    server := mcpmcp.New("engram", version)
    mcpmcp.RegisterBuiltinTools(server)

    // 6. Register all tools
    searchtools.RegisterTools(server, store, emb, repoRoot)
    memory.RegisterTools(server, store)
    architecture.RegisterTools(server, store, repoRoot, goModulePath(repoRoot))

    // 7. Start stdio transport
    fmt.Fprintf(os.Stderr, "Engram MCP server starting (version %s, transport: %s, repo: %s)\n", version, transport, repoRoot)
    return server.ServeStdio()
}
```

### Helper Functions to Add

**`detectRepoRoot() (string, error)`** — Walk up from the current working directory looking for `.git/`. Return the directory containing `.git/`. If not found, return cwd with a warning.

**`goModulePath(repoRoot string) string`** — Read `go.mod` in repoRoot, parse the `module` line, return the module path. Return empty string if not a Go project. This is needed by the architecture tool for import path resolution.

### Import Paths
The serve.go file needs these new imports:
```go
import (
    "github.com/TomOst-Sec/colony-project/internal/config"
    "github.com/TomOst-Sec/colony-project/internal/embeddings"
    "github.com/TomOst-Sec/colony-project/internal/storage"
    "github.com/TomOst-Sec/colony-project/internal/tools/architecture"
    "github.com/TomOst-Sec/colony-project/internal/tools/memory"
    searchtools "github.com/TomOst-Sec/colony-project/internal/tools/search"
)
```

### Error Handling
- If repo root detection fails: use cwd, log warning to stderr
- If config loading fails: use defaults, log warning to stderr
- If storage open fails: this is fatal — return error
- If embedder creation fails: log warning, continue without embeddings (FTS5-only)
- All diagnostic output to stderr (stdout is JSON-RPC transport)

## Acceptance Criteria
- [ ] `engram serve` opens the SQLite database in `~/.engram/<repo-hash>/engram.db`
- [ ] `engram serve` registers search_code, remember, recall, and get_architecture tools
- [ ] `engram serve` detects the git repo root from cwd
- [ ] `engram serve` continues working if embedder fails (graceful degradation)
- [ ] `engram serve` fails with clear error if database cannot be opened
- [ ] `engram serve` loads config from engram.json if present, uses defaults otherwise
- [ ] Database is properly closed on shutdown
- [ ] All diagnostic output goes to stderr, not stdout
- [ ] All tests pass with `go test ./cmd/engram/`

## Implementation Steps
1. Add `detectRepoRoot()` function to `cmd/engram/serve.go` (walk up from cwd looking for .git/)
2. Add `goModulePath()` function (read go.mod, parse module line)
3. Modify `runServe()`:
   - Add repo root detection
   - Add config loading
   - Add storage initialization with defer Close()
   - Add embedder creation with graceful degradation
   - Add all tool registrations
   - Update startup log message to include repo path
4. Update imports for storage, config, embeddings, tools packages
5. Update `cmd/engram/serve_test.go`:
   - Test: detectRepoRoot() finds .git/ from nested directory
   - Test: detectRepoRoot() returns cwd when no .git/ found
   - Test: goModulePath() parses go.mod correctly
   - Test: goModulePath() returns empty for non-Go project
6. Run `go test ./cmd/engram/` — all tests pass
7. Run `go build ./cmd/engram` — build succeeds

## Testing Requirements
- Unit test: detectRepoRoot() from project dir returns dir containing .git/
- Unit test: detectRepoRoot() from non-git dir returns cwd
- Unit test: goModulePath() with go.mod returns correct module path
- Unit test: goModulePath() without go.mod returns empty string
- Unit test: serve command still registers and is findable on root command

## Files to Create/Modify
- `cmd/engram/serve.go` — modify: add imports, detectRepoRoot, goModulePath, update runServe to wire up all components
- `cmd/engram/serve_test.go` — modify: add tests for new helper functions

## Notes
- The `architecture.RegisterTools` takes a `goModulePath` parameter for Go import resolution. Use the `goModulePath()` helper to extract it from go.mod. If it's not a Go project, pass empty string.
- The `search.RegisterTools` import alias must be `searchtools` to avoid collision with the built-in `search` package.
- The embedder `New("")` with empty path should attempt to find a default model. If the ONNX model isn't available (which it won't be until we bundle it), it will fail — that's expected. The NoOpEmbedder from TASK-006 handles this.
- Database path construction: use `config.DatabaseDir(repoRoot)` which returns `~/.engram/<repo-hash>/`. Create that directory with `os.MkdirAll` before opening the database. The storage.Open function may handle this, but be safe.
- Do NOT add new flags to the serve command in this task. The existing --transport and --log-level flags are sufficient.

---
## Completion Notes
- **Completed by:** alpha-2
- **Date:** 2026-03-15 15:45:53
- **Branch:** task/012
