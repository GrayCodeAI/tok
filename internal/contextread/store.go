package contextread

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/GrayCodeAI/tokman/internal/cache"
	"github.com/GrayCodeAI/tokman/internal/config"
)

const (
	maxSnapshots    = 128
	maxContentBytes = 256 * 1024
)

// Snapshot stores the last seen contents for a path so future reads can emit deltas.
type Snapshot struct {
	Path        string    `json:"path"`
	Fingerprint string    `json:"fingerprint"`
	Content     string    `json:"content"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Store is a small persisted state file used by smart read flows.
type Store struct {
	Snapshots map[string]Snapshot `json:"snapshots"`
}

// DefaultStorePath returns the persisted smart-read snapshot file path.
func DefaultStorePath() string {
	return filepath.Join(config.DataPath(), "read-state.json")
}

// Load reads the store from disk. Missing files are treated as empty state.
func Load(path string) (*Store, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Store{Snapshots: make(map[string]Snapshot)}, nil
		}
		return nil, err
	}

	store := &Store{}
	if err := json.Unmarshal(data, store); err != nil {
		return nil, err
	}
	if store.Snapshots == nil {
		store.Snapshots = make(map[string]Snapshot)
	}
	return store, nil
}

// Save writes the store back to disk atomically.
func (s *Store) Save(path string) error {
	if s.Snapshots == nil {
		s.Snapshots = make(map[string]Snapshot)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Get looks up a normalized path in the store.
func (s *Store) Get(path string) (Snapshot, bool) {
	if s == nil || s.Snapshots == nil {
		return Snapshot{}, false
	}
	key := normalizePath(path)
	snap, ok := s.Snapshots[key]
	return snap, ok
}

// Put updates the last-seen snapshot for a path.
func (s *Store) Put(path, content string) {
	if s.Snapshots == nil {
		s.Snapshots = make(map[string]Snapshot)
	}
	key := normalizePath(path)
	s.Snapshots[key] = Snapshot{
		Path:        key,
		Fingerprint: cache.ComputeFingerprint(content),
		Content:     trimContent(content),
		UpdatedAt:   time.Now(),
	}
	s.prune()
}

func (s *Store) prune() {
	if len(s.Snapshots) <= maxSnapshots {
		return
	}

	snaps := make([]Snapshot, 0, len(s.Snapshots))
	for _, snap := range s.Snapshots {
		snaps = append(snaps, snap)
	}
	sort.Slice(snaps, func(i, j int) bool {
		return snaps[i].UpdatedAt.Before(snaps[j].UpdatedAt)
	})

	toRemove := len(s.Snapshots) - maxSnapshots
	for i := 0; i < toRemove; i++ {
		delete(s.Snapshots, snaps[i].Path)
	}
}

func normalizePath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	if eval, err := filepath.EvalSymlinks(path); err == nil {
		path = eval
	}
	return filepath.Clean(path)
}

func trimContent(content string) string {
	if len(content) <= maxContentBytes {
		return content
	}
	return content[:maxContentBytes]
}
