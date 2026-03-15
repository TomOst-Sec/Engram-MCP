package conventions

import (
	"context"
	"encoding/json"

	conv "github.com/TomOst-Sec/colony-project/internal/conventions"
	"github.com/TomOst-Sec/colony-project/internal/storage"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// ConventionsTool implements the get_conventions MCP tool.
type ConventionsTool struct {
	store    *storage.Store
	analyzer *conv.Analyzer
}

// NewConventionsTool creates a new ConventionsTool.
func NewConventionsTool(store *storage.Store, repoRoot string) *ConventionsTool {
	return &ConventionsTool{
		store:    store,
		analyzer: conv.New(store, repoRoot),
	}
}

// Definition returns the MCP tool definition.
func (t *ConventionsTool) Definition() mcpgo.Tool {
	return mcpgo.NewTool("get_conventions",
		mcpgo.WithDescription("Get inferred code conventions and team patterns. Returns naming styles, error handling patterns, test structure, and more."),
		mcpgo.WithString("language",
			mcpgo.Description("Filter by language (go, python, typescript, etc.)"),
		),
		mcpgo.WithString("category",
			mcpgo.Description("Filter by category (naming, error_handling, testing, imports, structure, documentation)"),
		),
	)
}

// Handle processes a get_conventions request.
func (t *ConventionsTool) Handle(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	args := request.GetArguments()

	language, _ := args["language"].(string)
	category, _ := args["category"].(string)

	// Try to get stored conventions first
	conventions, err := t.analyzer.GetConventions(ctx, language, category)
	if err != nil || len(conventions) == 0 {
		// Run analysis on-the-fly
		result, err := t.analyzer.Analyze(ctx)
		if err != nil {
			return mcpgo.NewToolResultError("Failed to analyze conventions: " + err.Error()), nil
		}

		// Store the results
		conv.StoreConventions(t.store, result.Conventions)

		// Re-query with filters
		conventions, err = t.analyzer.GetConventions(ctx, language, category)
		if err != nil {
			return mcpgo.NewToolResultError("Failed to retrieve conventions: " + err.Error()), nil
		}
	}

	if len(conventions) == 0 {
		return mcpgo.NewToolResultText("No conventions detected. Make sure the codebase has been indexed with 'engram index' first."), nil
	}

	data, err := json.MarshalIndent(conventions, "", "  ")
	if err != nil {
		return mcpgo.NewToolResultError("Failed to serialize conventions: " + err.Error()), nil
	}

	return mcpgo.NewToolResultText(string(data)), nil
}
