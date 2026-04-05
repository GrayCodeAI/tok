package doctor

import (
	"strings"
	"testing"
)

func TestNewDoctor(t *testing.T) {
	d := NewDoctor()
	if d == nil {
		t.Fatal("NewDoctor returned nil")
	}
}

func TestRunChecks_ReturnsAllChecks(t *testing.T) {
	d := NewDoctor()
	results := d.RunChecks()

	expected := []string{"config", "database", "filters", "hooks"}

	for i, want := range expected {
		if i >= len(results) {
			t.Fatalf("expected %d checks, got %d", len(expected), len(results))
		}
		if results[i].Check != want {
			t.Errorf("Check[%d] = %q, want %q", i, results[i].Check, want)
		}
	}
}

func TestRunChecks_AllOK(t *testing.T) {
	d := NewDoctor()
	results := d.RunChecks()
	for _, r := range results {
		if r.Status != "ok" {
			t.Errorf("Check %q status = %q, want ok", r.Check, r.Status)
		}
	}
}

func TestFormat(t *testing.T) {
	d := NewDoctor()
	results := d.RunChecks()
	output := d.Format(results)

	for _, want := range []string{"TokMan Doctor", "config", "database", "filters", "hooks"} {
		if !strings.Contains(output, want) {
			t.Errorf("Format output missing %q", want)
		}
	}
	if !strings.Contains(output, "ok") {
		t.Error("Format output should contain 'ok'")
	}
}

func TestFormat_FailureStatus(t *testing.T) {
	d := NewDoctor()
	results := []Result{
		{Check: "test", Status: "fail", Message: "broken"},
	}
	output := d.Format(results)
	if !strings.Contains(output, "fail") {
		t.Errorf("Format should show 'fail' for failed checks, got: %q", output)
	}
}

func TestFormat_Empty(t *testing.T) {
	d := NewDoctor()
	output := d.Format([]Result{})
	if !strings.Contains(output, "TokMan Doctor") {
		t.Error("Format should include header even for empty results")
	}
}
