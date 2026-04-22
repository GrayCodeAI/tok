package filter

import (
	"testing"
)

func TestNewSemanticEquivalence(t *testing.T) {
	se := NewSemanticEquivalence()
	if se == nil {
		t.Fatal("expected non-nil SemanticEquivalence")
	}
}

func TestSemanticEquivalence_Check_NoErrors(t *testing.T) {
	se := NewSemanticEquivalence()
	original := "hello world"
	compressed := "hello"
	report := se.Check(original, compressed)

	if report.ErrorPreserved != true {
		t.Error("expected ErrorPreserved true when no errors in original")
	}
	if report.Score <= 0 {
		t.Error("expected positive score")
	}
}

func TestSemanticEquivalence_Check_WithErrors(t *testing.T) {
	se := NewSemanticEquivalence()
	original := "ERROR: something failed"
	compressed := "all good"
	report := se.Check(original, compressed)

	if report.ErrorPreserved {
		t.Error("expected ErrorPreserved false when errors were dropped")
	}
}

func TestSemanticEquivalence_Check_PreservesErrors(t *testing.T) {
	se := NewSemanticEquivalence()
	original := "ERROR: something failed"
	compressed := "ERROR: something failed"
	report := se.Check(original, compressed)

	if !report.ErrorPreserved {
		t.Error("expected ErrorPreserved true when errors preserved")
	}
}

func TestEquivalenceReport_IsGood(t *testing.T) {
	// Good report
	r1 := EquivalenceReport{Score: 0.8, ErrorPreserved: true}
	if !r1.IsGood() {
		t.Error("expected IsGood true for score 0.8 + errors preserved")
	}

	// Bad score
	r2 := EquivalenceReport{Score: 0.5, ErrorPreserved: true}
	if r2.IsGood() {
		t.Error("expected IsGood false for score 0.5")
	}

	// Missing errors
	r3 := EquivalenceReport{Score: 0.9, ErrorPreserved: false}
	if r3.IsGood() {
		t.Error("expected IsGood false when errors not preserved")
	}
}

func TestCheckCriticalNumbers(t *testing.T) {
	original := "exit code 42, line 100"
	compressed := "exit code 42"
	if !checkCriticalNumbers(original, compressed) {
		t.Error("expected numbers preserved when exit code kept")
	}

	compressed2 := "all done"
	if checkCriticalNumbers(original, compressed2) {
		t.Error("expected numbers not preserved when dropped")
	}
}

func TestCheckURLsPreserved(t *testing.T) {
	original := "See https://example.com for details"
	compressed := "See https://example.com for details"
	if !checkURLsPreserved(original, compressed) {
		t.Error("expected URLs preserved")
	}

	compressed2 := "See docs for details"
	if checkURLsPreserved(original, compressed2) {
		t.Error("expected URLs not preserved when dropped")
	}
}

func TestCheckPathsPreserved(t *testing.T) {
	original := "File: /tmp/test.go"
	compressed := "File: /tmp/test.go"
	if !checkPathsPreserved(original, compressed) {
		t.Error("expected paths preserved")
	}
}

func TestCheckExitCodes(t *testing.T) {
	original := "Process exited with exit code 1"
	compressed := "Process exited with exit code 1"
	if !checkExitCodes(original, compressed) {
		t.Error("expected exit codes preserved")
	}

	compressed2 := "Process done"
	if checkExitCodes(original, compressed2) {
		t.Error("expected exit codes not preserved when dropped")
	}
}

func TestComputeEquivalenceScore(t *testing.T) {
	original := "test output with error"
	compressed := "test output with error"
	score := computeEquivalenceScore(original, compressed)
	if score != 1.0 {
		t.Errorf("expected score 1.0 for identical content, got %f", score)
	}
}

func TestExtractCriticalNumbers(t *testing.T) {
	nums := extractCriticalNumbers("line 42, exit 1, port 8080")
	if len(nums) == 0 {
		t.Error("expected some numbers extracted")
	}
}

func TestExtractURLs(t *testing.T) {
	urls := extractURLs("Visit https://example.com and http://test.org")
	if len(urls) != 2 {
		t.Errorf("expected 2 URLs, got %d", len(urls))
	}
}

func TestExtractPaths(t *testing.T) {
	paths := extractPaths("File: /tmp/test.go\nDir: ./src")
	if len(paths) == 0 {
		t.Error("expected some paths extracted")
	}
}
