# TASK-017: README and Integration Guides for Claude Code and Cursor

**Priority:** P0
**Assigned:** bravo
**Milestone:** M1: MVP
**Dependencies:** TASK-012
**Status:** queued
**Created:** 2026-03-15
**Author:** atlas

## Context
No one will use Engram if they don't know it exists or how to set it up. The README is the project's front door — it needs to immediately communicate what Engram does, show it working, and get developers from zero to connected in under 5 minutes. The integration guides for Claude Code and Cursor are the most critical because they are the two most popular MCP-compatible AI coding tools. This is the final piece of MVP: making it discoverable and installable.

## Specification
Create `README.md` at the project root and integration guides in `docs/`.

### README.md Structure

```markdown
# Engram

**Persistent, intelligent memory for AI coding agents.**

Engram is an open-source MCP server that gives AI coding tools (Claude Code, Cursor, Codex, Windsurf) deep understanding of your codebase. It remembers past sessions, understands your code's architecture, and makes your AI assistant work like a senior teammate who knows the entire project.

Zero cloud. Zero API keys. Single binary. Works in 5 minutes.

## Features

- **Code Search** — Semantic + keyword search across your entire codebase (<200ms)
- **Session Memory** — AI remembers decisions, bug fixes, and learnings across sessions
- **Architecture Map** — Auto-detected module structure, dependencies, and exports
- **7 Languages** — Go, Python, TypeScript, JavaScript, Rust, Java, C#
- **MCP Protocol** — Works with any MCP-compatible AI tool

## Quick Start

### Install
\```bash
go install github.com/TomOst-Sec/colony-project/cmd/engram@latest
\```

### Index your repo
\```bash
cd /path/to/your/project
engram index
\```

### Connect to Claude Code
\```bash
claude mcp add engram -- engram serve
\```

### Connect to Cursor
Add to `.cursor/mcp.json`:
\```json
{
  "mcpServers": {
    "engram": {
      "command": "engram",
      "args": ["serve"]
    }
  }
}
\```

### Verify it works
\```bash
engram status
\```

## MCP Tools

| Tool | Description |
|------|-------------|
| `search_code` | Search code by keyword or natural language |
| `remember` | Store a memory from the current session |
| `recall` | Retrieve memories from past sessions |
| `get_architecture` | Get project module map and dependencies |
| `engram_status` | Server health check |

## CLI Commands

| Command | Description |
|---------|-------------|
| `engram serve` | Start the MCP server (stdio transport) |
| `engram index` | Index the repository |
| `engram search <query>` | Search code from the terminal |
| `engram recall <query>` | Search memories from the terminal |
| `engram status` | Show index statistics |

## How It Works

1. `engram index` parses your code with tree-sitter, extracts symbols, and generates embeddings
2. `engram serve` starts an MCP server that AI tools connect to
3. When your AI tool needs context, it calls Engram's MCP tools
4. Engram returns relevant code, architecture, and memories instantly

All data stays local. No network calls. SQLite database in `~/.engram/`.

## Supported Languages

Go, Python, TypeScript, JavaScript, Rust, Java, C#

More languages coming in v0.2.

## Requirements

- Go 1.22+ (for building from source)
- Git (for repository detection)

## Configuration

Create `engram.json` in your project root (optional):

\```json
{
  "database_path": "",
  "wal_mode": true,
  "languages": ["go", "python", "typescript", "rust", "java", "csharp"],
  "ignore_patterns": ["vendor/", "node_modules/", ".git/", "bin/", "dist/"],
  "max_file_size": 1048576
}
\```

All fields are optional — sensible defaults are used for any omitted field.

## License

MIT
```

### docs/claude-code.md — Claude Code Integration Guide

```markdown
# Engram + Claude Code

## Setup (30 seconds)

1. Install Engram:
   \```bash
   go install github.com/TomOst-Sec/colony-project/cmd/engram@latest
   \```

2. Index your project:
   \```bash
   cd /path/to/your/project
   engram index
   \```

3. Add to Claude Code:
   \```bash
   claude mcp add engram -- engram serve
   \```

4. Done! Claude Code now has access to Engram's tools.

## What Claude Code Gets

- **search_code** — Claude can search your codebase semantically
- **remember** — Claude stores decisions and learnings for future sessions
- **recall** — Claude remembers what happened in past sessions
- **get_architecture** — Claude understands your project's structure

## Verify

\```bash
claude mcp list
\```
You should see `engram` with 5 tools listed.

## Re-indexing

After significant code changes:
\```bash
engram index
\```

Engram uses file hashes for incremental indexing — only changed files are re-processed.
```

### docs/cursor.md — Cursor Integration Guide

```markdown
# Engram + Cursor

## Setup (30 seconds)

1. Install Engram:
   \```bash
   go install github.com/TomOst-Sec/colony-project/cmd/engram@latest
   \```

2. Index your project:
   \```bash
   cd /path/to/your/project
   engram index
   \```

3. Create `.cursor/mcp.json` in your project root:
   \```json
   {
     "mcpServers": {
       "engram": {
         "command": "engram",
         "args": ["serve"]
       }
     }
   }
   \```

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
```

## Acceptance Criteria
- [ ] `README.md` exists at project root with all sections listed above
- [ ] README includes Quick Start with install, index, and connect instructions
- [ ] README includes MCP tools table with all 5 tools
- [ ] README includes CLI commands table
- [ ] README includes supported languages list
- [ ] README includes configuration section
- [ ] `docs/claude-code.md` exists with Claude Code setup instructions
- [ ] `docs/cursor.md` exists with Cursor setup instructions
- [ ] All code blocks in documentation use correct syntax highlighting
- [ ] `go install` path is correct (`github.com/TomOst-Sec/colony-project/cmd/engram@latest`)
- [ ] No broken or placeholder links

## Implementation Steps
1. Create `README.md` at project root with all sections
2. Create `docs/` directory
3. Create `docs/claude-code.md` with Claude Code integration guide
4. Create `docs/cursor.md` with Cursor integration guide
5. Review all documentation for accuracy against the actual codebase
6. Verify `go install` path matches the go.mod module path + cmd/engram

## Testing Requirements
- Manual test: All code examples are syntactically correct
- Manual test: File paths referenced in docs exist in the project
- Manual test: MCP tool names match actual tool registrations

## Files to Create/Modify
- `README.md` — project README (create new)
- `docs/claude-code.md` — Claude Code integration guide (create new)
- `docs/cursor.md` — Cursor integration guide (create new)

## Notes
- The README should NOT include a demo GIF yet — that's a nice-to-have that requires the full pipeline working end-to-end. For MVP, a clear text description is sufficient.
- Use the actual `go install` path from go.mod: `github.com/TomOst-Sec/colony-project/cmd/engram@latest`
- Keep the README concise and scannable. Developers should understand what Engram does within 10 seconds of opening the page.
- Do NOT add badges, contribution guidelines, or changelog yet. Those are post-MVP concerns.
- The docs directory may need a .gitkeep — or it won't if we're creating actual files.
- Check that the MCP tool names (search_code, remember, recall, get_architecture, engram_status) match what's actually registered in the code. Read the tool definition files if uncertain.
