package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// HypernymCompressor implements Mercury-style word-level semantic compression.
// Research Source: "Hypernym Mercury: Token Optimization Through Semantic Field
// Constriction And Reconstruction From Hypernyms" (May 2025)
// Key Innovation: Replace detailed tokens with hypernym concepts when aggressive
// compression needed. 90%+ token reduction with controllable granularity.
//
// Example: "The quick brown fox jumps over the lazy dog" →
// "animal action location animal quality" (at hypernym level)
//
// This uses a built-in hypernym hierarchy for common concepts.
// The granularity is controlled by the compression mode.
type HypernymCompressor struct {
	config    HypernymConfig
	hierarchy map[string]string // word -> hypernym
	domain    map[string]string // domain-specific hypernyms
}

// HypernymConfig holds configuration for hypernym compression
type HypernymConfig struct {
	// Enabled controls whether the compressor is active
	Enabled bool

	// MinContentLength is minimum chars to apply
	MinContentLength int

	// MaxDetailLevel controls granularity: 1=most abstract, 3=most detailed
	MaxDetailLevel int

	// PreserveKeywords keeps technical terms uncompressed
	PreserveKeywords bool
}

// DefaultHypernymConfig returns default configuration
func DefaultHypernymConfig() HypernymConfig {
	return HypernymConfig{
		Enabled:          true,
		MinContentLength: 300,
		MaxDetailLevel:   2,
		PreserveKeywords: true,
	}
}

// NewHypernymCompressor creates a new hypernym compressor
func NewHypernymCompressor() *HypernymCompressor {
	return &HypernymCompressor{
		config:    DefaultHypernymConfig(),
		hierarchy: initHypernymHierarchy(),
		domain:    initDomainHypernyms(),
	}
}

// Name returns the filter name
func (h *HypernymCompressor) Name() string {
	return "hypernym"
}

// Apply applies hypernym-based concept compression
func (h *HypernymCompressor) Apply(input string, mode Mode) (string, int) {
	if !h.config.Enabled || mode == ModeNone {
		return input, 0
	}

	if len(input) < h.config.MinContentLength {
		return input, 0
	}

	originalTokens := core.EstimateTokens(input)

	// Only apply in aggressive mode or when content is very redundant
	if mode != ModeAggressive {
		// Check redundancy first
		redundancy := h.computeRedundancy(input)
		if redundancy < 0.4 {
			return input, 0
		}
	}

	// Process line by line
	lines := strings.Split(input, "\n")
	var result strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result.WriteString("\n")
			continue
		}

		compressed := h.compressLine(trimmed, mode)
		result.WriteString(compressed)
		result.WriteString("\n")
	}

	output := strings.TrimSpace(result.String())
	finalTokens := core.EstimateTokens(output)
	saved := originalTokens - finalTokens
	if saved < 5 {
		return input, 0
	}

	return output, saved
}

// compressLine compresses a single line using hypernym substitution
func (h *HypernymCompressor) compressLine(line string, mode Mode) string {
	words := strings.Fields(line)
	var compressed []string

	for _, word := range words {
		lower := strings.ToLower(strings.Trim(word, ".,;:!?\"'()[]{}"))

		// Preserve technical keywords
		if h.config.PreserveKeywords && h.isTechnicalTerm(lower) {
			compressed = append(compressed, word)
			continue
		}

		// Try hypernym substitution
		if hypernym, ok := h.hierarchy[lower]; ok {
			compressed = append(compressed, hypernym)
		} else if domainHyp, ok := h.domain[lower]; ok {
			compressed = append(compressed, domainHyp)
		} else {
			compressed = append(compressed, word)
		}
	}

	result := strings.Join(compressed, " ")

	// In aggressive mode, deduplicate consecutive identical concepts
	if mode == ModeAggressive {
		result = h.deduplicateConcepts(result)
	}

	return result
}

// computeRedundancy estimates how redundant the content is
func (h *HypernymCompressor) computeRedundancy(input string) float64 {
	words := strings.Fields(strings.ToLower(input))
	if len(words) < 10 {
		return 0
	}

	unique := make(map[string]bool)
	for _, w := range words {
		unique[w] = true
	}

	return 1.0 - (float64(len(unique)) / float64(len(words)))
}

// isTechnicalTerm checks if a word is a technical term that should be preserved
func (h *HypernymCompressor) isTechnicalTerm(word string) bool {
	technicalTerms := map[string]bool{
		"func": true, "function": true, "def": true, "class": true,
		"import": true, "package": true, "module": true, "struct": true,
		"interface": true, "type": true, "const": true, "var": true,
		"return": true, "if": true, "else": true, "for": true, "while": true,
		"error": true, "err": true, "nil": true, "null": true,
		"true": true, "false": true, "http": true, "api": true,
		"url": true, "json": true, "xml": true, "sql": true,
		"git": true, "docker": true, "npm": true, "cargo": true,
	}
	return technicalTerms[word]
}

// deduplicateConcepts removes consecutive duplicate hypernyms
func (h *HypernymCompressor) deduplicateConcepts(input string) string {
	words := strings.Fields(input)
	if len(words) < 2 {
		return input
	}

	var result []string
	prev := ""
	for _, w := range words {
		if w != prev {
			result = append(result, w)
			prev = w
		}
	}
	return strings.Join(result, " ")
}

// initHypernymHierarchy builds the word-to-hypernym mapping
func initHypernymHierarchy() map[string]string {
	return map[string]string{
		// Animals
		"fox": "animal", "dog": "animal", "cat": "animal", "bird": "animal",
		"fish": "animal", "horse": "animal", "cow": "animal", "mouse": "animal",
		"wolf": "animal", "bear": "animal", "lion": "animal", "tiger": "animal",
		"rabbit": "animal", "deer": "animal", "eagle": "animal", "hawk": "animal",

		// Actions
		"run": "action", "walk": "action", "jump": "action", "swim": "action",
		"fly": "action", "eat": "action", "drink": "action", "sleep": "action",
		"write": "action", "read": "action", "build": "action", "create": "action",
		"delete": "action", "update": "action", "modify": "action", "change": "action",
		"execute": "action", "compile": "action", "install": "action",
		"start": "action", "stop": "action", "open": "action", "close": "action",
		"send": "action", "receive": "action", "fetch": "action", "push": "action",
		"pull": "action", "merge": "action", "commit": "action", "deploy": "action",

		// Objects
		"file": "object", "directory": "object", "folder": "object", "path": "object",
		"database": "object", "table": "object", "column": "object", "row": "object",
		"server": "location", "client": "location", "endpoint": "location", "service": "location",
		"container": "object", "image": "object", "volume": "object", "network": "object",
		"config": "object", "setting": "object", "option": "object", "parameter": "object",
		"variable": "object", "constant": "object", "function": "object", "method": "object",

		// Qualities
		"quick": "quality", "slow": "quality", "fast": "quality", "lazy": "quality",
		"good": "quality", "bad": "quality", "big": "quality", "small": "quality",
		"long": "quality", "short": "quality", "high": "quality", "low": "quality",
		"new": "quality", "old": "quality", "important": "quality", "critical": "quality",
		"verbose": "quality", "compact": "quality", "clean": "quality", "dirty": "quality",

		// Locations
		"forest": "location", "field": "location", "river": "location", "mountain": "location",
		"home": "location", "office": "location", "cloud": "location",
		"repository": "location", "branch": "location", "remote": "location", "local": "location",
	}
}

// initDomainHypernyms builds software-domain specific hypernyms
func initDomainHypernyms() map[string]string {
	return map[string]string{
		// Common programming concepts
		"string": "data_type", "integer": "data_type", "float": "data_type",
		"boolean": "data_type", "array": "data_type", "map": "data_type",
		"slice": "data_type", "struct": "data_type", "interface": "data_type",
		"pointer": "data_type", "channel": "data_type",

		// Git concepts
		"commit": "git_op", "merge": "git_op", "rebase": "git_op",
		"branch": "git_op", "checkout": "git_op", "stash": "git_op",
		"diff": "git_op", "log": "git_op", "status": "git_op",

		// File operations
		"read": "io_op", "write": "io_op", "append": "io_op",
		"copy": "io_op", "move": "io_op", "rename": "io_op",
		"chmod": "io_op", "chown": "io_op",
	}
}
