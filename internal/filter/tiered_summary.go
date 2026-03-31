package filter

import "fmt"

// TieredSummary generates L0/L1/L2 tiered summaries of compressed content.
// Inspired by claw-compactor's tiered summary system.
type TieredSummary struct {
	L0 string // Ultra-compact (1-2 lines)
	L1 string // Compact (5-10 lines)
	L2 string // Detailed (full context with structure)
}

// GenerateTieredSummary creates multi-resolution summaries.
func GenerateTieredSummary(original, compressed string) TieredSummary {
	return TieredSummary{
		L0: generateL0(original, compressed),
		L1: generateL1(original, compressed),
		L2: generateL2(original, compressed),
	}
}

func generateL0(original, compressed string) string {
	origTokens := EstimateTokens(original)
	compTokens := EstimateTokens(compressed)
	saved := origTokens - compTokens
	pct := float64(saved) / float64(origTokens) * 100
	return fmt.Sprintf("Compressed: %d→%d tokens (%.0f%% saved)", origTokens, compTokens, pct)
}

func generateL1(original, compressed string) string {
	origTokens := EstimateTokens(original)
	compTokens := EstimateTokens(compressed)
	saved := origTokens - compTokens
	pct := float64(saved) / float64(origTokens) * 100
	return fmt.Sprintf("Original: %d tokens\nCompressed: %d tokens\nSaved: %d tokens (%.0f%%)\nRatio: %.2fx",
		origTokens, compTokens, saved, pct, float64(origTokens)/float64(compTokens))
}

func generateL2(original, compressed string) string {
	origTokens := EstimateTokens(original)
	compTokens := EstimateTokens(compressed)
	saved := origTokens - compTokens
	pct := float64(saved) / float64(origTokens) * 100
	return fmt.Sprintf("=== Compression Summary ===\nOriginal: %d tokens\nCompressed: %d tokens\nSaved: %d tokens (%.0f%%)\nRatio: %.2fx\nOriginal Size: %d bytes\nCompressed Size: %d bytes",
		origTokens, compTokens, saved, pct, float64(origTokens)/float64(compTokens), len(original), len(compressed))
}
