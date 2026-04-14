package filter

import (
	"strings"
)

// L15: ReasoningFilter - Unified filter for agent reasoning
// Merges: L26 LightThinker + L27 ThinkSwitcher + L28 GMSA + L29 CARL + L30 SlimInfer
//
// This filter optimizes agent reasoning traces and multi-step thinking:
// - LightThinker: Reduce thinking step verbosity
// - ThinkSwitcher: Switch between thinking modes
// - GMSA: Multi-step aggregation
// - CARL: Recursive reasoning compression
// - SlimInfer: Inference optimization

type ReasoningFilter struct{}

func NewReasoningFilter() *ReasoningFilter {
	return &ReasoningFilter{}
}

func (f *ReasoningFilter) Apply(input string, mode Mode) (string, int) {
	// Quick check: is this reasoning content?
	if !f.isReasoningContent(input) {
		return input, 0
	}

	// Apply unified reasoning compression
	compressed := f.compressReasoning(input)

	saved := len(input) - len(compressed)
	return compressed, saved
}

func (f *ReasoningFilter) isReasoningContent(input string) bool {
	// Check for reasoning patterns
	markers := []string{
		"Let me think", "Let's analyze", "Step by step",
		"First,", "Second,", "Third,", "Finally,",
		"Reasoning:", "Analysis:", "Thinking:",
		"I need to", "I should", "I'll",
	}

	for _, marker := range markers {
		if strings.Contains(input, marker) {
			return true
		}
	}
	return false
}

func (f *ReasoningFilter) compressReasoning(input string) string {
	lines := splitLines(input)
	if len(lines) < 10 {
		return input
	}

	var compressed []string
	inReasoning := false
	reasoningBuffer := []string{}

	for _, line := range lines {
		if f.isReasoningStart(line) {
			inReasoning = true
			reasoningBuffer = []string{line}
		} else if inReasoning && f.isReasoningEnd(line) {
			reasoningBuffer = append(reasoningBuffer, line)
			// Compress the reasoning block
			compressed = append(compressed, f.summarizeReasoning(reasoningBuffer))
			inReasoning = false
			reasoningBuffer = nil
		} else if inReasoning {
			reasoningBuffer = append(reasoningBuffer, line)
			// Prevent buffer from growing too large
			if len(reasoningBuffer) > 50 {
				compressed = append(compressed, f.summarizeReasoning(reasoningBuffer))
				reasoningBuffer = nil
				inReasoning = false
			}
		} else {
			compressed = append(compressed, line)
		}
	}

	if inReasoning && len(reasoningBuffer) > 0 {
		compressed = append(compressed, f.summarizeReasoning(reasoningBuffer))
	}

	return joinLines(compressed)
}

func (f *ReasoningFilter) isReasoningStart(line string) bool {
	starters := []string{
		"Let me think", "Let's analyze", "I need to",
		"Reasoning:", "Thinking:", "Analysis:",
		"Step 1:", "First,",
	}
	for _, s := range starters {
		if strings.Contains(line, s) {
			return true
		}
	}
	return false
}

func (f *ReasoningFilter) isReasoningEnd(line string) bool {
	enders := []string{
		"Therefore", "Conclusion:", "Answer:",
		"So,", "Thus,", "In conclusion",
		"The answer is", "Result:",
	}
	for _, e := range enders {
		if strings.Contains(line, e) {
			return true
		}
	}
	return false
}

func (f *ReasoningFilter) summarizeReasoning(lines []string) string {
	if len(lines) <= 3 {
		return joinLines(lines)
	}

	// Extract key steps
	var steps []string
	for i, line := range lines {
		if i == 0 || i == len(lines)-1 || f.isKeyStep(line) {
			steps = append(steps, line)
		}
		if len(steps) >= 5 {
			break
		}
	}

	if len(steps) < len(lines) {
		return "[Reasoning: " + joinLines(steps) + "]"
	}
	return joinLines(lines)
}

func (f *ReasoningFilter) isKeyStep(line string) bool {
	keyMarkers := []string{
		"Therefore", "Thus", "So", "Hence",
		"Key", "Important", "Critical",
		"However", "But", "Although",
		"Because", "Since", "As",
	}
	for _, marker := range keyMarkers {
		if strings.Contains(line, marker) {
			return true
		}
	}
	return false
}

func (f *ReasoningFilter) Name() string {
	return "L15_Reasoning"
}
