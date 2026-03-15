# Colony Roadmap

> Maintained by ATLAS. Updated every 30-minute cycle.
> Last updated: 2026-03-15 18:30

## Status: ALL MILESTONES SUBSTANTIALLY COMPLETE

44 of 45 tasks done. One P0 bug fix remaining (TASK-045: FTS5 build fix).
All 20 features from GOALS.md are implemented. No more feature tasks will be generated per CEO directive.

## Milestone Status

| Milestone | Status | Tasks | Done |
|-----------|--------|-------|------|
| M1: MVP | **COMPLETE** | 17 | 17 |
| M2: Core Features | **COMPLETE** (pending TASK-045 bug fix) | 18 | 17 |
| M3: Polish & Growth | **COMPLETE** | 8 | 8 |
| M4: Ecosystem | **COMPLETE** | 1 | 1 |

## Remaining Work

### TASK-045: FTS5 Build Fix (P0 — SHIP BLOCKER)
- Create `internal/storage/cgo_flags.go` with CGo CFLAGS directive
- Verify `go test ./...` passes without build tags
- In queue, assigned to alpha

### BUG-036: FTS5 Fix Incomplete
- TASK-036 fix was insufficient (Makefile/envrc only, no source file)
- TASK-045 is the correct re-do
- Resolved when TASK-045 merges

## Feature Coverage (20/20 GOALS.md Features)

| # | Feature | Tasks | Status |
|---|---------|-------|--------|
| 1 | MCP Server Core | 004, 012 | ✅ |
| 2 | search_code Tool | 010 | ✅ |
| 3 | Tree-Sitter Indexer (15 languages) | 005, 009, 014, 016, 021, 026, 027, 028, 037 | ✅ |
| 4 | ONNX Embeddings | 006 | ✅ |
| 5 | get_architecture Tool | 011 | ✅ |
| 6 | remember Tool | 007 | ✅ |
| 7 | recall Tool | 007 | ✅ |
| 8 | get_conventions Tool | 019 | ✅ |
| 9 | SQLite Storage | 003 | ✅ |
| 10 | Git History Analyzer | 018, 020 | ✅ |
| 11 | --watch Mode | 023 | ✅ |
| 12 | npx engram init | 030 | ✅ |
| 13 | Full CLI | 008, 013, 015, 031, 038 | ✅ |
| 14 | HTTP/SSE Transport | 032 | ✅ |
| 15 | Convention Prompts | 024 | ✅ |
| 16 | TUI Dashboard | 033, 041 | ✅ |
| 17 | Multi-Repo Support | 035 | ✅ |
| 18 | Ollama Integration | 034 | ✅ |
| 19 | Community Conventions | 039 | ✅ |
| 20 | CI/CD Memory Hook | 044 | ✅ |

## Additional Completed Work
- TASK-001: Project foundation
- TASK-002: Configuration system
- TASK-017, 029: Documentation + integration guides (Claude Code, Cursor, Codex, Windsurf, Copilot)
- TASK-022, 025: Wiring + integration
- TASK-036, 037: Bug fixes (FTS5 build tag, parser registration)
- TASK-040: Benchmark suite
- TASK-042: GoReleaser + packaging
- TASK-043: Docker image

## Colony Performance

| Metric | Value |
|--------|-------|
| Total tasks generated | 45 |
| Tasks completed | 44 |
| Tasks remaining | 1 (TASK-045) |
| Total duration | ~3.5 hours |
| Avg time per task | ~5 min (including review) |
| Rejection rate | ~15% (all resolved) |
| Languages supported | 15 |
| MCP tools | 7 |
| CLI commands | 10 |

## What Happens After TASK-045

Per CEO directive:
1. TASK-045 merges → FTS5 bug resolved
2. Final end-to-end validation
3. Generate release report
4. Colony work complete for v0.1.0
