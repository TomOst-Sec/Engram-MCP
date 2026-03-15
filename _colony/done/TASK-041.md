# TASK-041: TUI Dashboard — Conventions + Architecture Panels

**Priority:** P2
**Assigned:** alpha
**Milestone:** M3: Polish & Growth
**Dependencies:** TASK-033
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
TASK-033 built the TUI foundation with Status and Memories panels. This task completes the TUI with the remaining two panels: Conventions viewer (see detected patterns with confidence) and Architecture viewer (ASCII module dependency graph). After this, `engram tui` has all 4 panels functional.

## Specification

### Conventions Panel (Tab 3): `internal/tui/conventions.go`

```
┌─ Conventions ── [l] filter language ── [c] filter category ──┐
│                                                               │
│ Naming                                                        │
│   ✓ Go functions use camelCase         (95%)  [go]           │
│   ✓ Go types use PascalCase            (98%)  [go]           │
│   ✓ Python functions use snake_case    (92%)  [python]       │
│   ✓ TS components use PascalCase       (88%)  [typescript]   │
│                                                               │
│ Testing                                                       │
│   ✓ Go tests use Test prefix           (100%) [go]           │
│   ✓ Python tests use test_ prefix      (88%)  [python]       │
│                                                               │
│ Documentation                                                 │
│   ○ Go functions: 73% have doc comments       [go]           │
│   ○ Python: 55% docstring coverage            [python]       │
│                                                               │
│ 12 conventions │ showing all languages                        │
└───────────────────────────────────────────────────────────────┘
```

Features:
- Group conventions by category
- Show confidence as percentage with visual indicator (✓ high, ○ medium)
- Filter by language (`l` key)
- Filter by category (`c` key)
- Color-code confidence (green >80%, yellow 60-80%)

### Architecture Panel (Tab 4): `internal/tui/architecture.go`

```
┌─ Architecture ── [e] expand ── [d] show deps ────────────────┐
│                                                               │
│ Modules (12 detected)                                         │
│                                                               │
│ > cmd/engram/          [CLI]         complexity: 3            │
│   internal/config/     [Config]      complexity: 1            │
│   internal/storage/    [Storage]     complexity: 4            │
│   internal/parser/     [Parser]      complexity: 5            │
│   internal/embeddings/ [Embeddings]  complexity: 3            │
│   internal/mcp/        [MCP Server]  complexity: 2            │
│   internal/tools/      [MCP Tools]   complexity: 4            │
│   internal/git/        [Git]         complexity: 2            │
│   internal/conventions/[Conventions] complexity: 3            │
│   internal/tui/        [TUI]         complexity: 2            │
│   internal/indexer/    [Indexer]     complexity: 3            │
│   internal/cli/        [CLI Styles]  complexity: 1            │
│                                                               │
│ Press 'e' on a module to see exports, 'd' for dependencies   │
└───────────────────────────────────────────────────────────────┘
```

Features:
- List modules with description and complexity score
- `Enter` or `e` on a module shows its exports (functions, types)
- `d` shows module dependencies (imports from other modules)
- Vim-style navigation (j/k)

### Integration with App

Update `internal/tui/app.go`:
- Add ConventionsModel and ArchitectureModel as panel models
- Wire Tab 3 and Tab 4 to these panels
- Update tab names to ["Status", "Memories", "Conventions", "Architecture"]

## Acceptance Criteria
- [ ] Conventions panel shows conventions grouped by category
- [ ] Conventions panel displays confidence percentages
- [ ] Conventions panel filters by language and category
- [ ] Architecture panel lists modules with complexity scores
- [ ] Architecture panel shows module exports on expand
- [ ] Architecture panel shows dependencies on 'd' key
- [ ] Tab navigation works for all 4 panels
- [ ] All tests pass

## Implementation Steps
1. Create `internal/tui/conventions.go` — ConventionsModel
2. Create `internal/tui/architecture.go` — ArchitectureModel
3. Update `internal/tui/app.go` — wire new panels into tabs
4. Create `internal/tui/conventions_test.go` — convention panel tests
5. Create `internal/tui/architecture_test.go` — architecture panel tests
6. Run all tests

## Files to Create/Modify
- `internal/tui/conventions.go` — conventions panel
- `internal/tui/architecture.go` — architecture panel
- `internal/tui/app.go` — wire new panels
- `internal/tui/conventions_test.go` — tests
- `internal/tui/architecture_test.go` — tests

## Notes
- Query conventions table for Tab 3 data. Query architecture table for Tab 4 data.
- Reuse the bubbletea patterns from TASK-033 (StatusModel, MemoriesModel).
- The architecture ASCII graph is optional for this task — a simple list view is sufficient. An ASCII dependency tree can be a future enhancement.
- Open database read-only (consistent with TASK-033 TUI approach).

---
## Completion Notes
- **Completed by:** alpha-3
- **Date:** 2026-03-15 18:07:15
- **Branch:** task/041
