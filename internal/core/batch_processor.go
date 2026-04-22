package core

import "sync"

// Processor interface for batch processing
// The error return allows callers to detect failures per item.
type Processor interface {
	Process(string) (string, interface{}, error)
}

// BatchProcessor processes multiple inputs concurrently.
type BatchProcessor struct {
	coordinator Processor
	workers     int
}

func NewBatchProcessor(coordinator Processor, workers int) *BatchProcessor {
	return &BatchProcessor{coordinator: coordinator, workers: workers}
}

type BatchResult struct {
	Index  int
	Output string
	Stats  interface{}
	Error  error
}

// ProcessBatch runs the processor on each input concurrently.
// Errors from the processor are captured in BatchResult.Error.
func (bp *BatchProcessor) ProcessBatch(inputs []string) []BatchResult {
	results := make([]BatchResult, len(inputs))
	jobs := make(chan int, len(inputs))
	var wg sync.WaitGroup

	for w := 0; w < bp.workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				output, stats, err := bp.coordinator.Process(inputs[idx])
				results[idx] = BatchResult{
					Index:  idx,
					Output: output,
					Stats:  stats,
					Error:  err,
				}
			}
		}()
	}

	for i := range inputs {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	return results
}
