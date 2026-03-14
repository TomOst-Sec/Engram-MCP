#!/bin/bash
# Colony v2 status — print queue depths, active tasks, and recent activity
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

echo "╔══════════════════════════════════════════════╗"
echo "║  COLONY v2 STATUS — $(date '+%Y-%m-%d %H:%M:%S')      ║"
echo "╠══════════════════════════════════════════════╣"

# Pause check
if [ -f "_colony/PAUSE" ]; then
  echo "║  !! COLONY IS PAUSED !!                      ║"
  echo "╠══════════════════════════════════════════════╣"
fi

# Queue depths
QUEUED=$(find _colony/queue/ -name "TASK-*.md" 2>/dev/null | wc -l)
ACTIVE=$(find _colony/active/ -name "TASK-*.md" 2>/dev/null | wc -l)
REVIEW=$(find _colony/review/ -name "TASK-*.md" 2>/dev/null | wc -l)
DONE=$(find _colony/done/ -name "TASK-*.md" 2>/dev/null | wc -l)
BUGS=$(find _colony/bugs/ -name "*.md" 2>/dev/null | wc -l)

echo "║  TASK PIPELINE:                                "
echo "║    Queued:    $QUEUED"
echo "║    Active:    $ACTIVE"
echo "║    Review:    $REVIEW"
echo "║    Done:      $DONE"
echo "║    Bugs:      $BUGS"

# Active tasks detail
if [ "$ACTIVE" -gt 0 ]; then
  echo "║"
  echo "║  ACTIVE TASKS:"
  for f in _colony/active/TASK-*.md; do
    [ -f "$f" ] || continue
    NAME=$(basename "$f")
    TITLE=$(head -1 "$f" | sed 's/^# //')
    echo "║    $NAME — $TITLE"
  done
fi

# Review queue detail
if [ "$REVIEW" -gt 0 ]; then
  echo "║"
  echo "║  AWAITING REVIEW:"
  for f in _colony/review/TASK-*.md; do
    [ -f "$f" ] || continue
    NAME=$(basename "$f")
    TITLE=$(head -1 "$f" | sed 's/^# //')
    echo "║    $NAME — $TITLE"
  done
fi

# Agent status (tmux sessions)
echo "║"
echo "║  AGENTS:"
for role in atlas alpha bravo audit; do
  if tmux has-session -t "$role" 2>/dev/null; then
    echo "║    $role: ● RUNNING"
  else
    echo "║    $role: ○ STOPPED"
  fi
done

# Recent git activity
echo "║"
echo "║  RECENT COMMITS:"
git log --all --oneline -5 2>/dev/null | while read -r line; do
  echo "║    $line"
done

# Latest report
LATEST_REPORT=$(ls -t _colony/reports/*.md 2>/dev/null | head -1)
if [ -n "${LATEST_REPORT:-}" ]; then
  echo "║"
  echo "║  LATEST REPORT: $(basename "$LATEST_REPORT")"
fi

echo "╚══════════════════════════════════════════════╝"
