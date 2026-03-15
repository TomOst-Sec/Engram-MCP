#!/bin/bash
# Atomically claim a task — move from queue/ to active/ and commit
# Usage: ./claim-task.sh TASK-NNN.md <instance-name>
#   instance-name: alpha-1, alpha-2, alpha-3, bravo-1, bravo-2
# Exit 0 = claimed, Exit 1 = already taken or error

set -euo pipefail

TASK="${1:?Usage: ./claim-task.sh TASK-NNN.md <instance-name>}"
INSTANCE="${2:?Usage: ./claim-task.sh TASK-NNN.md <instance-name>}"

cd "$(git rev-parse --show-toplevel)"

# Extract team from instance name (alpha-1 → alpha, bravo-2 → bravo)
TEAM=$(echo "$INSTANCE" | sed 's/-[0-9]*$//')

# Pull latest to minimize race conditions
git pull origin main --rebase 2>/dev/null || true

# Check the task still exists in queue
if [ ! -f "_colony/queue/$TASK" ]; then
  echo "ERROR: $TASK not found in queue/ — already claimed or does not exist"
  exit 1
fi

# Verify this task is assigned to our team
ASSIGNED=$(grep -i "^\*\*Assigned:\*\*" "_colony/queue/$TASK" | head -1 | tr -d '* ' | cut -d: -f2 | xargs)
if [ -n "$ASSIGNED" ] && [ "$ASSIGNED" != "$TEAM" ]; then
  echo "ERROR: $TASK is assigned to $ASSIGNED, not $TEAM (instance: $INSTANCE)"
  exit 1
fi

# Move to active
mv "_colony/queue/$TASK" "_colony/active/$TASK"

# Update status in task file
sed -i "s/^\*\*Status:\*\*.*/\*\*Status:\*\* active/" "_colony/active/$TASK"

# Commit and push atomically
git add "_colony/queue/$TASK" "_colony/active/$TASK"
git commit -m "$INSTANCE: claimed $TASK"

if ! git push origin main; then
  # Push failed — someone else may have claimed it. Revert.
  git reset --soft HEAD~1
  git checkout -- _colony/
  echo "ERROR: Push failed — task may have been claimed by another instance"
  exit 1
fi

echo "OK: $INSTANCE claimed $TASK"
