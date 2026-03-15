package embeddings

import "math"

// CosineSimilarity computes the cosine similarity between two embedding vectors.
// Returns a value between -1.0 and 1.0, where 1.0 means identical direction.
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	denominator := math.Sqrt(normA) * math.Sqrt(normB)
	if denominator == 0 {
		return 0
	}

	return float32(dotProduct / denominator)
}

// TopK finds the k most similar vectors to the query from a set of candidates.
// Returns indices and scores, sorted by similarity descending.
func TopK(query []float32, candidates [][]float32, k int) (indices []int, scores []float32) {
	if len(candidates) == 0 || k <= 0 {
		return nil, nil
	}

	type scored struct {
		index int
		score float32
	}

	all := make([]scored, len(candidates))
	for i, c := range candidates {
		all[i] = scored{index: i, score: CosineSimilarity(query, c)}
	}

	// Simple selection sort for top-k (fine for typical candidate set sizes)
	for i := 0; i < k && i < len(all); i++ {
		maxIdx := i
		for j := i + 1; j < len(all); j++ {
			if all[j].score > all[maxIdx].score {
				maxIdx = j
			}
		}
		all[i], all[maxIdx] = all[maxIdx], all[i]
	}

	n := k
	if n > len(all) {
		n = len(all)
	}

	indices = make([]int, n)
	scores = make([]float32, n)
	for i := 0; i < n; i++ {
		indices[i] = all[i].index
		scores[i] = all[i].score
	}

	return indices, scores
}
