# Colony Roadmap

> Maintained by ATLAS. Updated every 30-minute cycle.
> Last updated: 2026-03-15 16:30

## Current Milestone
**Milestone 1: MVP** — "It works, it's fast, it's useful" (nearly complete)

## Milestone Status

| Milestone | Status | Progress | Tasks Total | Done | In Progress | Queued |
|-----------|--------|----------|-------------|------|-------------|--------|
| M1: MVP | Nearly Complete | 76% | 17 | 13 | 0 | 4 |
| M2: Core Features | Starting | 0% | 4 | 0 | 0 | 4 |
| M3: Polish & Growth | Not Started | 0% | 0 | 0 | 0 | 0 |
| M4: Ecosystem | Not Started | 0% | 0 | 0 | 0 | 0 |

## Task Summary

| Task | Title | Assigned | Status | Dependencies | Branch |
|------|-------|----------|--------|--------------|--------|
| TASK-001 | Project Foundation — Go Module, Dirs, Makefile | alpha | done | none | merged |
| TASK-002 | Configuration System — JSON Loading, Defaults | bravo | done | TASK-001 ✅ | merged |
| TASK-003 | SQLite Storage — Schema v1, WAL, Migrations, FTS5 | alpha | done | TASK-001 ✅ | merged |
| TASK-004 | MCP Server Core — JSON-RPC Stdio Transport | alpha | done | TASK-001 ✅ | merged |
| TASK-005 | Tree-Sitter Parser — Go and Python Grammars | bravo | done | TASK-003 ✅ | merged |
| TASK-006 | ONNX Embedding Pipeline — Vectors, Similarity | alpha | done | TASK-003 ✅ | merged |
| TASK-007 | Remember + Recall MCP Tools — Memory System | bravo | done | TASK-003 ✅, TASK-004 ✅ | merged |
| TASK-008 | CLI `serve` Command — MCP Server Startup | alpha | done | TASK-004 ✅ | merged |
| TASK-009 | TypeScript + JavaScript Tree-Sitter Grammars | bravo | done | TASK-005 ✅ | merged |
| TASK-010 | `search_code` MCP Tool — Hybrid FTS5 + Vector | alpha | done | TASK-003 ✅, TASK-005 ✅, TASK-006 ✅ | merged |
| TASK-011 | `get_architecture` MCP Tool — Module Detection | alpha | done | TASK-003 ✅, TASK-005 ✅ | merged |
| TASK-012 | Integration Wire-Up — Connect Tools to Serve | alpha | done | TASK-008 ✅ | merged |
| TASK-013 | `engram index` CLI — Full Repository Indexer | alpha | queued | TASK-003 ✅, TASK-005 ✅, TASK-006 ✅ | — |
| TASK-014 | Rust + Java Tree-Sitter Grammars | bravo | queued | TASK-005 ✅ | — |
| TASK-015 | CLI search, recall, status Commands | alpha | queued | TASK-003 ✅, TASK-008 ✅ | — |
| TASK-016 | C# Tree-Sitter Grammar | bravo | queued | TASK-005 ✅ | — |
| TASK-017 | README and Integration Guides | bravo | done | TASK-012 ✅ | merged |
| TASK-018 | Git History Analyzer — Blame, Hotspots | alpha | queued | TASK-003 ✅ | — |
| TASK-019 | `get_conventions` MCP Tool — Pattern Inference | bravo | queued | TASK-003 ✅, TASK-005 ✅ | — |
| TASK-020 | `get_history` MCP Tool — Git History via MCP | alpha | queued | TASK-018 | — |
| TASK-021 | Ruby + PHP Tree-Sitter Grammars | bravo | queued | TASK-005 ✅ | — |

## Dependency Graph

```
M1 (nearly complete):
  TASK-001 (done) ──┬──→ TASK-002 (done)
                    ├──→ TASK-003 (done) ──┬──→ TASK-005 (done) ──┬──→ TASK-009 (done)
                    │                      │                      ├──→ TASK-014 (queued)
                    │                      │                      ├──→ TASK-016 (queued)
                    │                      ├──→ TASK-006 (done)
                    │                      ├──→ TASK-007 (done)
                    │                      ├──→ TASK-010 (done)
                    │                      ├──→ TASK-011 (done)
                    │                      ├──→ TASK-013 (queued)
                    │                      └──→ TASK-015 (queued)
                    └──→ TASK-004 (done) ──┬──→ TASK-007 (done)
                                           ├──→ TASK-008 (done) ──→ TASK-012 (done) ──→ TASK-017 (done)
                                           └──→ TASK-015 (queued)

M2 (starting):
  TASK-003 (done) ──→ TASK-018 (queued) ──→ TASK-020 (queued)
  TASK-003+005 (done) ──→ TASK-019 (queued)
  TASK-005 (done) ──→ TASK-021 (queued)
```

## Team Allocation

| Team | Instances | Tasks Done | Tasks Queued | Total |
|------|-----------|------------|--------------|-------|
| Alpha (3) | alpha-1, alpha-2, alpha-3 | 8 | 4 (013, 015, 018, 020) | 12 (57%) |
| Bravo (2) | bravo-1, bravo-2 | 5 | 4 (014, 016, 019, 021) | 9 (43%) |

## Queue Status

### M1 Remaining (all unblocked — ready to claim NOW)
- TASK-013: `engram index` CLI (alpha) — deps met
- TASK-014: Rust + Java parsers (bravo) — deps met
- TASK-015: CLI search/recall/status (alpha) — deps met
- TASK-016: C# parser (bravo) — deps met

### M2 First Batch (3 of 4 unblocked)
- TASK-018: Git History Analyzer (alpha) — deps met, **can start now**
- TASK-019: `get_conventions` Tool (bravo) — deps met, **can start now**
- TASK-020: `get_history` MCP Tool (alpha) — blocked on TASK-018
- TASK-021: Ruby + PHP parsers (bravo) — deps met, **can start now**

## M1 Completion Criteria
All of these are done or queued:
- [x] MCP Server Core (Feature 1)
- [x] Tree-Sitter AST Indexer (Feature 3) — 4 languages done, Rust/Java/C# queued
- [x] ONNX Embedding Pipeline (Feature 4)
- [x] `search_code` Tool (Feature 2)
- [x] `get_architecture` Tool (Feature 5)
- [x] `remember` + `recall` Tools (Features 6+7)
- [x] Persistent SQLite Storage (Feature 9)
- [ ] CLI: `engram index` (TASK-013 queued)
- [ ] CLI: `engram search`, `engram recall`, `engram status` (TASK-015 queued)
- [x] README + Integration Guides (TASK-017 done)
- [ ] Rust + Java grammars (TASK-014 queued)
- [ ] C# grammar (TASK-016 queued)

## M2 Task Pipeline (not yet generated)
Future M2 tasks to generate next cycle:
- Incremental Re-Indexing / --watch mode (Feature 11) — depends on TASK-013
- Convention Enforcement Prompts (Feature 15) — depends on TASK-019
- Wire M2 tools into serve command (depends on TASK-018, 019, 020)
- Swift + Kotlin tree-sitter grammars
- C + C++ tree-sitter grammars
- Lua + Zig tree-sitter grammars
- `npx engram init` Bootstrap (Feature 12)
- Full CLI lipgloss styling (Feature 13)
- Integration guides: Codex, Windsurf, Copilot

## Velocity
- Session 1: 13 tasks completed in ~60 minutes (~4.6 min/task including review)
- Rejection rate: 31% first-attempt (4/13 rejected then fixed). Primary cause: go.mod merge conflicts.
- Colony is performing well. Both teams are productive.
