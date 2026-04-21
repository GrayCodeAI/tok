package filter

import (
	"regexp"
	"strings"

	"github.com/GrayCodeAI/tok/internal/core"
)

// SearchCrunchFilter deduplicates repeated search result lines and keeps top unique hits.
type SearchCrunchFilter struct{}

func NewSearchCrunchFilter() *SearchCrunchFilter { return &SearchCrunchFilter{} }

func (f *SearchCrunchFilter) Name() string { return "47_search_crunch" }

var searchPrefixPattern = regexp.MustCompile(`^\s*\d+[\.|\)]\s+`)

type searchResult struct {
	rank    int
	snippet string
	hash    uint64
}

func (f *SearchCrunchFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}
	lines := strings.Split(input, "\n")
	if len(lines) < 12 {
		return input, 0
	}

	results := parseSearchResults(lines)
	if len(results) == 0 {
		// Fallback to simple dedup
		return f.simpleDedup(lines, mode, input)
	}

	// Deduplicate by snippet similarity
	deduplicated := deduplicateBySnippet(results)

	maxResults := 60
	if mode == ModeAggressive {
		maxResults = 35
	}

	kept := deduplicated
	if len(kept) > maxResults {
		kept = kept[:maxResults]
	}

	if len(kept) == len(results) {
		return input, 0
	}

	out := formatSearchResults(kept)
	out = append(out, "[search-crunch: duplicate hits pruned]")
	output := strings.Join(out, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func (f *SearchCrunchFilter) simpleDedup(lines []string, mode Mode, input string) (string, int) {
	seen := map[string]bool{}
	out := make([]string, 0, len(lines))
	changed := false
	maxUnique := 60
	if mode == ModeAggressive {
		maxUnique = 35
	}
	unique := 0
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		if isErrorLine(line) || isWarningLine(line) {
			out = append(out, line)
			continue
		}
		norm := searchPrefixPattern.ReplaceAllString(strings.ToLower(trim), "")
		norm = strings.Join(strings.Fields(norm), " ")
		if seen[norm] {
			changed = true
			continue
		}
		seen[norm] = true
		if unique >= maxUnique {
			changed = true
			continue
		}
		unique++
		out = append(out, line)
	}

	if !changed {
		return input, 0
	}
	out = append(out, "[search-crunch: duplicate hits pruned]")
	output := strings.Join(out, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func parseSearchResults(lines []string) []searchResult {
	results := []searchResult{}
	rank := 0

	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}

		// Simple heuristic: lines with URLs or numbered prefixes
		if strings.Contains(line, "http") || searchPrefixPattern.MatchString(line) {
			rank++
			snippet := searchPrefixPattern.ReplaceAllString(trim, "")
			hash := SimHash(snippet)
			results = append(results, searchResult{
				rank:    rank,
				snippet: snippet,
				hash:    hash,
			})
		}
	}

	return results
}

func deduplicateBySnippet(results []searchResult) []searchResult {
	if len(results) == 0 {
		return results
	}

	kept := []searchResult{results[0]}
	for i := 1; i < len(results); i++ {
		isDuplicate := false
		for _, prev := range kept {
			if HammingDistance(results[i].hash, prev.hash) <= 3 {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			kept = append(kept, results[i])
		}
	}

	return kept
}

func formatSearchResults(results []searchResult) []string {
	out := make([]string, 0, len(results))
	for i, result := range results {
		out = append(out, string(rune(i+49))+". "+result.snippet)
	}
	return out
}
