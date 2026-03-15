---
description: "Beta testing skills — functional validation against project goals, regression detection, commit-level testing, failure report generation"
---

# Beta Testing & Goal Validation Skills

## 1. Commit Discovery — What Changed Since Last Run

### Tracking Last Run
- Read `_colony/logs/beta-tester.lastrun` for the commit SHA of the last tested state
- If no lastrun file exists, test the last 10 commits on the main branch
- After each cycle, write the current HEAD SHA to the lastrun file

### Gathering Changes
```bash
LAST_SHA=$(cat _colony/logs/beta-tester.lastrun 2>/dev/null || git log --oneline -10 --format='%H' | tail -1)
git log --oneline "$LAST_SHA"..HEAD
git diff --stat "$LAST_SHA"..HEAD
```

- List every file touched since last run
- Identify which modules, features, or subsystems were affected
- Group changes by area (e.g., MCP server, AST parsing, embeddings, storage, CLI)

## 2. Goal Alignment Validation

### Read the Source of Truth
Before testing, always read and internalize:
1. `_colony/GOALS.md` — the project objectives and success criteria
2. `_colony/ROADMAP.md` — what phase the project is in and current milestones
3. `_colony/CEO-DIRECTIVE.md` — any strategic overrides (if it exists)

### Alignment Check
For each new commit or group of commits, ask:
- **Does this change advance a stated goal?** Map the change to a specific goal or milestone.
- **Does this change contradict any goal?** Flag if a commit moves the project away from a stated objective.
- **Is this change in scope for the current phase?** Flag work that belongs to a future phase unless the CEO directive says otherwise.
- **Does the commit message reference a task?** Orphan commits (no TASK-NNN reference) may indicate out-of-process work.

## 3. Functional Testing (Go-specific)

### Test Execution Protocol
1. **Run the full test suite:** `go test ./... -v -count=1`
2. **Build check:** `go build ./...`
3. **Static analysis:** `go vet ./...`
4. **Read the output** — do not assume tests pass. Parse the actual output.

### Feature Verification
For each task referenced by new commits:
1. Find the task file in `_colony/done/` or `_colony/review/`
2. Read the acceptance criteria
3. Verify each criterion is actually met by the implementation
4. If a criterion cannot be verified automatically, note it as "MANUAL CHECK NEEDED"

### Regression Detection
- Run tests that existed BEFORE the new commits — do any old tests break?
- Check for removed tests — were tests deleted without replacement?
- Check for disabled tests — were tests commented out or skipped with `t.Skip()`?
- Scan for `TODO`, `FIXME`, `HACK`, `XXX` introduced in new commits

## 4. Failure Report Generation

When issues are found, write a report to `_colony/reports/beta-test-YYYY-MM-DD-HHMM.md`.

### Report Format
```markdown
# Beta Test Report — YYYY-MM-DD HH:MM

## Summary
- **Commits tested:** <first_sha>..<last_sha> (<N> commits)
- **Result:** PASS | FAIL | WARN
- **Tests run:** <count>
- **Tests passed:** <count>
- **Tests failed:** <count>
- **Goal alignment issues:** <count>

## Test Results
### Passing
- <summary of what works>

### Failing
For each failure:
- **Test:** <test name or file>
- **Error:** <actual error output>
- **Introduced by:** <commit SHA and message>
- **Related task:** TASK-NNN (if identifiable)

## Goal Alignment
### Aligned
- <commit> advances Goal X: <reason>

### Misaligned
- <commit> contradicts Goal Y: <reason>

### Untracked
- <commit> has no task reference — may be out-of-process work

## Regressions
- <description of any previously-passing test now failing>
- <description of removed or disabled tests>

## Recommendations
For each issue, write an actionable recommendation:
- **Priority:** P0 (blocks release) | P1 (should fix) | P2 (minor)
- **Action:** <specific fix instruction or task suggestion>
- **Assigned to:** <team name, if obvious from the affected area>
```

## 5. Auto-Filing Bug Tasks

When a P0 or P1 issue is found:
1. Get the next task number: `bash _colony/scripts/next-task-number.sh`
2. Check if a bug task already exists in `_colony/bugs/` for this issue
3. If not, write a new bug file and a corresponding queue task:

```markdown
# TASK-NNN: [BUG] <title>

**Status:** queued
**Priority:** P0
**Assigned:** alpha (or bravo, based on affected area)
**Found-By:** beta-tester
**Found-At:** <timestamp>
**Commit:** <SHA that introduced the issue>
**Related-Task:** TASK-NNN (if known)

## Description
<what is broken>

## Evidence
<test output, error messages>

## Expected Behavior
<what GOALS.md or the task spec says should happen>

## Actual Behavior
<what actually happens>

## Suggested Fix
<actionable steps to resolve>
```

## 6. Health Metrics

Track and report these in every cycle:
- **Build status** — does `go build ./...` succeed?
- **Test pass rate** — percentage of tests passing
- **Regression count** — tests that passed before but fail now
- **Goal coverage** — what percentage of stated goals have at least one implementation
- **Orphan commit rate** — commits without task references
- **Time since last green** — how long since all tests passed
