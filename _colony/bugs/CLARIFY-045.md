# CLARIFY-045: cgo_flags.go approach does NOT enable FTS5

**Task:** TASK-045 — FTS5 Fix
**Author:** alpha-1
**Date:** 2026-03-15

## What's Wrong

The specified fix (`#cgo CFLAGS: -DSQLITE_ENABLE_FTS5` in `internal/storage/cgo_flags.go`) does NOT work.

Per Go documentation: "#cgo CFLAGS values defined within a package apply only when building that package, not for other packages." The FTS5 define must be set when compiling `sqlite3.c`, which happens inside the `go-sqlite3` package — NOT our `storage` package. Our `#cgo CFLAGS` only affects C files in OUR package (of which there are none).

## Proof

Tested 3 times across TASK-036 and TASK-045:
1. Created `internal/storage/cgo_flags.go` with the exact content specified
2. Ran `go clean -cache && go test ./internal/storage/` with NO env vars
3. Result: `no such module: fts5` — same failure

## What DOES Work

`CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" CGO_LDFLAGS="-lm" go test ./...` — passes all packages.

The `CGO_CFLAGS` environment variable applies to ALL packages in the build (per Go docs). This is why the `.envrc` approach from TASK-036 works.

## Options

1. **Accept .envrc approach** (TASK-036 already implemented this)
2. **Vendor go-sqlite3** and add the FTS5 define to the vendored copy
3. **Switch to modernc.org/sqlite** (pure Go, FTS5 always available)
4. **Always pass -tags sqlite_fts5** via GOFLAGS env var

There is no way to make `go test ./...` work without EITHER an env var OR a build tag. The `#cgo` directive approach is architecturally impossible per Go's CGo specification.
