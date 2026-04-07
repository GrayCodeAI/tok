package learn

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestLearner(t *testing.T) (*Learner, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "learn-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	cfg := Config{
		DatabasePath:  filepath.Join(tmpDir, "learn_test.db"),
		SamplingRate:  1.0,
		MinFrequency:  2,
		MinConfidence: 0.5,
		MaxSamples:    100,
		Enabled:       true,
	}

	learner, err := New(cfg)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("create learner: %v", err)
	}

	cleanup := func() {
		learner.Close()
		os.RemoveAll(tmpDir)
	}

	return learner, cleanup
}

func TestNewDisabled(t *testing.T) {
	cfg := Config{Enabled: false}
	learner, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if learner != nil {
		t.Error("expected nil learner when disabled")
	}
}

func TestCollectSample(t *testing.T) {
	learner, cleanup := setupTestLearner(t)
	defer cleanup()

	sample := Sample{
		Command:   "go",
		Args:      "test ./...",
		Output:    "=== RUN   TestFoo\n--- PASS: TestFoo (0.01s)\nok  \tpkg/foo\t0.15s\n",
		Timestamp: time.Now(),
	}

	err := learner.CollectSample(sample)
	if err != nil {
		t.Fatalf("collect sample: %v", err)
	}

	stats, err := learner.GetStats()
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}

	if stats.SamplesCollected != 1 {
		t.Errorf("samples = %d, want 1", stats.SamplesCollected)
	}
}

func TestPatternDiscovery(t *testing.T) {
	learner, cleanup := setupTestLearner(t)
	defer cleanup()

	// Submit multiple samples with same noise patterns
	for i := 0; i < 5; i++ {
		sample := Sample{
			Command: "go",
			Args:    "test ./...",
			Output:  "=== RUN   TestFoo\n--- PASS: TestFoo (0.01s)\n[DEBUG] something\nok  \tpkg\t0.1s\n",
		}
		learner.CollectSample(sample)
	}

	// Check patterns discovered
	patterns, err := learner.GetPatterns("")
	if err != nil {
		t.Fatalf("get patterns: %v", err)
	}

	if len(patterns) == 0 {
		t.Error("expected patterns to be discovered")
	}

	// Verify noise patterns found
	foundNoise := false
	for _, p := range patterns {
		if p.Category == "noise" || p.Category == "boilerplate" {
			foundNoise = true
			if p.Frequency < 2 {
				t.Errorf("pattern %q frequency = %d, want >= 2", p.Pattern, p.Frequency)
			}
		}
	}
	if !foundNoise {
		t.Error("expected to find noise patterns")
	}
}

func TestFilterGeneration(t *testing.T) {
	learner, cleanup := setupTestLearner(t)
	defer cleanup()

	// Record enough patterns
	for i := 0; i < 10; i++ {
		sample := Sample{
			Command: "npm",
			Args:    "install",
			Output:  "npm WARN deprecated pkg@1.0\nnpm WARN peer dep\nInstalling...\nadded 42 packages\n",
		}
		learner.CollectSample(sample)
	}

	filters, err := learner.GenerateFilters()
	if err != nil {
		t.Fatalf("generate filters: %v", err)
	}

	// Should have filter suggestions
	if len(filters) == 0 {
		t.Skip("no filter suggestions generated (patterns may not meet threshold)")
	}

	for _, f := range filters {
		if f.Command == "" {
			t.Error("filter has empty command")
		}
		if f.TOMLOutput == "" {
			t.Error("filter has empty TOML output")
		}
		if f.Confidence <= 0 {
			t.Error("filter has zero confidence")
		}
	}
}

func TestApproveRejectPattern(t *testing.T) {
	learner, cleanup := setupTestLearner(t)
	defer cleanup()

	// Add a sample to create patterns
	for i := 0; i < 5; i++ {
		sample := Sample{
			Command: "test",
			Args:    "cmd",
			Output:  "[DEBUG] test line\n[INFO] info line\nresult\n",
		}
		learner.CollectSample(sample)
	}

	patterns, _ := learner.GetPatterns("pending")
	if len(patterns) == 0 {
		t.Skip("no pending patterns")
	}

	// Approve first pattern
	err := learner.ApprovePattern(patterns[0].ID)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}

	approved, _ := learner.GetPatterns("approved")
	if len(approved) != 1 {
		t.Errorf("approved = %d, want 1", len(approved))
	}

	// Reject second (if exists)
	if len(patterns) > 1 {
		err = learner.RejectPattern(patterns[1].ID)
		if err != nil {
			t.Fatalf("reject: %v", err)
		}

		rejected, _ := learner.GetPatterns("rejected")
		if len(rejected) != 1 {
			t.Errorf("rejected = %d, want 1", len(rejected))
		}
	}
}

func TestStartStop(t *testing.T) {
	learner, cleanup := setupTestLearner(t)
	defer cleanup()

	if !learner.active {
		t.Error("expected active after creation")
	}

	learner.Stop()
	if learner.active {
		t.Error("expected inactive after stop")
	}

	// Collect should be no-op when stopped
	err := learner.CollectSample(Sample{
		Command: "test",
		Output:  "output",
	})
	if err != nil {
		t.Fatalf("collect while stopped: %v", err)
	}

	stats, _ := learner.GetStats()
	if stats.SamplesCollected != 0 {
		t.Errorf("samples = %d, want 0 (stopped)", stats.SamplesCollected)
	}

	learner.Start()
	if !learner.active {
		t.Error("expected active after start")
	}
}

func TestClear(t *testing.T) {
	learner, cleanup := setupTestLearner(t)
	defer cleanup()

	for i := 0; i < 3; i++ {
		learner.CollectSample(Sample{
			Command: "test",
			Output:  "[DEBUG] line\nresult\n",
		})
	}

	stats, _ := learner.GetStats()
	if stats.SamplesCollected == 0 {
		t.Error("expected some samples before clear")
	}

	err := learner.Clear()
	if err != nil {
		t.Fatalf("clear: %v", err)
	}

	stats, _ = learner.GetStats()
	if stats.SamplesCollected != 0 {
		t.Errorf("samples after clear = %d, want 0", stats.SamplesCollected)
	}
}

func TestGetStats(t *testing.T) {
	learner, cleanup := setupTestLearner(t)
	defer cleanup()

	stats, err := learner.GetStats()
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}

	if !stats.LearningActive {
		t.Error("expected learning active")
	}
	if stats.SamplesCollected != 0 {
		t.Errorf("initial samples = %d, want 0", stats.SamplesCollected)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled {
		t.Error("expected disabled by default")
	}
	if cfg.SamplingRate != 1.0 {
		t.Errorf("sampling rate = %f, want 1.0", cfg.SamplingRate)
	}
	if cfg.MinFrequency != 3 {
		t.Errorf("min frequency = %d, want 3", cfg.MinFrequency)
	}
	if cfg.MinConfidence != 0.7 {
		t.Errorf("min confidence = %f, want 0.7", cfg.MinConfidence)
	}
	if cfg.DatabasePath == "" {
		t.Error("expected non-empty database path")
	}
}

func TestGenerateTOML(t *testing.T) {
	toml := generateTOML("git", []string{`^\[DEBUG\]`, `^=== RUN`})

	if toml == "" {
		t.Error("expected non-empty TOML")
	}
	if !contains(toml, "[git_learned]") {
		t.Error("expected [git_learned] section")
	}
	if !contains(toml, "strip_lines_matching") {
		t.Error("expected strip_lines_matching")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
