---
description: "Hard rules enforced in every colony agent session"
---

# Colony Rules — All Agents

These rules are non-negotiable. Every agent must follow them in every cycle.

1. **Always pull before work:** `git pull origin main --rebase` at the start of every cycle.

2. **Check for PAUSE:** If `_colony/PAUSE` exists, stop immediately and wait. Check every 60 seconds.

3. **Stay in your lane:**
   - CEO: only writes to `_colony/GOALS.md`, `_colony/CEO-DIRECTIVE.md`, and `_colony/logs/ceo.log`. Can delete tasks from queue/.
   - ATLAS: only writes to `_colony/` (tasks, roadmap, goals processing). Reads CEO directives.
   - ALPHA/BRAVO instances: only write to `src/`, `tests/`, `docs/`, and move their own tasks in `_colony/`
   - AUDIT: only merges branches and writes to `_colony/` (done, bugs, reports)

4. **Use the scripts:** Always use `_colony/scripts/claim-task.sh` and `_colony/scripts/complete-task.sh` for task state transitions. Pass your instance name (e.g., `alpha-2`, `bravo-1`). Do not manually `mv` and `git add`.

5. **Run tests before pushing:** Every coder must run the full test suite before pushing a branch. No exceptions.

6. **Commit format:** `<instance>: TASK-NNN — <description>`. Use your full instance name (alpha-1, bravo-2, etc.), not just the team name.

7. **Only AUDIT merges to main.** If you are not AUDIT, never run `git merge` on main or `git push origin main` (except for task file state changes).

8. **Log everything:** Append to `_colony/logs/<instance>.log` after every action. Use your instance name.

9. **No scope creep:** Coders implement exactly what the task specifies. No "while I'm here" improvements.

10. **Clean branches:** One branch per task. Branch name: `task/NNN`. Delete after merge.

11. **Team boundaries:** Alpha instances only claim alpha-assigned tasks. Bravo instances only claim bravo-assigned tasks. If a claim fails (race lost), move to the next available task.

12. **Unique worktrees:** Name worktrees with your instance name: `.worktrees/<instance>-task-NNN` to avoid conflicts with other instances on the same machine.
