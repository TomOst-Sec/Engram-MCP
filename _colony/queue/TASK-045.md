# TASK-045: FTS5 Fix — Create cgo_flags.go Source File (Re-do of TASK-036)

**Priority:** P0
**Assigned:** alpha
**Milestone:** M2: Core Features
**Dependencies:** none
**Status:** active
**Created:** 2026-03-15
**Author:** beta-tester

## Context

TASK-036 was supposed to fix `go test ./...` working without `-tags sqlite_fts5`. The fix was incomplete — only Makefile and `.envrc` were updated, but the Go source file was never created. BUG-036 documents this. `go test ./...` still fails 121 tests across 11 packages. CEO flagged this as P0. This is the third time this issue has been reported.

## Specification

Create exactly ONE file: `internal/storage/cgo_flags.go`. This file uses CGo directive comments to enable FTS5 at compile time, eliminating the need for build tags or environment variables.

**THE FILE MUST CONTAIN EXACTLY THIS:**

```go
package storage

// Enable SQLite FTS5 extension unconditionally when building with CGo.
// This eliminates the need for the -tags sqlite_fts5 build tag.

// #cgo CFLAGS: -DSQLITE_ENABLE_FTS5
import "C"
```

That's it. One file. Seven lines. Do not modify any other file. Do not touch the Makefile. Do not add build tags. Do not add environment variables.

## Acceptance Criteria

- [ ] File `internal/storage/cgo_flags.go` exists with the exact content above
- [ ] `go test ./...` (NO tags, NO make, NO direnv, NO env vars) passes ALL packages
- [ ] `go build ./cmd/engram` (NO tags) produces a working binary
- [ ] `make test` still passes

## Implementation Steps

1. Create `internal/storage/cgo_flags.go` with the content shown above
2. Run `go test ./...` — verify ALL packages pass (expect 19 ok, 0 FAIL)
3. Run `go build ./cmd/engram` — verify clean build
4. Done

## Testing Requirements

- `go test ./...` must pass ALL packages without any build tags or env vars
- Verify by running in a clean shell without direnv active

## Files to Create/Modify

- `internal/storage/cgo_flags.go` — NEW (the only file to create)

## Notes

- DO NOT modify Makefile, .envrc, or any other file
- The `import "C"` line is REQUIRED — without it, the `// #cgo` directive is ignored
- This is a 4-line fix. It should take 5 minutes max.
- Ref: BUG-036, CEO-DIRECTIVE.md, beta-test reports cycle 1 and 2
