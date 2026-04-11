package filter

import (
	"strings"
	"testing"
)

func TestNewGoalDrivenFilter(t *testing.T) {
	f := NewGoalDrivenFilter("debug error in function")
	if f == nil {
		t.Fatal("NewGoalDrivenFilter returned nil")
	}
	if f.Name() != "goal_driven" {
		t.Errorf("Name() = %q, want 'goal_driven'", f.Name())
	}
	if f.mode != GoalModeDebug {
		t.Errorf("mode = %v, want GoalModeDebug", f.mode)
	}
}

func TestParseGoalMode(t *testing.T) {
	tests := []struct {
		goal string
		want GoalMode
	}{
		{"debug the error", GoalModeDebug},
		{"fix the bug", GoalModeDebug},
		{"review the code", GoalModeReview},
		{"check for issues", GoalModeReview},
		{"deploy to prod", GoalModeDeploy},
		{"release version", GoalModeDeploy},
		{"search for patterns", GoalModeSearch},
		{"find the issue", GoalModeSearch},
		{"build the project", GoalModeBuild},
		{"compile the code", GoalModeBuild},
		{"run the tests", GoalModeTest},
		{"something random", GoalModeGeneric},
	}

	for _, tt := range tests {
		t.Run(tt.goal, func(t *testing.T) {
			got := parseGoalMode(tt.goal)
			if got != tt.want {
				t.Errorf("parseGoalMode(%q) = %v, want %v", tt.goal, got, tt.want)
			}
		})
	}
}

func TestGoalDrivenFilter_Apply_None(t *testing.T) {
	f := NewGoalDrivenFilter("debug")
	input := "error: something failed\nwarning: deprecated\ninfo: running"
	output, saved := f.Apply(input, ModeNone)
	if output != input {
		t.Error("ModeNone should not modify input")
	}
	if saved != 0 {
		t.Errorf("ModeNone should save 0, got %d", saved)
	}
}

func TestGoalDrivenFilter_Apply_Minimal(t *testing.T) {
	f := NewGoalDrivenFilter("debug error")
	input := strings.Repeat("error: something failed\nwarning: deprecated\ninfo: running\n", 20)
	output, saved := f.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("output should not be empty")
	}
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
}

func TestGoalDrivenFilter_Apply_Aggressive(t *testing.T) {
	f := NewGoalDrivenFilter("debug error")
	input := strings.Repeat("error: something failed\nwarning: deprecated\ninfo: running\nsome filler content\n", 30)
	output, saved := f.Apply(input, ModeAggressive)
	if output == "" {
		t.Error("output should not be empty")
	}
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
}

func TestGoalDrivenFilter_ErrorLineDetection(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"Error: something failed", true},
		{"Exception in thread main", true},
		{"panic: runtime error", true},
		{"fatal: could not read", true},
		{"test failed: assertion", true},
		{"info: running normally", false},
		{"hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.line[:min(20, len(tt.line))], func(t *testing.T) {
			got := isErrorLine(tt.line)
			if got != tt.want {
				t.Errorf("isErrorLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestGoalDrivenFilter_WarningLineDetection(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"Warning: deprecated API", true},
		{"warn: low memory", true},
		{"Caution: this is risky", true},
		{"info: all good", false},
	}

	for _, tt := range tests {
		t.Run(tt.line[:min(20, len(tt.line))], func(t *testing.T) {
			got := isWarningLine(tt.line)
			if got != tt.want {
				t.Errorf("isWarningLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestGoalDrivenFilter_HeadingLineDetection(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"# Heading", true},
		{"## Subheading", true},
		{"----", true},
		{"========", true},
		{"not a heading", false},
		{"-", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := isHeadingLine(tt.line)
			if got != tt.want {
				t.Errorf("isHeadingLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestGoalDrivenFilter_CodeLineDetection(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"func main() {", true},
		{"def hello():", true},
		{"class Foo:", true},
		{"if x > 0:", true},
		{"for i in range(10):", true},
		{"return value", true},
		{"import os", true},
		{"const x = 1", true},
		{"let y = 2", true},
		{"var z = 3", true},
		{"hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := isCodeLine(tt.line)
			if got != tt.want {
				t.Errorf("isCodeLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestGoalDrivenFilter_CRFDecode(t *testing.T) {
	f := NewGoalDrivenFilter("debug")

	// Empty scores
	decisions := f.crfDecode([]float64{})
	if len(decisions) != 0 {
		t.Errorf("crfDecode empty scores returned %d decisions, want 0", len(decisions))
	}

	// All zero scores
	decisions = f.crfDecode([]float64{0, 0, 0})
	if len(decisions) != 3 {
		t.Errorf("crfDecode returned %d decisions, want 3", len(decisions))
	}
}

func TestGoalDrivenFilter_EnsureCoherence(t *testing.T) {
	f := NewGoalDrivenFilter("debug")

	// Test that first and last scored lines are kept
	scores := []float64{0, 1.0, 2.0, 3.0, 0}
	decisions := make([]bool, len(scores))
	f.ensureCoherence(decisions, scores)

	if !decisions[1] {
		t.Error("first scored line should be kept")
	}
	if !decisions[3] {
		t.Error("last scored line should be kept")
	}
}

func TestGoalDrivenFilter_ScoreLine(t *testing.T) {
	f := NewGoalDrivenFilter("debug error")

	// Error line should score higher
	errorScore := f.scoreLine("Error: something failed badly")
	normalScore := f.scoreLine("this is normal output text")

	if errorScore <= normalScore {
		t.Errorf("error line score (%f) should be > normal line score (%f)", errorScore, normalScore)
	}
}

func BenchmarkGoalDrivenFilter_Apply(b *testing.B) {
	f := NewGoalDrivenFilter("debug error")
	input := strings.Repeat("error: something failed\nwarning: deprecated\ninfo: running\nsome filler content\n", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Apply(input, ModeMinimal)
	}
}
