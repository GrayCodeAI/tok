package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// LogCrunchFilter folds repetitive INFO/DEBUG logs while preserving warnings/errors and state transitions.
type LogCrunchFilter struct{}

func NewLogCrunchFilter() *LogCrunchFilter { return &LogCrunchFilter{} }

func (f *LogCrunchFilter) Name() string { return "46_log_crunch" }

func (f *LogCrunchFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}
	lines := strings.Split(input, "\n")
	if len(lines) < 20 {
		return input, 0
	}

	seen := map[string]int{}
	out := make([]string, 0, len(lines))
	changed := false
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		if isErrorLine(line) || isWarningLine(line) {
			out = append(out, line)
			continue
		}
		norm := normalizeLogLine(line)
		seen[norm]++
		limit := 2
		if mode == ModeAggressive {
			limit = 1
		}
		if seen[norm] <= limit {
			out = append(out, line)
			continue
		}
		changed = true
	}
	if !changed {
		return input, 0
	}
	out = append(out, "[log-crunch: repetitive logs folded]")
	output := strings.Join(out, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func normalizeLogLine(line string) string {
	lower := strings.ToLower(strings.TrimSpace(line))
	lower = strings.ReplaceAll(lower, "\t", " ")
	parts := strings.Fields(lower)
	if len(parts) > 8 {
		parts = parts[:8]
	}
	return strings.Join(parts, " ")
}
