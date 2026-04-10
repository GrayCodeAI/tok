package filter

import (
	"regexp"
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Paper: "LightThinker: Thinking Step-by-Step Compression"
// EMNLP 2025 — Zhang et al., Zhejiang University
// https://arxiv.org/abs/2502.15589
//
// LightThinkerFilter compresses reasoning output at step granularity.
// Unlike CoTCompressFilter which truncates whole blocks, LightThinker
// retains all steps but skeletonises each one to its single most
// informative sentence — a sketch of each reasoning step.
//
// Algorithm:
//  1. Segment input into reasoning steps (numbered/labelled sequences)
//  2. For each step with ≥ minStepLines, extract the "key sentence":
//     the line with the highest unique-term density relative to the step
//  3. Replace the step body with: step header + key sentence + stub
//  4. Pass non-step lines through unchanged
//
// Key insight: keeping one sentence per step (the conclusion/observation)
// preserves the logical trajectory while cutting 60-80% of step content.
type LightThinkerFilter struct {
	stepHeaderRe  *regexp.Regexp
	ordinalRe     *regexp.Regexp
	minStepLines  int // minimum body lines before compressing a step
}

// NewLightThinkerFilter creates a new LightThinker step-level compressor.
func NewLightThinkerFilter() *LightThinkerFilter {
	return &LightThinkerFilter{
		stepHeaderRe: regexp.MustCompile(`(?i)^(step\s+\d+[:.)]|\d+[.)\s]\s+)`),
		ordinalRe:    regexp.MustCompile(`(?i)^(first[,:]?|second[,:]?|third[,:]?|fourth[,:]?|fifth[,:]?|finally[,:]?|lastly[,:]?|next[,:]?)`),
		minStepLines: 3,
	}
}

// Name returns the filter name.
func (f *LightThinkerFilter) Name() string { return "26_lightthinker" }

// Apply skeletonises reasoning steps to one key sentence each.
func (f *LightThinkerFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")
	if len(lines) < f.minStepLines*2 {
		return input, 0
	}

	steps := f.segmentSteps(lines)
	if len(steps) == 0 {
		return input, 0
	}

	// Mark lines to replace
	type replacement struct {
		startBody int
		endBody   int
		keySent   string
		dropped   int
	}
	var replacements []replacement

	for _, s := range steps {
		body := lines[s.bodyStart : s.bodyEnd+1]
		if len(body) < f.minStepLines {
			continue
		}
		key := f.extractKeySentence(body)
		if key == "" || key == body[0] {
			continue
		}
		replacements = append(replacements, replacement{
			startBody: s.bodyStart,
			endBody:   s.bodyEnd,
			keySent:   key,
			dropped:   len(body) - 1,
		})
	}

	if len(replacements) == 0 {
		return input, 0
	}

	suppress := make(map[int]bool)
	inject := make(map[int]string) // after this line, inject replacement

	for _, r := range replacements {
		// Keep first line of body (often contains key info in first sentence)
		// Suppress body[1..end], inject key sentence + stub after body[0]
		for i := r.startBody + 1; i <= r.endBody; i++ {
			suppress[i] = true
		}
		if r.dropped > 1 {
			inject[r.startBody] = r.keySent + "\n[~" + itoa(r.dropped) + " step lines compressed]"
		} else {
			inject[r.startBody] = r.keySent
		}
	}

	var result []string
	for i, line := range lines {
		if suppress[i] {
			continue
		}
		result = append(result, line)
		if inj, ok := inject[i]; ok {
			result = append(result, inj)
		}
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

type stepSpan struct {
	headerLine int
	bodyStart  int
	bodyEnd    int
}

// segmentSteps identifies numbered/ordinal reasoning step sequences.
func (f *LightThinkerFilter) segmentSteps(lines []string) []stepSpan {
	var steps []stepSpan
	i := 0
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if f.stepHeaderRe.MatchString(trimmed) || f.ordinalRe.MatchString(trimmed) {
			// Find end of step body (next step header or blank line after content)
			bodyStart := i + 1
			j := bodyStart
			for j < len(lines) {
				t := strings.TrimSpace(lines[j])
				if f.stepHeaderRe.MatchString(t) || f.ordinalRe.MatchString(t) {
					break
				}
				if t == "" && j > bodyStart+1 {
					// blank line terminates step body
					break
				}
				j++
			}
			bodyEnd := j - 1
			// Skip trailing blank lines
			for bodyEnd > bodyStart && strings.TrimSpace(lines[bodyEnd]) == "" {
				bodyEnd--
			}
			if bodyEnd >= bodyStart {
				steps = append(steps, stepSpan{
					headerLine: i,
					bodyStart:  bodyStart,
					bodyEnd:    bodyEnd,
				})
			}
			i = j
		} else {
			i++
		}
	}
	return steps
}

// extractKeySentence returns the most informative line from a step body.
// "Most informative" = highest ratio of unique terms to line length.
func (f *LightThinkerFilter) extractKeySentence(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	// Build term frequency across the step
	termFreq := make(map[string]int)
	for _, line := range lines {
		for _, t := range ltTokenize(line) {
			termFreq[t]++
		}
	}

	// Score each line: sum of (1/freq) for each term it contains — rare terms score higher
	bestScore := -1.0
	bestLine := lines[0]
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		terms := ltTokenize(line)
		if len(terms) == 0 {
			continue
		}
		score := 0.0
		for _, t := range terms {
			if f := termFreq[t]; f > 0 {
				score += 1.0 / float64(f)
			}
		}
		score /= float64(len(terms)) // normalise by line length
		if score > bestScore {
			bestScore = score
			bestLine = trimmed
		}
	}
	return bestLine
}

// ltTokenize splits a line into lowercase tokens ≥ 3 chars.
func ltTokenize(line string) []string {
	var terms []string
	var word strings.Builder
	for _, ch := range strings.ToLower(line) {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' {
			word.WriteRune(ch)
		} else if word.Len() > 0 {
			if w := word.String(); len(w) >= 3 {
				terms = append(terms, w)
			}
			word.Reset()
		}
	}
	if word.Len() >= 3 {
		terms = append(terms, word.String())
	}
	return terms
}
