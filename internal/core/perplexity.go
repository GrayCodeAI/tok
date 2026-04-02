// Package core provides perplexity-based filtering using statistical language models.
package core

import (
	"math"
	"strings"
	"sync"
)

// PerplexityFilter uses n-gram models to estimate content information density.
// High perplexity indicates unpredictable/noisy content that can be compressed.
type PerplexityFilter struct {
	maxNgram  int
	threshold float64
	minLength int
	model     *NGramModel
	useGPU    bool
}

// NGramModel represents a statistical language model.
type NGramModel struct {
	mu       sync.RWMutex
	unigrams map[string]int
	bigrams  map[string]int
	rigrams  map[string]int
	total    int
}

// PerplexityConfig configures the perplexity filter.
type PerplexityConfig struct {
	MaxNgram  int
	Threshold float64 // Perplexity threshold for filtering
	MinLength int     // Minimum content length to apply
	UseGPU    bool    // Enable GPU acceleration
}

// DefaultPerplexityConfig returns default configuration.
func DefaultPerplexityConfig() PerplexityConfig {
	return PerplexityConfig{
		MaxNgram:  3,
		Threshold: 100.0, // Perplexity > 100 = high entropy
		MinLength: 100,
		UseGPU:    false,
	}
}

// NewPerplexityFilter creates a new perplexity filter.
func NewPerplexityFilter(config PerplexityConfig) *PerplexityFilter {
	return &PerplexityFilter{
		maxNgram:  config.MaxNgram,
		threshold: config.Threshold,
		minLength: config.MinLength,
		useGPU:    config.UseGPU,
		model:     NewNGramModel(),
	}
}

// NewNGramModel creates a new n-gram model.
func NewNGramModel() *NGramModel {
	return &NGramModel{
		unigrams: make(map[string]int),
		bigrams:  make(map[string]int),
		rigrams:  make(map[string]int),
	}
}

// Train trains the model on sample text.
func (m *NGramModel) Train(text string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Tokenize into characters for simplicity
	chars := []rune(strings.ToLower(text))

	for i := 0; i < len(chars); i++ {
		c := string(chars[i])
		m.unigrams[c]++
		m.total++

		if i < len(chars)-1 {
			bigram := string(chars[i : i+2])
			m.bigrams[bigram]++
		}

		if i < len(chars)-2 {
			trigram := string(chars[i : i+3])
			m.rigrams[trigram]++
		}
	}
}

// Perplexity calculates perplexity of text using the model.
func (m *NGramModel) Perplexity(text string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.total == 0 {
		return math.MaxFloat64
	}

	chars := []rune(strings.ToLower(text))
	if len(chars) < 3 {
		return 1.0
	}

	logProb := 0.0
	count := 0

	// Use trigram model with backoff
	for i := 2; i < len(chars); i++ {
		trigram := string(chars[i-2 : i+1])
		bigram := string(chars[i-2 : i])
		unigram := string(chars[i])

		prob := m.getTrigramProb(trigram, bigram, unigram)
		logProb += math.Log2(prob)
		count++
	}

	if count == 0 {
		return 1.0
	}

	// Perplexity = 2^(-logP / N)
	entropy := -logProb / float64(count)
	return math.Pow(2, entropy)
}

func (m *NGramModel) getTrigramProb(trigram, bigram, unigram string) float64 {
	// Trigram probability with Katz backoff
	trigramCount := m.rigrams[trigram]
	bigramCount := m.bigrams[bigram]

	if trigramCount > 0 && bigramCount > 0 {
		return float64(trigramCount) / float64(bigramCount)
	}

	// Backoff to bigram
	unigramCount := m.unigrams[unigram]
	if bigramCount > 0 && unigramCount > 0 {
		return 0.4 * float64(bigramCount) / float64(unigramCount)
	}

	// Backoff to unigram
	if m.total > 0 {
		return 0.4 * 0.4 * float64(m.unigrams[unigram]) / float64(m.total)
	}

	return 1.0 / 256.0 // Uniform probability
}

// TrainOnCorpus trains the model on a corpus of code/text.
func (f *PerplexityFilter) TrainOnCorpus(samples []string) {
	for _, sample := range samples {
		f.model.Train(sample)
	}
}

// ScoreResult contains perplexity analysis results.
type ScoreResult struct {
	Perplexity  float64
	EntropyType EntropyType
	ShouldKeep  bool
	Confidence  float64
}

// Score scores content for information density.
func (f *PerplexityFilter) Score(content string) ScoreResult {
	if len(content) < f.minLength {
		return ScoreResult{
			Perplexity:  1.0,
			EntropyType: MediumEntropy,
			ShouldKeep:  true,
			Confidence:  1.0,
		}
	}

	perplexity := f.model.Perplexity(content)

	// Classify entropy level
	var entropyType EntropyType
	switch {
	case perplexity > f.threshold*2:
		entropyType = HighEntropy
	case perplexity < f.threshold/2:
		entropyType = LowEntropy
	default:
		entropyType = MediumEntropy
	}

	// Determine if content should be kept
	shouldKeep := entropyType != HighEntropy
	confidence := math.Min(perplexity/f.threshold, 1.0)

	return ScoreResult{
		Perplexity:  perplexity,
		EntropyType: entropyType,
		ShouldKeep:  shouldKeep,
		Confidence:  confidence,
	}
}

// Filter applies perplexity-based filtering.
func (f *PerplexityFilter) Filter(content string) (string, FilterInfo) {
	lines := strings.Split(content, "\n")
	var kept []string
	perplexities := make([]float64, 0, len(lines))
	highEntropyLines := 0

	// Score each line
	for _, line := range lines {
		if len(line) < 10 {
			kept = append(kept, line)
			continue
		}

		score := f.Score(line)
		perplexities = append(perplexities, score.Perplexity)

		if score.EntropyType == HighEntropy {
			highEntropyLines++
			// Summarize high entropy lines
			if len(line) > 50 {
				kept = append(kept, summarizeLine(line))
			}
		} else {
			kept = append(kept, line)
		}
	}

	result := strings.Join(kept, "\n")
	avgPerplexity := average(perplexities)

	return result, FilterInfo{
		OriginalSize:     len(content),
		FilteredSize:     len(result),
		Perplexity:       avgPerplexity,
		HighEntropyLines: highEntropyLines,
	}
}

// FilterInfo contains filtering statistics.
type FilterInfo struct {
	OriginalSize     int
	FilteredSize     int
	Perplexity       float64
	HighEntropyLines int
}

// ReductionPercent returns the percentage reduction.
func (f FilterInfo) ReductionPercent() float64 {
	if f.OriginalSize == 0 {
		return 0
	}
	return float64(f.OriginalSize-f.FilteredSize) / float64(f.OriginalSize) * 100
}

// LineScore contains per-line perplexity score.
type LineScore struct {
	Line       int
	Text       string
	Perplexity float64
	Entropy    EntropyType
}

// AnalyzeLines analyzes each line separately.
func (f *PerplexityFilter) AnalyzeLines(content string) []LineScore {
	lines := strings.Split(content, "\n")
	scores := make([]LineScore, 0, len(lines))

	for i, line := range lines {
		if len(line) < 10 {
			scores = append(scores, LineScore{
				Line:       i + 1,
				Text:       line,
				Perplexity: 0,
				Entropy:    LowEntropy,
			})
			continue
		}

		score := f.Score(line)
		scores = append(scores, LineScore{
			Line:       i + 1,
			Text:       line,
			Perplexity: score.Perplexity,
			Entropy:    score.EntropyType,
		})
	}

	return scores
}

// BatchProcess processes multiple documents in parallel.
func (f *PerplexityFilter) BatchProcess(documents []string) []FilterInfo {
	results := make([]FilterInfo, len(documents))

	var wg sync.WaitGroup
	for i, doc := range documents {
		wg.Add(1)
		go func(idx int, content string) {
			defer wg.Done()
			_, info := f.Filter(content)
			results[idx] = info
		}(i, doc)
	}
	wg.Wait()

	return results
}

// PretrainedModel provides a basic pretrained model for common code patterns.
func PretrainedModel() *NGramModel {
	model := NewNGramModel()

	// Train on common code patterns
	patterns := []string{
		"func main() {\n\tfmt.Println(\"hello\")\n}",
		"package main\n\nimport \"fmt\"",
		"if err != nil {\n\treturn err\n}",
		"for i := 0; i < n; i++ {",
		"return nil",
		"struct {\n\tName string\n}",
		"interface {\n\tMethod()\n}",
		"// TODO: implement",
		"log.Printf(\"%s\", msg)",
		"http.Get(url)",
	}

	for _, pattern := range patterns {
		model.Train(pattern)
	}

	return model
}

func summarizeLine(line string) string {
	// Remove excessive whitespace
	line = strings.Join(strings.Fields(line), " ")

	// Truncate long lines but preserve structure
	if len(line) > 80 {
		return line[:77] + "..."
	}
	return line
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
