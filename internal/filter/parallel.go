package filter

import (
	"context"
	"runtime"
	"sync"
)

// ParallelFilterResult holds result from parallel filter execution
type ParallelFilterResult struct {
	Output string
	Saved  int
	Error  error
}

// ExecuteFiltersParallel runs filters in parallel on the same input.
// WARNING: This is only safe for truly independent filters that do not depend
// on each other's output. For sequential filter chains (where each layer
// transforms the output of the previous layer), use ExecuteFiltersSequential.
//
// Historically this function attempted to run sequential filters in parallel,
// which produced semantically incorrect output because later filters received
// the original input instead of the transformed output from previous layers.
// It now safely delegates to sequential execution unless all filters are
// explicitly marked as independent (future enhancement).
func ExecuteFiltersParallel(filters []filterLayer, input string, mode Mode) (string, int) {
	// Sequential execution is the only safe default for filter chains where
	// each layer depends on the previous layer's output.
	return ExecuteFiltersSequential(filters, input, mode)
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
	mu             sync.Mutex
	LayerStats     map[string]LayerStat
	TotalSaved     int
	ParallelTime   int64
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

// ParallelProcessor handles parallel compression of multiple inputs
// Uses worker pool pattern for optimal CPU utilization
type ParallelProcessor struct {
	workers int
	sem     chan struct{}
}

// NewParallelProcessor creates a new parallel processor
// Automatically determines optimal worker count based on CPU cores
func NewParallelProcessor() *ParallelProcessor {
	workers := runtime.NumCPU()
	if workers < 4 {
		workers = 4
	}
	return &ParallelProcessor{
		workers: workers,
		sem:     make(chan struct{}, workers),
	}
}

// NewParallelProcessorWithWorkers creates processor with specific worker count
func NewParallelProcessorWithWorkers(workers int) *ParallelProcessor {
	if workers < 1 {
		workers = 1
	}
	if workers > runtime.NumCPU()*4 {
		workers = runtime.NumCPU() * 4
	}
	return &ParallelProcessor{
		workers: workers,
		sem:     make(chan struct{}, workers),
	}
}

// ProcessItems processes multiple items in parallel
// Each item is processed by the provided function
func (p *ParallelProcessor) ProcessItems(items []string, processFn func(string) (string, int)) []ParallelProcessResult {
	if len(items) == 0 {
		return nil
	}

	// For small batches, process sequentially (avoid overhead)
	if len(items) < 4 {
		results := make([]ParallelProcessResult, len(items))
		for i, item := range items {
			output, saved := processFn(item)
			results[i] = ParallelProcessResult{
				Input:  item,
				Output: output,
				Saved:  saved,
				Index:  i,
			}
		}
		return results
	}

	// Process in parallel
	results := make([]ParallelProcessResult, len(items))
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		p.sem <- struct{}{} // Acquire semaphore

		go func(index int, input string) {
			defer wg.Done()
			defer func() { <-p.sem }() // Release semaphore

			output, saved := processFn(input)
			results[index] = ParallelProcessResult{
				Input:  input,
				Output: output,
				Saved:  saved,
				Index:  index,
			}
		}(i, item)
	}

	wg.Wait()
	return results
}

// ProcessItemsContext processes items with context cancellation support
func (p *ParallelProcessor) ProcessItemsContext(ctx context.Context, items []string, processFn func(context.Context, string) (string, int)) ([]ParallelProcessResult, error) {
	if len(items) == 0 {
		return nil, nil
	}

	// Check context before starting
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	results := make([]ParallelProcessResult, len(items))
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for i, item := range items {
		wg.Add(1)

		go func(index int, input string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				select {
				case errCh <- ctx.Err():
				default:
				}
				return
			default:
			}

			output, saved := processFn(ctx, input)
			results[index] = ParallelProcessResult{
				Input:  input,
				Output: output,
				Saved:  saved,
				Index:  index,
			}
		}(i, item)
	}

	wg.Wait()
	close(errCh)

	if err := <-errCh; err != nil {
		return results, err
	}
	return results, nil
}

// ParallelProcessResult holds the result of a parallel processing operation
type ParallelProcessResult struct {
	Input  string
	Output string
	Saved  int
	Index  int
}

// TotalSaved calculates total tokens saved from results
func TotalSaved(results []ParallelProcessResult) int {
	total := 0
	for _, r := range results {
		total += r.Saved
	}
	return total
}

// CollectOutputs collects all outputs from results in order
func CollectOutputs(results []ParallelProcessResult) []string {
	outputs := make([]string, len(results))
	for _, r := range results {
		outputs[r.Index] = r.Output
	}
	return outputs
}

// ParallelCompressor provides high-level parallel compression interface
type ParallelCompressor struct {
	processor *ParallelProcessor
	engine    *PipelineCoordinator
}

// NewParallelCompressor creates a new parallel compressor
func NewParallelCompressor(config PipelineConfig) *ParallelCompressor {
	return &ParallelCompressor{
		processor: NewParallelProcessor(),
		engine:    NewPipelineCoordinator(config),
	}
}

// Compress compresses a single input
func (pc *ParallelCompressor) Compress(input string) (string, int) {
	output, stats, err := pc.engine.Process(input)
	if err != nil {
		return input, 0
	}
	if stats == nil {
		return output, 0
	}
	return output, stats.TotalSaved
}

// CompressBatch compresses multiple inputs in parallel
func (pc *ParallelCompressor) CompressBatch(inputs []string) []ParallelProcessResult {
	return pc.processor.ProcessItems(inputs, func(item string) (string, int) {
		output, stats, err := pc.engine.Process(item)
		if err != nil {
			return item, 0
		}
		if stats == nil {
			return output, 0
		}
		return output, stats.TotalSaved
	})
}

// CompressBatchContext compresses with context support
func (pc *ParallelCompressor) CompressBatchContext(ctx context.Context, inputs []string) ([]ParallelProcessResult, error) {
	return pc.processor.ProcessItemsContext(ctx, inputs, func(ctx context.Context, item string) (string, int) {
		output, stats, err := pc.engine.Process(item)
		if err != nil {
			return item, 0
		}
		if stats == nil {
			return output, 0
		}
		return output, stats.TotalSaved
	})
}

// WorkerCount returns the number of workers
func (pc *ParallelCompressor) WorkerCount() int {
	return pc.processor.workers
}
