package embeddings

import (
	"github.com/TomOst-Sec/colony-project/internal/config"
)

// NewFromConfig creates the appropriate embedder based on configuration.
// If config specifies "ollama" as the embedding model or provides an Ollama endpoint,
// it returns an OllamaEmbedder. Otherwise, it returns the default ONNX Embedder.
func NewFromConfig(cfg *config.Config) (EmbedderInterface, error) {
	if cfg.EmbeddingModel == "ollama" || (cfg.OllamaEndpoint != "" && cfg.OllamaEndpoint != "http://localhost:11434") {
		return NewOllamaEmbedder(cfg.OllamaEndpoint, cfg.OllamaModel)
	}
	// Default: ONNX bundled model
	emb, err := New(cfg.EmbeddingModel)
	if err != nil {
		return nil, err
	}
	return emb, nil
}
