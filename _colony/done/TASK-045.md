# TASK-045: FTS5 Build Configuration — Proper Developer Setup

**Priority:** P0
**Assigned:** alpha
**Milestone:** M2: Core Features
**Dependencies:** none
**Status:** done
**Created:** 2026-03-15
**Author:** atlas (REWRITTEN — cgo_flags.go approach proven incorrect by BUG-045 + CLARIFY-045)

## Context

The `cgo_flags.go` approach was attempted by two coders (alpha-2, alpha-3) and independently confirmed to NOT WORK. CGo CFLAGS directives only affect C code compiled within that package. Since `mattn/go-sqlite3` compiles SQLite within its own package, our storage package's CGo flags have no effect.

**Root cause:** `go-sqlite3` gates FTS5 behind the `sqlite_fts5` build tag in `sqlite3_opt_fts5.go`. This is the library's design. There is no way to bypass it.

**For end users:** Pre-built binaries (npx, Homebrew, GoReleaser, Docker) are compiled WITH the tag. No issue.

**For developers:** Must use `make test` or `go test -tags sqlite_fts5 ./...`. This task ensures the developer experience is smooth.

## Specification

### 1. Remove broken cgo_flags.go if it exists
```bash
rm -f internal/storage/cgo_flags.go
```

### 2. Ensure `.envrc` sets GOFLAGS
Create or update `.envrc` at repo root:
```bash
# Loaded automatically by direnv (https://direnv.net/)
# Enables SQLite FTS5 for all go commands
export GOFLAGS="-tags=sqlite_fts5"
```

### 3. Ensure Makefile uses build tag consistently
All go commands should use a shared GOTAGS variable:
```makefile
GOTAGS := -tags sqlite_fts5
```
Applied to: build, test, install, bench, lint targets.

### 4. Update README.md
Add a "Building from Source" section:
```
## Building from Source

Engram requires the `sqlite_fts5` build tag for SQLite full-text search:

    make build     # Recommended
    make test      # Run all tests

Or manually:

    go build -tags sqlite_fts5 ./cmd/engram
    go test -tags sqlite_fts5 ./...

If you use direnv, the included .envrc sets this automatically.
```

### 5. Verify `.goreleaser.yml`
Confirm the builds section includes:
```yaml
flags:
  - -tags=sqlite_fts5
```

## Acceptance Criteria
- [ ] `internal/storage/cgo_flags.go` does NOT exist
- [ ] `.envrc` exists with `GOFLAGS="-tags=sqlite_fts5"`
- [ ] `Makefile` GOTAGS variable applies to all go commands
- [ ] `make test` passes ALL packages (0 failures)
- [ ] `make build` produces a working binary
- [ ] `go test -tags sqlite_fts5 ./...` passes ALL packages
- [ ] README has "Building from Source" section
- [ ] `.goreleaser.yml` includes sqlite_fts5 flag

## Implementation Steps
1. Delete `internal/storage/cgo_flags.go` if it exists
2. Create/verify `.envrc` with GOFLAGS
3. Verify/update Makefile GOTAGS variable
4. Add "Building from Source" to README.md
5. Verify `.goreleaser.yml` build flags
6. Run `make test` — ALL packages pass
7. Run `make build` — binary builds

## Files to Create/Modify
- `internal/storage/cgo_flags.go` — DELETE if exists
- `.envrc` — create/update
- `Makefile` — verify GOTAGS used consistently
- `README.md` — add build-from-source docs

## Notes
- The original `cgo_flags.go` approach was proven incorrect by two independent coders (BUG-045, CLARIFY-045). The CGo CFLAGS directive only affects C code compiled within the same package, not within dependency packages like go-sqlite3.
- `make test` is the canonical developer test command. This is standard Go practice — many projects with CGo deps use Makefile wrappers.
- `.envrc` with `GOFLAGS` gives direnv users transparent `go test ./...` support without flags.
- Delete BUG-045 and CLARIFY-045 after this task completes.

---
## Completion Notes
- **Completed by:** alpha-2
- **Date:** 2026-03-15 19:15:46
- **Branch:** task/045
- **Reviewed by:** audit
- **Merged:** 2026-03-15 20:12
- **Review notes:** All 8 acceptance criteria verified. make test passes 19/19 packages, make build produces working binary. Correct use of sqlite_fts5 build tag via GOTAGS variable. Replaces broken CGO_CFLAGS approach with proper build tag propagation.
