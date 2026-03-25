package test

import (
	"strings"
	"testing"
)

func TestFilterMinitestOutput_AllPass(t *testing.T) {
	input := `# Running:

...

Finished in 0.05s

4 runs, 8 assertions, 0 failures, 0 errors, 0 skips
`
	result := filterMinitestOutput(input)
	if !strings.Contains(result, "ok rake test:") {
		t.Errorf("expected 'ok rake test:', got: %s", result)
	}
	if !strings.Contains(result, "4 runs") {
		t.Errorf("expected run count, got: %s", result)
	}
	if !strings.Contains(result, "0 failures") {
		t.Errorf("expected failure count, got: %s", result)
	}
}

func TestFilterMinitestOutput_WithFailures(t *testing.T) {
	input := `# Running:

...

Finished in 0.08s

4 runs, 6 assertions, 1 failures, 0 errors, 0 skips

1) Failure:
UserTest#test_validates_name [/test/models/user_test.rb:15]:
Expected name to be present
`
	result := filterMinitestOutput(input)
	if !strings.Contains(result, "1 failures") {
		t.Errorf("expected failure count, got: %s", result)
	}
	if !strings.Contains(result, "UserTest#test_validates_name") {
		t.Errorf("expected failure description, got: %s", result)
	}
}

func TestFilterMinitestOutput_NoTests(t *testing.T) {
	result := filterMinitestOutput("")
	if !strings.Contains(result, "no tests ran") {
		t.Errorf("expected 'no tests ran', got: %s", result)
	}
}

func TestFilterMinitestOutput_WithSkips(t *testing.T) {
	input := `# Running:

...

Finished in 0.03s

10 runs, 15 assertions, 0 failures, 0 errors, 2 skips
`
	result := filterMinitestOutput(input)
	if !strings.Contains(result, "2 skips") {
		t.Errorf("expected skip count, got: %s", result)
	}
}

func TestFilterMinitestOutput_StripAnsi(t *testing.T) {
	input := "\x1b[32m# Running:\x1b[0m\n\n\x1b[32m...\x1b[0m\n\nFinished in 0.05s\n\n4 runs, 8 assertions, 0 failures, 0 errors, 0 skips\n"
	result := filterMinitestOutput(input)
	if strings.Contains(result, "\x1b[") {
		t.Errorf("should strip ANSI codes, got: %s", result)
	}
	if !strings.Contains(result, "ok rake test:") {
		t.Errorf("expected 'ok rake test:', got: %s", result)
	}
}

func TestFilterMinitestOutput_ManyFailures(t *testing.T) {
	var failures []string
	for i := 0; i < 12; i++ {
		failures = append(failures, string(rune('1'+i))+`) Failure:
Test#test_`+string(rune('a'+i))+` [test.rb:`+string(rune('1'+i))+`]:
Expected result`)
	}
	input := `# Running:

Finished in 0.5s

12 runs, 0 assertions, 12 failures, 0 errors, 0 skips

` + strings.Join(failures, "\n\n")

	result := filterMinitestOutput(input)
	if !strings.Contains(result, "... +2 more failures") {
		t.Errorf("expected overflow message, got: %s", result)
	}
}

func TestSelectRunner_NoTest(t *testing.T) {
	tool, args := selectRunner([]string{"db:migrate"})
	if tool != "rake" {
		t.Errorf("expected 'rake', got: %s", tool)
	}
	if len(args) != 1 || args[0] != "db:migrate" {
		t.Errorf("expected args unchanged, got: %v", args)
	}
}

func TestSelectRunner_TestNoFile(t *testing.T) {
	tool, _ := selectRunner([]string{"test"})
	if tool != "rake" {
		t.Errorf("expected 'rake' for bare test, got: %s", tool)
	}
}

func TestSelectRunner_TestWithFile(t *testing.T) {
	tool, _ := selectRunner([]string{"test", "test/models/user_test.rb"})
	if tool != "rails" {
		t.Errorf("expected 'rails' for file arg, got: %s", tool)
	}
}

func TestSelectRunner_TestWithLine(t *testing.T) {
	tool, _ := selectRunner([]string{"test", "test/models/user_test.rb:15"})
	if tool != "rails" {
		t.Errorf("expected 'rails' for line arg, got: %s", tool)
	}
}

func TestSelectRunner_TestWithEnv(t *testing.T) {
	tool, _ := selectRunner([]string{"test", "TEST=test/models/user_test.rb"})
	if tool != "rake" {
		t.Errorf("expected 'rake' for env var, got: %s", tool)
	}
}

func TestLooksLikeTestPath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"test/models/user_test.rb", true},
		{"spec/models/user_spec.rb", true},
		{"test/models/user_test.rb:15", true},
		{"app_test.rb", true},
		{"db:migrate", false},
		{"--verbose", false},
		{"TEST=foo", false},
	}
	for _, tt := range tests {
		got := looksLikeTestPath(tt.input)
		if got != tt.want {
			t.Errorf("looksLikeTestPath(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseMinitestSummary(t *testing.T) {
	runs, assertions, failures, errors, skips := parseMinitestSummary("10 runs, 25 assertions, 2 failures, 1 errors, 3 skips")
	if runs != 10 {
		t.Errorf("runs = %d, want 10", runs)
	}
	if assertions != 25 {
		t.Errorf("assertions = %d, want 25", assertions)
	}
	if failures != 2 {
		t.Errorf("failures = %d, want 2", failures)
	}
	if errors != 1 {
		t.Errorf("errors = %d, want 1", errors)
	}
	if skips != 3 {
		t.Errorf("skips = %d, want 3", skips)
	}
}

func TestParseMinitestSummary_Singular(t *testing.T) {
	runs, _, failures, _, _ := parseMinitestSummary("1 run, 1 assertion, 1 failure, 0 errors, 0 skips")
	if runs != 1 {
		t.Errorf("runs = %d, want 1", runs)
	}
	if failures != 1 {
		t.Errorf("failures = %d, want 1", failures)
	}
}
