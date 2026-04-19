package filter

import (
	"fmt"
	"strings"
)

// AdaptiveBufferSizer dynamically adjusts buffer sizes
type AdaptiveBufferSizer struct {
	minSize int
	maxSize int
	current int
}

func NewAdaptiveBufferSizer(min, max int) *AdaptiveBufferSizer {
	return &AdaptiveBufferSizer{minSize: min, maxSize: max, current: min}
}

func (abs *AdaptiveBufferSizer) Adjust(inputSize int) int {
	if inputSize > abs.current {
		abs.current = minInt(inputSize*2, abs.maxSize)
	}
	return abs.current
}

// CircuitBreaker stops processing on quality drop
type CircuitBreaker struct {
	threshold float64
	failures  int
	maxFails  int
}

func NewCircuitBreaker(threshold float64, maxFails int) *CircuitBreaker {
	return &CircuitBreaker{threshold: threshold, maxFails: maxFails}
}

func (cb *CircuitBreaker) Check(quality float64) bool {
	if quality < cb.threshold {
		cb.failures++
		return cb.failures < cb.maxFails
	}
	cb.failures = 0
	return true
}

// PerplexityOptimizer uses beam search
type PerplexityOptimizer struct {
	beamWidth int
}

func NewPerplexityOptimizer(width int) *PerplexityOptimizer {
	return &PerplexityOptimizer{beamWidth: width}
}

// Optimize ranks tokens by frequency; common words have lower perplexity
// and can be dropped during compression. Returns tokens sorted by
// ascending frequency (rarest first, highest perplexity first).
func (po *PerplexityOptimizer) Optimize(tokens []string) []string {
	freq := make(map[string]int)
	for _, t := range tokens {
		freq[strings.ToLower(t)]++
	}
	type tokenScore struct {
		token string
		score int
	}
	var scored []tokenScore
	for _, t := range tokens {
		scored = append(scored, tokenScore{token: t, score: freq[strings.ToLower(t)]})
	}
	// Sort by score ascending (rare tokens first = higher perplexity = keep)
	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score < scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}
	result := make([]string, len(scored))
	for i, s := range scored {
		result[i] = s.token
	}
	return result
}

// HeatmapGenerator visualizes compression
type HeatmapGenerator struct {
	data map[int]float64
}

func NewHeatmapGenerator() *HeatmapGenerator {
	return &HeatmapGenerator{data: make(map[int]float64)}
}

func (hg *HeatmapGenerator) Record(pos int, ratio float64) {
	hg.data[pos] = ratio
}

// Output returns the heatmap as a formatted string.
func (hg *HeatmapGenerator) Output() string {
	var sb strings.Builder
	sb.WriteString("Compression Heatmap:\n")
	for pos, ratio := range hg.data {
		sb.WriteString(fmt.Sprintf("  pos %d: %.2f\n", pos, ratio))
	}
	return sb.String()
}

// DifferentialCompressor compresses deltas
type DifferentialCompressor struct {
	previous string
}

func NewDifferentialCompressor() *DifferentialCompressor {
	return &DifferentialCompressor{}
}

func (dc *DifferentialCompressor) Compress(current string) string {
	if dc.previous == "" {
		dc.previous = current
		return current
	}
	// Simple line-level diff: keep only changed lines
	origLines := strings.Split(dc.previous, "\n")
	currLines := strings.Split(current, "\n")
	origSet := make(map[string]bool)
	for _, line := range origLines {
		origSet[strings.TrimSpace(line)] = true
	}
	var changed []string
	for _, line := range currLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !origSet[trimmed] {
			changed = append(changed, line)
		}
	}
	dc.previous = current
	return strings.Join(changed, "\n")
}

// RatioPredictor predicts compression ratio
type RatioPredictor struct {
	history []float64
}

func NewRatioPredictor() *RatioPredictor {
	return &RatioPredictor{history: make([]float64, 0, 100)}
}

func (rp *RatioPredictor) Predict() float64 {
	if len(rp.history) == 0 {
		return 0.5
	}
	sum := 0.0
	for _, v := range rp.history {
		sum += v
	}
	return sum / float64(len(rp.history))
}

func (rp *RatioPredictor) Learn(ratio float64) {
	rp.history = append(rp.history, ratio)
	if len(rp.history) > 100 {
		rp.history = rp.history[1:]
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
