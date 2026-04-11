package filter

import (
	"strings"
	"testing"
)

func TestCodingAgentCtxFilter_Name(t *testing.T) {
	f := NewCodingAgentContextFilter()
	if f.Name() != "24_coding_agent_ctx" {
		t.Errorf("unexpected name: %s", f.Name())
	}
}

func TestCodingAgentCtxFilter_ModeNone(t *testing.T) {
	f := NewCodingAgentContextFilter()
	input := strings.Repeat("some line of output\n", 20)
	out, saved := f.Apply(input, ModeNone)
	if out != input || saved != 0 {
		t.Error("ModeNone should return input unchanged")
	}
}

func TestCodingAgentCtxFilter_ShortInput(t *testing.T) {
	f := NewCodingAgentContextFilter()
	input := "line1\nline2\nline3"
	out, _ := f.Apply(input, ModeMinimal)
	if out != input {
		t.Error("short input should pass through unchanged")
	}
}

func TestCodingAgentCtxFilter_GitDiff(t *testing.T) {
	f := NewCodingAgentContextFilter()
	lines := []string{
		"diff --git a/src/main.rs b/src/main.rs",
		"index abc123..def456 100644",
		"--- a/src/main.rs",
		"+++ b/src/main.rs",
		"@@ -10,7 +10,7 @@ fn main() {",
		" let x = 1;",
		" let y = 2;",
		"-    println!(\"old\");",
		"+    println!(\"new\");",
		" let z = 3;",
		" }",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeAggressive)

	if saved <= 0 {
		t.Error("expected savings on git diff in aggressive mode")
	}
	if !strings.Contains(out, `println!("new")`) {
		t.Error("added lines (+) must be preserved")
	}
	if !strings.Contains(out, `println!("old")`) {
		t.Error("removed lines (-) must be preserved")
	}
	// Context lines should be dropped in aggressive mode
	if strings.Contains(out, " let x = 1;") {
		t.Error("context lines should be dropped in aggressive diff mode")
	}
}

func TestCodingAgentCtxFilter_CompileLog(t *testing.T) {
	f := NewCodingAgentContextFilter()
	lines := []string{
		"Compiling myproject v0.1.0 (/home/user/project)",
		"warning: unused variable `x` --> src/main.rs:5:9",
		"warning: unused variable `y` --> src/main.rs:6:9",
		"warning: unused variable `z` --> src/main.rs:7:9",
		"error[E0308]: mismatched types",
		" --> src/main.rs:10:5",
		"  |",
		"10 |     foo(\"hello\");",
		"  |         ^^^^^^^ expected i32, found &str",
		"error: aborting due to previous error",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Error("expected savings on compile log")
	}
	if !strings.Contains(out, "error[E0308]") {
		t.Error("error lines must be preserved")
	}
	if !strings.Contains(out, "aborting") {
		t.Error("summary error must be preserved")
	}
}

func TestCodingAgentCtxFilter_TestOutput(t *testing.T) {
	f := NewCodingAgentContextFilter()
	lines := []string{
		"running 5 tests",
		"test test_addition ... ok",
		"test test_subtraction ... ok",
		"test test_division ... FAILED",
		"test test_multiplication ... ok",
		"test test_modulo ... ok",
		"",
		"failures:",
		"",
		"---- test_division stdout ----",
		"thread 'test_division' panicked at assertion failed: result == expected",
		"left:  2",
		"right: 3",
		"",
		"test result: FAILED. 4 passed; 1 failed;",
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Error("expected savings on test output")
	}
	if !strings.Contains(out, "FAILED") {
		t.Error("FAILED lines must be preserved")
	}
	if !strings.Contains(out, "panicked") {
		t.Error("assertion failure must be preserved")
	}
}

func TestCodingAgentCtxFilter_FileRead(t *testing.T) {
	f := NewCodingAgentContextFilter()
	f.headLines = 5
	var lines []string
	for i := 0; i < 50; i++ {
		lines = append(lines, "    some code line number in the file for reading context")
	}
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Error("expected savings on long file read")
	}
	if !strings.Contains(out, "omitted") {
		t.Error("expected omission stub in compressed file read")
	}
}

func TestCodingAgentCtxFilter_BashOutputTruncation(t *testing.T) {
	f := NewCodingAgentContextFilter()
	f.tailLines = 10
	var lines []string
	for i := 0; i < 40; i++ {
		lines = append(lines, "standard bash output line with various content here")
	}
	lines = append(lines, "Process exited with code 0")
	input := strings.Join(lines, "\n")
	out, saved := f.Apply(input, ModeMinimal)

	if saved <= 0 {
		t.Error("expected savings on long bash output")
	}
	if !strings.Contains(out, "Process exited") {
		t.Error("last lines (most recent) must be preserved")
	}
}
