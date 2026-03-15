# Project Goals

> **Instructions:** Human fills this file with project requirements.
> ATLAS reads this to generate the roadmap and tasks.
> Update this file whenever goals change — ATLAS will pick up changes on its next cycle.

## Product Description

Engram is an open-source, single-binary MCP (Model Context Protocol) server written in Go that gives every AI coding agent — Cursor, Claude Code, Codex, Windsurf, Copilot, and any MCP-compatible client — a persistent, intelligent memory of your entire codebase, team conventions, architectural decisions, and project history. It works by scanning your repository to build a local semantic index (tree-sitter AST parsing + bundled ONNX embeddings), analyzing git history for decision context, inferring team conventions from actual code patterns, and storing compressed session memories in a local SQLite database. Install with `npx engram init`, connect any MCP-compatible tool with zero configuration, and your AI instantly understands your codebase the way a senior teammate would — remembering past sessions, enforcing your team's patterns, and providing deep architectural context. Zero dependencies, zero cloud services, zero API keys. Everything runs locally on your machine.

## Tech Stack

**Core Runtime:**
- Go 1.22+ — single static binary, cross-compiled for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- CGo enabled only for SQLite (mattn/go-sqlite3) — all other deps are pure Go

**Storage:**
- SQLite 3.45+ with FTS5 (full-text search), JSON1, and WAL mode for concurrent reads
- Schema: `memories` table (session decisions, compressed context), `code_index` table (AST nodes, embeddings, file hashes), `conventions` table (inferred patterns), `architecture` table (module graph, dependency map), `git_context` table (blame/log decision trails)

**AST Parsing:**
- tree-sitter via go bindings (smacker/go-tree-sitter) — bundled grammars for: TypeScript, JavaScript, Python, Go, Rust, Java, C#, Ruby, PHP, Swift, Kotlin, C, C++, Lua, Zig
- Custom tree-sitter queries per language to extract: function signatures, type definitions, imports, exports, class hierarchies, error handling patterns, test structures

**Embeddings:**
- ONNX Runtime via onnxruntime-go — bundled all-MiniLM-L6-v2 model (~23MB quantized INT8)
- 384-dimensional vectors stored in SQLite as BLOB, indexed with custom HNSW implementation in pure Go (or fallback brute-force for repos <10K nodes)
- Optional: Ollama integration for users who want to use a local LLM for embedding instead

**MCP Server:**
- JSON-RPC 2.0 over stdio (primary) and HTTP/SSE (secondary, for remote/team setups)
- Implements MCP spec 2025-03-26: tools, resources, prompts, sampling
- Uses mark3labs/mcp-go SDK

**CLI & TUI:**
- CLI: cobra + lipgloss for styled terminal output
- TUI: charmbracelet/bubbletea for interactive dashboard (memory browser, convention viewer, index status)

**Build & Distribution:**
- GoReleaser for cross-compilation + checksums + changelogs
- npm wrapper package (`npx engram init`) that downloads the correct binary for the platform
- Homebrew tap, AUR package, Scoop bucket, GitHub Releases with .deb/.rpm/.apk
- Docker image (alpine-based, ~30MB) for CI/CD integration

**Testing:**
- Go standard testing + testify for assertions
- Golden file tests for AST extraction (snapshot tree-sitter output per language)
- Integration tests: spin up MCP server, connect mock client, assert tool responses
- Benchmark suite: index time, query latency, memory usage per repo size (1K, 10K, 100K, 1M LOC)

## Features

1. **MCP Server Core** — JSON-RPC 2.0 stdio transport implementing the MCP spec. Exposes tool discovery, tool invocation, and resource listing. Handles concurrent requests. Graceful shutdown. Health check endpoint. This is the skeleton everything hangs on.

2. **`search_code` Tool** — Semantic codebase search combining FTS5 keyword search with ONNX vector similarity. Input: natural language query + optional filters (language, directory, symbol type). Output: ranked list of code snippets with file paths, line numbers, relevance scores, and surrounding context. Must return results in <200ms for repos up to 100K LOC.

3. **Tree-Sitter AST Indexer** — Parses every source file in the repo using language-specific tree-sitter grammars. Extracts a normalized representation: functions (name, params, return type, body hash, docstring), types/classes (fields, methods, inheritance), imports/exports, error handling patterns, test functions. Stores in SQLite with file hash for incremental re-indexing. Supports 15 languages at launch.

4. **ONNX Embedding Pipeline** — Generates 384-dim vectors for every extracted code symbol using bundled all-MiniLM-L6-v2. Batched inference (32 items/batch). Vectors stored alongside AST nodes in SQLite. Cosine similarity search with optional HNSW index for large repos. The embedding pipeline must be fully offline — no network calls ever.

5. **`get_architecture` Tool** — Returns a structured map of the project: top-level modules, their responsibilities (inferred from directory names, file patterns, and docstrings), inter-module dependencies (import graph), entry points, and key abstractions. Output format: JSON with `modules[]`, each containing `name`, `path`, `description`, `dependencies[]`, `exports[]`, `complexity_score`. For monorepos, detects and maps sub-projects.

6. **`remember` Tool** — Stores a new memory from the current coding session. Input: `content` (what happened), `type` (decision | bugfix | refactor | learning | convention), optional `tags[]`, optional `related_files[]`. The memory is timestamped, embedded for semantic search, and compressed using an extractive summarization pass (via the bundled model or simple TF-IDF extraction). Memories are append-only with soft-delete.

7. **`recall` Tool** — Retrieves relevant memories from past sessions. Input: natural language query + optional filters (type, date range, tags). Uses hybrid search: FTS5 keyword match + vector similarity on memory embeddings. Returns ranked memories with timestamps, types, and related file paths. Critical for session continuity — this is what makes the AI "remember yesterday."

8. **`get_conventions` Tool** — Analyzes the codebase to infer team conventions and returns them as structured rules. Detects: naming patterns (camelCase vs snake_case per context), error handling style (try/catch vs Result types vs error codes), test structure (describe/it vs test functions, fixture patterns), import ordering, file organization patterns, comment/docstring style, state management patterns, API response formats. Output: JSON array of `{ pattern, confidence, examples[], description }`. Confidence scored by consistency across the codebase (>80% = high).

9. **Persistent SQLite Storage Layer** — All data lives in `~/.engram/<repo-hash>/engram.db`. WAL mode for concurrent read access from multiple MCP clients. Auto-vacuum. Schema migrations via embedded SQL files. Backup/export to JSON. Import from JSON (for sharing). Database size target: <100MB for a 100K LOC repo.

10. **Git History Analyzer** — Parses `git log` and `git blame` to extract decision context. For each function/type in the index, stores: last modified date, author, commit message (as decision context), frequency of changes (hotspot detection), co-change patterns (files that always change together). Exposes via `get_history` tool: "Why does this function exist? When was it last changed? Who owns it?"

11. **Incremental Re-Indexing (`--watch` mode)** — Uses fsnotify to watch the repo for file changes. On save, re-parses only the changed file, updates AST nodes, regenerates embeddings for changed symbols, updates the architecture graph if imports changed. Must complete incremental update in <500ms for a single file change. Background goroutine, no CLI interaction needed.

12. **`npx engram init` Bootstrap** — npm wrapper package that: detects platform/arch, downloads the correct Go binary from GitHub Releases, runs initial full index, generates `engram.json` config file, outputs connection instructions for Cursor/Claude Code/Codex/Windsurf. The entire flow from `npx engram init` to "your AI tool is connected" must take <60 seconds on a 50K LOC repo.

13. **CLI Interface** — `engram serve` (start MCP server), `engram index` (full re-index), `engram search <query>` (CLI search), `engram recall <query>` (CLI memory search), `engram status` (index stats, memory count, last indexed), `engram conventions` (print detected conventions), `engram export` (dump DB to JSON), `engram import` (load from JSON). All commands use lipgloss for clean terminal output.

14. **HTTP/SSE Transport** — Secondary MCP transport for remote and team scenarios. `engram serve --http :3333` starts an HTTP server implementing MCP over SSE. Supports bearer token auth. Enables: team-shared Engram instances, CI/CD integration, remote development setups. CORS configurable.

15. **Convention Enforcement Prompts** — MCP `prompts` resource that AI tools can request before generating code. Returns a system prompt fragment containing the project's inferred conventions, architectural patterns, and relevant memories. The AI tool includes this in its context window automatically. This is the "automatic quality injection" — no manual prompt engineering needed.

16. **TUI Dashboard** — `engram tui` launches an interactive terminal UI (bubbletea). Panels: memory browser (search, filter, delete memories), convention viewer (see all detected patterns with confidence scores), index status (files indexed, languages, last update, DB size), architecture viewer (ASCII module dependency graph). Keyboard-driven, vim-style navigation.

17. **Multi-Repo Support** — `engram.json` can reference multiple repos (monorepo sub-projects or related microservices). The MCP server merges indexes and provides cross-repo search. Architecture tool shows inter-repo dependencies. Memories are scoped per-repo but cross-searchable.

18. **Ollama Integration** — Optional: `engram serve --embeddings ollama:nomic-embed-text` uses a local Ollama instance for embeddings instead of the bundled ONNX model. For users who want higher-quality embeddings or want to use a custom model. Also enables optional memory compression via local LLM (summarize verbose session logs into concise memories).

19. **Community Convention Modules** — Registry of shareable convention packs: `engram conventions add react-typescript`, `engram conventions add go-clean-arch`, `engram conventions add rails-standard`. Packs are JSON files in a community GitHub repo. Users can contribute their team's conventions as a pack. Engram merges community conventions with locally-inferred ones (local always wins on conflict).

20. **CI/CD Memory Hook** — `engram ci-hook` reads CI/CD output (test failures, build errors, deployment logs) and stores them as memories. When a developer starts a new session, the AI knows "the last 3 builds failed because of a timeout in the payment service." Supports GitHub Actions, GitLab CI, and generic stdin parsing.

## Milestones

### Milestone 1: MVP (Week 1-2) — "It works, it's fast, it's useful"
- MCP Server Core (Feature 1) — stdio transport, tool discovery, basic error handling
- Tree-Sitter AST Indexer (Feature 3) — 6 languages first: TypeScript, Python, Go, Rust, Java, C#
- ONNX Embedding Pipeline (Feature 4) — bundled model, batch inference, vector storage
- `search_code` Tool (Feature 2) — hybrid FTS5 + vector search, <200ms response
- `get_architecture` Tool (Feature 5) — module map, import graph, entry points
- `remember` Tool (Feature 6) — store session memories with type and tags
- `recall` Tool (Feature 7) — hybrid search over memories
- Persistent SQLite Storage (Feature 9) — WAL mode, schema v1, auto-vacuum
- CLI: `engram serve`, `engram index`, `engram search`, `engram status`
- README with viral demo GIF (split-screen: AI without Engram vs with Engram)
- Integration guide: Claude Code (`claude mcp add engram -- engram serve`)
- Integration guide: Cursor (`.cursor/mcp.json` config)
- **Ship criteria:** A developer can `go install` the binary, run `engram index && engram serve`, connect Claude Code or Cursor, and get meaningfully better AI responses within 5 minutes. Memories persist across sessions.

### Milestone 2: Core Features (Week 3-4) — "Every developer needs this"
- `get_conventions` Tool (Feature 8) — pattern inference, confidence scoring, 10+ pattern types
- Git History Analyzer (Feature 10) — blame context, hotspot detection, decision trails
- Incremental Re-Indexing / `--watch` mode (Feature 11) — fsnotify, <500ms updates
- `npx engram init` Bootstrap (Feature 12) — npm wrapper, platform detection, guided setup
- Full CLI Interface (Feature 13) — all commands, lipgloss styling, help text
- Convention Enforcement Prompts (Feature 15) — auto-injected context for AI tools
- Remaining 9 languages for tree-sitter: JavaScript, Ruby, PHP, Swift, Kotlin, C, C++, Lua, Zig
- Integration guide: Codex, Windsurf, Copilot
- **Ship criteria:** `npx engram init` works on macOS/Linux/Windows. Conventions are inferred automatically. Git history provides decision context. Watch mode keeps index fresh. Works with 5+ AI coding tools.

### Milestone 3: Polish & Growth (Month 2) — "Teams adopt this"
- HTTP/SSE Transport (Feature 14) — remote/team MCP server with auth
- TUI Dashboard (Feature 16) — interactive memory/convention/architecture browser
- Multi-Repo Support (Feature 17) — monorepo and microservice scenarios
- Ollama Integration (Feature 18) — optional local LLM for embeddings and compression
- Community Convention Modules (Feature 19) — registry, `engram conventions add <pack>`
- Homebrew tap, AUR package, Scoop bucket, .deb/.rpm packages
- Docker image for CI/CD environments
- Benchmark suite published: index time, query latency, memory usage at scale
- **Ship criteria:** Teams of 5+ can share an Engram instance. TUI provides full visibility. Community has contributed 10+ convention packs. Performance validated at 500K LOC.

### Milestone 4: Ecosystem (Month 3) — "It's the standard"
- CI/CD Memory Hook (Feature 20) — GitHub Actions, GitLab CI, generic stdin
- VS Code extension — sidebar showing Engram status, memories, conventions (read-only viewer, Engram still runs as MCP server)
- Architectural diagram generation — ASCII and Mermaid output from the architecture graph
- Memory decay — old, low-relevance memories auto-archive to reduce noise
- Cross-session learning — detect recurring patterns in memories (e.g., "developer keeps fixing the same type of bug") and surface proactive suggestions
- Plugin system — Go plugin interface for custom convention detectors, custom memory processors, custom MCP tools
- **Ship criteria:** Engram is the default MCP context server recommended in AI coding tool documentation. 10K+ GitHub stars. Active community contributing convention packs and plugins.

## Constraints

- **Zero network calls.** Engram must never phone home, send telemetry, or make any network request during normal operation. The ONNX model is bundled. Ollama integration is opt-in and local-only. The npm installer downloads the binary from GitHub Releases — that's the only network call in the entire lifecycle.
- **Single binary.** The Go binary must be fully self-contained. No runtime dependencies (except libc for SQLite CGo). No Docker required. No Python. No Node.js runtime (the npm package is just a thin downloader). A developer should be able to `curl` the binary and run it.
- **Sub-200ms tool responses.** Every MCP tool call must return in <200ms for repos up to 100K LOC. The `search_code` tool must return in <100ms for cached queries. Indexing is the only operation that can be slow (target: <30s for 100K LOC full index, <500ms incremental).
- **<100MB disk for 100K LOC repo.** The SQLite database (AST index + embeddings + memories + git context) must stay under 100MB for a typical 100K LOC repository. Embeddings are quantized INT8 (384 dims x 1 byte = 384 bytes per vector). Convention and memory compression keeps text storage lean.
- **<200MB RAM steady-state.** The running MCP server (with index loaded) must use less than 200MB RSS for a 100K LOC repo. ONNX model loaded on-demand for embedding generation, unloaded after idle timeout. SQLite uses mmap for reads.
- **Privacy by default.** No code leaves the machine. No API keys required for core functionality. Memories stored locally. Git history analysis runs locally. The only optional external integration is Ollama (which is also local). Suitable for corporate environments, HIPAA-adjacent codebases, and developers who simply don't want their code in the cloud.
- **MCP spec compliance.** Strict adherence to MCP spec 2025-03-26. Must pass the official MCP conformance test suite. Interoperable with any compliant MCP client without special handling.
- **Backward-compatible schema.** SQLite schema changes must use additive migrations only. Never drop columns. Never rename tables. A newer Engram version must read databases created by any older version. Forward compatibility is best-effort.
- **Cross-platform.** First-class support for macOS (Intel + Apple Silicon), Linux (x86_64 + ARM64), and Windows (x86_64). All features work identically on all platforms. CI tests on all six platform/arch combinations.
- **No CGo beyond SQLite.** tree-sitter bindings must use the pure-Go WASM approach (or maintain CGo bindings with static linking). The ONNX runtime uses CGo but bundles the dylib. Goal: minimize CGo surface to keep cross-compilation simple.
- **Graceful degradation.** If tree-sitter grammar is missing for a language, fall back to regex-based symbol extraction. If ONNX runtime fails to load, fall back to FTS5-only search (no vectors). If git is not available, skip history analysis. Every feature should degrade, never crash.
- **Idempotent indexing.** Running `engram index` twice produces identical results. File hashes determine re-indexing. No stale data accumulates. Clean re-index (`engram index --force`) drops and rebuilds everything.
- **MIT License.** Fully permissive. No CLA. No contributor restrictions. Corporate-friendly from day one.

## Out of Scope

- **Not an AI coding agent.** Engram does not generate code, review PRs, or make edits. It is a context/memory layer that makes existing agents better. It never invokes an LLM to produce code suggestions.
- **Not a cloud service.** No SaaS. No hosted version. No account creation. No billing. If someone wants to host it for their team, they run the binary on their own infrastructure.
- **Not an IDE plugin (beyond MCP).** We don't build a VS Code extension that duplicates MCP functionality. The optional VS Code extension (Milestone 4) is a read-only status viewer — all intelligence flows through the MCP protocol.
- **Not a code review tool.** Engram provides context to AI agents that may do code review, but Engram itself does not review code, flag issues, or generate review comments.
- **Not a replacement for `.cursorrules` / `CLAUDE.md` / etc.** Engram supplements these files. If a developer has a hand-written `CLAUDE.md`, Engram's conventions are additive. We don't auto-generate or overwrite tool-specific config files.
- **Not a vector database.** The embedded HNSW index is intentionally simple. We don't support ANN benchmarking, custom distance functions, or multi-tenant vector search. If you need that, use a real vector DB.
- **Not a git client.** Engram reads git history via `git log` and `git blame` shell commands. It does not implement git operations, manage branches, or interact with GitHub/GitLab APIs.
- **Not a language server.** Engram does AST-level parsing but does not provide LSP features (go-to-definition, autocomplete, diagnostics). It's complementary to LSP, not a replacement.
- **Not supporting non-MCP AI tools.** If a tool doesn't support MCP, we don't build custom integrations for it. MCP is the universal protocol — tools that don't support it are out of scope. (Exception: the npm bootstrap can generate legacy config file stubs as a courtesy.)
- **Not handling binary files.** Images, compiled assets, PDFs, and other binary files are excluded from indexing. Only source code files matching known extensions are processed.
- **Not multi-tenant.** One Engram instance serves one developer (or one team sharing a repo). There is no user management, RBAC, or workspace isolation within a single instance.
