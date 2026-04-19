package filter

import (
	"strings"

	"github.com/lakshmanpatel/tok/internal/core"
)

// StructuralCollapseFilter compacts repetitive structural boilerplate while preserving semantic anchors.
type StructuralCollapseFilter struct{}

func NewStructuralCollapseFilter() *StructuralCollapseFilter { return &StructuralCollapseFilter{} }

func (f *StructuralCollapseFilter) Name() string { return "49_structural_collapse" }

func (f *StructuralCollapseFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}
	lines := strings.Split(input, "\n")
	if len(lines) < 16 {
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
		if isErrorLine(line) || isWarningLine(line) || isCodeLine(line) {
			out = append(out, line)
			continue
		}
		norm := strings.ToLower(strings.Join(strings.Fields(trim), " "))
		if strings.HasPrefix(norm, "import ") || strings.HasPrefix(norm, "from ") || strings.HasPrefix(norm, "package ") || strings.HasPrefix(norm, "module ") || strings.HasPrefix(norm, "section ") || strings.HasPrefix(norm, "###") {
			seen[norm]++
			limit := 1
			if mode == ModeMinimal {
				limit = 2
			}
			if seen[norm] > limit {
				changed = true
				continue
			}
		}
		out = append(out, line)
	}

	if !changed {
		return input, 0
	}
	out = append(out, "[structural-collapse: repeated boilerplate pruned]")
	output := strings.Join(out, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}
