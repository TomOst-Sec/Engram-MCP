# TASK-033: TUI Dashboard Foundation — Bubbletea Interactive Browser

**Priority:** P2
**Assigned:** alpha
**Milestone:** M3: Polish & Growth
**Dependencies:** TASK-003
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
Feature 16 from GOALS.md. `engram tui` launches an interactive terminal UI using bubbletea. The TUI provides panels for: memory browser (search, filter, delete), convention viewer (patterns with confidence), index status (files, languages, DB size), and architecture viewer (module dependency graph). This task builds the foundation: the TUI framework, tab navigation, and the first two panels (index status + memory browser).

## Specification

### Dependencies
- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/bubbles` — UI components (table, textinput, viewport)
- `github.com/charmbracelet/lipgloss` — styling (already in use)

### TUI Structure: `internal/tui/`

```
internal/tui/
├── app.go          — Main bubbletea model, tab switching
├── tabs.go         — Tab bar component
├── status.go       — Index status panel
├── memories.go     — Memory browser panel
├── styles.go       — TUI-specific lipgloss styles
└── keymap.go       — Keyboard bindings
```

### App Model

```go
type App struct {
    store    *storage.Store
    repoRoot string
    tabs     []string       // ["Status", "Memories", "Conventions", "Architecture"]
    activeTab int

    // Panel models
    statusPanel   StatusModel
    memoriesPanel MemoriesModel

    width, height int
}
```

### Tab Navigation
- `Tab` / `Shift+Tab` — switch between panels
- `1-4` — jump to specific tab
- `q` / `Ctrl+C` — quit
- `j/k` or `↑/↓` — navigate within panel
- `/` — search within panel

### Status Panel (Tab 1)
Displays real-time index statistics:
```
┌─ Index Status ──────────────────────────────────┐
│ Repository: /path/to/repo                       │
│ Database:   ~/.engram/abc123/engram.db (4.2 MB)  │
│                                                  │
│ Files indexed:    342                             │
│ Symbols:          2,847                           │
│ Languages:        Go (120), Python (85), TS (137) │
│ Embeddings:       2,847 / 2,847 (100%)           │
│ Last indexed:     2026-03-15 14:30:00            │
│                                                  │
│ Memories:         42                             │
│ Conventions:      8 patterns                     │
│ Git history:      342 files analyzed             │
└──────────────────────────────────────────────────┘
```

### Memories Panel (Tab 2)
Interactive memory browser with search and filter:
```
┌─ Memories ── [/] search ── [t] filter type ─────┐
│ Search: _                                        │
│                                                  │
│ > #42  [decision]  2026-03-14                   │
│   Decided to use SQLite with FTS5 for storage    │
│                                                  │
│   #38  [learning]  2026-03-13                   │
│   FTS5 rank returns negative scores              │
│                                                  │
│   #35  [bugfix]   2026-03-12                    │
│   Fixed race condition in concurrent writes      │
│                                                  │
│ 42 memories │ showing all types                  │
└──────────────────────────────────────────────────┘
```

Features:
- Search memories by text (live filtering)
- Filter by type (decision, bugfix, refactor, learning, convention)
- `d` to soft-delete a memory
- `Enter` to view full memory details

### CLI Command: `cmd/engram/tui.go`

```go
var tuiCmd = &cobra.Command{
    Use:   "tui",
    Short: "Launch interactive dashboard",
    RunE:  runTUI,
}
```

## Acceptance Criteria
- [ ] `engram tui` launches an interactive terminal UI
- [ ] Tab bar shows panel names with active tab highlighted
- [ ] Tab/Shift+Tab switches between panels
- [ ] Status panel shows index stats from the database
- [ ] Memories panel lists memories with type and date
- [ ] Memories panel supports text search (live filtering)
- [ ] `q` or Ctrl+C quits the TUI
- [ ] TUI handles terminal resize
- [ ] All tests pass

## Implementation Steps
1. `go get github.com/charmbracelet/bubbletea github.com/charmbracelet/bubbles`
2. Create `internal/tui/styles.go` — TUI color scheme and border styles
3. Create `internal/tui/keymap.go` — key bindings
4. Create `internal/tui/tabs.go` — tab bar component
5. Create `internal/tui/status.go` — StatusModel with Init, Update, View
6. Create `internal/tui/memories.go` — MemoriesModel with search, filter, list
7. Create `internal/tui/app.go` — main App model, tab switching, keyboard dispatch
8. Create `cmd/engram/tui.go` — tui subcommand
9. Register in `cmd/engram/main.go`
10. Create `internal/tui/app_test.go`:
    - Test: App initializes with correct tab count
    - Test: Tab switching cycles through panels
    - Test: Status panel queries database for stats
11. Run all tests

## Testing Requirements
- Unit test: App model initializes correctly
- Unit test: Tab switching with key events
- Unit test: StatusModel queries produce expected output format
- Unit test: MemoriesModel filters by type correctly
- Unit test: TUI command is registered

## Files to Create/Modify
- `internal/tui/app.go` — main TUI app model
- `internal/tui/tabs.go` — tab bar
- `internal/tui/status.go` — status panel
- `internal/tui/memories.go` — memory browser
- `internal/tui/styles.go` — TUI styles
- `internal/tui/keymap.go` — key bindings
- `internal/tui/app_test.go` — tests
- `cmd/engram/tui.go` — CLI subcommand
- `cmd/engram/main.go` — register tui command

## Notes
- This task only implements 2 of 4 panels (Status + Memories). Conventions and Architecture panels will be separate tasks.
- Study bubbletea examples for the tab-switching pattern. The `tabs` example in the bubbletea repo is a good reference.
- The TUI should open the database in read-only mode (no WAL locking issues while serve is running).
- Memory deletion (`d` key) should use soft-delete (set deleted_at timestamp), not hard delete.
- Keep the TUI simple for V1 — no fancy animations or transitions. Just clean, readable panels with vim-style navigation.

---
## Completion Notes
- **Completed by:** alpha-3
- **Date:** 2026-03-15 17:47:25
- **Branch:** task/033
