package tui

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"
)

// LogEntry is one captured slog record, flattened to the fields the
// Logs section renders. We keep our own shape rather than embedding
// slog.Record so callers (including tests) don't depend on slog's
// internals and so the ring can be snapshotted without locking.
type LogEntry struct {
	Time    time.Time
	Level   slog.Level
	Message string
	// Attrs are flattened "key=value" pairs preserved in the order
	// slog emitted them. Keeping them as strings is enough for our
	// single-line rendering; structured values are %v-formatted.
	Attrs []string
}

// ringHandler is a slog.Handler that retains the most recent N records
// in memory. Intended for TUI consumption: cheap to install, cheap to
// tear down, and bounded memory regardless of how noisy the process is.
type ringHandler struct {
	mu       sync.Mutex
	capacity int
	buffer   []LogEntry
	next     int
	size     int
	// Delegate lets the handler tee records to a second handler (e.g.
	// the original JSON-to-stderr handler) so TUI log capture doesn't
	// silence logs for anything else that's reading them.
	delegate slog.Handler
	minLevel slog.Level
}

// NewRingHandler returns a ring handler with the given capacity. If
// delegate is non-nil, every captured record is also forwarded. Set
// minLevel to slog.LevelInfo to drop Debug in the ring; callers can
// still forward every level to the delegate.
func NewRingHandler(capacity int, minLevel slog.Level, delegate slog.Handler) *ringHandler {
	if capacity <= 0 {
		capacity = 256
	}
	return &ringHandler{
		capacity: capacity,
		buffer:   make([]LogEntry, capacity),
		delegate: delegate,
		minLevel: minLevel,
	}
}

// Enabled reports whether the handler accepts records at level. We
// accept anything at or above minLevel; we also always consult the
// delegate so that downstream filters still run.
func (h *ringHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if level >= h.minLevel {
		return true
	}
	if h.delegate != nil {
		return h.delegate.Enabled(ctx, level)
	}
	return false
}

func (h *ringHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= h.minLevel {
		attrs := make([]string, 0, r.NumAttrs())
		r.Attrs(func(a slog.Attr) bool {
			attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value.Any()))
			return true
		})
		entry := LogEntry{
			Time:    r.Time,
			Level:   r.Level,
			Message: r.Message,
			Attrs:   attrs,
		}
		h.mu.Lock()
		h.buffer[h.next] = entry
		h.next = (h.next + 1) % h.capacity
		if h.size < h.capacity {
			h.size++
		}
		h.mu.Unlock()
	}
	if h.delegate != nil {
		return h.delegate.Handle(ctx, r)
	}
	return nil
}

func (h *ringHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var d slog.Handler
	if h.delegate != nil {
		d = h.delegate.WithAttrs(attrs)
	}
	return &ringHandler{
		capacity: h.capacity,
		buffer:   h.buffer,
		delegate: d,
		minLevel: h.minLevel,
	}
}

func (h *ringHandler) WithGroup(name string) slog.Handler {
	var d slog.Handler
	if h.delegate != nil {
		d = h.delegate.WithGroup(name)
	}
	return &ringHandler{
		capacity: h.capacity,
		buffer:   h.buffer,
		delegate: d,
		minLevel: h.minLevel,
	}
}

// Snapshot returns a copy of the buffer in chronological order (oldest
// first). Safe to call from any goroutine; the returned slice is owned
// by the caller.
func (h *ringHandler) Snapshot() []LogEntry {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]LogEntry, 0, h.size)
	start := 0
	if h.size == h.capacity {
		start = h.next
	}
	for i := 0; i < h.size; i++ {
		idx := (start + i) % h.capacity
		out = append(out, h.buffer[idx])
	}
	// Defensive: if clocks jitter we still want time-ordered output.
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Time.Before(out[j].Time)
	})
	return out
}

// Clear empties the ring without disturbing the delegate.
func (h *ringHandler) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.next = 0
	h.size = 0
}

// formatLevel returns a short, colorable label for a slog level.
func formatLevel(l slog.Level) string {
	switch {
	case l >= slog.LevelError:
		return "ERR"
	case l >= slog.LevelWarn:
		return "WRN"
	case l >= slog.LevelInfo:
		return "INF"
	default:
		return "DBG"
	}
}

// formatEntry one-liner suitable for table-style log rows.
func formatEntry(e LogEntry) string {
	b := strings.Builder{}
	b.WriteString(e.Time.Format("15:04:05"))
	b.WriteString(" ")
	b.WriteString(formatLevel(e.Level))
	b.WriteString("  ")
	b.WriteString(e.Message)
	if len(e.Attrs) > 0 {
		b.WriteString("  ")
		b.WriteString(strings.Join(e.Attrs, " "))
	}
	return b.String()
}
