package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// LightMemFilter reuses previously seen high-signal facts with short references.
type LightMemFilter struct{}

// NewLightMemFilter creates the lightweight memory-augmentation filter.
func NewLightMemFilter() *LightMemFilter { return &LightMemFilter{} }

// Name returns the filter name.
func (f *LightMemFilter) Name() string { return "43_lightmem" }

// Apply detects repeated facts and replaces duplicates with lightweight references.
func (f *LightMemFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")
	if len(lines) < 8 {
		return input, 0
	}

	seen := make(map[string]int, 16)
	out := make([]string, 0, len(lines))
	changed := false
	memID := 1
	for _, line := range lines {
		norm := lightMemNormalize(line)
		if norm == "" {
			out = append(out, line)
			continue
		}
		if id, ok := seen[norm]; ok {
			out = append(out, "[lightmem: reuse #"+itoa(id)+"]")
			changed = true
			continue
		}
		seen[norm] = memID
		memID++
		out = append(out, line)
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

func lightMemNormalize(line string) string {
	trimmed := strings.ToLower(strings.TrimSpace(line))
	if trimmed == "" {
		return ""
	}
	if !(isErrorLine(trimmed) || isWarningLine(trimmed) || isCodeLine(line) || strings.Contains(trimmed, ":") || strings.Contains(trimmed, "file") || strings.Contains(trimmed, "path")) {
		return ""
	}
	toks := ltTokenize(trimmed)
	if len(toks) < 3 {
		return ""
	}
	if len(toks) > 12 {
		toks = toks[:12]
	}
	return strings.Join(toks, " ")
}
