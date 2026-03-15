package embeddings

// NoOpEmbedder is a fallback embedder that returns nil embeddings.
// Used when the ONNX model is unavailable, allowing the system to
// gracefully degrade to FTS5-only search.
type NoOpEmbedder struct{}

// Embed returns nil without error.
func (e *NoOpEmbedder) Embed(text string) ([]float32, error) {
	return nil, nil
}

// EmbedBatch returns nil slices for each input without error.
func (e *NoOpEmbedder) EmbedBatch(texts []string, batchSize int) ([][]float32, error) {
	result := make([][]float32, len(texts))
	return result, nil
}

// Close is a no-op.
func (e *NoOpEmbedder) Close() error {
	return nil
}
