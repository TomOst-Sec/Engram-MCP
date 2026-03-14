#!/bin/bash
# Atomically claim a task — move from queue/ to active/ and commit
# Usage: ./claim-task.sh TASK-NNN.md <role>
# Exit 0 = claimed, Exit 1 = already taken or error

set -euo pipefail

TASK="${1:?Usage: ./claim-task.sh TASK-NNN.md <role>}"
ROLE="${2:?Usage: ./claim-task.sh TASK-NNN.md <role>}"

cd "$(git rev-parse --show-toplevel)"

# Pull latest to minimize race conditions
git pull origin main --rebase 2>/dev/null || true

# Check the task still exists in queue
if [ ! -f "_colony/queue/$TASK" ]; then
  echo "ERROR: $TASK not found in queue/ — already claimed or does not exist"
  exit 1
fi

# Verify this task is assigned to our role
ASSIGNED=$(grep -i "^\\*\\*Assigned:\\*\\*" "_colony/queue/$TASK" | head -1 | tr -d '* ' | cut -d: -f2 | xargs)
if [ -n "$ASSIGNED" ] && [ "$ASSIGNED" != "$ROLE" ]; then
  echo "ERROR: $TASK is assigned to $ASSIGNED, not $ROLE"
  exit 1
fi

# Move to active
mv "_colony/queue/$TASK" "_colony/active/$TASK"

# Update status in task file
sed -i "s/^\\*\\*Status:\\*\\*.*/\\*\\*Status:\\*\\* active/" "_colony/active/$TASK"

# Commit and push atomically
git add "_colony/queue/$TASK" "_colony/active/$TASK"
git commit -m "$ROLE: claimed $TASK"

if ! git push origin main; then
  # Push failed — someone else may have claimed it. Revert.
  git reset --soft HEAD~1
  git checkout -- _colony/
  echo "ERROR: Push failed — task may have been claimed by another machine"
  exit 1
fi

echo "OK: $ROLE claimed $TASK"
