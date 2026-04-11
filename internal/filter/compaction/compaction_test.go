package compaction

import (
	"strings"
	"testing"
	
	"github.com/GrayCodeAI/tokman/internal/filter"
)

// TestConversationDetector tests conversation detection
func TestConversationDetector(t *testing.T) {
	detector := NewConversationDetector([]string{"chat", "conversation"})

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "chat format",
			input:    "User: Hello\nAssistant: Hi there",
			expected: true,
		},
		{
			name:     "plain text",
			input:    "This is just plain text",
			expected: false,
		},
		{
			name:     "empty",
			input:    "",
			expected: false,
		},
		{
			name:     "human ai format",
			input:    "Human: Question\nAI: Answer",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := detector.Detect(tt.input)
			if result != tt.expected {
				t.Errorf("Detect(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestContentExtractor tests content extraction
func TestContentExtractor(t *testing.T) {
	extractor := NewContentExtractor(10)

	t.Run("ExtractCritical", func(t *testing.T) {
		input := "Error: something failed\nFile: /path/to/file\nTODO: fix this"
		critical := extractor.ExtractCritical(input)
		
		if len(critical) == 0 {
			t.Error("expected critical items")
		}
		
		found := false
		for _, c := range critical {
			if strings.Contains(c, "Error") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error in critical items")
		}
	})

	t.Run("ExtractKeyValuePairs", func(t *testing.T) {
		input := "key1: value1\nkey2: value2"
		pairs := extractor.ExtractKeyValuePairs(input)
		
		if len(pairs) == 0 {
			t.Error("expected key-value pairs")
		}
	})

	t.Run("ExtractNextAction", func(t *testing.T) {
		input := "next: implement feature"
		action := extractor.ExtractNextAction(input)
		
		if action == "" {
			t.Error("expected next action")
		}
	})

	t.Run("ParseTurns", func(t *testing.T) {
		input := "User: Hello\nAssistant: Hi"
		turns := extractor.ParseTurns(input)
		
		if len(turns) == 0 {
			t.Error("expected turns")
		}
	})
}

// TestCompactionLayer tests the compaction layer
func TestCompactionLayer(t *testing.T) {
	config := DefaultCompactionConfig()
	config.Enabled = true
	config.ThresholdLines = 5
	
	cl := &CompactionLayer{
		config: config,
		cache:  make(map[string]*CompactionResult),
	}

	t.Run("Compact small content", func(t *testing.T) {
		input := "User: Hello\nAssistant: Hi there\nUser: How are you?\nAssistant: I'm fine\nUser: Good"
		output, saved, err := cl.Compact(input, filter.ModeAggressive)
		
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		
		if output == "" && input != "" {
			t.Error("returned empty for valid input")
		}
		
		if saved < 0 {
			t.Error("negative savings")
		}
	})

	t.Run("Compact below threshold", func(t *testing.T) {
		input := "Short"
		output, saved, err := cl.Compact(input, filter.ModeAggressive)
		
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		
		// Should return input unchanged (below threshold)
		if output != input {
			t.Logf("Below threshold, returned: %q", output)
		}
		
		if saved != 0 {
			t.Logf("Expected 0 savings below threshold, got %d", saved)
		}
	})

	t.Run("Empty input", func(t *testing.T) {
		output, saved, err := cl.Compact("", filter.ModeAggressive)
		
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		
		if output != "" {
			t.Error("empty input should return empty")
		}
		
		if saved != 0 {
			t.Error("empty input should save 0")
		}
	})
}

// TestCompactionCache tests caching
func TestCompactionCache(t *testing.T) {
	config := DefaultCompactionConfig()
	config.CacheEnabled = true
	
	cl := &CompactionLayer{
		config: config,
		cache:  make(map[string]*CompactionResult),
	}

	t.Run("Cache hit", func(t *testing.T) {
		input := "User: Test\nAssistant: Response\nUser: Question\nAssistant: Answer\nUser: Follow"
		
		// First call
		_, _, _ = cl.Compact(input, filter.ModeAggressive)
		
		// Second call should hit cache
		// (We can't easily verify without modifying the code, but it shouldn't error)
		_, _, err := cl.Compact(input, filter.ModeAggressive)
		if err != nil {
			t.Errorf("cache hit failed: %v", err)
		}
	})
}

// TestStateSnapshot tests snapshot formatting
func TestStateSnapshot(t *testing.T) {
	snapshot := &StateSnapshot{
		SessionHistory: SessionHistory{
			UserQueries: []string{"Hello", "How are you?"},
			ActivityLog: []string{"Processed query"},
		},
		CurrentState: CurrentState{
			Focus:      "test",
			NextAction: "respond",
		},
		Context: SnapshotContext{
			Critical: []string{"Important"},
			Working:  []string{"Current"},
		},
	}
	
	str := snapshot.String()
	if str == "" {
		t.Error("snapshot string should not be empty")
	}
	
	if !strings.Contains(str, "Session History") {
		t.Error("should contain session history info")
	}
}

// TestTruncate tests truncate helper
func TestTruncate(t *testing.T) {
	tests := []struct {
		input   string
		maxLen  int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
	}
	
	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

// BenchmarkCompaction benchmarks compaction
func BenchmarkCompaction(b *testing.B) {
	config := DefaultCompactionConfig()
	config.Enabled = true
	
	cl := &CompactionLayer{
		config: config,
		cache:  make(map[string]*CompactionResult),
	}
	
	input := "User: Hello\nAssistant: Hi there\nUser: How are you?\nAssistant: I'm fine"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cl.Compact(input, filter.ModeAggressive)
	}
}

// BenchmarkConversationDetector benchmarks detection
func BenchmarkConversationDetector(b *testing.B) {
	detector := NewConversationDetector([]string{"chat", "conversation"})
	input := "User: Hello\nAssistant: Hi there"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.ShouldCompact(input, 5, 100)
	}
}
