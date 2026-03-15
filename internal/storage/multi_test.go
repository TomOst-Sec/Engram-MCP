package storage

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	return store
}

func seedSymbol(t *testing.T, store *Store, filePath, name, symType, lang, signature string) {
	t.Helper()
	_, err := store.DB().Exec(
		`INSERT INTO code_index (file_path, file_hash, language, symbol_name, symbol_type, signature, docstring, start_line, end_line)
		 VALUES (?, 'hash', ?, ?, ?, ?, '', 1, 10)`,
		filePath, lang, name, symType, signature,
	)
	require.NoError(t, err)
}

func TestMultiStoreSearchAcrossStores(t *testing.T) {
	primary := openTestStore(t)
	additional := openTestStore(t)

	seedSymbol(t, primary, "main.go", "login", "function", "go", "func login() handles user login")
	seedSymbol(t, additional, "handler.go", "login_check", "function", "go", "func login_check() validates login")

	ms := NewMultiStore(primary, []*Store{additional}, []string{"api"})

	results, err := ms.SearchCode("login", "", "", 10)
	require.NoError(t, err)
	require.Len(t, results, 2)

	// Both results should be present
	names := make([]string, len(results))
	for i, r := range results {
		names[i] = r.SymbolName
	}
	assert.Contains(t, names, "login")
	assert.Contains(t, names, "login_check")
}

func TestMultiStoreRepoNamePrefix(t *testing.T) {
	primary := openTestStore(t)
	additional := openTestStore(t)

	seedSymbol(t, primary, "main.go", "authenticate", "function", "go", "func authenticate() primary auth")
	seedSymbol(t, additional, "api.go", "authorize", "function", "go", "func authorize() api auth")

	ms := NewMultiStore(primary, []*Store{additional}, []string{"api-service"})

	results, err := ms.SearchCode("auth", "", "", 10)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(results), 2)

	for _, r := range results {
		if r.SymbolName == "authenticate" {
			assert.Empty(t, r.RepoName, "primary repo should have no prefix")
		}
		if r.SymbolName == "authorize" {
			assert.Equal(t, "api-service", r.RepoName, "additional repo should have repo name")
		}
	}
}

func TestMultiStoreCloseAll(t *testing.T) {
	primary := openTestStore(t)
	additional1 := openTestStore(t)
	additional2 := openTestStore(t)

	ms := NewMultiStore(primary, []*Store{additional1, additional2}, []string{"a", "b"})
	err := ms.Close()
	assert.NoError(t, err)
}

func TestMultiStoreSingleStore(t *testing.T) {
	primary := openTestStore(t)
	seedSymbol(t, primary, "main.go", "bootstrap", "function", "go", "func bootstrap() start server")

	ms := NewMultiStore(primary, nil, nil)

	results, err := ms.SearchCode("bootstrap", "", "", 10)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "bootstrap", results[0].SymbolName)
	assert.Empty(t, results[0].RepoName)
}

func TestMultiStoreLanguageFilter(t *testing.T) {
	primary := openTestStore(t)
	seedSymbol(t, primary, "main.go", "handler", "function", "go", "func handler() process request")
	seedSymbol(t, primary, "main.py", "handler", "function", "python", "def handler() process request")

	ms := NewMultiStore(primary, nil, nil)

	results, err := ms.SearchCode("handler", "go", "", 10)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "go", results[0].Language)
}

func TestMultiStoreLimit(t *testing.T) {
	primary := openTestStore(t)
	for i := range 5 {
		seedSymbol(t, primary, fmt.Sprintf("file%d.go", i), "validate", "function", "go",
			fmt.Sprintf("func validate() check input %d", i))
	}

	ms := NewMultiStore(primary, nil, nil)

	results, err := ms.SearchCode("validate", "", "", 2)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMultiStorePrimaryAccessor(t *testing.T) {
	primary := openTestStore(t)
	ms := NewMultiStore(primary, nil, nil)
	assert.Equal(t, primary, ms.Primary())
}
