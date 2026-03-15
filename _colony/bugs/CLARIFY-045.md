# CLARIFY-045: cgo_flags.go approach doesn't fix FTS5

**Task:** TASK-045 — FTS5 Fix
**Filed by:** alpha-2
**Date:** 2026-03-15

## Issue

The specified `internal/storage/cgo_flags.go` with `#cgo CFLAGS: -DSQLITE_ENABLE_FTS5` does NOT fix the FTS5 issue. CGo CFLAGS only apply to C code compiled within the package they're defined in. Since `internal/storage` doesn't compile any C code (sqlite3 is compiled by `github.com/mattn/go-sqlite3`), the CFLAGS have no effect.

## Evidence

- `go test ./...` (with cgo_flags.go) → still fails: "no such module: fts5"
- `go test -tags sqlite_fts5 ./...` → ALL 18 packages pass (0 failures)

## Root Cause

`go-sqlite3` enables FTS5 via `sqlite3_opt_fts5.go` which is gated behind `//go:build sqlite_fts5`. The `-DSQLITE_ENABLE_FTS5` flag must be set when compiling the sqlite3 C code, which is in the go-sqlite3 package, not in internal/storage.

## Correct Fix Options

1. **GOFLAGS env var**: Add `GOFLAGS=-tags=sqlite_fts5` to .envrc (simplest)
2. **Makefile**: Add `-tags sqlite_fts5` to test/build targets
3. **go.env file**: Create `go.env` with `GOFLAGS=-tags=sqlite_fts5` (Go 1.21+)

Options 1 or 3 require no build tag in commands. Option 3 is checked into repo and works for everyone automatically.
