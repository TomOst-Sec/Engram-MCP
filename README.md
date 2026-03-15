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

```bash
go install github.com/TomOst-Sec/colony-project/cmd/engram@latest
```

### Index your repo

```bash
cd /path/to/your/project
engram index
```

### Connect to Claude Code

```bash
claude mcp add engram -- engram serve
```

### Connect to Cursor

Add to `.cursor/mcp.json`:

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

### Verify it works

```bash
engram status
```

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

```json
{
  "database_path": "",
  "wal_mode": true,
  "languages": ["go", "python", "typescript", "rust", "java", "csharp"],
  "ignore_patterns": ["vendor/", "node_modules/", ".git/", "bin/", "dist/"],
  "max_file_size": 1048576
}
```

All fields are optional — sensible defaults are used for any omitted field.

## License

MIT
