package filter

import (
	"regexp"
	"strings"
)

var tokenizeRe = regexp.MustCompile(`[\s\p{P}\p{S}]+`)

// tokenize splits text into words, handling code and natural language
func tokenize(text string) []string {
	// Split on whitespace and punctuation, keeping words together
	// This is a simple tokenizer suitable for compression algorithms

	// Pre-allocate with estimated capacity to avoid reallocations
	estimatedCount := len(text) / 5 // Rough estimate: average word length ~5
	result := make([]string, 0, estimatedCount)

	// Replace common separators with spaces
	words := tokenizeRe.Split(text, -1)

	// Filter empty strings
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word != "" {
			result = append(result, word)
		}
	}

	return result
}
