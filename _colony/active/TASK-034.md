# TASK-034: Ollama Integration — Optional Local LLM Embeddings

**Priority:** P2
**Assigned:** bravo
**Milestone:** M3: Polish & Growth
**Dependencies:** TASK-006
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
Feature 18 from GOALS.md. Engram currently uses a bundled ONNX model (all-MiniLM-L6-v2) for embeddings. Some users may want higher-quality embeddings using a local Ollama instance. This task adds Ollama as an alternative embedding backend. Config already has `ollama_endpoint` and `ollama_model` fields defined but unused.

## Specification

### Ollama Embedder: `internal/embeddings/ollama.go`

```go
type OllamaEmbedder struct {
    endpoint string  // e.g., "http://localhost:11434"
    model    string  // e.g., "nomic-embed-text"
    client   *http.Client
}

func NewOllamaEmbedder(endpoint, model string) (*OllamaEmbedder, error)

// Embed generates an embedding for a single text using Ollama.
func (o *OllamaEmbedder) Embed(text string) ([]float32, error)

// EmbedBatch generates embeddings for multiple texts.
func (o *OllamaEmbedder) EmbedBatch(texts []string, batchSize int) ([][]float32, error)

// Close is a no-op for Ollama (HTTP client doesn't need cleanup).
func (o *OllamaEmbedder) Close() error
```

### Ollama API
Call the Ollama embeddings API:

```
POST http://localhost:11434/api/embed
Content-Type: application/json

{
  "model": "nomic-embed-text",
  "input": ["text to embed"]
}

Response:
{
  "embeddings": [[0.1, 0.2, ...]]
}
```

### Embedder Interface

Refactor to use an interface so both ONNX and Ollama can be used interchangeably:

```go
// internal/embeddings/embedder.go
type EmbedderInterface interface {
    Embed(text string) ([]float32, error)
    EmbedBatch(texts []string, batchSize int) ([][]float32, error)
    Close() error
}
```

Both `Embedder` (ONNX) and `OllamaEmbedder` implement this interface.

### Factory Function

```go
// NewFromConfig creates the appropriate embedder based on config.
func NewFromConfig(cfg *config.Config) (EmbedderInterface, error) {
    if cfg.EmbeddingModel == "ollama" || cfg.OllamaEndpoint != "" {
        return NewOllamaEmbedder(cfg.OllamaEndpoint, cfg.OllamaModel)
    }
    // Default: ONNX bundled model
    return New(cfg.EmbeddingModel)
}
```

### Configuration

Already in config struct:
```go
OllamaEndpoint string  // "http://localhost:11434"
OllamaModel    string  // "nomic-embed-text"
```

Usage: set `embedding_model: "ollama"` in engram.json, or:
```bash
engram serve --embedding-model ollama
```

### Dimension Handling
- ONNX model: 384 dimensions (fixed)
- Ollama models: varies (nomic-embed-text = 768 dims)
- Store dimension count alongside vectors
- Search must handle different dimensions (cosine similarity works with any dimension)

## Acceptance Criteria
- [ ] `OllamaEmbedder` calls Ollama API and returns embeddings
- [ ] `OllamaEmbedder` handles batch requests
- [ ] `EmbedderInterface` is implemented by both ONNX and Ollama embedders
- [ ] `NewFromConfig` selects correct embedder based on config
- [ ] Ollama connection failure returns clear error message
- [ ] Ollama timeout (5s) prevents hanging
- [ ] Existing ONNX embedder still works unchanged
- [ ] All tests pass

## Implementation Steps
1. Create `internal/embeddings/interface.go` — EmbedderInterface definition
2. Verify existing Embedder satisfies EmbedderInterface
3. Create `internal/embeddings/ollama.go` — OllamaEmbedder implementation
4. Create `internal/embeddings/factory.go` — NewFromConfig factory
5. Create `internal/embeddings/ollama_test.go`:
   - Test: NewOllamaEmbedder with invalid endpoint returns error
   - Test: Embed with mock HTTP server returns correct vector
   - Test: EmbedBatch with mock server returns correct vectors
   - Test: Timeout handling (slow server)
6. Create `internal/embeddings/factory_test.go`:
   - Test: NewFromConfig with ollama config returns OllamaEmbedder
   - Test: NewFromConfig with default config returns ONNX Embedder
7. Run all tests

## Testing Requirements
- Unit test: OllamaEmbedder constructs with valid endpoint
- Unit test: Embed calls correct API endpoint with correct payload (use httptest)
- Unit test: EmbedBatch processes multiple texts
- Unit test: HTTP errors return meaningful error messages
- Unit test: Factory selects correct embedder type
- Regression test: existing ONNX embedder tests pass

## Files to Create/Modify
- `internal/embeddings/interface.go` — EmbedderInterface (create new)
- `internal/embeddings/ollama.go` — Ollama embedder (create new)
- `internal/embeddings/factory.go` — embedder factory (create new)
- `internal/embeddings/ollama_test.go` — Ollama tests (create new)
- `internal/embeddings/factory_test.go` — factory tests (create new)

## Notes
- Use `net/http/httptest` for testing — create a mock Ollama server that returns known embeddings.
- The Ollama API at `/api/embed` accepts an `input` field with an array of strings and returns `embeddings` as an array of arrays.
- Set a 5-second timeout on the HTTP client to prevent hanging if Ollama is down.
- Do NOT modify the existing `Embedder` struct (ONNX). Only add new files and an interface.
- The dimension mismatch between ONNX (384) and Ollama models (varies) means you should NOT hardcode 384 dims anywhere in the search/similarity code. Check that `internal/embeddings/similarity.go` works with arbitrary dimensions.
- If updating any callers from `*Embedder` to `EmbedderInterface`, make sure to update serve.go and index.go type references. But prefer not changing callers in this task — just add the interface and Ollama implementation.
