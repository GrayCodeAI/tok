package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// plannedHeuristicFilter provides baseline implementations for planned layers 30-49.
// They are intentionally conservative and disabled by default.
type plannedHeuristicFilter struct {
	id string
}

func (f *plannedHeuristicFilter) Name() string { return f.id }

func (f *plannedHeuristicFilter) Apply(input string, mode Mode) (string, int) {
	if input == "" || mode == ModeNone {
		return input, 0
	}
	output := input

	// Alias IDs are consolidated to canonical behaviors to avoid duplicate work.
	switch plannedLayerCanonicalID(f.id) {
	case "30_salience_graph", "38_semantic_dedup":
		output = dedupLines(output)
	case "31_trace_preserve", "36_stacktrace_focus", "41_error_window":
		output = keepSignalNeighborhood(output, 2)
	case "32_ast_diff_focus":
		output = keepDiffFocused(output)
	case "33_unit_test_focus":
		output = keepMatchingAndContext(output, []string{"fail", "assert", "expected", "actual", "test"}, 1)
	case "34_symbol_table":
		output = keepMatchingAndContext(output, []string{"func ", "class ", "def ", "interface ", "type "}, 0)
	case "35_path_anchor":
		output = keepMatchingAndContext(output, []string{".go:", ".ts:", ".py:", ".rs:", ".java:", ".rb:"}, 0)
	case "37_exit_signal_keep":
		output = keepMatchingAndContext(output, []string{"exit code", "status ", "failed", "success"}, 0)
	case "39_recall_booster", "49_repair_pass":
		// Conservative no-op baseline; upgraded versions can perform loss-aware repair.
	case "40_log_cluster":
		output = dedupLines(output)
	case "42_dependency_focus":
		output = keepMatchingAndContext(output, []string{"import ", "require(", "go.mod", "package.json", "cargo.toml"}, 0)
	case "43_symbolic_patch":
		output = keepDiffFocused(output)
	case "44_runtime_anchor":
		output = keepHeadTail(output, 25, 25)
	case "45_multiturn_merge":
		output = keepMatchingAndContext(output, []string{"user:", "assistant:", "human:", "ai:"}, 1)
	case "46_context_cache":
		output = dedupLines(output)
	case "47_confidence_gate", "48_loss_guard":
		// Gate placeholders; no-op until confidence models are wired.
	}

	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

// plannedLayerCanonicalID maps overlapping planned layers to a canonical executor.
// This keeps the registry IDs for tracking while avoiding redundant runtime passes.
func plannedLayerCanonicalID(id string) string {
	switch id {
	case "36_stacktrace_focus", "41_error_window":
		return "31_trace_preserve"
	case "38_semantic_dedup", "40_log_cluster", "46_context_cache":
		return "30_salience_graph"
	case "43_symbolic_patch":
		return "32_ast_diff_focus"
	case "49_repair_pass":
		return "39_recall_booster"
	case "47_confidence_gate":
		return "48_loss_guard"
	default:
		return id
	}
}

func (p *PipelineCoordinator) initPlannedLayers() {
	if !p.config.EnablePlannedLayers {
		return
	}
	ids := []string{
		"30_salience_graph",
		"31_trace_preserve",
		"32_ast_diff_focus",
		"33_unit_test_focus",
		"34_symbol_table",
		"35_path_anchor",
		"36_stacktrace_focus",
		"37_exit_signal_keep",
		"38_semantic_dedup",
		"39_recall_booster",
		"40_log_cluster",
		"41_error_window",
		"42_dependency_focus",
		"43_symbolic_patch",
		"44_runtime_anchor",
		"45_multiturn_merge",
		"46_context_cache",
		"47_confidence_gate",
		"48_loss_guard",
		"49_repair_pass",
	}

	p.plannedLayers = make([]filterLayer, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		canonical := plannedLayerCanonicalID(id)
		if _, ok := seen[canonical]; ok {
			continue
		}
		seen[canonical] = struct{}{}
		p.plannedLayers = append(p.plannedLayers, filterLayer{
			filter: &plannedHeuristicFilter{id: canonical},
			name:   canonical,
		})
	}
}

func dedupLines(input string) string {
	lines := strings.Split(input, "\n")
	seen := make(map[string]struct{}, len(lines))
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		key := strings.TrimSpace(l)
		if key == "" {
			out = append(out, l)
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, l)
	}
	return strings.Join(out, "\n")
}

func keepDiffFocused(input string) string {
	return keepMatchingAndContext(input, []string{"diff --git", "@@ ", "--- ", "+++ ", "+", "-"}, 0)
}

func keepSignalNeighborhood(input string, radius int) string {
	return keepMatchingAndContext(input, []string{"error", "panic", "traceback", "exception", "failed", "fatal"}, radius)
}

func keepMatchingAndContext(input string, markers []string, radius int) string {
	lines := strings.Split(input, "\n")
	keep := make([]bool, len(lines))
	for i, line := range lines {
		l := strings.ToLower(line)
		matched := false
		for _, m := range markers {
			if strings.Contains(l, strings.ToLower(m)) {
				matched = true
				break
			}
		}
		if matched {
			start := i - radius
			if start < 0 {
				start = 0
			}
			end := i + radius
			if end >= len(lines) {
				end = len(lines) - 1
			}
			for j := start; j <= end; j++ {
				keep[j] = true
			}
		}
	}
	any := false
	for _, k := range keep {
		if k {
			any = true
			break
		}
	}
	if !any {
		return input
	}
	out := make([]string, 0, len(lines))
	for i, line := range lines {
		if keep[i] {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func keepHeadTail(input string, head, tail int) string {
	lines := strings.Split(input, "\n")
	if len(lines) <= head+tail {
		return input
	}
	out := make([]string, 0, head+tail+1)
	out = append(out, lines[:head]...)
	out = append(out, "[... trimmed ...]")
	out = append(out, lines[len(lines)-tail:]...)
	return strings.Join(out, "\n")
}
