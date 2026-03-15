---
description: "Coder skills for ALPHA and BRAVO teams — TDD, debugging, git worktrees, clean code"
---

# Coder Skills — ALPHA & BRAVO Teams

You are a coder. These are your core capabilities.

## 1. Test-Driven Development (TDD)

The iron rule: **Write the test FIRST. Watch it FAIL. Then write code.**

### The Cycle
```
RED    → Write a test that fails (proves the feature doesn't exist yet)
GREEN  → Write the MINIMUM code to make the test pass (ugly is fine)
REFACTOR → Clean up while keeping tests green (no new functionality)
```

### Discipline
- **Never skip RED.** If you write code before the test, you're doing it wrong.
- **Minimum code only.** Don't anticipate future needs. Solve THIS test.
- **Run tests after every change.** Not "I think it passes" — actually run them.
- **One behavior per test.** If a test name has "and" in it, split it.

### Common Rationalizations to Resist
- "This is too simple to test" → Test it anyway. Simple things break too.
- "I'll write tests after" → No. RED first. Always.
- "The test is obvious" → Then it's easy to write. Do it.
- "I need to see the implementation first" → That's the point of TDD. Trust the process.

### Verification Before Completion
Before marking a task as done:
1. Run the FULL test suite (not just your new tests)
2. Read the actual test output — don't assume it passed
3. Verify each acceptance criterion from the task has a corresponding test
4. Check: if you deleted your implementation, would ALL your tests fail? If not, they're testing the wrong thing.

## 2. Systematic Debugging

When tests fail unexpectedly, follow this process:

### Phase 1: Observe (don't guess)
- Read the FULL error message. What file? What line? What expected vs actual?
- Check: is this YOUR code failing, or a dependency?
- Check: does this test pass on main? (git stash, run test, git stash pop)

### Phase 2: Isolate
- Run ONLY the failing test. Does it fail in isolation?
- Add print/log statements at key points to trace data flow
- Reduce the test to the minimal reproduction

### Phase 3: Hypothesize and Test
- Form ONE hypothesis about the root cause
- Write a test that would confirm or deny your hypothesis
- If confirmed, fix it. If denied, go back to Phase 1.

### Phase 4: Fix and Verify
- Fix the root cause, not the symptom
- Run ALL tests (not just the one that was failing)
- If you've tried 3 fixes and none work, it's an architectural problem. Write a bug report and move on.

## 3. Git Worktree Usage

Each task gets its own worktree for isolation:

```bash
INSTANCE=$(echo $COLONY_ROLE)  # e.g., alpha-1
NNN=<task number>

# Create
git worktree add .worktrees/$INSTANCE-task-$NNN -b task/$NNN main

# Work in it
cd .worktrees/$INSTANCE-task-$NNN

# When done — push, then clean up
git push origin task/$NNN
cd ../..
git worktree remove .worktrees/$INSTANCE-task-$NNN
```

### Setup Checklist (after creating worktree)
- [ ] `go mod download` (or `npm install`, `pip install -r requirements.txt`)
- [ ] Run existing tests to establish baseline — they MUST pass before you start
- [ ] If baseline tests fail, do NOT proceed. Report as a bug.

## 4. Clean Code Principles

### Do
- Name functions by WHAT they do, not HOW (`getUserByEmail` not `queryDatabaseForUserRecord`)
- One function = one responsibility. If it's >40 lines, split it.
- Return early for errors — avoid deep nesting
- Use the language's idioms (Go: error returns, not exceptions. Python: list comprehensions.)

### Don't
- Don't add code "just in case." YAGNI.
- Don't refactor code you weren't asked to touch. Stay in scope.
- Don't leave TODOs — either do it or don't. Task scope is your boundary.
- Don't add comments that say WHAT the code does. Only comment WHY if it's non-obvious.
- Don't suppress errors or use empty catch blocks.

## 5. Task Execution Process

For every task:
1. **Read the FULL task file** — Context, spec, acceptance criteria, implementation steps, files to modify
2. **Check dependencies** — Are the tasks this depends on merged to main?
3. **Create worktree** — Isolated workspace
4. **Setup** — Install deps, verify baseline tests
5. **TDD cycle** — For each acceptance criterion: RED → GREEN → REFACTOR
6. **Full test suite** — Run ALL tests, not just yours
7. **Commit** — `$COLONY_ROLE: TASK-NNN — <description>`
8. **Push** — `git push origin task/NNN`
9. **Complete** — `_colony/scripts/complete-task.sh TASK-NNN.md $COLONY_ROLE`
10. **Cleanup** — Remove worktree

## 6. When to Give Up

If after a good-faith effort:
- Tests won't pass after 3 different approaches → Write `_colony/bugs/BUG-NNN.md`
- Task is unclear or contradictory → Write `_colony/bugs/CLARIFY-NNN.md`
- Dependency isn't ready → Skip, try the next task
- Git conflict you can't resolve → Write `_colony/bugs/CONFLICT-NNN.md`

Move the task back to queue, commit, push, and move on. Don't waste time.
