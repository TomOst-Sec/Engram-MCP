# CEO Directive

> Last updated: 2026-03-15
> Status: ACTIVE

## Strategic Priority: Milestone 1 MVP — Foundation First

ATLAS, this is Day 0. No code exists yet. Your first batch of tasks must lay the foundation before anything else can happen. Do not jump to features — build the skeleton first.

## Task Generation Order

Generate tasks in this dependency order. Each batch should only contain tasks whose dependencies are met.

### Batch 1: Project Foundation (generate NOW)
1. **Go project initialization** — `go mod init`, directory structure (`cmd/engram/`, `internal/`, `pkg/`), basic `main.go` with cobra CLI skeleton, `Makefile` with build/test/lint targets. This is task zero — everything depends on it.
2. **SQLite storage layer** — Database connection manager, WAL mode setup, schema v1 migration (memories, code_index, conventions, architecture, git_context tables), connection pooling, basic CRUD operations. Use `mattn/go-sqlite3`.
3. **MCP server skeleton** — JSON-RPC 2.0 stdio transport using `mark3labs/mcp-go`. Tool registration framework. Health check. Graceful shutdown. No actual tools yet — just the server that can start, discover tools, and respond to pings.
4. **Configuration system** — `engram.json` config file loading, CLI flag parsing, sensible defaults. Config struct with repo path, DB path (`~/.engram/<repo-hash>/`), log level, server mode.

### Batch 2: Core Indexing (after Batch 1 merges)
5. **Tree-sitter AST parser** — Go + TypeScript + Python grammars first. Extract functions, types, imports. Store in code_index table.
6. **ONNX embedding pipeline** — Bundled all-MiniLM-L6-v2, batch inference, vector storage in SQLite.
7. **`engram index` command** — Full repo indexer that walks files, parses AST, generates embeddings, stores everything.

### Batch 3: Tools (after Batch 2 merges)
8. **search_code tool** — Hybrid FTS5 + vector similarity search
9. **remember/recall tools** — Memory storage and retrieval
10. **get_architecture tool** — Import graph, module map

### Batch 4: Polish MVP
11. **CLI commands** — serve, search, status with lipgloss output
12. **README + integration guides**

## Team Assignment Guidance

- **Alpha team (3 instances, opus):** Complex systems work — MCP server, tree-sitter integration, ONNX embeddings, search_code tool. These require deep reasoning.
- **Bravo team (2 instances, sonnet):** Storage layer, config system, CLI commands, remember/recall tools, README. Important but more straightforward implementation.

## Critical Rules

1. **No file conflicts between teams.** Alpha and Bravo must not touch the same files in the same batch. Plan the directory structure so ownership is clear.
2. **Tests are mandatory.** Every task must include test files. TDD is not optional.
3. **Keep tasks small.** Each task should be completable in 30-60 minutes by a single coder. If a feature is too large, split it into 2-3 tasks.
4. **Dependencies must be explicit.** If Task B requires Task A's code to exist, say so in the Dependencies field.
5. **Batch 1 is the priority.** Generate Batch 1 immediately. Do not generate Batch 2 tasks until Batch 1 is fully merged.

## What I'm Watching For

- Velocity: Are tasks getting completed and merged, or stuck in review?
- Quality: Is AUDIT rejecting too many tasks? If so, tasks may be underspecified.
- Balance: Are both teams busy, or is one idle?
- Scope: Are coders staying within task boundaries?

I'll check back in 60 minutes. Get the foundation laid.

— CEO
