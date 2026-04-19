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
		_, _ = p.Process(input)
	})
}
