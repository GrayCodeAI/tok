package filter

import (
	"sort"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// AgentOCRHistoryFilter compacts older conversation turns while preserving recent turns.
type AgentOCRHistoryFilter struct {
	recentTurns int
}

// NewAgentOCRHistoryFilter creates the history-focused AgentOCR extension.
func NewAgentOCRHistoryFilter() *AgentOCRHistoryFilter {
	return &AgentOCRHistoryFilter{recentTurns: 3}
}

// Name returns the filter name.
func (f *AgentOCRHistoryFilter) Name() string { return "41_agent_ocr_history" }

// Apply compresses old low-density turns and preserves recent high-signal turns.
func (f *AgentOCRHistoryFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}
	lines := strings.Split(input, "\n")
	turns := parseRoleTurns(lines)
	if len(turns) < 4 {
		return input, 0
	}

	keepRecent := f.recentTurns
	if mode == ModeAggressive {
		keepRecent = 2
	}
	cut := len(turns) - keepRecent
	if cut < 1 {
		cut = 1
	}

	out := make([]string, 0, len(lines))
	changed := false
	for i, t := range turns {
		seg := lines[t.start : t.end+1]
		if i >= cut {
			out = append(out, seg...)
			continue
		}

		out = append(out, lines[t.start])
		kept := ocrHistoryTopLines(seg[1:], mode)
		if len(kept) > 0 {
			out = append(out, kept...)
		}
		omitted := len(seg) - 1 - len(kept)
		if omitted > 0 {
			out = append(out, "[agent-ocr-history: "+itoa(omitted)+" lines compacted]")
			changed = true
		}
	}

	if !changed {
		return input, 0
	}
	output := strings.Join(out, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func ocrHistoryTopLines(lines []string, mode Mode) []string {
	if len(lines) == 0 {
		return nil
	}
	type cand struct {
		idx   int
		score float64
	}
	cands := make([]cand, 0, len(lines))
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		score := 0.0
		if isErrorLine(line) || isWarningLine(line) || isCodeLine(line) {
			score += 2.0
		}
		if strings.ContainsAny(line, ":=/") {
			score += 0.7
		}
		score += float64(len(ltTokenize(line))) / 10.0
		cands = append(cands, cand{idx: i, score: score})
	}

	if len(cands) == 0 {
		return nil
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].score > cands[j].score })

	limit := 2
	if mode == ModeAggressive {
		limit = 1
	}
	if limit > len(cands) {
		limit = len(cands)
	}

	pick := make(map[int]bool, limit)
	for i := 0; i < limit; i++ {
		pick[cands[i].idx] = true
	}
	out := make([]string, 0, limit)
	for i, line := range lines {
		if pick[i] {
			out = append(out, line)
		}
	}
	return out
}
