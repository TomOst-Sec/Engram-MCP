package memory

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

func newTestStore(t *testing.T) *storage.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	return store
}

func makeCallRequest(args map[string]any) mcpgo.CallToolRequest {
	return mcpgo.CallToolRequest{
		Params: mcpgo.CallToolParams{
			Arguments: args,
		},
	}
}

func TestRememberStoresMemory(t *testing.T) {
	store := newTestStore(t)
	tool := NewRememberTool(store)

	req := makeCallRequest(map[string]any{
		"content": "We decided to use JWT for authentication",
		"type":    "decision",
		"tags":    []any{"auth", "security"},
		"related_files": []any{"internal/auth/handler.go"},
	})

	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	// Parse response
	var resp map[string]any
	text := result.Content[0].(mcpgo.TextContent).Text
	require.NoError(t, json.Unmarshal([]byte(text), &resp))

	assert.Equal(t, "stored", resp["status"])
	assert.NotZero(t, resp["id"])
	assert.NotEmpty(t, resp["created_at"])
	assert.Contains(t, resp["content_preview"], "We decided to use JWT")

	// Verify in database
	var content, memType, tags, files, sessionID string
	err = store.DB().QueryRow("SELECT content, type, tags, related_files, session_id FROM memories WHERE id = ?", int64(resp["id"].(float64))).
		Scan(&content, &memType, &tags, &files, &sessionID)
	require.NoError(t, err)
	assert.Equal(t, "We decided to use JWT for authentication", content)
	assert.Equal(t, "decision", memType)
	assert.Equal(t, tool.SessionID(), sessionID)
}

func TestRememberRejectsEmptyContent(t *testing.T) {
	store := newTestStore(t)
	tool := NewRememberTool(store)

	req := makeCallRequest(map[string]any{
		"content": "",
		"type":    "decision",
	})

	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestRememberRejectsInvalidType(t *testing.T) {
	store := newTestStore(t)
	tool := NewRememberTool(store)

	req := makeCallRequest(map[string]any{
		"content": "some content",
		"type":    "foo",
	})

	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestRememberSessionIDConsistent(t *testing.T) {
	store := newTestStore(t)
	tool := NewRememberTool(store)

	for i := 0; i < 3; i++ {
		req := makeCallRequest(map[string]any{
			"content": "memory content",
			"type":    "learning",
		})
		result, err := tool.Handle(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, result.IsError)
	}

	// All 3 should have the same session_id
	rows, err := store.DB().Query("SELECT DISTINCT session_id FROM memories")
	require.NoError(t, err)
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		require.NoError(t, rows.Scan(&id))
		ids = append(ids, id)
	}
	assert.Len(t, ids, 1, "all memories should have the same session ID")
	assert.Equal(t, tool.SessionID(), ids[0])
}

func TestRememberTagsStoredAsJSON(t *testing.T) {
	store := newTestStore(t)
	tool := NewRememberTool(store)

	req := makeCallRequest(map[string]any{
		"content":       "tagged memory",
		"type":          "convention",
		"tags":          []any{"naming", "go"},
		"related_files": []any{"internal/config/config.go", "pkg/api/handler.go"},
	})

	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	var tagsJSON, filesJSON string
	err = store.DB().QueryRow("SELECT tags, related_files FROM memories WHERE id = 1").Scan(&tagsJSON, &filesJSON)
	require.NoError(t, err)

	var tags, files []string
	require.NoError(t, json.Unmarshal([]byte(tagsJSON), &tags))
	require.NoError(t, json.Unmarshal([]byte(filesJSON), &files))
	assert.Equal(t, []string{"naming", "go"}, tags)
	assert.Equal(t, []string{"internal/config/config.go", "pkg/api/handler.go"}, files)
}
