package filter

import (
	"bufio"
	"io"
)

// StreamingPipeline processes input via io.Reader/Writer for large files
type StreamingPipeline struct {
	coordinator *PipelineCoordinator
	bufferSize  int
}

// NewStreamingPipeline creates a streaming pipeline wrapper
func NewStreamingPipeline(cfg PipelineConfig) *StreamingPipeline {
	return &StreamingPipeline{
		coordinator: NewPipelineCoordinator(&cfg),
		bufferSize:  64 * 1024, // 64KB chunks
	}
}

// ProcessStream compresses input stream to output stream
func (sp *StreamingPipeline) ProcessStream(r io.Reader, w io.Writer) (*PipelineStats, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, sp.bufferSize), sp.bufferSize)

	var totalStats *PipelineStats

	for scanner.Scan() {
		line := scanner.Text()
		compressed, stats := sp.coordinator.Process(line)

		if _, err := w.Write([]byte(compressed + "\n")); err != nil {
			return totalStats, err
		}

		if totalStats == nil {
			totalStats = stats
		} else {
			totalStats.OriginalTokens += stats.OriginalTokens
			totalStats.FinalTokens += stats.FinalTokens
			totalStats.TotalSaved += stats.TotalSaved
		}
	}

	return totalStats, scanner.Err()
}
