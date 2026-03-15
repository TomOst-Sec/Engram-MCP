package search

// FTSResult represents a single FTS5 search result.
type FTSResult struct {
	ID         int64
	FilePath   string
	SymbolName string
	SymbolType string
	Language   string
	Signature  string
	StartLine  int
	EndLine    int
	Rank       float64 // FTS5 rank (more negative = better match)
}

// VectorResult represents a single vector similarity search result.
type VectorResult struct {
	ID         int64
	FilePath   string
	SymbolName string
	SymbolType string
	Language   string
	Signature  string
	StartLine  int
	EndLine    int
	Score      float32 // cosine similarity 0..1
}

// RankedResult represents a final ranked search result after hybrid merging.
type RankedResult struct {
	ID         int64
	FilePath   string
	SymbolName string
	SymbolType string
	Language   string
	Signature  string
	StartLine  int
	EndLine    int
	Score      float64
}

// NormalizeFTSScores normalizes FTS5 rank values to 0.0–1.0 range.
// FTS5 rank values are negative (more negative = better).
func NormalizeFTSScores(results []FTSResult) []FTSResult {
	if len(results) == 0 {
		return results
	}

	// Find min and max rank values
	minRank := results[0].Rank
	maxRank := results[0].Rank
	for _, r := range results[1:] {
		if r.Rank < minRank {
			minRank = r.Rank
		}
		if r.Rank > maxRank {
			maxRank = r.Rank
		}
	}

	rangeVal := maxRank - minRank
	normalized := make([]FTSResult, len(results))
	copy(normalized, results)

	for i := range normalized {
		if rangeVal == 0 {
			normalized[i].Rank = 1.0 // all same rank → score 1.0
		} else {
			// Invert: most negative (best) → highest score
			normalized[i].Rank = (maxRank - normalized[i].Rank) / rangeVal
		}
	}

	return normalized
}

// HybridRank merges FTS5 and vector results using weighted scoring.
// Weights: 0.4 * FTS5 + 0.6 * vector. Deduplicates by symbol ID.
func HybridRank(ftsResults []FTSResult, vectorResults []VectorResult) []RankedResult {
	type merged struct {
		result      RankedResult
		ftsScore    float64
		vectorScore float64
		hasFTS      bool
		hasVector   bool
	}

	byID := make(map[int64]*merged)

	// Add FTS results
	for _, f := range ftsResults {
		m := &merged{
			result: RankedResult{
				ID:         f.ID,
				FilePath:   f.FilePath,
				SymbolName: f.SymbolName,
				SymbolType: f.SymbolType,
				Language:   f.Language,
				Signature:  f.Signature,
				StartLine:  f.StartLine,
				EndLine:    f.EndLine,
			},
			ftsScore: f.Rank, // already normalized to 0..1
			hasFTS:   true,
		}
		byID[f.ID] = m
	}

	// Add/merge vector results
	for _, v := range vectorResults {
		if m, ok := byID[v.ID]; ok {
			m.vectorScore = float64(v.Score)
			m.hasVector = true
		} else {
			byID[v.ID] = &merged{
				result: RankedResult{
					ID:         v.ID,
					FilePath:   v.FilePath,
					SymbolName: v.SymbolName,
					SymbolType: v.SymbolType,
					Language:   v.Language,
					Signature:  v.Signature,
					StartLine:  v.StartLine,
					EndLine:    v.EndLine,
				},
				vectorScore: float64(v.Score),
				hasVector:   true,
			}
		}
	}

	// Compute final scores and collect results
	results := make([]RankedResult, 0, len(byID))
	for _, m := range byID {
		switch {
		case m.hasFTS && m.hasVector:
			m.result.Score = 0.4*m.ftsScore + 0.6*m.vectorScore
		case m.hasFTS:
			m.result.Score = m.ftsScore
		case m.hasVector:
			m.result.Score = m.vectorScore
		}
		results = append(results, m.result)
	}

	// Sort by score descending
	sortRankedResults(results)

	return results
}

// sortRankedResults sorts results by score descending using insertion sort
// (fine for typical result set sizes < 100).
func sortRankedResults(results []RankedResult) {
	for i := 1; i < len(results); i++ {
		key := results[i]
		j := i - 1
		for j >= 0 && results[j].Score < key.Score {
			results[j+1] = results[j]
			j--
		}
		results[j+1] = key
	}
}
