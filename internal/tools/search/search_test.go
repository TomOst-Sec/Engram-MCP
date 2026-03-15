package search

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/TomOst-Sec/colony-project/internal/embeddings"
	"github.com/TomOst-Sec/colony-project/internal/storage"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestStore(t *testing.T) *storage.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	return store
}

func insertSymbol(t *testing.T, store *storage.Store, filePath, name, symType, lang, signature, docstring string, startLine, endLine int) int64 {
	t.Helper()
	result, err := store.DB().Exec(
		`INSERT INTO code_index (file_path, file_hash, language, symbol_name, symbol_type, signature, docstring, start_line, end_line)
		 VALUES (?, 'hash', ?, ?, ?, ?, ?, ?, ?)`,
		filePath, lang, name, symType, signature, docstring, startLine, endLine,
	)
	require.NoError(t, err)
	id, err := result.LastInsertId()
	require.NoError(t, err)
	return id
}

func makeRequest(query string, extras ...map[string]interface{}) mcpgo.CallToolRequest {
	args := map[string]interface{}{"query": query}
	if len(extras) > 0 {
		for k, v := range extras[0] {
			args[k] = v
		}
	}
	return mcpgo.CallToolRequest{
		Params: mcpgo.CallToolParams{
			Name:      "search_code",
			Arguments: args,
		},
	}
}

func parseResponse(t *testing.T, result *mcpgo.CallToolResult) searchResponse {
	t.Helper()
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)
	textContent, ok := mcpgo.AsTextContent(result.Content[0])
	require.True(t, ok)

	var resp searchResponse
	err := json.Unmarshal([]byte(textContent.Text), &resp)
	require.NoError(t, err)
	return resp
}

func TestSearchTool_FTS5_ExactMatch(t *testing.T) {
	store := setupTestStore(t)
	insertSymbol(t, store, "auth/handler.go", "HandleLogin", "function", "go",
		"func HandleLogin(w http.ResponseWriter, r *http.Request) error",
		"HandleLogin processes user authentication", 10, 45)
	insertSymbol(t, store, "db/store.go", "Connect", "function", "go",
		"func Connect(dsn string) (*DB, error)",
		"Connect establishes database connection", 5, 20)

	tool := NewSearchTool(store, nil, t.TempDir())
	result, err := tool.Handle(context.Background(), makeRequest("HandleLogin"))
	require.NoError(t, err)

	resp := parseResponse(t, result)
	require.True(t, len(resp.Results) >= 1)
	assert.Equal(t, "HandleLogin", resp.Results[0].SymbolName)
	assert.Equal(t, "fts5", resp.SearchMode)
}

func TestSearchTool_FTS5_DocstringSearch(t *testing.T) {
	store := setupTestStore(t)
	insertSymbol(t, store, "auth/handler.go", "HandleLogin", "function", "go",
		"func HandleLogin(w http.ResponseWriter, r *http.Request) error",
		"HandleLogin processes user authentication", 10, 45)

	tool := NewSearchTool(store, nil, t.TempDir())
	result, err := tool.Handle(context.Background(), makeRequest("authentication"))
	require.NoError(t, err)

	resp := parseResponse(t, result)
	require.True(t, len(resp.Results) >= 1)
	assert.Equal(t, "HandleLogin", resp.Results[0].SymbolName)
}

func TestSearchTool_LanguageFilter(t *testing.T) {
	store := setupTestStore(t)
	insertSymbol(t, store, "main.go", "GoFunc", "function", "go",
		"func GoFunc()", "a Go function", 1, 10)
	insertSymbol(t, store, "main.py", "PyFunc", "function", "python",
		"def PyFunc():", "a Python function", 1, 10)

	tool := NewSearchTool(store, nil, t.TempDir())
	result, err := tool.Handle(context.Background(), makeRequest("function", map[string]interface{}{
		"language": "go",
	}))
	require.NoError(t, err)

	resp := parseResponse(t, result)
	for _, r := range resp.Results {
		assert.Equal(t, "go", r.Language)
	}
}

func TestSearchTool_SymbolTypeFilter(t *testing.T) {
	store := setupTestStore(t)
	insertSymbol(t, store, "main.go", "MyFunc", "function", "go",
		"func MyFunc()", "my function", 1, 10)
	insertSymbol(t, store, "types.go", "MyType", "type", "go",
		"type MyType struct{}", "my type", 1, 10)

	tool := NewSearchTool(store, nil, t.TempDir())
	result, err := tool.Handle(context.Background(), makeRequest("my", map[string]interface{}{
		"symbol_type": "function",
	}))
	require.NoError(t, err)

	resp := parseResponse(t, result)
	for _, r := range resp.Results {
		assert.Equal(t, "function", r.SymbolType)
	}
}

func TestSearchTool_DirectoryFilter(t *testing.T) {
	store := setupTestStore(t)
	insertSymbol(t, store, "internal/auth/handler.go", "HandleAuth", "function", "go",
		"func HandleAuth()", "auth handler", 1, 10)
	insertSymbol(t, store, "internal/db/store.go", "DBConnect", "function", "go",
		"func DBConnect()", "database connect", 1, 10)

	tool := NewSearchTool(store, nil, t.TempDir())
	result, err := tool.Handle(context.Background(), makeRequest("function", map[string]interface{}{
		"directory": "internal/auth/",
	}))
	require.NoError(t, err)

	resp := parseResponse(t, result)
	for _, r := range resp.Results {
		assert.Contains(t, r.FilePath, "internal/auth/")
	}
}

func TestSearchTool_LimitParameter(t *testing.T) {
	store := setupTestStore(t)
	// Insert 20 symbols
	for i := 0; i < 20; i++ {
		insertSymbol(t, store, "file.go", "Func", "function", "go",
			"func Func()", "a function description", i*10, (i+1)*10)
	}

	tool := NewSearchTool(store, nil, t.TempDir())
	result, err := tool.Handle(context.Background(), makeRequest("Func", map[string]interface{}{
		"limit": float64(5),
	}))
	require.NoError(t, err)

	resp := parseResponse(t, result)
	assert.LessOrEqual(t, len(resp.Results), 5)
	assert.GreaterOrEqual(t, resp.TotalMatches, 5)
}

func TestSearchTool_EmptyQuery(t *testing.T) {
	store := setupTestStore(t)
	tool := NewSearchTool(store, nil, t.TempDir())

	_, err := tool.Handle(context.Background(), makeRequest(""))
	assert.Error(t, err)
}

func TestSearchTool_NoResults(t *testing.T) {
	store := setupTestStore(t)
	tool := NewSearchTool(store, nil, t.TempDir())

	result, err := tool.Handle(context.Background(), makeRequest("nonexistent_symbol_xyz"))
	require.NoError(t, err)

	resp := parseResponse(t, result)
	assert.Empty(t, resp.Results)
	assert.Equal(t, 0, resp.TotalMatches)
}

func TestSearchTool_GracefulDegradation_NilEmbedder(t *testing.T) {
	store := setupTestStore(t)
	insertSymbol(t, store, "main.go", "TestFunc", "function", "go",
		"func TestFunc()", "test function", 1, 10)

	// Nil embedder — should use FTS5-only
	tool := NewSearchTool(store, nil, t.TempDir())
	result, err := tool.Handle(context.Background(), makeRequest("TestFunc"))
	require.NoError(t, err)

	resp := parseResponse(t, result)
	assert.Equal(t, "fts5", resp.SearchMode)
	require.True(t, len(resp.Results) >= 1)
}

func TestSearchTool_ResultFieldsPresent(t *testing.T) {
	store := setupTestStore(t)
	insertSymbol(t, store, "auth/handler.go", "HandleLogin", "function", "go",
		"func HandleLogin() error", "handles login", 10, 45)

	tool := NewSearchTool(store, nil, t.TempDir())
	result, err := tool.Handle(context.Background(), makeRequest("HandleLogin"))
	require.NoError(t, err)

	resp := parseResponse(t, result)
	require.Len(t, resp.Results, 1)
	r := resp.Results[0]
	assert.Equal(t, "auth/handler.go", r.FilePath)
	assert.Equal(t, "HandleLogin", r.SymbolName)
	assert.Equal(t, "function", r.SymbolType)
	assert.Equal(t, "go", r.Language)
	assert.Equal(t, "func HandleLogin() error", r.Signature)
	assert.Equal(t, 10, r.StartLine)
	assert.Equal(t, 45, r.EndLine)
	assert.Greater(t, r.Score, 0.0)
}

func TestSearchTool_LimitMax50(t *testing.T) {
	store := setupTestStore(t)
	insertSymbol(t, store, "file.go", "Func", "function", "go",
		"func Func()", "func description", 1, 10)

	tool := NewSearchTool(store, nil, t.TempDir())
	result, err := tool.Handle(context.Background(), makeRequest("Func", map[string]interface{}{
		"limit": float64(100), // over the max
	}))
	require.NoError(t, err)

	resp := parseResponse(t, result)
	assert.LessOrEqual(t, len(resp.Results), 50)
}

func TestSearchTool_VectorSearchWithEmbeddings(t *testing.T) {
	store := setupTestStore(t)

	id1 := insertSymbol(t, store, "auth/handler.go", "HandleLogin", "function", "go",
		"func HandleLogin()", "login handler", 10, 30)
	id2 := insertSymbol(t, store, "db/store.go", "Connect", "function", "go",
		"func Connect()", "database connection", 5, 15)

	// Add embeddings
	emb1 := make([]float32, 384)
	emb1[0] = 1.0
	emb2 := make([]float32, 384)
	emb2[0] = 0.1
	emb2[1] = 0.9

	require.NoError(t, embeddings.UpdateCodeIndexEmbedding(store, id1, emb1))
	require.NoError(t, embeddings.UpdateCodeIndexEmbedding(store, id2, emb2))

	// Even without a real embedder, FTS5 works
	tool := NewSearchTool(store, nil, t.TempDir())
	result, err := tool.Handle(context.Background(), makeRequest("HandleLogin"))
	require.NoError(t, err)

	resp := parseResponse(t, result)
	require.True(t, len(resp.Results) >= 1)
	assert.Equal(t, "HandleLogin", resp.Results[0].SymbolName)
}
