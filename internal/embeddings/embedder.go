package embeddings

import (
	"fmt"
	"log"
	"os"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

const (
	embeddingDim     = 384
	defaultBatchSize = 32
)

// Embedder generates vector embeddings from text using an ONNX model.
type Embedder struct {
	modelPath string
	session   *ort.DynamicAdvancedSession
	mu        sync.Mutex
}

// New creates an Embedder, loading the ONNX model from the given path.
// If the model file is not found or ONNX Runtime fails to initialize,
// it logs a warning and returns a nil Embedder. Callers should check
// for nil and use NoOpEmbedder as a fallback.
func New(modelPath string) (*Embedder, error) {
	if modelPath == "" {
		return nil, fmt.Errorf("model path is empty")
	}

	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		log.Printf("WARNING: ONNX model not found at %s — embeddings disabled", modelPath)
		return nil, nil
	}

	// Initialize ONNX Runtime shared library
	ort.SetSharedLibraryPath(findOrtLibrary())
	if err := ort.InitializeEnvironment(); err != nil {
		// May already be initialized — that's OK
		log.Printf("WARNING: ONNX Runtime init: %v", err)
	}

	// Create a dynamic session for variable input sizes
	session, err := ort.NewDynamicAdvancedSession(
		modelPath,
		[]string{"input_ids", "attention_mask", "token_type_ids"},
		[]string{"last_hidden_state"},
		nil,
	)
	if err != nil {
		log.Printf("WARNING: Failed to create ONNX session: %v — embeddings disabled", err)
		return nil, nil
	}

	return &Embedder{
		modelPath: modelPath,
		session:   session,
	}, nil
}

// Embed generates a 384-dimensional embedding vector for a single text input.
func (e *Embedder) Embed(text string) ([]float32, error) {
	results, err := e.EmbedBatch([]string{text}, 1)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no results returned")
	}
	return results[0], nil
}

// EmbedBatch generates embeddings for multiple texts.
// Processes in chunks of batchSize for memory efficiency.
func (e *Embedder) EmbedBatch(texts []string, batchSize int) ([][]float32, error) {
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	results := make([][]float32, 0, len(texts))

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]

		batchResults, err := e.embedBatchInternal(batch)
		if err != nil {
			return nil, fmt.Errorf("batch %d: %w", i/batchSize, err)
		}
		results = append(results, batchResults...)
	}

	return results, nil
}

// Close releases ONNX runtime resources.
func (e *Embedder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.session != nil {
		e.session.Destroy()
		e.session = nil
	}
	return nil
}

func (e *Embedder) embedBatchInternal(texts []string) ([][]float32, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.session == nil {
		return nil, fmt.Errorf("session is closed")
	}

	batchSize := len(texts)

	// Tokenize all texts
	maxLen := 0
	tokenized := make([][]string, batchSize)
	for i, text := range texts {
		tokenized[i] = tokenize(text, defaultMaxTokens)
		if len(tokenized[i]) > maxLen {
			maxLen = len(tokenized[i])
		}
	}
	if maxLen == 0 {
		maxLen = 1 // at least 1 token
	}
	// Add 2 for [CLS] and [SEP] tokens
	seqLen := maxLen + 2

	// Build input tensors (simplified token IDs using hash-based mapping)
	inputIDs := make([]int64, batchSize*seqLen)
	attentionMask := make([]int64, batchSize*seqLen)
	tokenTypeIDs := make([]int64, batchSize*seqLen)

	for i := 0; i < batchSize; i++ {
		offset := i * seqLen
		// [CLS] token = 101
		inputIDs[offset] = 101
		attentionMask[offset] = 1
		tokenTypeIDs[offset] = 0

		for j, token := range tokenized[i] {
			pos := offset + j + 1
			inputIDs[pos] = hashToken(token)
			attentionMask[pos] = 1
			tokenTypeIDs[pos] = 0
		}

		// [SEP] token = 102
		sepPos := offset + len(tokenized[i]) + 1
		inputIDs[sepPos] = 102
		attentionMask[sepPos] = 1
		tokenTypeIDs[sepPos] = 0
	}

	// Create ONNX tensors
	shape := ort.NewShape(int64(batchSize), int64(seqLen))

	inputIDsTensor, err := ort.NewTensor(shape, inputIDs)
	if err != nil {
		return nil, fmt.Errorf("creating input_ids tensor: %w", err)
	}
	defer inputIDsTensor.Destroy()

	attentionMaskTensor, err := ort.NewTensor(shape, attentionMask)
	if err != nil {
		return nil, fmt.Errorf("creating attention_mask tensor: %w", err)
	}
	defer attentionMaskTensor.Destroy()

	tokenTypeIDsTensor, err := ort.NewTensor(shape, tokenTypeIDs)
	if err != nil {
		return nil, fmt.Errorf("creating token_type_ids tensor: %w", err)
	}
	defer tokenTypeIDsTensor.Destroy()

	// Run inference — outputs auto-allocated by passing nil
	outputs := []ort.Value{nil}
	err = e.session.Run(
		[]ort.Value{inputIDsTensor, attentionMaskTensor, tokenTypeIDsTensor},
		outputs,
	)
	if err != nil {
		return nil, fmt.Errorf("ONNX inference: %w", err)
	}
	defer func() {
		for _, o := range outputs {
			if o != nil {
				o.Destroy()
			}
		}
	}()

	if outputs[0] == nil {
		return nil, fmt.Errorf("no output tensor returned")
	}

	// Extract the output — shape [batch, seqLen, 384]
	// We use mean pooling over the sequence dimension
	outputTensor, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return nil, fmt.Errorf("unexpected output tensor type")
	}

	outputData := outputTensor.GetData()
	results := make([][]float32, batchSize)

	for i := 0; i < batchSize; i++ {
		embedding := make([]float32, embeddingDim)
		// Mean pooling: average over sequence positions (only attended positions)
		var count float32
		for j := 0; j < seqLen; j++ {
			if attentionMask[i*seqLen+j] == 0 {
				continue
			}
			count++
			for d := 0; d < embeddingDim; d++ {
				embedding[d] += outputData[i*seqLen*embeddingDim+j*embeddingDim+d]
			}
		}
		if count > 0 {
			for d := 0; d < embeddingDim; d++ {
				embedding[d] /= count
			}
		}
		results[i] = embedding
	}

	return results, nil
}

// hashToken provides a simplified token-to-ID mapping using string hashing.
// This is a placeholder — a proper WordPiece tokenizer uses a vocabulary file.
func hashToken(token string) int64 {
	var h int64 = 0
	for _, c := range token {
		h = h*31 + int64(c)
	}
	// Map to vocab range [1000, 30000) — avoid special token IDs
	h = (h % 29000)
	if h < 0 {
		h += 29000
	}
	return h + 1000
}

// findOrtLibrary attempts to find the ONNX Runtime shared library.
func findOrtLibrary() string {
	candidates := []string{
		"libonnxruntime.so",
		"/usr/lib/libonnxruntime.so",
		"/usr/local/lib/libonnxruntime.so",
		"/usr/lib/x86_64-linux-gnu/libonnxruntime.so",
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return "libonnxruntime.so" // fallback to LD_LIBRARY_PATH search
}
