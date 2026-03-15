package embeddings

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCosineSimilarity_IdenticalVectors(t *testing.T) {
	v := []float32{1.0, 2.0, 3.0, 4.0}
	sim := CosineSimilarity(v, v)
	assert.InDelta(t, 1.0, float64(sim), 0.001)
}

func TestCosineSimilarity_OrthogonalVectors(t *testing.T) {
	a := []float32{1.0, 0.0, 0.0}
	b := []float32{0.0, 1.0, 0.0}
	sim := CosineSimilarity(a, b)
	assert.InDelta(t, 0.0, float64(sim), 0.001)
}

func TestCosineSimilarity_OppositeVectors(t *testing.T) {
	a := []float32{1.0, 2.0, 3.0}
	b := []float32{-1.0, -2.0, -3.0}
	sim := CosineSimilarity(a, b)
	assert.InDelta(t, -1.0, float64(sim), 0.001)
}

func TestCosineSimilarity_DifferentLengths(t *testing.T) {
	a := []float32{1.0, 2.0}
	b := []float32{1.0, 2.0, 3.0}
	sim := CosineSimilarity(a, b)
	assert.Equal(t, float32(0), sim)
}

func TestCosineSimilarity_EmptyVectors(t *testing.T) {
	sim := CosineSimilarity([]float32{}, []float32{})
	assert.Equal(t, float32(0), sim)
}

func TestCosineSimilarity_ZeroVector(t *testing.T) {
	a := []float32{0, 0, 0}
	b := []float32{1, 2, 3}
	sim := CosineSimilarity(a, b)
	assert.Equal(t, float32(0), sim)
}

func TestCosineSimilarity_SimilarVectors(t *testing.T) {
	a := []float32{1.0, 2.0, 3.0}
	b := []float32{1.1, 2.1, 3.1}
	sim := CosineSimilarity(a, b)
	assert.True(t, sim > 0.99, "similar vectors should have high similarity, got %f", sim)
}

func TestTopK_BasicUsage(t *testing.T) {
	query := []float32{1.0, 0.0, 0.0}
	candidates := [][]float32{
		{0.0, 1.0, 0.0}, // orthogonal
		{1.0, 0.0, 0.0}, // identical
		{0.5, 0.5, 0.0}, // partial match
		{-1.0, 0.0, 0.0}, // opposite
	}

	indices, scores := TopK(query, candidates, 3)
	require.Len(t, indices, 3)
	require.Len(t, scores, 3)

	// Best match should be index 1 (identical vector)
	assert.Equal(t, 1, indices[0])
	assert.InDelta(t, 1.0, float64(scores[0]), 0.001)

	// Scores should be in descending order
	for i := 1; i < len(scores); i++ {
		assert.True(t, scores[i-1] >= scores[i], "scores should be descending")
	}
}

func TestTopK_KLargerThanCandidates(t *testing.T) {
	query := []float32{1.0, 0.0}
	candidates := [][]float32{
		{1.0, 0.0},
		{0.0, 1.0},
	}

	indices, scores := TopK(query, candidates, 5)
	assert.Len(t, indices, 2)
	assert.Len(t, scores, 2)
}

func TestTopK_EmptyCandidates(t *testing.T) {
	query := []float32{1.0, 0.0}
	indices, scores := TopK(query, nil, 3)
	assert.Nil(t, indices)
	assert.Nil(t, scores)
}

func TestTopK_ZeroK(t *testing.T) {
	query := []float32{1.0, 0.0}
	candidates := [][]float32{{1.0, 0.0}}
	indices, scores := TopK(query, candidates, 0)
	assert.Nil(t, indices)
	assert.Nil(t, scores)
}

func TestTopK_CorrectOrdering(t *testing.T) {
	query := []float32{1.0, 0.0, 0.0}
	candidates := [][]float32{
		{0.0, 0.0, 1.0},             // cos = 0
		{1.0, 0.0, 0.0},             // cos = 1
		{float32(1 / math.Sqrt(2)), float32(1 / math.Sqrt(2)), 0.0}, // cos ≈ 0.707
	}

	indices, scores := TopK(query, candidates, 3)
	assert.Equal(t, 1, indices[0]) // identical
	assert.Equal(t, 2, indices[1]) // partial
	assert.Equal(t, 0, indices[2]) // orthogonal
	_ = scores
}
