package filter

import (
	"strings"
	"testing"
)

func TestNewGistFilter(t *testing.T) {
	f := NewGistFilter()
	if f == nil {
		t.Fatal("NewGistFilter returned nil")
	}
	if f.Name() != "gist" {
		t.Errorf("Name() = %q, want 'gist'", f.Name())
	}
}

func TestGistFilter_Apply_None(t *testing.T) {
	f := NewGistFilter()
	input := "some test content here"
	output, saved := f.Apply(input, ModeNone)
	if output != input {
		t.Error("ModeNone should not modify input")
	}
	if saved != 0 {
		t.Errorf("ModeNone should save 0, got %d", saved)
	}
}

func TestGistFilter_Apply_Minimal(t *testing.T) {
	f := NewGistFilter()
	input := `Traceback (most recent call last):
  File "test.py", line 10, in main
    result = compute()
  File "test.py", line 5, in compute
    return x + y
Error: something went wrong
Some additional context here that provides more information about the error.`
	output, saved := f.Apply(input, ModeMinimal)
	if output == "" {
		t.Error("output should not be empty")
	}
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
}

func TestGistFilter_Apply_Aggressive(t *testing.T) {
	f := NewGistFilter()
	input := `Traceback (most recent call last):
  File "test.py", line 10, in main
    result = compute()
  File "test.py", line 5, in compute
    return x + y
Error: something went wrong
More context line 1
More context line 2
More context line 3
More context line 4
More context line 5
More context line 6
More context line 7
More context line 8
More context line 9
More context line 10
More context line 11
More context line 12
More context line 13
More context line 14
More context line 15
More context line 16
More context line 17
More context line 18
More context line 19
More context line 20
More context line 21`
	output, saved := f.Apply(input, ModeAggressive)
	if saved < 0 {
		t.Errorf("saved should be >= 0, got %d", saved)
	}
	_ = output
}

func TestGistFilter_StackTraceDetection(t *testing.T) {
	f := NewGistFilter()

	tests := []struct {
		line string
		want bool
	}{
		{"Traceback (most recent call last):", true},
		{"stack traceback:", true},
		{"goroutine 1 [running]:", true},
		{"some normal line", false},
	}

	for _, tt := range tests {
		t.Run(tt.line[:min(20, len(tt.line))], func(t *testing.T) {
			got := f.isStackTraceStart(tt.line)
			if got != tt.want {
				t.Errorf("isStackTraceStart(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestGistFilter_ImportBlockDetection(t *testing.T) {
	f := NewGistFilter()

	tests := []struct {
		line string
		want bool
	}{
		{"import (", true},
		{"import \"fmt\"", true},
		{"from os import path", true},
		{"require('lodash')", true},
		{"some code", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := f.isImportBlockStart(tt.line)
			if got != tt.want {
				t.Errorf("isImportBlockStart(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestGistFilter_TestOutputDetection(t *testing.T) {
	f := NewGistFilter()

	tests := []struct {
		line string
		want bool
	}{
		{"=== RUN   TestFoo", true},
		{"test session starts", true},
		{"PASS ok github.com/example 0.5s", true},
		{"some output", false},
	}

	for _, tt := range tests {
		t.Run(tt.line[:min(20, len(tt.line))], func(t *testing.T) {
			got := f.isTestOutputStart(tt.line)
			if got != tt.want {
				t.Errorf("isTestOutputStart(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestGistFilter_GitDiffDetection(t *testing.T) {
	f := NewGistFilter()

	tests := []struct {
		line string
		want bool
	}{
		{"diff --git a/file.go b/file.go", true},
		{"index 1234567..abcdef", true},
		{"--- a/file.go", true},
		{"some diff content", false},
	}

	for _, tt := range tests {
		t.Run(tt.line[:min(20, len(tt.line))], func(t *testing.T) {
			got := f.isGitDiffStart(tt.line)
			if got != tt.want {
				t.Errorf("isGitDiffStart(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestGistFilter_SetMaxChunkSize(t *testing.T) {
	f := NewGistFilter()
	f.SetMaxChunkSize(1000)
	if f.maxChunkSize != 1000 {
		t.Errorf("maxChunkSize = %d, want 1000", f.maxChunkSize)
	}
}

func TestGistFilter_GistForType(t *testing.T) {
	f := NewGistFilter()

	tests := []struct {
		blockType string
		want      string
	}{
		{"stack_trace", "[stack trace]"},
		{"import_block", "[imports]"},
		{"test_output", "[test results]"},
		{"build_log", "[build output]"},
		{"git_diff", "[diff]"},
		{"json_block", "[json]"},
		{"unknown", "[...]"},
	}

	for _, tt := range tests {
		t.Run(tt.blockType, func(t *testing.T) {
			got := f.gistForType(tt.blockType)
			if got != tt.want {
				t.Errorf("gistForType(%q) = %q, want %q", tt.blockType, got, tt.want)
			}
		})
	}
}

func TestGistFilter_BuildLogDetection(t *testing.T) {
	f := NewGistFilter()

	tests := []struct {
		line string
		want bool
	}{
		{"Building project...", true},
		{"Compiling module", true},
		{"[BUILD] step 1", true},
		{"running tests", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := f.isBuildLogStart(tt.line)
			if got != tt.want {
				t.Errorf("isBuildLogStart(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestGistFilter_JSONBlockDetection(t *testing.T) {
	f := NewGistFilter()

	tests := []struct {
		line string
		want bool
	}{
		{"{", true},
		{"  {", true},
		{"[", true},
		{"  [", true},
		{"some text", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := f.isJSONBlockStart(tt.line)
			if got != tt.want {
				t.Errorf("isJSONBlockStart(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestGistFilter_BlockEnd(t *testing.T) {
	f := NewGistFilter()

	tests := []struct {
		line      string
		blockType string
		want      bool
	}{
		{")", "import_block", true},
		{"}", "json_block", true},
		{"]", "json_block", true},
		{"", "stack_trace", true},
		{"some content", "stack_trace", false},
		{"", "test_output", true},
	}

	for _, tt := range tests {
		t.Run(tt.blockType+"_"+tt.line, func(t *testing.T) {
			got := f.isBlockEnd(tt.line, tt.blockType)
			if got != tt.want {
				t.Errorf("isBlockEnd(%q, %q) = %v, want %v", tt.line, tt.blockType, got, tt.want)
			}
		})
	}
}

func BenchmarkGistFilter_Apply(b *testing.B) {
	f := NewGistFilter()
	input := strings.Repeat("Traceback (most recent call last):\n  File \"test.py\", line 10\nError: something went wrong\n", 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Apply(input, ModeAggressive)
	}
}
