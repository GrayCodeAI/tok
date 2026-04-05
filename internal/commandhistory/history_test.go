package commandhistory

import (
	"testing"
	"time"
)

func TestNewHistoryManager(t *testing.T) {
	hm := NewHistoryManager(100)
	if hm == nil {
		t.Fatal("NewHistoryManager returned nil")
	}
}

func TestHistoryManager_AddEntry(t *testing.T) {
	hm := NewHistoryManager(10)
	entry := HistoryEntry{
		Command:   "git status",
		Timestamp: time.Now(),
		Duration:  time.Second,
	}
	hm.AddEntry(entry)

	recent := hm.GetRecent(10)
	if len(recent) != 1 {
		t.Errorf("GetRecent(10) = %d entries, want 1", len(recent))
	}
}

func TestHistoryManager_GetRecent(t *testing.T) {
	hm := NewHistoryManager(10)

	for i := 0; i < 5; i++ {
		hm.AddEntry(HistoryEntry{
			Command:   "cmd",
			Timestamp: time.Now(),
			Duration:  time.Second,
		})
	}

	recent := hm.GetRecent(3)
	if len(recent) != 3 {
		t.Errorf("GetRecent(3) = %d, want 3", len(recent))
	}
}

func TestHistoryManager_Search(t *testing.T) {
	hm := NewHistoryManager(10)
	hm.AddEntry(HistoryEntry{Command: "git status", Timestamp: time.Now(), Duration: time.Second})
	hm.AddEntry(HistoryEntry{Command: "git log", Timestamp: time.Now(), Duration: time.Second})
	hm.AddEntry(HistoryEntry{Command: "npm test", Timestamp: time.Now(), Duration: time.Second})

	results := hm.Search("git")
	if len(results) != 2 {
		t.Errorf("Search('git') = %d results, want 2", len(results))
	}
}

func TestHistoryManager_GetByTag(t *testing.T) {
	hm := NewHistoryManager(10)
	hm.AddEntry(HistoryEntry{Command: "cmd1", Timestamp: time.Now(), Duration: time.Second})

	results := hm.GetByTag("git")
	_ = results // should not panic
}

func TestHistoryManager_MaxSize(t *testing.T) {
	hm := NewHistoryManager(3)
	for i := 0; i < 5; i++ {
		hm.AddEntry(HistoryEntry{Command: "cmd", Timestamp: time.Now(), Duration: time.Second})
	}

	recent := hm.GetRecent(10)
	if len(recent) != 3 {
		t.Errorf("history capped at 3, got %d", len(recent))
	}
}

func TestHistoryManager_LimitGreaterThanSize(t *testing.T) {
	hm := NewHistoryManager(3)
	hm.AddEntry(HistoryEntry{Command: "cmd1", Timestamp: time.Now(), Duration: time.Second})

	recent := hm.GetRecent(100)
	if len(recent) != 1 {
		t.Errorf("GetRecent(100) = %d, want 1", len(recent))
	}
}
