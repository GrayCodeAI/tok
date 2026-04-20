package tui

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

// TestLiveEndToEndWriteToEvent drives the full path: a real Tracker
// writes a CommandRecord via Record(), and a trackingLiveSource
// listening via SubscribeCommands observes the event. The whole
// journey must complete in under 500ms — anything slower and the TUI
// "live" branding isn't credible.
func TestLiveEndToEndWriteToEvent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TOK_DATABASE_PATH", filepath.Join(dir, "tracking.db"))

	tracker, err := tracking.NewTracker(filepath.Join(dir, "tracking.db"))
	if err != nil {
		t.Skipf("tracker init failed — likely a sandbox without sqlite: %v", err)
	}
	t.Cleanup(func() { _ = tracker.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	source := &trackingLiveSource{
		dbPath:        "",       // skip fsnotify; subscribe path only
		fallbackEvery: time.Hour, // no tick noise in this test
	}
	events := source.Start(ctx)
	time.Sleep(20 * time.Millisecond) // let the goroutine register the sub

	start := time.Now()
	writeErr := tracker.Record(&tracking.CommandRecord{
		Command:        "echo e2e",
		OriginalTokens: 100,
		FilteredTokens: 20,
		SavedTokens:    80,
		ExecTimeMs:     5,
	})
	if writeErr != nil {
		t.Fatalf("Record failed: %v", writeErr)
	}

	select {
	case ev := <-events:
		latency := time.Since(start)
		if ev.Source != "subscribe" {
			t.Errorf("got source %q, want subscribe", ev.Source)
		}
		if ev.Record == nil || ev.Record.Command != "echo e2e" {
			t.Errorf("got record %+v, want 'echo e2e'", ev.Record)
		}
		if latency > 500*time.Millisecond {
			t.Errorf("live latency %v exceeds 500ms target", latency)
		}
		t.Logf("live event delivered in %v", latency)
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for live event (1s) — the subscribe path is broken")
	}
}
