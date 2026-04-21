package filter

import (
	"strings"

	"github.com/GrayCodeAI/tok/internal/core"
)

// Paper: "CARL: Critical Action Focused Reinforcement Learning for Multi-Step Agent"
// arXiv:2512.04949 — 2025
//
// CARLFilter identifies "critical" vs "non-critical" tool-call entries in
// agent output sequences and drops the non-critical ones.
//
// Criticality is defined as: did this action cause an observable state change?
//
//	Critical:     error output, file writes/deletes, test failures, non-empty diffs,
//	              non-zero exit codes, assertion failures, CRUD operations
//	Non-critical: empty results, successful no-ops, pure info queries,
//	              repeated identical results, health checks
//
// CARL's key insight (from RL perspective): in a long agent trajectory, most
// actions are "maintenance" (checking state, listing files, echoing info) and
// carry no new causal information. Keeping only critical actions and their
// immediate context preserves the trajectory's causal skeleton at a fraction
// of the token cost.
type CARLFilter struct {
	criticalPatterns    []string
	nonCriticalPatterns []string
	contextLines        int // lines of context to keep around critical entries
	entryHeaderRe       []string
}

// NewCARLFilter creates a new CARL critical-action filter.
func NewCARLFilter() *CARLFilter {
	return &CARLFilter{
		criticalPatterns: []string{
			"error", "fail", "failed", "exception", "panic", "fatal",
			"assert", "expected", "got ", "mismatch", "undefined",
			"permission denied", "no such file", "not found",
			"exit code", "exit status", "returncode",
			"created", "deleted", "removed", "written", "saved",
			"diff --git", "--- a/", "+++ b/", "@@ -", "@@ +",
			"test failed", "assertion failed",
		},
		nonCriticalPatterns: []string{
			"(no output)", "(empty)", "total 0",
			"nothing to commit", "up to date",
			"ok\n", "ok  \t",
			"200 ok", "status: ok", "health: ok",
			"already exists", "already up to date",
		},
		contextLines: 2,
		entryHeaderRe: []string{
			"tool:", "result:", "output:", "stdout:", "stderr:",
			"<tool_result>", "<result>",
		},
	}
}

// Name returns the filter name.
func (f *CARLFilter) Name() string { return "29_carl" }

// Apply drops non-critical agent tool-call entries.
func (f *CARLFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")

	// If not agent-like output, skip
	if !f.looksLikeAgentOutput(lines) {
		return input, 0
	}

	entries := f.parseEntries(lines)
	if len(entries) < 2 {
		return input, 0
	}

	critThreshold := 0.3
	if mode == ModeAggressive {
		critThreshold = 0.5
	}

	suppress := make(map[int]bool)
	for _, e := range entries {
		score := f.criticalityScore(lines[e.start : e.end+1])
		if score < critThreshold {
			for i := e.start; i <= e.end; i++ {
				suppress[i] = true
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

type agentEntry struct{ start, end int }

// parseEntries segments the output into tool-call result blocks.
func (f *CARLFilter) parseEntries(lines []string) []agentEntry {
	var entries []agentEntry
	inEntry := false
	start := 0

	for i, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		isHeader := false
		for _, h := range f.entryHeaderRe {
			if strings.HasPrefix(lower, h) {
				isHeader = true
				break
			}
		}
		if isHeader {
			if inEntry && i > start {
				entries = append(entries, agentEntry{start, i - 1})
			}
			start = i
			inEntry = true
		}
	}
	if inEntry {
		entries = append(entries, agentEntry{start, len(lines) - 1})
	}
	return entries
}

// criticalityScore returns 0.0..1.0 for how critical an entry is.
func (f *CARLFilter) criticalityScore(lines []string) float64 {
	if len(lines) == 0 {
		return 0
	}

	text := strings.ToLower(strings.Join(lines, "\n"))

	// Non-critical patterns immediately signal low criticality
	for _, p := range f.nonCriticalPatterns {
		if strings.Contains(text, p) {
			return 0.05
		}
	}

	// Empty-ish entries
	contentLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			contentLines++
		}
	}
	if contentLines <= 1 {
		return 0.1
	}

	// Score by critical pattern hits
	hits := 0
	for _, p := range f.criticalPatterns {
		if strings.Contains(text, p) {
			hits++
		}
	}

	score := float64(hits) / float64(len(f.criticalPatterns))
	if score > 1.0 {
		score = 1.0
	}

	// Bonus for diff markers or explicit error/fail lines
	if strings.Contains(text, "error") || strings.Contains(text, "fail") {
		score += 0.3
	}
	if strings.Contains(text, "diff --git") || strings.Contains(text, "--- a/") ||
		strings.Contains(text, "+++ b/") {
		score += 0.4
	}
	if score > 1.0 {
		score = 1.0
	}
	return score
}

// looksLikeAgentOutput returns true if the input seems to contain agent tool results.
func (f *CARLFilter) looksLikeAgentOutput(lines []string) bool {
	count := 0
	for _, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		for _, h := range f.entryHeaderRe {
			if strings.HasPrefix(lower, h) {
				count++
				break
			}
		}
		if count >= 2 {
			return true
		}
	}
	return false
}
