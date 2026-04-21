package filter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/GrayCodeAI/tok/internal/core"
)

// QuantumLockFilter stabilizes system prompts for KV-cache alignment.
type QuantumLockFilter struct {
	patterns []dynamicPattern
}

type dynamicPattern struct {
	name        string
	regex       *regexp.Regexp
	placeholder string
}

type dynamicFragment struct {
	name        string
	original    string
	placeholder string
}

// NewQuantumLockFilter creates a new KV-cache alignment filter.
func NewQuantumLockFilter() *QuantumLockFilter {
	return &QuantumLockFilter{
		patterns: []dynamicPattern{
			{"iso_date", regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?`), "<DATE>"},
			{"time", regexp.MustCompile(`\b\d{2}:\d{2}:\d{2}\b`), "<TIME>"},
			{"jwt", regexp.MustCompile(`\beyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`), "<JWT>"},
			{"api_key", regexp.MustCompile(`\b(?:sk|rk)-[A-Za-z0-9_-]{16,}\b`), "<API_KEY>"},
			{"uuid", regexp.MustCompile(`\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\b`), "<UUID>"},
			{"unix_ts", regexp.MustCompile(`\b1[5-9]\d{8}\b`), "<TIMESTAMP>"},
			{"hex_id", regexp.MustCompile(`\b[0-9a-fA-F]{32,64}\b`), "<HEX_ID>"},
		},
	}
}

func (f *QuantumLockFilter) Name() string { return "0_quantum_lock" }

// Apply stabilizes content by replacing dynamic fragments with placeholders.
func (f *QuantumLockFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	fragments := f.extractDynamic(input)
	if len(fragments) == 0 {
		return input, 0
	}

	originalTokens := core.EstimateTokens(input)
	stabilized := f.stabilize(input, fragments)
	compressedTokens := core.EstimateTokens(stabilized)

	saved := originalTokens - compressedTokens
	if saved < 0 {
		saved = 0
	}

	return stabilized, saved
}

// extractDynamic finds all dynamic content in the input.
func (f *QuantumLockFilter) extractDynamic(content string) []dynamicFragment {
	seen := make(map[string]dynamicFragment)

	for _, pattern := range f.patterns {
		matches := pattern.regex.FindAllString(content, -1)
		for _, match := range matches {
			if _, exists := seen[match]; !exists {
				seen[match] = dynamicFragment{
					name:        pattern.name,
					original:    match,
					placeholder: pattern.placeholder,
				}
			}
		}
	}

	// Convert map to slice
	fragments := make([]dynamicFragment, 0, len(seen))
	for _, frag := range seen {
		fragments = append(fragments, frag)
	}

	return fragments
}

// stabilize replaces dynamic content and appends context block.
func (f *QuantumLockFilter) stabilize(content string, fragments []dynamicFragment) string {
	stabilized := content

	// Replace each fragment with placeholder
	for _, frag := range fragments {
		stabilized = strings.ReplaceAll(stabilized, frag.original, frag.placeholder)
	}

	// Append dynamic context block
	stabilized += "\n\n---\n<DYNAMIC_CONTEXT>\n"
	for _, frag := range fragments {
		stabilized += fmt.Sprintf("%s: %s\n", frag.name, frag.original)
	}
	stabilized += "</DYNAMIC_CONTEXT>"

	return stabilized
}
