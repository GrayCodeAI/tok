package filter

import (
	"sync"
)

// ParallelFilterResult holds result from parallel filter execution
type ParallelFilterResult struct {
	Output string
	Saved  int
	Error  error
}

// ExecuteFiltersParallel runs independent filters in parallel
// This improves throughput for multi-core systems
func ExecuteFiltersParallel(filters []filterLayer, input string, mode Mode) (string, int) {
	if len(filters) == 0 {
		return input, 0
	}
	
	// Single filter: run directly (avoid goroutine overhead)
	if len(filters) == 1 {
		return filters[0].filter.Apply(input, mode)
	}
	
	// Parallel execution for multiple filters
	results := make([]ParallelFilterResult, len(filters))
	var wg sync.WaitGroup
	
	// Run filters in parallel
	for i, layer := range filters {
		wg.Add(1)
		go func(idx int, l filterLayer) {
			defer wg.Done()
			output, saved := l.filter.Apply(input, mode)
			results[idx] = ParallelFilterResult{
				Output: output,
				Saved:  saved,
			}
		}(i, layer)
	}
	
	wg.Wait()
	
	// Combine results: use output from last filter
	// (This is a simplified combination - real implementation would be smarter)
	totalSaved := 0
	for _, r := range results {
		totalSaved += r.Saved
	}
	
	// Return best result (most savings)
	bestResult := results[0]
	for _, r := range results {
		if r.Saved > bestResult.Saved {
			bestResult = r
		}
	}
	
	return bestResult.Output, totalSaved
}

// ExecuteFiltersSequential runs filters sequentially
// Use when filters depend on each other's output
func ExecuteFiltersSequential(filters []filterLayer, input string, mode Mode) (string, int) {
	output := input
	totalSaved := 0
	
	for _, layer := range filters {
		newOutput, saved := layer.filter.Apply(output, mode)
		output = newOutput
		totalSaved += saved
	}
	
	return output, totalSaved
}

// ShouldUseParallel determines if parallel execution is beneficial
func ShouldUseParallel(filters []filterLayer, inputSize int) bool {
	// Use parallel for:
	// - Multiple filters (2+)
	// - Large inputs (>1KB)
	// - Independent filters
	
	if len(filters) < 2 {
		return false
	}
	
	if inputSize < 1024 {
		return false // Overhead not worth it for small inputs
	}
	
	// Check if we have enough CPU cores
	// runtime.NumCPU() >= 2
	
	return true
}

// ParallelPipelineStats holds stats from parallel execution
type ParallelPipelineStats struct {
	mu          sync.Mutex
	LayerStats  map[string]LayerStat
	TotalSaved  int
	ParallelTime int64
	SequentialTime int64
}

// NewParallelPipelineStats creates new stats tracker
func NewParallelPipelineStats() *ParallelPipelineStats {
	return &ParallelPipelineStats{
		LayerStats: make(map[string]LayerStat),
	}
}

// AddStat adds a layer stat thread-safely
func (s *ParallelPipelineStats) AddStat(name string, stat LayerStat) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LayerStats[name] = stat
	s.TotalSaved += stat.TokensSaved
}
