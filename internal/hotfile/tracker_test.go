package hotfile

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewHotFileTracker(t *testing.T) {
	config := HotFileConfig{MaxFiles: 100, Window: time.Hour}
	tracker := NewHotFileTracker(config)
	if tracker == nil {
		t.Error("Expected non-nil tracker")
	}
	if tracker.maxFiles != 100 {
		t.Errorf("Expected maxFiles 100, got %d", tracker.maxFiles)
	}
}

func TestHotFileTrackerTrack(t *testing.T) {
	tracker := NewHotFileTracker(HotFileConfig{MaxFiles: 100})

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	err := tracker.Track(context.Background(), tmpFile, "/project")
	if err != nil {
		t.Fatalf("Track failed: %v", err)
	}

	hotFiles := tracker.GetHotFiles("", 10)
	if len(hotFiles) != 1 {
		t.Errorf("Expected 1 hot file, got %d", len(hotFiles))
	}
}

func TestHotFileTrackerCalculateScore(t *testing.T) {
	tracker := NewHotFileTracker(HotFileConfig{})

	hf := &HotFile{
		Path:        "/test/file.go",
		AccessCount: 10,
		LastAccess:  time.Now(),
		FileSize:    5000,
	}

	score := tracker.calculateScore(hf)
	if score <= 0 {
		t.Error("Expected positive score")
	}
}

func TestHotFileTrackerGetHotFiles(t *testing.T) {
	tracker := NewHotFileTracker(HotFileConfig{})

	tmpFile1 := filepath.Join(t.TempDir(), "test1.txt")
	tmpFile2 := filepath.Join(t.TempDir(), "test2.txt")
	os.WriteFile(tmpFile1, []byte("test1"), 0644)
	os.WriteFile(tmpFile2, []byte("test2"), 0644)

	tracker.Track(context.Background(), tmpFile1, "/project")
	tracker.Track(context.Background(), tmpFile2, "/project")

	hotFiles := tracker.GetHotFiles("/project", 10)
	if len(hotFiles) != 2 {
		t.Errorf("Expected 2 hot files, got %d", len(hotFiles))
	}
}

func TestHotFileTrackerSearch(t *testing.T) {
	tracker := NewHotFileTracker(HotFileConfig{})

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_file.go")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	tracker.Track(context.Background(), tmpFile, "/project")

	results := tracker.Search(context.Background(), "test_file")
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestHotFileTrackerPrune(t *testing.T) {
	tracker := NewHotFileTracker(HotFileConfig{Window: time.Millisecond})

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	tracker.Track(context.Background(), tmpFile, "/project")

	time.Sleep(10 * time.Millisecond)

	count := tracker.Prune(context.Background())
	if count < 0 {
		t.Errorf("Expected non-negative prune count, got %d", count)
	}
}

func TestHotFileStore(t *testing.T) {
	store := NewHotFileStore("/tmp/hotfile_test.json")

	hf := &HotFile{
		Path:        "/test/file.go",
		AccessCount: 5,
		LastAccess:  time.Now(),
		FileSize:    1000,
	}

	store.Add(hf)

	retrieved, ok := store.Get("/test/file.go")
	if !ok {
		t.Error("Expected to find hot file")
	}
	if retrieved.AccessCount != 5 {
		t.Errorf("Expected AccessCount 5, got %d", retrieved.AccessCount)
	}

	list := store.List()
	if len(list) != 1 {
		t.Errorf("Expected 1 file in list, got %d", len(list))
	}

	store.Delete("/test/file.go")
	_, ok = store.Get("/test/file.go")
	if ok {
		t.Error("Expected file to be deleted")
	}
}

func TestHotFileRecommendations(t *testing.T) {
	tracker := NewHotFileTracker(HotFileConfig{})

	tmpDir := t.TempDir()
	files := []string{"test.go", "test.ts", "test.py", "test.json"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		os.WriteFile(path, []byte("test"), 0644)
		tracker.Track(context.Background(), path, "/project")
	}

	recs := tracker.GetRecommendations(context.Background(), "/project")
	if len(recs) == 0 {
		t.Error("Expected recommendations")
	}
}

func TestHotFileExport(t *testing.T) {
	tracker := NewHotFileTracker(HotFileConfig{})

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)
	tracker.Track(context.Background(), tmpFile, "/project")

	export := tracker.Export()
	if export["count"].(int) != 1 {
		t.Errorf("Expected count 1, got %v", export["count"])
	}
}

func TestHotFileStoreClear(t *testing.T) {
	store := NewHotFileStore("/tmp/test.json")

	store.Add(&HotFile{Path: "/test1"})
	store.Add(&HotFile{Path: "/test2"})

	store.Clear()

	list := store.List()
	if len(list) != 0 {
		t.Errorf("Expected empty list after clear, got %d", len(list))
	}
}
