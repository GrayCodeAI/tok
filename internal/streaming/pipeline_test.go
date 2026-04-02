// Package streaming provides tests for large file streaming.
package streaming

import (
	"strings"
	"testing"

	"github.com/GrayCodeAI/tokman/internal/filter"
)

// testProcessor is a simple processor for testing
type testProcessor struct{}

func (p *testProcessor) Process(chunk []byte) ([]byte, int, error) {
	// Simple pass-through with minor transformation
	return chunk, 0, nil
}

func (p *testProcessor) Name() string {
	return "test_processor"
}

func TestProcessLargeContent(t *testing.T) {
	// Create large content
	content := strings.Repeat("This is a line of text that will be repeated many times to simulate large content.\n", 100)

	processor := &testProcessor{}
	result, err := ProcessLargeContent(content, processor)
	if err != nil {
		t.Fatalf("ProcessLargeContent failed: %v", err)
	}

	if len(result) == 0 {
		t.Error("result is empty")
	}
}

func TestPipelineWithFilter(t *testing.T) {
	// Just verify FilterProcessor exists and can process
	filterProcessor := &FilterProcessor{engine: filter.NewEngine(filter.ModeMinimal)}

	chunk := []byte("line1\nline2\nline3\n")
	processed, saved, err := filterProcessor.Process(chunk)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Should return processed content
	if len(processed) == 0 {
		t.Error("processed is empty")
	}

	t.Logf("Processed %d bytes, saved %d bytes", len(processed), saved)
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		content  string
		expected int
	}{
		{"hello world", 3},
		{"", 0},
		{strings.Repeat("a", 4000), 1000},
	}

	for _, tc := range tests {
		tokens := EstimateTokens(tc.content)
		// Allow 10% variance
		if tokens < int(float64(tc.expected)*0.9) || tokens > int(float64(tc.expected)*1.1) {
			t.Errorf("expected ~%d tokens for %q, got %d", tc.expected, tc.content, tokens)
		}
	}
}
