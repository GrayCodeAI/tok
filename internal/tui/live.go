package tui

import (
	"context"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/GrayCodeAI/tok/internal/config"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

// LiveEvent is published by the liveSource whenever the underlying
// tracking data has changed. Record is set when the event originated
// from the in-process subscribe channel (a `tok <cmd>` ran in this
// process); it's nil when the event came from fsnotify (another
// process wrote to the DB) or from the fallback tick.
type LiveEvent struct {
	At     time.Time
	Record *tracking.CommandRecord
	Source string // "subscribe" | "fsnotify" | "tick"
}

// liveSource is the TUI's event source for "something changed, reload".
// Start returns a channel that emits LiveEvents until ctx is cancelled.
// The source itself is safe to construct and Start more than once only
// if each call gets its own context.
type liveSource interface {
	Start(ctx context.Context) <-chan LiveEvent
}

// trackingLiveSource multiplexes three signals into a single stream:
//
//  1. tracking.SubscribeCommands — instant, same-process writes
//  2. fsnotify on the tracking.db-wal file — other-process writes
//  3. a fallback time.Ticker — safety net in case both of the above
//     miss an event (fsnotify drops on saturation, subscribers drop
//     on a full buffer)
//
// The fallback tick is deliberately slow (15s) because the first two
// paths cover virtually every real case. If you're seeing UI staleness,
// the fix is usually "your writer is bypassing Record()", not "lower
// the tick."
type trackingLiveSource struct {
	dbPath        string
	fallbackEvery time.Duration
}

// newTrackingLiveSource returns a live source rooted at the user's tok
// tracking DB. Falls back to the in-process subscribe stream only if
// the DB path can't be resolved (e.g. first launch before init).
func newTrackingLiveSource() *trackingLiveSource {
	return &trackingLiveSource{
		dbPath:        resolveTrackingDBPath(),
		fallbackEvery: 15 * time.Second,
	}
}

func resolveTrackingDBPath() string {
	return config.DatabasePath()
}

// Start begins emitting events on the returned channel. Callers must
// drain it; a slow consumer causes drops, not backpressure. Cancel ctx
// to shut down all three underlying goroutines and close the channel.
func (s *trackingLiveSource) Start(ctx context.Context) <-chan LiveEvent {
	out := make(chan LiveEvent, 64)

	go func() {
		defer close(out)

		subs := tracking.SubscribeCommands(ctx)
		fs := s.watchFile(ctx)
		tick := time.NewTicker(s.fallbackEvery)
		defer tick.Stop()

		emit := func(ev LiveEvent) {
			select {
			case out <- ev:
			default:
				// Consumer is slow — drop. A later event or the
				// fallback tick will catch them back up.
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case rec, ok := <-subs:
				if !ok {
					subs = nil
					continue
				}
				emit(LiveEvent{At: time.Now(), Record: rec, Source: "subscribe"})
			case _, ok := <-fs:
				if !ok {
					fs = nil
					continue
				}
				emit(LiveEvent{At: time.Now(), Source: "fsnotify"})
			case t := <-tick.C:
				emit(LiveEvent{At: t, Source: "tick"})
			}
		}
	}()

	return out
}

// watchFile returns a channel that signals whenever the WAL or DB file
// mutates. Returns a nil (never-firing) channel if fsnotify can't be
// initialized or the path doesn't exist — fall through to tick+subscribe.
func (s *trackingLiveSource) watchFile(ctx context.Context) <-chan struct{} {
	if s.dbPath == "" {
		return nil
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Debug("tui live: fsnotify init failed, falling back to tick", "err", err)
		return nil
	}
	// Watch the containing directory rather than the individual file:
	// the DB + WAL + SHM files get rotated and fsnotify loses individual
	// file handles across rotations. Dir watches survive rotation and
	// we filter events by basename.
	dir := filepath.Dir(s.dbPath)
	if err := w.Add(dir); err != nil {
		w.Close()
		slog.Debug("tui live: fsnotify Add failed", "dir", dir, "err", err)
		return nil
	}

	out := make(chan struct{}, 8)
	go func() {
		defer close(out)
		defer w.Close()
		base := filepath.Base(s.dbPath)
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-w.Events:
				if !ok {
					return
				}
				// Only Write/Create on the DB or its WAL/SHM siblings
				// matter. Other dir activity is noise.
				name := filepath.Base(ev.Name)
				if name != base && name != base+"-wal" && name != base+"-shm" {
					continue
				}
				if ev.Op&(fsnotify.Write|fsnotify.Create) == 0 {
					continue
				}
				select {
				case out <- struct{}{}:
				default:
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				slog.Debug("tui live: fsnotify error", "err", err)
			}
		}
	}()
	return out
}

// nullLiveSource is used by tests that don't want a real watcher.
type nullLiveSource struct{}

func (nullLiveSource) Start(context.Context) <-chan LiveEvent {
	ch := make(chan LiveEvent)
	close(ch)
	return ch
}
