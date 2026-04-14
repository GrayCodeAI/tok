package simd

import (
	"bytes"
	"strings"
	"unsafe"
)

// SIMD-optimized operations for maximum performance
// Uses Go compiler auto-vectorization and unsafe operations for zero-copy

// FastHasANSI checks for ANSI sequences using SIMD-friendly loop
func FastHasANSI(data string) bool {
	if len(data) == 0 {
		return false
	}

	// Check 16 bytes at a time for better SIMD utilization
	n := len(data)
	for i := 0; i < n-15; i += 16 {
		// Unrolled 16-byte check
		if data[i] == 0x1b || data[i+1] == 0x1b || data[i+2] == 0x1b || data[i+3] == 0x1b ||
			data[i+4] == 0x1b || data[i+5] == 0x1b || data[i+6] == 0x1b || data[i+7] == 0x1b ||
			data[i+8] == 0x1b || data[i+9] == 0x1b || data[i+10] == 0x1b || data[i+11] == 0x1b ||
			data[i+12] == 0x1b || data[i+13] == 0x1b || data[i+14] == 0x1b || data[i+15] == 0x1b {
			return true
		}
	}

	// Check remaining bytes
	for i := (n / 16) * 16; i < n; i++ {
		if data[i] == 0x1b {
			return true
		}
	}
	return false
}

// StripANSI removes ANSI sequences using optimized byte operations
func StripANSI(input string) string {
	if !FastHasANSI(input) {
		return input
	}

	// Pre-allocate output buffer (worst case: same size as input)
	var buf bytes.Buffer
	buf.Grow(len(input))

	inEscape := false
	for i := 0; i < len(input); i++ {
		c := input[i]
		if inEscape {
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '~' {
				inEscape = false
			}
		} else if c == 0x1b {
			inEscape = true
		} else {
			buf.WriteByte(c)
		}
	}
	return buf.String()
}

// FastCountBytes counts bytes using SIMD-optimized loop
func FastCountBytes(data string, target byte) int {
	if len(data) == 0 {
		return 0
	}

	count := 0
	n := len(data)

	// Process 16 bytes at a time with unrolled loop
	for i := 0; i < n-15; i += 16 {
		// Unrolled comparisons
		if data[i] == target {
			count++
		}
		if data[i+1] == target {
			count++
		}
		if data[i+2] == target {
			count++
		}
		if data[i+3] == target {
			count++
		}
		if data[i+4] == target {
			count++
		}
		if data[i+5] == target {
			count++
		}
		if data[i+6] == target {
			count++
		}
		if data[i+7] == target {
			count++
		}
		if data[i+8] == target {
			count++
		}
		if data[i+9] == target {
			count++
		}
		if data[i+10] == target {
			count++
		}
		if data[i+11] == target {
			count++
		}
		if data[i+12] == target {
			count++
		}
		if data[i+13] == target {
			count++
		}
		if data[i+14] == target {
			count++
		}
		if data[i+15] == target {
			count++
		}
	}

	// Remaining bytes
	for i := (n / 16) * 16; i < n; i++ {
		if data[i] == target {
			count++
		}
	}
	return count
}

// FastLower ASCII lowercase conversion (optimized with unsafe)
func FastLower(s string) string {
	if len(s) == 0 {
		return s
	}

	// Check if already lowercase
	hasUpper := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return s
	}

	// Convert using unsafe for zero-copy when possible
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		c := b[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return *(*string)(unsafe.Pointer(&b))
}

// FastEqual compares strings with early exit
func FastEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	if len(a) == 0 {
		return true
	}

	// Compare 16 bytes at a time
	n := len(a)
	for i := 0; i < n-15; i += 16 {
		if a[i] != b[i] || a[i+1] != b[i+1] || a[i+2] != b[i+2] || a[i+3] != b[i+3] ||
			a[i+4] != b[i+4] || a[i+5] != b[i+5] || a[i+6] != b[i+6] || a[i+7] != b[i+7] ||
			a[i+8] != b[i+8] || a[i+9] != b[i+9] || a[i+10] != b[i+10] || a[i+11] != b[i+11] ||
			a[i+12] != b[i+12] || a[i+13] != b[i+13] || a[i+14] != b[i+14] || a[i+15] != b[i+15] {
			return false
		}
	}

	// Remaining bytes
	for i := (n / 16) * 16; i < n; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// FastContains checks if string contains substring (optimized for small patterns)
func FastContains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}

	// For small patterns, use bytes.Index
	if len(substr) <= 16 {
		return bytes.Index([]byte(s), []byte(substr)) >= 0
	}
	return strings.Contains(s, substr)
}

// ContainsAny checks if string contains any of the substrings
func ContainsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// ContainsWord checks if string contains word (space-delimited)
func ContainsWord(s, word string) bool {
	if len(word) == 0 {
		return false
	}

	lowerS := FastLower(s)
	lowerWord := FastLower(word)

	// Check word boundaries
	idx := strings.Index(lowerS, lowerWord)
	if idx < 0 {
		return false
	}

	// Check if it's a whole word
	wordLen := len(lowerWord)
	sLen := len(lowerS)

	for idx >= 0 {
		before := idx == 0 || !IsWordChar(lowerS[idx-1])
		after := idx+wordLen >= sLen || !IsWordChar(lowerS[idx+wordLen])
		if before && after {
			return true
		}
		// Look for next occurrence
		if idx+1 >= sLen {
			break
		}
		nextIdx := strings.Index(lowerS[idx+1:], lowerWord)
		if nextIdx < 0 {
			break
		}
		idx += nextIdx + 1
	}
	return false
}

// IsWordChar checks if byte is a word character
func IsWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// SplitWords splits string into words (optimized)
func SplitWords(input string) []string {
	if len(input) == 0 {
		return nil
	}

	// Pre-allocate slice (estimate ~10% of length for words)
	estimatedWords := len(input) / 10
	if estimatedWords < 8 {
		estimatedWords = 8
	}
	words := make([]string, 0, estimatedWords)

	start := -1
	for i := 0; i < len(input); i++ {
		c := input[i]
		if IsWordChar(c) || c == '\'' {
			if start < 0 {
				start = i
			}
		} else {
			if start >= 0 {
				words = append(words, input[start:i])
				start = -1
			}
		}
	}
	if start >= 0 {
		words = append(words, input[start:])
	}
	return words
}

// HasANSI is an alias for FastHasANSI for backward compatibility
func HasANSI(data string) bool {
	return FastHasANSI(data)
}

// Process is an alias for StripANSI for backward compatibility
func Process(data string) string {
	return StripANSI(data)
}
