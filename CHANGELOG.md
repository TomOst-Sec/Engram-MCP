# Changelog

All notable changes to Engram are documented in this file.

## [0.1.0] - 2026-03-15

Initial release. All 20 features from the project specification delivered.

### MCP Server
- JSON-RPC 2.0 over stdio transport (primary)
- HTTP/SSE transport with bearer token auth and CORS support (`engram serve --transport http`)
- 8 MCP tools: `search_code`, `remember`, `recall`, `get_architecture`, `get_history`, `get_conventions`, `get_conventions_prompt`, `engram_status`
- Graceful shutdown and health check endpoint

### Code Intelligence
- Tree-sitter AST indexing for 15 languages: Go, Python, TypeScript, JavaScript, Rust, Java, C#, Ruby, PHP, Swift, Kotlin, C, C++, Lua, Zig
- Extracts functions, methods, classes, types, imports, exports, docstrings
- ONNX embedding pipeline (bundled all-MiniLM-L6-v2, 384-dim vectors, batch inference)
- Hybrid search combining FTS5 full-text search with vector similarity ranking
- Incremental re-indexing via file hash comparison

### Memory System
- `remember` tool stores session memories with type, tags, and related files
- `recall` tool retrieves memories via hybrid FTS5 + semantic search
- Memory types: decision, bugfix, refactor, learning, convention
- Soft-delete support

### Architecture Analysis
- Auto-detected module structure from import graph
- Dependency mapping, export detection, complexity scoring
- Go module path awareness

### Git History
- Blame context, commit message extraction, last-author tracking
- Hotspot detection (most-changed files)
- Co-changed file patterns

### Convention Detection
- Naming pattern inference (camelCase, snake_case, PascalCase) per language
- Test coverage analysis
- Documentation coverage analysis
- Community convention pack registry (`engram conventions add <pack>`)
- Convention enforcement prompt for AI context injection

### CLI
- `engram serve` — Start MCP server (stdio or HTTP)
- `engram index` — Full repository indexing (`--watch` for live re-indexing via fsnotify)
- `engram search <query>` — Terminal code search
- `engram recall <query>` — Terminal memory search
- `engram status` — Index statistics
- `engram init` — Generate configuration
- `engram tui` — Interactive terminal dashboard (bubbletea)
- `engram export` / `engram import` — JSON data portability
- `engram ci-hook` — Parse CI/CD output (GitHub Actions, GitLab CI, generic)
- `engram conventions` — Manage convention packs
- Lipgloss-styled terminal output

### Storage
- SQLite with FTS5 full-text search, WAL mode, auto-vacuum
- Tables: `code_index`, `memories`, `conventions`, `architecture`, `git_context`
- Schema versioning with additive migrations
- Database per project in `~/.engram/<repo-hash>/`

### TUI Dashboard
- Status panel: index statistics, language breakdown, embedding count
- Memories panel: search and browse session memories
- Conventions panel: view detected patterns per language
- Architecture panel: module dependency visualization

### Integrations
- Ollama integration for optional local LLM embeddings
- Multi-repo support (cross-repository search)
- Docker image (Alpine 3.19, ~30MB, non-root)
- GoReleaser config for cross-compilation (linux/darwin, amd64/arm64)
- Integration guides for Claude Code, Cursor, Codex, Windsurf, Copilot

### Build
- Requires `sqlite_fts5` build tag (`make build` handles this)
- `.envrc` for direnv users sets `GOFLAGS` automatically
- Benchmark suite for indexing, search, memory, and tool performance
