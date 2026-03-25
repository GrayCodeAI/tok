package test

import (
	"strings"
	"testing"
)

func TestFilterRspecOutput_AllPass(t *testing.T) {
	input := `{
  "examples": [
    {"full_description": "User validates email", "status": "passed", "file_path": "./spec/models/user_spec.rb", "line_number": 10},
    {"full_description": "User validates name", "status": "passed", "file_path": "./spec/models/user_spec.rb", "line_number": 15}
  ],
  "summary": {"duration": 0.05, "example_count": 2, "failure_count": 0, "pending_count": 0, "errors_outside_of_examples_count": 0}
}`
	result := filterRspecOutput(input)
	if !strings.Contains(result, "✓ RSpec: 2 passed") {
		t.Errorf("expected pass message, got: %s", result)
	}
}

func TestFilterRspecOutput_WithFailures(t *testing.T) {
	bt := "block (2 levels) in <top (required)>"
	input := `{
  "examples": [
    {"full_description": "User validates email", "status": "passed", "file_path": "./spec/models/user_spec.rb", "line_number": 10},
    {"full_description": "User validates name", "status": "failed", "file_path": "./spec/models/user_spec.rb", "line_number": 20,
     "exception": {"class": "RSpec::Expectations::ExpectationNotMetError", "message": "expected name to be present", "backtrace": ["./spec/models/user_spec.rb:20:in '` + bt + `'", "/gems/rspec-core-3.13.0/lib/rspec/core/example.rb:180"]}}
  ],
  "summary": {"duration": 0.08, "example_count": 2, "failure_count": 1, "pending_count": 0, "errors_outside_of_examples_count": 0}
}`
	result := filterRspecOutput(input)
	if !strings.Contains(result, "1 passed, 1 failed") {
		t.Errorf("expected failure count, got: %s", result)
	}
	if !strings.Contains(result, "ExpectationNotMetError") {
		t.Errorf("expected exception class, got: %s", result)
	}
	// Should strip RSpec:: prefix
	if strings.Contains(result, "RSpec::Expectations::ExpectationNotMetError") {
		t.Errorf("should shorten exception class, got: %s", result)
	}
}

func TestFilterRspecOutput_NoExamples(t *testing.T) {
	input := `{"examples": [], "summary": {"duration": 0.0, "example_count": 0, "failure_count": 0, "pending_count": 0, "errors_outside_of_examples_count": 0}}`
	result := filterRspecOutput(input)
	if !strings.Contains(result, "No examples found") {
		t.Errorf("expected 'No examples found', got: %s", result)
	}
}

func TestFilterRspecOutput_Empty(t *testing.T) {
	result := filterRspecOutput("")
	if !strings.Contains(result, "No output") {
		t.Errorf("expected 'No output', got: %s", result)
	}
}

func TestFilterRspecOutput_WithManyFailures(t *testing.T) {
	examples := make([]string, 8)
	for i := 0; i < 8; i++ {
		examples[i] = `{"full_description": "Test ` + string(rune('A'+i)) + `", "status": "failed", "file_path": "./spec/test.rb", "line_number": ` + string(rune('1'+i)) + `, "exception": {"class": "Error", "message": "fail", "backtrace": []}}`
	}
	input := `{"examples": [` + strings.Join(examples, ",") + `], "summary": {"duration": 0.1, "example_count": 8, "failure_count": 8, "pending_count": 0, "errors_outside_of_examples_count": 0}}`
	result := filterRspecOutput(input)
	if !strings.Contains(result, "... +3 more failures") {
		t.Errorf("expected overflow message, got: %s", result)
	}
}

func TestStripRspecNoise(t *testing.T) {
	input := `Running via Spring preloader in process 12345
DEPRECATION WARNING: something old
Coverage report generated
All Files   95.2%

Finished in 0.05 seconds
User validates email
`
	result := stripRspecNoise(input)
	if strings.Contains(result, "Spring preloader") {
		t.Error("should strip Spring line")
	}
	if strings.Contains(result, "DEPRECATION") {
		t.Error("should strip deprecation line")
	}
	if strings.Contains(result, "Coverage report") {
		t.Error("should strip SimpleCov block")
	}
	if strings.Contains(result, "Finished in") {
		t.Error("should strip Finished line")
	}
	if !strings.Contains(result, "User validates email") {
		t.Error("should keep test output")
	}
}

func TestStripRspecNoise_Screenshot(t *testing.T) {
	input := `Test failed
saved screenshot to /tmp/screenshot.png
Next test
`
	result := stripRspecNoise(input)
	if !strings.Contains(result, "[screenshot: /tmp/screenshot.png]") {
		t.Errorf("expected compact screenshot, got: %s", result)
	}
}

func TestFilterRspecText_Passing(t *testing.T) {
	input := `Run options: --seed 12345

# Running:

..

Finished in 0.01234 seconds
2 examples, 0 failures
`
	result := filterRspecText(input)
	if !strings.Contains(result, "2 examples, 0 failures") {
		t.Errorf("expected summary, got: %s", result)
	}
}

func TestFilterRspecText_WithFailures(t *testing.T) {
	input := `Run options: --seed 12345

# Running:

.F

Failures:

  1) User validates name
     Failure/Error: expect(user.name).to eq("John")
       expected: "John"
            got: nil
     # ./spec/models/user_spec.rb:20

Finished in 0.05 seconds
2 examples, 1 failure
`
	result := filterRspecText(input)
	if !strings.Contains(result, "2 examples, 1 failure") {
		t.Errorf("expected summary, got: %s", result)
	}
	if !strings.Contains(result, "User validates name") {
		t.Errorf("expected failure description, got: %s", result)
	}
}

func TestIsNumberedFailure(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"1) User test", true},
		{"10) Some test", true},
		{"  3) Another test", true},
		{"Failures:", false},
		{"not a failure", false},
		{") broken", false},
		{"abc) broken", false},
	}
	for _, tt := range tests {
		got := isNumberedFailure(tt.input)
		if got != tt.want {
			t.Errorf("isNumberedFailure(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsGemBacktrace(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"# /gems/rspec-core-3.13.0/lib/rspec/core/example.rb:180", true},
		{"# lib/rspec/expectations/handler.rb:10", true},
		{"# lib/ruby/3.2.0/set.rb:50", true},
		{"# vendor/bundle/ruby/3.2.0/gems/rails-7.1.0/lib/rails.rb:10", true},
		{"# ./spec/models/user_spec.rb:20", false},
		{"# ./app/models/user.rb:15", false},
	}
	for _, tt := range tests {
		got := isGemBacktrace(tt.input)
		if got != tt.want {
			t.Errorf("isGemBacktrace(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
