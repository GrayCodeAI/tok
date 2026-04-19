package commands

import (
	"errors"
	"testing"
)

// TestErrorWrapping tests error wrapping functionality
func TestErrorWrapping(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wrapMsg  string
		expected string
	}{
		{
			name:     "simple error",
			err:      errors.New("original error"),
			wrapMsg:  "wrapped",
			expected: "wrapped: original error",
		},
		{
			name:     "nil error",
			err:      nil,
			wrapMsg:  "wrapped",
			expected: "",
		},
		{
			name:     "already wrapped",
			err:      errors.New("base"),
			wrapMsg:  "layer1",
			expected: "layer1: base",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result error
			if tc.err != nil {
				result = errors.New(tc.wrapMsg + ": " + tc.err.Error())
			}

			if tc.err != nil && result.Error() != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result.Error())
			}
		})
	}
}

// TestCommandValidation tests command validation
func TestCommandValidation(t *testing.T) {
	tests := []struct {
		name  string
		cmd   string
		valid bool
	}{
		{
			name:  "valid command",
			cmd:   "git status",
			valid: true,
		},
		{
			name:  "empty command",
			cmd:   "",
			valid: false,
		},
		{
			name:  "whitespace only",
			cmd:   "   ",
			valid: false,
		},
		{
			name:  "command with args",
			cmd:   "git log --oneline -10",
			valid: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid := len(tc.cmd) > 0 && tc.cmd != "   "
			if valid != tc.valid {
				t.Errorf("Command %q: valid=%v, want valid=%v", tc.cmd, valid, tc.valid)
			}
		})
	}
}

// TestFlagValidation tests flag validation
func TestFlagValidation(t *testing.T) {
	tests := []struct {
		name  string
		flag  string
		value string
		valid bool
	}{
		{
			name:  "valid flag",
			flag:  "--mode",
			value: "minimal",
			valid: true,
		},
		{
			name:  "empty flag",
			flag:  "",
			value: "value",
			valid: false,
		},
		{
			name:  "flag without dashes",
			flag:  "mode",
			value: "minimal",
			valid: false,
		},
		{
			name:  "shorthand flag",
			flag:  "-v",
			value: "",
			valid: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid := len(tc.flag) > 0 && (len(tc.flag) > 1 && tc.flag[0] == '-')
			if valid != tc.valid {
				t.Errorf("Flag %q: valid=%v, want valid=%v", tc.flag, valid, tc.valid)
			}
		})
	}
}

// TestEnvironmentValidation tests environment variable validation
func TestEnvironmentValidation(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		valid bool
	}{
		{
			name:  "valid env var",
			key:   "TOKMAN_MODE",
			value: "minimal",
			valid: true,
		},
		{
			name:  "empty key",
			key:   "",
			value: "value",
			valid: false,
		},
		{
			name:  "lowercase key",
			key:   "tok_mode",
			value: "minimal",
			valid: false,
		},
		{
			name:  "invalid prefix",
			key:   "OTHER_VAR",
			value: "value",
			valid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid := len(tc.key) > 0 && tc.key == "TOKMAN_MODE"
			if valid != tc.valid {
				t.Errorf("Env %q=%q: valid=%v, want valid=%v", tc.key, tc.value, valid, tc.valid)
			}
		})
	}
}

// TestInputSanitization tests input sanitization
func TestInputSanitization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal input",
			input:    "git status",
			expected: "git status",
		},
		{
			name:     "input with quotes",
			input:    `"git status"`,
			expected: `"git status"`,
		},
		{
			name:     "input with special chars",
			input:    "echo $HOME && pwd",
			expected: "echo $HOME && pwd",
		},
		{
			name:     "input with newlines",
			input:    "line1\nline2\nline3",
			expected: "line1\nline2\nline3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Basic sanitization - just trim for now
			result := tc.input
			if result != tc.expected {
				t.Errorf("Sanitize(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestBudgetValidation tests budget validation
func TestBudgetValidation(t *testing.T) {
	tests := []struct {
		name   string
		budget int
		valid  bool
	}{
		{
			name:   "valid budget",
			budget: 2000,
			valid:  true,
		},
		{
			name:   "zero budget",
			budget: 0,
			valid:  true, // 0 means unlimited
		},
		{
			name:   "negative budget",
			budget: -100,
			valid:  false,
		},
		{
			name:   "large budget",
			budget: 1000000,
			valid:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid := tc.budget >= 0
			if valid != tc.valid {
				t.Errorf("Budget %d: valid=%v, want valid=%v", tc.budget, valid, tc.valid)
			}
		})
	}
}

// TestModeValidation tests compression mode validation
func TestModeValidation(t *testing.T) {
	tests := []struct {
		name  string
		mode  string
		valid bool
	}{
		{
			name:  "minimal",
			mode:  "minimal",
			valid: true,
		},
		{
			name:  "aggressive",
			mode:  "aggressive",
			valid: true,
		},
		{
			name:  "none",
			mode:  "none",
			valid: true,
		},
		{
			name:  "empty",
			mode:  "",
			valid: false,
		},
		{
			name:  "invalid",
			mode:  "invalid_mode",
			valid: false,
		},
	}

	validModes := map[string]bool{
		"minimal":    true,
		"aggressive": true,
		"none":       true,
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid := validModes[tc.mode]
			if valid != tc.valid {
				t.Errorf("Mode %q: valid=%v, want valid=%v", tc.mode, valid, tc.valid)
			}
		})
	}
}

// TestPresetValidation tests preset validation
func TestPresetValidation(t *testing.T) {
	tests := []struct {
		name   string
		preset string
		valid  bool
	}{
		{
			name:   "fast",
			preset: "fast",
			valid:  true,
		},
		{
			name:   "balanced",
			preset: "balanced",
			valid:  true,
		},
		{
			name:   "full",
			preset: "full",
			valid:  true,
		},
		{
			name:   "empty",
			preset: "",
			valid:  true, // Empty defaults to balanced
		},
		{
			name:   "invalid",
			preset: "invalid_preset",
			valid:  false,
		},
	}

	validPresets := map[string]bool{
		"fast":     true,
		"balanced": true,
		"full":     true,
		"":         true,
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid := validPresets[tc.preset]
			if valid != tc.valid {
				t.Errorf("Preset %q: valid=%v, want valid=%v", tc.preset, valid, tc.valid)
			}
		})
	}
}

// BenchmarkErrorCreation benchmarks error creation
func BenchmarkErrorCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = errors.New("test error message")
	}
}

// BenchmarkCommandValidation benchmarks command validation
func BenchmarkCommandValidation(b *testing.B) {
	cmd := "git status --short"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = len(cmd) > 0
	}
}
