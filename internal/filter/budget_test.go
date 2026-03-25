package filter

import (
	"strings"
	"testing"
)

func TestNewBudgetEnforcer(t *testing.T) {
	b := NewBudgetEnforcer(100)
	if b == nil {
		t.Fatal("NewBudgetEnforcer returned nil")
	}
	if b.Name() != "budget" {
		t.Errorf("Name() = %q, want 'budget'", b.Name())
	}
}

func TestBudgetEnforcer_NoLimit(t *testing.T) {
	b := NewBudgetEnforcer(0)
	input := "line1\nline2\nline3"
	output, saved := b.Apply(input, ModeMinimal)
	if output != input {
		t.Error("no budget limit should pass through unchanged")
	}
	if saved != 0 {
		t.Errorf("no budget should save 0, got %d", saved)
	}
}

func TestBudgetEnforcer_UnderBudget(t *testing.T) {
	b := NewBudgetEnforcer(1000)
	input := "short output\n"
	output, saved := b.Apply(input, ModeMinimal)
	if output != input {
		t.Error("under budget should pass through unchanged")
	}
	if saved != 0 {
		t.Errorf("under budget should save 0, got %d", saved)
	}
}

func TestBudgetEnforcer_OverBudget(t *testing.T) {
	b := NewBudgetEnforcer(50)
	// Generate content well over budget
	lines := make([]string, 200)
	for i := range lines {
		lines[i] = "This is a verbose log line with information that may not be critical for debugging purposes."
	}
	input := strings.Join(lines, "\n")
	output, saved := b.Apply(input, ModeMinimal)
	if saved <= 0 {
		t.Error("over budget should save tokens")
	}
	// Output should be shorter than input
	if len(output) >= len(input) {
		t.Error("output should be shorter than input when over budget")
	}
}

func TestBudgetEnforcer_PreservesErrors(t *testing.T) {
	b := NewBudgetEnforcer(30)
	input := strings.Repeat("debug: verbose info\n", 50) +
		"ERROR: critical failure in database connection\n" +
		strings.Repeat("debug: more verbose info\n", 50)
	output, _ := b.Apply(input, ModeMinimal)
	if !strings.Contains(output, "ERROR") {
		t.Error("should preserve ERROR lines")
	}
}

func TestBudgetEnforcer_SetBudget(t *testing.T) {
	b := NewBudgetEnforcer(100)
	b.SetBudget(50)
	if b.budget != 50 {
		t.Errorf("budget = %d, want 50", b.budget)
	}
}

func TestNewBudgetEnforcerWithConfig(t *testing.T) {
	cfg := BudgetConfig{Budget: 200}
	b := NewBudgetEnforcerWithConfig(cfg)
	if b.budget != 200 {
		t.Errorf("budget = %d, want 200", b.budget)
	}
}
