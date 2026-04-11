package filter

import (
	"strings"
	"testing"
)

func TestCoTCompressFilter_Name(t *testing.T) {
	f := NewCoTCompressFilter()
	if f.Name() != "23_cot_compress" {
		t.Errorf("unexpected name: %s", f.Name())
	}
}

func TestCoTCompressFilter_ModeNone(t *testing.T) {
	f := NewCoTCompressFilter()
	input := "<think>Let me think about this carefully step by step.\nFirst, I consider the problem.\nThen I analyze options.\nFinally I decide.\n</think>\nThe answer is 42."
	out, saved := f.Apply(input, ModeNone)
	if out != input || saved != 0 {
		t.Error("ModeNone should return input unchanged")
	}
}

func TestCoTCompressFilter_AggressiveRemovesThinkBlock(t *testing.T) {
	f := NewCoTCompressFilter()
	think := "<think>\n" + strings.Repeat("I need to carefully analyze this problem step by step.\n", 15) + "</think>"
	input := think + "\nFinal answer: the solution is X."
	out, saved := f.Apply(input, ModeAggressive)

	if saved <= 0 {
		t.Error("expected positive savings on aggressive CoT compression")
	}
	if strings.Contains(out, "carefully analyze") {
		t.Error("think block content should be removed in aggressive mode")
	}
	if !strings.Contains(out, "[thinking:") {
		t.Error("expected [thinking: ...] stub in aggressive output")
	}
	if !strings.Contains(out, "Final answer") {
		t.Error("content after think block must be preserved")
	}
}

func TestCoTCompressFilter_MinimalTruncatesThinkBlock(t *testing.T) {
	f := NewCoTCompressFilter()
	thinkLines := make([]string, 20)
	for i := range thinkLines {
		thinkLines[i] = "Reasoning step line number in the thinking block for analysis."
	}
	input := "<think>\n" + strings.Join(thinkLines, "\n") + "\n</think>\nConclusion reached."
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Error("expected savings in minimal mode for large think block")
	}
	if !strings.Contains(out, "<think>") {
		t.Error("minimal mode should preserve think tags")
	}
	if !strings.Contains(out, "omitted") {
		t.Error("minimal mode should annotate omitted content")
	}
	if !strings.Contains(out, "Conclusion reached") {
		t.Error("content after block must be preserved")
	}
}

func TestCoTCompressFilter_NoCoTPassthrough(t *testing.T) {
	f := NewCoTCompressFilter()
	input := "Here is a direct answer.\nNo reasoning traces here.\nJust plain output lines."
	out, saved := f.Apply(input, ModeMinimal)
	if saved != 0 {
		t.Errorf("no CoT to compress, expected 0 savings, got %d", saved)
	}
	if out != input {
		t.Error("input without CoT should be unchanged")
	}
}

func TestCoTCompressFilter_MarkdownReasoningSteps(t *testing.T) {
	f := NewCoTCompressFilter()
	lines := []string{
		"Step 1: Identify the root cause of the issue.",
		"Step 2: Analyze the stack trace for relevant frames.",
		"Step 3: Check the database connection pool settings.",
		"Step 4: Verify the environment variables are correct.",
		"Step 5: Review the recent deployment changes.",
		"Step 6: Formulate a hypothesis about the failure.",
		"",
		"The answer is: increase pool size to 20.",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeAggressive)

	if saved <= 0 {
		t.Error("expected savings on markdown reasoning steps")
	}
	if !strings.Contains(out, "The answer is") {
		t.Error("conclusion must be preserved")
	}
}
