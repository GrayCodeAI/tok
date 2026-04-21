package filter

import (
	"strings"

	"github.com/GrayCodeAI/tok/internal/core"
)

// LatentCollabFilter approximates latent-space collaboration by collapsing
// semantically equivalent multi-agent turns into compact markers.
type LatentCollabFilter struct {
	similarityThreshold float64
}

// NewLatentCollabFilter creates a latent-collaboration inspired filter.
func NewLatentCollabFilter() *LatentCollabFilter {
	return &LatentCollabFilter{similarityThreshold: 0.62}
}

// Name returns the filter name.
func (f *LatentCollabFilter) Name() string { return "37_latent_collab" }

// Apply merges highly similar adjacent agent turns.
func (f *LatentCollabFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")
	turns := parseRoleTurns(lines)
	if len(turns) < 2 {
		return input, 0
	}

	thresh := f.similarityThreshold
	if mode == ModeAggressive {
		thresh = 0.62
	}

	type signature struct {
		role  string
		terms map[string]bool
	}
	var kept []signature
	var out []string
	changed := false

	for _, t := range turns {
		segment := lines[t.start : t.end+1]
		terms := latentTermSet(segment)
		if len(terms) == 0 {
			out = append(out, segment...)
			continue
		}

		merged := false
		for i := len(kept) - 1; i >= 0; i-- {
			if kept[i].role != t.role {
				continue
			}
			if jaccardOverlap(kept[i].terms, terms) >= thresh {
				out = append(out, lines[t.start])
				out = append(out, "[latent-collab: merged]")
				merged = true
				changed = true
				break
			}
		}
		if merged {
			continue
		}

		out = append(out, segment...)
		kept = append(kept, signature{role: t.role, terms: terms})
	}

	if !changed {
		return input, 0
	}

	output := strings.Join(out, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func latentTermSet(lines []string) map[string]bool {
	set := make(map[string]bool)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if isErrorLine(line) || isWarningLine(line) || isCodeLine(line) {
			for _, t := range ltTokenize(line) {
				set[t] = true
			}
			continue
		}
		for _, t := range ltTokenize(line) {
			if len(t) >= 4 {
				set[t] = true
			}
		}
	}
	return set
}
