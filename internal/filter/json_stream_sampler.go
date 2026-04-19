package filter

import (
	"strings"

	"github.com/lakshmanpatel/tok/internal/core"
)

// JSONSamplerFilter down-samples dense JSON line streams while preserving anchors.
type JSONSamplerFilter struct{}

func NewJSONSamplerFilter() *JSONSamplerFilter { return &JSONSamplerFilter{} }

func (f *JSONSamplerFilter) Name() string { return "45_json_sampler" }

func (f *JSONSamplerFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}
	lines := strings.Split(input, "\n")
	if len(lines) < 20 {
		return input, 0
	}

	jsonLike := 0
	for _, line := range lines {
		if isJSONLikeLine(line) {
			jsonLike++
		}
	}
	if float64(jsonLike)/float64(len(lines)) < 0.55 {
		return input, 0
	}

	stride := 4
	if mode == ModeAggressive {
		stride = 6
	}

	out := make([]string, 0, len(lines)/2)
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if i < 4 || i >= len(lines)-4 || isErrorLine(line) || isWarningLine(line) {
			out = append(out, line)
			continue
		}
		if !isJSONLikeLine(line) {
			out = append(out, line)
			continue
		}
		if i%stride == 0 || strings.Contains(trim, "\"error\"") {
			out = append(out, line)
		}
	}

	if len(out) >= len(lines) {
		return input, 0
	}
	out = append(out, "[json-sampler: sampled JSON lines]")
	output := strings.Join(out, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func isJSONLikeLine(line string) bool {
	trim := strings.TrimSpace(line)
	if trim == "" {
		return false
	}
	if strings.HasPrefix(trim, "{") || strings.HasPrefix(trim, "[") || strings.HasPrefix(trim, "}") || strings.HasPrefix(trim, "]") {
		return true
	}
	if strings.Contains(trim, "\":") || strings.Contains(trim, "\",") {
		return true
	}
	return false
}
