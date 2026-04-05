package historybloat_test

import (
	"strings"
	"testing"

	"github.com/GrayCodeAI/tokman/internal/historybloat"
)

func TestNewHistoryAnalyzer(t *testing.T) {
	a := historybloat.NewHistoryAnalyzer()
	if a == nil {
		t.Fatal("NewHistoryAnalyzer returned nil")
	}
}

func TestNewHistoryAnalyzerWithThreshold(t *testing.T) {
	a := historybloat.NewHistoryAnalyzerWithThreshold(0.5)
	if a == nil {
		t.Fatal("NewHistoryAnalyzerWithThreshold returned nil")
	}
}

func TestAnalyze_CleanInput(t *testing.T) {
	a := historybloat.NewHistoryAnalyzer()
	report := a.Analyze("hello world no repeated lines")
	if report == nil {
		t.Fatal("Analyze returned nil")
	}
	if report.TotalTokens <= 0 {
		t.Error("expected positive token count")
	}
}

func TestAnalyze_RedundantInput(t *testing.T) {
	a := historybloat.NewHistoryAnalyzer()
	repeatedLine := "This is a very long line that repeats many times in conversation history\n"
	input := strings.Repeat(repeatedLine, 20)
	report := a.Analyze(input)
	if report.RedundantEntries <= 0 {
		t.Errorf("expected redundancy in repeated input, got %d", report.RedundantEntries)
	}
}

func TestAnalyze_CompilationErrors(t *testing.T) {
	a := historybloat.NewHistoryAnalyzer()
	input := "error[E0001]: something went wrong\nerror[E0001]: same error again\n"
	report := a.Analyze(input)
	if report == nil {
		t.Fatal("Analyze returned nil")
	}
}

func TestAnalyze_EmptyInput(t *testing.T) {
	a := historybloat.NewHistoryAnalyzer()
	report := a.Analyze("")
	if report == nil {
		t.Fatal("Analyze returned nil for empty input")
	}
	if report.TotalTokens != 0 {
		t.Errorf("expected 0 tokens for empty input, got %d", report.TotalTokens)
	}
}

func TestAnalyze_RecommendationNotEmpty(t *testing.T) {
	a := historybloat.NewHistoryAnalyzerWithThreshold(0.1)
	input := strings.Repeat("line\n", 50)
	report := a.Analyze(input)
	if report.IsBloated {
		t.Log("input was correctly identified as bloated")
	}
	if report.Recommendation == "" && report.IsBloated {
		t.Error("bloated report should have a recommendation")
	}
}
