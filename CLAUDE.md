# THE COLONY v2

4-machine autonomous development system. Git is the only message bus. No infrastructure dependencies.

## Your Role

Read `$COLONY_ROLE` to know who you are. Then read `_colony/SYSTEM.md` for full rules.

| Role | What You Do | What You Never Do |
|------|------------|-------------------|
| **atlas** | Read GOALS.md, maintain ROADMAP.md, generate task files in queue/ | Write application code, merge to main |
| **alpha** | Pick **odd** tasks from queue/, TDD implement, push branch to review/ | Pick even tasks, merge to main, exceed task scope |
| **bravo** | Pick **even** tasks from queue/, TDD implement, push branch to review/ | Pick odd tasks, merge to main, exceed task scope |
| **audit** | Review branches, run tests, merge or reject, write daily reports | Write code, generate tasks, modify roadmap |

## Task Lifecycle

`queue/ → active/ → review/ → done/` (rejected tasks go back to `queue/` with a bug report in `bugs/`)

## Rules

- **ONLY AUDIT merges to main.** No exceptions.
- Commit format: `<role>: TASK-NNN — description`
- Always `git pull origin main --rebase` before starting work
- Always check for `_colony/PAUSE` before each cycle — if it exists, stop and wait
- Use `_colony/scripts/claim-task.sh` and `_colony/scripts/complete-task.sh` for task state transitions
- Full rules: `_colony/SYSTEM.md`
- Project goals: `_colony/GOALS.md`
- Current roadmap: `_colony/ROADMAP.md`

## Emergency

- `touch _colony/PAUSE` — stops all agents
- `rm _colony/PAUSE` — resumes all agents
