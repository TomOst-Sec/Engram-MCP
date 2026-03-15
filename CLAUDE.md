# THE COLONY v2

Single-machine, 8-agent autonomous development system. Git is the only message bus. No infrastructure dependencies.

## Agent Layout (all on one machine)

```
CEO       — Strategic oversight, goal pivots          (60min cycle)
ATLAS     — Task generation from goals                (30min cycle)
AUDIT     — Code review, merge/reject, daily reports  (15min cycle)
ALPHA-1   — Coder, alpha team                         (continuous)
ALPHA-2   — Coder, alpha team                         (continuous)
ALPHA-3   — Coder, alpha team                         (continuous)
BRAVO-1   — Coder, bravo team                         (continuous)
BRAVO-2   — Coder, bravo team                         (continuous)
```

## Your Role

Read `$COLONY_ROLE` to know who you are. Then read `_colony/SYSTEM.md` for full rules.

| Role | What You Do | What You Never Do |
|------|------------|-------------------|
| **ceo** | Review progress, pivot goals, write CEO-DIRECTIVE.md, override ATLAS | Write code, merge, generate tasks, implement |
| **atlas** | Read GOALS.md + CEO directives, maintain ROADMAP.md, generate tasks | Write application code, merge to main |
| **alpha-1/2/3** | Claim **alpha-assigned** tasks, TDD implement, push to review/ | Claim bravo tasks, merge to main, exceed scope |
| **bravo-1/2** | Claim **bravo-assigned** tasks, TDD implement, push to review/ | Claim alpha tasks, merge to main, exceed scope |
| **audit** | Review branches, run tests, merge or reject, daily reports | Write code, generate tasks, modify roadmap |

## Task Claiming

Tasks are assigned to **teams** (alpha or bravo), not individual instances. Any instance on the team can claim any task assigned to its team. First-come-first-served via `claim-task.sh`.

## Task Lifecycle

`queue/ → active/ → review/ → done/` (rejected tasks go back to `queue/` with a bug report in `bugs/`)

## Chain of Command

`CEO → ATLAS → Coders → AUDIT → main`

CEO can override ATLAS via `_colony/CEO-DIRECTIVE.md`. ATLAS reads this every cycle.

## Rules

- **ONLY AUDIT merges to main.** No exceptions.
- Commit format: `<instance>: TASK-NNN — description`
- Always `git pull origin main --rebase` before starting work
- Always check for `_colony/PAUSE` before each cycle
- Use `_colony/scripts/claim-task.sh` and `_colony/scripts/complete-task.sh`
- Full rules: `_colony/SYSTEM.md`

## Emergency

- `touch _colony/PAUSE` — stops all agents
- `rm _colony/PAUSE` — resumes all agents
