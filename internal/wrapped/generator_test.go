package wrapped

import "testing"

func TestWrappedGenerator(t *testing.T) {
	gen := NewWrappedGenerator()

	stats := &WrappedStats{
		TotalTokens:       100000,
		TotalSaved:        60000,
		SavingsPercentage: 60.0,
		TopModel:          "gpt-4o",
		TopCommand:        "git status",
		TotalCommands:     500,
		TotalSessions:     50,
	}

	output := gen.Generate(stats)
	if output == "" {
		t.Error("Expected non-empty output")
	}
}

func TestWrappedSVG(t *testing.T) {
	gen := NewWrappedGenerator()

	stats := &WrappedStats{
		TotalTokens:   100000,
		TotalSaved:    60000,
		TopModel:      "gpt-4o",
		TotalCommands: 500,
		TotalSessions: 50,
	}

	svg := gen.GenerateSVG(stats)
	if svg == "" {
		t.Error("Expected non-empty SVG")
	}
}
