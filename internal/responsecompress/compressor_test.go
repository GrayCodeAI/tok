package responsecompress

import "testing"

func TestResponseCompressor(t *testing.T) {
	c := NewResponseCompressor()

	input := "Here is the code   \nNote that this works   \n\n\nThe output\n\n\n"
	compressed := c.Compress(input)

	if len(compressed) >= len(input) {
		t.Error("Expected shorter output")
	}
}

func TestStripWhitespace(t *testing.T) {
	input := "hello   \n\n\nworld\n\n"
	output := stripWhitespace(input)
	if output != "hello\n\nworld" {
		t.Errorf("Expected 'hello\\n\\nworld', got %q", output)
	}
}

func TestTruncateToTokens(t *testing.T) {
	long := string(make([]byte, 1000))
	truncated := truncateToTokens(long, 100)
	if len(truncated) >= len(long) {
		t.Error("Expected truncated output")
	}

	short := "hello"
	result := truncateToTokens(short, 100)
	if result != short {
		t.Error("Expected unchanged output for short input")
	}
}

func TestCalculateResponseMetrics(t *testing.T) {
	metrics := CalculateResponseMetrics("long original text here", "short")
	if metrics.OriginalTokens == 0 {
		t.Error("Expected non-zero original tokens")
	}
	if metrics.Savings <= 0 {
		t.Error("Expected positive savings")
	}
}
