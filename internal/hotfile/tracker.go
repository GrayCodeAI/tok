package hotfile

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type HotFile struct {
	Path        string
	AccessCount int64
	LastAccess  time.Time
	FileSize    int64
	FileType    string
	ProjectPath string
	Score       float64
}

type HotFileTracker struct {
	mu           sync.RWMutex
	files        map[string]*HotFile
	projectIndex map[string][]string
	maxFiles     int
	window       time.Duration
}

type HotFileConfig struct {
	MaxFiles       int
	Window         time.Duration
	MinAccessCount int
	TrackTypes     []string
}

func NewHotFileTracker(config HotFileConfig) *HotFileTracker {
	if config.MaxFiles == 0 {
		config.MaxFiles = 1000
	}
	if config.Window == 0 {
		config.Window = 24 * time.Hour
	}
	if config.MinAccessCount == 0 {
		config.MinAccessCount = 3
	}

	return &HotFileTracker{
		files:        make(map[string]*HotFile),
		projectIndex: make(map[string][]string),
		maxFiles:     config.MaxFiles,
		window:       config.Window,
	}
}

func (t *HotFileTracker) Track(ctx context.Context, path, projectPath string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	now := time.Now()

	hotFile, exists := t.files[absPath]
	if !exists {
		hotFile = &HotFile{
			Path:        absPath,
			LastAccess:  now,
			FileType:    filepath.Ext(absPath),
			ProjectPath: projectPath,
		}
		t.files[absPath] = hotFile
	}

	hotFile.AccessCount++
	hotFile.LastAccess = now
	hotFile.FileSize = stat.Size()
	hotFile.Score = t.calculateScore(hotFile)

	t.updateProjectIndex(absPath, projectPath)

	return nil
}

func (t *HotFileTracker) calculateScore(hf *HotFile) float64 {
	recencyScore := float64(t.window.Seconds() - time.Since(hf.LastAccess).Seconds())
	if recencyScore < 0 {
		recencyScore = 0
	}

	frequencyScore := float64(hf.AccessCount) * 10

	sizeScore := 1.0
	if hf.FileSize > 10000 {
		sizeScore = 0.8
	}
	if hf.FileSize > 100000 {
		sizeScore = 0.5
	}

	return (recencyScore * 0.4) + (frequencyScore * 0.4) + (sizeScore * 0.2)
}

func (t *HotFileTracker) updateProjectIndex(path, projectPath string) {
	if projectPath == "" {
		return
	}

	t.projectIndex[projectPath] = append(t.projectIndex[projectPath], path)
}

func (t *HotFileTracker) GetHotFiles(projectPath string, limit int) []*HotFile {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var result []*HotFile

	if projectPath != "" {
		for _, path := range t.projectIndex[projectPath] {
			if hf, ok := t.files[path]; ok {
				result = append(result, hf)
			}
		}
	} else {
		for _, hf := range t.files {
			result = append(result, hf)
		}
	}

	for i := range result {
		result[i].Score = t.calculateScore(result[i])
	}

	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Score > result[i].Score {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result
}

func (t *HotFileTracker) GetRecommendations(ctx context.Context, projectPath string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	hotFiles := t.GetHotFiles(projectPath, 20)
	var recommendations []string

	extCounts := make(map[string]int)
	for _, hf := range hotFiles {
		extCounts[hf.FileType]++
	}

	typePriority := map[string]int{
		".go":   10,
		".ts":   9,
		".tsx":  9,
		".js":   8,
		".jsx":  8,
		".py":   7,
		".rs":   7,
		".java": 6,
		".json": 5,
		".yaml": 5,
		".yml":  5,
	}

	for _, hf := range hotFiles {
		priority := typePriority[hf.FileType]
		if priority > 0 {
			recommendations = append(recommendations, hf.Path)
		}
	}

	return recommendations
}

func (t *HotFileTracker) Search(ctx context.Context, query string) []*HotFile {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var result []*HotFile
	lowerQuery := query

	for _, hf := range t.files {
		if len(result) >= 50 {
			break
		}
		if contains(hf.Path, lowerQuery) {
			result = append(result, hf)
		}
	}

	return result
}

func (t *HotFileTracker) Export() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[string]interface{})
	result["files"] = t.files
	result["count"] = len(t.files)
	result["projects"] = len(t.projectIndex)

	return result
}

func (t *HotFileTracker) Prune(ctx context.Context) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	cutoff := time.Now().Add(-t.window)
	var toDelete []string

	for path, hf := range t.files {
		if hf.LastAccess.Before(cutoff) || hf.AccessCount < 3 {
			toDelete = append(toDelete, path)
		}
	}

	for _, path := range toDelete {
		delete(t.files, path)
	}

	return len(toDelete)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

type HotFileStore struct {
	mu    sync.RWMutex
	files map[string]*HotFile
	path  string
}

func NewHotFileStore(path string) *HotFileStore {
	return &HotFileStore{
		files: make(map[string]*HotFile),
		path:  path,
	}
}

func (s *HotFileStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return nil
}

func (s *HotFileStore) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return nil
}

func (s *HotFileStore) Add(hf *HotFile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[hf.Path] = hf
}

func (s *HotFileStore) Get(path string) (*HotFile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	hf, ok := s.files[path]
	return hf, ok
}

func (s *HotFileStore) List() []*HotFile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*HotFile, 0, len(s.files))
	for _, hf := range s.files {
		result = append(result, hf)
	}
	return result
}

func (s *HotFileStore) Delete(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.files, path)
}

func (s *HotFileStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files = make(map[string]*HotFile)
}
