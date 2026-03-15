# TASK-001: Project Foundation — Go Module, Directory Structure, Makefile

**Priority:** P0
**Assigned:** alpha
**Milestone:** M1: MVP
**Dependencies:** none
**Status:** done
**Created:** 2026-03-15
**Author:** atlas

## Context
Engram is a greenfield Go project. Nothing exists yet — no go.mod, no source files, no build system. Every other task in the colony depends on this foundation being in place. This task creates the complete project skeleton so all 5 coders can work in parallel after it merges.

## Specification
Create the Go module, directory structure, entry point, and build system for the Engram MCP server.

**Go module:** `github.com/anthropics/engram` (or use the actual repo URL from `git remote -v`)
**Go version:** 1.22+

**Directory structure to create:**
```
cmd/
  engram/
    main.go              — entry point, delegates to cobra (stub for now)
internal/
  config/                — configuration loading (empty .go placeholder)
  storage/               — SQLite storage layer (empty .go placeholder)
  mcp/                   — MCP server core (empty .go placeholder)
  parser/                — tree-sitter AST parsing (empty .go placeholder)
  embeddings/            — ONNX embedding pipeline (empty .go placeholder)
  tools/                 — MCP tool implementations (empty .go placeholder)
tests/
  integration/           — integration test directory
```

**main.go** must:
- Import and initialize a cobra root command
- Support `--version` flag that prints `engram v0.1.0-dev`
- Exit cleanly with code 0

**Makefile** targets:
- `build` — `go build -o bin/engram ./cmd/engram`
- `test` — `go test ./...`
- `lint` — `golangci-lint run` (if available, skip gracefully if not)
- `clean` — remove `bin/`
- `run` — `go run ./cmd/engram`

**Dependencies to add via go get:**
- `github.com/spf13/cobra` (CLI framework)
- `github.com/stretchr/testify` (test assertions)

## Acceptance Criteria
- [ ] `go build ./cmd/engram` succeeds with zero errors
- [ ] `./bin/engram --version` prints version string containing "engram" and "0.1.0"
- [ ] `go test ./...` passes (at least one test exists)
- [ ] `make build` produces `bin/engram` binary
- [ ] `make test` runs successfully
- [ ] All directories listed above exist
- [ ] Every package directory has at least a `.go` file with the correct package declaration (can be a `doc.go` with just the package statement)

## Implementation Steps
1. Run `go mod init` with the correct module path (check `git remote -v` for the URL)
2. Create directory structure: `cmd/engram/`, `internal/{config,storage,mcp,parser,embeddings,tools}/`, `tests/integration/`
3. Create `cmd/engram/main.go` with cobra root command and `--version` flag
4. Create placeholder `doc.go` files in each `internal/` subdirectory with correct package names
5. Run `go get github.com/spf13/cobra github.com/stretchr/testify`
6. Create `Makefile` with build, test, lint, clean, run targets
7. Write `cmd/engram/main_test.go` — test that the root command executes without error
8. Run `go mod tidy`
9. Run `make build && make test` to verify everything works

## Testing Requirements
- Unit test: `cmd/engram/main_test.go` — execute root command, assert no error, assert version output contains expected string
- Build test: `go build ./...` succeeds for all packages

## Files to Create/Modify
- `go.mod` — Go module definition
- `go.sum` — dependency checksums (auto-generated)
- `cmd/engram/main.go` — entry point with cobra root command
- `cmd/engram/main_test.go` — root command tests
- `internal/config/doc.go` — package config placeholder
- `internal/storage/doc.go` — package storage placeholder
- `internal/mcp/doc.go` — package mcp placeholder
- `internal/parser/doc.go` — package parser placeholder
- `internal/embeddings/doc.go` — package embeddings placeholder
- `internal/tools/doc.go` — package tools placeholder
- `Makefile` — build system
- `.gitignore` — update to include `bin/`, `*.exe`, `*.db`

## Notes
- Check the actual git remote URL for the module path. If it's `github.com/TomOst-Sec/colony-project`, use that as the module path.
- Do NOT add any application logic beyond the version flag. This is purely scaffolding.
- The placeholder doc.go files should just contain `// Package <name> provides ...` and `package <name>`.
- Keep the cobra setup minimal — just a root command with version. Other commands (serve, index, etc.) will be added by TASK-008.

---
## Completion Notes
- **Completed by:** alpha-1
- **Date:** 2026-03-15 14:52:44
- **Branch:** task/001
