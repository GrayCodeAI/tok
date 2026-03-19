package filter

import (
	"regexp"
	"strings"
)

// tokenize splits text into words, handling code and natural language
func tokenize(text string) []string {
	// Split on whitespace and punctuation, keeping words together
	// This is a simple tokenizer suitable for compression algorithms
	
	// Replace common separators with spaces
	re := regexp.MustCompile(`[\s\p{P}\p{S}]+`)
	words := re.Split(text, -1)
	
	// Filter empty strings
	var result []string
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word != "" {
			result = append(result, word)
		}
	}
	
	return result
}
