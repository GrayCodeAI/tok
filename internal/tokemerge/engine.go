package tokemerge

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
)

type TokenMergingEngine struct {
	mu       sync.RWMutex
	config   EngineConfig
	stats    EngineStats
	matchers map[string]Matcher
}

type EngineConfig struct {
	MergeRatio        float64
	EnableBipartite   bool
	EnableProgressive bool
	EnableReverse     bool
	SimilarityMetric  string
}

type EngineStats struct {
	TotalMerges     int64
	TokensMerged    int64
	TokensKept      int64
	AvgReduction    float64
	MergeOperations map[string]int64
}

type Matcher interface {
	Match(ctx context.Context, tokens []Token, mergeRatio float64) (*MergeResult, error)
	Name() string
}

type Token struct {
	ID         string
	Embedding  []float64
	Importance float64
	Position   int
	Metadata   map[string]interface{}
}

type MergeResult struct {
	MergedTokens  []Token
	OriginalCount int
	MergedCount   int
	Reduction     float64
	Merges        []Merge
}

type Merge struct {
	SourceIDs []string
	TargetID  string
	Weight    float64
}

func NewTokenMergingEngine(config EngineConfig) *TokenMergingEngine {
	e := &TokenMergingEngine{
		config: config,
		stats: EngineStats{
			MergeOperations: make(map[string]int64),
		},
		matchers: make(map[string]Matcher),
	}

	e.registerMatchers()

	return e
}

func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		MergeRatio:        0.5,
		EnableBipartite:   true,
		EnableProgressive: true,
		EnableReverse:     true,
		SimilarityMetric:  "cosine",
	}
}

func (e *TokenMergingEngine) registerMatchers() {
	e.matchers["bipartite"] = &BipartiteMatcher{}
	e.matchers["kmeans"] = &KMeansMatcher{}
	e.matchers["greedy"] = &GreedyMatcher{}
}

func (e *TokenMergingEngine) Merge(ctx context.Context, tokens []Token) (*MergeResult, error) {
	if len(tokens) == 0 {
		return &MergeResult{}, nil
	}

	mergeCount := int(float64(len(tokens)) * e.config.MergeRatio)
	if mergeCount < 1 {
		mergeCount = 1
	}

	matcherName := "bipartite"
	if e.config.EnableBipartite {
		matcherName = "bipartite"
	} else {
		matcherName = "greedy"
	}

	matcher, ok := e.matchers[matcherName]
	if !ok {
		matcher = e.matchers["greedy"]
	}

	result, err := matcher.Match(ctx, tokens, float64(mergeCount)/float64(len(tokens)))
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	e.stats.TotalMerges++
	e.stats.TokensMerged += int64(result.OriginalCount - result.MergedCount)
	e.stats.TokensKept += int64(result.MergedCount)
	e.stats.MergeOperations[matcherName]++

	if e.stats.TotalMerges > 1 {
		e.stats.AvgReduction = (e.stats.AvgReduction*float64(e.stats.TotalMerges-1) + result.Reduction) / float64(e.stats.TotalMerges)
	} else {
		e.stats.AvgReduction = result.Reduction
	}
	e.mu.Unlock()

	return result, nil
}

func (e *TokenMergingEngine) Unmerge(ctx context.Context, mergedTokens []Token, originalCount int) []Token {
	if !e.config.EnableReverse || len(mergedTokens) == 0 {
		return mergedTokens
	}

	result := make([]Token, 0, originalCount)

	for _, token := range mergedTokens {
		if splits, ok := token.Metadata["splits"].([]string); ok && len(splits) > 0 {
			for _, splitID := range splits {
				result = append(result, Token{
					ID:         splitID,
					Embedding:  token.Embedding,
					Importance: token.Importance / float64(len(splits)),
				})
			}
		} else {
			result = append(result, token)
		}
	}

	return result
}

func (e *TokenMergingEngine) GetStats() EngineStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stats
}

type BipartiteMatcher struct{}

func (m *BipartiteMatcher) Name() string { return "bipartite" }

func (m *BipartiteMatcher) Match(ctx context.Context, tokens []Token, mergeRatio float64) (*MergeResult, error) {
	if len(tokens) < 2 {
		return &MergeResult{
			MergedTokens:  tokens,
			OriginalCount: len(tokens),
			MergedCount:   len(tokens),
			Reduction:     0,
		}, nil
	}

	n := len(tokens)
	mid := n / 2

	left := tokens[:mid]
	right := tokens[mid:]

	similarityMatrix := computeSimilarityMatrix(left, right)

	merges := make([]Merge, 0)
	mergedRight := make(map[int]bool)

	type pairScore struct {
		leftIdx  int
		rightIdx int
		score    float64
	}

	var pairs []pairScore
	for i := range left {
		for j := range right {
			if mergedRight[j] {
				continue
			}
			pairs = append(pairs, pairScore{
				leftIdx:  i,
				rightIdx: j,
				score:    similarityMatrix[i][j],
			})
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].score > pairs[j].score
	})

	numMerges := int(float64(len(right)) * mergeRatio)
	if numMerges > len(pairs) {
		numMerges = len(pairs)
	}

	for i := 0; i < numMerges; i++ {
		p := pairs[i]

		if mergedRight[p.rightIdx] {
			continue
		}

		mergedRight[p.rightIdx] = true

		merges = append(merges, Merge{
			SourceIDs: []string{left[p.leftIdx].ID, right[p.rightIdx].ID},
			TargetID:  fmt.Sprintf("merged_%d_%d", p.leftIdx, p.rightIdx),
			Weight:    p.score,
		})
	}

	mergedTokens := make([]Token, 0, len(merges)+len(left)+len(right)-numMerges)

	for _, token := range left {
		mergedTokens = append(mergedTokens, token)
	}

	for j, token := range right {
		if mergedRight[j] {
			continue
		}
		mergedTokens = append(mergedTokens, token)
	}

	for _, merge := range merges {
		mergedEmbedding := computeMergedEmbedding(tokens, merge.SourceIDs)
		mergedImportance := computeMergedImportance(tokens, merge.SourceIDs)

		mergedTokens = append(mergedTokens, Token{
			ID:         merge.TargetID,
			Embedding:  mergedEmbedding,
			Importance: mergedImportance,
			Metadata:   map[string]interface{}{"splits": merge.SourceIDs},
		})
	}

	result := &MergeResult{
		MergedTokens:  mergedTokens,
		OriginalCount: n,
		MergedCount:   len(mergedTokens),
		Merges:        merges,
	}

	if n > 0 {
		result.Reduction = float64(n-len(mergedTokens)) / float64(n)
	}

	return result, nil
}

func computeSimilarityMatrix(a, b []Token) [][]float64 {
	n, m := len(a), len(b)
	matrix := make([][]float64, n)

	for i := range a {
		matrix[i] = make([]float64, m)
		for j := range b {
			matrix[i][j] = cosineSimilarity(a[i].Embedding, b[j].Embedding)
		}
	}

	return matrix
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	dot := 0.0
	normA := 0.0
	normB := 0.0

	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func computeMergedEmbedding(tokens []Token, ids []string) []float64 {
	if len(ids) == 0 {
		return nil
	}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	var result []float64
	for _, token := range tokens {
		if idSet[token.ID] {
			if result == nil {
				result = make([]float64, len(token.Embedding))
			}
			for i, v := range token.Embedding {
				result[i] += v
			}
		}
	}

	if result != nil {
		divisor := float64(len(ids))
		for i := range result {
			result[i] /= divisor
		}
	}

	return result
}

func computeMergedImportance(tokens []Token, ids []string) float64 {
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	var total float64
	for _, token := range tokens {
		if idSet[token.ID] {
			total += token.Importance
		}
	}

	return total / float64(len(ids))
}

type KMeansMatcher struct{}

func (m *KMeansMatcher) Name() string { return "kmeans" }

func (m *KMeansMatcher) Match(ctx context.Context, tokens []Token, mergeRatio float64) (*MergeResult, error) {
	if len(tokens) < 2 {
		return &MergeResult{
			MergedTokens:  tokens,
			OriginalCount: len(tokens),
			MergedCount:   len(tokens),
			Reduction:     0,
		}, nil
	}

	k := int(float64(len(tokens)) * (1 - mergeRatio))
	if k < 1 {
		k = 1
	}

	centroids := make([][]float64, k)
	for i := range centroids {
		centroids[i] = make([]float64, len(tokens[0].Embedding))
		copy(centroids[i], tokens[i].Embedding)
	}

	clusters := make([][]Token, k)

	for iter := 0; iter < 10; iter++ {
		for i := range clusters {
			clusters[i] = clusters[i][:0]
		}

		for _, token := range tokens {
			minDist := math.MaxFloat64
			closest := 0

			for i, centroid := range centroids {
				dist := euclideanDist(token.Embedding, centroid)
				if dist < minDist {
					minDist = dist
					closest = i
				}
			}

			clusters[closest] = append(clusters[closest], token)
		}

		for i := range centroids {
			if len(clusters[i]) > 0 {
				for dim := range centroids[i] {
					var sum float64
					for _, token := range clusters[i] {
						sum += token.Embedding[dim]
					}
					centroids[i][dim] = sum / float64(len(clusters[i]))
				}
			}
		}
	}

	merges := make([]Merge, 0)
	mergedTokens := make([]Token, 0)

	for i, cluster := range clusters {
		if len(cluster) == 0 {
			continue
		}

		if len(cluster) > 1 {
			ids := make([]string, len(cluster))
			for j, t := range cluster {
				ids[j] = t.ID
			}

			merges = append(merges, Merge{
				SourceIDs: ids,
				TargetID:  fmt.Sprintf("cluster_%d", i),
				Weight:    1.0,
			})

			mergedTokens = append(mergedTokens, Token{
				ID:         fmt.Sprintf("cluster_%d", i),
				Embedding:  centroids[i],
				Importance: computeMergedImportance(tokens, ids),
				Metadata:   map[string]interface{}{"splits": ids},
			})
		} else {
			mergedTokens = append(mergedTokens, cluster[0])
		}
	}

	result := &MergeResult{
		MergedTokens:  mergedTokens,
		OriginalCount: len(tokens),
		MergedCount:   len(mergedTokens),
		Merges:        merges,
	}

	if len(tokens) > 0 {
		result.Reduction = float64(len(tokens)-len(mergedTokens)) / float64(len(tokens))
	}

	return result, nil
}

func euclideanDist(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.MaxFloat64
	}

	sum := 0.0
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}

	return math.Sqrt(sum)
}

type GreedyMatcher struct{}

func (m *GreedyMatcher) Name() string { return "greedy" }

func (m *GreedyMatcher) Match(ctx context.Context, tokens []Token, mergeRatio float64) (*MergeResult, error) {
	if len(tokens) < 2 {
		return &MergeResult{
			MergedTokens:  tokens,
			OriginalCount: len(tokens),
			MergedCount:   len(tokens),
			Reduction:     0,
		}, nil
	}

	numMerges := int(float64(len(tokens)-1) * mergeRatio)
	if numMerges < 1 {
		numMerges = 1
	}

	merged := make(map[int]bool)
	merges := make([]Merge, 0)

	type tokenSimilarity struct {
		i, j       int
		similarity float64
	}

	var similarities []tokenSimilarity
	for i := 0; i < len(tokens); i++ {
		for j := i + 1; j < len(tokens); j++ {
			if merged[i] || merged[j] {
				continue
			}
			sim := cosineSimilarity(tokens[i].Embedding, tokens[j].Embedding)
			similarities = append(similarities, tokenSimilarity{i, j, sim})
		}
	}

	sort.Slice(similarities, func(a, b int) bool {
		return similarities[a].similarity > similarities[b].similarity
	})

	for _, sim := range similarities {
		if len(merges) >= numMerges {
			break
		}
		if merged[sim.i] || merged[sim.j] {
			continue
		}

		merged[sim.i] = true

		merges = append(merges, Merge{
			SourceIDs: []string{tokens[sim.i].ID, tokens[sim.j].ID},
			TargetID:  fmt.Sprintf("merged_%d_%d", sim.i, sim.j),
			Weight:    sim.similarity,
		})
	}

	mergedTokens := make([]Token, 0)

	for i, token := range tokens {
		if !merged[i] {
			mergedTokens = append(mergedTokens, token)
		}
	}

	for _, merge := range merges {
		mergedEmbedding := computeMergedEmbedding(tokens, merge.SourceIDs)

		mergedTokens = append(mergedTokens, Token{
			ID:         merge.TargetID,
			Embedding:  mergedEmbedding,
			Importance: computeMergedImportance(tokens, merge.SourceIDs),
			Metadata:   map[string]interface{}{"splits": merge.SourceIDs},
		})
	}

	result := &MergeResult{
		MergedTokens:  mergedTokens,
		OriginalCount: len(tokens),
		MergedCount:   len(mergedTokens),
		Merges:        merges,
	}

	if len(tokens) > 0 {
		result.Reduction = float64(len(tokens)-len(mergedTokens)) / float64(len(tokens))
	}

	return result, nil
}
