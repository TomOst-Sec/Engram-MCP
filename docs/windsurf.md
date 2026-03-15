# Engram + Windsurf

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

- **search_code** — Semantic code search across your codebase
- **remember** — Store decisions and learnings for future sessions
- **recall** — Retrieve memories from past sessions
- **get_architecture** — Project structure and module dependencies
- **get_history** — Git history and change context
- **get_conventions** — Team coding patterns and conventions

## Re-indexing

After significant code changes, run `engram index` from the terminal. Windsurf does not need to be restarted — the MCP server re-reads from the updated database.
