package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// RegisterBuiltinTools registers all built-in tools with the server.
func RegisterBuiltinTools(s *Server) {
	tool := mcpgo.NewTool("engram_status",
		mcpgo.WithDescription("Returns Engram server status including version, uptime, and health"),
	)
	s.RegisterTool(tool, s.handleStatus)
}

// handleStatus returns the server's current status.
func (s *Server) handleStatus(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	uptime := time.Since(s.startTime).Seconds()
	response := map[string]interface{}{
		"version":        s.version,
		"status":         "healthy",
		"uptime_seconds": int(uptime),
	}
	data, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal status: %w", err)
	}
	return mcpgo.NewToolResultText(string(data)), nil
}
