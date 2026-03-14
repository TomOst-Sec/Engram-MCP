# THE COLONY — Project Memory

## Architecture
This project uses a 4-machine autonomous development system.
- GENESIS (Machine 1): Creates stories, writes code, generates prompts
- SENTINEL (Machine 2): Validates logic, checks architecture coherence
- FORGE (Machine 3): Runs tests, debugs, writes missing test coverage
- ORACLE (Machine 4): Reviews PRs, merges, deploys, monitors

## Coordination Protocol
- All work happens on feature branches via git worktrees
- Task queue lives in `_colony/tasks/` as YAML files
- Each machine polls for tasks matching its role
- Completed work is pushed as PRs for ORACLE to review
- ORACLE merges and feeds new tasks back to GENESIS

## Standards
- All code must have tests before merge (FORGE enforces)
- All architecture changes require SENTINEL approval
- All PRs require ORACLE review
- GENESIS never merges its own code

## Infrastructure
- InsForge: Task queue + BaaS (runs on ORACLE machine, port 3000)
- OpenViking: Shared context database (runs on ORACLE machine, port 8100)
- DeerFlow: Orchestration dashboard (runs on ORACLE machine, port 3001)

## Agent Personas
Sourced from ~/colony/agency-agents/ for engineering, testing, and security roles.

## Emergency Controls
- `touch _colony/PAUSE` — halts all agents
- `rm _colony/PAUSE` — resumes all agents
- `tmux kill-server` — kills all agents on a machine
