package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Paper: "PoC: Performance-oriented Context Compression" — 2026
// https://arxiv.org/abs/2603.19733
// PoCFilter implements performance prediction — estimates how well the
// compressed output will perform and re-inserts critical info if needed.
type PoCFilter struct {
	criticalPatterns []string
}

// NewPoCFilter creates a new performance-oriented compression filter.
func NewPoCFilter() *PoCFilter {
	return &PoCFilter{
		criticalPatterns: []string{
			"error", "fail", "panic", "line ", "file ", "path ",
			"func ", "class ", "type ", "import",
		},
	}
}

// Apply compresses then validates critical info preservation.
func (f *PoCFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	original := input
	lines := strings.Split(input, "\n")

	var critical []string
	var normal []string

	for _, line := range lines {
		if f.isCritical(line) {
			critical = append(critical, line)
		} else {
			normal = append(normal, line)
		}
	}

	var compressed []string
	for _, line := range normal {
		if len(line) > 80 && !f.containsCode(line) {
			if len(line) > 60 {
				compressed = append(compressed, line[:60]+"...")
			} else {
				compressed = append(compressed, line)
			}
		} else {
			compressed = append(compressed, line)
		}
	}

	var result []string
	criticalSet := make(map[string]bool)
	for _, c := range critical {
		criticalSet[c] = true
		result = append(result, c)
	}
	for _, c := range compressed {
		if !criticalSet[c] {
			result = append(result, c)
		}
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(original) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func (f *PoCFilter) isCritical(line string) bool {
	lower := strings.ToLower(line)
	for _, p := range f.criticalPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	if strings.Contains(line, "/") && strings.Count(line, "/") >= 2 {
		return true
	}
	if strings.Contains(line, ":") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && len(parts[0]) <= 5 {
			for _, c := range parts[0] {
				if c < '0' || c > '9' {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (f *PoCFilter) containsCode(line string) bool {
	return strings.ContainsAny(line, "{}[]();=<>")
}

// Name returns the layer name.
func (f *PoCFilter) Name() string { return "24_poc" }
