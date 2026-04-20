package tracking

import (
	"context"
	"sync"
)

// commandSub is one live subscriber. Record() pushes a non-blocking copy
// of each newly-written record to every sub's channel. If the channel is
// full we drop rather than back-pressure the write path — live UIs are
// allowed to miss events; the SQL row is still canonical.
type commandSub struct {
	ch chan *CommandRecord
}

// subscribers is a package-level registry so the global Tracker and any
// user-owned Tracker share the same fan-out. Record() calls
// notifySubscribers after a successful INSERT.
var (
	subsMu sync.RWMutex
	subs   []*commandSub
)

// SubscribeCommands returns a channel that receives a pointer to every
// CommandRecord written by any Tracker in this process after the call.
// The channel is buffered (capacity 32); full sends are dropped.
//
// The subscription is tied to ctx — when ctx is canceled the channel is
// closed and the subscriber is removed. Callers MUST drain until close
// or cancel ctx; leaking a subscriber does not panic but wastes memory.
//
// Multiple concurrent subscribers are supported. Each gets its own copy
// of the record pointer; the underlying value must not be mutated by
// subscribers.
func SubscribeCommands(ctx context.Context) <-chan *CommandRecord {
	sub := &commandSub{ch: make(chan *CommandRecord, 32)}
	subsMu.Lock()
	subs = append(subs, sub)
	subsMu.Unlock()

	go func() {
		<-ctx.Done()
		subsMu.Lock()
		for i, s := range subs {
			if s == sub {
				subs = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		subsMu.Unlock()
		close(sub.ch)
	}()

	return sub.ch
}

// notifySubscribers fans a freshly-recorded command out to all live
// subscribers. Non-blocking: if a subscriber's buffer is full we drop
// the event for that sub. This keeps Record() latency bounded even when
// a slow TUI is holding the channel.
func notifySubscribers(rec *CommandRecord) {
	if rec == nil {
		return
	}
	subsMu.RLock()
	defer subsMu.RUnlock()
	for _, s := range subs {
		select {
		case s.ch <- rec:
		default:
			// Subscriber lagging — drop. The TUI resync on next tick.
		}
	}
}

// subscriberCount is a test-only helper reporting active subscriber count.
func subscriberCount() int {
	subsMu.RLock()
	defer subsMu.RUnlock()
	return len(subs)
}
