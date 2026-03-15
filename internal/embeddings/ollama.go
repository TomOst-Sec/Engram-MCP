package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaEmbedder generates embeddings using a local Ollama instance.
type OllamaEmbedder struct {
	endpoint string
	model    string
	client   *http.Client
}

// NewOllamaEmbedder creates a new Ollama-backed embedder.
func NewOllamaEmbedder(endpoint, model string) (*OllamaEmbedder, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("ollama endpoint is empty")
	}
	if model == "" {
		return nil, fmt.Errorf("ollama model is empty")
	}
	return &OllamaEmbedder{
		endpoint: endpoint,
		model:    model,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

type ollamaEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type ollamaEmbedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// Embed generates an embedding for a single text using Ollama.
func (o *OllamaEmbedder) Embed(text string) ([]float32, error) {
	results, err := o.callAPI([]string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("ollama returned empty embeddings")
	}
	return results[0], nil
}

// EmbedBatch generates embeddings for multiple texts in batches.
func (o *OllamaEmbedder) EmbedBatch(texts []string, batchSize int) ([][]float32, error) {
	if batchSize <= 0 {
		batchSize = 32
	}

	var allEmbeddings [][]float32
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]
		embeddings, err := o.callAPI(batch)
		if err != nil {
			return nil, fmt.Errorf("batch %d-%d failed: %w", i, end, err)
		}
		allEmbeddings = append(allEmbeddings, embeddings...)
	}
	return allEmbeddings, nil
}

// Close is a no-op for Ollama (HTTP client doesn't need cleanup).
func (o *OllamaEmbedder) Close() error {
	return nil
}

func (o *OllamaEmbedder) callAPI(input []string) ([][]float32, error) {
	reqBody := ollamaEmbedRequest{
		Model: o.model,
		Input: input,
	}
	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := o.endpoint + "/api/embed"
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Embeddings, nil
}
