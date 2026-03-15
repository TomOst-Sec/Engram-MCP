package parser

import (
	"path/filepath"
	"testing"

	"github.com/TomOst-Sec/colony-project/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreSymbols(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	symbols := []Symbol{
		{
			Name:      "HandleRequest",
			Type:      "function",
			Language:  "go",
			Signature: "func HandleRequest(w http.ResponseWriter, r *http.Request) error",
			Docstring: "Handles incoming requests",
			StartLine: 10,
			EndLine:   20,
			FilePath:  "internal/handler.go",
			BodyHash:  "abc123",
		},
		{
			Name:      "User",
			Type:      "type",
			Language:  "go",
			Signature: "type User struct",
			StartLine: 1,
			EndLine:   5,
			FilePath:  "internal/handler.go",
			BodyHash:  "def456",
		},
	}

	err := StoreSymbols(store, "filehash123", symbols)
	require.NoError(t, err)

	// Verify rows were inserted
	var count int
	err = store.DB().QueryRow("SELECT COUNT(*) FROM code_index WHERE file_path = ?", "internal/handler.go").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestStoreSymbolsFTS5Search(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	symbols := []Symbol{
		{
			Name:      "HandleRequest",
			Type:      "function",
			Language:  "go",
			Signature: "func HandleRequest(w http.ResponseWriter, r *http.Request) error",
			Docstring: "Handles incoming HTTP requests",
			StartLine: 10,
			EndLine:   20,
			FilePath:  "internal/handler.go",
			BodyHash:  "abc123",
		},
	}

	err := StoreSymbols(store, "filehash123", symbols)
	require.NoError(t, err)

	// Search via FTS5
	var name string
	err = store.DB().QueryRow("SELECT symbol_name FROM code_index_fts WHERE code_index_fts MATCH ?", "HandleRequest").Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "HandleRequest", name)
}

func TestDeleteFileSymbols(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	symbols := []Symbol{
		{Name: "Foo", Type: "function", Language: "go", StartLine: 1, EndLine: 2, FilePath: "a.go", BodyHash: "x"},
	}
	require.NoError(t, StoreSymbols(store, "h1", symbols))

	err := DeleteFileSymbols(store, "a.go")
	require.NoError(t, err)

	var count int
	store.DB().QueryRow("SELECT COUNT(*) FROM code_index WHERE file_path = ?", "a.go").Scan(&count)
	assert.Equal(t, 0, count)
}

func TestGetFileHash(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	symbols := []Symbol{
		{Name: "Bar", Type: "function", Language: "go", StartLine: 1, EndLine: 2, FilePath: "b.go", BodyHash: "y"},
	}
	require.NoError(t, StoreSymbols(store, "myhash42", symbols))

	hash, err := GetFileHash(store, "b.go")
	require.NoError(t, err)
	assert.Equal(t, "myhash42", hash)
}

func openTestStore(t *testing.T) *storage.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	return store
}
