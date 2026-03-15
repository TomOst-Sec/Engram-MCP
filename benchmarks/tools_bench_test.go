package benchmarks

import (
	"context"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/TomOst-Sec/colony-project/internal/tools/memory"
	searchtools "github.com/TomOst-Sec/colony-project/internal/tools/search"
)

func BenchmarkSearchCodeTool(b *testing.B) {
	store := setupIndexedStore(b, 100)
	defer store.Close()

	tool := searchtools.NewSearchTool(store, nil, b.TempDir())
	req := mcpgo.CallToolRequest{
		Params: mcpgo.CallToolParams{
			Name: "search_code",
			Arguments: map[string]interface{}{
				"query": "handler",
				"limit": float64(10),
			},
		},
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := tool.Handle(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRememberTool(b *testing.B) {
	store := setupIndexedStore(b, 10)
	defer store.Close()

	tool := memory.NewRememberTool(store)
	req := mcpgo.CallToolRequest{
		Params: mcpgo.CallToolParams{
			Name: "remember",
			Arguments: map[string]interface{}{
				"content": "We decided to use SQLite with FTS5 for the storage layer",
				"type":    "decision",
			},
		},
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := tool.Handle(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRecallTool(b *testing.B) {
	store := setupIndexedStore(b, 10)
	defer store.Close()

	// Seed some memories first
	remTool := memory.NewRememberTool(store)
	for _, content := range []string{
		"Decided to use SQLite for storage",
		"Fixed race condition in concurrent writes",
		"Learned about FTS5 ranking functions",
	} {
		req := mcpgo.CallToolRequest{
			Params: mcpgo.CallToolParams{
				Arguments: map[string]interface{}{
					"content": content,
					"type":    "decision",
				},
			},
		}
		remTool.Handle(context.Background(), req)
	}

	recallTool := memory.NewRecallTool(store)
	req := mcpgo.CallToolRequest{
		Params: mcpgo.CallToolParams{
			Name: "recall",
			Arguments: map[string]interface{}{
				"query": "sqlite storage",
				"limit": float64(10),
			},
		},
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := recallTool.Handle(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
