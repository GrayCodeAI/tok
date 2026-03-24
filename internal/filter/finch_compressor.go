package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// FINCHCompressor implements prompt-guided KV cache compression.
// Research: "FINCH: Prompt-guided Key-Value Cache Compression" (TACL 2024)
// Key Innovation: Given a prompt/query, iteratively identifies the most relevant
// content using attention-like scoring, achieving up to 93x compression.
//
// In TokMan: When a query/intent is provided, use it to guide which content
// to keep. Content relevant to the query gets higher scores. This complements
// our existing goal-driven filter with a more sophisticated scoring mechanism.
type FINCHCompressor struct {
	config FINCHConfig
}

// FINCHConfig holds configuration for FINCH compression
type FINCHConfig struct {
	Enabled          bool
	QueryWeight      float64 // Weight of query relevance in scoring
	ChunkSize        int     // Lines per chunk for chunked scoring
	MaxChunks        int     // Maximum chunks to process
	MinContentLength int
}

// DefaultFINCHConfig returns default configuration
func DefaultFINCHConfig() FINCHConfig {
	return FINCHConfig{
		Enabled:          true,
		QueryWeight:      0.6,
		ChunkSize:        5,
		MaxChunks:        100,
		MinContentLength: 300,
	}
}

// NewFINCHCompressor creates a new FINCH compressor
func NewFINCHCompressor() *FINCHCompressor {
	return &FINCHCompressor{config: DefaultFINCHConfig()}
}

// Name returns the filter name
func (f *FINCHCompressor) Name() string { return "finch" }

// Apply applies FINCH-style query-guided compression
func (f *FINCHCompressor) Apply(input string, mode Mode) (string, int) {
	if !f.config.Enabled || mode == ModeNone {
		return input, 0
	}

	if len(input) < f.config.MinContentLength {
		return input, 0
	}

	originalTokens := core.EstimateTokens(input)

	// Split into chunks
	chunks := f.splitChunks(input)
	if len(chunks) < 3 {
		return input, 0
	}

	// Score chunks (without query, use content importance)
	scores := f.scoreChunks(chunks, mode)

	// Select top chunks
	output := f.selectChunks(chunks, scores, mode)

	finalTokens := core.EstimateTokens(output)
	saved := originalTokens - finalTokens
	if saved < 5 {
		return input, 0
	}

	return output, saved
}

// ApplyWithQuery applies FINCH with a specific query for guided compression
func (f *FINCHCompressor) ApplyWithQuery(input, query string, mode Mode) (string, int) {
	if !f.config.Enabled || mode == ModeNone {
		return input, 0
	}

	if len(input) < f.config.MinContentLength {
		return input, 0
	}

	originalTokens := core.EstimateTokens(input)

	chunks := f.splitChunks(input)
	if len(chunks) < 3 {
		return input, 0
	}

	// Score chunks with query guidance
	scores := f.scoreChunksWithQuery(chunks, query, mode)

	output := f.selectChunks(chunks, scores, mode)

	finalTokens := core.EstimateTokens(output)
	saved := originalTokens - finalTokens
	if saved < 5 {
		return input, 0
	}

	return output, saved
}

// finchChunk represents a chunk of content
type finchChunk struct {
	content string
	lines   []string
	words   map[string]bool
}

// splitChunks splits input into chunks
func (f *FINCHCompressor) splitChunks(input string) []finchChunk {
	lines := strings.Split(input, "\n")
	var chunks []finchChunk

	for i := 0; i < len(lines); i += f.config.ChunkSize {
		end := i + f.config.ChunkSize
		if end > len(lines) {
			end = len(lines)
		}

		chunkLines := lines[i:end]
		content := strings.Join(chunkLines, "\n")

		chunks = append(chunks, finchChunk{
			content: content,
			lines:   chunkLines,
			words:   f.extractWordSet(content),
		})

		if len(chunks) >= f.config.MaxChunks {
			break
		}
	}

	return chunks
}

// scoreChunks scores chunks without query
func (f *FINCHCompressor) scoreChunks(chunks []finchChunk, mode Mode) []float64 {
	scores := make([]float64, len(chunks))

	for i, chunk := range chunks {
		score := 0.5

		// Information density
		if len(chunk.lines) > 0 {
			uniqueWords := len(chunk.words)
			totalWords := 0
			for _, line := range chunk.lines {
				totalWords += len(strings.Fields(line))
			}
			if totalWords > 0 {
				score += float64(uniqueWords) / float64(totalWords) * 0.3
			}
		}

		// Structural importance
		for _, line := range chunk.lines {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "error") || strings.Contains(lower, "fail") {
				score += 0.2
			}
			if strings.Contains(line, "func ") || strings.Contains(line, "class ") {
				score += 0.15
			}
		}

		// First and last chunks are often important
		if i == 0 || i == len(chunks)-1 {
			score += 0.1
		}

		scores[i] = score
	}

	return scores
}

// scoreChunksWithQuery scores chunks with query guidance
func (f *FINCHCompressor) scoreChunksWithQuery(chunks []finchChunk, query string, mode Mode) []float64 {
	queryWords := f.extractWordSet(query)
	queryWeight := f.config.QueryWeight

	scores := make([]float64, len(chunks))

	for i, chunk := range chunks {
		// Content importance (same as without query)
		contentScore := 0.5
		if len(chunk.lines) > 0 {
			uniqueWords := len(chunk.words)
			totalWords := 0
			for _, line := range chunk.lines {
				totalWords += len(strings.Fields(line))
			}
			if totalWords > 0 {
				contentScore = float64(uniqueWords) / float64(totalWords)
			}
		}

		// Query relevance (Jaccard similarity)
		queryRelevance := 0.0
		if len(queryWords) > 0 && len(chunk.words) > 0 {
			intersection := 0
			for w := range queryWords {
				if chunk.words[w] {
					intersection++
				}
			}
			union := len(queryWords) + len(chunk.words) - intersection
			if union > 0 {
				queryRelevance = float64(intersection) / float64(union)
			}
		}

		// Combined score
		scores[i] = contentScore*(1-queryWeight) + queryRelevance*queryWeight
	}

	return scores
}

// selectChunks selects top chunks based on scores
func (f *FINCHCompressor) selectChunks(chunks []finchChunk, scores []float64, mode Mode) string {
	// Determine how many chunks to keep
	keepRatio := 0.5
	if mode == ModeAggressive {
		keepRatio = 0.3
	}

	keepCount := int(float64(len(chunks)) * keepRatio)
	if keepCount < 1 {
		keepCount = 1
	}

	// Select top chunks by score (preserve order)
	keepMask := make([]bool, len(chunks))

	for i := 0; i < keepCount; i++ {
		bestIdx := -1
		bestScore := -1.0
		for j := 0; j < len(chunks); j++ {
			if !keepMask[j] && scores[j] > bestScore {
				bestScore = scores[j]
				bestIdx = j
			}
		}
		if bestIdx >= 0 {
			keepMask[bestIdx] = true
		}
	}

	// Reconstruct in original order
	var result strings.Builder
	for i, chunk := range chunks {
		if keepMask[i] {
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			result.WriteString(chunk.content)
		}
	}

	return strings.TrimSpace(result.String())
}

// extractWordSet extracts word set from text
func (f *FINCHCompressor) extractWordSet(text string) map[string]bool {
	words := make(map[string]bool)
	for _, w := range strings.Fields(strings.ToLower(text)) {
		cleaned := strings.Trim(w, ".,;:!?\"'()[]{}")
		if len(cleaned) > 2 {
			words[cleaned] = true
		}
	}
	return words
}
