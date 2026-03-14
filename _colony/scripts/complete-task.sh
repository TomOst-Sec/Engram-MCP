#!/bin/bash
# Complete a task — move from active/ to review/ and commit
# Usage: ./complete-task.sh TASK-NNN.md <role>

set -euo pipefail

TASK="${1:?Usage: ./complete-task.sh TASK-NNN.md <role>}"
ROLE="${2:?Usage: ./complete-task.sh TASK-NNN.md <role>}"

cd "$(git rev-parse --show-toplevel)"

# Pull latest
git pull origin main --rebase 2>/dev/null || true

# Verify task is in active
if [ ! -f "_colony/active/$TASK" ]; then
  echo "ERROR: $TASK not found in active/"
  exit 1
fi

# Move to review
mv "_colony/active/$TASK" "_colony/review/$TASK"

# Update status in task file
sed -i "s/^\\*\\*Status:\\*\\*.*/\\*\\*Status:\\*\\* review/" "_colony/review/$TASK"

# Append completion notes
TASK_NUM=$(echo "$TASK" | grep -oP '\d+')
cat >> "_colony/review/$TASK" << EOF

---
## Completion Notes
- **Completed by:** $ROLE
- **Date:** $(date '+%Y-%m-%d %H:%M:%S')
- **Branch:** task/$TASK_NUM
EOF

# Commit and push
git add "_colony/active/$TASK" "_colony/review/$TASK"
git commit -m "$ROLE: completed $TASK — moved to review"
git push origin main

echo "OK: $TASK moved to review/"
