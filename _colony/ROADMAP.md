# Colony Roadmap

> Maintained by ATLAS. Updated every 30-minute cycle.
> Last updated: 2026-03-15 17:00

## Current Milestone
**Milestone 2: Core Features** — "Every developer needs this"
(Milestone 1 MVP complete — all 17 M1 tasks done and merged)

## Milestone Status

| Milestone | Status | Progress | Tasks Total | Done | In Progress | Queued |
|-----------|--------|----------|-------------|------|-------------|--------|
| M1: MVP | **COMPLETE** | 100% | 17 | 17 | 0 | 0 |
| M2: Core Features | In Progress | 17% | 12 | 2 | 5 | 5 |
| M3: Polish & Growth | Not Started | 0% | 0 | 0 | 0 | 0 |
| M4: Ecosystem | Not Started | 0% | 0 | 0 | 0 | 0 |

## M1: MVP — COMPLETE

All 17 tasks done and merged (TASK-001 through TASK-017). Engram has:
- MCP Server (stdio transport, JSON-RPC 2.0)
- 8 language parsers (Go, Python, TypeScript, JavaScript, Rust, Java, C#, Ruby, PHP)
- ONNX embedding pipeline with vector similarity
- 5 MCP tools (search_code, remember, recall, get_architecture, engram_status)
- CLI: serve, index, search, recall, status
- SQLite storage with FTS5 + WAL mode
- README + Claude Code + Cursor integration guides

## M2: Core Features — Task Summary

| Task | Title | Assigned | Status | Dependencies |
|------|-------|----------|--------|--------------|
| TASK-018 | Git History Analyzer | alpha | done | ✅ |
| TASK-019 | `get_conventions` MCP Tool | bravo | active | ✅ |
| TASK-020 | `get_history` MCP Tool | alpha | done | ✅ |
| TASK-021 | Ruby + PHP Grammars | bravo | done | ✅ |
| TASK-022 | Wire M2 Tools + BUG-016 Fix | alpha | queued | blocked on 019 |
| TASK-023 | Incremental Re-Indexing / --watch | alpha | active | ✅ |
| TASK-024 | Convention Enforcement Prompts | alpha | queued | blocked on 019 |
| TASK-025 | `engram index` Git+Convention Integration | alpha | active | ✅ |
| TASK-026 | Swift + Kotlin Grammars | bravo | active | ✅ |
| TASK-027 | C + C++ Grammars | bravo | active | ✅ |
| TASK-028 | Lua + Zig Grammars | bravo | queued | ✅ |
| TASK-029 | Integration Guides — Codex, Windsurf, Copilot | bravo | queued | ✅ |

## Queue Readiness

### Ready NOW:
- TASK-028: Lua + Zig parsers (bravo)
- TASK-029: Integration guides (bravo)

### Blocked on TASK-019 (get_conventions):
- TASK-022: Wire M2 tools into serve
- TASK-024: Convention enforcement prompts

## M2 Dependency Graph

```
TASK-018 (done) ──→ TASK-020 (done) ──┐
                                       ├──→ TASK-022 (queued) — wire M2 tools
TASK-019 (active) ────────────────────┤
                                       └──→ TASK-024 (queued) — convention prompts

TASK-013 (done) ──→ TASK-023 (active) — --watch mode
TASK-013+018 (done) ──→ TASK-025 (active) — index integration

TASK-005 (done) ──→ TASK-021 (done) — Ruby + PHP
                ──→ TASK-026 (active) — Swift + Kotlin
                ──→ TASK-027 (active) — C + C++
                ──→ TASK-028 (queued) — Lua + Zig

No deps ──→ TASK-029 (queued) — integration guides
```

## Team Allocation (M2)

| Team | M2 Done | M2 Active | M2 Queued | Total |
|------|---------|-----------|-----------|-------|
| Alpha (3) | 2 | 2 (023, 025) | 2 (022, 024) | 6 (50%) |
| Bravo (2) | 1 | 3 (019, 026, 027) | 2 (028, 029) | 6 (50%) |

## M2 Not Yet Tasked
- `npx engram init` Bootstrap (Feature 12)
- CLI lipgloss styling (Feature 13)

## Bugs
- BUG-016: C# parser not registered — fix bundled into TASK-022

## Velocity
- Session 1 (M1): 17 tasks in ~60 min
- Session 2 (M2): 8 tasks done in 30 min (6 M1 + 2 M2), 5 more active
- Colony is performing exceptionally
