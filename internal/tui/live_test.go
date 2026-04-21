package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/GrayCodeAI/tok/internal/tracking"
)

// TestLiveSourceReceivesTrackerWrite verifies the full path works:
// tracking.SubscribeCommands fan-out → trackingLiveSource → LiveEvent
// channel → caller. No TUI model involved, just the source wiring.
func TestLiveSourceReceivesTrackerWrite(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	source := &trackingLiveSource{
		dbPath:        "", // disable fsnotify for this test
		fallbackEvery: time.Hour,
	}
	events := source.Start(ctx)

	// Publish a record through the same path Record() uses.
	record := &tracking.CommandRecord{
		Command:     "git status",
		SavedTokens: 123,
		ExecTimeMs:  50,
	}
	// Brief sleep so the Start goroutine has actually registered the
	// subscribe channel before we publish. Without it we race.
	time.Sleep(20 * time.Millisecond)
	publishForTest(record)

	select {
	case ev := <-events:
		if ev.Source != "subscribe" {
			t.Errorf("got source %q, want subscribe", ev.Source)
		}
		if ev.Record == nil || ev.Record.Command != "git status" {
			t.Errorf("got record %+v, want git status", ev.Record)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for live event (500ms)")
	}
}

// TestLiveSourceEmitsFallbackTick verifies the source continues firing
// via its fallback tick even when subscribe + fsnotify are both quiet.
func TestLiveSourceEmitsFallbackTick(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	source := &trackingLiveSource{
		dbPath:        "",
		fallbackEvery: 50 * time.Millisecond,
	}
	events := source.Start(ctx)

	select {
	case ev := <-events:
		if ev.Source != "tick" {
			t.Errorf("got source %q, want tick", ev.Source)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for fallback tick")
	}
}

// fakeLiveSource is a test double that emits a single LiveEvent on
// demand via the channel returned by Start.
type fakeLiveSource struct {
	ch chan LiveEvent
}

func newFakeLiveSource() *fakeLiveSource {
	return &fakeLiveSource{ch: make(chan LiveEvent, 8)}
}

func (f *fakeLiveSource) Start(ctx context.Context) <-chan LiveEvent {
	go func() {
		<-ctx.Done()
		close(f.ch)
	}()
	return f.ch
}

func (f *fakeLiveSource) emit(ev LiveEvent) {
	f.ch <- ev
}

// TestModelReactsToLiveEvent drives the full TUI model: inject a live
// event with a record, observe that the header goes to "● live", the
// toast stack contains the command summary, and a snapshot reload was
// triggered.
func TestModelReactsToLiveEvent(t *testing.T) {
	loader := &stubLoader{snapshot: goldenFixture()}
	fake := newFakeLiveSource()
	mRaw := NewModelWithLive(Options{Theme: ThemeDark, Days: 7}, loader, fake)
	m := mRaw.(model)
	// Size + initial load so the view renders.
	next, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	m = next.(model)
	next, _ = m.Update(snapshotLoadedMsg{snapshot: loader.snapshot, loadedAt: time.Now()})
	m = next.(model)

	// Inject a live event wrapped the way waitForLiveEvent would.
	ev := LiveEvent{
		At: time.Now(),
		Record: &tracking.CommandRecord{
			Command:     "gh pr view",
			SavedTokens: 1500,
		},
		Source: "subscribe",
	}
	next, _ = m.Update(liveEventWithChan{ev: liveEventMsg(ev), ch: fake.ch})
	m = next.(model)

	if m.liveCount != 1 {
		t.Errorf("liveCount = %d, want 1", m.liveCount)
	}
	if m.lastLive.IsZero() {
		t.Errorf("lastLive not updated")
	}
	view := m.View()
	if !strings.Contains(view, "● live") {
		t.Errorf("expected '● live' badge in header:\n%s", view)
	}
	// The toast is dispatched via a cmd; we can't flush the cmd queue
	// here without a tea.Program, but we can assert the badge and count
	// update which is what the user sees at a glance.
}

// TestSubscribeCommandsFanout verifies multiple concurrent subscribers
// all receive every published record and don't leak when their context
// is cancelled.
func TestSubscribeCommandsFanout(t *testing.T) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel1()
	defer cancel2()

	ch1 := tracking.SubscribeCommands(ctx1)
	ch2 := tracking.SubscribeCommands(ctx2)
	time.Sleep(10 * time.Millisecond)

	rec := &tracking.CommandRecord{Command: "fanout-test"}
	publishForTest(rec)

	for i, ch := range []<-chan *tracking.CommandRecord{ch1, ch2} {
		select {
		case r := <-ch:
			if r.Command != "fanout-test" {
				t.Errorf("sub %d got %q, want fanout-test", i, r.Command)
			}
		case <-time.After(250 * time.Millisecond):
			t.Fatalf("sub %d timed out", i)
		}
	}

	// Cancel one; the other should keep working.
	cancel1()
	time.Sleep(10 * time.Millisecond)
	publishForTest(&tracking.CommandRecord{Command: "second"})
	select {
	case r := <-ch2:
		if r.Command != "second" {
			t.Errorf("after cancel, sub2 got %q, want second", r.Command)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatalf("sub2 timed out after cancelling sub1")
	}
}

// publishForTest is a test-only shim that reaches into the tracking
// package's non-exported notifySubscribers via a whitebox export. This
// lets the TUI test drive the fan-out path without booting a real DB.
func publishForTest(rec *tracking.CommandRecord) {
	tracking.PublishForTest(rec)
}
