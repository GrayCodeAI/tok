package filter

import (
	"strings"
	"sync"
)

// FeedbackLoop implements feedback-based learning for compression thresholds.
type FeedbackLoop struct {
	mu          sync.RWMutex
	thresholds  map[string]float64
	adjustments map[string]float64
	samples     map[string][]float64
}

// NewFeedbackLoop creates a new feedback loop learner.
func NewFeedbackLoop() *FeedbackLoop {
	return &FeedbackLoop{
		thresholds:  make(map[string]float64),
		adjustments: make(map[string]float64),
		samples:     make(map[string][]float64),
	}
}

// Record records a feedback sample for a language/content type.
func (fl *FeedbackLoop) Record(key string, quality float64) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.samples[key] = append(fl.samples[key], quality)
	if len(fl.samples[key]) > 100 {
		fl.samples[key] = fl.samples[key][len(fl.samples[key])-100:]
	}
	if quality < 0.7 {
		fl.adjustments[key] -= 0.05
	} else if quality > 0.9 {
		fl.adjustments[key] += 0.02
	}
}

// GetThreshold returns the learned threshold for a key.
func (fl *FeedbackLoop) GetThreshold(key string, base float64) float64 {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	return base + fl.adjustments[key]
}

// InformationBottleneck filters content by entropy and task-relevance.
type InformationBottleneck struct {
	config IBConfig
}

// IBConfig holds configuration for information bottleneck.
type IBConfig struct {
	Enabled            bool
	EntropyThreshold   float64
	RelevanceThreshold float64
}

// DefaultIBConfig returns default IB configuration.
func DefaultIBConfig() IBConfig {
	return IBConfig{Enabled: true, EntropyThreshold: 0.5, RelevanceThreshold: 0.3}
}

// NewInformationBottleneck creates a new information bottleneck filter.
func NewInformationBottleneck(cfg IBConfig) *InformationBottleneck {
	return &InformationBottleneck{config: cfg}
}

// Process filters content by information bottleneck principle.
func (ib *InformationBottleneck) Process(content, query string) string {
	if !ib.config.Enabled {
		return content
	}
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		entropy := lineEntropy(line)
		relevance := lineRelevance(line, query)
		if entropy > ib.config.EntropyThreshold || relevance > ib.config.RelevanceThreshold {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

func lineEntropy(line string) float64 {
	if len(line) == 0 {
		return 0
	}
	freq := make(map[rune]int)
	for _, r := range line {
		freq[r]++
	}
	var entropy float64
	n := float64(len(line))
	for _, count := range freq {
		p := float64(count) / n
		if p > 0 {
			entropy -= p * log2f(p)
		}
	}
	maxEntropy := log2f(float64(len(freq)))
	if maxEntropy == 0 {
		return 0
	}
	return entropy / maxEntropy
}

func lineRelevance(line, query string) float64 {
	if query == "" {
		return 0.5
	}
	queryWords := strings.Fields(strings.ToLower(query))
	lineLower := strings.ToLower(line)
	matches := 0
	for _, w := range queryWords {
		if strings.Contains(lineLower, w) {
			matches++
		}
	}
	return float64(matches) / float64(len(queryWords))
}

func log2f(x float64) float64 {
	if x <= 0 {
		return 0
	}
	result := 0.0
	for x > 1 {
		x /= 2.71828
		result++
	}
	return result + (x - 1)
}
