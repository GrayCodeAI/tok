package agent

import "testing"

func TestAgentManager(t *testing.T) {
	am := NewAgentManager()
	if am == nil {
		t.Fatal("Expected non-nil manager")
	}

	providers := am.ListProviders()
	if len(providers) != 0 {
		t.Error("Expected no providers initially")
	}
}

func TestRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	if cfg.MaxRetries != 3 {
		t.Errorf("Expected 3 retries, got %d", cfg.MaxRetries)
	}
}

func TestStuckLoopDetector(t *testing.T) {
	d := NewStuckLoopDetector(3)

	if d.Record("bash -c ls") {
		t.Error("Should not detect loop on first call")
	}
	if d.Record("bash -c ls") {
		t.Error("Should not detect loop on second call")
	}
	if !d.Record("bash -c ls") {
		t.Error("Should detect loop on third call")
	}

	d.Reset()
	if d.Record("bash -c ls") {
		t.Error("Should not detect loop after reset")
	}
}

func TestContextTracker(t *testing.T) {
	ct := &ContextTracker{MaxTokens: 1000}

	if !ct.AddTokens(500) {
		t.Error("Should be within limit")
	}
	if ct.NeedsCompaction() {
		t.Error("Should not need compaction yet")
	}

	ct.AddTokens(600)
	if !ct.NeedsCompaction() {
		t.Error("Should need compaction after exceeding limit")
	}
}
