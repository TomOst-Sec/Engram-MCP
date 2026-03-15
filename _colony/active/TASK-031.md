# TASK-031: CLI Lipgloss Styling — Polished Terminal Output

**Priority:** P2
**Assigned:** alpha
**Milestone:** M2: Core Features
**Dependencies:** TASK-015
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
Feature 13 from GOALS.md. The CLI commands currently use plain `fmt.Printf` for output. This task adds lipgloss styling for clean, professional terminal output across all commands: `engram status`, `engram search`, `engram recall`, `engram index`. Lipgloss is already in go.mod (dependency of cobra). This is polish work — the commands work fine without it, but styled output makes Engram feel professional.

## Specification

### Styling Package: `internal/cli/styles.go`

Create a shared styles package:

```go
package cli

import "github.com/charmbracelet/lipgloss"

var (
    // Colors
    Primary   = lipgloss.Color("#7C3AED")  // Purple
    Secondary = lipgloss.Color("#06B6D4")  // Cyan
    Success   = lipgloss.Color("#22C55E")  // Green
    Warning   = lipgloss.Color("#EAB308")  // Yellow
    Error     = lipgloss.Color("#EF4444")  // Red
    Muted     = lipgloss.Color("#6B7280")  // Gray

    // Styles
    Title = lipgloss.NewStyle().
        Bold(true).
        Foreground(Primary)

    Subtitle = lipgloss.NewStyle().
        Foreground(Secondary)

    Label = lipgloss.NewStyle().
        Foreground(Muted)

    Value = lipgloss.NewStyle().
        Bold(true)

    FilePath = lipgloss.NewStyle().
        Foreground(Secondary).
        Underline(true)

    SymbolName = lipgloss.NewStyle().
        Bold(true).
        Foreground(Primary)

    SuccessText = lipgloss.NewStyle().
        Foreground(Success)

    ErrorText = lipgloss.NewStyle().
        Foreground(Error)

    WarningText = lipgloss.NewStyle().
        Foreground(Warning)
)
```

### Apply Styling to Commands

**`engram status`:**
```
  Engram v0.1.0                          (styled title)
  Repository: /path/to/repo             (label: value)
  Database:   ~/.engram/abc123/engram.db (4.2 MB)

  Index                                  (styled subtitle)
    Files indexed:    342
    Symbols:          2,847
    Languages:        Go (120), Python (85), TypeScript (137)

  Memories                               (styled subtitle)
    Total:            42
    Types:            decision (15), bugfix (10), learning (12)
```

**`engram search`:**
```
  internal/auth/handler.go:42            (filepath styled)
  HandleLogin  function                  (symbol styled)
  func HandleLogin(w http.ResponseWriter, r *http.Request) error

  Found 2 results for "auth" (0.8ms)    (muted footer)
```

**`engram index`:**
```
  Engram indexing /path/to/repo...       (title)
  342/342 files processed                (progress)
  ✓ Index complete                       (success)
    Files: 342  Symbols: 2,847  Duration: 12.3s
```

### Dependencies
- `github.com/charmbracelet/lipgloss` — already in go.mod (cobra dependency)

## Acceptance Criteria
- [ ] `internal/cli/styles.go` defines shared color constants and text styles
- [ ] `engram status` output is styled with colors and formatting
- [ ] `engram search` output highlights file paths and symbol names
- [ ] `engram recall` output highlights memory types and dates
- [ ] `engram index` output shows styled progress and completion
- [ ] Plain text fallback when terminal doesn't support colors (lipgloss handles this)
- [ ] All existing tests still pass
- [ ] No functional changes — only cosmetic improvements

## Implementation Steps
1. Add lipgloss to go.mod if not present: `go get github.com/charmbracelet/lipgloss`
2. Create `internal/cli/styles.go` — color constants and style definitions
3. Update `cmd/engram/status.go` — apply styles to output
4. Update `cmd/engram/search.go` — apply styles to search results
5. Update `cmd/engram/recall.go` — apply styles to memory results
6. Update `cmd/engram/index.go` — apply styles to indexing output
7. Create `internal/cli/styles_test.go`:
   - Test: styles render without error
   - Test: styled strings contain expected content
8. Run all tests

## Testing Requirements
- Unit test: Style constants are defined (not zero values)
- Unit test: Title.Render("test") produces non-empty output
- Regression test: all existing command tests pass

## Files to Create/Modify
- `internal/cli/styles.go` — shared lipgloss styles (create new)
- `internal/cli/styles_test.go` — style tests (create new)
- `cmd/engram/status.go` — add styling
- `cmd/engram/search.go` — add styling
- `cmd/engram/recall.go` — add styling
- `cmd/engram/index.go` — add styling

## Notes
- Lipgloss automatically detects terminal capabilities. In CI/piped output, it strips ANSI codes. No need for manual `--no-color` flag.
- Keep the styling subtle — professional, not flashy. Use color for structure (labels, file paths, symbols) not decoration.
- Do NOT change the information content of any command output. Only change how it's displayed.
- The search and recall commands write to stdout (for piping). Use lipgloss on stdout — it auto-detects whether stdout is a terminal.
- If lipgloss is not in go.mod, add it. If it IS already there as a transitive dependency, make sure to import it properly.
