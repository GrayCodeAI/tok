package filter

import (
	"strings"
	"testing"
)

func TestMarginalInfoGainFilter_Name(t *testing.T) {
	f := NewMarginalInfoGainFilter()
	if f.Name() != "21_marginal_info_gain" {
		t.Errorf("unexpected name: %s", f.Name())
	}
}

func TestMarginalInfoGainFilter_ModeNone(t *testing.T) {
	f := NewMarginalInfoGainFilter()
	input := strings.Repeat("hello world foo bar baz\n", 20)
	out, saved := f.Apply(input, ModeNone)
	if out != input || saved != 0 {
		t.Error("ModeNone should return input unchanged")
	}
}

func TestMarginalInfoGainFilter_ShortInput(t *testing.T) {
	f := NewMarginalInfoGainFilter()
	input := "line1\nline2\nline3"
	out, _ := f.Apply(input, ModeMinimal)
	if out != input {
		t.Error("short input should pass through unchanged")
	}
}

func TestMarginalInfoGainFilter_ReducesRedundantLines(t *testing.T) {
	f := NewMarginalInfoGainFilter()
	// Build input with many near-identical lines that carry no new information
	var lines []string
	lines = append(lines, "ERROR: database connection failed at host:5432")
	for i := 0; i < 30; i++ {
		lines = append(lines, "INFO: retrying connection to database host port 5432 attempt number")
	}
	lines = append(lines, "FATAL: max retries exceeded giving up on database host")

	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Error("expected positive token savings on redundant input")
	}
	outLines := strings.Split(out, "\n")
	if len(outLines) >= len(lines) {
		t.Errorf("expected fewer lines: got %d, want < %d", len(outLines), len(lines))
	}
	// Error and fatal lines must be preserved
	if !strings.Contains(out, "ERROR:") {
		t.Error("error line must be preserved")
	}
	if !strings.Contains(out, "FATAL:") {
		t.Error("fatal line must be preserved")
	}
}

func TestMarginalInfoGainFilter_AggressiveMoreReduction(t *testing.T) {
	f := NewMarginalInfoGainFilter()
	var lines []string
	for i := 0; i < 40; i++ {
		lines = append(lines, "warning: unused import in file src/module foo bar baz qux")
	}
	lines = append(lines, "error: build failed with 40 warnings")
	input := strings.Join(lines, "\n")

	_, savedMin := f.Apply(input, ModeMinimal)
	_, savedAgg := f.Apply(input, ModeAggressive)
	if savedAgg < savedMin {
		t.Error("aggressive mode should save at least as many tokens as minimal")
	}
}

func TestMarginalInfoGainFilter_AnchorsPreserved(t *testing.T) {
	f := NewMarginalInfoGainFilter()
	var lines []string
	lines = append(lines, "FIRST LINE UNIQUE HEADER unique_start_marker")
	for i := 0; i < 20; i++ {
		lines = append(lines, "middle line with same content repeated over and over again")
	}
	lines = append(lines, "LAST LINE UNIQUE FOOTER unique_end_marker")

	input := strings.Join(lines, "\n")
	out, _ := f.Apply(input, ModeMinimal)

	if !strings.Contains(out, "FIRST LINE") {
		t.Error("first line anchor must be preserved")
	}
	if !strings.Contains(out, "LAST LINE") {
		t.Error("last line anchor must be preserved")
	}
}
