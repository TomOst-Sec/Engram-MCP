# Engram + Windsurf

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

3. Configure MCP in Windsurf:
   - Open Windsurf Settings
   - Navigate to the MCP Servers section
   - Add a new server:
     - Name: `engram`
     - Command: `engram`
     - Args: `serve`

   Or create `.windsurf/mcp.json` in your project root:
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

4. Restart Windsurf.

5. Done! Windsurf's AI now has access to Engram's tools.

## What Windsurf Gets

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

For a manual full re-index, run `engram index` from the terminal. Use `--force` to rebuild everything. Windsurf does not need to be restarted -- the MCP server re-reads from the updated database.
