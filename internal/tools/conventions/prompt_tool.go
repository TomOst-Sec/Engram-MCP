package conventions

import (
	"context"
	"fmt"
	"strings"

	conv "github.com/TomOst-Sec/colony-project/internal/conventions"
	"github.com/TomOst-Sec/colony-project/internal/storage"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// ConventionsPromptTool implements the get_conventions_prompt MCP tool.
type ConventionsPromptTool struct {
	store *storage.Store
}

// NewConventionsPromptTool creates a new ConventionsPromptTool.
func NewConventionsPromptTool(store *storage.Store) *ConventionsPromptTool {
	return &ConventionsPromptTool{store: store}
}

// Definition returns the MCP tool definition.
func (t *ConventionsPromptTool) Definition() mcpgo.Tool {
	return mcpgo.NewTool("get_conventions_prompt",
		mcpgo.WithDescription("Returns a formatted system prompt fragment containing this project's coding conventions. AI tools should include this in their context for style-consistent code generation."),
		mcpgo.WithString("language",
			mcpgo.Description("Restrict to conventions for this language (optional)"),
		),
	)
}

// Handle processes a get_conventions_prompt request.
func (t *ConventionsPromptTool) Handle(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	args := request.GetArguments()
	language, _ := args["language"].(string)

	conventions, err := conv.GetConventions(t.store, language, "")
	if err != nil {
		return mcpgo.NewToolResultError("Failed to retrieve conventions: " + err.Error()), nil
	}

	if len(conventions) == 0 {
		return mcpgo.NewToolResultText("No conventions detected. Run 'engram index' to analyze the codebase first."), nil
	}

	// Limit to top 15 by confidence (already sorted DESC from GetConventions)
	if len(conventions) > 15 {
		conventions = conventions[:15]
	}

	prompt := formatConventionsPrompt(conventions)
	return mcpgo.NewToolResultText(prompt), nil
}

// formatConventionsPrompt formats conventions into a system prompt fragment.
func formatConventionsPrompt(conventions []conv.Convention) string {
	var sb strings.Builder

	sb.WriteString("# Project Coding Conventions\n\n")
	sb.WriteString("Follow these conventions when writing code for this project. These patterns were inferred from the existing codebase with high confidence.\n\n")

	// Group by category
	categories := make(map[string][]conv.Convention)
	categoryOrder := []string{}
	for _, c := range conventions {
		if _, exists := categories[c.Category]; !exists {
			categoryOrder = append(categoryOrder, c.Category)
		}
		categories[c.Category] = append(categories[c.Category], c)
	}

	categoryTitles := map[string]string{
		"naming":        "Naming Conventions",
		"error_handling": "Error Handling",
		"testing":       "Testing",
		"documentation": "Documentation",
		"imports":       "Imports",
		"structure":     "Structure",
	}

	for _, cat := range categoryOrder {
		convs := categories[cat]
		title, ok := categoryTitles[cat]
		if !ok {
			title = strings.Title(cat)
		}
		sb.WriteString("## " + title + "\n")

		for _, c := range convs {
			sb.WriteString(fmt.Sprintf("- %s (confidence: %.0f%%", c.Description, c.Confidence*100))
			if len(c.Examples) > 0 {
				sb.WriteString(", e.g., " + strings.Join(c.Examples, ", "))
			}
			sb.WriteString(")\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("These conventions are automatically detected. Confidence scores indicate how consistently the pattern appears in the codebase.\n")

	return sb.String()
}
