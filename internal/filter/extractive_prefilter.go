package filter

import (
	"sort"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

type ExtractivePrefilterConfig struct {
	MaxLines    int
	HeadLines   int
	TailLines   int
	SignalLines int
}

type ExtractivePrefilter struct {
	cfg ExtractivePrefilterConfig
}

func NewExtractivePrefilter(cfg ExtractivePrefilterConfig) *ExtractivePrefilter {
	if cfg.MaxLines <= 0 {
		cfg.MaxLines = 400
	}
	if cfg.HeadLines <= 0 {
		cfg.HeadLines = 80
	}
	if cfg.TailLines <= 0 {
		cfg.TailLines = 60
	}
	if cfg.SignalLines <= 0 {
		cfg.SignalLines = 120
	}
	return &ExtractivePrefilter{cfg: cfg}
}

func (e *ExtractivePrefilter) Apply(input string) (string, int) {
	lines := strings.Split(input, "\n")
	if len(lines) <= e.cfg.MaxLines {
		return input, 0
	}

	keep := make(map[int]struct{}, e.cfg.HeadLines+e.cfg.TailLines+e.cfg.SignalLines)

	for i := 0; i < e.cfg.HeadLines && i < len(lines); i++ {
		keep[i] = struct{}{}
	}
	startTail := len(lines) - e.cfg.TailLines
	if startTail < 0 {
		startTail = 0
	}
	for i := startTail; i < len(lines); i++ {
		keep[i] = struct{}{}
	}

	signalBudget := e.cfg.SignalLines
	for i, line := range lines {
		if signalBudget <= 0 {
			break
		}
		if isSignalLine(line) {
			if _, exists := keep[i]; !exists {
				keep[i] = struct{}{}
				signalBudget--
			}
		}
	}

	indexes := make([]int, 0, len(keep))
	for i := range keep {
		indexes = append(indexes, i)
	}
	sort.Ints(indexes)

	outLines := make([]string, 0, len(indexes)+8)
	last := -2
	for _, i := range indexes {
		if i-last > 1 {
			outLines = append(outLines, "[... omitted by extractive prefilter ...]")
		}
		outLines = append(outLines, lines[i])
		last = i
	}

	output := strings.Join(outLines, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func isSignalLine(line string) bool {
	l := strings.ToLower(line)
	return containsAny(l,
		"error", "failed", "panic", "exception", "traceback", "warning",
		"fatal", "undefined", "cannot", "permission denied", "timeout",
		"diff --git", "@@", "assert", "expected", "actual",
	)
}
