package waste

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestWhitespaceBloatDetector(t *testing.T) {
	d := NewWhitespaceBloatDetector()

	input := "hello   \nworld\n\n\n\nfoo\t\t\nbar"
	findings := d.Detect(input)
	if len(findings) == 0 {
		t.Error("Expected whitespace findings")
	}
}

func TestFillerDetector(t *testing.T) {
	d := NewFillerDetector()

	input := "Here is the code you requested.\nNote that this is important.\nLet me explain.\nOf course, this works."
	findings := d.Detect(input)
	if len(findings) == 0 {
		t.Error("Expected filler findings")
	}
}

func TestRedundantInstructionDetector(t *testing.T) {
	d := NewRedundantInstructionDetector()

	input := "The quick brown fox jumps over the lazy dog\nSome other line\nThe quick brown fox jumps over the lazy dog"
	findings := d.Detect(input)
	if len(findings) == 0 {
		t.Error("Expected redundant findings")
	}
}

func TestOutputUtilizationTracker(t *testing.T) {
	tracker := NewOutputUtilizationTracker()

	finding := tracker.Track(1000, 100)
	if finding.Type != WasteOutput {
		t.Error("Expected low utilization finding")
	}

	finding = tracker.Track(100, 80)
	if finding.Type == WasteOutput {
		t.Error("Expected no finding for good utilization")
	}
}

func TestWasteScoreCalculator(t *testing.T) {
	c := NewWasteScoreCalculator()

	findings := []WasteFinding{
		{Savings: 50},
		{Savings: 30},
	}
	score := c.Calculate(1000, findings)
	if score != 8.0 {
		t.Errorf("Expected score 8.0, got %.2f", score)
	}

	score = c.Calculate(0, findings)
	if score != 0 {
		t.Errorf("Expected score 0 for zero tokens, got %.2f", score)
	}
}

func TestWasteAnalyzer(t *testing.T) {
	a := NewWasteAnalyzer()

	input := "Here is the code   \nNote that this works   \nThe quick brown fox jumps over the lazy dog\nThe quick brown fox jumps over the lazy dog\nfix this bug"
	report := a.Analyze(input)

	if report.TotalTokens == 0 {
		t.Error("Expected non-zero tokens")
	}
	if report.WasteScore < 0 || report.WasteScore > 100 {
		t.Errorf("Invalid waste score: %.2f", report.WasteScore)
	}
	if len(report.Recommendations) == 0 {
		t.Error("Expected recommendations")
	}
}

func TestWasteStore(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skip("SQLite not available")
	}
	defer db.Close()

	store := NewWasteStore(db)
	if err := store.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}

	report := &WasteReport{
		TotalTokens: 1000,
		WasteTokens: 100,
		WasteScore:  10.0,
		Findings:    []WasteFinding{{Type: WasteFiller}},
	}

	if err := store.Record("git status", report); err != nil {
		t.Fatalf("Record error: %v", err)
	}

	records, err := store.GetTrend(7)
	if err != nil {
		t.Fatalf("GetTrend error: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}

	avg, err := store.GetAverageScore(7)
	if err != nil {
		t.Fatalf("GetAverageScore error: %v", err)
	}
	if avg != 10.0 {
		t.Errorf("Expected avg 10.0, got %.2f", avg)
	}
}
