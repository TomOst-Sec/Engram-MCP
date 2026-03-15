# TASK-029: Integration Guides — Codex, Windsurf, Copilot

**Priority:** P2
**Assigned:** bravo
**Milestone:** M2: Core Features
**Dependencies:** none
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
TASK-017 created integration guides for Claude Code and Cursor. Milestone 2 targets three more AI coding tools: OpenAI Codex CLI, Windsurf (Codeium), and GitHub Copilot. These guides help developers connect Engram to their preferred AI tool. The guides follow the same pattern as the existing Claude Code and Cursor guides.

## Specification
Create integration guides in `docs/` and update the README.

### docs/codex.md — Codex CLI Integration Guide

```markdown
# Engram + Codex CLI

## Setup

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
- **search_code** — Semantic code search
- **remember** — Persist session context
- **recall** — Retrieve past session memories
- **get_architecture** — Project structure understanding
- **get_history** — Git history and change context
- **get_conventions** — Team coding patterns

## Notes
- Codex CLI uses stdio transport — same as Claude Code
- Check Codex CLI docs for the latest MCP configuration format
```

### docs/windsurf.md — Windsurf Integration Guide

```markdown
# Engram + Windsurf

## Setup

1. Install Engram (same as other tools)

2. Index your project

3. Configure MCP in Windsurf:
   - Open Windsurf Settings
   - Navigate to MCP Servers section
   - Add a new server:
     - Name: `engram`
     - Command: `engram`
     - Args: `serve`

   Or create `.windsurf/mcp.json`:
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

## What Windsurf Gets
(same tool list as other guides)
```

### docs/copilot.md — Copilot Integration Guide

```markdown
# Engram + GitHub Copilot

## Setup

GitHub Copilot's MCP support is evolving. Check the latest Copilot documentation for MCP server configuration.

### VS Code with Copilot

If Copilot supports MCP in VS Code, configure via `.vscode/mcp.json`:
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

### Copilot CLI

If Copilot CLI supports MCP:
```bash
gh copilot --mcp-server "engram serve"
```

## Notes
- Copilot MCP support may be experimental. Check GitHub's documentation.
- Engram follows the MCP spec exactly, so it will work with any compliant MCP client.
```

### README.md Updates
Add a "Supported AI Tools" section:
```markdown
## Supported AI Tools
- [Claude Code](docs/claude-code.md)
- [Cursor](docs/cursor.md)
- [Codex CLI](docs/codex.md)
- [Windsurf](docs/windsurf.md)
- [GitHub Copilot](docs/copilot.md)
- Any MCP-compatible tool
```

Update the MCP tools table to include get_history and get_conventions.

## Acceptance Criteria
- [ ] `docs/codex.md` exists with Codex CLI setup instructions
- [ ] `docs/windsurf.md` exists with Windsurf setup instructions
- [ ] `docs/copilot.md` exists with Copilot setup instructions
- [ ] README.md updated with supported AI tools section
- [ ] README.md MCP tools table includes get_history and get_conventions
- [ ] All guides include install, index, and configure steps
- [ ] All guides list available tools
- [ ] No broken links

## Implementation Steps
1. Create `docs/codex.md`
2. Create `docs/windsurf.md`
3. Create `docs/copilot.md`
4. Update `README.md` — add supported tools section, update tools table
5. Review all docs for accuracy

## Testing Requirements
- Manual: all file paths in docs exist
- Manual: tool names match actual registrations

## Files to Create/Modify
- `docs/codex.md` — Codex CLI integration guide (create new)
- `docs/windsurf.md` — Windsurf integration guide (create new)
- `docs/copilot.md` — Copilot integration guide (create new)
- `README.md` — update supported tools + tools table

## Notes
- Follow the exact same format as `docs/claude-code.md` and `docs/cursor.md`. Study those for the pattern.
- The Codex CLI MCP configuration path may differ — use `~/.codex/mcp.json` as a reasonable guess and note that users should check Codex docs.
- Copilot MCP support is still maturing. Be honest about this in the guide — don't promise features that don't exist yet.
- Update the README tools table to 7 tools (add get_history and get_conventions to the existing 5).
- Keep each guide concise — under 60 lines.

---
## Completion Notes
- **Completed by:** bravo-2
- **Date:** 2026-03-15 16:52:58
- **Branch:** task/029
