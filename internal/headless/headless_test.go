package headless

import "testing"

func TestHeadlessMode(t *testing.T) {
	h := NewHeadlessMode()

	h.Record("git status", 100, 50, 50)
	h.Record("git log", 200, 80, 120)

	metrics := h.Metrics()
	if metrics.CommandsRun != 2 {
		t.Errorf("Expected 2 commands, got %d", metrics.CommandsRun)
	}
	if metrics.TokensSaved != 170 {
		t.Errorf("Expected 170 saved, got %d", metrics.TokensSaved)
	}
}

func TestHeadlessModeReport(t *testing.T) {
	h := NewHeadlessMode()
	h.Record("git status", 100, 50, 50)

	report := h.Report()
	if report == "" {
		t.Error("Expected non-empty report")
	}

	h.SetFormat("text")
	report = h.Report()
	if report == "" {
		t.Error("Expected non-empty text report")
	}
}

func TestHeadlessModeReset(t *testing.T) {
	h := NewHeadlessMode()
	h.Record("cmd", 100, 50, 50)
	h.Reset()

	if h.Metrics().CommandsRun != 0 {
		t.Error("Expected 0 after reset")
	}
}
