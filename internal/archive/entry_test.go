package archive

import (
	"testing"
	"time"
)

func TestNewArchiveEntry(t *testing.T) {
	content := []byte("test content")
	command := "echo test"

	entry := NewArchiveEntry(content, command)

	if entry.Command != command {
		t.Errorf("Command = %v, want %v", entry.Command, command)
	}

	if entry.OriginalSize != int64(len(content)) {
		t.Errorf("OriginalSize = %v, want %v", entry.OriginalSize, len(content))
	}

	if entry.Category != CategoryCommand {
		t.Errorf("Category = %v, want %v", entry.Category, CategoryCommand)
	}

	if entry.Compression != CompressionNone {
		t.Errorf("Compression = %v, want %v", entry.Compression, CompressionNone)
	}
}

func TestArchiveEntry_CompressionRatio(t *testing.T) {
	tests := []struct {
		name           string
		originalSize   int64
		compressedSize int64
		expectedRatio  float64
	}{
		{
			name:           "no compression",
			originalSize:   1000,
			compressedSize: 1000,
			expectedRatio:  1.0,
		},
		{
			name:           "50% compression",
			originalSize:   1000,
			compressedSize: 500,
			expectedRatio:  0.5,
		},
		{
			name:           "zero size",
			originalSize:   0,
			compressedSize: 0,
			expectedRatio:  1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &ArchiveEntry{
				OriginalSize:   tt.originalSize,
				CompressedSize: tt.compressedSize,
			}

			ratio := entry.CompressionRatio()
			if ratio != tt.expectedRatio {
				t.Errorf("CompressionRatio() = %v, want %v", ratio, tt.expectedRatio)
			}
		})
	}
}

func TestArchiveEntry_SpaceSaved(t *testing.T) {
	entry := &ArchiveEntry{
		OriginalSize:   1000,
		CompressedSize: 400,
	}

	saved := entry.SpaceSaved()
	if saved != 600 {
		t.Errorf("SpaceSaved() = %v, want 600", saved)
	}
}

func TestArchiveEntry_IsExpired(t *testing.T) {
	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := []struct {
		name     string
		expires  *time.Time
		expected bool
	}{
		{
			name:     "not expired",
			expires:  &future,
			expected: false,
		},
		{
			name:     "expired",
			expires:  &past,
			expected: true,
		},
		{
			name:     "no expiration",
			expires:  nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &ArchiveEntry{
				ExpiresAt: tt.expires,
			}

			if got := entry.IsExpired(); got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestArchiveEntry_MarkAccessed(t *testing.T) {
	entry := NewArchiveEntry([]byte("test"), "echo test")

	if entry.AccessCount != 0 {
		t.Error("Initial AccessCount should be 0")
	}

	if entry.AccessedAt != nil {
		t.Error("Initial AccessedAt should be nil")
	}

	entry.MarkAccessed()

	if entry.AccessCount != 1 {
		t.Errorf("AccessCount = %v, want 1", entry.AccessCount)
	}

	if entry.AccessedAt == nil {
		t.Error("AccessedAt should be set after MarkAccessed")
	}

	// Access again
	entry.MarkAccessed()

	if entry.AccessCount != 2 {
		t.Errorf("AccessCount = %v, want 2", entry.AccessCount)
	}
}

func TestArchiveEntry_Tags(t *testing.T) {
	entry := NewArchiveEntry([]byte("test"), "echo test")

	// Test AddTag
	entry.AddTag("test")
	if !entry.HasTag("test") {
		t.Error("Should have tag 'test' after AddTag")
	}

	// Test duplicate AddTag
	entry.AddTag("test")
	if len(entry.Tags) != 1 {
		t.Errorf("Should not add duplicate tag, got %d tags", len(entry.Tags))
	}

	// Test RemoveTag
	entry.RemoveTag("test")
	if entry.HasTag("test") {
		t.Error("Should not have tag 'test' after RemoveTag")
	}

	// Test RemoveTag on non-existent tag
	entry.RemoveTag("nonexistent") // Should not panic
}

func TestArchiveEntry_WithMethods(t *testing.T) {
	entry := NewArchiveEntry([]byte("test"), "echo test")

	// Test WithCategory
	entry.WithCategory(CategorySession)
	if entry.Category != CategorySession {
		t.Errorf("Category = %v, want %v", entry.Category, CategorySession)
	}

	// Test WithAgent
	entry.WithAgent("claude")
	if entry.Agent != "claude" {
		t.Errorf("Agent = %v, want 'claude'", entry.Agent)
	}

	// Test WithWorkingDirectory
	entry.WithWorkingDirectory("/home/user")
	if entry.WorkingDirectory != "/home/user" {
		t.Errorf("WorkingDirectory = %v, want '/home/user'", entry.WorkingDirectory)
	}

	// Test WithTags
	entry.WithTags("tag1", "tag2")
	if len(entry.Tags) != 2 {
		t.Errorf("Tags length = %v, want 2", len(entry.Tags))
	}

	// Test WithExpiration
	entry.WithExpiration(time.Hour)
	if entry.ExpiresAt == nil {
		t.Error("ExpiresAt should be set")
	}

	// Test WithMetadata
	entry.WithMetadata("key", "value")
	if entry.Metadata.Custom["key"] != "value" {
		t.Error("Custom metadata should be set")
	}
}

func TestArchiveEntry_Summary(t *testing.T) {
	entry := &ArchiveEntry{
		Hash:           "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		Command:        "echo hello world",
		OriginalSize:   1000,
		CompressedSize: 500,
	}

	summary := entry.Summary()

	// Should contain hash prefix
	if summary == "" {
		t.Error("Summary should not be empty")
	}

	// Should contain command (possibly truncated)
	if len(summary) < 10 {
		t.Error("Summary should be meaningful")
	}
}

func TestArchiveEntry_Metadata(t *testing.T) {
	entry := NewArchiveEntry([]byte("test"), "echo test")

	// Set some metadata
	entry.Metadata.UserID = "user123"
	entry.Metadata.Hostname = "myhost"
	entry.Metadata.GitBranch = "main"
	entry.Metadata.Environment = "dev"

	// Marshal
	jsonStr, err := entry.MarshalMetadata()
	if err != nil {
		t.Fatalf("MarshalMetadata() error = %v", err)
	}

	// Create new entry and unmarshal
	newEntry := NewArchiveEntry([]byte("test"), "echo test")
	err = newEntry.UnmarshalMetadata(jsonStr)
	if err != nil {
		t.Fatalf("UnmarshalMetadata() error = %v", err)
	}

	// Verify
	if newEntry.Metadata.UserID != "user123" {
		t.Errorf("UserID = %v, want 'user123'", newEntry.Metadata.UserID)
	}
	if newEntry.Metadata.Hostname != "myhost" {
		t.Errorf("Hostname = %v, want 'myhost'", newEntry.Metadata.Hostname)
	}
	if newEntry.Metadata.GitBranch != "main" {
		t.Errorf("GitBranch = %v, want 'main'", newEntry.Metadata.GitBranch)
	}
}

func TestArchiveEntry_ToMap(t *testing.T) {
	entry := NewArchiveEntry([]byte("test"), "echo test")
	entry.Hash = "abc123"
	entry.OriginalSize = 100
	entry.CompressedSize = 50

	m := entry.ToMap()

	// Check required fields
	if m["hash"] != "abc123" {
		t.Errorf("hash = %v, want 'abc123'", m["hash"])
	}

	if m["original_size"] != int64(100) {
		t.Errorf("original_size = %v, want 100", m["original_size"])
	}

	// Check computed fields
	if m["compression_ratio"] != 0.5 {
		t.Errorf("compression_ratio = %v, want 0.5", m["compression_ratio"])
	}

	if m["space_saved"] != int64(50) {
		t.Errorf("space_saved = %v, want 50", m["space_saved"])
	}
}

func TestDefaultListOptions(t *testing.T) {
	opts := DefaultListOptions()

	if opts.Limit != 100 {
		t.Errorf("Limit = %v, want 100", opts.Limit)
	}

	if opts.Offset != 0 {
		t.Errorf("Offset = %v, want 0", opts.Offset)
	}

	if opts.SortBy != "created_at" {
		t.Errorf("SortBy = %v, want 'created_at'", opts.SortBy)
	}

	if opts.SortOrder != "desc" {
		t.Errorf("SortOrder = %v, want 'desc'", opts.SortOrder)
	}
}

func TestArchiveEntryStats_CalculateAvgCompressionRatio(t *testing.T) {
	tests := []struct {
		name           string
		originalSize   int64
		compressedSize int64
		expectedRatio  float64
	}{
		{
			name:           "no compression",
			originalSize:   1000,
			compressedSize: 1000,
			expectedRatio:  1.0,
		},
		{
			name:           "50% compression",
			originalSize:   1000,
			compressedSize: 500,
			expectedRatio:  0.5,
		},
		{
			name:           "zero size",
			originalSize:   0,
			compressedSize: 0,
			expectedRatio:  1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &ArchiveEntryStats{
				TotalOriginalSize:   tt.originalSize,
				TotalCompressedSize: tt.compressedSize,
			}

			stats.CalculateAvgCompressionRatio()

			if stats.AvgCompressionRatio != tt.expectedRatio {
				t.Errorf("AvgCompressionRatio = %v, want %v", stats.AvgCompressionRatio, tt.expectedRatio)
			}
		})
	}
}

func BenchmarkNewArchiveEntry(b *testing.B) {
	content := []byte("benchmark content for archive entry creation")
	command := "echo benchmark"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewArchiveEntry(content, command)
	}
}
