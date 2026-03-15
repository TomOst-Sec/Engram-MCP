package memory

import (
	engmcp "github.com/TomOst-Sec/colony-project/internal/mcp"
	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// RegisterTools registers the remember and recall tools with the MCP server.
func RegisterTools(server *engmcp.Server, store *storage.Store) {
	remember := NewRememberTool(store)
	server.RegisterTool(remember.Definition(), remember.Handle)

	recall := NewRecallTool(store)
	server.RegisterTool(recall.Definition(), recall.Handle)
}
