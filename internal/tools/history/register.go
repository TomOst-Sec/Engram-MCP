package history

import (
	engmcp "github.com/TomOst-Sec/colony-project/internal/mcp"
	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// RegisterTools registers the get_history tool with the MCP server.
func RegisterTools(server *engmcp.Server, store *storage.Store) {
	tool := NewHistoryTool(store)
	server.RegisterTool(tool.Definition(), tool.Handle)
}
