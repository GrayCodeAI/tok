// Package core provides entropy-based filtering for compression optimization.
package core

import (
	"math"
	"strings"
)

// EntropyFilter implements Shannon entropy-based content filtering.
// It identifies high-entropy regions (random data, hashes, keys) that
// are less compressible and can be aggressively deduplicated.
type EntropyFilter struct {
	windowSize    int
	entropyThresh float64
	minRepeat     int
}

// EntropyConfig configures the entropy filter.
type EntropyConfig struct {
	WindowSize        int
	HighEntropyThresh float64 // Threshold for high entropy (random data)
	LowEntropyThresh  float64 // Threshold for low entropy (repetitive data)
	MinRepeat         int     // Minimum repeats to consider deduplication
}

// DefaultEntropyConfig returns default configuration.
func DefaultEntropyConfig() EntropyConfig {
	return EntropyConfig{
		WindowSize:        64,
		HighEntropyThresh: 4.5, // Shannon entropy > 4.5 bits/byte = high entropy
		LowEntropyThresh:  2.0, // Shannon entropy < 2.0 bits/byte = low entropy
		MinRepeat:         3,
	}
}

// NewEntropyFilter creates a new entropy filter.
func NewEntropyFilter(config EntropyConfig) *EntropyFilter {
	return &EntropyFilter{
		windowSize:    config.WindowSize,
		entropyThresh: config.HighEntropyThresh,
		minRepeat:     config.MinRepeat,
	}
}

// Region represents a region of content with entropy characteristics.
type Region struct {
	Start       int
	End         int
	Entropy     float64
	EntropyType EntropyType
	Content     string
}

// EntropyType classifies the entropy level.
type EntropyType int

const (
	LowEntropy    EntropyType = iota // Repetitive, highly compressible
	MediumEntropy                    // Normal text/code
	HighEntropy                      // Random data, hashes, keys
)

// Analyze analyzes content and returns regions with entropy information.
func (f *EntropyFilter) Analyze(content string) []Region {
	if len(content) == 0 {
		return nil
	}

	var regions []Region
	lines := strings.Split(content, "\n")
	pos := 0

	for _, line := range lines {
		lineLen := len(line)
		if lineLen == 0 {
			pos++ // newline
			continue
		}

		// Calculate entropy for the line
		entropy := calculateEntropy(line)
		entType := classifyEntropy(entropy, f.entropyThresh)

		regions = append(regions, Region{
			Start:       pos,
			End:         pos + lineLen,
			Entropy:     entropy,
			EntropyType: entType,
			Content:     line,
		})

		pos += lineLen + 1 // +1 for newline
	}

	return regions
}

// FindDuplicates finds duplicate high-entropy regions.
func (f *EntropyFilter) FindDuplicates(content string) map[string][]int {
	regions := f.Analyze(content)
	duplicates := make(map[string][]int)

	for _, region := range regions {
		if region.EntropyType == HighEntropy {
			key := hashContent(region.Content)
			duplicates[key] = append(duplicates[key], region.Start)
		}
	}

	// Filter to only those with min repeats
	result := make(map[string][]int)
	for key, positions := range duplicates {
		if len(positions) >= f.minRepeat {
			result[key] = positions
		}
	}

	return result
}

// Compress applies entropy-based compression.
func (f *EntropyFilter) Compress(content string) (string, CompressionInfo) {
	regions := f.Analyze(content)
	if len(regions) == 0 {
		return content, CompressionInfo{OriginalSize: len(content), CompressedSize: len(content)}
	}

	var builder strings.Builder
	totalSaved := 0

	for _, region := range regions {
		switch region.EntropyType {
		case HighEntropy:
			// For high entropy, use aggressive deduplication
			compressed := compressHighEntropy(region.Content)
			builder.WriteString(compressed)
			totalSaved += len(region.Content) - len(compressed)

		case LowEntropy:
			// For low entropy, aggressive whitespace compression
			compressed := compressLowEntropy(region.Content)
			builder.WriteString(compressed)
			totalSaved += len(region.Content) - len(compressed)

		default:
			// Medium entropy - pass through with minor normalization
			builder.WriteString(region.Content)
		}
		builder.WriteByte('\n')
	}

	compressed := builder.String()
	return compressed, CompressionInfo{
		OriginalSize:   len(content),
		CompressedSize: len(compressed),
		Saved:          totalSaved,
		Regions:        len(regions),
	}
}

// CompressionInfo contains compression statistics.
type CompressionInfo struct {
	OriginalSize   int
	CompressedSize int
	Saved          int
	Regions        int
}

// calculateEntropy calculates Shannon entropy in bits per byte.
func calculateEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	// Count character frequencies
	freq := make(map[byte]int)
	for i := 0; i < len(s); i++ {
		freq[s[i]]++
	}

	// Calculate entropy
	length := float64(len(s))
	var entropy float64
	for _, count := range freq {
		p := float64(count) / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

func classifyEntropy(entropy, threshold float64) EntropyType {
	if entropy > threshold {
		return HighEntropy
	}
	if entropy < threshold/2 {
		return LowEntropy
	}
	return MediumEntropy
}

func hashContent(s string) string {
	// Simple hash for deduplication
	var h uint64 = 14695981039346656037
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return string(rune(h & 0xFFFFFFFF))
}

func compressHighEntropy(s string) string {
	// For high entropy data (UUIDs, hashes, keys), truncate with marker
	if looksLikeUUID(s) {
		return "[UUID:" + s[:8] + "]"
	}
	if looksLikeHash(s) {
		return "[HASH:" + s[:16] + "]"
	}
	if looksLikeBase64(s) {
		return "[B64:" + s[:20] + "]"
	}
	return s
}

func compressLowEntropy(s string) string {
	// For low entropy, compress repeated patterns
	result := strings.ReplaceAll(s, "    ", "\t")
	result = strings.ReplaceAll(result, "  ", " ")
	return strings.TrimRight(result, " \t")
}

func looksLikeUUID(s string) bool {
	// Simple UUID pattern check
	if len(s) < 32 {
		return false
	}
	// Check for UUID-like structure (8-4-4-4-12)
	if len(s) >= 36 && s[8] == '-' && s[13] == '-' && s[18] == '-' && s[23] == '-' {
		return true
	}
	return false
}

func looksLikeHash(s string) bool {
	// Check for hash patterns (hex strings of specific lengths)
	s = strings.TrimSpace(s)
	switch len(s) {
	case 32, 40, 64: // MD5, SHA1, SHA256
		for _, c := range s {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
		return true
	}
	return false
}

func looksLikeBase64(s string) bool {
	// Check for base64 characteristics
	s = strings.TrimSpace(s)
	if len(s) < 20 {
		return false
	}
	// Check for base64 character set
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
			return false
		}
	}
	return len(s)%4 == 0
}

// EntropyStats provides statistics about content entropy.
type EntropyStats struct {
	AvgEntropy      float64
	HighEntropyPct  float64
	LowEntropyPct   float64
	DuplicateHashes int
}

// GetStats returns entropy statistics for content.
func (f *EntropyFilter) GetStats(content string) EntropyStats {
	regions := f.Analyze(content)
	if len(regions) == 0 {
		return EntropyStats{}
	}

	var totalEntropy float64
	highCount, lowCount := 0, 0

	for _, r := range regions {
		totalEntropy += r.Entropy
		switch r.EntropyType {
		case HighEntropy:
			highCount++
		case LowEntropy:
			lowCount++
		}
	}

	duplicates := f.FindDuplicates(content)
	totalDupes := 0
	for _, positions := range duplicates {
		totalDupes += len(positions) - 1 // Count duplicates beyond first occurrence
	}

	total := float64(len(regions))
	return EntropyStats{
		AvgEntropy:      totalEntropy / total,
		HighEntropyPct:  float64(highCount) / total * 100,
		LowEntropyPct:   float64(lowCount) / total * 100,
		DuplicateHashes: totalDupes,
	}
}
