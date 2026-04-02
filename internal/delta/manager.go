// Package delta implements file version tracking and diff-based compression.
package delta

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/GrayCodeAI/tokman/internal/filter"
)

// Version represents a stored file version.
type Version struct {
	Hash      string    `json:"hash"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Size      int64     `json:"size"`
	DiffFrom  string    `json:"diff_from,omitempty"` // Hash of parent version
}

// Manager tracks file versions and computes deltas.
type Manager struct {
	mu       sync.RWMutex
	versions map[string][]*Version // path -> versions
	baseDir  string
}

// NewManager creates a new delta manager.
func NewManager(baseDir string) *Manager {
	return &Manager{
		versions: make(map[string][]*Version),
		baseDir:  baseDir,
	}
}

// Track adds a new version of a file.
func (m *Manager) Track(path string, content string) (*Version, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hash := computeHash(content)
	versions := m.versions[path]

	// Check if identical to latest
	if len(versions) > 0 && versions[len(versions)-1].Hash == hash {
		return versions[len(versions)-1], nil
	}

	version := &Version{
		Hash:      hash,
		Content:   content,
		Timestamp: time.Now(),
		Size:      int64(len(content)),
	}

	// If we have previous versions, compute diff
	if len(versions) > 0 {
		last := versions[len(versions)-1]
		delta := filter.ComputeDelta(last.Content, content)

		// Store diff if it saves space
		diffSize := estimateDeltaSize(delta)
		if diffSize < len(content)/2 {
			version.DiffFrom = last.Hash
			// Don't store full content if we have a good diff
			version.Content = ""
		}
	}

	m.versions[path] = append(versions, version)

	// Limit versions per file (keep last 10)
	if len(m.versions[path]) > 10 {
		m.versions[path] = m.versions[path][len(m.versions[path])-10:]
	}

	return version, nil
}

// Get retrieves a specific version by hash.
func (m *Manager) Get(path string, hash string) (*Version, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	versions, ok := m.versions[path]
	if !ok {
		return nil, fmt.Errorf("no versions for path: %s", path)
	}

	// Find the version
	var target *Version
	for _, v := range versions {
		if strings.HasPrefix(v.Hash, hash) || v.Hash == hash {
			target = v
			break
		}
	}

	if target == nil {
		return nil, fmt.Errorf("version not found: %s", hash)
	}

	// If content is empty, reconstruct from diff
	if target.Content == "" && target.DiffFrom != "" {
		parent := m.findVersion(path, target.DiffFrom)
		if parent != nil {
			delta := filter.ComputeDelta(parent.Content, "")
			target.Content = applyDelta(parent.Content, delta)
		}
	}

	return target, nil
}

// GetLatest retrieves the latest version.
func (m *Manager) GetLatest(path string) (*Version, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	versions, ok := m.versions[path]
	if !ok || len(versions) == 0 {
		return nil, fmt.Errorf("no versions for path: %s", path)
	}

	latest := versions[len(versions)-1]

	// Reconstruct if needed
	if latest.Content == "" && latest.DiffFrom != "" {
		parent := m.findVersion(path, latest.DiffFrom)
		if parent != nil {
			delta := filter.ComputeDelta(parent.Content, "")
			latest.Content = applyDelta(parent.Content, delta)
		}
	}

	return latest, nil
}

// ComputeDiff returns the diff between two versions.
func (m *Manager) ComputeDiff(path string, fromHash, toHash string) (filter.IncrementalDelta, error) {
	from, err := m.Get(path, fromHash)
	if err != nil {
		return filter.IncrementalDelta{}, err
	}

	to, err := m.Get(path, toHash)
	if err != nil {
		return filter.IncrementalDelta{}, err
	}

	return filter.ComputeDelta(from.Content, to.Content), nil
}

// GetVersions returns all versions for a path.
func (m *Manager) GetVersions(path string) []*Version {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Version, len(m.versions[path]))
	copy(result, m.versions[path])
	return result
}

// ShouldCompress determines if delta compression should be used.
func (m *Manager) ShouldCompress(path string, content string) (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	versions, ok := m.versions[path]
	if !ok || len(versions) == 0 {
		return false, ""
	}

	latest := versions[len(versions)-1]
	if latest.Hash == computeHash(content) {
		return false, latest.Hash // Identical
	}

	delta := filter.ComputeDelta(latest.Content, content)
	diffRatio := float64(len(delta.Added)+len(delta.Removed)) / float64(len(content))

	// Use delta if changes are less than 50% of content
	if diffRatio < 0.5 {
		return true, latest.Hash
	}

	return false, ""
}

// Cleanup removes old versions.
func (m *Manager) Cleanup(path string, keep int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	versions, ok := m.versions[path]
	if !ok || len(versions) <= keep {
		return nil
	}

	m.versions[path] = versions[len(versions)-keep:]
	return nil
}

// Save persists versions to disk.
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(m.baseDir, 0755); err != nil {
		return err
	}

	// Save each file's versions
	for path, versions := range m.versions {
		safePath := sanitizePath(path)
		filePath := filepath.Join(m.baseDir, safePath+".versions")

		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		// Simple text format for now
		var data strings.Builder
		for _, v := range versions {
			fmt.Fprintf(&data, "%s|%d|%s|%s\n", v.Hash, v.Timestamp.Unix(), v.DiffFrom, v.Content)
		}

		if err := os.WriteFile(filePath, []byte(data.String()), 0644); err != nil {
			return err
		}
	}

	return nil
}

// Load loads versions from disk.
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".versions") {
			continue
		}

		path := strings.TrimSuffix(entry.Name(), ".versions")
		data, err := os.ReadFile(filepath.Join(m.baseDir, entry.Name()))
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			parts := strings.Split(line, "|")
			if len(parts) < 4 {
				continue
			}

			ts, _ := parseInt64(parts[1])
			version := &Version{
				Hash:      parts[0],
				Timestamp: time.Unix(ts, 0),
				DiffFrom:  parts[2],
				Content:   parts[3],
			}
			m.versions[path] = append(m.versions[path], version)
		}
	}

	return nil
}

// Stats returns manager statistics.
func (m *Manager) Stats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := Stats{
		FilesTracked: len(m.versions),
	}

	for _, versions := range m.versions {
		stats.TotalVersions += len(versions)
		for _, v := range versions {
			stats.TotalSize += v.Size
		}
	}

	return stats
}

func (m *Manager) findVersion(path, hash string) *Version {
	versions := m.versions[path]
	for _, v := range versions {
		if v.Hash == hash || strings.HasPrefix(v.Hash, hash) {
			return v
		}
	}
	return nil
}

// Stats provides statistics about the manager.
type Stats struct {
	FilesTracked  int
	TotalVersions int
	TotalSize     int64
}

// CompressionDecision determines how to compress content.
type CompressionDecision struct {
	UseDelta         bool
	BaseHash         string
	FullContent      bool
	EstimatedSavings int
}

// DecideCompression determines the best compression strategy.
func (m *Manager) DecideCompression(path string, content string) CompressionDecision {
	decision := CompressionDecision{
		FullContent: true,
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	versions, ok := m.versions[path]
	if !ok || len(versions) == 0 {
		return decision
	}

	// Get most similar version
	var bestMatch *Version
	bestSimilarity := 0.0

	for _, v := range versions {
		sim := similarity(v.Content, content)
		if sim > bestSimilarity {
			bestSimilarity = sim
			bestMatch = v
		}
	}

	if bestMatch != nil && bestSimilarity > 0.7 {
		delta := filter.ComputeDelta(bestMatch.Content, content)
		deltaSize := estimateDeltaSize(delta)
		fullSize := len(content)

		if deltaSize < fullSize*2/3 {
			decision.UseDelta = true
			decision.BaseHash = bestMatch.Hash
			decision.FullContent = false
			decision.EstimatedSavings = fullSize - deltaSize
		}
	}

	return decision
}

// GitIntegration provides git-aware delta tracking.
type GitIntegration struct {
	manager *Manager
}

// NewGitIntegration creates git integration.
func NewGitIntegration(manager *Manager) *GitIntegration {
	return &GitIntegration{manager: manager}
}

// TrackChangedFiles tracks files changed since a ref.
func (g *GitIntegration) TrackChangedFiles(repoPath string, sinceRef string) ([]string, error) {
	// This would use git commands to find changed files
	// Placeholder implementation
	return nil, fmt.Errorf("git integration not fully implemented")
}

// GetCommitDelta returns delta for a commit.
func (g *GitIntegration) GetCommitDelta(repoPath, commitHash string) (*CommitDelta, error) {
	// Placeholder
	return nil, fmt.Errorf("git integration not fully implemented")
}

// CommitDelta represents changes in a commit.
type CommitDelta struct {
	CommitHash string
	ParentHash string
	Files      []FileDelta
}

// FileDelta represents a file change.
type FileDelta struct {
	Path       string
	OldHash    string
	NewHash    string
	ChangeType string // added, modified, deleted
}

// Helper functions

func computeHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

func estimateDeltaSize(delta filter.IncrementalDelta) int {
	size := 0
	for _, a := range delta.Added {
		size += len(a)
	}
	for _, r := range delta.Removed {
		size += len(r)
	}
	return size
}

func applyDelta(base string, delta filter.IncrementalDelta) string {
	// Simple reconstruction
	lines := strings.Split(base, "\n")
	result := make([]string, 0, len(lines))

	removed := make(map[string]bool)
	for _, r := range delta.Removed {
		removed[r] = true
	}

	for _, line := range lines {
		if !removed[line] {
			result = append(result, line)
		}
	}

	result = append(result, delta.Added...)
	return strings.Join(result, "\n")
}

func similarity(a, b string) float64 {
	if a == "" && b == "" {
		return 1.0
	}
	if a == "" || b == "" {
		return 0.0
	}

	// Simple line-based similarity
	linesA := strings.Split(a, "\n")
	linesB := strings.Split(b, "\n")

	setA := make(map[string]bool)
	for _, line := range linesA {
		setA[strings.TrimSpace(line)] = true
	}

	common := 0
	for _, line := range linesB {
		if setA[strings.TrimSpace(line)] {
			common++
		}
	}

	maxLen := len(linesA)
	if len(linesB) > maxLen {
		maxLen = len(linesB)
	}

	if maxLen == 0 {
		return 1.0
	}

	return float64(common) / float64(maxLen)
}

func sanitizePath(path string) string {
	// Replace path separators with safe characters
	return strings.ReplaceAll(strings.ReplaceAll(path, "/", "_"), "\\", "_")
}

func parseInt64(s string) (int64, error) {
	var n int64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid number")
		}
		n = n*10 + int64(ch-'0')
	}
	return n, nil
}
