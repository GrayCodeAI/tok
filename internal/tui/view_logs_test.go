package tui

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func seededLogRing() *ringHandler {
	ring := NewRingHandler(16, slog.LevelDebug, nil)
	logger := slog.New(ring)
	logger.Info("snapshot loaded", "count", 42)
	logger.Warn("parse failure", "command", "gh pr view")
	logger.Error("db locked", "op", "insert")
	logger.Debug("trace", "step", 1)
	return ring
}

func TestLogsSectionRenders(t *testing.T) {
	ring := seededLogRing()
	ctx := fixtureDashCtxWithTrends()
	ctx.Logs = ring

	s := newLogsSection()
	view := s.View(ctx)
	// Default level is Info — debug entry should be hidden, error+warn+info visible.
	if strings.Contains(view, "trace") {
		t.Fatalf("expected debug to be hidden at default level:\n%s", view)
	}
	for _, want := range []string{"snapshot loaded", "parse failure", "db locked"} {
		if !strings.Contains(view, want) {
			t.Fatalf("logs view missing %q:\n%s", want, view)
		}
	}
}

func TestLogsSectionLevelToggle(t *testing.T) {
	ring := seededLogRing()
	ctx := fixtureDashCtxWithTrends()
	ctx.Logs = ring

	s := newLogsSection()
	// Switch to debug — the trace entry should appear.
	s.Update(ctx, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	view := s.View(ctx)
	if !strings.Contains(view, "trace") {
		t.Fatalf("debug view should include trace:\n%s", view)
	}

	// Switch to error — only the error entry remains.
	s.Update(ctx, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	view = s.View(ctx)
	if strings.Contains(view, "snapshot loaded") || strings.Contains(view, "parse failure") {
		t.Fatalf("error-only view leaked lower levels:\n%s", view)
	}
	if !strings.Contains(view, "db locked") {
		t.Fatalf("error-only view missing error entry:\n%s", view)
	}
}

func TestRingHandlerCapacityAndOrder(t *testing.T) {
	ring := NewRingHandler(3, slog.LevelDebug, nil)
	logger := slog.New(ring)
	logger.Info("first")
	time.Sleep(time.Millisecond) // ensure ordering via timestamp
	logger.Info("second")
	time.Sleep(time.Millisecond)
	logger.Info("third")
	time.Sleep(time.Millisecond)
	logger.Info("fourth") // should push "first" out

	entries := ring.Snapshot()
	if len(entries) != 3 {
		t.Fatalf("snapshot len = %d, want 3", len(entries))
	}
	want := []string{"second", "third", "fourth"}
	for i, e := range entries {
		if e.Message != want[i] {
			t.Errorf("entries[%d] = %q, want %q", i, e.Message, want[i])
		}
	}
}
