package pattern

import (
	"os"
	"testing"
	"time"
)

func setupTestEngine(t *testing.T) (*PatternDiscoveryEngine, func()) {
	tmpDir, err := os.MkdirTemp("", "pattern-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Override data path
	oldDataPath := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)

	engine, err := NewPatternDiscoveryEngine()
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create engine: %v", err)
	}

	cleanup := func() {
		engine.Close()
		os.RemoveAll(tmpDir)
		os.Setenv("XDG_DATA_HOME", oldDataPath)
	}

	return engine, cleanup
}

func TestNewPatternDiscoveryEngine(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	if engine == nil {
		t.Fatal("expected non-nil engine")
	}
	if engine.db == nil {
		t.Error("expected non-nil database")
	}
	if engine.patterns == nil {
		t.Error("expected initialized patterns map")
	}
	if engine.sampleQueue == nil {
		t.Error("expected initialized sample queue")
	}
	if engine.stopChan == nil {
		t.Error("expected initialized stop channel")
	}
	if engine.minFrequency != 5 {
		t.Errorf("expected minFrequency=5, got %d", engine.minFrequency)
	}
	if engine.minConfidence != 0.7 {
		t.Errorf("expected minConfidence=0.7, got %f", engine.minConfidence)
	}
}

func TestPatternDiscoveryEngine_StartStop(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	// Start should not panic
	engine.Start()

	if !engine.running {
		t.Error("expected engine to be running after Start()")
	}

	// Stop should not panic
	engine.Stop()

	if engine.running {
		t.Error("expected engine to not be running after Stop()")
	}

	// Multiple stops should not panic
	engine.Stop()
}

func TestPatternDiscoveryEngine_RestartAfterStop(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	engine.Start()
	engine.Stop()
	engine.Start()
	defer engine.Stop()

	if !engine.running {
		t.Fatal("expected engine to be running after restart")
	}

	engine.SubmitSample("2024-01-15 INFO restarted", "restart.log")
	time.Sleep(50 * time.Millisecond)
}

func TestPatternDiscoveryEngine_SubmitSample(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	engine.Start()
	defer engine.Stop()

	// Should not panic or block
	engine.SubmitSample("test content", "test.txt")
	engine.SubmitSample("2024-01-15 INFO Starting application", "app.log")
	engine.SubmitSample("error: connection failed", "error.log")

	// Give worker time to process
	time.Sleep(100 * time.Millisecond)
}

func TestPatternDiscoveryEngine_RecordPattern(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	// Record a pattern
	engine.recordPattern("test_type", "test_pattern", `test.*regex`, "source.txt")

	// Verify it was recorded
	engine.mu.RLock()
	count := len(engine.patterns)
	engine.mu.RUnlock()

	if count != 1 {
		t.Errorf("expected 1 pattern, got %d", count)
	}

	// Record same pattern again (should update frequency)
	engine.recordPattern("test_type", "test_pattern", `test.*regex`, "source.txt")

	engine.mu.RLock()
	for _, p := range engine.patterns {
		if p.Frequency != 2 {
			t.Errorf("expected frequency=2, got %d", p.Frequency)
		}
	}
	engine.mu.RUnlock()
}

func TestPatternDiscoveryEngine_CalculateConfidence(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	tests := []struct {
		frequency int
		minConf   float64
	}{
		{1, 0.5},
		{5, 0.7},
		{10, 0.9},
	}

	for _, tt := range tests {
		conf := engine.calculateConfidence(tt.frequency)
		if conf < 0 || conf > 1 {
			t.Errorf("confidence %f out of range [0,1] for frequency %d", conf, tt.frequency)
		}
	}
}

func TestPatternDiscoveryEngine_GetPatterns(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	// Add some patterns
	engine.recordPattern("type1", "pattern1", `regex1`, "source1")
	engine.recordPattern("type2", "pattern2", `regex2`, "source2")

	// Manually boost frequency to pass threshold
	engine.mu.Lock()
	for _, p := range engine.patterns {
		p.Frequency = 10
		p.Confidence = 0.9
	}
	engine.mu.Unlock()

	patterns := engine.GetPatterns(0.0)
	if len(patterns) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(patterns))
	}

	// Test with high confidence threshold
	patterns = engine.GetPatterns(0.95)
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns with high threshold, got %d", len(patterns))
	}
}

func TestPatternDiscoveryEngine_GetPatternByID(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	// Record pattern
	engine.recordPattern("test", "content", `regex`, "source")

	// Get pattern ID
	engine.mu.RLock()
	var patternID string
	for id := range engine.patterns {
		patternID = id
		break
	}
	engine.mu.RUnlock()

	// Retrieve it
	p, found := engine.GetPatternByID(patternID)
	if !found {
		t.Error("expected to find pattern")
	}
	if p == nil {
		t.Fatal("expected non-nil pattern")
	}
	if p.Type != "test" {
		t.Errorf("expected type='test', got '%s'", p.Type)
	}

	// Try non-existent
	_, found = engine.GetPatternByID("non-existent")
	if found {
		t.Error("expected not to find non-existent pattern")
	}
}

func TestPatternDiscoveryEngine_DeletePattern(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	// Record and get pattern ID
	engine.recordPattern("test", "content", `regex`, "source")

	engine.mu.RLock()
	var patternID string
	for id := range engine.patterns {
		patternID = id
		break
	}
	engine.mu.RUnlock()

	// Delete it
	err := engine.DeletePattern(patternID)
	if err != nil {
		t.Errorf("delete failed: %v", err)
	}

	// Verify deletion
	_, found := engine.GetPatternByID(patternID)
	if found {
		t.Error("expected pattern to be deleted")
	}
}

func TestDiscoveredPattern_GenerateFilter(t *testing.T) {
	tests := []struct {
		name     string
		pattern  DiscoveredPattern
		expected string
	}{
		{
			name: "timestamp pattern",
			pattern: DiscoveredPattern{
				Type:  "timestamp",
				Regex: `\d{4}-\d{2}-\d{2}`,
			},
			expected: "remove_lines_matching",
		},
		{
			name: "hash pattern",
			pattern: DiscoveredPattern{
				Type:  "hash_md5",
				Regex: `[a-f0-9]{32}`,
			},
			expected: "replace_pattern",
		},
		{
			name: "unknown pattern",
			pattern: DiscoveredPattern{
				Type:       "unknown",
				Pattern:    "test",
				Confidence: 0.8,
			},
			expected: "# Pattern:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pattern.GenerateFilter()
			if result == "" {
				t.Error("expected non-empty filter")
			}
			if len(result) < len(tt.expected) || result[:len(tt.expected)] != tt.expected {
				t.Errorf("expected filter to start with '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		slice []string
		item  string
		want  bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{[]string{"a"}, "a", true},
	}

	for _, tt := range tests {
		got := contains(tt.slice, tt.item)
		if got != tt.want {
			t.Errorf("contains(%v, %s) = %v, want %v", tt.slice, tt.item, got, tt.want)
		}
	}
}

func TestDetectPatterns(t *testing.T) {
	engine, cleanup := setupTestEngine(t)
	defer cleanup()

	testCases := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "log line with timestamp",
			content:  "2024-01-15 10:30:00 INFO Application started",
			expected: "timestamp",
		},
		{
			name:     "error line",
			content:  "ERROR: Connection timeout",
			expected: "error",
		},
		{
			name:     "HTTP request",
			content:  "GET /api/v1/users HTTP/1.1",
			expected: "http",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			engine.analyzeSample(&ContentSample{
				Content: tc.content,
				Source:  "test.log",
			})
		})
	}
}
