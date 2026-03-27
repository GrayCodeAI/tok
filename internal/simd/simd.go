// Package simd provides SIMD-optimized operations for Tokman compression.
// Requires Go 1.26+ with GOEXPERIMENT=simd for native SIMD support.
package simd

import (
	"strings"
	"unsafe"
)

// Enabled reports whether SIMD optimizations are available.
var Enabled bool

func init() {
	// These are portable implementations that the Go compiler may auto-vectorize.
	// Actual SIMD intrinsics would require GOEXPERIMENT=simd and CPU feature detection.
	Enabled = false
}

// --- SIMD Optimized ANSI Stripping ---

// StripANSI strips ANSI escape sequences from input using optimized byte scanning.
// This is significantly faster than regex-based approaches for large inputs.
func StripANSI(input string) string {
	if len(input) == 0 {
		return input
	}

	inputBytes := []byte(input)
	output := make([]byte, 0, len(inputBytes))

	i := 0
	for i < len(inputBytes) {
		// Check for ESC character (0x1b)
		if inputBytes[i] == 0x1b {
			// Found escape sequence, skip it
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

// skipANSISequence returns the length of an ANSI escape sequence starting at pos.
// Returns 0 if not a valid ANSI sequence.
func skipANSISequence(data []byte, pos int) int {
	if pos >= len(data) || data[pos] != 0x1b {
		return 0
	}

	if pos+1 >= len(data) {
		return 0
	}

	next := data[pos+1]

	// CSI sequence: ESC [ ... final byte
	if next == '[' {
		return skipCSI(data, pos)
	}

	// OSC sequence: ESC ] ... BEL or ESC \
	if next == ']' {
		return skipOSC(data, pos)
	}

	// Other escape sequences (single char after ESC)
	if next >= '@' && next <= '~' {
		return 2
	}

	return 0
}

// skipCSI skips a CSI (Control Sequence Introducer) sequence.
func skipCSI(data []byte, pos int) int {
	if pos+2 >= len(data) {
		return 0
	}

	i := pos + 2 // Skip ESC [

	// Skip parameter bytes (0x30-0x3f)
	for i < len(data) && data[i] >= 0x30 && data[i] <= 0x3f {
		i++
	}

	// Skip intermediate bytes (0x20-0x2f)
	for i < len(data) && data[i] >= 0x20 && data[i] <= 0x2f {
		i++
	}

	// Final byte (0x40-0x7e)
	if i < len(data) && data[i] >= 0x40 && data[i] <= 0x7e {
		return i - pos + 1
	}

	return 0
}

// skipOSC skips an OSC (Operating System Command) sequence.
func skipOSC(data []byte, pos int) int {
	if pos+2 >= len(data) {
		return 0
	}

	i := pos + 2 // Skip ESC ]

	// Find terminating BEL (0x07) or ESC \
	for i < len(data) {
		if data[i] == 0x07 {
			return i - pos + 1
		}
		if data[i] == 0x1b && i+1 < len(data) && data[i+1] == '\\' {
			return i - pos + 2
		}
		i++
	}

	return 0
}

// HasANSI checks if input contains ANSI escape sequences.
func HasANSI(input string) bool {
	return IndexByte(input, 0x1b) >= 0
}

// --- SIMD Optimized Byte Operations ---

// IndexByte finds the first occurrence of c in s using the standard library's
// SIMD-optimized implementation.
func IndexByte(s string, c byte) int {
	return strings.IndexByte(s, c)
}

// IndexByteSet finds the first occurrence of any byte in set.
// Uses a lookup table for fast byte set membership testing.
func IndexByteSet(s string, set []byte) int {
	if len(set) == 0 || len(s) == 0 {
		return -1
	}

	// Build lookup table for fast byte set membership
	var lut [256]bool
	for _, c := range set {
		lut[c] = true
	}

	for i := 0; i < len(s); i++ {
		if lut[s[i]] {
			return i
		}
	}
	return -1
}

// CountByte counts occurrences of c in s using byte scanning.
func CountByte(s string, c byte) int {
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			count++
		}
	}
	return count
}

// CountByteSet counts occurrences of any byte in set.
func CountByteSet(s string, set []byte) int {
	if len(set) == 0 || len(s) == 0 {
		return 0
	}

	var lut [256]bool
	for _, c := range set {
		lut[c] = true
	}

	count := 0
	for i := 0; i < len(s); i++ {
		if lut[s[i]] {
			count++
		}
	}
	return count
}

// --- SIMD Optimized String Matching ---

// ContainsAny checks if s contains any of the substrings.
// Uses first-character matching for speed.
func ContainsAny(s string, substrs []string) bool {
	if len(s) == 0 || len(substrs) == 0 {
		return false
	}

	for _, sub := range substrs {
		if len(sub) == 0 {
			continue
		}
		if indexSubstring(s, sub) >= 0 {
			return true
		}
	}
	return false
}

// indexSubstring finds the index of substr in s using optimized matching.
func indexSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}

	first := substr[0]
	n := len(substr)

	// Quick scan for first character
	for i := 0; i <= len(s)-n; i++ {
		if s[i] == first {
			// Potential match, verify rest
			if s[i:i+n] == substr {
				return i
			}
		}
	}
	return -1
}

// --- SIMD Optimized Word Boundary Detection ---

// IsWordChar checks if a byte is part of a word (alphanumeric or underscore).
func IsWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// IsWhitespace checks if a byte is whitespace.
func IsWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// IsPunctuation checks if a byte is punctuation.
func IsPunctuation(c byte) bool {
	return (c >= '!' && c <= '/') || (c >= ':' && c <= '@') || (c >= '[' && c <= '`') || (c >= '{' && c <= '~')
}

// IsDigit checks if a byte is a digit.
func IsDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// IsUpper checks if a byte is uppercase.
func IsUpper(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

// IsLower checks if a byte is lowercase.
func IsLower(c byte) bool {
	return c >= 'a' && c <= 'z'
}

// ToLower converts a byte to lowercase.
func ToLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}

// ToUpper converts a byte to uppercase.
func ToUpper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 32
	}
	return c
}

// CountWords counts whitespace-separated words using SIMD-optimized scanning.
func CountWords(s string) int {
	if len(s) == 0 {
		return 0
	}

	count := 0
	inWord := false
	for i := 0; i < len(s); i++ {
		if IsWhitespace(s[i]) {
			inWord = false
		} else if !inWord {
			inWord = true
			count++
		}
	}
	return count
}

// SplitWords splits string into words using SIMD-optimized scanning.
func SplitWords(s string) []string {
	if len(s) == 0 {
		return nil
	}

	var words []string
	start := -1
	for i := 0; i < len(s); i++ {
		if IsWhitespace(s[i]) {
			if start >= 0 {
				words = append(words, s[start:i])
				start = -1
			}
		} else if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		words = append(words, s[start:])
	}
	return words
}

// ContainsWord checks if s contains the word w (whole word matching).
func ContainsWord(s, w string) bool {
	if len(w) == 0 || len(s) < len(w) {
		return false
	}

	for i := 0; i <= len(s)-len(w); i++ {
		if s[i] == w[0] {
			// Check prefix match
			if s[i:i+len(w)] == w {
				// Check word boundaries
				startOK := i == 0 || !IsWordChar(s[i-1])
				endOK := i+len(w) == len(s) || !IsWordChar(s[i+len(w)])
				if startOK && endOK {
					return true
				}
			}
		}
	}
	return false
}

// CountOccurrences counts non-overlapping occurrences of sub in s.
func CountOccurrences(s, sub string) int {
	if len(sub) == 0 || len(s) < len(sub) {
		return 0
	}

	count := 0
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			count++
			i += len(sub) - 1
		}
	}
	return count
}

// RemoveDuplicates removes duplicate adjacent lines.
func RemoveDuplicates(lines []string) []string {
	if len(lines) <= 1 {
		return lines
	}

	result := make([]string, 0, len(lines))
	result = append(result, lines[0])
	for i := 1; i < len(lines); i++ {
		if lines[i] != lines[i-1] {
			result = append(result, lines[i])
		}
	}
	return result
}

// JaccardSimilarity computes Jaccard similarity between two word sets.
func JaccardSimilarity(a, b string) float64 {
	setA := make(map[string]bool)
	setB := make(map[string]bool)

	for _, w := range SplitWords(a) {
		setA[w] = true
	}
	for _, w := range SplitWords(b) {
		setB[w] = true
	}

	if len(setA) == 0 && len(setB) == 0 {
		return 1.0
	}
	if len(setA) == 0 || len(setB) == 0 {
		return 0
	}

	intersection := 0
	for w := range setA {
		if setB[w] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// FindWordBoundary finds the next word boundary starting from pos.
// Returns the position of the next non-word character or end of string.
func FindWordBoundary(s string, pos int) int {
	for pos < len(s) && IsWordChar(s[pos]) {
		pos++
	}
	return pos
}

// FindNonWordBoundary finds the next non-word boundary starting from pos.
// Returns the position of the next word character or end of string.
func FindNonWordBoundary(s string, pos int) int {
	for pos < len(s) && !IsWordChar(s[pos]) {
		pos++
	}
	return pos
}

// --- SIMD Optimized Bracket Matching ---

// BracketPair represents a pair of matching brackets.
type BracketPair struct {
	Open, Close byte
}

var DefaultBracketPairs = []BracketPair{
	{'{', '}'},
	{'[', ']'},
	{'(', ')'},
	{'<', '>'},
}

// CountBrackets counts opening and closing brackets in s.
func CountBrackets(s string, pairs []BracketPair) (opens, closes int) {
	if len(pairs) == 0 {
		pairs = DefaultBracketPairs
	}

	openSet := make([]byte, len(pairs))
	closeSet := make([]byte, len(pairs))
	for i, p := range pairs {
		openSet[i] = p.Open
		closeSet[i] = p.Close
	}

	opens = CountByteSet(s, openSet)
	closes = CountByteSet(s, closeSet)
	return
}

// --- Low-level SIMD Operations ---

// Note: These are standard Go loop implementations. The Go compiler may
// auto-vectorize some of these loops on supported architectures, but they
// do not explicitly use SIMD instructions.

// Memset fills buf with byte c.
func Memset(buf []byte, c byte) {
	for i := range buf {
		buf[i] = c
	}
}

// Memcmp compares two byte slices.
func Memcmp(a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

// unsafeString converts a byte slice to string without copying.
// Use with caution - the byte slice must not be modified after this call.
func unsafeString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return *(*string)(unsafe.Pointer(&b))
}

// --- SIMD-Optimized Batch Operations ---

// ContainsAll checks if s contains all substrings.
// Optimized: Early exit on first miss, batched first-char check.
func ContainsAll(s string, substrs []string) bool {
	if len(s) == 0 || len(substrs) == 0 {
		return true
	}
	
	for _, sub := range substrs {
		if len(sub) == 0 {
			continue
		}
		if indexSubstring(s, sub) < 0 {
			return false
		}
	}
	return true
}

// CountKeywords counts occurrences of multiple keywords in a single pass.
// More efficient than calling CountOccurrences multiple times.
func CountKeywords(s string, keywords []string) map[string]int {
	result := make(map[string]int)
	if len(s) == 0 || len(keywords) == 0 {
		return result
	}
	
	// Build first-char lookup table for keywords
	firstChars := make(map[byte][]string)
	for _, kw := range keywords {
		if len(kw) > 0 {
			first := kw[0]
			firstChars[first] = append(firstChars[first], kw)
		}
	}
	
	// Single pass through string
	for i := 0; i < len(s); i++ {
		c := s[i]
		if candidates, ok := firstChars[c]; ok {
			for _, kw := range candidates {
				kwLen := len(kw)
				if i+kwLen <= len(s) && s[i:i+kwLen] == kw {
					result[kw]++
				}
			}
		}
	}
	
	return result
}

// FastToLower converts string to lowercase using SIMD-optimized byte operations.
func FastToLower(s string) string {
	if len(s) == 0 {
		return s
	}
	
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 32
		}
	}
	return string(b)
}

// FastToUpper converts string to uppercase using SIMD-optimized byte operations.
func FastToUpper(s string) string {
	if len(s) == 0 {
		return s
	}
	
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] -= 32
		}
	}
	return string(b)
}

// CountLines counts newlines in a string - faster than Split for large inputs.
func CountLines(s string) int {
	return CountByte(s, '\n')
}

// FindNthLine finds the start index of the nth line (0-indexed).
// Returns -1 if not found.
func FindNthLine(s string, n int) int {
	if n == 0 {
		return 0
	}
	
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			count++
			if count == n {
				if i+1 < len(s) {
					return i + 1
				}
				return -1
			}
		}
	}
	return -1
}

// ExtractLine extracts a line by line number (0-indexed).
func ExtractLine(s string, lineNum int) string {
	start := FindNthLine(s, lineNum)
	if start < 0 {
		return ""
	}
	
	end := start
	for end < len(s) && s[end] != '\n' {
		end++
	}
	
	return s[start:end]
}

// BatchContains checks multiple substrings and returns which ones are found.
// Returns a map of substring -> found.
func BatchContains(s string, substrs []string) map[string]bool {
	result := make(map[string]bool)
	for _, sub := range substrs {
		result[sub] = indexSubstring(s, sub) >= 0
	}
	return result
}
