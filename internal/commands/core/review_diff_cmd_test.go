package core

import (
	"strings"
	"testing"
)

const reviewDiffTODO = `diff --git a/src/a.go b/src/a.go
--- a/src/a.go
+++ b/src/a.go
@@ -1,1 +1,2 @@
 package src
+// TODO: fix this later
`

const reviewDiffSecret = `diff --git a/src/cfg.go b/src/cfg.go
--- a/src/cfg.go
+++ b/src/cfg.go
@@ -1,1 +1,2 @@
 package src
+const apiKey = "sk_1234567890abcdef1234567890abcdef"
`

const reviewDiffConsoleLog = `diff --git a/src/ui.js b/src/ui.js
--- a/src/ui.js
+++ b/src/ui.js
@@ -1,1 +1,2 @@
 function foo() {}
+console.log("debug");
`

const reviewDiffDebugger = `diff --git a/src/ui.js b/src/ui.js
--- a/src/ui.js
+++ b/src/ui.js
@@ -1,1 +1,2 @@
 function foo() {}
+debugger;
`

var reviewDiffLongLine = `diff --git a/src/long.go b/src/long.go
--- a/src/long.go
+++ b/src/long.go
@@ -1,1 +1,2 @@
 package src
+var x = "` + strings.Repeat("x", 130) + `"
`

const reviewDiffClean = `diff --git a/src/ok.go b/src/ok.go
--- a/src/ok.go
+++ b/src/ok.go
@@ -1,1 +1,2 @@
 package src
+func Foo() int { return 42 }
`

func TestScanDiff_FlagsTODO(t *testing.T) {
	findings := scanDiff(reviewDiffTODO)
	if len(findings) == 0 {
		t.Fatal("expected TODO finding")
	}
	if !strings.Contains(findings[0], "TODO") {
		t.Errorf("finding does not mention TODO: %q", findings[0])
	}
	if !strings.Contains(findings[0], "src/a.go:") {
		t.Errorf("finding missing file:line: %q", findings[0])
	}
}

func TestScanDiff_FlagsHardcodedSecret(t *testing.T) {
	findings := scanDiff(reviewDiffSecret)
	found := false
	for _, f := range findings {
		if strings.Contains(f, "credential") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected credential finding, got %v", findings)
	}
}

func TestScanDiff_FlagsConsoleLog(t *testing.T) {
	findings := scanDiff(reviewDiffConsoleLog)
	found := false
	for _, f := range findings {
		if strings.Contains(f, "console.log") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected console.log finding, got %v", findings)
	}
}

func TestScanDiff_FlagsDebugger(t *testing.T) {
	findings := scanDiff(reviewDiffDebugger)
	found := false
	for _, f := range findings {
		if strings.Contains(f, "debugger") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected debugger finding, got %v", findings)
	}
}

func TestScanDiff_FlagsLongLine(t *testing.T) {
	findings := scanDiff(reviewDiffLongLine)
	found := false
	for _, f := range findings {
		if strings.Contains(f, "120 chars") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected long-line finding, got %v", findings)
	}
}

func TestScanDiff_NoFalsePositives(t *testing.T) {
	findings := scanDiff(reviewDiffClean)
	if len(findings) != 0 {
		t.Errorf("clean diff should have no findings, got %v", findings)
	}
}

func TestScanDiff_LineNumbersAreCorrect(t *testing.T) {
	findings := scanDiff(reviewDiffTODO)
	if len(findings) == 0 {
		t.Fatal("expected findings")
	}
	// TODO is on line 2 of the new file (hunk starts at +1, first line is "package src", TODO is line 2)
	if !strings.Contains(findings[0], ":2 ") {
		t.Errorf("expected line :2, got %q", findings[0])
	}
}

func TestScanDiff_SeverityPrefixes(t *testing.T) {
	// Secret should be bug severity
	findings := scanDiff(reviewDiffSecret)
	sawBug := false
	for _, f := range findings {
		if strings.Contains(f, "credential") && strings.Contains(f, "🔴 bug") {
			sawBug = true
		}
	}
	if !sawBug {
		t.Errorf("hardcoded credential should be marked 🔴 bug; findings=%v", findings)
	}

	// TODO should be risk severity
	findings = scanDiff(reviewDiffTODO)
	sawRisk := false
	for _, f := range findings {
		if strings.Contains(f, "TODO") && strings.Contains(f, "🟡 risk") {
			sawRisk = true
		}
	}
	if !sawRisk {
		t.Errorf("TODO should be marked 🟡 risk; findings=%v", findings)
	}
}
