package filter

import (
	"regexp"
	"strings"

	"github.com/GrayCodeAI/tok/internal/core"
)

// Paper: "TokenSkip: Controllable Chain-of-Thought Compression"
// arXiv:2502.12067 — 2025
//
// CoTCompressFilter detects chain-of-thought reasoning traces in output and
// applies token-budget-controlled compression:
//   - ModeMinimal: truncate CoT to first 30% + summary marker
//   - ModeAggressive: replace entire CoT block with a token-count stub
//
// Applicable when tok wraps tools that emit LLM reasoning output
// (e.g. claude --verbose, agent traces, reasoning model output).
//
// Patterns detected:
//   - XML-style: <think>...</think>, <reasoning>...</reasoning>
//   - Markdown step blocks: "Step 1:", "Let me think", numbered reasoning
//   - Reflection loops: "Wait,", "Actually,", "Let me reconsider"
type CoTCompressFilter struct {
	xmlThinkRe     *regexp.Regexp
	xmlReasoningRe *regexp.Regexp
	stepPrefixRe   *regexp.Regexp
	reflectionRe   *regexp.Regexp
	minBlockLines  int // minimum CoT block size before compressing
}

// NewCoTCompressFilter creates a new TokenSkip-inspired chain-of-thought compressor.
func NewCoTCompressFilter() *CoTCompressFilter {
	return &CoTCompressFilter{
		xmlThinkRe:     regexp.MustCompile(`(?s)<think>(.*?)</think>`),
		xmlReasoningRe: regexp.MustCompile(`(?s)<reasoning>(.*?)</reasoning>`),
		stepPrefixRe:   regexp.MustCompile(`(?i)^(step\s+\d+[:.)]|let me (think|consider|analyze)|firstly,|secondly,|thirdly,|finally,)`),
		reflectionRe:   regexp.MustCompile(`(?i)^(wait,|actually,|let me reconsider|hmm,|on second thought)`),
		minBlockLines:  4,
	}
}

// Name returns the filter name.
func (f *CoTCompressFilter) Name() string { return "23_cot_compress" }

// Apply compresses chain-of-thought blocks according to mode.
func (f *CoTCompressFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	output := input

	// Handle XML-style think blocks first (most common in modern LLM output)
	output = f.compressXMLBlocks(output, mode)

	// Handle markdown-style reasoning sections
	output = f.compressMarkdownCoT(output, mode)

	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

func (f *CoTCompressFilter) compressXMLBlocks(input string, mode Mode) string {
	for _, re := range []*regexp.Regexp{f.xmlThinkRe, f.xmlReasoningRe} {
		input = re.ReplaceAllStringFunc(input, func(match string) string {
			// Extract inner content
			inner := re.FindStringSubmatch(match)
			if len(inner) < 2 {
				return match
			}
			content := inner[1]
			toks := core.EstimateTokens(content)

			if mode == ModeAggressive {
				return "[thinking: " + tokLabel(toks) + " compressed]"
			}

			// ModeMinimal: keep first 30%
			lines := strings.Split(strings.TrimSpace(content), "\n")
			if len(lines) < f.minBlockLines {
				return match
			}
			keep := len(lines) * 30 / 100
			if keep < 2 {
				keep = 2
			}
			tag := xmlTagName(re)
			truncated := strings.Join(lines[:keep], "\n")
			return "<" + tag + ">\n" + truncated + "\n[... " + tokLabel(toks*70/100) + " omitted]\n</" + tag + ">"
		})
	}
	return input
}

func (f *CoTCompressFilter) compressMarkdownCoT(input string, mode Mode) string {
	lines := strings.Split(input, "\n")

	// Find runs of reasoning lines (step prefixes or reflection markers)
	type run struct{ start, end int }
	var runs []run
	inRun := false
	runStart := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		isReasoning := f.stepPrefixRe.MatchString(trimmed) || f.reflectionRe.MatchString(trimmed)
		if isReasoning && !inRun {
			inRun = true
			runStart = i
		} else if !isReasoning && trimmed == "" && inRun {
			// Blank line ends a reasoning run
			if i-runStart >= f.minBlockLines {
				runs = append(runs, run{runStart, i - 1})
			}
			inRun = false
		}
	}
	if inRun && len(lines)-runStart >= f.minBlockLines {
		runs = append(runs, run{runStart, len(lines) - 1})
	}

	if len(runs) == 0 {
		return input
	}

	suppress := make(map[int]bool)
	annotation := make(map[int]string)

	for _, r := range runs {
		block := lines[r.start : r.end+1]
		toks := core.EstimateTokens(strings.Join(block, "\n"))

		if mode == ModeAggressive {
			// Replace entire run with a single stub line
			annotation[r.start] = "[reasoning: " + tokLabel(toks) + " compressed]"
			for i := r.start + 1; i <= r.end; i++ {
				suppress[i] = true
			}
		} else {
			// ModeMinimal: keep first 30%, suppress the rest
			keep := (r.end - r.start + 1) * 30 / 100
			if keep < 2 {
				keep = 2
			}
			cutoff := r.start + keep
			omitted := r.end - cutoff + 1
			if omitted > 0 {
				annotation[cutoff] = "[... " + tokLabel(toks*(100-30)/100) + " reasoning omitted]"
				for i := cutoff + 1; i <= r.end; i++ {
					suppress[i] = true
				}
			}
		}
	}

	var result []string
	for i, line := range lines {
		if suppress[i] {
			continue
		}
		if ann, ok := annotation[i]; ok {
			if mode == ModeAggressive {
				result = append(result, ann)
			} else {
				result = append(result, line)
				result = append(result, ann)
			}
			continue
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

func tokLabel(n int) string {
	if n < 1000 {
		return "~" + cotItoa(n) + " tok"
	}
	return "~" + cotItoa(n/1000) + "k tok"
}

func cotItoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}

func xmlTagName(re *regexp.Regexp) string {
	src := re.String()
	if strings.Contains(src, "reasoning") {
		return "reasoning"
	}
	return "think"
}
