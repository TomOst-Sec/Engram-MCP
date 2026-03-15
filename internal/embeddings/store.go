package embeddings

import (
	"database/sql"
	"fmt"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// SearchResult represents a single search result from vector similarity search.
type SearchResult struct {
	ID         int64
	FilePath   string
	SymbolName string
	SymbolType string
	Score      float32
	StartLine  int
	EndLine    int
}

// SearchFilter allows narrowing vector search results.
type SearchFilter struct {
	Language   string
	SymbolType string
	Directory  string
}

// UpdateCodeIndexEmbedding sets the embedding BLOB for a code_index row.
func UpdateCodeIndexEmbedding(store *storage.Store, id int64, embedding []float32) error {
	blob := SerializeVector(embedding)
	_, err := store.DB().Exec(
		"UPDATE code_index SET embedding = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		blob, id,
	)
	if err != nil {
		return fmt.Errorf("updating code_index embedding for id %d: %w", id, err)
	}
	return nil
}

// UpdateMemoryEmbedding sets the embedding BLOB for a memories row.
func UpdateMemoryEmbedding(store *storage.Store, id int64, embedding []float32) error {
	blob := SerializeVector(embedding)
	_, err := store.DB().Exec(
		"UPDATE memories SET embedding = ? WHERE id = ?",
		blob, id,
	)
	if err != nil {
		return fmt.Errorf("updating memory embedding for id %d: %w", id, err)
	}
	return nil
}

// SearchByVector performs brute-force cosine similarity search against code_index embeddings.
// Returns top-k results with scores, sorted by similarity descending.
func SearchByVector(store *storage.Store, query []float32, k int, filters ...SearchFilter) ([]SearchResult, error) {
	if len(query) == 0 || k <= 0 {
		return nil, nil
	}

	// Build the query with optional filters
	sqlQuery := `SELECT id, file_path, symbol_name, symbol_type, embedding, start_line, end_line
		FROM code_index WHERE embedding IS NOT NULL`
	var args []interface{}

	if len(filters) > 0 {
		f := filters[0]
		if f.Language != "" {
			sqlQuery += " AND language = ?"
			args = append(args, f.Language)
		}
		if f.SymbolType != "" {
			sqlQuery += " AND symbol_type = ?"
			args = append(args, f.SymbolType)
		}
		if f.Directory != "" {
			sqlQuery += " AND file_path LIKE ?"
			args = append(args, f.Directory+"%")
		}
	}

	rows, err := store.DB().Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("querying code_index: %w", err)
	}
	defer rows.Close()

	type candidate struct {
		result    SearchResult
		embedding []float32
	}

	var candidates []candidate
	for rows.Next() {
		var c candidate
		var embBlob []byte
		err := rows.Scan(
			&c.result.ID,
			&c.result.FilePath,
			&c.result.SymbolName,
			&c.result.SymbolType,
			&embBlob,
			&c.result.StartLine,
			&c.result.EndLine,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		c.embedding = DeserializeVector(embBlob)
		if c.embedding != nil {
			candidates = append(candidates, c)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	// Compute cosine similarities
	embeddings := make([][]float32, len(candidates))
	for i, c := range candidates {
		embeddings[i] = c.embedding
	}

	indices, scores := TopK(query, embeddings, k)

	results := make([]SearchResult, len(indices))
	for i, idx := range indices {
		results[i] = candidates[idx].result
		results[i].Score = scores[i]
	}

	return results, nil
}

// GetCodeIndexEmbedding retrieves the embedding for a code_index row.
// Returns nil if no embedding is stored.
func GetCodeIndexEmbedding(store *storage.Store, id int64) ([]float32, error) {
	var blob sql.NullString
	err := store.DB().QueryRow("SELECT embedding FROM code_index WHERE id = ?", id).Scan(&blob)
	if err != nil {
		return nil, fmt.Errorf("getting embedding for id %d: %w", id, err)
	}
	if !blob.Valid {
		return nil, nil
	}
	return DeserializeVector([]byte(blob.String)), nil
}
