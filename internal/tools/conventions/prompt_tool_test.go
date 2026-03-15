package conventions

import (
	"context"
	"path/filepath"
	"testing"

	conv "github.com/TomOst-Sec/colony-project/internal/conventions"
	"github.com/TomOst-Sec/colony-project/internal/storage"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPromptTestStore(t *testing.T) *storage.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	return store
}

func TestPromptToolDefinition(t *testing.T) {
	tool := NewConventionsPromptTool(nil)
	def := tool.Definition()
	assert.Equal(t, "get_conventions_prompt", def.Name)
}

func TestPromptToolEmptyConventions(t *testing.T) {
	store := setupPromptTestStore(t)
	tool := NewConventionsPromptTool(store)

	req := mcpgo.CallToolRequest{}
	req.Params.Name = "get_conventions_prompt"

	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	text := result.Content[0].(mcpgo.TextContent).Text
	assert.Contains(t, text, "engram index")
}

func TestPromptToolWithConventions(t *testing.T) {
	store := setupPromptTestStore(t)

	conventions := []conv.Convention{
		{
			Pattern:     "camelCase function names",
			Description: "go functions predominantly use camelCase naming",
			Category:    "naming",
			Confidence:  0.95,
			Examples:    []string{"handleRequest", "parseConfig"},
			Language:    "go",
		},
		{
			Pattern:     "test coverage pattern",
			Description: "go: 15% of symbols are tests",
			Category:    "testing",
			Confidence:  0.9,
			Language:    "go",
		},
	}
	err := conv.StoreConventions(store, conventions)
	require.NoError(t, err)

	tool := NewConventionsPromptTool(store)

	req := mcpgo.CallToolRequest{}
	req.Params.Name = "get_conventions_prompt"

	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	text := result.Content[0].(mcpgo.TextContent).Text
	assert.Contains(t, text, "Project Coding Conventions")
	assert.Contains(t, text, "camelCase")
	assert.Contains(t, text, "95%")
	assert.Contains(t, text, "handleRequest")
	assert.Contains(t, text, "Testing")
}

func TestPromptToolLanguageFilter(t *testing.T) {
	store := setupPromptTestStore(t)

	conventions := []conv.Convention{
		{
			Pattern:     "camelCase function names",
			Description: "go functions use camelCase",
			Category:    "naming",
			Confidence:  0.95,
			Language:    "go",
		},
		{
			Pattern:     "snake_case function names",
			Description: "python functions use snake_case",
			Category:    "naming",
			Confidence:  0.92,
			Language:    "python",
		},
	}
	err := conv.StoreConventions(store, conventions)
	require.NoError(t, err)

	tool := NewConventionsPromptTool(store)

	req := mcpgo.CallToolRequest{}
	req.Params.Name = "get_conventions_prompt"
	req.Params.Arguments = map[string]interface{}{"language": "go"}

	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	text := result.Content[0].(mcpgo.TextContent).Text
	assert.Contains(t, text, "camelCase")
	assert.NotContains(t, text, "snake_case")
}

func TestFormatConventionsPromptLimit(t *testing.T) {
	// Test that formatConventionsPrompt handles many conventions
	var conventions []conv.Convention
	for i := 0; i < 20; i++ {
		conventions = append(conventions, conv.Convention{
			Pattern:     "pattern",
			Description: "description",
			Category:    "naming",
			Confidence:  0.9,
			Language:    "go",
		})
	}

	// The Handle function caps at 15, not formatConventionsPrompt directly
	// But we test the formatter handles arbitrary counts
	result := formatConventionsPrompt(conventions)
	assert.Contains(t, result, "Project Coding Conventions")
}
