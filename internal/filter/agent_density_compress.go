package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Paper: "AgentOCR: Content-Density Aware Compression for Multi-Turn Agent Trajectories"
// arXiv 2026
//
// AgentOCRFilter operates on multi-turn agent outputs (tool call sequences, conversation turns).
// It measures the "content density" of each turn — the ratio of information-bearing lines
// to total lines — and collapses low-density turns into a single summary stub while
// preserving high-density turns in full.
//
// Content density signals:
//   - Information-bearing: error lines, code lines, unique-term-rich lines
//   - Filler: empty lines, repeated terms, pure-whitespace lines
//
// Turn detection: turns are separated by patterns like:
//   - "Human:", "Assistant:", "User:", "Agent:", "System:"
//   - "<human>", "<assistant>", markdown "## Turn N"
//
// Density thresholds:
//   - Low density (<0.30): collapse to "[Turn summary: N lines, M tokens]"
//   - Medium density (0.30–0.65): keep first+last few lines + omission marker
//   - High density (>0.65): preserve fully
type AgentOCRFilter struct {
	lowDensityThreshold  float64
	highDensityThreshold float64
	contextLines         int // lines to keep at start/end of medium-density turns
}

// NewAgentOCRFilter creates a new agent turn content-density filter.
func NewAgentOCRFilter() *AgentOCRFilter {
	return &AgentOCRFilter{
		lowDensityThreshold:  0.30,
		highDensityThreshold: 0.65,
		contextLines:         3,
	}
}

// Name returns the filter name.
func (f *AgentOCRFilter) Name() string { return "34_agent_ocr" }

// Apply collapses low-density agent turns, trims medium-density ones.
func (f *AgentOCRFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")

	turns := f.parseTurns(lines)
	if len(turns) < 2 {
		return input, 0
	}

	lowThresh := f.lowDensityThreshold
	highThresh := f.highDensityThreshold
	if mode == ModeAggressive {
		lowThresh = 0.45
		highThresh = 0.75
	}

	var resultLines []string
	changed := false

	for _, t := range turns {
		turnLines := lines[t.start : t.end+1]
		density := f.contentDensity(turnLines)

		if density >= highThresh {
			// High density: preserve fully
			resultLines = append(resultLines, turnLines...)
		} else if density >= lowThresh {
			// Medium density: keep head + tail + marker
			ctx := f.contextLines
			if len(turnLines) <= ctx*2+1 {
				resultLines = append(resultLines, turnLines...)
			} else {
				resultLines = append(resultLines, turnLines[:ctx]...)
				omitted := len(turnLines) - ctx*2
				resultLines = append(resultLines, "[... "+itoa(omitted)+" lines omitted (density="+aocFmtPct(density)+") ...]")
				resultLines = append(resultLines, turnLines[len(turnLines)-ctx:]...)
				changed = true
			}
		} else {
			// Low density: collapse to stub
			tokens := core.EstimateTokens(strings.Join(turnLines, "\n"))
			resultLines = append(resultLines, turnLines[0]) // keep the header line
			resultLines = append(resultLines, "[collapsed: "+itoa(len(turnLines)-1)+" lines / ~"+itoa(tokens)+" tokens (low density)]")
			changed = true
		}
	}

	if !changed {
		return input, 0
	}

	output := strings.Join(resultLines, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

type agentTurn struct{ start, end int }

var aocTurnHeaders = []string{
	"human:", "assistant:", "user:", "agent:", "system:",
	"<human>", "<assistant>", "<user>", "<agent>",
	"## turn ", "# turn ",
}

// parseTurns splits lines into agent turns by header patterns.
func (f *AgentOCRFilter) parseTurns(lines []string) []agentTurn {
	var headers []int
	for i, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		for _, h := range aocTurnHeaders {
			if strings.HasPrefix(lower, h) {
				headers = append(headers, i)
				break
			}
		}
	}

	if len(headers) < 2 {
		return nil
	}

	var turns []agentTurn
	for k := 0; k < len(headers); k++ {
		end := len(lines) - 1
		if k+1 < len(headers) {
			end = headers[k+1] - 1
		}
		turns = append(turns, agentTurn{headers[k], end})
	}
	return turns
}

// contentDensity measures the fraction of information-bearing lines in a turn.
func (f *AgentOCRFilter) contentDensity(lines []string) float64 {
	if len(lines) == 0 {
		return 1.0
	}
	infoBearing := 0
	for _, line := range lines {
		if f.isInfoBearing(line) {
			infoBearing++
		}
	}
	return float64(infoBearing) / float64(len(lines))
}

// isInfoBearing returns true if a line carries substantive information.
func (f *AgentOCRFilter) isInfoBearing(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	if isErrorLine(line) || isWarningLine(line) || isCodeLine(line) {
		return true
	}
	// Lines with ≥4 distinct tokens are likely substantive
	terms := ltTokenize(line)
	unique := make(map[string]bool)
	for _, t := range terms {
		unique[t] = true
	}
	return len(unique) >= 4
}

// aocFmtPct formats a float as a short percentage string.
func aocFmtPct(f float64) string {
	pct := int(f * 100)
	return itoa(pct) + "%"
}
