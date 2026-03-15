# TASK-008: CLI `serve` Command — Start MCP Server via Stdio

**Priority:** P0
**Assigned:** alpha
**Milestone:** M1: MVP
**Dependencies:** TASK-004
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
The `engram serve` command is how users start the MCP server. It's the primary entry point for AI coding tools (Cursor, Claude Code, Codex) to connect to Engram. This task wires up the cobra CLI to the MCP server core (TASK-004), creating the `serve` subcommand that starts the stdio transport and handles graceful shutdown. Without this command, there's no way to launch Engram from the terminal.

## Specification
Add an `engram serve` subcommand to the cobra CLI that starts the MCP server.

### What to Build

**`cmd/engram/serve.go`** — The serve subcommand:
```go
var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the Engram MCP server",
    Long:  "Start the Engram MCP server using stdio transport. AI coding tools connect to this server to access codebase intelligence.",
    RunE:  runServe,
}
```

**`runServe` function behavior:**
1. Load configuration using `config.Load()` from the current working directory
2. Create the MCP server using `mcp.New("engram", version)`
3. Register all built-in tools (engram_status from TASK-004)
4. Log startup info to stderr: `"Engram MCP server starting (version %s, transport: stdio)"`
5. Start the stdio transport: `server.ServeStdio()`
6. Handle graceful shutdown on SIGINT/SIGTERM:
   - Log `"Shutting down..."` to stderr
   - Call `server.Shutdown()`
   - Exit cleanly with code 0

**Key design decisions:**
- All diagnostic output goes to **stderr** (stdout is reserved for JSON-RPC)
- The version string comes from a build-time variable: `var version = "0.1.0-dev"`
- The command registers itself with the root command in an `init()` function
- Configuration loading is optional at this stage — if config.Load fails, use defaults and log a warning

### CLI Flags for `serve`
- `--transport` — `stdio` (default) or `http` (placeholder, not implemented yet — print "HTTP transport not yet implemented" and exit)
- `--log-level` — `debug`, `info` (default), `warn`, `error` — controls verbosity of stderr output

### Integration with main.go
The serve command must register itself with the root command. Update `cmd/engram/main.go` to import the serve command's package or use cobra's `AddCommand` pattern.

## Acceptance Criteria
- [ ] `engram serve` starts the MCP server and blocks (reading stdin for JSON-RPC)
- [ ] `engram serve` prints startup message to stderr, NOT stdout
- [ ] Ctrl+C (SIGINT) causes graceful shutdown with exit code 0
- [ ] `--transport stdio` is the default behavior
- [ ] `--transport http` prints "not yet implemented" and exits with code 1
- [ ] `--log-level` flag is accepted (debug/info/warn/error)
- [ ] `engram serve --help` shows proper usage text
- [ ] The serve command registers correctly with the root cobra command
- [ ] All tests pass with `go test ./cmd/engram/`

## Implementation Steps
1. Create `cmd/engram/serve.go` with the serve command definition
2. Implement `runServe()`:
   - Set up signal handling (SIGINT, SIGTERM) via `signal.NotifyContext`
   - Create MCP server with `mcp.New()`
   - Register built-in tools
   - Log startup to stderr
   - Call `server.ServeStdio()` with the context
   - On context cancellation, call `server.Shutdown()`
3. Update `cmd/engram/main.go` to add the serve command to root: `rootCmd.AddCommand(serveCmd)`
4. Create `cmd/engram/serve_test.go`:
   - Test: serve command is registered on root command
   - Test: --help flag produces output containing "serve" and "MCP"
   - Test: --transport flag with invalid value returns error
   - Test: signal handling triggers shutdown (use a goroutine + signal send)
5. Run `go test ./cmd/engram/` — all pass

## Testing Requirements
- Unit test: serveCmd is registered and findable via rootCmd.Find("serve")
- Unit test: --help output contains "MCP server" and "stdio"
- Unit test: runServe with --transport=invalid returns error
- Unit test: version variable is set to a non-empty string

## Files to Create/Modify
- `cmd/engram/serve.go` — serve subcommand, runServe function, flag definitions
- `cmd/engram/serve_test.go` — serve command tests
- `cmd/engram/main.go` — add serveCmd to root command (minimal change)

## Notes
- Do NOT import `internal/storage` in this task — the serve command doesn't open the database yet. That integration comes later when tools that need storage are wired in.
- The MCP server's `ServeStdio()` blocks reading from stdin. Tests should NOT call ServeStdio() directly — test the command registration, flag parsing, and error handling instead.
- All log/diagnostic output MUST go to stderr. stdout is the JSON-RPC transport channel. Mixing diagnostic output into stdout will break MCP protocol communication.
- Use `fmt.Fprintf(os.Stderr, ...)` for logging in MVP. A proper logger integration is a later task.
- The `--transport http` flag exists as a placeholder for Milestone 3. For now, it should return an error explaining HTTP is not yet implemented.
