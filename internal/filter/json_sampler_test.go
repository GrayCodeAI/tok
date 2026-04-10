package filter

import (
	"strings"
	"testing"
)

func TestJSONSamplerFilter_SamplesDenseJSON(t *testing.T) {
	f := NewJSONSamplerFilter()
	var lines []string
	lines = append(lines, "[")
	for i := 0; i < 40; i++ {
		lines = append(lines, "{\"id\": "+itoa(i)+", \"path\": \"/api/v1/items\", \"value\": \"abcdef\"},")
	}
	lines = append(lines, "]")
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)
	if saved < 0 {
		t.Fatalf("expected non-negative savings, got %d", saved)
	}
	if !strings.Contains(out, "json-sampler") {
		t.Fatalf("expected json-sampler marker")
	}
}
