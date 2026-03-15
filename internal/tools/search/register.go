package search

import (
	"github.com/TomOst-Sec/colony-project/internal/embeddings"
	engmcp "github.com/TomOst-Sec/colony-project/internal/mcp"
	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// RegisterTools registers the search_code tool with the MCP server.
func RegisterTools(server *engmcp.Server, store *storage.Store, embedder *embeddings.Embedder, repoRoot string) {
	tool := NewSearchTool(store, embedder, repoRoot)
	server.RegisterTool(tool.Definition(), tool.Handle)
}
