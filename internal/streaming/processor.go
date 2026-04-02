package streaming

import (
	"bufio"
	"strings"
)

type StreamingProcessor struct {
	chunkSize int
	threshold int
	adaptive  bool
}

func NewStreamingProcessor() *StreamingProcessor {
	return &StreamingProcessor{
		chunkSize: 4096,
		threshold: 500000,
		adaptive:  true,
	}
}

func (sp *StreamingProcessor) ShouldStream(input string) bool {
	tokens := len(input) / 4
	if sp.adaptive {
		sp.threshold = tokens / 2
	}
	return tokens > sp.threshold
}

func (sp *StreamingProcessor) Process(input string, processor func(string) string) string {
	if !sp.ShouldStream(input) {
		return processor(input)
	}

	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Buffer(make([]byte, sp.chunkSize), sp.chunkSize)

	for scanner.Scan() {
		line := scanner.Text()
		processed := processor(line)
		result.WriteString(processed)
		result.WriteString("\n")
	}

	return result.String()
}

func (sp *StreamingProcessor) SetChunkSize(size int) {
	sp.chunkSize = size
}

func (sp *StreamingProcessor) SetThreshold(threshold int) {
	sp.threshold = threshold
}

func (sp *StreamingProcessor) SetAdaptive(adaptive bool) {
	sp.adaptive = adaptive
}

type StreamingMetrics struct {
	TotalChunks    int     `json:"total_chunks"`
	ProcessedBytes int     `json:"processed_bytes"`
	DurationMs     float64 `json:"duration_ms"`
	ThroughputMBs  float64 `json:"throughput_mbs"`
}

func (sp *StreamingProcessor) ProcessWithMetrics(input string, processor func(string) string) (string, *StreamingMetrics) {
	if !sp.ShouldStream(input) {
		result := processor(input)
		return result, &StreamingMetrics{
			TotalChunks:    1,
			ProcessedBytes: len(input),
		}
	}

	var result strings.Builder
	var metrics StreamingMetrics
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Buffer(make([]byte, sp.chunkSize), sp.chunkSize)

	for scanner.Scan() {
		line := scanner.Text()
		processed := processor(line)
		result.WriteString(processed)
		result.WriteString("\n")
		metrics.TotalChunks++
		metrics.ProcessedBytes += len(line)
	}

	return result.String(), &metrics
}

type BackpressureController struct {
	maxPending int
	pending    int
}

func NewBackpressureController(maxPending int) *BackpressureController {
	return &BackpressureController{maxPending: maxPending}
}

func (b *BackpressureController) TryAcquire() bool {
	if b.pending >= b.maxPending {
		return false
	}
	b.pending++
	return true
}

func (b *BackpressureController) Release() {
	if b.pending > 0 {
		b.pending--
	}
}

func (b *BackpressureController) IsBlocked() bool {
	return b.pending >= b.maxPending
}
