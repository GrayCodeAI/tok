package tui

// historyEntry captures the complete state needed to restore a view.
// Used for forward/back navigation across sections and drill-downs.
type historyEntry struct {
	SectionIndex int    // which section (0-based)
	ScrollOffset int    // vertical scroll position
	DrillKey     string // drill-down identifier (e.g., session ID), empty for list view
}

// historyStack implements a bounded navigation history with back/forward.
// Max depth prevents unbounded growth during long sessions.
type historyStack struct {
	entries []historyEntry
	pos     int // current position in entries, -1 if empty
	max     int // maximum depth
}

// newHistoryStack creates a stack with a maximum depth (default 50).
func newHistoryStack(maxDepth int) *historyStack {
	if maxDepth <= 0 {
		maxDepth = 50
	}
	return &historyStack{
		entries: make([]historyEntry, 0, maxDepth),
		pos:     -1,
		max:     maxDepth,
	}
}

// CanGoBack returns true if there are entries before current position.
func (h *historyStack) CanGoBack() bool {
	return h.pos > 0
}

// CanGoForward returns true if there are entries after current position.
func (h *historyStack) CanGoForward() bool {
	return h.pos >= 0 && h.pos < len(h.entries)-1
}

// Back moves to the previous entry and returns it.
// Returns zero value and false if at the beginning.
func (h *historyStack) Back() (historyEntry, bool) {
	if !h.CanGoBack() {
		return historyEntry{}, false
	}
	h.pos--
	return h.entries[h.pos], true
}

// Forward moves to the next entry and returns it.
// Returns zero value and false if at the end.
func (h *historyStack) Forward() (historyEntry, bool) {
	if !h.CanGoForward() {
		return historyEntry{}, false
	}
	h.pos++
	return h.entries[h.pos], true
}

// Current returns the current entry without moving.
// Returns zero value and false if history is empty.
func (h *historyStack) Current() (historyEntry, bool) {
	if h.pos < 0 || h.pos >= len(h.entries) {
		return historyEntry{}, false
	}
	return h.entries[h.pos], true
}

// Push adds a new entry after the current position, truncating any
// forward history. Duplicate of the current entry is ignored.
func (h *historyStack) Push(e historyEntry) {
	// Don't push duplicates of current position
	if cur, ok := h.Current(); ok && cur == e {
		return
	}

	// Truncate forward history if we're not at the end
	if h.pos >= 0 && h.pos < len(h.entries)-1 {
		h.entries = h.entries[:h.pos+1]
	}

	// Append new entry
	h.entries = append(h.entries, e)
	h.pos++

	// Trim from front if over capacity
	if len(h.entries) > h.max {
		h.entries = h.entries[1:]
		h.pos--
	}
}

// PushSection is a convenience helper for simple section jumps.
func (h *historyStack) PushSection(sectionIdx, scrollOffset int) {
	h.Push(historyEntry{
		SectionIndex: sectionIdx,
		ScrollOffset: scrollOffset,
		DrillKey:     "",
	})
}

// Len returns the total number of entries in history.
func (h *historyStack) Len() int {
	return len(h.entries)
}

// Pos returns the current position (0-based), -1 if empty.
func (h *historyStack) Pos() int {
	return h.pos
}
