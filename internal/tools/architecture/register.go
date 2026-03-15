package architecture

import (
	engrammcp "github.com/TomOst-Sec/colony-project/internal/mcp"
	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// RegisterTools registers the get_architecture tool with the MCP server.
func RegisterTools(server *engrammcp.Server, store *storage.Store, repoRoot, goModulePath string) {
	tool := NewArchitectureTool(store, repoRoot, goModulePath)
	server.RegisterTool(tool.Definition(), tool.Handle)
}
