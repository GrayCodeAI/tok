package extractivesum

import (
	"context"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type AdvancedExtractiveEngine struct {
	mu         sync.RWMutex
	config     EngineConfig
	algorithms map[string]Extractor
	stats      EngineStats
}

type EngineConfig struct {
	DefaultAlgorithm  string
	MaxSummaryLength  int
	MinSentenceLength int
	EnableMultiDoc    bool
	EnableQueryFocus  bool
}

type EngineStats struct {
	TotalSummaries       int64
	TotalInputSentences  int64
	TotalOutputSentences int64
	AvgReduction         float64
	AlgorithmUsage       map[string]int64
}

type Extractor interface {
	Extract(ctx context.Context, sentences []Sentence, options ExtractOptions) ([]Sentence, error)
	Name() string
}

type Sentence struct {
	Text       string
	Score      float64
	Position   int
	Importance float64
	Tokens     []string
	Metadata   map[string]interface{}
}

type ExtractOptions struct {
	MaxSentences int
	Query        string
	Topic        string
	MinScore     float64
}

func NewAdvancedExtractiveEngine(config EngineConfig) *AdvancedExtractiveEngine {
	e := &AdvancedExtractiveEngine{
		config:     config,
		algorithms: make(map[string]Extractor),
		stats: EngineStats{
			AlgorithmUsage: make(map[string]int64),
		},
	}

	e.registerAlgorithms()

	return e
}

func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		DefaultAlgorithm:  "sumbasic",
		MaxSummaryLength:  10,
		MinSentenceLength: 20,
		EnableMultiDoc:    true,
		EnableQueryFocus:  true,
	}
}

func (e *AdvancedExtractiveEngine) registerAlgorithms() {
	e.RegisterExtractor(&SumBasicExtractor{})
	e.RegisterExtractor(&KLSumExtractor{})
	e.RegisterExtractor(&MMRExtractor{})
	e.RegisterExtractor(&CentroidExtractor{})
	e.RegisterExtractor(&SubmodularExtractor{})
	e.RegisterExtractor(&PositionExtractor{})
}

func (e *AdvancedExtractiveEngine) RegisterExtractor(ext Extractor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.algorithms[ext.Name()] = ext
}

func (e *AdvancedExtractiveEngine) Extract(ctx context.Context, documents []string, options ExtractOptions) ([]Sentence, error) {
	allSentences := make([]Sentence, 0)

	for docIdx, doc := range documents {
		sentences := e.tokenize(doc)
		for pos, sent := range sentences {
			sent.Position = pos
			sent.Metadata = map[string]interface{}{"doc": docIdx}
			allSentences = append(allSentences, sent)
		}
	}

	if len(allSentences) == 0 {
		return []Sentence{}, nil
	}

	algoName := e.config.DefaultAlgorithm
	if options.Query != "" && e.config.EnableQueryFocus {
		algoName = "mmr"
	}

	e.mu.RLock()
	algo, ok := e.algorithms[algoName]
	e.mu.RUnlock()

	if !ok {
		algo = e.algorithms["sumbasic"]
	}

	result, err := algo.Extract(ctx, allSentences, options)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	e.stats.TotalSummaries++
	e.stats.TotalInputSentences += int64(len(allSentences))
	e.stats.TotalOutputSentences += int64(len(result))

	if e.stats.TotalSummaries > 1 {
		e.stats.AvgReduction = (e.stats.AvgReduction*float64(e.stats.TotalSummaries-1) + float64(len(allSentences)-len(result))/float64(len(allSentences))) / float64(e.stats.TotalSummaries)
	} else {
		e.stats.AvgReduction = float64(len(allSentences)-len(result)) / float64(len(allSentences))
	}

	e.stats.AlgorithmUsage[algoName]++
	e.mu.Unlock()

	return result, nil
}

func (e *AdvancedExtractiveEngine) tokenize(text string) []Sentence {
	sentences := regexp.MustCompile(`[.!?]+\s+`).Split(text, -1)
	result := make([]Sentence, 0, len(sentences))

	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if len(s) < e.config.MinSentenceLength {
			continue
		}

		tokens := extractTokens(s)
		if len(tokens) < 3 {
			continue
		}

		result = append(result, Sentence{
			Text:   s,
			Tokens: tokens,
		})
	}

	return result
}

func extractTokens(s string) []string {
	re := regexp.MustCompile(`[a-zA-Z]+`)
	return re.FindAllString(s, -1)
}

func (e *AdvancedExtractiveEngine) GetStats() EngineStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stats
}

type SumBasicExtractor struct{}

func (e *SumBasicExtractor) Name() string { return "sumbasic" }

func (e *SumBasicExtractor) Extract(ctx context.Context, sentences []Sentence, options ExtractOptions) ([]Sentence, error) {
	if len(sentences) == 0 {
		return []Sentence{}, nil
	}

	wordFreq := make(map[string]float64)
	totalWords := 0

	for _, sent := range sentences {
		for _, word := range sent.Tokens {
			lowercase := strings.ToLower(word)
			wordFreq[lowercase]++
			totalWords++
		}
	}

	for word := range wordFreq {
		wordFreq[word] /= float64(totalWords)
	}

	for i := range sentences {
		score := 0.0
		seen := make(map[string]bool)
		for _, word := range sentences[i].Tokens {
			lowercase := strings.ToLower(word)
			if !seen[lowercase] {
				score += wordFreq[lowercase]
				seen[lowercase] = true
			}
		}
		sentences[i].Score = score
		sentences[i].Importance = score
	}

	maxSentences := options.MaxSentences
	if maxSentences <= 0 {
		maxSentences = 5
	}

	sorted := make([]Sentence, len(sentences))
	copy(sorted, sentences)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	result := sorted[:min(len(sorted), maxSentences)]

	sort.Slice(result, func(i, j int) bool {
		return result[i].Position < result[j].Position
	})

	return result, nil
}

type KLSumExtractor struct{}

func (e *KLSumExtractor) Name() string { return "kl_sum" }

func (e *KLSumExtractor) Extract(ctx context.Context, sentences []Sentence, options ExtractOptions) ([]Sentence, error) {
	if len(sentences) == 0 {
		return []Sentence{}, nil
	}

	docFreq := make(map[string]float64)
	totalDocs := float64(len(sentences))

	for _, sent := range sentences {
		seen := make(map[string]bool)
		for _, word := range sent.Tokens {
			lowercase := strings.ToLower(word)
			if !seen[lowercase] {
				docFreq[lowercase]++
				seen[lowercase] = true
			}
		}
	}

	for word := range docFreq {
		docFreq[word] = math.Log(totalDocs / (1 + docFreq[word]))
	}

	for i := range sentences {
		score := 0.0
		seen := make(map[string]bool)
		for _, word := range sentences[i].Tokens {
			lowercase := strings.ToLower(word)
			if !seen[lowercase] {
				score += docFreq[lowercase]
				seen[lowercase] = true
			}
		}
		sentences[i].Score = score
		sentences[i].Importance = score
	}

	maxSentences := options.MaxSentences
	if maxSentences <= 0 {
		maxSentences = 5
	}

	sorted := make([]Sentence, len(sentences))
	copy(sorted, sentences)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	return sorted[:min(len(sorted), maxSentences)], nil
}

type MMRExtractor struct{}

func (e *MMRExtractor) Name() string { return "mmr" }

func (e *MMRExtractor) Extract(ctx context.Context, sentences []Sentence, options ExtractOptions) ([]Sentence, error) {
	if len(sentences) == 0 {
		return []Sentence{}, nil
	}

	maxSentences := options.MaxSentences
	if maxSentences <= 0 {
		maxSentences = 5
	}

	queryWords := extractTokens(strings.ToLower(options.Query))

	for i := range sentences {
		score := 0.0
		lowerText := strings.ToLower(sentences[i].Text)

		for _, qw := range queryWords {
			if strings.Contains(lowerText, qw) {
				score += 1.0
			}
		}

		sentences[i].Score = score
		sentences[i].Importance = score
	}

	result := make([]Sentence, 0, maxSentences)
	selected := make(map[int]bool)

	lambda := 0.5

	for len(result) < maxSentences && len(result) < len(sentences) {
		bestScore := -math.MaxFloat64
		bestIdx := -1

		for i, sent := range sentences {
			if selected[i] {
				continue
			}

			relevance := sent.Score

			maxSim := 0.0
			for _, sel := range result {
				sim := sentenceSimilarity(sent, sel)
				if sim > maxSim {
					maxSim = sim
				}
			}

			mmrScore := lambda*relevance - (1-lambda)*maxSim

			if mmrScore > bestScore {
				bestScore = mmrScore
				bestIdx = i
			}
		}

		if bestIdx == -1 {
			break
		}

		selected[bestIdx] = true
		result = append(result, sentences[bestIdx])
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Position < result[j].Position
	})

	return result, nil
}

func sentenceSimilarity(a, b Sentence) float64 {
	if len(a.Tokens) == 0 || len(b.Tokens) == 0 {
		return 0
	}

	wordsA := make(map[string]bool)
	wordsB := make(map[string]bool)

	for _, w := range a.Tokens {
		wordsA[strings.ToLower(w)] = true
	}
	for _, w := range b.Tokens {
		wordsB[strings.ToLower(w)] = true
	}

	intersection := 0
	for w := range wordsA {
		if wordsB[w] {
			intersection++
		}
	}

	union := len(wordsA) + len(wordsB) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

type CentroidExtractor struct{}

func (e *CentroidExtractor) Name() string { return "centroid" }

func (e *CentroidExtractor) Extract(ctx context.Context, sentences []Sentence, options ExtractOptions) ([]Sentence, error) {
	if len(sentences) == 0 {
		return []Sentence{}, nil
	}

	centroidFreq := make(map[string]float64)
	wordCount := 0

	for _, sent := range sentences {
		for _, word := range sent.Tokens {
			lowercase := strings.ToLower(word)
			centroidFreq[lowercase]++
			wordCount++
		}
	}

	for word := range centroidFreq {
		centroidFreq[word] /= float64(wordCount)
	}

	for i := range sentences {
		score := 0.0
		seen := make(map[string]bool)
		for _, word := range sentences[i].Tokens {
			lowercase := strings.ToLower(word)
			if !seen[lowercase] {
				score += centroidFreq[lowercase]
				seen[lowercase] = true
			}
		}
		sentences[i].Score = score
		sentences[i].Importance = score
	}

	maxSentences := options.MaxSentences
	if maxSentences <= 0 {
		maxSentences = 5
	}

	sorted := make([]Sentence, len(sentences))
	copy(sorted, sentences)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	return sorted[:min(len(sorted), maxSentences)], nil
}

type SubmodularExtractor struct{}

func (e *SubmodularExtractor) Name() string { return "submodular" }

func (e *SubmodularExtractor) Extract(ctx context.Context, sentences []Sentence, options ExtractOptions) ([]Sentence, error) {
	if len(sentences) == 0 {
		return []Sentence{}, nil
	}

	wordDocFreq := make(map[string]int)
	for _, sent := range sentences {
		seen := make(map[string]bool)
		for _, word := range sent.Tokens {
			lowercase := strings.ToLower(word)
			if !seen[lowercase] {
				wordDocFreq[lowercase]++
				seen[lowercase] = true
			}
		}
	}

	n := float64(len(sentences))

	for i := range sentences {
		score := 0.0
		seen := make(map[string]bool)
		for _, word := range sentences[i].Tokens {
			lowercase := strings.ToLower(word)
			if !seen[lowercase] {
				df := float64(wordDocFreq[lowercase])
				idf := math.Log(n / (1 + df))
				score += idf
				seen[lowercase] = true
			}
		}

		lengthPenalty := 1.0 + math.Log(float64(len(sentences[i].Tokens)))
		sentences[i].Score = score / lengthPenalty
		sentences[i].Importance = sentences[i].Score
	}

	maxSentences := options.MaxSentences
	if maxSentences <= 0 {
		maxSentences = 5
	}

	result := submodularSelect(sentences, maxSentences)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Position < result[j].Position
	})

	return result, nil
}

func submodularSelect(sentences []Sentence, k int) []Sentence {
	if len(sentences) <= k {
		return sentences
	}

	sorted := make([]Sentence, len(sentences))
	copy(sorted, sentences)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	result := sorted[:k]

	for iter := 0; iter < 5; iter++ {
		improved := false

		for i := 0; i < k; i++ {
			for j := k; j < len(sorted); j++ {
				newResult := make([]Sentence, k)
				copy(newResult, result)
				newResult[i] = sorted[j]

				oldCoverage := coverage(result)
				newCoverage := coverage(newResult)

				if newCoverage > oldCoverage {
					result = newResult
					improved = true
				}
			}
		}

		if !improved {
			break
		}
	}

	return result
}

func coverage(sentences []Sentence) float64 {
	wordWeights := make(map[string]float64)
	totalWeight := 0.0

	for _, sent := range sentences {
		weight := sent.Importance
		for _, word := range sent.Tokens {
			lowercase := strings.ToLower(word)
			wordWeights[lowercase] += weight
			totalWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	coverage := 0.0
	for _, w := range wordWeights {
		coverage += w / totalWeight
	}

	return coverage
}

type PositionExtractor struct{}

func (e *PositionExtractor) Name() string { return "position" }

func (e *PositionExtractor) Extract(ctx context.Context, sentences []Sentence, options ExtractOptions) ([]Sentence, error) {
	if len(sentences) == 0 {
		return []Sentence{}, nil
	}

	totalSentences := len(sentences)

	for i := range sentences {
		positionScore := 1.0

		if i == 0 {
			positionScore = 1.5
		} else if i < 3 {
			positionScore = 1.2
		} else if i >= totalSentences-2 {
			positionScore = 1.1
		}

		lengthScore := float64(len(sentences[i].Tokens)) / 50.0
		if lengthScore > 1 {
			lengthScore = 1
		}

		sentences[i].Score = positionScore * (0.7 + 0.3*lengthScore)
		sentences[i].Importance = sentences[i].Score
	}

	maxSentences := options.MaxSentences
	if maxSentences <= 0 {
		maxSentences = 5
	}

	sorted := make([]Sentence, len(sentences))
	copy(sorted, sentences)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	return sorted[:min(len(sorted), maxSentences)], nil
}
