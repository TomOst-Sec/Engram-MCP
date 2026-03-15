package mcp

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// Server wraps the mcp-go MCPServer to provide Engram's MCP interface.
type Server struct {
	mcpServer *mcpserver.MCPServer
	version   string
	startTime time.Time
}

// New creates a new MCP server with the given name and version.
func New(name, version string) *Server {
	return &Server{
		mcpServer: mcpserver.NewMCPServer(name, version,
			mcpserver.WithToolCapabilities(false),
		),
		version:   version,
		startTime: time.Now(),
	}
}

// RegisterTool adds a tool and its handler to the server.
func (s *Server) RegisterTool(tool mcpgo.Tool, handler mcpserver.ToolHandlerFunc) {
	s.mcpServer.AddTool(tool, handler)
}

// ServeStdio starts the server on stdin/stdout with graceful shutdown
// on SIGINT or SIGTERM.
func (s *Server) ServeStdio() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	stdioServer := mcpserver.NewStdioServer(s.mcpServer)
	return stdioServer.Listen(ctx, os.Stdin, os.Stdout)
}

// Shutdown signals the server to stop. In stdio mode, closing stdin
// or sending SIGINT achieves the same effect.
func (s *Server) Shutdown() error {
	return nil
}
