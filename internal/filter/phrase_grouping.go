package filter

import (
	"regexp"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// PhraseGroupingFilter implements dependency-based phrase grouping for compression.
// Research Source: "CompactPrompt: A Unified Pipeline for Prompt Data Compression" (Oct 2025)
// Key Innovation: Group related tokens using syntactic dependency analysis before
// compression, preserving semantic coherence better than token-level pruning.
//
// This identifies noun phrases, verb phrases, and prepositional phrases as
// atomic compression units, preventing the separation of semantically linked tokens.
type PhraseGroupingFilter struct {
	config PhraseGroupConfig
}

// PhraseGroupConfig holds configuration for phrase grouping
type PhraseGroupConfig struct {
	// Enabled controls whether the filter is active
	Enabled bool

	// MinContentLength is the minimum character length to apply
	MinContentLength int

	// MaxPhraseSize is maximum tokens in a phrase group
	MaxPhraseSize int
}

// DefaultPhraseGroupConfig returns default configuration
func DefaultPhraseGroupConfig() PhraseGroupConfig {
	return PhraseGroupConfig{
		Enabled:          true,
		MinContentLength: 200,
		MaxPhraseSize:    8,
	}
}

// NewPhraseGroupingFilter creates a new phrase grouping filter
func NewPhraseGroupingFilter() *PhraseGroupingFilter {
	return &PhraseGroupingFilter{
		config: DefaultPhraseGroupConfig(),
	}
}

// Name returns the filter name
func (f *PhraseGroupingFilter) Name() string {
	return "phrase_group"
}

// phraseGroup represents a semantic phrase unit
type phraseGroup struct {
	words []string
	score float64
}

// Apply applies dependency-based phrase grouping
func (f *PhraseGroupingFilter) Apply(input string, mode Mode) (string, int) {
	if !f.config.Enabled || mode == ModeNone {
		return input, 0
	}

	if len(input) < f.config.MinContentLength {
		return input, 0
	}

	originalTokens := core.EstimateTokens(input)

	// Process line by line
	lines := strings.Split(input, "\n")
	var result strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result.WriteString("\n")
			continue
		}

		groups := f.groupPhrases(trimmed)
		compressed := f.compressGroups(groups, mode)
		result.WriteString(compressed)
		result.WriteString("\n")
	}

	output := strings.TrimSpace(result.String())
	finalTokens := core.EstimateTokens(output)
	saved := originalTokens - finalTokens
	if saved < 3 {
		return input, 0
	}

	return output, saved
}

// groupPhrases groups words into semantic phrases
func (f *PhraseGroupingFilter) groupPhrases(line string) []phraseGroup {
	words := strings.Fields(line)
	var groups []phraseGroup

	i := 0
	for i < len(words) {
		// Try to match phrase patterns starting at this position
		matched := false

		// Noun phrase: article + adjective* + noun
		if np := f.matchNounPhrase(words, i); np != nil {
			groups = append(groups, *np)
			i += len(np.words)
			matched = true
		}

		// Verb phrase: adverb? + verb + particle/preposition?
		if !matched {
			if vp := f.matchVerbPhrase(words, i); vp != nil {
				groups = append(groups, *vp)
				i += len(vp.words)
				matched = true
			}
		}

		// Prepositional phrase: preposition + noun phrase
		if !matched {
			if pp := f.matchPrepPhrase(words, i); pp != nil {
				groups = append(groups, *pp)
				i += len(pp.words)
				matched = true
			}
		}

		// Single word group
		if !matched {
			groups = append(groups, phraseGroup{
				words: []string{words[i]},
				score: 0.5,
			})
			i++
		}
	}

	return groups
}

// matchNounPhrase detects noun phrases (determiner + adjectives + noun)
func (f *PhraseGroupingFilter) matchNounPhrase(words []string, start int) *phraseGroup {
	if start >= len(words) {
		return nil
	}

	// Check for determiner
	hasDet := false
	word := strings.ToLower(words[start])
	if word == "the" || word == "a" || word == "an" || word == "this" ||
		word == "that" || word == "these" || word == "those" ||
		word == "my" || word == "your" || word == "its" || word == "our" {
		hasDet = true
	}

	if !hasDet {
		// Check for adjective + noun pattern
		if start+1 < len(words) && f.isAdjective(words[start]) && f.isNoun(words[start+1]) {
			return &phraseGroup{
				words: words[start : start+2],
				score: 0.7,
			}
		}
		return nil
	}

	// Collect adjective* + noun sequence
	end := start + 1
	for end < len(words) && end-start < f.config.MaxPhraseSize {
		if f.isAdjective(words[end]) {
			end++
		} else if f.isNoun(words[end]) {
			end++
			break
		} else {
			break
		}
	}

	if end > start+1 {
		return &phraseGroup{
			words: words[start:end],
			score: 0.8,
		}
	}
	return nil
}

// matchVerbPhrase detects verb phrases (adverb? + verb + particle?)
func (f *PhraseGroupingFilter) matchVerbPhrase(words []string, start int) *phraseGroup {
	if start >= len(words) {
		return nil
	}

	// Check for adverb + verb
	if start+1 < len(words) && f.isAdverb(words[start]) && f.isVerb(words[start+1]) {
		end := start + 2
		// Optional particle
		if end < len(words) && f.isParticle(words[end]) {
			end++
		}
		return &phraseGroup{
			words: words[start:end],
			score: 0.7,
		}
	}

	// Just verb + particle
	if f.isVerb(words[start]) && start+1 < len(words) && f.isParticle(words[start+1]) {
		return &phraseGroup{
			words: words[start : start+2],
			score: 0.6,
		}
	}

	return nil
}

// matchPrepPhrase detects prepositional phrases (preposition + noun phrase)
func (f *PhraseGroupingFilter) matchPrepPhrase(words []string, start int) *phraseGroup {
	if start >= len(words) || !f.isPreposition(words[start]) {
		return nil
	}

	// Look for noun phrase after preposition
	np := f.matchNounPhrase(words, start+1)
	if np != nil {
		allWords := append([]string{words[start]}, np.words...)
		return &phraseGroup{
			words: allWords,
			score: 0.6,
		}
	}

	// Preposition + single noun
	if start+1 < len(words) && f.isNoun(words[start+1]) {
		return &phraseGroup{
			words: words[start : start+2],
			score: 0.5,
		}
	}

	return nil
}

// compressGroups compresses phrase groups based on mode
func (f *PhraseGroupingFilter) compressGroups(groups []phraseGroup, mode Mode) string {
	var parts []string

	for _, g := range groups {
		if len(g.words) == 1 {
			parts = append(parts, g.words[0])
			continue
		}

		// In aggressive mode, compress low-scoring groups
		if mode == ModeAggressive && g.score < 0.5 {
			// Keep only first and last word of the group
			compressed := g.words[0] + "…" + g.words[len(g.words)-1]
			parts = append(parts, compressed)
		} else {
			parts = append(parts, strings.Join(g.words, " "))
		}
	}

	return strings.Join(parts, " ")
}

// Simple heuristic word class detection (no ML required)
var (
	articles     = map[string]bool{"the": true, "a": true, "an": true}
	determiners  = map[string]bool{"the": true, "a": true, "an": true, "this": true, "that": true, "these": true, "those": true, "my": true, "your": true, "its": true, "our": true, "their": true}
	prepositions = map[string]bool{"in": true, "on": true, "at": true, "to": true, "for": true, "with": true, "from": true, "by": true, "of": true, "about": true, "into": true, "through": true, "during": true, "before": true, "after": true, "above": true, "below": true, "between": true}
	adverbs      = map[string]bool{"very": true, "quickly": true, "slowly": true, "easily": true, "often": true, "never": true, "always": true, "already": true, "just": true, "also": true, "still": true, "now": true, "then": true, "here": true, "there": true}
	particles    = map[string]bool{"up": true, "out": true, "off": true, "away": true, "back": true, "down": true, "over": true, "around": true}
	verbSuffixes = []string{"ing", "ed", "ize", "ise", "ate", "ify"}
	nounSuffixes = []string{"tion", "sion", "ment", "ness", "ity", "er", "or", "ist", "ism"}
	adjSuffixes  = []string{"able", "ible", "ful", "less", "ous", "ive", "ial", "al", "ent", "ant"}
)

func (f *PhraseGroupingFilter) isNoun(w string) bool {
	w = strings.ToLower(w)
	if determiners[w] || prepositions[w] {
		return false
	}
	// Heuristic: ends with noun suffix or is capitalized
	for _, s := range nounSuffixes {
		if strings.HasSuffix(w, s) {
			return true
		}
	}
	return len(w) > 2 && w[0] >= 'A' && w[0] <= 'Z'
}

func (f *PhraseGroupingFilter) isVerb(w string) bool {
	w = strings.ToLower(w)
	commonVerbs := map[string]bool{
		"is": true, "are": true, "was": true, "were": true,
		"have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true,
		"will": true, "would": true, "shall": true, "should": true,
		"can": true, "could": true, "may": true, "might": true, "must": true,
		"make": true, "get": true, "go": true, "come": true,
		"take": true, "see": true, "know": true, "think": true, "give": true,
		"find": true, "tell": true, "become": true, "leave": true, "put": true,
		"mean": true, "keep": true, "let": true, "begin": true, "show": true,
		"set": true, "add": true, "use": true, "call": true, "try": true,
		"need": true, "want": true, "create": true, "build": true, "run": true,
		"compile": true, "execute": true, "install": true, "update": true,
		"delete": true, "remove": true, "start": true, "stop": true,
	}
	if commonVerbs[w] {
		return true
	}
	for _, s := range verbSuffixes {
		if strings.HasSuffix(w, s) && len(w) > len(s)+2 {
			return true
		}
	}
	return false
}

func (f *PhraseGroupingFilter) isAdjective(w string) bool {
	w = strings.ToLower(w)
	for _, s := range adjSuffixes {
		if strings.HasSuffix(w, s) && len(w) > len(s)+2 {
			return true
		}
	}
	commonAdj := map[string]bool{
		"new": true, "old": true, "good": true, "bad": true, "big": true,
		"small": true, "long": true, "short": true, "high": true, "low": true,
		"first": true, "last": true, "next": true, "main": true, "other": true,
	}
	return commonAdj[w]
}

func (f *PhraseGroupingFilter) isAdverb(w string) bool {
	w = strings.ToLower(w)
	return adverbs[w] || strings.HasSuffix(w, "ly")
}

func (f *PhraseGroupingFilter) isPreposition(w string) bool {
	return prepositions[strings.ToLower(w)]
}

func (f *PhraseGroupingFilter) isParticle(w string) bool {
	return particles[strings.ToLower(w)]
}

// Dependency patterns for common CLI/code output structures
var (
	// "function_name(args)" → treat as single unit
	funcCallPattern = regexp.MustCompile(`\w+\([^)]*\)`)
	// "key=value" → treat as single unit
	kvPattern = regexp.MustCompile(`\w+=\S+`)
	// "path/to/file" → treat as single unit
	pathPattern = regexp.MustCompile(`[\w./\\]+\.\w+`)
)

