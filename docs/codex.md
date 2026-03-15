# Engram + Codex CLI

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

- **search_code** — Semantic code search across your codebase
- **remember** — Store decisions and learnings for future sessions
- **recall** — Retrieve memories from past sessions
- **get_architecture** — Project structure and module dependencies
- **get_history** — Git history and change context
- **get_conventions** — Team coding patterns and conventions

## Re-indexing

After significant code changes:
```bash
engram index
```

## Notes

- Codex CLI uses stdio transport — same as Claude Code
- Check Codex CLI docs for the latest MCP configuration format
