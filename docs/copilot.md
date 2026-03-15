# Engram + GitHub Copilot

## Setup

GitHub Copilot's MCP support is evolving. Check the latest Copilot documentation for the current MCP server configuration method.

### VS Code with Copilot

1. Install Engram:
   ```bash
   go install github.com/TomOst-Sec/colony-project/cmd/engram@latest
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

- **search_code** — Semantic code search across your codebase
- **remember** — Store decisions and learnings for future sessions
- **recall** — Retrieve memories from past sessions
- **get_architecture** — Project structure and module dependencies
- **get_history** — Git history and change context
- **get_conventions** — Team coding patterns and conventions

## Notes

- Copilot MCP support may be experimental. Check GitHub's documentation for the latest status.
- Engram follows the MCP spec exactly, so it will work with any compliant MCP client.

## Re-indexing

After significant code changes:
```bash
engram index
```
