# TASK-022: Wire M2 Tools into Serve Command — History + Conventions

**Priority:** P0
**Assigned:** alpha
**Milestone:** M2: Core Features
**Dependencies:** TASK-019, TASK-020
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
TASK-020 created `internal/tools/history/` (get_history tool) and TASK-019 created `internal/tools/conventions/` (get_conventions tool). Neither is wired into the serve command yet. This task adds both to `cmd/engram/serve.go` so they're available to MCP clients. Same pattern as the existing tool registrations. Also fixes BUG-016: C# parser not registered in `NewDefaultRegistry()`.

## Specification

### 1. Add tool registrations to serve.go

Add after the existing `architecture.RegisterTools(...)` line:

```go
import (
    // add these to existing imports
    historytools "github.com/TomOst-Sec/colony-project/internal/tools/history"
    conventiontools "github.com/TomOst-Sec/colony-project/internal/tools/conventions"
)

// In runServe(), after architecture.RegisterTools:
historytools.RegisterTools(server, store)
conventiontools.RegisterTools(server, store, repoRoot)
```

Check the actual `RegisterTools` signatures in the history and conventions packages — match whatever parameters they require.

### 2. Fix BUG-016: Register C# parser

Add to `internal/parser/registry.go` in `NewDefaultRegistry()`:
```go
r.Register(NewCSharpParser())
```

After the existing Java parser registration.

### 3. Update serve startup message

Update the startup message to show the total tool count:
```go
fmt.Fprintf(os.Stderr, "Engram MCP server starting (version %s, transport: %s, repo: %s, tools: 7)\n", ...)
```

## Acceptance Criteria
- [ ] `engram serve` registers get_history tool
- [ ] `engram serve` registers get_conventions tool
- [ ] MCP clients can discover and call both new tools
- [ ] `NewDefaultRegistry()` includes `NewCSharpParser()` (BUG-016 fixed)
- [ ] Existing tools still work (search_code, remember, recall, get_architecture, engram_status)
- [ ] All tests pass: `go test ./cmd/engram/ ./internal/parser/`

## Implementation Steps
1. Read `internal/tools/history/register.go` to confirm RegisterTools signature
2. Read `internal/tools/conventions/register.go` to confirm RegisterTools signature
3. Add imports to `cmd/engram/serve.go`
4. Add RegisterTools calls after architecture registration
5. Add `r.Register(NewCSharpParser())` to `internal/parser/registry.go`
6. Run `go test ./cmd/engram/ ./internal/parser/` — all pass
7. Run `go build ./cmd/engram` — build succeeds

## Testing Requirements
- Unit test: serve command still builds and serve_test.go passes
- Unit test: registry_test.go confirms C# extension (.cs) has a registered parser
- Build test: `go build ./cmd/engram` succeeds with no import errors

## Files to Create/Modify
- `cmd/engram/serve.go` — add imports + RegisterTools calls for history + conventions
- `internal/parser/registry.go` — add `r.Register(NewCSharpParser())` to NewDefaultRegistry

## Notes
- This is a small wiring task. Do NOT create new packages or modify tool implementations.
- If the history or conventions RegisterTools signatures differ from what's described here, adapt to match their actual signatures. Read the register.go files.
- After this task, `engram serve` should expose 7 tools: engram_status, search_code, remember, recall, get_architecture, get_history, get_conventions.
- Delete `_colony/bugs/BUG-016.md` after fixing — ATLAS will clean it up.

---
## Completion Notes
- **Completed by:** alpha-1
- **Date:** 2026-03-15 17:07:22
- **Branch:** task/022
