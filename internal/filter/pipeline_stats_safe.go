package filter

import (
	"sync"
	"sync/atomic"
)

// SafePipelineStats provides thread-safe pipeline statistics.
// Replaces the unsafe PipelineStats in concurrent contexts.
type SafePipelineStats struct {
	mu             sync.RWMutex
	layerStats     map[string]LayerStat
	runningSaved   int64
	originalTokens int
	finalTokens    int
	totalSaved     int
	reductionPct   float64
}

// NewSafePipelineStats creates a new thread-safe stats tracker.
func NewSafePipelineStats(originalTokens int) *SafePipelineStats {
	return &SafePipelineStats{
		layerStats:     make(map[string]LayerStat),
		originalTokens: originalTokens,
	}
}

// AddLayerStat records a layer's statistics thread-safely.
func (s *SafePipelineStats) AddLayerStat(name string, stat LayerStat) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.layerStats == nil {
		s.layerStats = make(map[string]LayerStat)
	}
	s.layerStats[name] = stat
	atomic.AddInt64(&s.runningSaved, int64(stat.TokensSaved))
}

// RunningSaved returns the current tokens saved (thread-safe).
func (s *SafePipelineStats) RunningSaved() int {
	return int(atomic.LoadInt64(&s.runningSaved))
}

// ToPipelineStats converts to immutable PipelineStats for reporting.
func (s *SafePipelineStats) ToPipelineStats() *PipelineStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy of layer stats
	layerStatsCopy := make(map[string]LayerStat, len(s.layerStats))
	for k, v := range s.layerStats {
		layerStatsCopy[k] = v
	}

	return &PipelineStats{
		OriginalTokens:   s.originalTokens,
		FinalTokens:      s.finalTokens,
		TotalSaved:       s.totalSaved,
		ReductionPercent: s.reductionPct,
		LayerStats:       layerStatsCopy,
		runningSaved:     int(s.runningSaved),
	}
}

// SetFinalResult sets the final compression results.
func (s *SafePipelineStats) SetFinalResult(finalTokens, totalSaved int, reductionPct float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.finalTokens = finalTokens
	s.totalSaved = totalSaved
	s.reductionPct = reductionPct
}
