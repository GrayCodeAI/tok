package core

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestOSCommandRunnerRun(t *testing.T) {
	runner := NewOSCommandRunner()

	tests := []struct {
		name       string
		args       []string
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "echo simple",
			args:       []string{"echo", "hello"},
			wantOutput: "hello",
			wantErr:    false,
		},
		{
			name:       "printf",
			args:       []string{"printf", "test"},
			wantOutput: "test",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			output, exitCode, err := runner.Run(ctx, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output = %q, want to contain %q", output, tt.wantOutput)
			}
			if exitCode != 0 {
				t.Errorf("exitCode = %d, want 0", exitCode)
			}
		})
	}
}

func TestOSCommandRunnerLookupPath(t *testing.T) {
	runner := NewOSCommandRunner()

	// Should find common commands
	path, err := runner.LookPath("echo")
	if err != nil {
		t.Fatalf("LookPath(echo): %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path for echo")
	}
}

func TestOSCommandRunnerTimeout(t *testing.T) {
	runner := NewOSCommandRunner()

	// Command that should timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, _, err := runner.Run(ctx, []string{"sleep", "5"})
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestEstimateTokensConsistent(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"non-empty input returns positive", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateTokens(tt.input)
			if got <= 0 {
				t.Errorf("EstimateTokens(%q) = %d, want > 0", tt.input, got)
			}
		})
	}
}

func TestCommandRunnerInterface(t *testing.T) {
	// Verify OSCommandRunner implements CommandRunner
	var _ CommandRunner = (*OSCommandRunner)(nil)
}
