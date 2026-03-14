#!/bin/bash
# Get the next available task number (zero-padded to 3 digits)
# Usage: ./next-task-number.sh
# Output: 001, 002, etc.

set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

# Find the highest TASK-NNN number across all directories
MAX=0
for dir in _colony/queue _colony/active _colony/review _colony/done; do
  if [ -d "$dir" ]; then
    for f in "$dir"/TASK-*.md; do
      [ -f "$f" ] || continue
      NUM=$(basename "$f" | grep -oP '\d+' | head -1)
      NUM=$((10#$NUM))  # Remove leading zeros for arithmetic
      if [ "$NUM" -gt "$MAX" ]; then
        MAX=$NUM
      fi
    done
  fi
done

NEXT=$((MAX + 1))
printf "%03d\n" "$NEXT"
