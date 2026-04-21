package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/GrayCodeAI/tok/internal/tracking"
)

// TestLiveFeedPanelShowsOnEvent injects a live subscribe event, renders
// Home, and asserts the Live Feed panel appears with the command.
func TestLiveFeedPanelShowsOnEvent(t *testing.T) {
	loader := &stubLoader{snapshot: goldenFixture()}
	fake := newFakeLiveSource()
	mRaw := NewModelWithLive(Options{Theme: ThemeDark, Days: 7}, loader, fake)
	m := mRaw.(model)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 180, Height: 60})
	m = next.(model)
	next, _ = m.Update(snapshotLoadedMsg{snapshot: loader.snapshot, loadedAt: time.Now()})
	m = next.(model)

	ev := LiveEvent{
		At: time.Now(),
		Record: &tracking.CommandRecord{
			Command:     "git status --short",
			SavedTokens: 420,
		},
		Source: "subscribe",
	}
	next, _ = m.Update(liveEventWithChan{ev: liveEventMsg(ev), ch: fake.ch})
	m = next.(model)

	view := m.View()
	if !strings.Contains(view, "Live Feed") {
		t.Fatalf("expected 'Live Feed' panel in Home view:\n%s", view)
	}
	if !strings.Contains(view, "git status --short") {
		t.Errorf("expected command text in Live Feed")
	}
	if !strings.Contains(view, "+420") {
		t.Errorf("expected '+420' saved-tokens in Live Feed")
	}
}

// TestLiveFeedCapsAt20 verifies the ring buffer bound holds.
func TestLiveFeedCapsAt20(t *testing.T) {
	loader := &stubLoader{snapshot: goldenFixture()}
	fake := newFakeLiveSource()
	mRaw := NewModelWithLive(Options{Theme: ThemeDark, Days: 7}, loader, fake)
	m := mRaw.(model)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	m = next.(model)
	next, _ = m.Update(snapshotLoadedMsg{snapshot: loader.snapshot, loadedAt: time.Now()})
	m = next.(model)

	for i := 0; i < 30; i++ {
		rec := &tracking.CommandRecord{
			Command:     "cmd",
			SavedTokens: i,
		}
		ev := LiveEvent{At: time.Now(), Record: rec, Source: "subscribe"}
		next, _ = m.Update(liveEventWithChan{ev: liveEventMsg(ev), ch: fake.ch})
		m = next.(model)
	}

	if got := len(m.liveFeed); got != maxLiveFeed {
		t.Fatalf("liveFeed len = %d, want %d", got, maxLiveFeed)
	}
	// Newest first; with SavedTokens=29 as the last insert.
	if m.liveFeed[0].SavedTokens != 29 {
		t.Errorf("newest entry SavedTokens = %d, want 29", m.liveFeed[0].SavedTokens)
	}
}
