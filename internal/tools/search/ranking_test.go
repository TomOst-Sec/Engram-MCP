package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeFTSScores_Empty(t *testing.T) {
	result := NormalizeFTSScores(nil)
	assert.Nil(t, result)
}

func TestNormalizeFTSScores_SingleResult(t *testing.T) {
	results := []FTSResult{{ID: 1, Rank: -5.0}}
	normalized := NormalizeFTSScores(results)
	require.Len(t, normalized, 1)
	assert.Equal(t, 1.0, normalized[0].Rank) // single result → 1.0
}

func TestNormalizeFTSScores_Range(t *testing.T) {
	results := []FTSResult{
		{ID: 1, Rank: -10.0}, // best match (most negative)
		{ID: 2, Rank: -5.0},  // middle
		{ID: 3, Rank: -1.0},  // worst match
	}
	normalized := NormalizeFTSScores(results)
	require.Len(t, normalized, 3)

	// Most negative should have highest normalized score
	assert.InDelta(t, 1.0, normalized[0].Rank, 0.001)
	// Least negative should have lowest normalized score
	assert.InDelta(t, 0.0, normalized[2].Rank, 0.001)
	// Middle should be between
	assert.True(t, normalized[1].Rank > 0.0 && normalized[1].Rank < 1.0)

	// All values should be in [0,1]
	for _, r := range normalized {
		assert.True(t, r.Rank >= 0.0 && r.Rank <= 1.0,
			"rank should be in [0,1], got %f", r.Rank)
	}
}

func TestHybridRank_FTSOnly(t *testing.T) {
	fts := []FTSResult{
		{ID: 1, SymbolName: "Func1", Rank: 0.9},
		{ID: 2, SymbolName: "Func2", Rank: 0.5},
	}
	results := HybridRank(fts, nil)
	require.Len(t, results, 2)
	// FTS-only: score = ftsScore
	assert.Equal(t, "Func1", results[0].SymbolName)
	assert.InDelta(t, 0.9, results[0].Score, 0.001)
}

func TestHybridRank_VectorOnly(t *testing.T) {
	vec := []VectorResult{
		{ID: 1, SymbolName: "Func1", Score: 0.95},
		{ID: 2, SymbolName: "Func2", Score: 0.7},
	}
	results := HybridRank(nil, vec)
	require.Len(t, results, 2)
	// Vector-only: score = vectorScore
	assert.Equal(t, "Func1", results[0].SymbolName)
	assert.InDelta(t, 0.95, results[0].Score, 0.001)
}

func TestHybridRank_Combined(t *testing.T) {
	fts := []FTSResult{
		{ID: 1, SymbolName: "Func1", Rank: 0.8},
	}
	vec := []VectorResult{
		{ID: 1, SymbolName: "Func1", Score: 0.9},
	}
	results := HybridRank(fts, vec)
	require.Len(t, results, 1)
	// Combined: 0.4 * 0.8 + 0.6 * 0.9 = 0.32 + 0.54 = 0.86
	assert.InDelta(t, 0.86, results[0].Score, 0.001)
}

func TestHybridRank_Deduplication(t *testing.T) {
	fts := []FTSResult{
		{ID: 1, SymbolName: "Func1", Rank: 0.8},
		{ID: 2, SymbolName: "Func2", Rank: 0.5},
	}
	vec := []VectorResult{
		{ID: 1, SymbolName: "Func1", Score: 0.9}, // duplicate with ID 1
		{ID: 3, SymbolName: "Func3", Score: 0.7}, // unique
	}
	results := HybridRank(fts, vec)
	require.Len(t, results, 3) // deduped: IDs 1, 2, 3

	// Check ID 1 appears only once
	idCount := 0
	for _, r := range results {
		if r.ID == 1 {
			idCount++
		}
	}
	assert.Equal(t, 1, idCount)
}

func TestHybridRank_SortedDescending(t *testing.T) {
	fts := []FTSResult{
		{ID: 1, Rank: 0.3},
		{ID: 2, Rank: 0.9},
		{ID: 3, Rank: 0.6},
	}
	results := HybridRank(fts, nil)
	for i := 1; i < len(results); i++ {
		assert.True(t, results[i-1].Score >= results[i].Score,
			"results should be descending: %f >= %f", results[i-1].Score, results[i].Score)
	}
}

func TestHybridRank_WeightCorrectness(t *testing.T) {
	// FTS weight 0.4, vector weight 0.6
	fts := []FTSResult{
		{ID: 1, Rank: 1.0}, // normalized FTS score
	}
	vec := []VectorResult{
		{ID: 1, Score: 1.0}, // vector score
	}
	results := HybridRank(fts, vec)
	require.Len(t, results, 1)
	// 0.4 * 1.0 + 0.6 * 1.0 = 1.0
	assert.InDelta(t, 1.0, results[0].Score, 0.001)

	fts2 := []FTSResult{
		{ID: 1, Rank: 0.0},
	}
	vec2 := []VectorResult{
		{ID: 1, Score: 1.0},
	}
	results2 := HybridRank(fts2, vec2)
	// 0.4 * 0.0 + 0.6 * 1.0 = 0.6
	assert.InDelta(t, 0.6, results2[0].Score, 0.001)
}

func TestHybridRank_Empty(t *testing.T) {
	results := HybridRank(nil, nil)
	assert.Empty(t, results)
}
