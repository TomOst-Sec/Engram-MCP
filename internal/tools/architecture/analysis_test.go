package architecture

import (
	"path/filepath"
	"testing"

	"github.com/TomOst-Sec/colony-project/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *storage.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	return store
}

func insertSymbol(t *testing.T, store *storage.Store, filePath, lang, name, symType string, startLine, endLine int) {
	t.Helper()
	_, err := store.DB().Exec(`
		INSERT INTO code_index (file_path, file_hash, language, symbol_name, symbol_type, start_line, end_line)
		VALUES (?, 'hash', ?, ?, ?, ?, ?)
	`, filePath, lang, name, symType, startLine, endLine)
	require.NoError(t, err)
}

func insertImport(t *testing.T, store *storage.Store, filePath, lang, importPath string) {
	t.Helper()
	insertSymbol(t, store, filePath, lang, importPath, "import", 1, 1)
}

func TestDetectModulesGroupsCorrectly(t *testing.T) {
	store := setupTestDB(t)

	// Insert symbols across 3 modules
	insertSymbol(t, store, "internal/auth/handler.go", "go", "HandleLogin", "function", 10, 30)
	insertSymbol(t, store, "internal/auth/middleware.go", "go", "AuthMiddleware", "function", 5, 20)
	insertSymbol(t, store, "internal/storage/store.go", "go", "Open", "function", 10, 50)
	insertSymbol(t, store, "internal/storage/store.go", "go", "Close", "function", 52, 60)
	insertSymbol(t, store, "internal/mcp/server.go", "go", "New", "function", 5, 15)

	// 3 files across 3 directories at depth 2
	modules, err := DetectModules(store, 2)
	require.NoError(t, err)
	assert.Len(t, modules, 3)

	// Verify module data
	moduleMap := make(map[string]Module)
	for _, m := range modules {
		moduleMap[m.Path] = m
	}

	assert.Contains(t, moduleMap, "internal/auth")
	assert.Equal(t, 2, moduleMap["internal/auth"].Files)
	assert.Equal(t, 2, moduleMap["internal/auth"].Symbols)

	assert.Contains(t, moduleMap, "internal/storage")
	assert.Equal(t, 1, moduleMap["internal/storage"].Files)
	assert.Equal(t, 2, moduleMap["internal/storage"].Symbols)

	assert.Contains(t, moduleMap, "internal/mcp")
	assert.Equal(t, 1, moduleMap["internal/mcp"].Files)
}

func TestDetectModulesDepth1(t *testing.T) {
	store := setupTestDB(t)

	insertSymbol(t, store, "internal/auth/handler.go", "go", "HandleLogin", "function", 10, 30)
	insertSymbol(t, store, "internal/storage/store.go", "go", "Open", "function", 10, 50)
	insertSymbol(t, store, "cmd/engram/main.go", "go", "main", "function", 1, 10)

	modules, err := DetectModules(store, 1)
	require.NoError(t, err)
	assert.Len(t, modules, 2) // "internal" and "cmd"

	moduleMap := make(map[string]Module)
	for _, m := range modules {
		moduleMap[m.Path] = m
	}
	assert.Contains(t, moduleMap, "internal")
	assert.Contains(t, moduleMap, "cmd")
}

func TestBuildDependencyGraph(t *testing.T) {
	store := setupTestDB(t)

	// Set up modules
	insertSymbol(t, store, "internal/auth/handler.go", "go", "HandleLogin", "function", 10, 30)
	insertSymbol(t, store, "internal/storage/store.go", "go", "Open", "function", 10, 50)
	insertSymbol(t, store, "internal/config/config.go", "go", "Load", "function", 5, 20)

	// Add imports: auth imports storage and config
	insertImport(t, store, "internal/auth/handler.go", "go", "github.com/TomOst-Sec/colony-project/internal/storage")
	insertImport(t, store, "internal/auth/handler.go", "go", "github.com/TomOst-Sec/colony-project/internal/config")

	// Add external import
	insertImport(t, store, "internal/auth/handler.go", "go", "github.com/spf13/cobra")

	modules, err := DetectModules(store, 2)
	require.NoError(t, err)

	internalDeps, externalDeps, err := BuildDependencyGraph(store, modules, "github.com/TomOst-Sec/colony-project")
	require.NoError(t, err)

	// auth should depend on storage and config
	assert.Contains(t, internalDeps["internal/auth"], "internal/storage")
	assert.Contains(t, internalDeps["internal/auth"], "internal/config")

	// auth should have external dep on cobra
	assert.Contains(t, externalDeps["internal/auth"], "github.com/spf13/cobra")
}

func TestBuildDependencyGraphSeparatesInternalExternal(t *testing.T) {
	store := setupTestDB(t)

	insertSymbol(t, store, "internal/mcp/server.go", "go", "New", "function", 5, 15)
	insertImport(t, store, "internal/mcp/server.go", "go", "github.com/mark3labs/mcp-go/server")
	insertImport(t, store, "internal/mcp/server.go", "go", "github.com/TomOst-Sec/colony-project/internal/storage")

	insertSymbol(t, store, "internal/storage/store.go", "go", "Open", "function", 10, 50)

	modules, err := DetectModules(store, 2)
	require.NoError(t, err)

	internalDeps, externalDeps, err := BuildDependencyGraph(store, modules, "github.com/TomOst-Sec/colony-project")
	require.NoError(t, err)

	assert.Contains(t, internalDeps["internal/mcp"], "internal/storage")
	assert.Contains(t, externalDeps["internal/mcp"], "github.com/mark3labs/mcp-go/server")
}

func TestComputeComplexityInRange(t *testing.T) {
	score := ComputeComplexity(10, 5, 3)
	assert.GreaterOrEqual(t, score, 0.0)
	assert.LessOrEqual(t, score, 10.0)

	// 10*0.5 + 5*0.3 + 3*0.2 = 5.0 + 1.5 + 0.6 = 7.1
	assert.InDelta(t, 7.1, score, 0.01)
}

func TestComputeComplexityCapsAt10(t *testing.T) {
	score := ComputeComplexity(100, 50, 30)
	assert.Equal(t, 10.0, score)
}

func TestComputeComplexityZero(t *testing.T) {
	score := ComputeComplexity(0, 0, 0)
	assert.Equal(t, 0.0, score)
}

func TestDetectModulesEmptyDB(t *testing.T) {
	store := setupTestDB(t)

	modules, err := DetectModules(store, 2)
	require.NoError(t, err)
	assert.Empty(t, modules)
}

func TestFindDependents(t *testing.T) {
	depGraph := map[string][]string{
		"internal/auth":    {"internal/storage", "internal/config"},
		"internal/mcp":     {"internal/tools", "internal/config"},
		"internal/tools":   {"internal/storage"},
	}

	dependents := FindDependents("internal/storage", depGraph)
	assert.Contains(t, dependents, "internal/auth")
	assert.Contains(t, dependents, "internal/tools")
	assert.Len(t, dependents, 2)
}

func TestIsExported(t *testing.T) {
	assert.True(t, isExported("HandleLogin", "go"))
	assert.False(t, isExported("handleLogin", "go"))
	assert.True(t, isExported("my_func", "python"))
	assert.False(t, isExported("_private", "python"))
	assert.True(t, isExported("handler", "typescript"))
}
