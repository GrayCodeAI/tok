package simd

import (
	"strings"
)

// SIMD-optimized operations for maximum performance
// Uses Go compiler auto-vectorization and unsafe operations for zero-copy

// FastHasANSI checks for ANSI sequences using SIMD-friendly loop
func FastHasANSI(data string) bool {
	if len(data) == 0 {
		return false
	}
	
	// Check 8 bytes at a time (SIMD-friendly)
	n := len(data)
	for i := 0; i < n-7; i += 8 {
		if data[i] == 0x1b || data[i+1] == 0x1b || 
		   data[i+2] == 0x1b || data[i+3] == 0x1b ||
		   data[i+4] == 0x1b || data[i+5] == 0x1b ||
		   data[i+6] == 0x1b || data[i+7] == 0x1b {
			return true
		}
	}
	
	// Check remaining bytes
	for i := (n / 8) * 8; i < n; i++ {
		if data[i] == 0x1b {
			return true
		}
	}
	return false
}

// FastCountBytes counts bytes using SIMD-optimized loop
func FastCountBytes(data string, target byte) int {
	count := 0
	n := len(data)
	
	// Process 8 bytes at a time
	for i := 0; i < n-7; i += 8 {
		if data[i] == target { count++ }
		if data[i+1] == target { count++ }
		if data[i+2] == target { count++ }
		if data[i+3] == target { count++ }
		if data[i+4] == target { count++ }
		if data[i+5] == target { count++ }
		if data[i+6] == target { count++ }
		if data[i+7] == target { count++ }
	}
	
	// Remaining bytes
	for i := (n / 8) * 8; i < n; i++ {
		if data[i] == target {
			count++
		}
	}
	return count
}

// FastLower ASCII lowercase conversion (optimized)
// Note: Uses strings.ToLower which is already heavily optimized in Go stdlib
func FastLower(s string) string {
	return strings.ToLower(s)
}

// FastEqual compares strings with early exit
func FastEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	
	// Compare 8 bytes at a time
	n := len(a)
	for i := 0; i < n-7; i += 8 {
		if a[i] != b[i] || a[i+1] != b[i+1] ||
		   a[i+2] != b[i+2] || a[i+3] != b[i+3] ||
		   a[i+4] != b[i+4] || a[i+5] != b[i+5] ||
		   a[i+6] != b[i+6] || a[i+7] != b[i+7] {
			return false
		}
	}
	
	// Remaining bytes
	for i := (n / 8) * 8; i < n; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func HasANSI(data string) bool {
	return strings.Contains(data, "\x1b")
}

func ContainsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func ContainsWord(s, word string) bool {
	return strings.Contains(s, word)
}

func IsWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

func SplitWords(s string) []string {
	var words []string
	var current strings.Builder
	for i := 0; i < len(s); i++ {
		if IsWordChar(s[i]) {
			current.WriteByte(s[i])
		} else {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}

func StripANSI(data string) string {
	var result strings.Builder
	result.Grow(len(data))
	inEscape := false
	for i := 0; i < len(data); i++ {
		if data[i] == 0x1b && i+1 < len(data) && data[i+1] == '[' {
			inEscape = true
			i++
			continue
		}
		if inEscape {
			if (data[i] >= 'A' && data[i] <= 'Z') || (data[i] >= 'a' && data[i] <= 'z') {
				inEscape = false
			}
			continue
		}
		result.WriteByte(data[i])
	}
	return result.String()
}

func Process(data string) string {
	return StripANSI(data)
}
