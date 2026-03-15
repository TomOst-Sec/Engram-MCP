# TASK-032: HTTP/SSE Transport — Remote MCP Server

**Priority:** P1
**Assigned:** alpha
**Milestone:** M3: Polish & Growth
**Dependencies:** TASK-004, TASK-022
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
Feature 14 from GOALS.md. The `engram serve --http :3333` command starts an HTTP server implementing MCP over SSE (Server-Sent Events). This enables team-shared Engram instances, CI/CD integration, and remote development setups. Currently `engram serve` only supports stdio transport. This task adds the HTTP/SSE alternative.

## Specification

### HTTP Transport: `internal/mcp/http.go`

Implement HTTP/SSE transport following the MCP spec:

```go
// ServeHTTP starts the HTTP/SSE MCP transport.
func (s *Server) ServeHTTP(addr string, token string) error
```

### MCP over HTTP/SSE Protocol
The MCP spec defines HTTP/SSE transport:

**Endpoints:**
- `POST /mcp` — JSON-RPC request/response (standard MCP tool calls)
- `GET /mcp/sse` — SSE stream for server-initiated messages (notifications)
- `GET /health` — Health check endpoint (returns 200 with server info)

**Request flow:**
1. Client sends JSON-RPC request via `POST /mcp`
2. Server processes and returns JSON-RPC response
3. For streaming responses, client connects to `GET /mcp/sse`

### Authentication
- Optional bearer token auth: `Authorization: Bearer <token>`
- Token set via `--http-token` flag or `http_token` config field
- If no token configured, server is open (suitable for localhost/team use)

### CORS
- Configurable CORS headers for browser-based MCP clients
- Default: allow all origins (development mode)
- Production: set `--cors-origin` flag

### CLI Integration

Update `cmd/engram/serve.go`:
```go
// When transport == "http":
fmt.Fprintf(os.Stderr, "Engram MCP server starting on %s\n", httpAddr)
return server.ServeHTTP(httpAddr, httpToken)
```

Add flags:
```go
cmd.Flags().StringVar(&httpAddr, "http-addr", ":3333", "HTTP server address (when --transport=http)")
cmd.Flags().StringVar(&httpToken, "http-token", "", "Bearer token for HTTP auth (optional)")
cmd.Flags().StringVar(&corsOrigin, "cors-origin", "*", "CORS allowed origin")
```

### Health Check Response
```json
{
  "name": "engram",
  "version": "0.1.0",
  "transport": "http",
  "tools": 7,
  "uptime_seconds": 3600
}
```

## Acceptance Criteria
- [ ] `engram serve --transport http` starts an HTTP server on the configured address
- [ ] `POST /mcp` accepts JSON-RPC requests and returns responses
- [ ] `GET /health` returns server status JSON
- [ ] Bearer token auth works when configured
- [ ] Requests without valid token return 401 when auth is enabled
- [ ] CORS headers are set correctly
- [ ] Server shuts down gracefully on SIGINT
- [ ] All existing stdio transport tests still pass
- [ ] All new HTTP transport tests pass

## Implementation Steps
1. Create `internal/mcp/http.go`:
   - HTTP handler for `POST /mcp` (JSON-RPC dispatch)
   - SSE handler for `GET /mcp/sse`
   - Health check handler for `GET /health`
   - Auth middleware for bearer token
   - CORS middleware
   - ServeHTTP method on Server
2. Update `cmd/engram/serve.go`:
   - Add --http-addr, --http-token, --cors-origin flags
   - Replace "not yet implemented" error with actual ServeHTTP call
3. Create `internal/mcp/http_test.go`:
   - Test: POST /mcp with valid JSON-RPC returns response
   - Test: GET /health returns server info
   - Test: Auth middleware rejects missing token
   - Test: Auth middleware accepts valid token
   - Test: CORS headers present in response
   - Test: Invalid JSON-RPC request returns error response
4. Run all tests

## Testing Requirements
- Unit test: HTTP handler dispatches JSON-RPC requests to tool handlers
- Unit test: Health endpoint returns correct JSON structure
- Unit test: Auth middleware blocks unauthenticated requests
- Unit test: Auth middleware passes valid bearer tokens
- Unit test: CORS headers set in response
- Integration test: Full round-trip — send tool_call via HTTP, receive response

## Files to Create/Modify
- `internal/mcp/http.go` — HTTP/SSE transport implementation
- `internal/mcp/http_test.go` — HTTP transport tests
- `cmd/engram/serve.go` — add HTTP flags, wire up ServeHTTP

## Notes
- Use Go's `net/http` standard library. No external HTTP framework needed.
- For SSE, use `text/event-stream` content type with `Transfer-Encoding: chunked`.
- The JSON-RPC dispatch should reuse the same tool handler infrastructure as stdio. The only difference is the transport layer — stdio reads/writes JSON-RPC from stdin/stdout, HTTP reads/writes from HTTP request/response.
- Check how `mark3labs/mcp-go` handles HTTP transport. It may already have HTTP support that we can use directly. If so, use it instead of building from scratch.
- The `--http-token` value should NOT be logged or printed to stderr for security.
- Set `ReadTimeout` and `WriteTimeout` on the HTTP server (30 seconds each).
