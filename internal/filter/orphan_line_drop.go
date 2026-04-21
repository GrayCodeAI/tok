package filter

import (
	"strings"

	"github.com/GrayCodeAI/tok/internal/core"
)

// Paper: "SlimInfer: Accelerating Long-Context LLM Inference via Dynamic Token Pruning"
// arXiv:2508.06447 — 2025
//
// SlimInferFilter implements the core insight from SlimInfer: drop "orphan" lines —
// lines whose terms do not appear in any other line of the output.
//
// This is the complement to MarginalInfoGainFilter:
//
//	MIG  — drops lines that contribute NO NEW terms to the retained set
//	Slim — drops lines whose terms are REFERENCED BY NO OTHER line
//
// Together they remove both ends of the information graph:
//   - Lines that are fully covered by others  (MIG)
//   - Lines that are fully isolated from others (Slim)
//
// Algorithm:
//  1. Build a term → {line indices} inverted index
//  2. For each line, compute refScore = number of OTHER lines that share ≥1 term
//  3. Lines with refScore < threshold are "orphans" → prune
//  4. Structural lines (errors, headings, first/last) are always kept
//
// Threshold:
//
//	ModeMinimal:   refScore ≥ 1  (at least one other line shares a term)
//	ModeAggressive: refScore ≥ 2  (at least two other lines share a term)
type SlimInferFilter struct {
	minLineLen int // lines shorter than this are always kept
}

// NewSlimInferFilter creates a new SlimInfer orphan-line pruner.
func NewSlimInferFilter() *SlimInferFilter {
	return &SlimInferFilter{minLineLen: 12}
}

// Name returns the filter name.
func (f *SlimInferFilter) Name() string { return "30_slim_infer" }

// Apply drops orphan lines (low inter-line term reference count).
func (f *SlimInferFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")
	if len(lines) < 6 {
		return input, 0
	}

	threshold := 1
	if mode == ModeAggressive {
		threshold = 2
	}

	// Build term sets per line
	termSets := make([]map[string]bool, len(lines))
	for i, line := range lines {
		termSets[i] = siTermSet(line)
	}

	// Build inverted index: term → set of line indices
	invIdx := make(map[string][]int)
	for i, ts := range termSets {
		for t := range ts {
			invIdx[t] = append(invIdx[t], i)
		}
	}

	// Compute refScore for each line
	refScore := make([]int, len(lines))
	for i, ts := range termSets {
		seen := make(map[int]bool)
		for t := range ts {
			for _, j := range invIdx[t] {
				if j != i && !seen[j] {
					seen[j] = true
					refScore[i]++
				}
			}
		}
	}

	// Determine keep set
	keep := make([]bool, len(lines))
	// Always keep first and last non-empty lines
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			keep[i] = true
			break
		}
	}
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			keep[i] = true
			break
		}
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			keep[i] = true // preserve blank lines (structure)
			continue
		}
		if len(trimmed) < f.minLineLen {
			keep[i] = true // too short to assess
			continue
		}
		if isErrorLine(line) || isWarningLine(line) || isHeadingLine(line) {
			keep[i] = true
			continue
		}
		if refScore[i] >= threshold {
			keep[i] = true
		}
	}

	var result []string
	for i, line := range lines {
		if keep[i] {
			result = append(result, line)
		}
	}
	if len(result) == 0 {
		return input, 0
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

// siTermSet builds a set of significant terms for a line.
func siTermSet(line string) map[string]bool {
	set := make(map[string]bool)
	var word strings.Builder
	for _, ch := range strings.ToLower(line) {
		if (ch >= 'a' && ch <= 'z') || ch == '_' {
			word.WriteRune(ch)
		} else if word.Len() > 0 {
			if w := word.String(); len(w) >= 4 { // ≥4 chars to avoid noise
				set[w] = true
			}
			word.Reset()
		}
	}
	if word.Len() >= 4 {
		set[word.String()] = true
	}
	return set
}
