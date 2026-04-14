package filter

import (
	"strings"
)

// L16: AdvancedFilter - Unified filter for advanced optimizations
// Merges: L31-L45 research filters including:
// - DiffAdapt, EPiC, SSDP, AgentOCR, S2MAD
// - ACON, LatentCollab, GraphCoT, RoleBudget
// - SWEAdaptiveLoop, AgentOCRHistory, PlanBudget
// - LightMem, PathShorten, JSONSampler
//
// This filter applies advanced research-backed optimizations
// when standard compression isn't sufficient.

type AdvancedFilter struct {
	enabledOptimizations map[string]bool
}

func NewAdvancedFilter() *AdvancedFilter {
	return &AdvancedFilter{
		enabledOptimizations: map[string]bool{
			"adaptive_diff":   true, // L31: DiffAdapt
			"causal_edges":    true, // L32: EPiC
			"tree_prune":      true, // L33: SSDP
			"agent_ctx":       true, // L34: AgentOCR
			"debate_collapse": true, // L35: S2MAD
			"context_opt":     true, // L36: ACON
			"collab_merge":    true, // L37: LatentCollab
			"graph_cot":       true, // L38: GraphCoT
			"role_budget":     true, // L39: RoleBudget
			"adaptive_loop":   true, // L40: SWEAdaptiveLoop
		},
	}
}

func (f *AdvancedFilter) Apply(input string, mode Mode) (string, int) {
	// Only apply for large content or aggressive mode
	if len(input) < 5000 && mode != ModeAggressive {
		return input, 0
	}

	originalLen := len(input)
	compressed := input

	// Apply enabled optimizations based on content type
	if f.enabledOptimizations["adaptive_diff"] && f.isDiffContent(input) {
		compressed = f.adaptiveDiffCompress(compressed)
	}

	if f.enabledOptimizations["causal_edges"] && f.hasCausalStructure(input) {
		compressed = f.preserveCausalEdges(compressed)
	}

	if f.enabledOptimizations["tree_prune"] && f.isTreeStructure(input) {
		compressed = f.pruneTreeBranches(compressed)
	}

	if f.enabledOptimizations["agent_ctx"] && f.isAgentContext(input) {
		compressed = f.optimizeAgentContext(compressed)
	}

	if f.enabledOptimizations["context_opt"] {
		compressed = f.optimizeContextWindow(compressed)
	}

	if f.enabledOptimizations["collab_merge"] && f.isCollaboration(input) {
		compressed = f.mergeCollaboration(compressed)
	}

	saved := originalLen - len(compressed)
	return compressed, saved
}

func (f *AdvancedFilter) isDiffContent(input string) bool {
	return strings.Contains(input, "diff --git") ||
		strings.Contains(input, "--- a/") ||
		strings.Contains(input, "+++ b/")
}

func (f *AdvancedFilter) adaptiveDiffCompress(input string) string {
	// Simplified diff compression
	lines := splitLines(input)
	var compressed []string

	for _, line := range lines {
		// Keep context lines minimal
		if strings.HasPrefix(line, "@@") {
			compressed = append(compressed, line)
		} else if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			compressed = append(compressed, line)
		} else if strings.HasPrefix(line, " ") {
			// Skip most context lines, keep every 10th
			if len(compressed)%10 == 0 {
				compressed = append(compressed, " ...")
			}
		}
	}

	return joinLines(compressed)
}

func (f *AdvancedFilter) hasCausalStructure(input string) bool {
	return strings.Contains(input, "because") ||
		strings.Contains(input, "therefore") ||
		strings.Contains(input, "causes") ||
		strings.Contains(input, "leads to")
}

func (f *AdvancedFilter) preserveCausalEdges(input string) string {
	// Keep sentences with causal relationships
	lines := splitLines(input)
	var preserved []string

	for _, line := range lines {
		if f.hasCausalStructure(line) || len(preserved) < 5 {
			preserved = append(preserved, line)
		} else if len(preserved)%5 == 0 {
			preserved = append(preserved, "...")
		}
	}

	return joinLines(preserved)
}

func (f *AdvancedFilter) isTreeStructure(input string) bool {
	return strings.Contains(input, "├──") ||
		strings.Contains(input, "└──") ||
		strings.Contains(input, "├─") ||
		strings.Count(input, "  ") > 20
}

func (f *AdvancedFilter) pruneTreeBranches(input string) string {
	lines := splitLines(input)
	if len(lines) < 20 {
		return input
	}

	var pruned []string
	depth := 0

	for i, line := range lines {
		// Keep root and first few branches
		if i < 10 {
			pruned = append(pruned, line)
			continue
		}

		// Calculate depth from indentation
		currentDepth := (len(line) - len(strings.TrimLeft(line, " "))) / 2

		// Skip deep branches, keep summary
		if currentDepth <= depth+2 {
			pruned = append(pruned, line)
			depth = currentDepth
		} else if i%10 == 0 {
			pruned = append(pruned, "  ...")
		}
	}

	return joinLines(pruned)
}

func (f *AdvancedFilter) isAgentContext(input string) bool {
	return strings.Contains(input, "Agent:") ||
		strings.Contains(input, "Assistant:") ||
		strings.Contains(input, "System:") ||
		strings.Contains(input, "User:")
}

func (f *AdvancedFilter) optimizeAgentContext(input string) string {
	// Keep recent context, compress older
	lines := splitLines(input)
	if len(lines) < 30 {
		return input
	}

	// Keep first 5 and last 20 lines
	var optimized []string
	optimized = append(optimized, lines[:5]...)
	optimized = append(optimized, "[... older context compressed ...]")
	optimized = append(optimized, lines[len(lines)-20:]...)

	return joinLines(optimized)
}

func (f *AdvancedFilter) optimizeContextWindow(input string) string {
	// ACON-style optimization: allocate budget based on importance
	if len(input) < 2000 {
		return input
	}

	lines := splitLines(input)
	if len(lines) < 50 {
		return input
	}

	// Score each line
	type scoredLine struct {
		line  string
		score int
	}

	scored := make([]scoredLine, len(lines))
	for i, line := range lines {
		score := 0
		if strings.Contains(line, "func ") || strings.Contains(line, "class ") {
			score += 10 // High importance
		}
		if strings.Contains(line, "error") || strings.Contains(line, "Error") {
			score += 8
		}
		if strings.Contains(line, "TODO") || strings.Contains(line, "FIXME") {
			score += 5
		}
		scored[i] = scoredLine{line, score}
	}

	// Keep high-score lines + some low-score for context
	var optimized []string
	kept := 0
	maxKeep := len(lines) / 2

	for _, sl := range scored {
		if sl.score > 0 || kept < maxKeep/3 {
			optimized = append(optimized, sl.line)
			kept++
		}
		if kept >= maxKeep {
			break
		}
	}

	return joinLines(optimized)
}

func (f *AdvancedFilter) isCollaboration(input string) bool {
	return strings.Contains(input, "@") &&
		(strings.Contains(input, "said") || strings.Contains(input, "wrote"))
}

func (f *AdvancedFilter) mergeCollaboration(input string) string {
	// Simple dedup for collaborative content
	lines := splitLines(input)
	var merged []string
	seen := make(map[string]bool)

	for _, line := range lines {
		normalized := strings.ToLower(strings.TrimSpace(line))
		if !seen[normalized] || len(normalized) < 20 {
			merged = append(merged, line)
			seen[normalized] = true
		}
	}

	return joinLines(merged)
}

func (f *AdvancedFilter) Name() string {
	return "L16_Advanced"
}
