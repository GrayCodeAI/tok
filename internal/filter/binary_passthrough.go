package filter

import (
	"unicode/utf8"
)

// BinaryPassthrough detects binary output and passes it through unchanged.
// Compression on binary data (images, compiled binaries, PDFs) is harmful
// and can corrupt the output. This filter detects binary content and skips
// all compression layers.
type BinaryPassthrough struct {
	config BinaryConfig
}

// BinaryConfig holds configuration for binary detection
type BinaryConfig struct {
	Enabled          bool
	MaxBinaryRatio   float64 // Max ratio of non-UTF8 bytes before marking as binary
	MinCheckBytes    int     // Minimum bytes to check
}

// DefaultBinaryConfig returns default configuration
func DefaultBinaryConfig() BinaryConfig {
	return BinaryConfig{
		Enabled:        true,
		MaxBinaryRatio: 0.1, // >10% non-UTF8 = binary
		MinCheckBytes:  512,
	}
}

// NewBinaryPassthrough creates a new binary detector
func NewBinaryPassthrough() *BinaryPassthrough {
	return &BinaryPassthrough{
		config: DefaultBinaryConfig(),
	}
}

// Name returns the filter name
func (b *BinaryPassthrough) Name() string {
	return "binary_passthrough"
}

// IsBinary checks if content is binary (without modifying)
func (b *BinaryPassthrough) IsBinary(input string) bool {
	if !b.config.Enabled {
		return false
	}

	if len(input) == 0 {
		return false
	}

	// Quick check: null bytes are a strong binary indicator
	for i := 0; i < len(input) && i < b.config.MinCheckBytes; i++ {
		if input[i] == 0 {
			return true
		}
	}

	// Check UTF-8 validity ratio
	checkLen := len(input)
	if checkLen > b.config.MinCheckBytes*10 {
		checkLen = b.config.MinCheckBytes * 10
	}

	invalidCount := 0
	validCount := 0
	for i := 0; i < checkLen; {
		r, size := utf8.DecodeRuneInString(input[i:])
		if r == utf8.RuneError && size == 1 {
			invalidCount++
		} else {
			validCount++
		}
		i += size
	}

	total := invalidCount + validCount
	if total == 0 {
		return false
	}

	ratio := float64(invalidCount) / float64(total)
	return ratio > b.config.MaxBinaryRatio
}

// Apply passes binary content through unchanged
func (b *BinaryPassthrough) Apply(input string, mode Mode) (string, int) {
	if b.IsBinary(input) {
		// Pass through unchanged - don't compress binary data
		return input, 0
	}
	return input, 0
}
