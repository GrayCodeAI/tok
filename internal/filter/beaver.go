package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Paper: "BEAVER: Structure-Aware Page Selection" — 2026
// https://arxiv.org/abs/2603.19635
// BEAVERFilter implements structure-aware hierarchical compression — treats
// content as pages/sections and selects based on structural importance.
type BEAVERFilter struct {
	maxSections int
}

// NewBEAVERFilter creates a new structure-aware page selection filter.
func NewBEAVERFilter() *BEAVERFilter {
	return &BEAVERFilter{maxSections: 5}
}

// Apply selects structurally important sections.
func (f *BEAVERFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	original := input
	sections := f.splitSections(input)

	type section struct {
		content string
		score   float64
	}
	scored := make([]section, 0, len(sections))
	for _, s := range sections {
		scored = append(scored, section{content: s, score: f.sectionScore(s)})
	}

	for i := 1; i < len(scored); i++ {
		for j := i; j > 0 && scored[j].score > scored[j-1].score; j-- {
			scored[j], scored[j-1] = scored[j-1], scored[j]
		}
	}

	keep := f.maxSections
	if len(scored) < keep {
		keep = len(scored)
	}
	var result []string
	for i := 0; i < keep; i++ {
		result = append(result, scored[i].content)
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(original) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func (f *BEAVERFilter) splitSections(input string) []string {
	lines := strings.Split(input, "\n")
	var sections []string
	var current []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "###") || strings.HasPrefix(trimmed, "---") ||
			strings.HasPrefix(trimmed, "===") || strings.HasPrefix(trimmed, "```") {
			if len(current) > 0 {
				sections = append(sections, strings.Join(current, "\n"))
				current = nil
			}
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		sections = append(sections, strings.Join(current, "\n"))
	}
	return sections
}

func (f *BEAVERFilter) sectionScore(section string) float64 {
	score := 0.0
	lines := strings.Split(section, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "###") || strings.HasPrefix(trimmed, "##") {
			score += 3.0
		}
		if strings.HasPrefix(trimmed, "```") {
			score += 2.0
		}
		if strings.Contains(trimmed, "error") || strings.Contains(trimmed, "fail") {
			score += 2.0
		}
		if strings.Contains(trimmed, "func ") || strings.Contains(trimmed, "class ") {
			score += 1.5
		}
		if len(trimmed) > 0 {
			score += 0.1
		}
	}

	if strings.Contains(section, "```") {
		score += 5.0
	}

	return score / float64(len(lines)+1)
}

// Name returns the layer name.
func (f *BEAVERFilter) Name() string { return "25_beaver" }
