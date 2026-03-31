package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	tele := New(true, 50)
	if !tele.IsEnabled() {
		t.Error("expected enabled")
	}
	if tele.EventCount() != 0 {
		t.Error("expected 0 pending events")
	}

	teleDisabled := New(false, 50)
	if teleDisabled.IsEnabled() {
		t.Error("expected disabled")
	}
}

func TestNewDefault(t *testing.T) {
	os.Setenv("TOKMAN_TELEMETRY", "false")
	tele := NewDefault()
	if tele.IsEnabled() {
		t.Error("expected disabled when TOKMAN_TELEMETRY=false")
	}

	os.Setenv("TOKMAN_TELEMETRY", "true")
	tele = NewDefault()
	if !tele.IsEnabled() {
		t.Error("expected enabled when TOKMAN_TELEMETRY=true")
	}

	os.Unsetenv("TOKMAN_TELEMETRY")
}

func TestRecord_Disabled(t *testing.T) {
	tele := New(false, 100)
	tele.Record("test", nil)
	if tele.EventCount() != 0 {
		t.Error("expected no events when disabled")
	}
}

func TestRecord_Enabled(t *testing.T) {
	tele := New(true, 100)
	tele.Record("test_event", map[string]interface{}{"key": "value"})
	if tele.EventCount() != 1 {
		t.Errorf("expected 1 event, got %d", tele.EventCount())
	}
}

func TestRecord_AutoFlush(t *testing.T) {
	tele := New(true, 3)
	tele.Record("e1", nil)
	tele.Record("e2", nil)
	if tele.EventCount() != 2 {
		t.Errorf("expected 2 events, got %d", tele.EventCount())
	}

	// 3rd event triggers flush
	tele.Record("e3", nil)
	if tele.EventCount() != 0 {
		t.Errorf("expected 0 events after auto-flush, got %d", tele.EventCount())
	}
}

func TestFlush(t *testing.T) {
	tele := New(true, 100)
	tele.Record("e1", nil)
	tele.Record("e2", nil)
	tele.Flush()
	if tele.EventCount() != 0 {
		t.Errorf("expected 0 events after flush, got %d", tele.EventCount())
	}
}

func TestFlush_Empty(t *testing.T) {
	tele := New(true, 100)
	tele.Flush() // should not panic
	if tele.EventCount() != 0 {
		t.Error("expected 0 events")
	}
}

func TestSetOutput(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "telemetry.json")

	tele := New(true, 100)
	tele.SetOutput(outputPath)
	tele.Record("test", map[string]interface{}{"key": "value"})
	tele.Flush()

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("failed to parse event: %v", err)
	}
	if event.Type != "test" {
		t.Errorf("expected type 'test', got %q", event.Type)
	}
	if event.Properties["key"] != "value" {
		t.Errorf("expected key='value', got %v", event.Properties["key"])
	}
}

func TestCommandTelemetryEvent(t *testing.T) {
	tele := New(true, 100)
	tele.CommandTelemetryEvent("git status", 500, 120)
	if tele.EventCount() != 1 {
		t.Errorf("expected 1 event, got %d", tele.EventCount())
	}
}

func TestFilterTelemetryEvent(t *testing.T) {
	tele := New(true, 100)
	layers := map[string]int{"entropy": 100, "compaction": 200}
	tele.FilterTelemetryEvent(1000, 700, layers)
	if tele.EventCount() != 1 {
		t.Errorf("expected 1 event, got %d", tele.EventCount())
	}
}

func TestString(t *testing.T) {
	tele := New(true, 100)
	s := tele.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
