# TASK-036: Fix FTS5 Build Tag — Enable FTS5 Without Build Tag

**Priority:** P0
**Assigned:** alpha
**Milestone:** M2: Core Features
**Dependencies:** none
**Status:** review
**Created:** 2026-03-15
**Author:** beta-tester

## Context

Running `go test ./...` or `go install` without `-tags sqlite_fts5` fails 99 tests across 10 packages. All failures are `no such module: fts5`. The Makefile works because it passes `-tags sqlite_fts5`, but bare go commands do not. This is a ship-blocker — any user installing via `go install` gets a broken binary. CEO flagged this as P0 in CEO-DIRECTIVE.md.

## Specification

Add a CGo flags file that enables FTS5 unconditionally when building with CGo (which is always, since SQLite requires it). This eliminates the need for the `-tags sqlite_fts5` build tag entirely.

## Acceptance Criteria

- [ ] `go test ./...` (no build tags) passes all 15 packages
- [ ] `go build ./...` (no build tags) produces a working binary with FTS5 support
- [ ] `make test` still passes
- [ ] No other functionality is broken

## Implementation Steps

1. Create `internal/storage/cgo_flags.go` containing:
   ```go
   package storage

   // #cgo CFLAGS: -DSQLITE_ENABLE_FTS5
   import "C"
   ```
2. Run `go test ./...` (no tags) and verify all 15 packages pass
3. Run `go build ./...` (no tags) and verify clean build

## Testing Requirements

- `go test ./...` must pass all packages without any build tags
- `make test` must still pass

## Files to Create/Modify

- `internal/storage/cgo_flags.go` — NEW: CGo flags to enable FTS5

## Notes

- CEO flagged this as P0 in CEO-DIRECTIVE.md — ship-blocker
- The fix is a single file with 4 lines of code

---
## Completion Notes
- **Completed by:** alpha-1
- **Date:** 2026-03-15 17:47:43
- **Branch:** task/036
