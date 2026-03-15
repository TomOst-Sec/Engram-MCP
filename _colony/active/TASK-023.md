# TASK-023: Incremental Re-Indexing with `--watch` Mode

**Priority:** P1
**Assigned:** alpha
**Milestone:** M2: Core Features
**Dependencies:** TASK-013
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
Feature 11 from GOALS.md. Currently `engram index` does a full repo scan each time. The `--watch` mode uses fsnotify to monitor file changes and re-indexes only changed files in real-time (<500ms per file change). This keeps the index fresh without manual re-runs. It runs as a background goroutine alongside the MCP server, or as a standalone `engram index --watch` command.

## Specification

### Watcher Package: `internal/indexer/watcher.go`

```go
type Watcher struct {
    indexer  *Indexer
    repoRoot string
    config   *config.Config
    done     chan struct{}
}

func NewWatcher(indexer *Indexer, repoRoot string, cfg *config.Config) *Watcher

// Start begins watching the repository for file changes.
// Blocks until ctx is cancelled or Stop() is called.
func (w *Watcher) Start(ctx context.Context) error

// Stop signals the watcher to stop and waits for cleanup.
func (w *Watcher) Stop()
```

### Behavior
1. Use `github.com/fsnotify/fsnotify` to watch the repo root recursively
2. On file CREATE/WRITE events:
   a. Check if file has a registered parser (skip unknown extensions)
   b. Check ignore patterns (skip vendor/, node_modules/, .git/, etc.)
   c. Read file contents
   d. Call `indexer.IndexFile(ctx, filePath, source)` to re-index
   e. Log: `"Re-indexed %s (%d symbols, %dms)"`
3. On file REMOVE events:
   a. Delete symbols for that file from code_index: `DELETE FROM code_index WHERE file_path = ?`
4. On file RENAME events:
   a. Treat as REMOVE old + CREATE new
5. Debounce: if the same file triggers multiple events within 100ms, only process once

### Recursive Directory Watching
fsnotify doesn't watch subdirectories automatically. Walk the repo root at startup and add each directory (respecting ignore patterns). Watch for new directory creation and add those too.

### CLI Integration: `engram index --watch`
Add a `--watch` flag to the index command:
```go
indexCmd.Flags().Bool("watch", false, "Watch for file changes and re-index automatically")
```

When `--watch` is set:
1. Run full index first (existing behavior)
2. Create Watcher and start watching
3. Block until SIGINT/SIGTERM

### Dependencies
- `github.com/fsnotify/fsnotify` v1.7+ — add to go.mod

## Acceptance Criteria
- [ ] `engram index --watch` runs full index then watches for changes
- [ ] File saves trigger re-indexing of only the changed file
- [ ] File deletions remove symbols from the index
- [ ] Ignored directories (vendor/, .git/) are not watched
- [ ] New directories created after watch starts are also watched
- [ ] Re-indexing a single file completes in <500ms
- [ ] Watcher stops cleanly on SIGINT
- [ ] Debouncing prevents duplicate processing for rapid saves
- [ ] All tests pass

## Implementation Steps
1. `go get github.com/fsnotify/fsnotify`
2. Create `internal/indexer/watcher.go` — Watcher struct, NewWatcher, Start, Stop
3. Create `internal/indexer/watcher_test.go`:
   - Test: Watcher detects file creation and calls IndexFile
   - Test: Watcher detects file modification and re-indexes
   - Test: Watcher detects file deletion and removes symbols
   - Test: Watcher ignores files in .git/ directory
   - Test: Watcher ignores files with unregistered extensions
   - Test: Debouncing works (rapid saves → single index call)
4. Add `--watch` flag to `cmd/engram/index.go`
5. Update runIndex to start watcher when --watch is set
6. Run all tests

## Testing Requirements
- Unit test: NewWatcher creates without error
- Integration test: Write a file in temp dir → watcher calls IndexFile within 1s
- Integration test: Delete a file → symbols removed from code_index
- Unit test: Ignore patterns exclude .git/ and vendor/
- Unit test: Debounce timer coalesces rapid events
- Unit test: Stop() cleanly shuts down watcher

## Files to Create/Modify
- `internal/indexer/watcher.go` — file system watcher
- `internal/indexer/watcher_test.go` — watcher tests
- `cmd/engram/index.go` — add --watch flag and integration

## Notes
- The Indexer from TASK-013 has an `IndexFile(ctx, filePath, source)` method. Use it directly.
- For symbol deletion on file remove, you need to add a `DeleteFileSymbols(store, filePath)` function if it doesn't exist in the parser package. Check `internal/parser/store.go` — it likely has this already.
- fsnotify on Linux uses inotify which has a per-user watch limit (default ~8192). For large repos, this may need to be increased via `fs.inotify.max_user_watches`. Log a warning if adding a watch fails.
- The debounce implementation: use a `map[string]*time.Timer` where each file path has a pending timer. On event, reset the timer. When timer fires, do the indexing.
- Do NOT start the watcher from `engram serve`. That's a future enhancement. For now, only `engram index --watch` uses it.

---
## Completion Notes
- **Completed by:** alpha-3
- **Date:** 2026-03-15 17:51:13
- **Branch:** task/023
