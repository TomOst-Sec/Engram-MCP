# TASK-006: ONNX Embedding Pipeline — Model Loading, Batch Inference, Vector Storage

**Priority:** P1
**Assigned:** alpha
**Milestone:** M1: MVP
**Dependencies:** TASK-003
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
Engram uses semantic embeddings to power similarity search across code symbols and memories. The embedding pipeline takes text (function signatures, docstrings, memory content) and produces 384-dimensional vectors using the all-MiniLM-L6-v2 model via ONNX Runtime. These vectors are stored in SQLite alongside the source data and used for cosine similarity search. The pipeline must be fully offline — no network calls ever.

## Specification
Create the `internal/embeddings` package implementing an ONNX-based embedding pipeline.

### Dependencies
- `github.com/yalue/onnxruntime_go` — Go bindings for ONNX Runtime

### Core Types and Functions
```go
// Embedder generates vector embeddings from text
type Embedder struct {
    session *ort.Session  // ONNX runtime session
    // tokenizer state
}

// New creates an Embedder, loading the ONNX model from the given path
// If modelPath is empty, uses the bundled default model
func New(modelPath string) (*Embedder, error)

// Close releases ONNX runtime resources
func (e *Embedder) Close() error

// Embed generates a 384-dimensional embedding vector for a single text input
func (e *Embedder) Embed(text string) ([]float32, error)

// EmbedBatch generates embeddings for multiple texts in a single batch
// Processes in chunks of batchSize (default 32) for memory efficiency
func (e *Embedder) EmbedBatch(texts []string, batchSize int) ([][]float32, error)

// CosineSimilarity computes similarity between two embedding vectors
func CosineSimilarity(a, b []float32) float32

// TopK finds the k most similar vectors to the query from a set of candidates
// Returns indices and scores, sorted by similarity descending
func TopK(query []float32, candidates [][]float32, k int) (indices []int, scores []float32)
```

### Tokenizer
The all-MiniLM-L6-v2 model uses a WordPiece tokenizer. For MVP, implement a **simplified tokenizer**:
1. Lowercase the input
2. Split on whitespace and punctuation
3. Truncate to 128 tokens (model max is 512, but 128 is sufficient for code symbols)
4. Pad to uniform length within a batch
5. Create attention mask (1 for real tokens, 0 for padding)

Note: This simplified tokenizer won't match HuggingFace's exact tokenization, but will produce usable embeddings for code search. A proper WordPiece tokenizer can be added later.

### Vector Storage Helpers
```go
// SerializeVector converts float32 slice to bytes for SQLite BLOB storage
func SerializeVector(v []float32) []byte

// DeserializeVector converts bytes from SQLite BLOB back to float32 slice
func DeserializeVector(b []byte) []float32

// UpdateCodeIndexEmbedding sets the embedding BLOB for a code_index row
func UpdateCodeIndexEmbedding(store *storage.Store, id int64, embedding []float32) error

// UpdateMemoryEmbedding sets the embedding BLOB for a memories row
func UpdateMemoryEmbedding(store *storage.Store, id int64, embedding []float32) error

// SearchByVector performs brute-force cosine similarity search against code_index embeddings
// Returns top-k results with scores
func SearchByVector(store *storage.Store, query []float32, k int, filters ...SearchFilter) ([]SearchResult, error)

type SearchResult struct {
    ID         int64
    FilePath   string
    SymbolName string
    SymbolType string
    Score      float32
    StartLine  int
    EndLine    int
}

type SearchFilter struct {
    Language   string
    SymbolType string
    Directory  string
}
```

### Graceful Degradation
If the ONNX model file is not found or ONNX Runtime fails to initialize:
- Log a warning
- Return a `NoOpEmbedder` that returns nil embeddings
- All embedding-dependent features (vector search) gracefully degrade to FTS5-only search
- The system continues to function without embeddings

## Acceptance Criteria
- [ ] Embedder loads the ONNX model (or returns NoOpEmbedder if unavailable)
- [ ] Embed() produces a 384-dimensional float32 vector for any text input
- [ ] EmbedBatch() processes multiple texts and returns correct number of vectors
- [ ] CosineSimilarity(v, v) returns ~1.0 (same vector = identical)
- [ ] CosineSimilarity returns higher score for semantically similar texts than dissimilar ones
- [ ] SerializeVector/DeserializeVector round-trips correctly
- [ ] SearchByVector returns results sorted by similarity score descending
- [ ] TopK returns correct top-k indices
- [ ] Graceful degradation: NoOpEmbedder works when model is missing
- [ ] All tests pass

## Implementation Steps
1. `go get github.com/yalue/onnxruntime_go`
2. Create `internal/embeddings/embedder.go` — Embedder struct, New, Embed, EmbedBatch, Close
3. Create `internal/embeddings/tokenizer.go` — simplified tokenizer (lowercase, split, truncate, pad, attention mask)
4. Create `internal/embeddings/similarity.go` — CosineSimilarity, TopK functions
5. Create `internal/embeddings/vector.go` — SerializeVector, DeserializeVector
6. Create `internal/embeddings/store.go` — UpdateCodeIndexEmbedding, UpdateMemoryEmbedding, SearchByVector
7. Create `internal/embeddings/noop.go` — NoOpEmbedder for graceful degradation
8. Create `internal/embeddings/embedder_test.go` — test Embed, EmbedBatch (skip if no model file available)
9. Create `internal/embeddings/similarity_test.go` — test CosineSimilarity, TopK
10. Create `internal/embeddings/vector_test.go` — test serialize/deserialize round-trip
11. Create `internal/embeddings/store_test.go` — test storage helpers with SQLite
12. Run all tests

## Testing Requirements
- Unit test: CosineSimilarity of identical vectors returns ~1.0
- Unit test: CosineSimilarity of orthogonal vectors returns ~0.0
- Unit test: TopK(query, candidates, 3) returns 3 results in descending score order
- Unit test: SerializeVector → DeserializeVector round-trip preserves values exactly
- Unit test: NoOpEmbedder.Embed returns nil without error
- Unit test: SearchByVector returns results sorted by score
- Conditional test: If ONNX model available, Embed() returns 384-dim vector (use `t.Skip` if no model)
- Conditional test: If model available, similar texts have higher cosine similarity than dissimilar texts

## Files to Create/Modify
- `internal/embeddings/embedder.go` — Embedder struct, New, Embed, EmbedBatch, Close
- `internal/embeddings/tokenizer.go` — simplified WordPiece tokenizer
- `internal/embeddings/similarity.go` — CosineSimilarity, TopK
- `internal/embeddings/vector.go` — vector serialization
- `internal/embeddings/store.go` — database helpers for embedding storage and search
- `internal/embeddings/noop.go` — NoOpEmbedder for degraded mode
- `internal/embeddings/embedder_test.go` — embedder tests
- `internal/embeddings/similarity_test.go` — similarity tests
- `internal/embeddings/vector_test.go` — serialization tests
- `internal/embeddings/store_test.go` — storage integration tests

## Notes
- The ONNX model file (all-MiniLM-L6-v2 quantized INT8, ~23MB) will be bundled later. For now, the model path is configurable. Tests that need the model should `t.Skip("ONNX model not available")` if the file is missing.
- The simplified tokenizer is a deliberate MVP choice. A proper WordPiece tokenizer with the model's vocabulary is a Milestone 2 enhancement.
- For vector storage, use `encoding/binary` with `binary.LittleEndian` to convert float32 slices to/from byte slices. Each float32 is 4 bytes, so a 384-dim vector = 1536 bytes.
- The brute-force search (scanning all vectors) is fine for repos < 10K symbols. HNSW indexing is a Milestone 2 feature.
- ONNX Runtime requires the shared library to be available. Tests may need to `t.Skip` on systems where it's not installed. Design the test suite so that similarity tests (pure math) always run, and ONNX-dependent tests are conditional.

---
## Completion Notes
- **Completed by:** alpha-3
- **Date:** 2026-03-15 15:12:47
- **Branch:** task/006

---
## Completion Notes
- **Completed by:** alpha-3
- **Date:** 2026-03-15 15:19:47
- **Branch:** task/006
