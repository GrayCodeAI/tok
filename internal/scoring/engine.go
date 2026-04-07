package scoring

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// ScoringEngine provides semantic signal scoring for content
type ScoringEngine struct {
	config    ScoringConfig
	weights   ScoringWeights
	keywords  map[string]float64
	userPrefs *UserPreferences
	mu        sync.RWMutex
}

// ScoringConfig holds engine configuration
type ScoringConfig struct {
	EnablePositionScoring  bool
	EnableKeywordScoring   bool
	EnableFrequencyScoring bool
	EnableRecencyScoring   bool
	EnableSemanticScoring  bool
	EnableQueryAware       bool
	DefaultTier            SignalTier
	Threshold              float64
}

// DefaultScoringConfig returns default configuration
func DefaultScoringConfig() ScoringConfig {
	return ScoringConfig{
		EnablePositionScoring:  true,
		EnableKeywordScoring:   true,
		EnableFrequencyScoring: true,
		EnableRecencyScoring:   true,
		EnableSemanticScoring:  true,
		EnableQueryAware:       true,
		DefaultTier:            TierImportant,
		Threshold:              0.5,
	}
}

// ScoringWeights holds configurable weights for different scoring factors
type ScoringWeights struct {
	Position  float64 `json:"position"`
	Keyword   float64 `json:"keyword"`
	Frequency float64 `json:"frequency"`
	Recency   float64 `json:"recency"`
	Semantic  float64 `json:"semantic"`
	Query     float64 `json:"query"`
}

// DefaultScoringWeights returns default weights
func DefaultScoringWeights() ScoringWeights {
	return ScoringWeights{
		Position:  0.15,
		Keyword:   0.25,
		Frequency: 0.20,
		Recency:   0.15,
		Semantic:  0.15,
		Query:     0.10,
	}
}

// SignalTier represents the importance tier of a signal
type SignalTier string

const (
	TierCritical   SignalTier = "critical"
	TierImportant  SignalTier = "important"
	TierNiceToHave SignalTier = "nice_to_have"
	TierNoise      SignalTier = "noise"
)

// NewScoringEngine creates a new scoring engine
func NewScoringEngine() *ScoringEngine {
	return NewScoringEngineWithConfig(DefaultScoringConfig())
}

// NewScoringEngineWithConfig creates engine with custom config
func NewScoringEngineWithConfig(config ScoringConfig) *ScoringEngine {
	return &ScoringEngine{
		config:    config,
		weights:   DefaultScoringWeights(),
		keywords:  make(map[string]float64),
		userPrefs: NewUserPreferences(),
	}
}

// SetWeights updates scoring weights
func (se *ScoringEngine) SetWeights(weights ScoringWeights) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.weights = weights
}

// AddKeyword adds a keyword with its importance score
func (se *ScoringEngine) AddKeyword(keyword string, score float64) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.keywords[strings.ToLower(keyword)] = score
}

// ScoreContent scores a piece of content
func (se *ScoringEngine) ScoreContent(content string, opts ScoringOptions) *ScoredContent {
	lines := strings.Split(content, "\n")
	scoredLines := make([]*ScoredLine, 0, len(lines))

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		score := se.calculateLineScore(line, i, len(lines), opts)

		scoredLines = append(scoredLines, &ScoredLine{
			LineNumber: i + 1,
			Content:    line,
			Score:      score,
			Tier:       se.determineTier(score),
		})
	}

	// Sort by score (descending)
	sort.Slice(scoredLines, func(i, j int) bool {
		return scoredLines[i].Score > scoredLines[j].Score
	})

	// Calculate aggregate statistics
	var totalScore, avgScore float64
	tierCounts := make(map[SignalTier]int)

	for _, line := range scoredLines {
		totalScore += line.Score
		tierCounts[line.Tier]++
	}

	if len(scoredLines) > 0 {
		avgScore = totalScore / float64(len(scoredLines))
	}

	return &ScoredContent{
		Lines:      scoredLines,
		TotalLines: len(lines),
		AvgScore:   avgScore,
		MaxScore:   scoredLines[0].Score,
		MinScore:   scoredLines[len(scoredLines)-1].Score,
		TierCounts: tierCounts,
	}
}

// calculateLineScore calculates the score for a single line
func (se *ScoringEngine) calculateLineScore(line string, index, total int, opts ScoringOptions) float64 {
	var score float64

	// Position-based scoring (higher at beginning and end)
	if se.config.EnablePositionScoring {
		score += se.calculatePositionScore(index, total) * se.weights.Position
	}

	// Keyword-based scoring
	if se.config.EnableKeywordScoring {
		score += se.calculateKeywordScore(line) * se.weights.Keyword
	}

	// Frequency-based scoring
	if se.config.EnableFrequencyScoring && opts.DocumentFreq != nil {
		score += se.calculateFrequencyScore(line, opts.DocumentFreq) * se.weights.Frequency
	}

	// Recency scoring
	if se.config.EnableRecencyScoring && opts.Timestamp != nil {
		score += se.calculateRecencyScore(*opts.Timestamp) * se.weights.Recency
	}

	// Semantic similarity scoring
	if se.config.EnableSemanticScoring && opts.Query != "" {
		score += se.calculateSemanticScore(line, opts.Query) * se.weights.Semantic
	}

	// Query-aware scoring
	if se.config.EnableQueryAware && opts.Query != "" {
		score += se.calculateQueryAwareScore(line, opts.Query) * se.weights.Query
	}

	// Apply tier boost
	tierBoost := se.getTierBoost(line)
	score *= (1 + tierBoost)

	return math.Min(score, 1.0)
}

// calculatePositionScore scores based on position in document
func (se *ScoringEngine) calculatePositionScore(index, total int) float64 {
	if total <= 1 {
		return 1.0
	}

	// Normalize position to 0-1
	pos := float64(index) / float64(total-1)

	// Higher scores at beginning (0) and end (1), lower in middle (0.5)
	distance := math.Abs(pos-0.5) * 2
	return 0.3 + 0.7*distance
}

// calculateKeywordScore scores based on important keywords
func (se *ScoringEngine) calculateKeywordScore(line string) float64 {
	se.mu.RLock()
	defer se.mu.RUnlock()

	lineLower := strings.ToLower(line)
	var score float64

	for keyword, weight := range se.keywords {
		if strings.Contains(lineLower, keyword) {
			score += weight
		}
	}

	// Also check for common important patterns
	importantPatterns := []string{
		"error", "fail", "panic", "exception",
		"warning", "deprecated",
		"todo", "fixme",
		"func ", "class ", "def ",
		"import", "from ",
		"return", "yield",
	}

	for _, pattern := range importantPatterns {
		if strings.Contains(lineLower, pattern) {
			score += 0.1
		}
	}

	return math.Min(score, 1.0)
}

// calculateFrequencyScore scores based on term frequency
func (se *ScoringEngine) calculateFrequencyScore(line string, docFreq map[string]int) float64 {
	words := tokenize(line)
	if len(words) == 0 {
		return 0
	}

	var score float64
	for _, word := range words {
		if freq, ok := docFreq[word]; ok {
			// TF-IDF-like scoring
			score += 1.0 / (1.0 + math.Log(float64(freq)))
		}
	}

	return math.Min(score/float64(len(words)), 1.0)
}

// calculateRecencyScore scores based on recency
func (se *ScoringEngine) calculateRecencyScore(timestamp time.Time) float64 {
	age := time.Since(timestamp)

	// Exponential decay over 24 hours
	hours := age.Hours()
	if hours > 24 {
		return 0.1
	}

	return math.Exp(-hours / 12.0)
}

// calculateSemanticScore calculates semantic similarity
func (se *ScoringEngine) calculateSemanticScore(line, query string) float64 {
	// Simple word overlap similarity
	lineWords := set(tokenize(line))
	queryWords := set(tokenize(query))

	if len(lineWords) == 0 || len(queryWords) == 0 {
		return 0
	}

	// Calculate Jaccard similarity
	intersection := 0
	for word := range lineWords {
		if queryWords[word] {
			intersection++
		}
	}

	union := len(lineWords) + len(queryWords) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

// calculateQueryAwareScore scores based on query relevance
func (se *ScoringEngine) calculateQueryAwareScore(line, query string) float64 {
	queryLower := strings.ToLower(query)
	lineLower := strings.ToLower(line)

	// Exact match
	if strings.Contains(lineLower, queryLower) {
		return 1.0
	}

	// Partial matches
	queryWords := tokenize(queryLower)
	if len(queryWords) == 0 {
		return 0
	}

	matches := 0
	for _, word := range queryWords {
		if strings.Contains(lineLower, word) {
			matches++
		}
	}

	return float64(matches) / float64(len(queryWords))
}

// getTierBoost returns a boost factor based on content patterns
func (se *ScoringEngine) getTierBoost(line string) float64 {
	// Critical patterns
	criticalPatterns := []string{
		`(?i)error[:\s]`,
		`(?i)fail(?:ed|ure|ing)`,
		`(?i)panic`,
		`(?i)fatal`,
	}

	for _, pattern := range criticalPatterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return 0.5 // 50% boost for critical content
		}
	}

	// Important patterns
	importantPatterns := []string{
		`(?i)warning`,
		`(?i)deprecated`,
		`(?i)todo\b`,
		`(?i)fixme\b`,
	}

	for _, pattern := range importantPatterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return 0.2 // 20% boost for important content
		}
	}

	return 0
}

// determineTier determines the tier based on score
func (se *ScoringEngine) determineTier(score float64) SignalTier {
	switch {
	case score >= 0.85:
		return TierCritical
	case score >= 0.65:
		return TierImportant
	case score >= se.config.Threshold:
		return TierNiceToHave
	default:
		return TierNoise
	}
}

// FilterByTier returns only lines matching the specified tier or higher
func (sc *ScoredContent) FilterByTier(minTier SignalTier) []*ScoredLine {
	tierOrder := map[SignalTier]int{
		TierCritical:   4,
		TierImportant:  3,
		TierNiceToHave: 2,
		TierNoise:      1,
	}

	minTierLevel := tierOrder[minTier]

	var filtered []*ScoredLine
	for _, line := range sc.Lines {
		if tierOrder[line.Tier] >= minTierLevel {
			filtered = append(filtered, line)
		}
	}

	return filtered
}

// GetTopN returns the top N highest-scored lines
func (sc *ScoredContent) GetTopN(n int) []*ScoredLine {
	if n >= len(sc.Lines) {
		return sc.Lines
	}
	return sc.Lines[:n]
}

// tokenize splits text into words
func tokenize(text string) []string {
	// Simple tokenization
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	words := re.Split(strings.ToLower(text), -1)

	var result []string
	for _, word := range words {
		if len(word) > 2 {
			result = append(result, word)
		}
	}

	return result
}

// set creates a set from a slice
func set(items []string) map[string]bool {
	s := make(map[string]bool)
	for _, item := range items {
		s[item] = true
	}
	return s
}
