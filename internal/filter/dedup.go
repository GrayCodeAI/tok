package filter

import (
	"hash/fnv"
	"strings"
	"sync"
)

// SimHash computes a 64-bit fingerprint for content deduplication.
// Uses character n-gram hashing with Hamming distance for near-duplicate detection.
func SimHash(content string) uint64 {
	var v [64]int

	// Generate character trigrams
	for i := 0; i+2 < len(content); i++ {
		ngram := content[i : i+3]
		h := fnv.New64a()
		h.Write([]byte(ngram))
		hash := h.Sum64()

		for j := 0; j < 64; j++ {
			if hash&(1<<uint(j)) != 0 {
				v[j]++
			} else {
				v[j]--
			}
		}
	}

	// Build final hash
	var result uint64
	for j := 0; j < 64; j++ {
		if v[j] > 0 {
			result |= 1 << uint(j)
		}
	}
	return result
}

// HammingDistance returns the number of differing bits between two hashes.
func HammingDistance(a, b uint64) int {
	xor := a ^ b
	dist := 0
	for xor != 0 {
		dist++
		xor &= xor - 1
	}
	return dist
}

// IsNearDuplicate returns true if two content blocks are near-duplicates.
// Uses SimHash with configurable Hamming distance threshold.
func IsNearDuplicate(a, b string, threshold int) bool {
	hashA := SimHash(a)
	hashB := SimHash(b)
	return HammingDistance(hashA, hashB) <= threshold
}

// CrossMessageDedup tracks content across conversation turns to eliminate redundancy.
type CrossMessageDedup struct {
	mu        sync.RWMutex
	seen      map[uint64]string // hash -> original content
	threshold int               // max Hamming distance for near-duplicate
}

// NewCrossMessageDedup creates a new cross-message deduplication tracker.
func NewCrossMessageDedup() *CrossMessageDedup {
	return &CrossMessageDedup{
		seen:      make(map[uint64]string),
		threshold: 3,
	}
}

// DedupMessage checks if a message is a duplicate of previously seen content.
// Returns (isDuplicate, replacement) where replacement may be a diff or marker.
func (d *CrossMessageDedup) DedupMessage(content string) (bool, string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	hash := SimHash(content)

	// Exact match
	if orig, ok := d.seen[hash]; ok && orig == content {
		return true, "[duplicate: previously sent]"
	}

	// Near-duplicate check
	for seenHash, orig := range d.seen {
		if HammingDistance(hash, seenHash) <= d.threshold {
			// Generate diff for similar content
			diff := generateDiff(orig, content)
			if diff != "" {
				return true, diff
			}
			return true, "[near-duplicate: " + orig[:min(len(orig), 50)] + "...]"
		}
	}

	// New content
	d.seen[hash] = content
	return false, content
}

// Clear resets the deduplication tracker.
func (d *CrossMessageDedup) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen = make(map[uint64]string)
}

// Count returns the number of unique content blocks tracked.
func (d *CrossMessageDedup) Count() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.seen)
}

// generateDiff produces a unified diff between two similar strings.
func generateDiff(old, new string) string {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var added, removed []string
	oldSet := make(map[string]bool)
	newSet := make(map[string]bool)

	for _, l := range oldLines {
		oldSet[l] = true
	}
	for _, l := range newLines {
		newSet[l] = true
	}

	for l := range oldSet {
		if !newSet[l] {
			removed = append(removed, "- "+l)
		}
	}
	for l := range newSet {
		if !oldSet[l] {
			added = append(added, "+ "+l)
		}
	}

	if len(added) == 0 && len(removed) == 0 {
		return ""
	}

	result := "[diff]\n"
	result += strings.Join(removed, "\n")
	if len(removed) > 0 && len(added) > 0 {
		result += "\n"
	}
	result += strings.Join(added, "\n")
	return result
}
