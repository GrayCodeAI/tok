package filter

import (
	"io"
	"strings"

	"github.com/lakshmanpatel/tok/internal/core"
)

// StreamingThreshold is the token count at which streaming mode activates.

// ChunkSize is the target size for each chunk in streaming mode.
const ChunkSize = 100000 // 100K tokens per chunk

// StreamingProcessor handles large inputs by processing in chunks.
type StreamingProcessor struct {
	coordinator *PipelineCoordinator
	chunkSize   int
}

// NewStreamingProcessor creates a new streaming processor.
func NewStreamingProcessor(coordinator *PipelineCoordinator) *StreamingProcessor {
	return &StreamingProcessor{
		coordinator: coordinator,
		chunkSize:   ChunkSize,
	}
}

// ShouldStream returns true if the input should be processed in streaming mode.
func ShouldStream(input string) bool {
	// Use fast estimation for the check
	estimatedTokens := core.EstimateTokens(input)
	return estimatedTokens > StreamingThreshold
}

// ProcessStream processes large input in chunks.
// This reduces memory usage for very large inputs.
func (sp *StreamingProcessor) ProcessStream(input string) (string, *PipelineStats) {
	if !ShouldStream(input) {
		// Not large enough for streaming, use normal processing
		return sp.coordinator.Process(input)
	}

	// Split input into chunks at natural boundaries
	chunks := sp.splitIntoChunks(input)

	// Process each chunk
	var results []string
	totalStats := &PipelineStats{
		LayerStats: make(map[string]LayerStat),
	}

	for _, chunk := range chunks {
		output, stats := sp.coordinator.Process(chunk)
		results = append(results, output)

		// Aggregate stats
		totalStats.OriginalTokens += stats.OriginalTokens
		totalStats.FinalTokens += stats.FinalTokens
		totalStats.TotalSaved += stats.TotalSaved

		// Merge layer stats
		for name, stat := range stats.LayerStats {
			if existing, ok := totalStats.LayerStats[name]; ok {
				totalStats.LayerStats[name] = LayerStat{
					TokensSaved: existing.TokensSaved + stat.TokensSaved,
					Duration:    existing.Duration + stat.Duration,
				}
			} else {
				totalStats.LayerStats[name] = stat
			}
		}
	}

	// Join results
	finalOutput := strings.Join(results, "\n")

	// Calculate final reduction with overflow protection
	if totalStats.OriginalTokens > 0 {
		if totalStats.TotalSaved < 0 {
			totalStats.TotalSaved = 0
		}
		totalStats.ReductionPercent = float64(totalStats.TotalSaved) / float64(totalStats.OriginalTokens) * 100
		// Clamp to valid range
		if totalStats.ReductionPercent < 0 {
			totalStats.ReductionPercent = 0
		} else if totalStats.ReductionPercent > 100 {
			totalStats.ReductionPercent = 100
		}
	}

	return finalOutput, totalStats
}

// splitIntoChunks splits input into chunks at natural boundaries.
func (sp *StreamingProcessor) splitIntoChunks(input string) []string {
	var chunks []string
	lines := strings.Split(input, "\n")

	var currentChunk strings.Builder
	currentTokens := 0

	for _, line := range lines {
		lineTokens := core.EstimateTokens(line)

		// Start new chunk if current would exceed target
		if currentTokens > 0 && currentTokens+lineTokens > sp.chunkSize {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
			currentTokens = 0
		}

		currentChunk.WriteString(line)
		currentChunk.WriteString("\n")
		currentTokens += lineTokens
	}

	// Add final chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

// ProcessStreamReader processes input from a reader in streaming mode.
func (sp *StreamingProcessor) ProcessStreamReader(reader io.Reader, writer io.Writer) error {
	// Read all input (for now - future: true streaming)
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	output, _ := sp.ProcessStream(string(data))
	_, err = writer.Write([]byte(output))
	return err
}

// StreamConfig holds configuration for streaming mode.
type StreamConfig struct {
	Enabled        bool
	Threshold      int
	ChunkSize      int
	ParallelChunks bool // Future: process chunks in parallel
}

// DefaultStreamConfig returns the default streaming configuration.
func DefaultStreamConfig() StreamConfig {
	return StreamConfig{
		Enabled:        true,
		Threshold:      StreamingThreshold,
		ChunkSize:      ChunkSize,
		ParallelChunks: false,
	}
}
