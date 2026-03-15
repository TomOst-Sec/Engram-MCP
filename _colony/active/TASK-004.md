# TASK-004: MCP Server Core — JSON-RPC 2.0 Stdio Transport, Tool Registration

**Priority:** P0
**Assigned:** alpha
**Milestone:** M1: MVP
**Dependencies:** TASK-001
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
The MCP (Model Context Protocol) server is the primary interface through which AI coding tools interact with Engram. It speaks JSON-RPC 2.0 over stdio (reading requests from stdin, writing responses to stdout). This task creates the server skeleton using the `mark3labs/mcp-go` SDK: server initialization, tool registration system, stdio transport, graceful shutdown, and a basic health/version tool to prove the pipeline works end-to-end.

## Specification
Create the `internal/mcp` package implementing the MCP server core.

### Dependencies
- `github.com/mark3labs/mcp-go` — MCP SDK for Go (handles JSON-RPC 2.0, MCP spec compliance)

### Server Structure
```go
// internal/mcp/server.go
type Server struct {
    mcpServer *server.MCPServer
    // future: store reference, config reference
}

func New(name, version string) *Server
func (s *Server) RegisterTool(tool mcp.Tool, handler server.ToolHandlerFunc)
func (s *Server) ServeStdio() error     // blocks, reads stdin, writes stdout
func (s *Server) Shutdown() error
```

### What to Implement
1. **Server initialization** — Create an MCP server with `server.NewMCPServer()`, set name="engram", version from build info
2. **Tool registration** — A method to register MCP tools with their handlers. Each tool has a name, description, JSON schema for inputs, and a handler function.
3. **Stdio transport** — Use `server.NewStdioServer()` from mcp-go to wrap the MCP server and serve over stdin/stdout
4. **Graceful shutdown** — Handle SIGINT/SIGTERM, drain in-flight requests, close cleanly
5. **Built-in `engram_status` tool** — A simple tool that returns server version, uptime, and a "healthy" status. This proves the full pipeline works: client sends request → server routes to handler → handler returns result.

### engram_status tool spec
- **Name:** `engram_status`
- **Description:** "Returns Engram server status including version, uptime, and health"
- **Input schema:** no required parameters
- **Response:**
```json
{
    "version": "0.1.0-dev",
    "status": "healthy",
    "uptime_seconds": 42
}
```

## Acceptance Criteria
- [ ] MCP server initializes without error
- [ ] `engram_status` tool is registered and callable
- [ ] Server can start stdio transport (reads from stdin, writes to stdout)
- [ ] Server handles graceful shutdown on SIGINT
- [ ] `engram_status` returns valid JSON with version, status, and uptime_seconds fields
- [ ] All tests pass with `go test ./internal/mcp/`

## Implementation Steps
1. Run `go get github.com/mark3labs/mcp-go`
2. Create `internal/mcp/server.go`:
   - Server struct wrapping mcp-go's MCPServer
   - New() constructor that creates the server with name and version
   - RegisterTool() method
   - ServeStdio() method using server.NewStdioServer
   - Shutdown() with signal handling
3. Create `internal/mcp/tools.go`:
   - Define the `engram_status` tool (name, description, schema)
   - Implement the handler: return version, "healthy" status, and uptime since server start
   - RegisterBuiltinTools() function that registers engram_status
4. Create `internal/mcp/server_test.go`:
   - Test: New() creates server without error
   - Test: RegisterTool adds a tool
   - Test: engram_status handler returns correct JSON structure
   - Test: Status response contains expected fields
5. Run tests

## Testing Requirements
- Unit test: New("engram", "0.1.0-dev") returns non-nil Server
- Unit test: engram_status handler returns map with "version", "status", "uptime_seconds" keys
- Unit test: status "version" field matches the version passed to New()
- Unit test: "status" field is "healthy"
- Unit test: "uptime_seconds" is a non-negative number
- Integration test (optional, if feasible): pipe a JSON-RPC initialize request through stdin, verify response on stdout

## Files to Create/Modify
- `internal/mcp/server.go` — Server struct, New, RegisterTool, ServeStdio, Shutdown
- `internal/mcp/tools.go` — engram_status tool definition and handler, RegisterBuiltinTools
- `internal/mcp/server_test.go` — all tests

## Notes
- Read the mcp-go documentation carefully. The SDK provides `server.NewMCPServer()` for creating the server and `server.NewStdioServer()` for the stdio transport. Use the SDK's conventions.
- For signal handling, use `os/signal` with `signal.NotifyContext` for clean context cancellation.
- The `engram_status` tool is intentionally simple — it's a proof-of-life for the MCP pipeline. Real tools (search, remember, recall) come in later tasks.
- Do NOT start the stdio server in tests (it blocks on stdin). Test the handler functions directly.
- Store the server start time in the Server struct to compute uptime.

---
## Completion Notes
- **Completed by:** alpha-1
- **Date:** 2026-03-15 15:05:24
- **Branch:** task/004

## Rejection Notes
- **Rejected by:** audit
- **Date:** 2026-03-15 15:08
- **Reason:** CODE IS APPROVED — merge conflict in go.mod/go.sum due to TASK-003 merge. Coder needs to rebase on current main (`git pull origin main --rebase`), resolve trivial dependency conflict, run `go mod tidy`, verify tests, and re-push. See `_colony/bugs/BUG-004.md`.
