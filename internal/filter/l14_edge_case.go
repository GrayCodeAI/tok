package filter

import (
	"fmt"
	"strings"
)

// L14: EdgeCaseFilter - Unified filter for rare/special cases
// Merges: L21 MarginalInfoGain + L22 NearDedup + L23 CoTCompress + L24 CodingAgentContext + L25 PerceptionCompress
//
// This filter handles edge cases that don't fit standard compression patterns:
// - Marginal info gain: removes info that doesn't add new knowledge
// - Near-dedup: fuzzy duplicate detection
// - CoT compression: chain-of-thought reasoning compression
// - Coding context: special handling for code reviews
// - Perception: image+text multi-modal hints

type EdgeCaseFilter struct {
	config *EdgeCaseConfig
}

type EdgeCaseConfig struct {
	NearDedupThreshold    float64 // 0.0-1.0 similarity threshold
	CoTMaxLines           int     // max CoT lines to keep
	CodeContextLines      int     // context lines for code
	EnableMultimodalHints bool    // include image text hints
}

func NewEdgeCaseFilter() *EdgeCaseFilter {
	return &EdgeCaseFilter{
		config: &EdgeCaseConfig{
			NearDedupThreshold:    0.85,
			CoTMaxLines:           20,
			CodeContextLines:      5,
			EnableMultimodalHints: true,
		},
	}
}

func (f *EdgeCaseFilter) Apply(input string, mode Mode) (string, int) {
	// Stage 1: Skip if content is normal (fast path)
	if !f.hasEdgeCaseContent(input) {
		return input, 0
	}

	// Stage 2: Apply near-deduplication (L22)
	deduped := f.nearDedup(input)

	// Stage 3: Compress CoT if present (L23)
	cotCompressed := f.compressCoT(deduped)

	// Stage 4: Optimize coding context (L24)
	codeOptimized := f.optimizeCodeContext(cotCompressed)

	// Stage 5: Add perception hints if multimodal (L25)
	final := f.addPerceptionHints(codeOptimized)

	saved := len(input) - len(final)
	return final, saved
}

func (f *EdgeCaseFilter) hasEdgeCaseContent(input string) bool {
	// Check for CoT patterns
	if contains(input, "Let's think", "Step 1:", "First,", "Therefore,") {
		return true
	}
	// Check for near-duplicate lines
	lines := splitLines(input)
	if len(lines) > 50 {
		similarCount := 0
		for i := 0; i < len(lines)-1 && i < 20; i++ {
			if similarity(lines[i], lines[i+1]) > 0.8 {
				similarCount++
			}
		}
		if similarCount > 5 {
			return true
		}
	}
	// Check for code review patterns
	if contains(input, "```", "func ", "class ", "def ") && len(input) > 5000 {
		return true
	}
	return false
}

func (f *EdgeCaseFilter) nearDedup(input string) string {
	lines := splitLines(input)
	if len(lines) < 20 {
		return input
	}

	var unique []string
	for _, line := range lines {
		isDup := false
		for _, existing := range unique {
			if similarity(line, existing) > f.config.NearDedupThreshold {
				isDup = true
				break
			}
		}
		if !isDup {
			unique = append(unique, line)
		}
	}

	if float64(len(unique)) < float64(len(lines))*0.9 {
		return joinLines(unique) + fmt.Sprintf("\n[near-dedup: %d→%d lines]", len(lines), len(unique))
	}
	return input
}

func (f *EdgeCaseFilter) compressCoT(input string) string {
	if !contains(input, "Let's think", "Step 1:", "Reasoning:") {
		return input
	}

	lines := splitLines(input)
	var compressed []string
	inCoT := false
	coTBuffer := []string{}

	for _, line := range lines {
		if isCoTStart(line) {
			inCoT = true
			coTBuffer = []string{line}
		} else if inCoT && isCoTEnd(line) {
			coTBuffer = append(coTBuffer, line)
			// Compress CoT section
			compressed = append(compressed, f.summarizeCoT(coTBuffer))
			inCoT = false
			coTBuffer = nil
		} else if inCoT {
			coTBuffer = append(coTBuffer, line)
			if len(coTBuffer) > f.config.CoTMaxLines {
				// Too long, compress now
				compressed = append(compressed, f.summarizeCoT(coTBuffer))
				coTBuffer = nil
				inCoT = false
			}
		} else {
			compressed = append(compressed, line)
		}
	}

	if inCoT && len(coTBuffer) > 0 {
		compressed = append(compressed, f.summarizeCoT(coTBuffer))
	}

	return joinLines(compressed)
}

func (f *EdgeCaseFilter) summarizeCoT(lines []string) string {
	if len(lines) <= 3 {
		return joinLines(lines)
	}
	// Keep first (setup), last (conclusion), and key intermediate steps
	keySteps := []string{lines[0]}

	// Add key intermediate steps (every Nth line)
	stepSize := len(lines) / 3
	if stepSize < 1 {
		stepSize = 1
	}
	for i := stepSize; i < len(lines)-1; i += stepSize {
		if len(keySteps) < 4 {
			keySteps = append(keySteps, lines[i])
		}
	}
	keySteps = append(keySteps, lines[len(lines)-1])

	return "[CoT: " + joinLines(keySteps) + "]"
}

func (f *EdgeCaseFilter) optimizeCodeContext(input string) string {
	if !contains(input, "```") {
		return input
	}

	// For code blocks, keep signature + first N lines + last N lines
	lines := splitLines(input)
	var optimized []string
	inCode := false
	codeBuffer := []string{}

	for _, line := range lines {
		if strings.HasPrefix(line, "```") && !inCode {
			inCode = true
			codeBuffer = []string{line}
		} else if strings.HasPrefix(line, "```") && inCode {
			codeBuffer = append(codeBuffer, line)
			// Process code block
			optimized = append(optimized, f.compressCodeBlock(codeBuffer))
			inCode = false
			codeBuffer = nil
		} else if inCode {
			codeBuffer = append(codeBuffer, line)
		} else {
			optimized = append(optimized, line)
		}
	}

	if inCode && len(codeBuffer) > 0 {
		optimized = append(optimized, f.compressCodeBlock(codeBuffer))
	}

	return joinLines(optimized)
}

func (f *EdgeCaseFilter) compressCodeBlock(lines []string) string {
	if len(lines) <= f.config.CodeContextLines*2+2 {
		return joinLines(lines)
	}

	// Keep: opening ```, signature, first N lines, [...], last N lines, closing ```
	result := []string{
		lines[0], // opening ```
		lines[1], // signature if present
	}

	// First N context lines
	for i := 2; i < len(lines)-1 && i < 2+f.config.CodeContextLines; i++ {
		result = append(result, lines[i])
	}

	result = append(result, fmt.Sprintf("... [%d lines] ...", len(lines)-2*f.config.CodeContextLines-2))

	// Last N context lines
	for i := len(lines) - 1 - f.config.CodeContextLines; i < len(lines)-1; i++ {
		if i > 2+f.config.CodeContextLines {
			result = append(result, lines[i])
		}
	}

	result = append(result, lines[len(lines)-1]) // closing ```
	return joinLines(result)
}

func (f *EdgeCaseFilter) addPerceptionHints(input string) string {
	if !f.config.EnableMultimodalHints {
		return input
	}
	// Add hints for image content if detected
	if contains(input, "[image", "<image", "![") {
		return "[multimodal content] " + input
	}
	return input
}

func (f *EdgeCaseFilter) Name() string {
	return "L14_EdgeCase"
}

// Helper functions
func isCoTStart(line string) bool {
	return contains(line, "Let's think", "Reasoning:", "Step 1:", "Analysis:")
}

func isCoTEnd(line string) bool {
	return contains(line, "Therefore", "Conclusion:", "Answer:", "Result:")
}

func similarity(a, b string) float64 {
	// Simple Jaccard similarity
	setA := make(map[string]bool)
	setB := make(map[string]bool)

	for _, w := range strings.Fields(strings.ToLower(a)) {
		setA[w] = true
	}
	for _, w := range strings.Fields(strings.ToLower(b)) {
		setB[w] = true
	}

	intersection := 0
	for w := range setA {
		if setB[w] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// Helper functions
func contains(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}
