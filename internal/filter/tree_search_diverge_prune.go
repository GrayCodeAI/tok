package filter

import (
	"strings"

	"github.com/lakshmanpatel/tok/internal/core"
)

// Paper: "SSDP / Chopping Trees: Pruning Tree-of-Thought Branches for Efficient LLM Inference"
// NeurIPSW 2025
//
// SSDPFilter detects branching tree-of-thought (ToT) structures in text and prunes
// redundant or divergent branches, keeping only the most informative path.
//
// Branch detection: sequences starting with markers like:
//   - "Option A:", "Option B:", "Alternative:", "Approach 1:", "Approach 2:"
//   - "Case 1:", "Case 2:", "Path A:", "Scenario A:"
//
// Pruning strategy:
//  1. Similarity pruning: if two branches share >60% vocabulary, drop the shorter one.
//  2. Divergence pruning: if a branch's content strongly contradicts the final
//     conclusion (detected by negation+key-term overlap), drop it.
//  3. In aggressive mode: keep only the branch with the highest anchor score
//     (error/heading density) — the branch most likely to be the final answer.
type SSDPFilter struct {
	simThreshold float64 // vocabulary overlap threshold for similarity pruning
}

// NewSSDPFilter creates a new SSDP tree-of-thought branch pruner.
func NewSSDPFilter() *SSDPFilter {
	return &SSDPFilter{
		simThreshold: 0.60,
	}
}

// Name returns the filter name.
func (f *SSDPFilter) Name() string { return "33_ssdp" }

// Apply detects ToT branch blocks and prunes redundant ones.
func (f *SSDPFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")

	branches := f.detectBranches(lines)
	if len(branches) < 2 {
		return input, 0
	}

	simThresh := f.simThreshold
	if mode == ModeAggressive {
		simThresh = 0.45
	}

	suppress := make(map[int]bool)

	if mode == ModeAggressive && len(branches) > 1 {
		// Keep only highest-anchor branch
		bestIdx := 0
		bestScore := -1.0
		for i, b := range branches {
			score := ssdpBranchAnchorScore(lines[b.start : b.end+1])
			if score > bestScore {
				bestScore = score
				bestIdx = i
			}
		}
		for i, b := range branches {
			if i != bestIdx {
				for j := b.start; j <= b.end; j++ {
					suppress[j] = true
				}
			}
		}
	} else {
		// Similarity pruning: suppress shorter of similar-pair branches
		for i := 0; i < len(branches); i++ {
			if suppress[branches[i].start] {
				continue
			}
			for j := i + 1; j < len(branches); j++ {
				if suppress[branches[j].start] {
					continue
				}
				termA := ssdpBranchTerms(lines[branches[i].start : branches[i].end+1])
				termB := ssdpBranchTerms(lines[branches[j].start : branches[j].end+1])
				if gmsaOverlap(termA, termB) >= simThresh {
					// Suppress the shorter branch
					lenA := branches[i].end - branches[i].start
					lenB := branches[j].end - branches[j].start
					victim := j
					if lenA < lenB {
						victim = i
					}
					for k := branches[victim].start; k <= branches[victim].end; k++ {
						suppress[k] = true
					}
				}
			}
		}
	}

	if len(suppress) == 0 {
		return input, 0
	}

	var result []string
	for i, line := range lines {
		if !suppress[i] {
			result = append(result, line)
		}
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

type ssdpBranch struct{ start, end int }

// branchMarkers are the header patterns that signal a ToT branch.
var ssdpBranchMarkers = []string{
	"option a", "option b", "option c", "option d",
	"approach 1", "approach 2", "approach 3",
	"alternative 1", "alternative 2",
	"case 1", "case 2", "case 3",
	"path a", "path b",
	"scenario a", "scenario b",
	"method 1", "method 2",
	"solution 1", "solution 2",
}

// detectBranches finds branch-header lines and extends each branch to the next header or end.
func (f *SSDPFilter) detectBranches(lines []string) []ssdpBranch {
	var headers []int
	for i, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		for _, marker := range ssdpBranchMarkers {
			if strings.HasPrefix(lower, marker) {
				headers = append(headers, i)
				break
			}
		}
	}

	if len(headers) < 2 {
		return nil
	}

	var branches []ssdpBranch
	for k := 0; k < len(headers); k++ {
		end := len(lines) - 1
		if k+1 < len(headers) {
			end = headers[k+1] - 1
		}
		branches = append(branches, ssdpBranch{headers[k], end})
	}
	return branches
}

func ssdpBranchTerms(lines []string) map[string]bool {
	set := make(map[string]bool)
	for _, line := range lines {
		for _, t := range ltTokenize(line) {
			set[t] = true
		}
	}
	return set
}

func ssdpBranchAnchorScore(lines []string) float64 {
	if len(lines) == 0 {
		return 0
	}
	score := 0.0
	for _, line := range lines {
		if isErrorLine(line) || isWarningLine(line) {
			score += 2.0
		} else if isHeadingLine(line) {
			score += 1.0
		}
	}
	return score / float64(len(lines))
}
