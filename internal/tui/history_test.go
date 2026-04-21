package tui

import (
	"testing"
)

func TestHistoryStackBasic(t *testing.T) {
	h := newHistoryStack(10)

	// Empty stack
	if h.CanGoBack() {
		t.Error("empty stack should not allow back")
	}
	if h.CanGoForward() {
		t.Error("empty stack should not allow forward")
	}

	// Push first entry
	h.PushSection(0, 0)
	if h.Len() != 1 {
		t.Errorf("expected len 1, got %d", h.Len())
	}
	if h.CanGoBack() {
		t.Error("single entry should not allow back")
	}

	// Push more entries
	h.PushSection(1, 10)
	h.PushSection(2, 20)
	if h.Len() != 3 {
		t.Errorf("expected len 3, got %d", h.Len())
	}

	// Go back
	if !h.CanGoBack() {
		t.Error("should allow back after pushing multiple")
	}
	entry, ok := h.Back()
	if !ok {
		t.Error("Back() should succeed")
	}
	if entry.SectionIndex != 1 {
		t.Errorf("expected section 1, got %d", entry.SectionIndex)
	}
	if entry.ScrollOffset != 10 {
		t.Errorf("expected offset 10, got %d", entry.ScrollOffset)
	}

	// Go forward
	if !h.CanGoForward() {
		t.Error("should allow forward after going back")
	}
	entry, ok = h.Forward()
	if !ok {
		t.Error("Forward() should succeed")
	}
	if entry.SectionIndex != 2 {
		t.Errorf("expected section 2, got %d", entry.SectionIndex)
	}
}

func TestHistoryStackDeduplication(t *testing.T) {
	h := newHistoryStack(10)

	// Push same entry twice - should dedupe
	h.PushSection(0, 0)
	h.PushSection(0, 0) // duplicate
	if h.Len() != 1 {
		t.Errorf("expected len 1 after dedupe, got %d", h.Len())
	}

	// Push different entry
	h.PushSection(1, 10)
	if h.Len() != 2 {
		t.Errorf("expected len 2, got %d", h.Len())
	}
}

func TestHistoryStackTruncateForward(t *testing.T) {
	h := newHistoryStack(10)

	// Build history: 0 -> 1 -> 2
	h.PushSection(0, 0)
	h.PushSection(1, 10)
	h.PushSection(2, 20)

	// Go back to position 1
	h.Back() // now at 1 (section 1)

	// Push new entry - should truncate forward history
	h.PushSection(3, 30)

	// Should be: 0 -> 1 -> 3
	if h.Len() != 3 {
		t.Errorf("expected len 3 after truncation, got %d", h.Len())
	}
	if h.CanGoForward() {
		t.Error("should not allow forward after truncation")
	}

	// Verify current position
	entry, ok := h.Current()
	if !ok {
		t.Fatal("Current() failed")
	}
	if entry.SectionIndex != 3 {
		t.Errorf("expected section 3, got %d", entry.SectionIndex)
	}
}

func TestHistoryStackMaxDepth(t *testing.T) {
	h := newHistoryStack(3)

	// Push 5 entries with max depth 3
	h.PushSection(0, 0)
	h.PushSection(1, 10)
	h.PushSection(2, 20)
	h.PushSection(3, 30)
	h.PushSection(4, 40)

	if h.Len() != 3 {
		t.Errorf("expected len 3 at max depth, got %d", h.Len())
	}

	// Oldest entry (0) should have been evicted
	// History should be: 2 -> 3 -> 4
	entry, ok := h.Back()
	if !ok {
		t.Fatal("Back() failed")
	}
	if entry.SectionIndex != 3 {
		t.Errorf("expected section 3 after eviction, got %d", entry.SectionIndex)
	}
}

func TestHistoryStackWithDrillKey(t *testing.T) {
	h := newHistoryStack(10)

	// Push entry with drill key
	h.Push(historyEntry{
		SectionIndex: 7, // Sessions section
		ScrollOffset: 5,
		DrillKey:     "session-abc123",
	})

	entry, ok := h.Current()
	if !ok {
		t.Fatal("Current() failed")
	}
	if entry.DrillKey != "session-abc123" {
		t.Errorf("expected drill key 'session-abc123', got %q", entry.DrillKey)
	}
}

func TestHistoryStackEmptyNavigation(t *testing.T) {
	h := newHistoryStack(10)

	_, ok := h.Back()
	if ok {
		t.Error("Back() should fail on empty stack")
	}

	_, ok = h.Forward()
	if ok {
		t.Error("Forward() should fail on empty stack")
	}

	_, ok = h.Current()
	if ok {
		t.Error("Current() should fail on empty stack")
	}
}
