package filter

import (
	"strings"
	"sync"
)

// StreamingProcessor handles large inputs (>500K tokens) with chunked processing.
// This reduces memory usage by processing content in chunks rather than loading
// everything into memory at once.
//
// Based on research:
// - DSPC (Sep 2025): Coarse filtering before expensive layers
// - MemGPT (UC Berkeley 2023): Memory-efficient context management
type StreamingProcessor struct {
	chunkSize    int   // tokens per chunk
	overlap      int   // token overlap between chunks
	maxChunks    int   // maximum chunks to process
	mode         Mode  // compression mode
	layerConfigs LayerConfigs

	// Pool for reusing string builders
	builderPool sync.Pool
}

// LayerConfigs holds configuration for which layers to apply in streaming mode
type LayerConfigs struct {
	EnableEntropy     bool
	EnableTFIDF       bool
	EnableH2O         bool
	EnableCompaction  bool
	EnableAttribution bool
}

// NewStreamingProcessor creates a new streaming processor
func NewStreamingProcessor(mode Mode, cfg LayerConfigs) *StreamingProcessor {
	return &StreamingProcessor{
		chunkSize:    50000, // ~50K tokens per chunk
		overlap:      500,   // 500 token overlap for context continuity
		maxChunks:    20,    // Max 20 chunks = ~1M tokens max
		mode:         mode,
		layerConfigs: cfg,
		builderPool: sync.Pool{
			New: func() interface{} {
				return &strings.Builder{}
			},
		},
	}
}

// ProcessStream processes large input in chunks with reduced memory footprint
func (sp *StreamingProcessor) ProcessStream(input string) (string, *PipelineStats) {
	stats := &PipelineStats{
		OriginalTokens: EstimateTokens(input),
		LayerStats:     make(map[string]LayerStat),
	}

	// Check if streaming is needed
	if stats.OriginalTokens < 500000 {
		// Small enough for standard processing
		pipeline := NewPipelineCoordinator(PipelineConfig{
			Mode:           sp.mode,
			EnableEntropy:  sp.layerConfigs.EnableEntropy,
			EnableTFIDF:    sp.layerConfigs.EnableTFIDF,
			EnableH2O:      sp.layerConfigs.EnableH2O,
			EnableCompaction: sp.layerConfigs.EnableCompaction,
		})
		return pipeline.Process(input)
	}

	// Split into chunks
	chunks := sp.splitIntoChunks(input)
	results := make([]chunkResult, 0, len(chunks))

	// Process each chunk
	for i, chunk := range chunks {
		if i >= sp.maxChunks {
			break
		}
		processed, chunkStats := sp.processChunk(chunk.content)
		results = append(results, chunkResult{
			content:     processed,
			tokensSaved: chunkStats.TotalSaved,
		})
		stats.TotalSaved += chunkStats.TotalSaved

		// Merge layer stats
		for layer, stat := range chunkStats.LayerStats {
			existing := stats.LayerStats[layer]
			existing.TokensSaved += stat.TokensSaved
			stats.LayerStats[layer] = existing
		}
	}

	// Combine results
	output := sp.combineResults(results)
	stats.FinalTokens = EstimateTokens(output)
	stats.ReductionPercent = float64(stats.TotalSaved) / float64(stats.OriginalTokens) * 100

	return output, stats
}

type chunkResult struct {
	content     string
	tokensSaved int
}

type inputChunk struct {
	content string
	index   int
}

// splitIntoChunks splits input into overlapping chunks for processing
func (sp *StreamingProcessor) splitIntoChunks(input string) []inputChunk {
	lines := strings.Split(input, "\n")
	var chunks []inputChunk

	currentChunk := sp.getBuilder()
	currentSize := 0
	chunkIndex := 0

	for _, line := range lines {
		lineTokens := EstimateTokens(line)

		if currentSize+lineTokens > sp.chunkSize && currentSize > 0 {
			// Save current chunk
			chunks = append(chunks, inputChunk{
				content: currentChunk.String(),
				index:   chunkIndex,
			})
			chunkIndex++

			// Start new chunk with overlap
			currentChunk.Reset()
			currentSize = 0

			// Add overlap lines from previous chunk for context
			overlapLines := sp.getLastLines(chunks[len(chunks)-1].content, sp.overlap)
			for _, overlapLine := range overlapLines {
				currentChunk.WriteString(overlapLine)
				currentChunk.WriteString("\n")
				currentSize += EstimateTokens(overlapLine)
			}
		}

		currentChunk.WriteString(line)
		currentChunk.WriteString("\n")
		currentSize += lineTokens
	}

	// Add final chunk
	if currentSize > 0 {
		chunks = append(chunks, inputChunk{
			content: currentChunk.String(),
			index:   chunkIndex,
		})
	}

	return chunks
}

// processChunk processes a single chunk through lightweight layers
func (sp *StreamingProcessor) processChunk(chunk string) (string, *PipelineStats) {
	cfg := PipelineConfig{
		Mode:           sp.mode,
		EnableEntropy:  sp.layerConfigs.EnableEntropy,
		EnableTFIDF:    sp.layerConfigs.EnableTFIDF,
		EnableH2O:      sp.layerConfigs.EnableH2O,
		EnableCompaction: sp.layerConfigs.EnableCompaction,
		EnableAttribution: sp.layerConfigs.EnableAttribution,
	}

	pipeline := NewPipelineCoordinator(cfg)
	return pipeline.Process(chunk)
}

// combineResults merges processed chunks, removing overlap duplicates
func (sp *StreamingProcessor) combineResults(results []chunkResult) string {
	if len(results) == 0 {
		return ""
	}

	builder := sp.getBuilder()
	defer sp.putBuilder(builder)

	for i, result := range results {
		if i > 0 {
			// Remove overlap prefix from subsequent chunks
			content := sp.removeOverlapPrefix(result.content)
			builder.WriteString(content)
		} else {
			builder.WriteString(result.content)
		}
	}

	return builder.String()
}

// getLastLines returns the last N tokens worth of lines
func (sp *StreamingProcessor) getLastLines(content string, maxTokens int) []string {
	lines := strings.Split(content, "\n")
	var result []string
	tokenCount := 0

	// Iterate backwards
	for i := len(lines) - 1; i >= 0 && tokenCount < maxTokens; i-- {
		if lines[i] == "" {
			continue
		}
		tokens := EstimateTokens(lines[i])
		if tokenCount+tokens > maxTokens {
			break
		}
		result = append([]string{lines[i]}, result...)
		tokenCount += tokens
	}

	return result
}

// removeOverlapPrefix removes the overlap section from chunk content
func (sp *StreamingProcessor) removeOverlapPrefix(content string) string {
	lines := strings.Split(content, "\n")
	tokenCount := 0
	startIndex := 0

	// Find where overlap ends
	for i, line := range lines {
		tokens := EstimateTokens(line)
		if tokenCount+tokens > sp.overlap {
			startIndex = i
			break
		}
		tokenCount += tokens
	}

	// Return content after overlap
	if startIndex < len(lines) {
		return strings.Join(lines[startIndex:], "\n")
	}
	return content
}

// getBuilder gets a string builder from the pool
func (sp *StreamingProcessor) getBuilder() *strings.Builder {
	return sp.builderPool.Get().(*strings.Builder)
}

// putBuilder returns a string builder to the pool
func (sp *StreamingProcessor) putBuilder(builder *strings.Builder) {
	builder.Reset()
	sp.builderPool.Put(builder)
}
