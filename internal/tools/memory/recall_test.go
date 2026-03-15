package memory

import (
	"context"
	"encoding/json"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertMemory(t *testing.T, tool *RememberTool, content, memType string) {
	t.Helper()
	req := makeCallRequest(map[string]any{
		"content": content,
		"type":    memType,
	})
	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)
	require.False(t, result.IsError)
}

type recallResponse struct {
	Memories     []map[string]any `json:"memories"`
	TotalMatches int              `json:"total_matches"`
	Query        string           `json:"query"`
}

func parseRecallResponse(t *testing.T, result *mcpgo.CallToolResult) recallResponse {
	t.Helper()
	var resp recallResponse
	text := result.Content[0].(mcpgo.TextContent).Text
	require.NoError(t, json.Unmarshal([]byte(text), &resp))
	return resp
}

func TestRecallFindsByKeyword(t *testing.T) {
	store := newTestStore(t)
	remember := NewRememberTool(store)
	recall := NewRecallTool(store)

	insertMemory(t, remember, "We chose JWT tokens for authentication", "decision")
	insertMemory(t, remember, "Fixed null pointer in user handler", "bugfix")
	insertMemory(t, remember, "Refactored database connection pooling", "refactor")
	insertMemory(t, remember, "Learned about Go context cancellation", "learning")
	insertMemory(t, remember, "Always use snake_case for JSON fields", "convention")

	req := makeCallRequest(map[string]any{
		"query": "authentication",
	})
	result, err := recall.Handle(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	resp := parseRecallResponse(t, result)
	assert.Greater(t, len(resp.Memories), 0)
	assert.Equal(t, "authentication", resp.Query)

	// The JWT/authentication memory should be in results
	found := false
	for _, m := range resp.Memories {
		if content, ok := m["content"].(string); ok {
			if content == "We chose JWT tokens for authentication" {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "should find the authentication memory")
}

func TestRecallFilterByType(t *testing.T) {
	store := newTestStore(t)
	remember := NewRememberTool(store)
	recall := NewRecallTool(store)

	insertMemory(t, remember, "authentication decision made", "decision")
	insertMemory(t, remember, "authentication bug fixed", "bugfix")

	req := makeCallRequest(map[string]any{
		"query": "authentication",
		"type":  "decision",
	})
	result, err := recall.Handle(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	resp := parseRecallResponse(t, result)
	for _, m := range resp.Memories {
		assert.Equal(t, "decision", m["type"])
	}
}

func TestRecallFilterByDate(t *testing.T) {
	store := newTestStore(t)
	remember := NewRememberTool(store)
	recall := NewRecallTool(store)

	insertMemory(t, remember, "old memory about testing", "learning")

	// Set one memory's created_at to the past
	_, err := store.DB().Exec("UPDATE memories SET created_at = datetime('2020-01-01') WHERE id = 1")
	require.NoError(t, err)

	insertMemory(t, remember, "new memory about testing", "learning")

	req := makeCallRequest(map[string]any{
		"query": "testing",
		"since": "2025-01-01",
	})
	result, err := recall.Handle(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	resp := parseRecallResponse(t, result)
	assert.Equal(t, 1, resp.TotalMatches)
	require.Len(t, resp.Memories, 1)
	assert.Equal(t, "new memory about testing", resp.Memories[0]["content"])
}

func TestRecallLimitParameter(t *testing.T) {
	store := newTestStore(t)
	remember := NewRememberTool(store)
	recall := NewRecallTool(store)

	for i := 0; i < 10; i++ {
		insertMemory(t, remember, "memory about database operations", "learning")
	}

	req := makeCallRequest(map[string]any{
		"query": "database",
		"limit": 3,
	})
	result, err := recall.Handle(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	resp := parseRecallResponse(t, result)
	assert.Len(t, resp.Memories, 3)
	assert.Equal(t, 10, resp.TotalMatches)
}

func TestRecallExcludesSoftDeleted(t *testing.T) {
	store := newTestStore(t)
	remember := NewRememberTool(store)
	recall := NewRecallTool(store)

	insertMemory(t, remember, "visible memory about routing", "learning")
	insertMemory(t, remember, "deleted memory about routing", "learning")

	// Soft-delete the second memory
	_, err := store.DB().Exec("UPDATE memories SET deleted_at = CURRENT_TIMESTAMP WHERE id = 2")
	require.NoError(t, err)

	req := makeCallRequest(map[string]any{
		"query": "routing",
	})
	result, err := recall.Handle(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	resp := parseRecallResponse(t, result)
	assert.Equal(t, 1, resp.TotalMatches)
	require.Len(t, resp.Memories, 1)
	assert.Equal(t, "visible memory about routing", resp.Memories[0]["content"])
}

func TestRecallEmptyQueryReturnsError(t *testing.T) {
	store := newTestStore(t)
	recall := NewRecallTool(store)

	req := makeCallRequest(map[string]any{
		"query": "",
	})
	result, err := recall.Handle(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
