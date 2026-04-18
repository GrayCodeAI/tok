package core

import "sync"

// Processor interface for batch processing
type Processor interface {
	Process(string) (string, interface{})
}

// BatchProcessor processes multiple inputs concurrently
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

func (bp *BatchProcessor) ProcessBatch(inputs []string) []BatchResult {
	results := make([]BatchResult, len(inputs))
	jobs := make(chan int, len(inputs))
	var wg sync.WaitGroup
	
	for w := 0; w < bp.workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				output, stats := bp.coordinator.Process(inputs[idx])
				results[idx] = BatchResult{
					Index:  idx,
					Output: output,
					Stats:  stats,
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
