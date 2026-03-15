package memory

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

var validTypes = map[string]bool{
	"decision":   true,
	"bugfix":     true,
	"refactor":   true,
	"learning":   true,
	"convention": true,
}

// RememberTool stores memories from coding sessions.
type RememberTool struct {
	store     *storage.Store
	sessionID string
}

// NewRememberTool creates a new RememberTool with a generated session ID.
func NewRememberTool(store *storage.Store) *RememberTool {
	return &RememberTool{
		store:     store,
		sessionID: generateSessionID(),
	}
}

// SessionID returns the tool's session ID (useful for testing).
func (t *RememberTool) SessionID() string {
	return t.sessionID
}

// Definition returns the MCP tool definition with input schema.
func (t *RememberTool) Definition() mcpgo.Tool {
	return mcpgo.NewTool("remember",
		mcpgo.WithDescription("Store a memory from the current coding session for future reference"),
		mcpgo.WithString("content",
			mcpgo.Required(),
			mcpgo.Description("What to remember (decision, learning, bug fix, etc.)"),
		),
		mcpgo.WithString("type",
			mcpgo.Required(),
			mcpgo.Description("Category of this memory"),
			mcpgo.Enum("decision", "bugfix", "refactor", "learning", "convention"),
		),
		mcpgo.WithArray("tags",
			mcpgo.Description("Optional tags for categorization"),
		),
		mcpgo.WithArray("related_files",
			mcpgo.Description("Optional file paths related to this memory"),
		),
	)
}

// Handle processes a remember tool call.
func (t *RememberTool) Handle(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	content, err := request.RequireString("content")
	if err != nil {
		return toolError("content is required"), nil
	}
	if content == "" {
		return toolError("content must not be empty"), nil
	}

	memType, err := request.RequireString("type")
	if err != nil {
		return toolError("type is required"), nil
	}
	if !validTypes[memType] {
		return toolError(fmt.Sprintf("invalid type %q: must be one of decision, bugfix, refactor, learning, convention", memType)), nil
	}

	tagsJSON := marshalStringSlice(request.GetStringSlice("tags", nil))
	filesJSON := marshalStringSlice(request.GetStringSlice("related_files", nil))

	result, err := t.store.DB().ExecContext(ctx,
		`INSERT INTO memories (content, type, tags, related_files, session_id) VALUES (?, ?, ?, ?, ?)`,
		content, memType, tagsJSON, filesJSON, t.sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("storing memory: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting memory ID: %w", err)
	}

	preview := content
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}

	response := map[string]any{
		"id":              id,
		"status":          "stored",
		"created_at":      time.Now().UTC().Format(time.RFC3339),
		"content_preview": preview,
	}
	data, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("marshaling response: %w", err)
	}

	return mcpgo.NewToolResultText(string(data)), nil
}

func generateSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func marshalStringSlice(s []string) string {
	if s == nil {
		return "[]"
	}
	data, _ := json.Marshal(s)
	return string(data)
}

func toolError(msg string) *mcpgo.CallToolResult {
	result := mcpgo.NewToolResultText(msg)
	result.IsError = true
	return result
}
