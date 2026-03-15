package conventions

import (
	engmcp "github.com/TomOst-Sec/colony-project/internal/mcp"
	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// RegisterTools registers all convention-related MCP tools.
func RegisterTools(server *engmcp.Server, store *storage.Store, repoRoot string) {
	tool := NewConventionsTool(store, repoRoot)
	server.RegisterTool(tool.Definition(), tool.Handle)

	promptTool := NewConventionsPromptTool(store)
	server.RegisterTool(promptTool.Definition(), promptTool.Handle)
}
