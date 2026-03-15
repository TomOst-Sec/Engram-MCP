package embeddings

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_EmptyPath(t *testing.T) {
	embedder, err := New("")
	assert.Nil(t, embedder)
	assert.Error(t, err)
}

func TestNew_NonExistentModel(t *testing.T) {
	embedder, err := New("/nonexistent/model.onnx")
	assert.Nil(t, embedder)
	assert.NoError(t, err) // returns nil,nil for missing model (graceful degradation)
}

func TestNoOpEmbedder_Embed(t *testing.T) {
	noop := &NoOpEmbedder{}
	result, err := noop.Embed("some text")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestNoOpEmbedder_EmbedBatch(t *testing.T) {
	noop := &NoOpEmbedder{}
	results, err := noop.EmbedBatch([]string{"text1", "text2", "text3"}, 2)
	assert.NoError(t, err)
	require.Len(t, results, 3)
	for _, r := range results {
		assert.Nil(t, r)
	}
}

func TestNoOpEmbedder_Close(t *testing.T) {
	noop := &NoOpEmbedder{}
	assert.NoError(t, noop.Close())
}

func TestTokenize_Basic(t *testing.T) {
	tokens := tokenize("Hello World", 128)
	assert.Equal(t, []string{"hello", "world"}, tokens)
}

func TestTokenize_Punctuation(t *testing.T) {
	tokens := tokenize("func HandleRequest(w http.ResponseWriter)", 128)
	assert.Contains(t, tokens, "func")
	assert.Contains(t, tokens, "handlerequest")
	assert.Contains(t, tokens, "(")
}

func TestTokenize_Truncation(t *testing.T) {
	tokens := tokenize("a b c d e f g h i j", 5)
	assert.Len(t, tokens, 5)
}

func TestTokenize_Empty(t *testing.T) {
	tokens := tokenize("", 128)
	assert.Empty(t, tokens)
}

func TestHashToken_Deterministic(t *testing.T) {
	a := hashToken("hello")
	b := hashToken("hello")
	assert.Equal(t, a, b)
}

func TestHashToken_InRange(t *testing.T) {
	tokens := []string{"hello", "world", "func", "HandleRequest", "123", "!@#"}
	for _, tok := range tokens {
		id := hashToken(tok)
		assert.True(t, id >= 1000, "token ID should be >= 1000, got %d for %q", id, tok)
		assert.True(t, id < 30000, "token ID should be < 30000, got %d for %q", id, tok)
	}
}

// Conditional test: only run if ONNX model is available
func TestEmbedder_WithModel(t *testing.T) {
	modelPath := os.Getenv("ONNX_MODEL_PATH")
	if modelPath == "" {
		t.Skip("ONNX_MODEL_PATH not set — skipping ONNX model tests")
	}

	embedder, err := New(modelPath)
	require.NoError(t, err)
	if embedder == nil {
		t.Skip("ONNX Runtime not available — skipping")
	}
	defer embedder.Close()

	t.Run("single_embed_384dim", func(t *testing.T) {
		result, err := embedder.Embed("function HandleRequest processes HTTP requests")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result, 384)
	})

	t.Run("batch_embed", func(t *testing.T) {
		texts := []string{
			"function HandleRequest processes HTTP requests",
			"class UserModel stores user data",
			"import database connection pool",
		}
		results, err := embedder.EmbedBatch(texts, 2)
		require.NoError(t, err)
		require.Len(t, results, 3)
		for i, r := range results {
			assert.Len(t, r, 384, "result %d should have 384 dimensions", i)
		}
	})

	t.Run("similar_texts_higher_similarity", func(t *testing.T) {
		v1, err := embedder.Embed("authentication login handler")
		require.NoError(t, err)

		v2, err := embedder.Embed("user login authentication")
		require.NoError(t, err)

		v3, err := embedder.Embed("database schema migration")
		require.NoError(t, err)

		simSimilar := CosineSimilarity(v1, v2)
		simDifferent := CosineSimilarity(v1, v3)

		assert.True(t, simSimilar > simDifferent,
			"similar texts should have higher similarity (%f) than dissimilar (%f)",
			simSimilar, simDifferent)
	})
}
