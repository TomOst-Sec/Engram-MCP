# Colony Roadmap

> Maintained by ATLAS. Updated every 30-minute cycle.
> Last updated: 2026-03-15 15:30

## Current Milestone
**Milestone 1: MVP** — "It works, it's fast, it's useful"

## Milestone Status

| Milestone | Status | Progress | Tasks Total | Done | In Progress | Queued |
|-----------|--------|----------|-------------|------|-------------|--------|
| M1: MVP | In Progress | 9% | 11 | 1 | 3 | 7 |
| M2: Core Features | Not Started | 0% | 0 | 0 | 0 | 0 |
| M3: Polish & Growth | Not Started | 0% | 0 | 0 | 0 | 0 |
| M4: Ecosystem | Not Started | 0% | 0 | 0 | 0 | 0 |

## Task Summary

| Task | Title | Assigned | Status | Dependencies | Branch |
|------|-------|----------|--------|--------------|--------|
| TASK-001 | Project Foundation — Go Module, Directory Structure, Makefile | alpha | done | none | task/001 (merged) |
| TASK-002 | Configuration System — engram.json Loading, Defaults, Validation | bravo | active | TASK-001 ✅ | task/002 |
| TASK-003 | SQLite Storage Layer — Schema v1, WAL Mode, Migrations | alpha | active | TASK-001 ✅ | task/003 |
| TASK-004 | MCP Server Core — JSON-RPC 2.0 Stdio Transport | alpha | active | TASK-001 ✅ | task/004 |
| TASK-005 | Tree-Sitter Parser Framework + Go and Python Grammars | bravo | queued | TASK-003 | — |
| TASK-006 | ONNX Embedding Pipeline — Model Loading, Batch Inference | alpha | queued | TASK-003 | — |
| TASK-007 | Remember and Recall MCP Tools — Memory Storage and Retrieval | bravo | queued | TASK-003, TASK-004 | — |
| TASK-008 | CLI `serve` Command — Start MCP Server via Stdio | alpha | queued | TASK-004 | — |
| TASK-009 | TypeScript and JavaScript Tree-Sitter Grammars | bravo | queued | TASK-005 | — |
| TASK-010 | `search_code` MCP Tool — Hybrid FTS5 + Vector Search | alpha | queued | TASK-003, TASK-005, TASK-006 | — |
| TASK-011 | `get_architecture` MCP Tool — Module Map and Import Graph | alpha | queued | TASK-003, TASK-005 | — |

## Dependency Graph

```
TASK-001 (done) ──┬──→ TASK-002 (active)
                  ├──→ TASK-003 (active) ──┬──→ TASK-005 (queued) ──┬──→ TASK-009 (queued)
                  │                        │                        ├──→ TASK-010 (queued)
                  │                        │                        └──→ TASK-011 (queued)
                  │                        ├──→ TASK-006 (queued) ──→ TASK-010 (queued)
                  │                        ├──→ TASK-007 (queued)
                  │                        └──→ TASK-010 (queued)
                  └──→ TASK-004 (active) ──┬──→ TASK-007 (queued)
                                           └──→ TASK-008 (queued)
```

## Team Allocation

| Team | Instances | Tasks Done | Tasks Active | Tasks Queued | Total |
|------|-----------|------------|--------------|--------------|-------|
| Alpha (3) | alpha-1, alpha-2, alpha-3 | 1 | 2 | 4 | 7 (64%) |
| Bravo (2) | bravo-1, bravo-2 | 0 | 1 | 3 | 4 (36%) |

## Batch Status

### Batch 1: Project Foundation (CEO Priority)
- ✅ TASK-001: Project Foundation — **DONE** (alpha-1)
- 🔧 TASK-002: Configuration System — **ACTIVE** (bravo)
- 🔧 TASK-003: SQLite Storage Layer — **ACTIVE** (alpha)
- 🔧 TASK-004: MCP Server Core — **ACTIVE** (alpha)

### Batch 2: Core Indexing + Server Wire-up
- ⏳ TASK-005: Tree-Sitter Parser Framework — blocked on TASK-003
- ⏳ TASK-006: ONNX Embedding Pipeline — blocked on TASK-003
- ⏳ TASK-008: CLI `serve` Command — blocked on TASK-004

### Batch 3: MCP Tools
- ⏳ TASK-007: Remember/Recall Tools — blocked on TASK-003 + TASK-004
- ⏳ TASK-009: TS/JS Grammars — blocked on TASK-005
- ⏳ TASK-010: `search_code` Tool — blocked on TASK-003 + TASK-005 + TASK-006
- ⏳ TASK-011: `get_architecture` Tool — blocked on TASK-003 + TASK-005

## Blocked Items
- All Batch 2/3 tasks blocked on TASK-003 (SQLite Storage) and TASK-004 (MCP Server Core). These are the critical path.
- 1 alpha instance and 1 bravo instance are currently idle waiting for Batch 1 to complete.

## Next Up
- When TASK-003 and TASK-004 merge: TASK-005, TASK-006, TASK-007, TASK-008 become actionable
- Future tasks needed for MVP (not yet generated):
  - `engram index` CLI command (full repo indexer — depends on parser + embeddings)
  - `engram status` CLI command (show index stats)
  - `engram search` CLI command (CLI search interface)
  - Rust and Java tree-sitter grammars (remaining MVP languages)
  - Git History Analyzer
  - README + integration guides
