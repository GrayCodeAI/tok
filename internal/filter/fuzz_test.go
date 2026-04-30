package filter

import (
	"strings"
	"testing"
)

func FuzzPipelineProcess(f *testing.F) {
	seeds := []string{
		"ERROR: something failed\n  at line 42\n  in file.go",
		"warning: deprecated function used",
		"ok  	github.com/example/pkg	0.123s",
		"PASS",
		"FAIL",
		"",
		strings.Repeat("x", 10000),
		strings.Repeat("a\n", 1000),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		p := NewPipelineCoordinator(PipelineConfig{})
		output, stats, err := p.Process(input)
		if err != nil {
			// Errors are acceptable for malformed input, but output must not be empty
			// when input is non-empty.
			if input != "" && output == "" {
				t.Fatalf("non-empty input produced empty output on error: %v", err)
			}
			return
		}

		// Invariants
		if stats == nil {
			t.Fatal("stats must not be nil")
		}
		if input == "" && output != "" {
			t.Fatalf("empty input produced non-empty output: %q", output)
		}
		// Output should never be larger than a reasonable multiple of input
		// (compression should not dramatically inflate)
		if len(input) > 0 && len(output) > len(input)*10 {
			t.Fatalf("output (%d) dramatically inflated vs input (%d)", len(output), len(input))
		}
		// Stats must be consistent
		if stats.OriginalTokens < 0 || stats.FinalTokens < 0 {
			t.Fatalf("negative token counts: orig=%d final=%d", stats.OriginalTokens, stats.FinalTokens)
		}
	})
}
