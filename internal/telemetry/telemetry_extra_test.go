package telemetry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GrayCodeAI/tok/internal/config"
)

func TestIsEnabled_EnvOverride(t *testing.T) {
	t.Setenv("TOK_TELEMETRY_DISABLED", "1")
	if IsEnabled() {
		t.Error("expected disabled when TOK_TELEMETRY_DISABLED=1")
	}
}

func TestGetConsent_Unknown(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	status := GetConsent()
	if status != ConsentUnknown {
		t.Errorf("expected ConsentUnknown, got %d", status)
	}
}

func TestSetConsent(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	if err := os.MkdirAll(config.DataPath(), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := SetConsent(true); err != nil {
		t.Fatalf("SetConsent(true) failed: %v", err)
	}
	if GetConsent() != ConsentEnabled {
		t.Error("expected ConsentEnabled after SetConsent(true)")
	}

	if err := SetConsent(false); err != nil {
		t.Fatalf("SetConsent(false) failed: %v", err)
	}
	if GetConsent() != ConsentDisabled {
		t.Error("expected ConsentDisabled after SetConsent(false)")
	}
}

func TestSend_Disabled(t *testing.T) {
	t.Setenv("TOK_TELEMETRY_DISABLED", "1")

	data := &TelemetryData{DeviceHash: "test"}
	err := Send(data)
	if err != nil {
		t.Errorf("Send should return nil when disabled, got %v", err)
	}
}

func TestTrackFeatureUsage_Disabled(t *testing.T) {
	t.Setenv("TOK_TELEMETRY_DISABLED", "1")

	err := TrackFeatureUsage("test", nil)
	if err != nil {
		t.Errorf("TrackFeatureUsage should return nil when disabled, got %v", err)
	}
}

func TestGetDeviceHash(t *testing.T) {
	hash1 := getDeviceHash()
	hash2 := getDeviceHash()
	if hash1 != hash2 {
		t.Error("getDeviceHash should return consistent results")
	}
	if len(hash1) != 32 {
		t.Errorf("expected 32 hex chars, got %d", len(hash1))
	}
}

func TestBreakdownKeys(t *testing.T) {
	result := breakdownKeys(nil, 5)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestBuildCommandCategories(t *testing.T) {
	result := buildCommandCategories(nil)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestTopKeys(t *testing.T) {
	counts := map[string]int{
		"a": 3,
		"b": 1,
		"c": 2,
	}
	result := topKeys(counts, 2)
	if len(result) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result))
	}
	// First should be "a (3)" since it has highest count
	if result[0] != "a (3)" {
		t.Errorf("expected first to be 'a (3)', got %q", result[0])
	}
}

func TestTopKeys_Empty(t *testing.T) {
	if topKeys(nil, 5) != nil {
		t.Error("expected nil for nil input")
	}
	if topKeys(map[string]int{}, 5) != nil {
		t.Error("expected nil for empty map")
	}
	if topKeys(map[string]int{"a": 1}, 0) != nil {
		t.Error("expected nil for zero limit")
	}
}

func TestRecentLocalEvents_Empty(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	events, err := RecentLocalEvents(10)
	if err != nil {
		t.Fatalf("RecentLocalEvents failed: %v", err)
	}
	if events != nil {
		t.Error("expected nil for non-existent file")
	}
}

func TestRecentLocalEvents(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	eventsPath := filepath.Join(config.DataPath(), "telemetry", LocalEventsFile)
	if err := os.MkdirAll(filepath.Dir(eventsPath), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	content := `{"feature":"test","timestamp":"2026-04-18T10:00:00Z"}
{"feature":"test2","timestamp":"2026-04-18T11:00:00Z"}`
	if err := os.WriteFile(eventsPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	events, err := RecentLocalEvents(10)
	if err != nil {
		t.Fatalf("RecentLocalEvents failed: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestRecentLocalEvents_Limit(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	eventsPath := filepath.Join(config.DataPath(), "telemetry", LocalEventsFile)
	if err := os.MkdirAll(filepath.Dir(eventsPath), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	content := `{"n":1}
{"n":2}
{"n":3}
{"n":4}
{"n":5}`
	if err := os.WriteFile(eventsPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	events, err := RecentLocalEvents(2)
	if err != nil {
		t.Fatalf("RecentLocalEvents failed: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events with limit, got %d", len(events))
	}
}

func TestGetLocalEventStats_Empty(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	stats, err := GetLocalEventStats()
	if err != nil {
		t.Fatalf("GetLocalEventStats failed: %v", err)
	}
	if stats.TotalEvents != 0 {
		t.Errorf("expected 0 events, got %d", stats.TotalEvents)
	}
}

func TestEventBatcher_AddEvent_Disabled(t *testing.T) {
	t.Setenv("TOK_TELEMETRY_DISABLED", "1")

	b := &EventBatcher{
		events:    make([]map[string]interface{}, 0),
		batchSize: 10,
	}
	b.AddEvent(map[string]interface{}{"type": "test"})
	if len(b.events) != 0 {
		t.Error("expected no events when disabled")
	}
}

func TestEventBatcher_Flush_Empty(t *testing.T) {
	b := &EventBatcher{
		events:    make([]map[string]interface{}, 0),
		batchSize: 10,
	}
	if err := b.Flush(); err != nil {
		t.Errorf("Flush on empty batcher should return nil, got %v", err)
	}
}

func TestGetBatcher(t *testing.T) {
	b1 := GetBatcher()
	b2 := GetBatcher()
	if b1 != b2 {
		t.Error("GetBatcher should return same singleton")
	}
}

func TestFlushBatcher(t *testing.T) {
	err := FlushBatcher()
	if err != nil {
		t.Errorf("FlushBatcher should not error, got %v", err)
	}
}
