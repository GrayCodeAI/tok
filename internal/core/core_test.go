package core

import (
	"context"
	"testing"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input       string
		minExpected int
	}{
		{"", 0},
		{"a", 1},
		{"hello world", 2},
	}

	for _, tt := range tests {
		got := EstimateTokens(tt.input)
		if got < tt.minExpected {
			t.Errorf("EstimateTokens(%q) = %d, want >= %d", tt.input, got, tt.minExpected)
		}
	}
}

func TestCalculateTokensSaved(t *testing.T) {
	tests := []struct {
		original string
		filtered string
		minSaved int
	}{
		{"hello world", "hello", 1},
		{"same", "same", 0},
		{"short", "longer than original", 0},
		{"a b c d e f g h", "a c e g", 1},
	}

	for _, tt := range tests {
		got := CalculateTokensSaved(tt.original, tt.filtered)
		if got < tt.minSaved {
			t.Errorf("CalculateTokensSaved(%q, %q) = %d, want >= %d",
				tt.original, tt.filtered, got, tt.minSaved)
		}
	}
}

func TestCalculateSavings(t *testing.T) {
	// Test known model
	savings := CalculateSavings(1000000, "gpt-4o")
	if savings <= 0 {
		t.Errorf("CalculateSavings returned %f, want > 0", savings)
	}

	// Test unknown model (falls back to default)
	savings = CalculateSavings(1000000, "unknown_model_xyz")
	if savings <= 0 {
		t.Errorf("CalculateSavings for unknown model returned %f, want > 0", savings)
	}
}

func TestOSCommandRunner_Run(t *testing.T) {
	runner := NewOSCommandRunner()

	tests := []struct {
		name     string
		args     []string
		wantExit int
	}{
		{"echo hello", []string{"echo", "hello"}, 0},
		{"echo empty", []string{"echo"}, 0},
		{"true", []string{"true"}, 0},
		{"false", []string{"false"}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, exitCode, _ := runner.Run(ctx, tt.args)
			if exitCode != tt.wantExit {
				t.Errorf("exitCode = %d, want %d", exitCode, tt.wantExit)
			}
		})
	}
}

func TestOSCommandRunner_Run_Empty(t *testing.T) {
	runner := NewOSCommandRunner()
	ctx := context.Background()
	output, exitCode, err := runner.Run(ctx, []string{})
	if output != "" || exitCode != 0 || err != nil {
		t.Errorf("empty run should return empty, 0, nil: got %v, %d, %v", output, exitCode, err)
	}
}

func TestOSCommandRunner_Run_NotFound(t *testing.T) {
	runner := NewOSCommandRunner()
	ctx := context.Background()
	_, exitCode, err := runner.Run(ctx, []string{"nonexistent_cmd_xyz"})
	if exitCode != 127 {
		t.Errorf("exitCode = %d, want 127", exitCode)
	}
	if err == nil {
		t.Error("expected error for nonexistent command")
	}
}

func TestOSCommandRunner_LookPath(t *testing.T) {
	runner := NewOSCommandRunner()
	path, err := runner.LookPath("ls")
	if err != nil || path == "" {
		t.Errorf("LookPath(ls) = %q, %v", path, err)
	}

	_, err = runner.LookPath("nonexistent_cmd_xyz_123")
	if err == nil {
		t.Error("LookPath for nonexistent should error")
	}
}

func TestCommonModelPricing(t *testing.T) {
	if len(CommonModelPricing) == 0 {
		t.Error("CommonModelPricing should not be empty")
	}

	// Test that prices are positive
	for name, pricing := range CommonModelPricing {
		if pricing.InputPerMillion <= 0 {
			t.Errorf("model %s has non-positive input price", name)
		}
		if pricing.OutputPerMillion <= 0 {
			t.Errorf("model %s has non-positive output price", name)
		}
	}
}
