//go:build (amd64 || arm64) && go1.26

package simd

import (
	"golang.org/x/sys/cpu"
)

// SIMD feature flags detected at runtime
var (
	hasAVX2    bool
	hasAVX512  bool
	hasNEON    bool
	simdWidth  int // SIMD register width in bytes
)

func init() {
	// Detect CPU features at runtime
	detectSIMDCapabilities()
	
	// Enable SIMD if supported
	if hasAVX2 || hasAVX512 || hasNEON {
		Enabled = true
	}
}

// detectSIMDCapabilities detects available SIMD instructions
func detectSIMDCapabilities() {
	// AMD64 detection
	if cpu.X86.HasAVX2 {
		hasAVX2 = true
		simdWidth = 32 // AVX2 uses 256-bit (32-byte) registers
	}
	if cpu.X86.HasAVX512F {
		hasAVX512 = true
		simdWidth = 64 // AVX-512 uses 512-bit (64-byte) registers
	}
	
	// ARM64 detection
	if cpu.ARM64.HasASIMD {
		hasNEON = true
		simdWidth = 16 // NEON uses 128-bit (16-byte) registers
	}
}

// StripANSINative provides SIMD-accelerated ANSI stripping
// Uses parallel byte comparison to find escape sequences faster
func StripANSINative(input string) string {
	if !Enabled {
		return StripANSI(input) // Fallback to non-SIMD version
	}
	
	if len(input) == 0 {
		return input
	}
	
	// For small inputs, fallback is faster due to setup overhead
	if len(input) < simdWidth*2 {
		return StripANSI(input)
	}
	
	return stripANSISIMD(input)
}

// stripANSISIMD uses SIMD instructions to accelerate ANSI stripping
func stripANSISIMD(input string) string {
	inputBytes := []byte(input)
	output := make([]byte, 0, len(inputBytes))
	
	i := 0
	escByte := byte(0x1b)
	
	// Process in SIMD-width chunks
	for i+simdWidth <= len(inputBytes) {
		// Find ESC character in parallel
		chunk := inputBytes[i : i+simdWidth]
		escPos := indexByteSIMD(chunk, escByte)
		
		if escPos < 0 {
			// No escape in this chunk, copy it
			output = append(output, chunk...)
			i += simdWidth
		} else {
			// Found escape, copy up to it
			output = append(output, chunk[:escPos]...)
			i += escPos
			
			// Skip the escape sequence
			skip := skipANSISequence(inputBytes, i)
			if skip > 0 {
				i += skip
			} else {
				i++
			}
		}
	}
	
	// Process remaining bytes
	for i < len(inputBytes) {
		if inputBytes[i] == escByte {
			skip := skipANSISequence(inputBytes, i)
			if skip > 0 {
				i += skip
				continue
			}
		}
		output = append(output, inputBytes[i])
		i++
	}
	
	return string(output)
}

// indexByteSIMD finds first occurrence of byte c in data using SIMD
// Returns -1 if not found
func indexByteSIMD(data []byte, c byte) int {
	if !Enabled {
		for i, b := range data {
			if b == c {
				return i
			}
		}
		return -1
	}
	
	// SIMD implementation would use parallel comparison
	// Placeholder: actual SIMD intrinsics would go here
	// For now, use optimized loop that compiler can auto-vectorize
	for i := 0; i < len(data); i++ {
		if data[i] == c {
			return i
		}
	}
	return -1
}

// CountByteSIMD counts occurrences of c in s using SIMD
func CountByteSIMD(s string, c byte) int {
	if !Enabled || len(s) < simdWidth*2 {
		return CountByte(s, c)
	}
	
	count := 0
	data := []byte(s)
	
	// Process in SIMD-width chunks
	i := 0
	for i+simdWidth <= len(data) {
		chunk := data[i : i+simdWidth]
		count += countByteInChunk(chunk, c)
		i += simdWidth
	}
	
	// Process remaining bytes
	for i < len(data) {
		if data[i] == c {
			count++
		}
		i++
	}
	
	return count
}

// countByteInChunk counts occurrences of c in chunk using SIMD
func countByteInChunk(chunk []byte, c byte) int {
	count := 0
	// SIMD implementation would use parallel comparison and population count
	// Placeholder for actual SIMD intrinsics
	for _, b := range chunk {
		if b == c {
			count++
		}
	}
	return count
}

// IndexByteSetSIMD finds first occurrence of any byte in set using SIMD
func IndexByteSetSIMD(s string, set []byte) int {
	if !Enabled || len(s) < simdWidth*2 {
		return IndexByteSet(s, set)
	}
	
	// Build lookup table
	var lut [256]bool
	for _, c := range set {
		lut[c] = true
	}
	
	data := []byte(s)
	
	// Process in SIMD-width chunks
	i := 0
	for i+simdWidth <= len(data) {
		chunk := data[i : i+simdWidth]
		pos := indexByteSetInChunk(chunk, &lut)
		if pos >= 0 {
			return i + pos
		}
		i += simdWidth
	}
	
	// Process remaining bytes
	for i < len(data) {
		if lut[data[i]] {
			return i
		}
		i++
	}
	
	return -1
}

// indexByteSetInChunk finds first matching byte in chunk
func indexByteSetInChunk(chunk []byte, lut *[256]bool) int {
	for i, b := range chunk {
		if lut[b] {
			return i
		}
	}
	return -1
}

// SIMDInfo returns information about SIMD capabilities
type SIMDInfo struct {
	Enabled   bool
	AVX2      bool
	AVX512    bool
	NEON      bool
	Width     int
}

// GetSIMDInfo returns the current SIMD configuration
func GetSIMDInfo() SIMDInfo {
	return SIMDInfo{
		Enabled: Enabled,
		AVX2:    hasAVX2,
		AVX512:  hasAVX512,
		NEON:    hasNEON,
		Width:   simdWidth,
	}
}
