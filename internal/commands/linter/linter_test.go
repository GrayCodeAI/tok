package linter

import (
	"strings"
	"testing"
)

func TestStripPmPrefix(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"no prefix", []string{"eslint", "src/"}, []string{"eslint", "src/"}},
		{"npx", []string{"npx", "eslint", "src/"}, []string{"eslint", "src/"}},
		{"bunx", []string{"bunx", "eslint"}, []string{"eslint"}},
		{"pnpm exec", []string{"pnpm", "exec", "ruff", "check"}, []string{"ruff", "check"}},
		{"all pm", []string{"npx"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripPmPrefix(tt.args)
			if len(got) != len(tt.want) {
				t.Errorf("stripPmPrefix(%v) = %v, want %v", tt.args, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("stripPmPrefix(%v)[%d] = %q, want %q", tt.args, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestDetectLinter(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		want     string
		explicit bool
	}{
		{"no args", []string{}, "eslint", false},
		{"path", []string{"src/"}, "eslint", false},
		{"flag", []string{"--fix"}, "eslint", false},
		{"explicit ruff", []string{"ruff", "check"}, "ruff", true},
		{"explicit eslint", []string{"eslint", "src/"}, "eslint", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, explicit := detectLinter(tt.args)
			if got != tt.want {
				t.Errorf("detectLinter(%v) = %q, want %q", tt.args, got, tt.want)
			}
			if explicit != tt.explicit {
				t.Errorf("detectLinter(%v) explicit = %v, want %v", tt.args, explicit, tt.explicit)
			}
		})
	}
}

func TestFilterEslintJSON_NoIssues(t *testing.T) {
	input := `[{"filePath":"src/index.ts","messages":[],"errorCount":0,"warningCount":0}]`
	got := filterEslintJSON(input)
	if !strings.Contains(got, "No issues") {
		t.Errorf("expected 'No issues' in output, got %q", got)
	}
}

func TestFilterEslintJSON_WithIssues(t *testing.T) {
	input := `[{"filePath":"src/index.ts","messages":[{"ruleId":"no-unused-vars","severity":1,"message":"'x' is defined but never used.","line":1,"column":5}],"errorCount":0,"warningCount":1}]`
	got := filterEslintJSON(input)
	if !strings.Contains(got, "warning") {
		t.Errorf("expected 'warning' in output, got %q", got)
	}
}

func TestFilterEslintJSON_InvalidJSON(t *testing.T) {
	got := filterEslintJSON("not json")
	if !strings.Contains(got, "parse failed") {
		t.Errorf("expected 'parse failed' in output, got %q", got)
	}
}

func TestFilterPylintJSON_NoIssues(t *testing.T) {
	got := filterPylintJSON(`[]`)
	if !strings.Contains(got, "No issues") {
		t.Errorf("expected 'No issues' in output, got %q", got)
	}
}

func TestFilterPylintJSON_WithIssues(t *testing.T) {
	input := `[{"type":"convention","module":"main","obj":"","line":1,"column":0,"path":"main.py","symbol":"missing-module-docstring","message":"Missing module docstring","message-id":"C0114"}]`
	got := filterPylintJSON(input)
	if !strings.Contains(got, "issues") {
		t.Errorf("expected 'issues' in output, got %q", got)
	}
}

func TestFilterLintUltraCompact(t *testing.T) {
	tests := []struct {
		name   string
		linter string
		stdout string
		want   string
	}{
		{"eslint ok", "eslint", `[{"messages":[],"errorCount":0,"warningCount":0}]`, "ok"},
		{"ruff ok", "ruff", `[]`, "ok"},
		{"pylint ok", "pylint", `[]`, "ok"},
		{"mypy ok", "mypy", "Success: no issues", "ok"},
		{"mypy errors", "mypy", "src/a.py:1: error: foo", "errors"},
		{"generic ok", "unknown", "all good", "ok"},
		{"generic errors", "unknown", "error: something failed", "errors"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterLintUltraCompact(tt.stdout, "", tt.linter)
			if !strings.Contains(got, tt.want) {
				t.Errorf("filterLintUltraCompact(%q) = %q, expected to contain %q", tt.linter, got, tt.want)
			}
		})
	}
}

func TestFilterGenericLint_NoIssues(t *testing.T) {
	got := filterGenericLint("all clean")
	if !strings.Contains(got, "No issues") {
		t.Errorf("expected 'No issues' in output, got %q", got)
	}
}

func TestFilterGenericLint_WithIssues(t *testing.T) {
	input := "warning: unused import\nerror: undefined variable"
	got := filterGenericLint(input)
	if !strings.Contains(got, "errors") {
		t.Errorf("expected 'errors' in output, got %q", got)
	}
}

func TestCompactPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"short", "src/index.ts", "src/index.ts"},
		{"long", "/home/user/project/src/components/Button/index.tsx", ".../Button/index.tsx"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compactPath(tt.input)
			if len(got) > len(tt.input) {
				t.Errorf("compactPath(%q) = %q, should not be longer than input", tt.input, got)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	input := "\x1b[31merror\x1b[0m"
	got := stripANSI(input)
	if got != "error" {
		t.Errorf("stripANSI(%q) = %q, want %q", input, got, "error")
	}
}
