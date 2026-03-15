# Engram + Codex CLI

## Setup (30 seconds)

1. Install Engram:
   ```bash
   git clone https://github.com/TomOst-Sec/colony-project.git && cd colony-project && make build
   # Copy bin/engram to your PATH
   ```

2. Index your project:
   ```bash
   cd /path/to/your/project
   engram index
   ```

3. Add to Codex CLI MCP configuration. Create or edit `~/.codex/mcp.json`:
   ```json
   {
     "mcpServers": {
       "engram": {
         "command": "engram",
         "args": ["serve"]
       }
     }
   }
   ```

   Or use the CLI flag:
   ```bash
   codex --mcp-server "engram serve"
   ```

4. Done! Codex now has access to Engram's tools.

## What Codex Gets

- **search_code** — Semantic + keyword code search
- **remember** — Store decisions and learnings for future sessions
- **recall** — Retrieve memories from past sessions
- **get_architecture** — Project module map and dependencies
- **get_history** — Git blame context, hotspots, co-change patterns
- **get_conventions** — Team coding patterns and conventions
- **get_conventions_prompt** — Auto-injected convention context
- **engram_status** — Server health and version info

## Watch Mode

Keep the index fresh automatically while you work:
```bash
engram index --watch
```

This uses fsnotify to re-index changed files in <500ms -- no manual re-indexing needed.

## Re-indexing

For a manual full re-index:
```bash
engram index
```

Use `--force` to rebuild everything.

## Notes

- Codex CLI uses stdio transport -- same as Claude Code
- Check Codex CLI docs for the latest MCP configuration format
