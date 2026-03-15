package embeddings

// EmbedderInterface defines the contract for embedding text into vectors.
// Both the ONNX Embedder and OllamaEmbedder implement this interface.
type EmbedderInterface interface {
	// Embed generates an embedding vector for a single text.
	Embed(text string) ([]float32, error)

	// EmbedBatch generates embeddings for multiple texts in batches.
	EmbedBatch(texts []string, batchSize int) ([][]float32, error)

	// Close releases any resources held by the embedder.
	Close() error
}
