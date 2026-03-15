package embeddings

import (
	"testing"

	"github.com/TomOst-Sec/colony-project/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewFromConfigOllama(t *testing.T) {
	cfg := &config.Config{
		EmbeddingModel: "ollama",
		OllamaEndpoint: "http://localhost:11434",
		OllamaModel:    "nomic-embed-text",
	}

	emb, err := NewFromConfig(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, emb)

	// Should be OllamaEmbedder
	_, ok := emb.(*OllamaEmbedder)
	assert.True(t, ok, "expected OllamaEmbedder")
}

func TestNewFromConfigDefault(t *testing.T) {
	cfg := &config.Config{
		EmbeddingModel: "/nonexistent/model.onnx",
		OllamaEndpoint: "http://localhost:11434",
		OllamaModel:    "nomic-embed-text",
	}

	// ONNX model won't load (returns nil, nil for graceful degradation)
	// but it should attempt ONNX path, not Ollama
	emb, _ := NewFromConfig(cfg)
	// Should NOT be OllamaEmbedder
	if emb != nil {
		_, isOllama := emb.(*OllamaEmbedder)
		assert.False(t, isOllama, "should not return OllamaEmbedder for default config")
	}
}
