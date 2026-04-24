package filter

import "sync"

// ParallelExecutor runs independent layers concurrently
type ParallelExecutor struct {
	pool *sync.Pool
}

// NewParallelExecutor creates a parallel execution engine
func NewParallelExecutor() *ParallelExecutor {
	return &ParallelExecutor{
		pool: &sync.Pool{
			New: func() interface{} {
				return &layerResult{}
			},
		},
	}
}

type layerResult struct {
	output string
	tokens int
}

// ExecuteParallel runs layers concurrently and merges results
func (pe *ParallelExecutor) ExecuteParallel(input string, layers []Filter) (string, int) {
	if len(layers) == 0 {
		return input, 0
	}

	results := make([]*layerResult, len(layers))
	var wg sync.WaitGroup

	for i, layer := range layers {
		wg.Add(1)
		go func(idx int, l Filter) {
			defer wg.Done()
			r := pe.pool.Get().(*layerResult)
			r.output, r.tokens = l.Apply(input, ModeMinimal)
			results[idx] = r
		}(i, layer)
	}

	wg.Wait()

	// Use best result (highest compression)
	best := results[0]
	for _, r := range results[1:] {
		if len(r.output) < len(best.output) {
			best = r
		}
	}

	// Copy before returning to pool — best points into results, which we Put below.
	bestOutput, bestTokens := best.output, best.tokens

	for _, r := range results {
		pe.pool.Put(r)
	}

	return bestOutput, bestTokens
}
