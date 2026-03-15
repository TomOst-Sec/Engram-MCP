# Engram + GitHub Copilot

## Setup

GitHub Copilot's MCP support is evolving. Check the latest Copilot documentation for the current MCP server configuration method.

### VS Code with Copilot

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

3. If Copilot supports MCP in VS Code, configure via `.vscode/mcp.json`:
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

4. Restart VS Code.

### Copilot CLI

If Copilot CLI supports MCP:
```bash
gh copilot --mcp-server "engram serve"
```

## What Copilot Gets

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

## Notes

- Copilot MCP support may be experimental. Check GitHub's documentation for the latest status.
- Engram follows the MCP spec exactly, so it will work with any compliant MCP client.

## Re-indexing

For a manual full re-index:
```bash
engram index
```

Use `--force` to rebuild everything.
