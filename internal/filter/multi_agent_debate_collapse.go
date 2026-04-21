package filter

import (
	"strings"

	"github.com/GrayCodeAI/tok/internal/core"
)

// Paper: "S2-MAD: Semantic-Similarity Multi-Agent Debate Compression"
// NAACL 2025
//
// S2MADFilter detects agreement phrases in multi-agent debate or review outputs
// ("I agree with", "As X mentioned", "Building on that", "This is correct") and
// collapses those agreement passages into compact markers while preserving the
// novel arguments in each agent turn.
//
// Multi-agent debate outputs are common in:
//   - LLM self-critique and revision loops
//   - Peer-review style agent pipelines
//   - RAG reranker debate outputs
//
// The filter operates in two stages:
//  1. Passage scoring: each line is checked for agreement/acknowledgement markers.
//     Lines with such markers score near 0; lines with novel claims score near 1.
//  2. Agreement run collapsing: consecutive agreement-heavy lines are merged into
//     a single "[agreement: N lines]" marker, preserving surrounding novel content.
type S2MADFilter struct {
	agreementThreshold float64 // agreement-marker density to trigger collapsing
	minRunLength       int     // minimum run of agreement lines to collapse
}

// NewS2MADFilter creates a new S2-MAD multi-agent debate compression filter.
func NewS2MADFilter() *S2MADFilter {
	return &S2MADFilter{
		agreementThreshold: 0.5,
		minRunLength:       2,
	}
}

// Name returns the filter name.
func (f *S2MADFilter) Name() string { return "35_s2_mad" }

// Apply collapses agreement passages and preserves novel arguments.
func (f *S2MADFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	lines := strings.Split(input, "\n")
	if len(lines) < 4 {
		return input, 0
	}

	if !s2madLooksLikeDebate(lines) {
		return input, 0
	}

	minRun := f.minRunLength
	if mode == ModeAggressive {
		minRun = 1
	}

	// Mark agreement lines
	isAgreement := make([]bool, len(lines))
	for i, line := range lines {
		isAgreement[i] = s2madIsAgreementLine(line)
	}

	// Collapse consecutive agreement runs of length ≥ minRun
	var result []string
	i := 0
	changed := false
	for i < len(lines) {
		if isAgreement[i] && strings.TrimSpace(lines[i]) != "" {
			// Measure run length
			j := i
			for j < len(lines) && isAgreement[j] {
				j++
			}
			runLen := j - i
			if runLen >= minRun {
				result = append(result, "[agreement: "+itoa(runLen)+" lines collapsed]")
				i = j
				changed = true
				continue
			}
		}
		result = append(result, lines[i])
		i++
	}

	if !changed {
		return input, 0
	}

	output := strings.Join(result, "\n")
	saved := core.EstimateTokens(input) - core.EstimateTokens(output)
	if saved < 0 {
		saved = 0
	}
	return output, saved
}

// agreementPhrases are markers that indicate a line is acknowledging/agreeing.
var s2madAgreementPhrases = []string{
	"i agree", "i agree with", "agreed,", "i concur",
	"as you mentioned", "as mentioned", "as stated",
	"as noted", "as pointed out", "as highlighted",
	"building on that", "building on this",
	"you are correct", "that is correct", "this is correct",
	"that's right", "that's a good point", "good point",
	"this aligns with", "this is consistent with",
	"echoing the", "supporting the view",
	"i think you're right", "i think that's right",
	"same as above", "similar to what",
	"to add to that", "adding to what",
}

// s2madIsAgreementLine returns true if the line is primarily an agreement expression.
func s2madIsAgreementLine(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	if lower == "" {
		return false
	}
	for _, phrase := range s2madAgreementPhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

// s2madLooksLikeDebate returns true if the input resembles debate/review output.
func s2madLooksLikeDebate(lines []string) bool {
	agreementCount := 0
	speakerCount := 0
	speakerMarkers := []string{
		"agent", "model", "reviewer", "critic", "expert",
		"assistant", "debater", "participant",
	}
	for _, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		if s2madIsAgreementLine(line) {
			agreementCount++
		}
		for _, m := range speakerMarkers {
			if strings.HasPrefix(lower, m) && (strings.Contains(lower, ":") || strings.Contains(lower, " 1") || strings.Contains(lower, " 2")) {
				speakerCount++
				break
			}
		}
	}
	return agreementCount >= 2 || speakerCount >= 2
}
