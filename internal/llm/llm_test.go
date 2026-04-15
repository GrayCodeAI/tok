package llm

import (
	"testing"
)

func TestNewSummarizer(t *testing.T) {
	s := NewSummarizer("openai", "gpt-4", "http://localhost:11434")
	if s == nil {
		t.Fatal("expected non-nil summarizer")
	}
	if s.Provider != "openai" {
		t.Errorf("expected provider='openai', got '%s'", s.Provider)
	}
	if s.Model != "gpt-4" {
		t.Errorf("expected model='gpt-4', got '%s'", s.Model)
	}
	if s.BaseURL != "http://localhost:11434" {
		t.Errorf("expected baseURL='http://localhost:11434', got '%s'", s.BaseURL)
	}
}

func TestNewSummarizerFromEnv(t *testing.T) {
	s := NewSummarizerFromEnv()
	if s == nil {
		t.Fatal("expected non-nil summarizer")
	}
	if s.Provider != "stub" {
		t.Errorf("expected provider='stub', got '%s'", s.Provider)
	}
	if s.Model != "stub" {
		t.Errorf("expected model='stub', got '%s'", s.Model)
	}
}

func TestSummarizer_Summarize(t *testing.T) {
	s := NewSummarizer("test", "test-model", "http://test")

	input := "This is a test text for summarization."
	result, err := s.Summarize(input, 100)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Stub returns input unchanged
	if result != input {
		t.Errorf("expected input to be returned unchanged, got '%s'", result)
	}
}

func TestSummarizer_SummarizeWithRequest(t *testing.T) {
	s := NewSummarizer("test", "test-model", "http://test")

	req := SummaryRequest{
		Text:      "Test content for summarization",
		Content:   "Additional content",
		Intent:    "summarize",
		MaxTokens: 100,
	}

	resp, err := s.SummarizeWithRequest(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Stub returns text unchanged
	if resp.Summary != req.Text {
		t.Errorf("expected summary='%s', got '%s'", req.Text, resp.Summary)
	}

	// Tokens should be length/4
	expectedTokens := len(req.Text) / 4
	if resp.Tokens != expectedTokens {
		t.Errorf("expected tokens=%d, got %d", expectedTokens, resp.Tokens)
	}
}

func TestSummarizer_IsAvailable(t *testing.T) {
	s := NewSummarizer("test", "test-model", "http://test")

	// Stub always returns false
	if s.IsAvailable() {
		t.Error("expected IsAvailable() to return false for stub")
	}
}

func TestSummaryRequest(t *testing.T) {
	req := SummaryRequest{
		Text:      "test text",
		Content:   "test content",
		Intent:    "test",
		MaxTokens: 50,
	}

	if req.Text != "test text" {
		t.Error("Text field not set correctly")
	}
	if req.MaxTokens != 50 {
		t.Error("MaxTokens field not set correctly")
	}
}

func TestSummaryResponse(t *testing.T) {
	resp := SummaryResponse{
		Summary: "test summary",
		Tokens:  100,
	}

	if resp.Summary != "test summary" {
		t.Error("Summary field not set correctly")
	}
	if resp.Tokens != 100 {
		t.Error("Tokens field not set correctly")
	}
}
