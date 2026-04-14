package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Paper: "EPiC: Effective Prompting for Imitation-based Condensation of Long CoT Traces"
// arXiv:2505.xxxxx — 2025
//
// EPiCFilter identifies "causal edge" lines in chain-of-thought reasoning traces —
// lines that explicitly reference or build upon conclusions from prior steps —
// and protects them from compression.
//
// Causal connectives are the load-bearing joints of a reasoning chain:
//   - "therefore", "thus", "so", "hence", "consequently"
//   - "because", "since", "given that", "due to"
//   - "this means", "which implies", "it follows that"
//   - "building on", "using the result from", "from step N"
//
// Without these connectives, a compressed trace loses its logical continuity.
// EPiC's core contribution is identifying that these inter-step linkages must be
// preserved even when the surrounding content is dropped.
//
// Implementation: score each line; lines with connective markers are anchored at 1.0.
// Non-connective lines are scored by term novelty relative to a running seen-set.
// Lines below the novelty threshold are dropped.
type EPiCFilter struct {
	noveltyThreshold float64 // min fraction of new terms to retain a line
}

// NewEPiCFilter creates a new EPiC causal-edge preservation filter.
func NewEPiCFilter() *EPiCFilter {
	return &EPiCFilter{
		noveltyThreshold: 0.35,
	}
}

// Name returns the filter name.
func (f *EPiCFilter) Name() string { return "32_epic" }

// Apply identifies causal edges and drops low-novelty non-causal lines.
func (f *EPiCFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")
	if len(lines) < 6 {
		return input, 0
	}

	// Only apply to reasoning-heavy content
	if !epicLooksLikeCoT(lines) {
		return input, 0
	}

	threshold := f.noveltyThreshold
	if mode == ModeAggressive {
		threshold = 0.50
	}

	seen := make(map[string]bool)
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result = append(result, line)
			continue
		}

		// Always preserve structural anchors
		if isErrorLine(line) || isWarningLine(line) || isHeadingLine(line) || isCodeLine(line) {
			result = append(result, line)
			epicAddTerms(seen, line)
			continue
		}

		// Causal edge lines are always preserved
		if epicIsCausalEdge(trimmed) {
			result = append(result, line)
			epicAddTerms(seen, line)
			continue
		}

		// Score novelty: fraction of this line's terms not yet seen
		terms := ltTokenize(line)
		if len(terms) == 0 {
			result = append(result, line)
			continue
		}
		newTerms := 0
		for _, t := range terms {
			if !seen[t] {
				newTerms++
			}
		}
		novelty := float64(newTerms) / float64(len(terms))

		if novelty >= threshold {
			result = append(result, line)
			epicAddTerms(seen, line)
		}
	}

	if len(result) == len(lines) {
		return input, 0
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

// epicIsCausalEdge returns true if the line contains a causal/logical connective.
func epicIsCausalEdge(line string) bool {
	lower := strings.ToLower(line)
	causalMarkers := []string{
		"therefore", "thus,", "hence,", "so,", "consequently",
		"because ", "since ", "given that", "due to",
		"this means", "which implies", "it follows",
		"building on", "from step ", "from the previous",
		"using this", "using the result", "we can conclude",
		"in conclusion", "as a result", "for this reason",
	}
	for _, marker := range causalMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// epicLooksLikeCoT returns true if the input has reasoning-trace characteristics.
func epicLooksLikeCoT(lines []string) bool {
	causalCount := 0
	stepCount := 0
	for _, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		if epicIsCausalEdge(lower) {
			causalCount++
		}
		if isReasoningLine(line) {
			stepCount++
		}
	}
	return causalCount >= 2 || stepCount >= 3
}

// epicAddTerms adds all tokens from a line to the seen set.
func epicAddTerms(seen map[string]bool, line string) {
	for _, t := range ltTokenize(line) {
		seen[t] = true
	}
}
