# Engram + Claude Code

## Setup (30 seconds)

1. Install Engram:
   ```bash
   go install github.com/TomOst-Sec/colony-project/cmd/engram@latest
   ```

2. Index your project:
   ```bash
   cd /path/to/your/project
   engram index
   ```

3. Add to Claude Code:
   ```bash
   claude mcp add engram -- engram serve
   ```

4. Done! Claude Code now has access to Engram's tools.

## What Claude Code Gets

- **search_code** — Claude can search your codebase semantically
- **remember** — Claude stores decisions and learnings for future sessions
- **recall** — Claude remembers what happened in past sessions
- **get_architecture** — Claude understands your project's structure

## Verify

```bash
claude mcp list
```
You should see `engram` with 5 tools listed.

## Re-indexing

After significant code changes:
```bash
engram index
```

Engram uses file hashes for incremental indexing — only changed files are re-processed.
