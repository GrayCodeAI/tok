package filter

import (
	"strings"
	"testing"
)

func TestThinkSwitcherFilter_Name(t *testing.T) {
	f := NewThinkSwitcherFilter()
	if f.Name() != "27_think_switcher" {
		t.Errorf("unexpected name: %s", f.Name())
	}
}

func TestThinkSwitcherFilter_ModeNone(t *testing.T) {
	f := NewThinkSwitcherFilter()
	input := strings.Repeat("Step 1: Do this.\nStep 2: Do that.\n", 10)
	out, saved := f.Apply(input, ModeNone)
	if out != input || saved != 0 {
		t.Error("ModeNone should return unchanged")
	}
}

func TestThinkSwitcherFilter_FastPathDirectAnswer(t *testing.T) {
	f := NewThinkSwitcherFilter()
	// Direct answer with no reasoning markers
	lines := []string{
		"The answer to your question is 42.",
		"This is based on the standard formula for this calculation.",
		"You can verify this by running the provided test case.",
		"The implementation is in src/math/calculator.go at line 45.",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	// Fast path: should not compress (reasoning density < threshold)
	if saved != 0 {
		t.Errorf("fast path: direct answer should not be compressed, got %d saved", saved)
	}
	if out != input {
		t.Error("fast path: output should equal input")
	}
}

func TestThinkSwitcherFilter_SlowPathHeavyReasoning(t *testing.T) {
	f := NewThinkSwitcherFilter()
	// Heavy reasoning output - many step/think markers
	lines := []string{
		"Let me think through this problem carefully.",
		"First, I need to understand the database schema.",
		"Second, I should check the migration history.",
		"Third, I need to verify the foreign key constraints.",
		"Fourth, let me consider the transaction boundaries.",
		"Fifth, I should look at the locking behavior.",
		"Let me reconsider the approach I was taking.",
		"Actually, the root issue is the index on user_id.",
		"Now, let me trace through the query execution plan.",
		"Then, I can identify where the bottleneck occurs.",
		"Therefore, the fix is to add a composite index.",
		"Thus, the query will use the index instead of full scan.",
		"",
		"Answer: Add composite index on (user_id, created_at).",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeAggressive)

	if saved <= 0 {
		t.Error("expected savings on heavy reasoning output")
	}
	if !strings.Contains(out, "Answer:") {
		t.Error("final answer must be preserved")
	}
}

func TestThinkSwitcherFilter_LightPath(t *testing.T) {
	f := NewThinkSwitcherFilter()
	// Moderate reasoning - some steps but not overwhelming
	lines := []string{
		"The server is returning 500 errors.",
		"Let me check the application logs.",
		"First, look at the error message in the logs.",
		"The error is a null pointer exception in UserService.",
		"The service is initialized before the database connection.",
		"Checking the startup order in the configuration.",
		"The fix is to delay UserService initialization.",
		"",
		"Root cause: initialization order dependency.",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	// Should apply some compression (light path)
	if saved < 0 {
		t.Error("light path should not increase token count")
	}
	if !strings.Contains(out, "Root cause") {
		t.Error("conclusion must be preserved")
	}
}

func TestThinkSwitcherFilter_AggressiveMoreReduction(t *testing.T) {
	f := NewThinkSwitcherFilter()
	var lines []string
	for i := 0; i < 5; i++ {
		lines = append(lines,
			"Let me think about this systematically.",
			"First, consider the architectural implications.",
			"Second, analyze the performance characteristics.",
			"Third, evaluate the maintenance overhead.",
		)
	}
	lines = append(lines, "Conclusion: use the simpler approach.")
	input := strings.Join(lines, "\n")

	_, savedMin := f.Apply(input, ModeMinimal)
	_, savedAgg := f.Apply(input, ModeAggressive)
	if savedAgg < savedMin {
		t.Error("aggressive should save >= minimal")
	}
}
