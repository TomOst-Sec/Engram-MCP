# TASK-018: Git History Analyzer — Blame Context, Hotspots, Decision Trails

**Priority:** P1
**Assigned:** alpha
**Milestone:** M2: Core Features
**Dependencies:** TASK-003
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
Feature 10 from GOALS.md. The Git History Analyzer parses `git log` and `git blame` to extract decision context for every symbol in the index. This powers the `get_history` MCP tool (TASK-020) which lets AI agents answer "Why does this function exist? Who changed it last? What files always change together?" This is a backend/library task — it builds the `internal/git` package that reads git history and populates the `git_context` table in SQLite. The `git_context` table already exists in the schema (TASK-003).

## Specification
Create the `internal/git` package with a `HistoryAnalyzer` that shells out to `git log` and `git blame` and stores results in the `git_context` table.

### Package: `internal/git/analyzer.go`

```go
type HistoryAnalyzer struct {
    store    *storage.Store
    repoRoot string
}

type FileHistory struct {
    FilePath          string
    LastAuthor        string
    LastCommitHash    string
    LastCommitMessage string
    LastModified      time.Time
    ChangeFrequency   int      // total commits touching this file
    CoChangedFiles    []string // files that changed in the same commits (top 10)
}

type SymbolHistory struct {
    FilePath          string
    SymbolName        string
    LastAuthor        string
    LastCommitHash    string
    LastCommitMessage string
    LastModified      time.Time
    ChangeFrequency   int
}

func New(store *storage.Store, repoRoot string) *HistoryAnalyzer

// AnalyzeAll scans git history for all indexed files and populates git_context table.
func (h *HistoryAnalyzer) AnalyzeAll(ctx context.Context) (*AnalysisStats, error)

// AnalyzeFile analyzes git history for a single file.
func (h *HistoryAnalyzer) AnalyzeFile(ctx context.Context, filePath string) (*FileHistory, error)

// GetHotspots returns files sorted by change frequency (most-changed first).
func (h *HistoryAnalyzer) GetHotspots(ctx context.Context, limit int) ([]FileHistory, error)

// GetCoChangedFiles returns files that frequently change together with the given file.
func (h *HistoryAnalyzer) GetCoChangedFiles(ctx context.Context, filePath string, limit int) ([]string, error)
```

### AnalyzeAll Algorithm

```
1. Query code_index for distinct file_path values (only analyze indexed files)
2. For each file_path:
   a. Run: git log --format="%H|%an|%s|%aI" --follow -- <file_path>
      Parse output into commits (hash, author, message, date)
   b. Count commits → change_frequency
   c. Take most recent commit → last_author, last_commit_hash, last_commit_message, last_modified
   d. Run: git log --format="%H" -- <file_path>
      For each commit hash, get other files in that commit:
        git diff-tree --no-commit-id --name-only -r <hash>
      Track co-occurrence counts, return top 10 by frequency → co_changed_files
   e. Upsert into git_context table
3. Return stats
```

### Git Command Execution

Create a helper `internal/git/exec.go`:
```go
// RunGit executes a git command in the repo root and returns stdout.
func (h *HistoryAnalyzer) RunGit(ctx context.Context, args ...string) (string, error)
```
- Use `exec.CommandContext` for cancellation support
- Set `Dir` to `h.repoRoot`
- Capture stderr for error reporting
- Timeout: 30 seconds per command

### Storage Operations

Create `internal/git/store.go`:
```go
// UpsertFileHistory stores or updates git context for a file.
func UpsertFileHistory(store *storage.Store, fh *FileHistory) error

// GetFileHistory retrieves git context for a file.
func GetFileHistory(store *storage.Store, filePath string) (*FileHistory, error)

// GetHotspots retrieves files ordered by change_frequency DESC.
func GetHotspots(store *storage.Store, limit int) ([]FileHistory, error)
```

Use `INSERT OR REPLACE INTO git_context ...` for upserts. The `co_changed_files` field is stored as JSON array text.

### AnalysisStats
```go
type AnalysisStats struct {
    FilesAnalyzed    int
    FilesSkipped     int     // files not in git history
    TotalCommits     int     // unique commits scanned
    HottestFile      string  // file with most changes
    HottestFrequency int
    Duration         time.Duration
}
```

### Performance Considerations
- AnalyzeAll can be slow on large repos (many git commands). This is acceptable — it runs during `engram index`, not during MCP tool calls.
- For co-change analysis, limit to last 100 commits per file to bound execution time.
- Cache commit→files mappings to avoid re-running `git diff-tree` for the same commit across multiple files.

## Acceptance Criteria
- [ ] `HistoryAnalyzer.AnalyzeFile` returns correct last author, commit hash, message, and date for a file
- [ ] `HistoryAnalyzer.AnalyzeFile` returns correct change frequency (commit count)
- [ ] `HistoryAnalyzer.AnalyzeAll` processes all indexed files and populates git_context table
- [ ] `GetHotspots` returns files sorted by change frequency descending
- [ ] `GetCoChangedFiles` returns files that frequently change alongside the given file
- [ ] Co-changed files are stored as JSON array in git_context
- [ ] Git commands use context for cancellation
- [ ] Analyzer gracefully handles files not in git history (skips, doesn't error)
- [ ] Analyzer works correctly in this project's repo
- [ ] All tests pass

## Implementation Steps
1. Create `internal/git/exec.go` — RunGit helper with context, timeout, stderr capture
2. Create `internal/git/store.go` — UpsertFileHistory, GetFileHistory, GetHotspots storage functions
3. Create `internal/git/analyzer.go` — HistoryAnalyzer struct, New, AnalyzeFile, AnalyzeAll, GetHotspots, GetCoChangedFiles
4. Create `internal/git/exec_test.go` — test RunGit with basic git commands
5. Create `internal/git/store_test.go` — test upsert and query functions with real SQLite
6. Create `internal/git/analyzer_test.go`:
   - Test: AnalyzeFile on a file in this repo returns non-empty history
   - Test: AnalyzeFile returns correct commit count (>0)
   - Test: AnalyzeAll populates git_context for multiple files
   - Test: GetHotspots returns files in descending frequency order
   - Test: GetCoChangedFiles returns related files
   - Test: AnalyzeFile on non-existent file returns empty/zero result (no error)
7. Run all tests

## Testing Requirements
- Unit test: RunGit executes `git --version` successfully
- Unit test: UpsertFileHistory stores and retrieves FileHistory correctly
- Unit test: GetHotspots returns correct ordering
- Integration test: AnalyzeFile on a real git repo file returns valid history
- Integration test: AnalyzeAll on a temp repo with known commits produces expected results
- Unit test: Co-changed files JSON serialization round-trips correctly
- Unit test: Non-git file is skipped gracefully

## Files to Create/Modify
- `internal/git/exec.go` — git command execution helper
- `internal/git/store.go` — git_context table storage operations
- `internal/git/analyzer.go` — HistoryAnalyzer: AnalyzeFile, AnalyzeAll, GetHotspots, GetCoChangedFiles
- `internal/git/exec_test.go` — exec tests
- `internal/git/store_test.go` — storage tests
- `internal/git/analyzer_test.go` — analyzer integration tests

## Notes
- The `git_context` table already exists in the schema from TASK-003. Columns: id, file_path, symbol_name, last_author, last_commit_hash, last_commit_message, last_modified, change_frequency, co_changed_files, created_at, updated_at. The `symbol_name` column can be empty for file-level history (populate it later when we have symbol-level blame).
- Use `git log --format` with a delimiter (pipe `|`) that's unlikely to appear in commit messages. If a commit message contains `|`, truncate at the first occurrence.
- For co-change analysis, the commit cache is important. Build a `map[string][]string` (commitHash → files) to avoid redundant `git diff-tree` calls.
- Do NOT add this to the serve command or create an MCP tool in this task. That's TASK-020.
- Do NOT modify any files outside `internal/git/`. This is a pure library package.

---
## Completion Notes
- **Completed by:** alpha-3
- **Date:** 2026-03-15 16:27:55
- **Branch:** task/018
