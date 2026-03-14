---
description: "Start a colony agent on this machine"
---

Start the colony agent loop for this machine's role. Usage:

```bash
# Set your role first (add to ~/.bashrc):
export COLONY_ROLE="genesis"  # or sentinel, forge, oracle

# Start agent in tmux:
~/colony/start-agent.sh $COLONY_ROLE

# Monitor:
tmux attach -t $COLONY_ROLE
```

For ORACLE (Machine 4), start infrastructure first:
```bash
~/colony/start-oracle.sh
```
