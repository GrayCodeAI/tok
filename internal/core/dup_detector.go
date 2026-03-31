package core

import (
	"strings"
	"sync"
	"time"
)

type DupEntry struct {
	Command   string
	Count     int
	FirstSeen time.Time
	LastSeen  time.Time
	TotalTokens int
}

type DupDetector struct {
	mu      sync.RWMutex
	entries map[string]*DupEntry
	window  time.Duration
}

func NewDupDetector(window time.Duration) *DupDetector {
	if window == 0 {
		window = 5 * time.Minute
	}
	return &DupDetector{
		entries: make(map[string]*DupEntry),
		window:  window,
	}
}

func (d *DupDetector) Record(command string, tokens int) *DupEntry {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := normalizeCommand(command)
	now := time.Now()

	if entry, exists := d.entries[key]; exists {
		entry.Count++
		entry.LastSeen = now
		entry.TotalTokens += tokens
		return entry
	}

	d.entries[key] = &DupEntry{
		Command:     command,
		Count:       1,
		FirstSeen:   now,
		LastSeen:    now,
		TotalTokens: tokens,
	}
	return nil
}

func (d *DupDetector) GetDuplicates() []*DupEntry {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var dups []*DupEntry
	for _, e := range d.entries {
		if e.Count > 1 {
			dups = append(dups, e)
		}
	}
	return dups
}

func (d *DupDetector) Cleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()

	cutoff := time.Now().Add(-d.window)
	for k, e := range d.entries {
		if e.LastSeen.Before(cutoff) {
			delete(d.entries, k)
		}
	}
}

func normalizeCommand(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}
	key := parts[0]
	if len(parts) > 1 {
		key += " " + parts[1]
	}
	return strings.ToLower(key)
}
