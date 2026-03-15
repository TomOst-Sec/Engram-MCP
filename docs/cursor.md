# Engram + Cursor

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

3. Create `.cursor/mcp.json` in your project root:
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

4. Restart Cursor.

5. Done! Cursor's AI now has access to Engram's tools.

## Verify

Open Cursor's MCP panel (Settings > MCP) and check that `engram` appears with 5 tools.

## What Cursor Gets

- **search_code** — AI searches your codebase with semantic understanding
- **remember** — AI stores context from the current session
- **recall** — AI retrieves memories from past sessions
- **get_architecture** — AI understands module structure and dependencies

## Re-indexing

After significant code changes, run `engram index` from the terminal. Cursor does not need to be restarted — the MCP server re-reads from the updated database.
