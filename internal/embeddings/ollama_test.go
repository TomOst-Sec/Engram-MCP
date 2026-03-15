package embeddings

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOllamaEmbedder(t *testing.T) {
	emb, err := NewOllamaEmbedder("http://localhost:11434", "nomic-embed-text")
	require.NoError(t, err)
	assert.NotNil(t, emb)
}

func TestNewOllamaEmbedderEmptyEndpoint(t *testing.T) {
	_, err := NewOllamaEmbedder("", "nomic-embed-text")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint is empty")
}

func TestNewOllamaEmbedderEmptyModel(t *testing.T) {
	_, err := NewOllamaEmbedder("http://localhost:11434", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model is empty")
}

func TestOllamaEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/embed", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req ollamaEmbedRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "nomic-embed-text", req.Model)
		assert.Len(t, req.Input, 1)

		resp := ollamaEmbedResponse{
			Embeddings: [][]float32{{0.1, 0.2, 0.3, 0.4}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	emb, err := NewOllamaEmbedder(server.URL, "nomic-embed-text")
	require.NoError(t, err)

	vec, err := emb.Embed("test text")
	require.NoError(t, err)
	assert.Len(t, vec, 4)
	assert.InDelta(t, 0.1, float64(vec[0]), 0.001)
}

func TestOllamaEmbedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ollamaEmbedRequest
		json.NewDecoder(r.Body).Decode(&req)

		embeddings := make([][]float32, len(req.Input))
		for i := range req.Input {
			embeddings[i] = []float32{float32(i) * 0.1, float32(i) * 0.2}
		}

		resp := ollamaEmbedResponse{Embeddings: embeddings}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	emb, err := NewOllamaEmbedder(server.URL, "nomic-embed-text")
	require.NoError(t, err)

	texts := []string{"text1", "text2", "text3"}
	results, err := emb.EmbedBatch(texts, 2)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestOllamaEmbedHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("model not found"))
	}))
	defer server.Close()

	emb, err := NewOllamaEmbedder(server.URL, "nomic-embed-text")
	require.NoError(t, err)

	_, err = emb.Embed("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 500")
}

func TestOllamaEmbedConnectionError(t *testing.T) {
	emb, err := NewOllamaEmbedder("http://localhost:1", "nomic-embed-text")
	require.NoError(t, err)

	_, err = emb.Embed("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request failed")
}

func TestOllamaClose(t *testing.T) {
	emb, err := NewOllamaEmbedder("http://localhost:11434", "nomic-embed-text")
	require.NoError(t, err)
	assert.NoError(t, emb.Close())
}

func TestOllamaImplementsInterface(t *testing.T) {
	emb, err := NewOllamaEmbedder("http://localhost:11434", "nomic-embed-text")
	require.NoError(t, err)

	// Compile-time check that OllamaEmbedder implements EmbedderInterface
	var _ EmbedderInterface = emb
}
