package filter

import (
	"strings"
	"testing"
)

func TestCARLFilter_Name(t *testing.T) {
	f := NewCARLFilter()
	if f.Name() != "29_carl" {
		t.Errorf("unexpected name: %s", f.Name())
	}
}

func TestCARLFilter_ModeNone(t *testing.T) {
	f := NewCARLFilter()
	input := "Tool: bash\nResult:\nsome output\n\nTool: bash\nResult:\nother output\n"
	out, saved := f.Apply(input, ModeNone)
	if out != input || saved != 0 {
		t.Error("ModeNone should return unchanged")
	}
}

func TestCARLFilter_NonAgentOutputPassthrough(t *testing.T) {
	f := NewCARLFilter()
	// No tool-call markers → not agent output → pass through
	input := "line 1\nline 2\nline 3\nline 4\nline 5\n"
	out, saved := f.Apply(input, ModeMinimal)
	if out != input || saved != 0 {
		t.Error("non-agent output should pass through unchanged")
	}
}

func TestCARLFilter_DropNonCriticalEntries(t *testing.T) {
	f := NewCARLFilter()
	lines := []string{
		"Tool: bash",
		"Result:",
		"(no output)",
		"",
		"Tool: bash",
		"Result:",
		"total 0",
		"",
		"Tool: bash",
		"Result:",
		"error: cannot find module 'missing-package' in node_modules",
		"npm ERR! code MODULE_NOT_FOUND",
		"",
		"Tool: bash",
		"Result:",
		"(no output)",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Error("expected savings by dropping non-critical empty results")
	}
	if !strings.Contains(out, "MODULE_NOT_FOUND") {
		t.Error("critical error entry must be preserved")
	}
	if !strings.Contains(out, "cannot find module") {
		t.Error("critical error details must be preserved")
	}
}

func TestCARLFilter_PreservesCriticalErrors(t *testing.T) {
	f := NewCARLFilter()
	lines := []string{
		"Tool: bash",
		"Result:",
		"Compiling main package",
		"error[E0308]: mismatched types expected i32 found str",
		"aborting due to previous error",
		"",
		"Tool: bash",
		"Result:",
		"already up to date",
		"",
		"Tool: bash",
		"Result:",
		"test failed: assertion failed left 2 right 3 in test_divide",
		"FAILED: 1 test failed in test suite",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Error("expected savings from dropping no-op entry")
	}
	if !strings.Contains(out, "E0308") {
		t.Error("compile error must be preserved")
	}
	if !strings.Contains(out, "test failed") {
		t.Error("test failure must be preserved")
	}
}

func TestCARLFilter_AggressiveDropsMore(t *testing.T) {
	f := NewCARLFilter()
	// Mix of empty results, low-info results, and one critical error
	lines := []string{
		"Tool: bash",
		"Result:",
		"total 4",
		"",
		"Tool: bash",
		"Result:",
		"health: ok",
		"",
		"Tool: bash",
		"Result:",
		"status: ok running version 1.2.3",
		"",
		"Tool: bash",
		"Result:",
		"200 ok service is healthy",
		"",
		"Tool: bash",
		"Result:",
		"error: permission denied writing to /etc/config",
		"",
	}
	input := strings.Join(lines, "\n")

	_, savedMin := f.Apply(input, ModeMinimal)
	_, savedAgg := f.Apply(input, ModeAggressive)
	if savedAgg < savedMin {
		t.Error("aggressive should save >= minimal")
	}
}

func TestCARLFilter_DiffIsCritical(t *testing.T) {
	f := NewCARLFilter()
	lines := []string{
		"Tool: bash",
		"Result:",
		"diff --git a/src/main.rs b/src/main.rs",
		"--- a/src/main.rs",
		"+++ b/src/main.rs",
		"@@ -10,3 +10,3 @@",
		"-    old_value = 42;",
		"+    new_value = 100;",
		"",
		"Tool: bash",
		"Result:",
		"(no output)",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Error("expected savings by dropping no-op entry")
	}
	if !strings.Contains(out, "diff --git") {
		t.Error("diff entry is critical and must be preserved")
	}
}
