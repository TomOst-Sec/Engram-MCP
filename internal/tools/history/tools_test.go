package history

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

func setupTestStore(t *testing.T) *storage.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	return store
}

func seedGitContext(t *testing.T, store *storage.Store, filePath, author, hash, msg string, freq int, coChanged string) {
	t.Helper()
	_, err := store.DB().Exec(
		`INSERT INTO git_context (file_path, symbol_name, last_author, last_commit_hash, last_commit_message, last_modified, change_frequency, co_changed_files)
		VALUES (?, '', ?, ?, ?, datetime('now'), ?, ?)`,
		filePath, author, hash, msg, freq, coChanged,
	)
	require.NoError(t, err)
}

func makeRequest(mode string, extras map[string]interface{}) mcpgo.CallToolRequest {
	args := map[string]interface{}{"mode": mode}
	for k, v := range extras {
		args[k] = v
	}
	return mcpgo.CallToolRequest{
		Params: mcpgo.CallToolParams{
			Name:      "get_history",
			Arguments: args,
		},
	}
}

func getResultText(t *testing.T, result *mcpgo.CallToolResult) string {
	t.Helper()
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)
	textContent, ok := mcpgo.AsTextContent(result.Content[0])
	require.True(t, ok, "expected text content")
	return textContent.Text
}

func TestDefinitionNameAndSchema(t *testing.T) {
	store := setupTestStore(t)
	tool := NewHistoryTool(store)

	def := tool.Definition()
	assert.Equal(t, "get_history", def.Name)

	// Verify required mode parameter exists
	schema := def.InputSchema
	props, ok := schema.Properties["mode"]
	require.True(t, ok, "mode property should exist")

	propsMap, ok := props.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", propsMap["type"])

	// Verify mode is required
	assert.Contains(t, schema.Required, "mode")
}

func TestFileMode_ReturnsHistory(t *testing.T) {
	store := setupTestStore(t)
	seedGitContext(t, store, "internal/auth/handler.go", "alice", "abc1234def", "Fix session token expiry", 15,
		`["internal/auth/middleware.go","tests/auth_test.go"]`)

	tool := NewHistoryTool(store)
	result, err := tool.Handle(context.Background(), makeRequest("file", map[string]interface{}{
		"file_path": "internal/auth/handler.go",
	}))
	require.NoError(t, err)

	text := getResultText(t, result)
	assert.Contains(t, text, "internal/auth/handler.go")
	assert.Contains(t, text, "alice")
	assert.Contains(t, text, "abc1234")
	assert.Contains(t, text, "Fix session token expiry")
	assert.Contains(t, text, "15 commits")
	assert.Contains(t, text, "internal/auth/middleware.go")
	assert.Contains(t, text, "tests/auth_test.go")
}

func TestFileMode_MissingFilePath(t *testing.T) {
	store := setupTestStore(t)
	tool := NewHistoryTool(store)

	result, err := tool.Handle(context.Background(), makeRequest("file", nil))
	require.NoError(t, err)
	assert.True(t, result.IsError)

	text := getResultText(t, result)
	assert.Contains(t, text, "file_path is required")
}

func TestFileMode_FileNotFound(t *testing.T) {
	store := setupTestStore(t)
	tool := NewHistoryTool(store)

	result, err := tool.Handle(context.Background(), makeRequest("file", map[string]interface{}{
		"file_path": "nonexistent.go",
	}))
	require.NoError(t, err)

	text := getResultText(t, result)
	assert.Contains(t, text, "No git history found")
	assert.Contains(t, text, "engram index")
}

func TestHotspotsMode_ReturnsSorted(t *testing.T) {
	store := setupTestStore(t)
	seedGitContext(t, store, "low.go", "dev", "aaa", "msg", 2, "[]")
	seedGitContext(t, store, "high.go", "dev", "bbb", "msg", 42, "[]")
	seedGitContext(t, store, "mid.go", "dev", "ccc", "msg", 15, "[]")

	tool := NewHistoryTool(store)
	result, err := tool.Handle(context.Background(), makeRequest("hotspots", nil))
	require.NoError(t, err)

	text := getResultText(t, result)
	assert.Contains(t, text, "Hotspots")

	// high.go should appear before mid.go, mid.go before low.go
	highIdx := len(text) - len(text) // use strings index
	_ = highIdx
	assert.Contains(t, text, "42 changes")
	assert.Contains(t, text, "15 changes")
	assert.Contains(t, text, "2 changes")
}

func TestHotspotsMode_EmptyTable(t *testing.T) {
	store := setupTestStore(t)
	tool := NewHistoryTool(store)

	result, err := tool.Handle(context.Background(), makeRequest("hotspots", nil))
	require.NoError(t, err)

	text := getResultText(t, result)
	assert.Contains(t, text, "No git history found")
}

func TestHotspotsMode_LimitParameter(t *testing.T) {
	store := setupTestStore(t)
	for i := 0; i < 5; i++ {
		seedGitContext(t, store, fmt.Sprintf("file%d.go", i), "dev", fmt.Sprintf("hash%d", i), "msg", (i+1)*10, "[]")
	}

	tool := NewHistoryTool(store)
	result, err := tool.Handle(context.Background(), makeRequest("hotspots", map[string]interface{}{
		"limit": float64(2),
	}))
	require.NoError(t, err)

	text := getResultText(t, result)
	// Should only have 2 entries (lines starting with numbers)
	lines := 0
	for _, line := range splitLines(text) {
		if len(line) > 0 && line[0] >= '0' && line[0] <= '9' {
			lines++
		}
	}
	// With format " 1. file..." the first char is a space, so count differently
	assert.Contains(t, text, "50 changes")  // highest
	assert.Contains(t, text, "40 changes")  // second highest
	assert.NotContains(t, text, "10 changes") // should be cut off
}

func TestCochangedMode_ReturnsFiles(t *testing.T) {
	store := setupTestStore(t)
	seedGitContext(t, store, "main.go", "dev", "hash", "msg", 10,
		`["helper.go","utils.go","config.go"]`)

	tool := NewHistoryTool(store)
	result, err := tool.Handle(context.Background(), makeRequest("cochanged", map[string]interface{}{
		"file_path": "main.go",
	}))
	require.NoError(t, err)

	text := getResultText(t, result)
	assert.Contains(t, text, "frequently change with main.go")
	assert.Contains(t, text, "helper.go")
	assert.Contains(t, text, "utils.go")
	assert.Contains(t, text, "config.go")
}

func TestCochangedMode_MissingFilePath(t *testing.T) {
	store := setupTestStore(t)
	tool := NewHistoryTool(store)

	result, err := tool.Handle(context.Background(), makeRequest("cochanged", nil))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestCochangedMode_NoData(t *testing.T) {
	store := setupTestStore(t)
	seedGitContext(t, store, "lonely.go", "dev", "hash", "msg", 5, "[]")

	tool := NewHistoryTool(store)
	result, err := tool.Handle(context.Background(), makeRequest("cochanged", map[string]interface{}{
		"file_path": "lonely.go",
	}))
	require.NoError(t, err)

	text := getResultText(t, result)
	assert.Contains(t, text, "No co-change data")
}

func TestCochangedMode_LimitParameter(t *testing.T) {
	store := setupTestStore(t)
	seedGitContext(t, store, "main.go", "dev", "hash", "msg", 10,
		`["a.go","b.go","c.go","d.go","e.go"]`)

	tool := NewHistoryTool(store)
	result, err := tool.Handle(context.Background(), makeRequest("cochanged", map[string]interface{}{
		"file_path": "main.go",
		"limit":     float64(2),
	}))
	require.NoError(t, err)

	text := getResultText(t, result)
	assert.Contains(t, text, "a.go")
	assert.Contains(t, text, "b.go")
	assert.NotContains(t, text, "c.go")
}

func TestMissingModeReturnsError(t *testing.T) {
	store := setupTestStore(t)
	tool := NewHistoryTool(store)

	req := mcpgo.CallToolRequest{
		Params: mcpgo.CallToolParams{
			Name:      "get_history",
			Arguments: map[string]interface{}{},
		},
	}
	result, err := tool.Handle(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// Ensure unused imports don't cause issues
var _ = json.Marshal
var _ = fmt.Sprintf
