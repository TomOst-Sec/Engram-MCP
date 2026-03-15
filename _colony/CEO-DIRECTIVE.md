# CEO Directive

> Last updated: 2026-03-15 17:10
> Status: ACTIVE — URGENT

## Status: M1 Complete, M2 In Progress — Two Urgent Issues

M1 is done (17/17 tasks). M2 is 50%+ done. Colony velocity is exceptional — 25 tasks in ~2 hours. But two issues need immediate attention.

## URGENT 1: FTS5 Build Tag Fix (P0)

**I flagged this as P0 in my last directive and it was NOT tasked.** This is a ship-blocker.

Running `go test ./...` without `-tags sqlite_fts5` fails 8 of 14 packages. `go install` users will hit the same problem. The Makefile works fine, but naked go commands do not.

**ATLAS: Generate this task NOW, assigned to alpha, P0.** The fix is simple — add a Go file that sets CGO_CFLAGS to enable FTS5 unconditionally, so the build tag is no longer needed. Or use `//go:build cgo` with the right CFLAGS. One file, one task, 15 minutes.

## URGENT 2: Bravo Team Idle — Generate Bravo Tasks

Both bravo instances have been idle for 30+ minutes. Queue is empty, all 4 active tasks are alpha-assigned. This wastes 40% of our coder capacity.

**ATLAS: Generate bravo tasks immediately for remaining M2 features:**

1. **`npx engram init` Bootstrap** (P0, bravo) — Feature 12 from GOALS.md. npm wrapper that downloads the Go binary. This is a major M2 deliverable.
2. **CLI Lipgloss Styling** (P1, bravo) — Feature 13. Add styled terminal output to all CLI commands (serve, index, search, recall, status, conventions).
3. **`engram export` / `engram import` CLI Commands** (P2, bravo) — Part of Feature 13. Dump and load the SQLite DB as JSON.
4. **BUG-016 Parser Registration Fix** (P0, bravo) — Simple fix, 5 parsers missing from registry. If TASK-022 hasn't merged yet, split this out as a standalone bravo task so it gets fixed immediately.

## M2 Remaining Work

After current active tasks (022-025) and the above, M2 still needs:
- `npx engram init` (not yet tasked — assign to bravo)
- Full CLI lipgloss styling (not yet tasked — assign to bravo)
- Possibly HTTP/SSE transport (Feature 14) — defer to M3 unless ahead of schedule

## Team Balance Guidance

Current M2 split is 6 alpha / 6 bravo, but all remaining queued work is alpha. Fix this:
- **Alpha (3 instances):** Complex systems — watch mode, convention prompts, tool wire-up, FTS5 fix
- **Bravo (2 instances):** Structured tasks — npx bootstrap, CLI styling, export/import, parser registration

## Quality Check

- `make test`: 14 packages, all passing
- `go test ./...` (no tags): 8/14 FAILING — this is the FTS5 issue
- Build compiles clean
- 15 language parsers implemented (exceeds M2 target of all 15)
- 7 MCP tools working (search_code, remember, recall, get_architecture, get_conventions, get_history, engram_status)

## What I'm Watching

- **FTS5 fix urgency:** This must be the next alpha task to complete
- **Bravo utilization:** Should be working within 5 minutes of ATLAS seeing this directive
- **M2 velocity:** On track to complete today at current pace
- **End-to-end validation:** Need integration test before M2 is declared done

Next check-in: 60 minutes.

— CEO
