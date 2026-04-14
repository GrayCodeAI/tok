package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

func (p *PipelineCoordinator) applyAdaptiveRouting(input string, stats *PipelineStats) string {
	output := input
	if p.config.EnablePolicyRouter {
		if p.runtimeQueryIntent == "" {
			p.runtimeQueryIntent = inferQueryIntentFromContent(output)
		}
		stats.LayerStats["0_policy_router"] = LayerStat{TokensSaved: 0}
	}

	if p.config.EnableExtractivePrefilter {
		filtered, saved := p.applyExtractivePrefilter(output)
		if saved > 0 {
			output = filtered
			stats.LayerStats["0_extractive_prefilter"] = LayerStat{TokensSaved: saved}
			stats.runningSaved += saved
		}
	}
	return output
}

func inferQueryIntentFromContent(input string) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "panic") || strings.Contains(lower, "error") || strings.Contains(lower, "failed"):
		return "debug"
	case strings.Contains(lower, "diff --git") || strings.Contains(lower, "@@") || strings.Contains(lower, "patch"):
		return "review"
	case strings.Contains(lower, "test") || strings.Contains(lower, "assert"):
		return "test"
	case strings.Contains(lower, "benchmark") || strings.Contains(lower, "latency"):
		return "optimize"
	default:
		return "summarize"
	}
}

func (p *PipelineCoordinator) applyExtractivePrefilter(input string) (string, int) {
	lines := strings.Split(input, "\n")
	maxLines := p.config.ExtractiveMaxLines
	if maxLines <= 0 {
		maxLines = 400
	}
	if len(lines) <= maxLines {
		return input, 0
	}

	head := p.config.ExtractiveHeadLines
	tail := p.config.ExtractiveTailLines
	signalBudget := p.config.ExtractiveSignalLines
	if head <= 0 {
		head = 80
	}
	if tail <= 0 {
		tail = 60
	}
	if signalBudget <= 0 {
		signalBudget = 120
	}

	keep := make(map[int]bool, head+tail+signalBudget)
	for i := 0; i < len(lines) && i < head; i++ {
		keep[i] = true
	}
	for i := len(lines) - tail; i < len(lines); i++ {
		if i >= 0 {
			keep[i] = true
		}
	}

	intent := p.effectiveQueryIntent()
	signals := 0
	for i, line := range lines {
		if signals >= signalBudget {
			break
		}
		if keep[i] {
			continue
		}
		l := strings.ToLower(line)
		if isErrorLine(line) || isWarningLine(line) || isCodeLine(line) || isReasoningLine(line) || epicIsCausalEdge(line) {
			keep[i] = true
			signals++
			continue
		}
		if intent != "" && strings.Contains(l, intent) {
			keep[i] = true
			signals++
		}
	}

	out := make([]string, 0, len(keep)+2)
	for i, line := range lines {
		if keep[i] {
			out = append(out, line)
		}
	}
	out = append(out, "[extractive-prefilter: reduced from "+itoa(len(lines))+" lines]")
	output := strings.Join(out, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}
