package storage

import (
	"fmt"
	"sort"
)

// MultiStore aggregates multiple Store instances for cross-repo search.
type MultiStore struct {
	primary    *Store
	additional []*Store
	names      []string // display names for additional repos
}

// SearchResult holds a search result with repo context.
type SearchResult struct {
	RepoName   string
	FilePath   string
	SymbolName string
	SymbolType string
	Language   string
	Signature  string
	StartLine  int
	EndLine    int
	Rank       float64
}

// NewMultiStore creates a MultiStore from a primary and optional additional stores.
func NewMultiStore(primary *Store, additional []*Store, names []string) *MultiStore {
	return &MultiStore{
		primary:    primary,
		additional: additional,
		names:      names,
	}
}

// Primary returns the primary store.
func (ms *MultiStore) Primary() *Store {
	return ms.primary
}

// SearchCode searches across all stores and merges results by relevance.
func (ms *MultiStore) SearchCode(query string, language string, symbolType string, limit int) ([]SearchResult, error) {
	var allResults []SearchResult

	// Search primary
	results, err := searchStore(ms.primary, "", query, language, symbolType)
	if err != nil {
		return nil, fmt.Errorf("searching primary store: %w", err)
	}
	allResults = append(allResults, results...)

	// Search additional stores
	for i, store := range ms.additional {
		name := ""
		if i < len(ms.names) {
			name = ms.names[i]
		}
		results, err := searchStore(store, name, query, language, symbolType)
		if err != nil {
			continue // skip failing additional repos
		}
		allResults = append(allResults, results...)
	}

	// Sort by rank (more negative = better match for FTS5)
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Rank < allResults[j].Rank
	})

	if limit > 0 && len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}

// Close closes all stores.
func (ms *MultiStore) Close() error {
	var firstErr error
	if err := ms.primary.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	for _, store := range ms.additional {
		if err := store.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func searchStore(store *Store, repoName string, query string, language string, symbolType string) ([]SearchResult, error) {
	sqlQuery := `SELECT ci.file_path, ci.symbol_name, ci.symbol_type, ci.language, ci.signature, ci.start_line, ci.end_line, rank
		FROM code_index_fts fts
		JOIN code_index ci ON ci.id = fts.rowid
		WHERE code_index_fts MATCH ?`
	args := []any{query}

	if language != "" {
		sqlQuery += " AND ci.language = ?"
		args = append(args, language)
	}
	if symbolType != "" {
		sqlQuery += " AND ci.symbol_type = ?"
		args = append(args, symbolType)
	}
	sqlQuery += " ORDER BY rank LIMIT 50"

	rows, err := store.DB().Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.FilePath, &r.SymbolName, &r.SymbolType, &r.Language, &r.Signature, &r.StartLine, &r.EndLine, &r.Rank); err != nil {
			continue
		}
		r.RepoName = repoName
		results = append(results, r)
	}
	return results, rows.Err()
}
