package filter

import (
	"strings"
	"testing"
)

func TestNearDedupFilter_Name(t *testing.T) {
	f := NewNearDedupFilter()
	if f.Name() != "22_near_dedup" {
		t.Errorf("unexpected name: %s", f.Name())
	}
}

func TestNearDedupFilter_ModeNone(t *testing.T) {
	f := NewNearDedupFilter()
	input := "warning: unused variable `x` at src/lib.rs:10\nwarning: unused variable `y` at src/lib.rs:11\n"
	out, saved := f.Apply(input, ModeNone)
	if out != input || saved != 0 {
		t.Error("ModeNone should return input unchanged")
	}
}

func TestNearDedupFilter_ShortInput(t *testing.T) {
	f := NewNearDedupFilter()
	input := "line1\nline2\n"
	out, _ := f.Apply(input, ModeMinimal)
	if out != input {
		t.Error("short input should pass through unchanged")
	}
}

func TestNearDedupFilter_CollapsesNearDuplicates(t *testing.T) {
	f := NewNearDedupFilter()
	// These lines are structurally identical except for file/line number
	lines := []string{
		"warning: unused variable `foo` at src/alpha.rs:10:5",
		"warning: unused variable `bar` at src/alpha.rs:11:5",
		"warning: unused variable `baz` at src/alpha.rs:12:5",
		"warning: unused variable `qux` at src/alpha.rs:13:5",
		"warning: unused variable `quux` at src/alpha.rs:14:5",
		"",
		"error: build failed",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Errorf("expected positive savings on near-duplicate warnings, got %d", saved)
	}
	if !strings.Contains(out, "error: build failed") {
		t.Error("non-duplicate lines must be preserved")
	}
	if !strings.Contains(out, "similar") {
		t.Error("expected [+N similar] annotation in output")
	}
}

func TestNearDedupFilter_PreservesDistinctLines(t *testing.T) {
	f := NewNearDedupFilter()
	lines := []string{
		"error: cannot find value `foobar` in this scope at src/main.rs:5",
		"warning: unused import `std::fmt` at src/lib.rs:3",
		"note: consider removing this import from your codebase",
		"error: mismatched types expected i32 found str at src/main.rs:10",
		"",
		"error[E0308]: build aborted due to 2 previous errors",
	}
	input := strings.Join(lines, "\n")
	out, _ := f.Apply(input, ModeMinimal)

	outLines := strings.Split(strings.TrimSpace(out), "\n")
	// All distinct lines should survive (no near-duplicates here)
	if len(outLines) < 4 {
		t.Errorf("distinct lines should not be collapsed; got %d lines", len(outLines))
	}
}

func TestNearDedupFilter_AggressiveThreshold(t *testing.T) {
	f := NewNearDedupFilter()
	// Build lines that are slightly less similar — aggressive mode should still group them
	lines := []string{
		"[2026-01-01 12:00:00] DEBUG handler=http path=/api/v1/users latency=12ms status=200",
		"[2026-01-01 12:00:01] DEBUG handler=http path=/api/v1/users latency=14ms status=200",
		"[2026-01-01 12:00:02] DEBUG handler=http path=/api/v1/users latency=11ms status=200",
		"[2026-01-01 12:00:03] DEBUG handler=http path=/api/v1/users latency=13ms status=200",
		"[2026-01-01 12:00:04] DEBUG handler=http path=/api/v1/users latency=10ms status=200",
		"",
		"INFO server shutting down gracefully after receiving signal",
	}
	input := strings.Join(lines, "\n")
	_, savedMin := f.Apply(input, ModeMinimal)
	_, savedAgg := f.Apply(input, ModeAggressive)
	if savedAgg < savedMin {
		t.Error("aggressive mode should save >= minimal mode")
	}
}
