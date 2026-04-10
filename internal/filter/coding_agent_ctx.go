package filter

import (
	"math"
	"sort"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Paper: "SWE-Pruner: Self-Adaptive Context Pruning for Coding Agents"
// arXiv:2601.16746 — Wang et al., Shanghai Jiao Tong, 2026
//
// CodingAgentContextFilter specialises context pruning for the structured
// tool outputs that coding agents (Claude Code, Cursor, etc.) receive:
// file reads, bash output, search results, git diffs, test output, compile logs.
//
// Unlike the general GoalDrivenFilter (CRF scoring against query terms),
// this filter is structure-aware: it identifies the output type and applies
// a type-specific compression strategy, then self-adjusts the compression
// ratio based on observed output density.
//
// Type-specific strategies:
//
//	file_read   — elide unchanged middle sections, keep head+tail
//	bash_output — keep last N lines (most recent = most relevant)
//	search_hits — one result per unique file path
//	git_diff    — keep ±-lines only, drop context lines in aggressive mode
//	test_output — keep FAIL/PASS summary + failing assertions
//	compile_log — keep error/warning lines, collapse repeated warnings
type CodingAgentContextFilter struct {
	headLines  int     // lines to keep at head of file reads
	tailLines  int     // lines to keep at tail of bash output
	maxResults int     // max search results per file path
	baseRatio  float64 // baseline keep ratio for unlabeled output
}

// NewCodingAgentContextFilter creates a self-adaptive coding agent context filter.
func NewCodingAgentContextFilter() *CodingAgentContextFilter {
	return &CodingAgentContextFilter{
		headLines:  30,
		tailLines:  50,
		maxResults: 3,
		baseRatio:  0.6,
	}
}

// Name returns the filter name.
func (f *CodingAgentContextFilter) Name() string { return "24_coding_agent_ctx" }

// Apply detects output type and applies the appropriate compression strategy.
func (f *CodingAgentContextFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")
	if len(lines) < 8 {
		return input, 0
	}

	outputType := f.detectType(lines)

	var output string
	switch outputType {
	case "git_diff":
		output = f.compressDiff(lines, mode)
	case "test_output":
		output = f.compressTestOutput(lines, mode)
	case "compile_log":
		output = f.compressCompileLog(lines, mode)
	case "bash_output":
		output = f.compressBashOutput(lines, mode)
	case "search_hits":
		output = f.compressSearchHits(lines, mode)
	case "file_read":
		output = f.compressFileRead(lines, mode)
	default:
		output = f.compressGeneric(lines, mode)
	}

	if output == "" || output == input {
		return input, 0
	}

	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

// detectType identifies the output type from leading lines.
func (f *CodingAgentContextFilter) detectType(lines []string) string {
	head := strings.Join(firstN(lines, 10), "\n")
	headL := strings.ToLower(head)

	if strings.Contains(head, "diff --git") || strings.HasPrefix(lines[0], "--- ") {
		return "git_diff"
	}
	if strings.Contains(headL, "=== run") || strings.Contains(headL, "--- fail") ||
		strings.Contains(headL, "--- pass") || strings.Contains(headL, "test session starts") ||
		strings.Contains(headL, "running ") && strings.Contains(headL, "test") {
		return "test_output"
	}
	if strings.Contains(headL, "error[e") || strings.Contains(headL, "compiling ") ||
		strings.Contains(headL, "building [") || strings.Contains(headL, ": error:") {
		return "compile_log"
	}
	// Search results: lines starting with "filepath:linenum:content"
	if looksLikeSearchResults(lines) {
		return "search_hits"
	}
	// File read: lots of indented/code lines, no diff markers
	if isLikelyFileRead(lines) {
		return "file_read"
	}
	// Long bash output: no special structure
	if len(lines) > 30 {
		return "bash_output"
	}
	return "generic"
}

// compressDiff keeps +/- lines; drops @@ context lines in aggressive mode.
func (f *CodingAgentContextFilter) compressDiff(lines []string, mode Mode) string {
	var result []string
	for _, line := range lines {
		if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "index ") ||
			strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") {
			result = append(result, line)
			continue
		}
		if strings.HasPrefix(line, "@@") {
			if mode != ModeAggressive {
				result = append(result, line)
			}
			continue
		}
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			result = append(result, line)
			continue
		}
		// Context line
		if mode != ModeAggressive {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// compressTestOutput keeps FAIL summary + failing test names + assertion lines.
func (f *CodingAgentContextFilter) compressTestOutput(lines []string, mode Mode) string {
	var result []string
	inFailBlock := false

	for _, line := range lines {
		lower := strings.ToLower(line)
		isSummary := strings.Contains(lower, "passed") || strings.Contains(lower, "failed") ||
			strings.Contains(lower, "ok") && strings.Contains(lower, "test") ||
			strings.HasPrefix(lower, "failures:") || strings.HasPrefix(lower, "test result")
		isFailLine := strings.Contains(lower, "fail") || strings.Contains(lower, "panic") ||
			strings.Contains(lower, "assert") || strings.Contains(lower, "expected") ||
			strings.Contains(lower, "got ") || strings.HasPrefix(line, "---") ||
			strings.HasPrefix(line, "FAIL") || strings.HasPrefix(line, "--- FAIL")

		if isSummary {
			result = append(result, line)
			inFailBlock = false
			continue
		}
		if isFailLine {
			result = append(result, line)
			inFailBlock = true
			continue
		}
		if inFailBlock && mode != ModeAggressive {
			result = append(result, line) // context within fail block
		}
	}
	return strings.Join(result, "\n")
}

// compressCompileLog keeps error/warning lines; collapses repeated warnings.
func (f *CodingAgentContextFilter) compressCompileLog(lines []string, mode Mode) string {
	var result []string
	seenWarnings := make(map[string]int) // pattern → count

	for _, line := range lines {
		lower := strings.ToLower(line)
		isErr := strings.Contains(lower, "error") || strings.Contains(lower, "fatal")
		isWarn := strings.Contains(lower, "warning") || strings.Contains(lower, "warn:")
		isNote := strings.Contains(lower, "note:") || strings.Contains(lower, "help:")
		isSummary := strings.HasPrefix(lower, "error[") || strings.Contains(lower, "aborting due") ||
			strings.Contains(lower, "build failed") || strings.Contains(lower, "build finished")

		if isErr || isSummary {
			result = append(result, line)
			continue
		}
		if isWarn {
			// Normalise warning to its pattern (strip line numbers)
			pattern := warnPattern(line)
			seenWarnings[pattern]++
			if seenWarnings[pattern] == 1 {
				result = append(result, line)
			} else if seenWarnings[pattern] == 2 && mode != ModeAggressive {
				result = append(result, line+" [repeated]")
			}
			continue
		}
		if isNote && mode != ModeAggressive {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// compressBashOutput keeps last tailLines lines (most recent = most relevant).
func (f *CodingAgentContextFilter) compressBashOutput(lines []string, mode Mode) string {
	keep := f.tailLines
	if mode == ModeAggressive {
		keep = keep / 2
	}
	if len(lines) <= keep {
		return strings.Join(lines, "\n")
	}
	omitted := len(lines) - keep
	stub := "[... " + itoa(omitted) + " lines omitted]\n"
	return stub + strings.Join(lines[len(lines)-keep:], "\n")
}

// compressSearchHits keeps at most maxResults matches per file path.
func (f *CodingAgentContextFilter) compressSearchHits(lines []string, mode Mode) string {
	maxPer := f.maxResults
	if mode == ModeAggressive {
		maxPer = 1
	}
	fileCounts := make(map[string]int)
	var result []string
	for _, line := range lines {
		path := extractFilePath(line)
		if path == "" {
			result = append(result, line)
			continue
		}
		fileCounts[path]++
		if fileCounts[path] <= maxPer {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// compressFileRead keeps head + tail, elides middle with a stub.
func (f *CodingAgentContextFilter) compressFileRead(lines []string, mode Mode) string {
	head := f.headLines
	tail := f.headLines
	if mode == ModeAggressive {
		head = head / 2
		tail = tail / 2
	}
	total := head + tail
	if len(lines) <= total {
		return strings.Join(lines, "\n")
	}
	omitted := len(lines) - total
	stub := "... (" + itoa(omitted) + " lines omitted) ..."
	result := append(lines[:head], stub)
	result = append(result, lines[len(lines)-tail:]...)
	return strings.Join(result, "\n")
}

// compressGeneric applies a simple keep-ratio to unlabeled output.
func (f *CodingAgentContextFilter) compressGeneric(lines []string, mode Mode) string {
	ratio := f.baseRatio
	if mode == ModeAggressive {
		ratio *= 0.6
	}
	keep := int(math.Ceil(float64(len(lines)) * ratio))
	if keep >= len(lines) {
		return strings.Join(lines, "\n")
	}
	// Keep structurally important lines first, then fill to budget
	type scored struct {
		idx   int
		score float64
	}
	scores := make([]scored, len(lines))
	for i, line := range lines {
		scores[i] = scored{idx: i, score: structuralBonus(line)}
	}
	sort.Slice(scores, func(a, b int) bool { return scores[a].score > scores[b].score })

	kept := make(map[int]bool)
	for _, s := range scores[:keep] {
		kept[s.idx] = true
	}
	var result []string
	for i, line := range lines {
		if kept[i] {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// -- helpers --

func firstN(lines []string, n int) []string {
	if n > len(lines) {
		n = len(lines)
	}
	return lines[:n]
}

func looksLikeSearchResults(lines []string) bool {
	matches := 0
	for _, line := range firstN(lines, 15) {
		if colonIdx := strings.Index(line, ":"); colonIdx > 0 {
			prefix := line[:colonIdx]
			if strings.Contains(prefix, "/") || strings.HasSuffix(prefix, ".go") ||
				strings.HasSuffix(prefix, ".rs") || strings.HasSuffix(prefix, ".ts") {
				matches++
			}
		}
	}
	return matches >= 3
}

func isLikelyFileRead(lines []string) bool {
	indented := 0
	for _, line := range firstN(lines, 20) {
		if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "    ") {
			indented++
		}
	}
	return indented >= 8
}

func warnPattern(line string) string {
	// Strip digits to normalize warning patterns
	var b strings.Builder
	for _, ch := range line {
		if ch >= '0' && ch <= '9' {
			b.WriteRune('#')
		} else {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func extractFilePath(line string) string {
	idx := strings.Index(line, ":")
	if idx <= 0 {
		return ""
	}
	prefix := line[:idx]
	if strings.Contains(prefix, "/") || strings.Contains(prefix, ".") {
		return prefix
	}
	return ""
}
