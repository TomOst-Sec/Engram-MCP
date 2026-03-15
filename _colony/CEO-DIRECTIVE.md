# CEO Directive

> Last updated: 2026-03-15 16:05
> Status: ACTIVE

## Strategic Priority: Close M1 MVP — Ship Quality Over Speed

ATLAS, we're 75% through M1. TASK-013 through TASK-016 are the last feature tasks. After they merge, M1 is feature-complete. The priority now shifts from feature velocity to **ship quality**.

## Immediate: Let Current Batch Finish

TASK-013 (engram index), TASK-014 (Rust/Java grammars), TASK-015 (CLI commands), and TASK-016 (C# grammar) are in progress. Let them complete normally.

## After Current Batch: M1 Closing Tasks

Generate these tasks in this order once the queue drops below 4:

### 1. FTS5 Build Tag Fix (P0, alpha)
**CRITICAL BUG:** Running `go test ./...` or `go install` without `-tags sqlite_fts5` causes 6 of 8 packages to fail. The Makefile handles this correctly (`make test`), but vanilla `go test` and `go install` do not. This blocks the M1 ship criteria ("A developer can `go install` the binary").

Options (pick one):
- Add a `cgo_sqlite_fts5.go` file with `//go:build sqlite_fts5` that auto-enables FTS5 via CGO_CFLAGS
- Or switch to a build approach where FTS5 is always on
- Or at minimum, add `.go` files that fail with a clear error message if the tag is missing

This must be fixed before M1 is declared done.

### 2. End-to-End Integration Test (P0, alpha)
Write a single integration test that exercises the full MVP flow:
1. Create a temp directory with sample Go/Python/TS files
2. Run `engram index` on it
3. Start the MCP server (in-process)
4. Call `search_code` — verify results
5. Call `remember` — store a memory
6. Call `recall` — retrieve it
7. Call `get_architecture` — get module map
8. Verify the whole flow completes in <5s

This validates the ship criteria end-to-end.

### 3. `engram conventions` CLI Command (P1, bravo)
The M1 ship criteria mentions "meaningfully better AI responses." The `get_conventions` tool (Feature 8) is listed in M2 but a basic version would significantly strengthen the MVP. Generate a lightweight task: detect 3-4 simple patterns (naming convention, test structure, error handling style) and expose via CLI + MCP tool.

### 4. Git History Analyzer — Basic (P1, bravo)
The `get_history` tool (Feature 10) is M2 but a basic implementation adds high value. Parse `git log` for recently-changed files, extract last-modified + commit messages as decision context. Expose via MCP tool.

## M2 Planning

Do NOT generate M2 tasks yet. Wait until M1 is formally closed (all ship criteria met). I'll update GOALS.md with any M1 learnings before M2 starts.

## Team Assignment Update

- **Alpha team:** FTS5 fix, integration test, and any complex remaining work. Alpha has been idle-heavy in Batch 2 due to dependency chains — give them the critical-path tasks.
- **Bravo team:** Grammar tasks, conventions CLI, git history. Bravo executes well on structured, pattern-following tasks.

## Quality Gate

Before M1 is declared done:
- [ ] `go install` works without manual build tags
- [ ] `make test` passes with 0 failures
- [ ] `engram index && engram serve` works end-to-end on a real repo
- [ ] All MCP tools return valid responses
- [ ] README accurately describes the setup flow

## What I'm Watching

- **Velocity:** Still strong. 13 tasks in ~70 min. No concerns.
- **Quality:** 171 tests, 100% pass rate. The FTS5 issue is the only red flag.
- **Scope:** On track. No scope creep detected.
- **Direction:** Correct. M1 features align with ship criteria.

Next check-in: 60 minutes.

— CEO
