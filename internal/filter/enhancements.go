package filter

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
	// TODO: compute diff
	dc.previous = current
	return current
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
