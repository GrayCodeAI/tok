package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// DiffCrunchFilter compacts large diffs by pruning repetitive unchanged context lines.
type DiffCrunchFilter struct{}

func NewDiffCrunchFilter() *DiffCrunchFilter { return &DiffCrunchFilter{} }

func (f *DiffCrunchFilter) Name() string { return "48_diff_crunch" }

func (f *DiffCrunchFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}
	lines := strings.Split(input, "\n")
	if len(lines) < 20 {
		return input, 0
	}
	if !looksLikeDiff(lines) {
		return input, 0
	}

	out := make([]string, 0, len(lines))
	contextRun := 0
	maxContext := 3
	if mode == ModeAggressive {
		maxContext = 2
	}
	changed := false
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(line, "diff --git") || strings.HasPrefix(line, "@@") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			contextRun = 0
			out = append(out, line)
			continue
		}
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			contextRun = 0
			out = append(out, line)
			continue
		}
		if trim == "" {
			continue
		}
		contextRun++
		if contextRun <= maxContext {
			out = append(out, line)
		} else {
			changed = true
		}
	}

	if !changed {
		return input, 0
	}
	out = append(out, "[diff-crunch: context folded]")
	output := strings.Join(out, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func looksLikeDiff(lines []string) bool {
	hits := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") || strings.HasPrefix(line, "@@") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			hits++
		}
	}
	return hits >= 2
}
