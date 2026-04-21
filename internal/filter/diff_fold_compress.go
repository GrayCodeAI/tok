package filter

import (
	"fmt"
	"strings"

	"github.com/GrayCodeAI/tok/internal/core"
)

// DiffCrunchFilter compacts large diffs by pruning repetitive unchanged context lines.
type DiffCrunchFilter struct{}

func NewDiffCrunchFilter() *DiffCrunchFilter { return &DiffCrunchFilter{} }

func (f *DiffCrunchFilter) Name() string { return "48_diff_crunch" }

type diffHunk struct {
	header      string
	contextPre  []string
	changes     []string
	contextPost []string
}

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

	hunks := parseUnifiedDiff(lines)
	if len(hunks) == 0 {
		return input, 0
	}

	contextWindow := 3
	if mode == ModeAggressive {
		contextWindow = 2
	}

	out := make([]string, 0, len(lines))
	changed := false

	for _, hunk := range hunks {
		out = append(out, hunk.header)

		// Fold pre-context
		if len(hunk.contextPre) > contextWindow {
			out = append(out, hunk.contextPre[:contextWindow]...)
			out = append(out, fmt.Sprintf("[... %d context lines folded ...]", len(hunk.contextPre)-contextWindow))
			changed = true
		} else {
			out = append(out, hunk.contextPre...)
		}

		// Always keep changes
		out = append(out, hunk.changes...)

		// Fold post-context
		if len(hunk.contextPost) > contextWindow {
			out = append(out, hunk.contextPost[:contextWindow]...)
			out = append(out, fmt.Sprintf("[... %d context lines folded ...]", len(hunk.contextPost)-contextWindow))
			changed = true
		} else {
			out = append(out, hunk.contextPost...)
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

func parseUnifiedDiff(lines []string) []diffHunk {
	hunks := []diffHunk{}
	var current *diffHunk
	inChanges := false

	for _, line := range lines {
		// Hunk header
		if strings.HasPrefix(line, "@@") {
			if current != nil {
				hunks = append(hunks, *current)
			}
			current = &diffHunk{header: line}
			inChanges = false
			continue
		}

		if current == nil {
			continue
		}

		// File headers
		if strings.HasPrefix(line, "diff --git") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			current.header = line
			continue
		}

		// Changes
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			current.changes = append(current.changes, line)
			inChanges = true
			continue
		}

		// Context lines
		if !inChanges {
			current.contextPre = append(current.contextPre, line)
		} else {
			current.contextPost = append(current.contextPost, line)
		}
	}

	if current != nil {
		hunks = append(hunks, *current)
	}

	return hunks
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
