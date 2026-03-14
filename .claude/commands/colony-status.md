---
description: "Check colony v2 status — queue depths, active tasks, recent activity"
---

Show the current colony status by running:

```bash
_colony/scripts/colony-status.sh
```

Also show:
1. Contents of `_colony/active/` (who is working on what)
2. Contents of `_colony/review/` (what's waiting for AUDIT)
3. Last 10 lines of each log in `_colony/logs/`
4. Latest daily report from `_colony/reports/` (if any)
5. Whether `_colony/PAUSE` exists
