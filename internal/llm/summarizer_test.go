package llm

import "testing"

func TestBuildPromptTruncatesLargeContent(t *testing.T) {
	s := NewSummarizer(DefaultConfig())
	req := SummaryRequest{
		Content:   string(make([]byte, 9000)),
		MaxTokens: 100,
		Intent:    "debug",
	}

	prompt := s.buildPrompt(req)
	if len(prompt) == 0 {
		t.Fatal("expected prompt")
	}
	if len(prompt) > 10000 {
		t.Fatalf("prompt unexpectedly large: %d", len(prompt))
	}
}
