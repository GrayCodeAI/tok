package filter

import (
	"testing"
)

func TestConversationDedup_DuplicateDetection(t *testing.T) {
	dedup := NewConversationDedup()

	messages := []Message{
		{Role: "user", Content: "Fix the login bug in the authentication module of auth.py file"},
		{Role: "assistant", Content: "I'll help you fix that issue"},
		{Role: "user", Content: "Fix the login bug in the authentication module of auth.py file"}, // Exact duplicate
	}

	result, stats := dedup.DeduplicateMessages(messages)

	if len(result) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(result))
	}

	if stats.MessagesDeduped != 1 {
		t.Errorf("Expected 1 deduped message, got %d", stats.MessagesDeduped)
	}

	if !containsString(result[2].Content, "similar to message 0") {
		t.Errorf("Expected reference to message 0, got: %s", result[2].Content)
	}
}

func TestConversationDedup_NoSimilarity(t *testing.T) {
	dedup := NewConversationDedup()

	messages := []Message{
		{Role: "user", Content: "Fix the login bug"},
		{Role: "assistant", Content: "Sure, I can help"},
		{Role: "user", Content: "What about the database schema?"},
	}

	result, stats := dedup.DeduplicateMessages(messages)

	if stats.MessagesDeduped != 0 {
		t.Errorf("Expected 0 deduped messages, got %d", stats.MessagesDeduped)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(result))
	}
}

func TestConversationDedup_ShortMessages(t *testing.T) {
	dedup := NewConversationDedup()

	messages := []Message{
		{Role: "user", Content: "Hi"},
		{Role: "assistant", Content: "Hello"},
		{Role: "user", Content: "Hi"}, // Too short to dedup
	}

	_, stats := dedup.DeduplicateMessages(messages)

	if stats.MessagesDeduped != 0 {
		t.Errorf("Expected 0 deduped (too short), got %d", stats.MessagesDeduped)
	}
}

func TestComputeShingles(t *testing.T) {
	text := "the quick brown fox jumps"
	shingles := computeShingles(text, 3)

	expected := []string{
		"the quick brown",
		"quick brown fox",
		"brown fox jumps",
	}

	if len(shingles) != len(expected) {
		t.Errorf("Expected %d shingles, got %d", len(expected), len(shingles))
	}

	for _, exp := range expected {
		if !shingles[exp] {
			t.Errorf("Missing shingle: %s", exp)
		}
	}
}

func TestJaccardSimilarity(t *testing.T) {
	a := map[string]bool{"a": true, "b": true, "c": true}
	b := map[string]bool{"b": true, "c": true, "d": true}

	sim := jaccardSimilarity(a, b)
	expected := 0.5 // 2 intersection / 4 union

	if sim != expected {
		t.Errorf("Expected similarity %.2f, got %.2f", expected, sim)
	}
}

func TestJaccardSimilarity_Identical(t *testing.T) {
	a := map[string]bool{"a": true, "b": true, "c": true}
	b := map[string]bool{"a": true, "b": true, "c": true}

	sim := jaccardSimilarity(a, b)

	if sim != 1.0 {
		t.Errorf("Expected similarity 1.0, got %.2f", sim)
	}
}

func TestJaccardSimilarity_NoOverlap(t *testing.T) {
	a := map[string]bool{"a": true, "b": true}
	b := map[string]bool{"c": true, "d": true}

	sim := jaccardSimilarity(a, b)

	if sim != 0.0 {
		t.Errorf("Expected similarity 0.0, got %.2f", sim)
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		(s == substr || len(s) >= len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
