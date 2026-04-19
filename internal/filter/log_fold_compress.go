package filter

import (
	"regexp"
	"strings"

	"github.com/lakshmanpatel/tok/internal/core"
)

// LogCrunchFilter folds repetitive INFO/DEBUG logs while preserving warnings/errors and state transitions.
type LogCrunchFilter struct {
	normalizeTimestamps bool
}

func NewLogCrunchFilter() *LogCrunchFilter {
	return &LogCrunchFilter{normalizeTimestamps: true}
}

func (f *LogCrunchFilter) Name() string { return "46_log_crunch" }

var (
	stackFramePattern  = regexp.MustCompile(`(?:^\s+at\s+|^\s+File\s+"|^\s+in\s+\w|Traceback|goroutine\s+\d+)`)
	stackIndentPattern = regexp.MustCompile(`^(\s{2,}|\t)`)
	timestampPattern   = regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?`)
)

func (f *LogCrunchFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}
	lines := strings.Split(input, "\n")
	if len(lines) < 20 {
		return input, 0
	}

	if f.normalizeTimestamps {
		lines = normalizeTimestamps(lines)
	}

	out := compressLogLines(lines, mode)
	if len(out) == len(lines) {
		return input, 0
	}

	out = append(out, "[log-crunch: repetitive logs folded]")
	output := strings.Join(out, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func compressLogLines(lines []string, mode Mode) []string {
	out := make([]string, 0, len(lines))
	inTrace := false
	var traceBuffer []string
	runNorm := ""
	runLines := []string{}
	runCount := 0

	flushRun := func() {
		if len(runLines) == 0 {
			return
		}
		if runCount >= 3 {
			out = append(out, runLines[0])
			out = append(out, "  [... repeated "+string(rune(runCount-2+48))+" more times ...]")
			out = append(out, runLines[len(runLines)-1])
		} else {
			out = append(out, runLines...)
		}
		runLines = nil
		runCount = 0
		runNorm = ""
	}

	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}

		// Detect stack trace start
		if !inTrace && stackFramePattern.MatchString(line) {
			flushRun()
			inTrace = true
			traceBuffer = []string{line}
			// Collect continuation lines
			for j := i + 1; j < len(lines); j++ {
				if stackIndentPattern.MatchString(lines[j]) || stackFramePattern.MatchString(lines[j]) {
					traceBuffer = append(traceBuffer, lines[j])
					i = j
				} else {
					break
				}
			}
			out = append(out, traceBuffer...)
			inTrace = false
			traceBuffer = nil
			continue
		}

		// Always keep errors/warnings
		if isErrorLine(line) || isWarningLine(line) {
			flushRun()
			out = append(out, line)
			continue
		}

		// Collapse repetitive INFO/DEBUG
		norm := normalizeLogLine(line)
		if norm == runNorm {
			runLines = append(runLines, line)
			runCount++
		} else {
			flushRun()
			runNorm = norm
			runLines = []string{line}
			runCount = 1
		}
	}

	flushRun()
	return out
}

func normalizeTimestamps(lines []string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = timestampPattern.ReplaceAllString(line, "[+T]")
	}
	return result
}

func normalizeLogLine(line string) string {
	lower := strings.ToLower(strings.TrimSpace(line))
	lower = strings.ReplaceAll(lower, "\t", " ")
	parts := strings.Fields(lower)
	if len(parts) > 8 {
		parts = parts[:8]
	}
	return strings.Join(parts, " ")
}
