# Engram

**Persistent, intelligent memory for AI coding agents.**

Engram is an open-source MCP server that gives AI coding tools (Claude Code, Cursor, Codex, Windsurf, Copilot) deep understanding of your codebase. It remembers past sessions, understands your code's architecture, and makes your AI assistant work like a senior teammate who knows the entire project.

Zero cloud. Zero API keys. Single binary. Works in 5 minutes.

## Features

- **Code Search** — Hybrid FTS5 + vector similarity search across your codebase (<200ms)
- **Session Memory** — AI remembers decisions, bug fixes, and learnings across sessions
- **Architecture Map** — Auto-detected module structure, dependencies, and exports
- **Convention Detection** — Infers naming patterns, test structure, error handling style from your code
- **Git History** — Blame context, hotspot detection, co-change patterns, decision trails
- **15 Languages** — Go, Python, TypeScript, JavaScript, Rust, Java, C#, Ruby, PHP, Swift, Kotlin, C, C++, Lua, Zig
- **Watch Mode** — Incremental re-indexing on file save via fsnotify (<500ms updates)
- **HTTP/SSE Transport** — Remote/team MCP server with bearer token auth and CORS
- **TUI Dashboard** — Interactive terminal UI for browsing memories, conventions, and architecture
- **Ollama Integration** — Optional local LLM for embeddings
- **CI/CD Hook** — Parse GitHub Actions / GitLab CI output into searchable memories
- **Community Conventions** — Shareable convention packs (`engram conventions add <pack>`)
- **Multi-Repo** — Cross-repository search for monorepos and microservices
- **MCP Protocol** — Works with any MCP-compatible AI tool

## Quick Start

### Install

**From source (recommended):**

```bash
git clone https://github.com/TomOst-Sec/colony-project.git
cd colony-project
make build
# Binary at bin/engram — copy to your PATH
```

> **Note:** Engram requires the `sqlite_fts5` build tag for full-text search. Plain `go install` will not include it. Use `make build` or pass `-tags sqlite_fts5` manually.

**With Docker:**

```bash
docker build -t engram .
docker run -v $(pwd):/workspace engram index
docker run -v $(pwd):/workspace engram serve
```

See [docs/docker.md](docs/docker.md) for team setups and CI/CD usage.

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

## Supported AI Tools

- [Claude Code](docs/claude-code.md) — `claude mcp add engram -- engram serve`
- [Cursor](docs/cursor.md) — `.cursor/mcp.json`
- [Codex CLI](docs/codex.md) — `~/.codex/mcp.json`
- [Windsurf](docs/windsurf.md) — Settings > MCP Servers or `.windsurf/mcp.json`
- [GitHub Copilot](docs/copilot.md) — `.vscode/mcp.json` (experimental)
- Any MCP-compatible tool via stdio or HTTP/SSE transport

## MCP Tools

| Tool | Description |
|------|-------------|
| `search_code` | Search code by keyword or natural language |
| `remember` | Store a memory from the current session |
| `recall` | Retrieve memories from past sessions |
| `get_architecture` | Get project module map and dependencies |
| `get_history` | Git history and change context |
| `get_conventions` | Team coding patterns and conventions |
| `get_conventions_prompt` | Auto-injected convention context for AI tools |
| `engram_status` | Server health check |

## CLI Commands

| Command | Description |
|---------|-------------|
| `engram serve` | Start the MCP server (stdio or HTTP transport) |
| `engram index` | Index the repository (`--watch` for live re-indexing, `--force` for full rebuild) |
| `engram search <query>` | Search code from the terminal (`-l` language, `-t` symbol type, `-n` limit) |
| `engram recall <query>` | Search memories (`-t` type filter, `-n` limit, `--since` date) |
| `engram status` | Show index statistics |
| `engram init` | Initialize Engram and generate configuration |
| `engram tui` | Interactive terminal dashboard (status, memories, conventions, architecture) |
| `engram export [file]` | Export database to JSON (`--tables`, `--pretty`) |
| `engram import [file]` | Import database from JSON (`--replace` to clear first) |
| `engram ci-hook` | Parse CI/CD output into memories (`-s` github-actions/gitlab-ci/generic) |
| `engram conventions list\|add\|remove` | Manage community convention packs |

### Server flags

```
engram serve --transport http --http-addr :3333 --http-token SECRET --cors-origin "*"
```

## How It Works

1. `engram index` parses your code with tree-sitter, extracts symbols, and generates embeddings
2. `engram serve` starts an MCP server that AI tools connect to
3. When your AI tool needs context, it calls Engram's MCP tools
4. Engram returns relevant code, architecture, and memories instantly

All data stays local. No network calls. SQLite database in `~/.engram/`.

## Supported Languages

Go, Python, TypeScript, JavaScript, Rust, Java, C#, Ruby, PHP, Swift, Kotlin, C, C++, Lua, Zig

## Building from Source

Engram requires the `sqlite_fts5` build tag for SQLite full-text search:

```bash
make build     # Recommended
make test      # Run all tests
```

Or manually:

```bash
go build -tags sqlite_fts5 ./cmd/engram
go test -tags sqlite_fts5 ./...
```

If you use [direnv](https://direnv.net/), the included `.envrc` sets this automatically.

## Requirements

- Go 1.22+ with CGo enabled (for building from source)
- GCC or compatible C compiler (for SQLite CGo bindings)
- Git (for repository detection and history analysis)

## Configuration

Create `engram.json` in your project root (optional):

```json
{
  "database_path": "",
  "wal_mode": true,
  "transport": "stdio",
  "languages": ["go", "python", "typescript", "javascript", "rust", "java", "csharp"],
  "ignore_patterns": ["vendor/", "node_modules/", ".git/", "bin/", "dist/"],
  "max_file_size": 1048576,
  "ollama_endpoint": "http://localhost:11434",
  "ollama_model": "nomic-embed-text",
  "additional_repos": []
}
```

All fields are optional — sensible defaults are used for any omitted field.

### Configuration Precedence

Environment variables (`ENGRAM_*`) > `engram.json` > defaults.

| Environment Variable | Default |
|---------------------|---------|
| `ENGRAM_DATABASE_PATH` | `~/.engram/<repo-hash>/engram.db` |
| `ENGRAM_TRANSPORT` | `stdio` |
| `ENGRAM_HTTP_ADDR` | `:3333` |
| `ENGRAM_HTTP_TOKEN` | (none) |
| `ENGRAM_EMBEDDING_MODEL` | `builtin` |
| `ENGRAM_OLLAMA_ENDPOINT` | `http://localhost:11434` |
| `ENGRAM_OLLAMA_MODEL` | `nomic-embed-text` |
| `ENGRAM_MAX_FILE_SIZE` | `1048576` |
| `ENGRAM_WAL_MODE` | `true` |
| `ENGRAM_EMBEDDING_BATCH_SIZE` | `32` |

## Docker

```bash
docker build -t engram .
docker run -v $(pwd):/workspace engram index
docker run -v $(pwd):/workspace engram serve
```

For team HTTP setups, use Docker Compose:

```bash
docker compose up -d
# Connect AI tools to http://localhost:3333
```

Image details: Alpine 3.19, ~30MB, non-root user. See [docs/docker.md](docs/docker.md).

## License

MIT
