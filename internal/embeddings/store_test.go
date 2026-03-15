package embeddings

import (
	"path/filepath"
	"testing"

	"github.com/TomOst-Sec/colony-project/internal/storage"
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

func insertTestSymbol(t *testing.T, store *storage.Store, filePath, name, symType, lang string, startLine, endLine int) int64 {
	t.Helper()
	result, err := store.DB().Exec(
		`INSERT INTO code_index (file_path, file_hash, language, symbol_name, symbol_type, signature, start_line, end_line)
		 VALUES (?, 'testhash', ?, ?, ?, ?, ?, ?)`,
		filePath, lang, name, symType, "func "+name+"()", startLine, endLine,
	)
	require.NoError(t, err)
	id, err := result.LastInsertId()
	require.NoError(t, err)
	return id
}

func TestUpdateCodeIndexEmbedding(t *testing.T) {
	store := setupTestStore(t)

	id := insertTestSymbol(t, store, "main.go", "HandleRequest", "function", "go", 10, 20)

	embedding := make([]float32, 384)
	for i := range embedding {
		embedding[i] = float32(i) * 0.001
	}

	err := UpdateCodeIndexEmbedding(store, id, embedding)
	require.NoError(t, err)

	// Verify it was stored
	var blob []byte
	err = store.DB().QueryRow("SELECT embedding FROM code_index WHERE id = ?", id).Scan(&blob)
	require.NoError(t, err)
	require.NotNil(t, blob)
	assert.Len(t, blob, 384*4)

	recovered := DeserializeVector(blob)
	require.Len(t, recovered, 384)
	for i := range embedding {
		assert.Equal(t, embedding[i], recovered[i])
	}
}

func TestUpdateMemoryEmbedding(t *testing.T) {
	store := setupTestStore(t)

	result, err := store.DB().Exec(
		`INSERT INTO memories (content, type) VALUES (?, ?)`,
		"test memory content", "decision",
	)
	require.NoError(t, err)
	id, err := result.LastInsertId()
	require.NoError(t, err)

	embedding := []float32{0.1, 0.2, 0.3}
	err = UpdateMemoryEmbedding(store, id, embedding)
	require.NoError(t, err)

	var blob []byte
	err = store.DB().QueryRow("SELECT embedding FROM memories WHERE id = ?", id).Scan(&blob)
	require.NoError(t, err)
	require.NotNil(t, blob)

	recovered := DeserializeVector(blob)
	assert.Equal(t, embedding, recovered)
}

func TestSearchByVector_Basic(t *testing.T) {
	store := setupTestStore(t)

	// Insert symbols with embeddings
	id1 := insertTestSymbol(t, store, "auth/handler.go", "HandleLogin", "function", "go", 10, 30)
	id2 := insertTestSymbol(t, store, "auth/middleware.go", "AuthMiddleware", "function", "go", 5, 15)
	id3 := insertTestSymbol(t, store, "db/store.go", "Connect", "function", "go", 1, 10)

	// Create embeddings that are progressively less similar to query
	query := make([]float32, 384)
	query[0] = 1.0

	emb1 := make([]float32, 384) // very similar to query
	emb1[0] = 0.9
	emb1[1] = 0.1

	emb2 := make([]float32, 384) // somewhat similar
	emb2[0] = 0.5
	emb2[1] = 0.5

	emb3 := make([]float32, 384) // least similar
	emb3[0] = 0.1
	emb3[1] = 0.9

	require.NoError(t, UpdateCodeIndexEmbedding(store, id1, emb1))
	require.NoError(t, UpdateCodeIndexEmbedding(store, id2, emb2))
	require.NoError(t, UpdateCodeIndexEmbedding(store, id3, emb3))

	results, err := SearchByVector(store, query, 3)
	require.NoError(t, err)
	require.Len(t, results, 3)

	// Results should be sorted by score descending
	assert.Equal(t, "HandleLogin", results[0].SymbolName)
	assert.True(t, results[0].Score >= results[1].Score)
	assert.True(t, results[1].Score >= results[2].Score)
}

func TestSearchByVector_WithLanguageFilter(t *testing.T) {
	store := setupTestStore(t)

	id1 := insertTestSymbol(t, store, "main.go", "GoFunc", "function", "go", 1, 10)
	id2 := insertTestSymbol(t, store, "main.py", "PyFunc", "function", "python", 1, 10)

	emb := make([]float32, 384)
	emb[0] = 1.0
	require.NoError(t, UpdateCodeIndexEmbedding(store, id1, emb))
	require.NoError(t, UpdateCodeIndexEmbedding(store, id2, emb))

	query := make([]float32, 384)
	query[0] = 1.0

	results, err := SearchByVector(store, query, 10, SearchFilter{Language: "go"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "GoFunc", results[0].SymbolName)
}

func TestSearchByVector_WithSymbolTypeFilter(t *testing.T) {
	store := setupTestStore(t)

	id1 := insertTestSymbol(t, store, "main.go", "MyFunc", "function", "go", 1, 10)
	id2 := insertTestSymbol(t, store, "types.go", "MyType", "type", "go", 1, 10)

	emb := make([]float32, 384)
	emb[0] = 1.0
	require.NoError(t, UpdateCodeIndexEmbedding(store, id1, emb))
	require.NoError(t, UpdateCodeIndexEmbedding(store, id2, emb))

	query := make([]float32, 384)
	query[0] = 1.0

	results, err := SearchByVector(store, query, 10, SearchFilter{SymbolType: "function"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "MyFunc", results[0].SymbolName)
}

func TestSearchByVector_WithDirectoryFilter(t *testing.T) {
	store := setupTestStore(t)

	id1 := insertTestSymbol(t, store, "internal/auth/handler.go", "HandleLogin", "function", "go", 1, 10)
	id2 := insertTestSymbol(t, store, "internal/db/store.go", "Connect", "function", "go", 1, 10)

	emb := make([]float32, 384)
	emb[0] = 1.0
	require.NoError(t, UpdateCodeIndexEmbedding(store, id1, emb))
	require.NoError(t, UpdateCodeIndexEmbedding(store, id2, emb))

	query := make([]float32, 384)
	query[0] = 1.0

	results, err := SearchByVector(store, query, 10, SearchFilter{Directory: "internal/auth/"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "HandleLogin", results[0].SymbolName)
}

func TestSearchByVector_EmptyQuery(t *testing.T) {
	store := setupTestStore(t)
	results, err := SearchByVector(store, nil, 10)
	assert.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchByVector_NoEmbeddings(t *testing.T) {
	store := setupTestStore(t)

	// Insert symbol without embedding
	insertTestSymbol(t, store, "main.go", "Func", "function", "go", 1, 10)

	query := make([]float32, 384)
	query[0] = 1.0

	results, err := SearchByVector(store, query, 10)
	assert.NoError(t, err)
	assert.Nil(t, results)
}

func TestSearchByVector_ResultsSortedDescending(t *testing.T) {
	store := setupTestStore(t)

	// Insert 5 symbols with progressively less similar embeddings
	for i := 0; i < 5; i++ {
		id := insertTestSymbol(t, store, "file.go", "Func"+string(rune('A'+i)), "function", "go", i*10, (i+1)*10)
		emb := make([]float32, 384)
		emb[0] = float32(5-i) * 0.2 // decreasing similarity
		emb[1] = float32(i) * 0.2   // increasing orthogonal component
		require.NoError(t, UpdateCodeIndexEmbedding(store, id, emb))
	}

	query := make([]float32, 384)
	query[0] = 1.0

	results, err := SearchByVector(store, query, 5)
	require.NoError(t, err)
	require.Len(t, results, 5)

	for i := 1; i < len(results); i++ {
		assert.True(t, results[i-1].Score >= results[i].Score,
			"results should be descending: %f >= %f at positions %d,%d",
			results[i-1].Score, results[i].Score, i-1, i)
	}
}

func TestGetCodeIndexEmbedding(t *testing.T) {
	store := setupTestStore(t)

	id := insertTestSymbol(t, store, "main.go", "Func", "function", "go", 1, 10)

	// Initially nil
	emb, err := GetCodeIndexEmbedding(store, id)
	require.NoError(t, err)
	assert.Nil(t, emb)

	// Set embedding
	expected := []float32{1.0, 2.0, 3.0}
	require.NoError(t, UpdateCodeIndexEmbedding(store, id, expected))

	emb, err = GetCodeIndexEmbedding(store, id)
	require.NoError(t, err)
	assert.Equal(t, expected, emb)
}
