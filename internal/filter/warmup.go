package filter

import "sync/atomic"

// Warmup pre-initializes pipeline components
func (p *PipelineCoordinator) Warmup() {
	dummy := "warmup test data"
	p.Process(dummy)
}

// LockFreeCounter atomic counter
type LockFreeCounter struct {
	value uint64
}

func (c *LockFreeCounter) Inc() {
	atomic.AddUint64(&c.value, 1)
}

func (c *LockFreeCounter) Get() uint64 {
	return atomic.LoadUint64(&c.value)
}

// StreamingResult for incremental output
type StreamingResult struct {
	chunk chan string
	done  chan struct{}
}

func NewStreamingResult() *StreamingResult {
	return &StreamingResult{
		chunk: make(chan string, 10),
		done:  make(chan struct{}),
	}
}

func (sr *StreamingResult) Send(s string) {
	sr.chunk <- s
}

func (sr *StreamingResult) Close() {
	close(sr.done)
	close(sr.chunk)
}

func (sr *StreamingResult) Receive() <-chan string {
	return sr.chunk
}
