package filter

import (
	"regexp"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// SearchCrunchFilter deduplicates repeated search result lines and keeps top unique hits.
type SearchCrunchFilter struct{}

func NewSearchCrunchFilter() *SearchCrunchFilter { return &SearchCrunchFilter{} }

func (f *SearchCrunchFilter) Name() string { return "47_search_crunch" }

var searchPrefixPattern = regexp.MustCompile(`^\s*\d+[\.|\)]\s+`)

func (f *SearchCrunchFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}
	lines := strings.Split(input, "\n")
	if len(lines) < 12 {
		return input, 0
	}

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
