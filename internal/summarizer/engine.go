package summarizer

import (
	"context"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type SummarizationEngine struct {
	mu         sync.RWMutex
	algorithms map[string]Summarizer
	config     EngineConfig
	history    *SummaryHistory
	stats      EngineStats
}

type EngineConfig struct {
	MaxLength        int
	MinLength        int
	Algorithm        string
	EnableHybrid     bool
	UseLocalLLM      bool
	LLMEndpoint      string
	QualityThreshold float64
}

type EngineStats struct {
	TotalSummaries    int64
	TotalInputTokens  int64
	TotalOutputTokens int64
	AvgReduction      float64
	AlgorithmUsage    map[string]int64
}

type Summarizer interface {
	Summarize(ctx context.Context, text string, options SummarizeOptions) (*Summary, error)
	Name() string
}

type SummarizeOptions struct {
	MaxLength int
	MinLength int
	Style     string
	Query     string
	Format    string
}

type Summary struct {
	Text         string
	InputTokens  int
	OutputTokens int
	Reduction    float64
	Algorithm    string
	Confidence   float64
	Sentences    []string
	Metadata     map[string]interface{}
}

type SummaryHistory struct {
	mu      sync.RWMutex
	entries []SummaryEntry
	maxSize int
}

type SummaryEntry struct {
	Timestamp   string
	OriginalLen int
	SummaryLen  int
	Reduction   float64
	Algorithm   string
}

func NewSummarizationEngine(config EngineConfig) *SummarizationEngine {
	e := &SummarizationEngine{
		algorithms: make(map[string]Summarizer),
		config:     config,
		history:    NewSummaryHistory(1000),
		stats: EngineStats{
			AlgorithmUsage: make(map[string]int64),
		},
	}

	e.registerAlgorithms()

	return e
}

func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		MaxLength:        500,
		MinLength:        50,
		Algorithm:        "textrank",
		EnableHybrid:     true,
		UseLocalLLM:      false,
		LLMEndpoint:      "http://localhost:11434",
		QualityThreshold: 0.7,
	}
}

func (e *SummarizationEngine) registerAlgorithms() {
	e.RegisterSummarizer(&ExtractiveSummarizer{})
	e.RegisterSummarizer(&TFIDFSummarizer{})
	e.RegisterSummarizer(&TextRankSummarizer{})
	e.RegisterSummarizer(&LexRankSummarizer{})
	e.RegisterSummarizer(&LSA{})
}

func (e *SummarizationEngine) RegisterSummarizer(s Summarizer) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.algorithms[s.Name()] = s
}

func (e *SummarizationEngine) Summarize(ctx context.Context, text string, opts SummarizeOptions) (*Summary, error) {
	if opts.MaxLength == 0 {
		opts.MaxLength = e.config.MaxLength
	}
	if opts.MinLength == 0 {
		opts.MinLength = e.config.MinLength
	}

	algoName := e.config.Algorithm
	if opts.Style != "" {
		algoName = opts.Style
	}

	e.mu.RLock()
	algo, ok := e.algorithms[algoName]
	e.mu.RUnlock()

	if !ok {
		algoName = "extractive"
		e.mu.RLock()
		algo = e.algorithms[algoName]
		e.mu.RUnlock()
	}

	summary, err := algo.Summarize(ctx, text, opts)
	if err != nil {
		return nil, err
	}

	e.history.Add(SummaryEntry{
		Timestamp:   "now",
		OriginalLen: len(text),
		SummaryLen:  len(summary.Text),
		Reduction:   summary.Reduction,
		Algorithm:   algoName,
	})

	e.mu.Lock()
	e.stats.TotalSummaries++
	e.stats.TotalInputTokens += int64(summary.InputTokens)
	e.stats.TotalOutputTokens += int64(summary.OutputTokens)
	e.stats.AlgorithmUsage[algoName]++
	e.mu.Unlock()

	return summary, nil
}

func (e *SummarizationEngine) GetStats() EngineStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stats
}

type ExtractiveSummarizer struct{}

func (s *ExtractiveSummarizer) Name() string { return "extractive" }

func (s *ExtractiveSummarizer) Summarize(ctx context.Context, text string, opts SummarizeOptions) (*Summary, error) {
	sentences := s.splitIntoSentences(text)

	if len(sentences) == 0 {
		return &Summary{Text: "", Reduction: 0}, nil
	}

	scored := make([]sentScore, len(sentences))
	for i, sent := range sentences {
		scored[i] = sentScore{
			text:  sent,
			score: s.scoreSentence(sent, sentences),
			index: i,
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	selected := scored[:len(scored)/2+1]
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].index < selected[j].index
	})

	result := &Summary{
		Algorithm:  "extractive",
		Confidence: 0.8,
		Sentences:  make([]string, len(selected)),
		Metadata:   map[string]interface{}{},
	}

	for i, ss := range selected {
		result.Sentences[i] = ss.text
		result.Text += ss.text + " "
	}

	result.Text = strings.TrimSpace(result.Text)
	result.InputTokens = len(text) / 4
	result.OutputTokens = len(result.Text) / 4
	if result.InputTokens > 0 {
		result.Reduction = 1.0 - float64(result.OutputTokens)/float64(result.InputTokens)
	}

	return result, nil
}

func (s *ExtractiveSummarizer) scoreSentence(sent string, all []string) float64 {
	score := float64(len(sent)) / 100.0

	importantPatterns := []string{
		"(?i)(important|critical|essential|key|main|primary)",
		"(?i)(error|fail|issue|problem|bug)",
		"(?i)(success|complete|done|finished)",
	}

	for _, pat := range importantPatterns {
		if regexp.MustCompile(pat).MatchString(sent) {
			score += 0.5
		}
	}

	firstThird := len(all) / 3
	for i, other := range all {
		if i/firstThird == 0 && i != 0 {
			score += float64(len(intersectionWords(sent, other))) * 0.1
		}
	}

	return score
}

func (s *ExtractiveSummarizer) splitIntoSentences(text string) []string {
	sentences := regexp.MustCompile(`[.!?]+\s+`).Split(text, -1)
	result := make([]string, 0, len(sentences))
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if len(s) > 20 {
			result = append(result, s)
		}
	}
	return result
}

type TFIDFSummarizer struct{}

func (s *TFIDFSummarizer) Name() string { return "tfidf" }

func (s *TFIDFSummarizer) Summarize(ctx context.Context, text string, opts SummarizeOptions) (*Summary, error) {
	sentences := s.splitIntoSentences(text)
	if len(sentences) == 0 {
		return &Summary{Text: "", Reduction: 0}, nil
	}

	docFreq := make(map[string]int)
	totalDocs := len(sentences)

	for _, sent := range sentences {
		words := getWords(sent)
		seen := make(map[string]bool)
		for _, w := range words {
			if !seen[w] {
				docFreq[w]++
				seen[w] = true
			}
		}
	}

	scored := make([]sentScore, len(sentences))
	for i, sent := range sentences {
		score := 0.0
		words := getWords(sent)
		seen := make(map[string]bool)
		for _, w := range words {
			if !seen[w] {
				df := docFreq[w]
				if df > 0 {
					idf := math.Log(float64(totalDocs) / float64(df))
					score += idf
				}
				seen[w] = true
			}
		}
		scored[i] = sentScore{text: sent, score: score, index: i}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	selected := scored[:len(scored)/2+1]
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].index < selected[j].index
	})

	result := &Summary{
		Algorithm:  "tfidf",
		Confidence: 0.75,
		Sentences:  make([]string, len(selected)),
		Metadata:   map[string]interface{}{},
	}

	for i, ss := range selected {
		result.Sentences[i] = ss.text
		result.Text += ss.text + " "
	}

	result.Text = strings.TrimSpace(result.Text)
	result.InputTokens = len(text) / 4
	result.OutputTokens = len(result.Text) / 4
	if result.InputTokens > 0 {
		result.Reduction = 1.0 - float64(result.OutputTokens)/float64(result.InputTokens)
	}

	return result, nil
}

func (s *TFIDFSummarizer) splitIntoSentences(text string) []string {
	sentences := regexp.MustCompile(`[.!?]+\s+`).Split(text, -1)
	result := make([]string, 0, len(sentences))
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if len(s) > 20 {
			result = append(result, s)
		}
	}
	return result
}

type TextRankSummarizer struct{}

func (s *TextRankSummarizer) Name() string { return "textrank" }

func (s *TextRankSummarizer) Summarize(ctx context.Context, text string, opts SummarizeOptions) (*Summary, error) {
	sentences := s.splitIntoSentences(text)
	if len(sentences) < 2 {
		return &Summary{Text: text, Reduction: 0, Algorithm: "textrank"}, nil
	}

	similarity := make([][]float64, len(sentences))
	for i := range similarity {
		similarity[i] = make([]float64, len(sentences))
	}

	for i := 0; i < len(sentences); i++ {
		for j := i + 1; j < len(sentences); j++ {
			sim := s.computeSimilarity(sentences[i], sentences[j])
			similarity[i][j] = sim
			similarity[j][i] = sim
		}
	}

	scores := make([]float64, len(sentences))
	for iter := 0; iter < 30; iter++ {
		newScores := make([]float64, len(sentences))
		for i := 0; i < len(sentences); i++ {
			for j := 0; j < len(sentences); j++ {
				if i != j && similarity[i][j] > 0 {
					newScores[i] += similarity[i][j] * scores[j]
				}
			}
		}
		for i := range newScores {
			if math.IsNaN(newScores[i]) {
				newScores[i] = 1.0 / float64(len(sentences))
			}
		}
		scores = newScores
	}

	scored := make([]sentScore, len(sentences))
	for i, score := range scores {
		scored[i] = sentScore{text: sentences[i], score: score, index: i}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	topN := len(sentences) / 2
	if topN < 1 {
		topN = 1
	}
	selected := scored[:topN]
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].index < selected[j].index
	})

	result := &Summary{
		Algorithm:  "textrank",
		Confidence: 0.85,
		Sentences:  make([]string, len(selected)),
		Metadata:   map[string]interface{}{},
	}

	for i, ss := range selected {
		result.Sentences[i] = ss.text
		result.Text += ss.text + " "
	}

	result.Text = strings.TrimSpace(result.Text)
	result.InputTokens = len(text) / 4
	result.OutputTokens = len(result.Text) / 4
	if result.InputTokens > 0 {
		result.Reduction = 1.0 - float64(result.OutputTokens)/float64(result.InputTokens)
	}

	return result, nil
}

func (s *TextRankSummarizer) computeSimilarity(a, b string) float64 {
	wordsA := getWords(a)
	wordsB := getWords(b)

	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0
	}

	match := 0
	for _, wa := range wordsA {
		for _, wb := range wordsB {
			if wa == wb {
				match++
				break
			}
		}
	}

	union := float64(len(wordsA) + len(wordsB) - match)
	if union == 0 {
		return 0
	}

	return float64(match) / union
}

func (s *TextRankSummarizer) splitIntoSentences(text string) []string {
	sentences := regexp.MustCompile(`[.!?]+\s+`).Split(text, -1)
	result := make([]string, 0, len(sentences))
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if len(s) > 20 {
			result = append(result, s)
		}
	}
	return result
}

type LexRankSummarizer struct{}

func (s *LexRankSummarizer) Name() string { return "lexrank" }

func (s *LexRankSummarizer) Summarize(ctx context.Context, text string, opts SummarizeOptions) (*Summary, error) {
	sentences := s.splitIntoSentences(text)
	if len(sentences) < 2 {
		return &Summary{Text: text, Reduction: 0, Algorithm: "lexrank"}, nil
	}

	return &Summary{
		Text:        text[:min(len(text)/2, opts.MaxLength)],
		Algorithm:   "lexrank",
		Confidence:  0.8,
		InputTokens: len(text) / 4,
		Reduction:   0.5,
	}, nil
}

func (s *LexRankSummarizer) splitIntoSentences(text string) []string {
	sentences := regexp.MustCompile(`[.!?]+\s+`).Split(text, -1)
	result := make([]string, 0, len(sentences))
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if len(s) > 20 {
			result = append(result, s)
		}
	}
	return result
}

type LSA struct{}

func (s *LSA) Name() string { return "lsa" }

func (s *LSA) Summarize(ctx context.Context, text string, opts SummarizeOptions) (*Summary, error) {
	sentences := s.splitIntoSentences(text)
	if len(sentences) < 2 {
		return &Summary{Text: text, Reduction: 0, Algorithm: "lsa"}, nil
	}

	words := make([][]float64, len(sentences))
	wordIdx := make(map[string]int)

	for i, sent := range sentences {
		sentWords := getWords(sent)
		words[i] = make([]float64, 100)
		for j, w := range sentWords {
			if j >= 100 {
				break
			}
			if _, ok := wordIdx[w]; !ok {
				wordIdx[w] = len(wordIdx)
			}
			words[i][wordIdx[w]]++
		}
	}

	scores := make([]float64, len(sentences))
	for i := range scores {
		scores[i] = 1.0
	}

	scored := make([]sentScore, len(sentences))
	for i, score := range scores {
		scored[i] = sentScore{text: sentences[i], score: score, index: i}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	selected := scored[:len(scored)/2+1]
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].index < selected[j].index
	})

	result := &Summary{
		Algorithm:  "lsa",
		Confidence: 0.7,
		Sentences:  make([]string, len(selected)),
		Metadata:   map[string]interface{}{},
	}

	for i, ss := range selected {
		result.Sentences[i] = ss.text
		result.Text += ss.text + " "
	}

	result.Text = strings.TrimSpace(result.Text)
	result.InputTokens = len(text) / 4
	result.OutputTokens = len(result.Text) / 4
	if result.InputTokens > 0 {
		result.Reduction = 1.0 - float64(result.OutputTokens)/float64(result.InputTokens)
	}

	return result, nil
}

func (s *LSA) splitIntoSentences(text string) []string {
	sentences := regexp.MustCompile(`[.!?]+\s+`).Split(text, -1)
	result := make([]string, 0, len(sentences))
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if len(s) > 20 {
			result = append(result, s)
		}
	}
	return result
}

type sentScore struct {
	text  string
	score float64
	index int
}

func NewSummaryHistory(maxSize int) *SummaryHistory {
	return &SummaryHistory{
		entries: make([]SummaryEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

func (h *SummaryHistory) Add(entry SummaryEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = append(h.entries, entry)
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[1:]
	}
}

func getWords(s string) []string {
	re := regexp.MustCompile(`[a-zA-Z]+`)
	return re.FindAllString(s, -1)
}

func intersectionWords(a, b string) []string {
	wordsA := getWords(a)
	wordsB := getWords(b)

	result := make([]string, 0)
	seen := make(map[string]bool)

	for _, wa := range wordsA {
		if seen[wa] {
			continue
		}
		for _, wb := range wordsB {
			if wa == wb {
				result = append(result, wa)
				seen[wa] = true
				break
			}
		}
	}

	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
