package summarizer

import (
	"context"
	"testing"
)

func TestSummarizationEngine(t *testing.T) {
	config := DefaultEngineConfig()
	engine := NewSummarizationEngine(config)

	if len(engine.algorithms) == 0 {
		t.Error("Expected algorithms to be registered")
	}

	algo, ok := engine.algorithms["textrank"]
	if !ok {
		t.Error("Expected textrank algorithm to exist")
	}

	if algo.Name() != "textrank" {
		t.Errorf("Expected textrank, got %s", algo.Name())
	}
}

func TestExtractiveSummarizer(t *testing.T) {
	summarizer := &ExtractiveSummarizer{}

	text := "This is the first sentence. This is the second important sentence. This is the third sentence."

	summary, err := summarizer.Summarize(context.Background(), text, SummarizeOptions{MaxLength: 100})
	if err != nil {
		t.Fatalf("Summarize failed: %v", err)
	}

	if summary.Algorithm != "extractive" {
		t.Errorf("Expected extractive, got %s", summary.Algorithm)
	}

	if summary.InputTokens == 0 {
		t.Error("Expected non-zero input tokens")
	}
}

func TestTFIDFSummarizer(t *testing.T) {
	summarizer := &TFIDFSummarizer{}

	text := "Machine learning is a subset of artificial intelligence. Deep learning is a subset of machine learning. Neural networks are used in deep learning."

	summary, err := summarizer.Summarize(context.Background(), text, SummarizeOptions{MaxLength: 100})
	if err != nil {
		t.Fatalf("Summarize failed: %v", err)
	}

	if summary.Algorithm != "tfidf" {
		t.Errorf("Expected tfidf, got %s", summary.Algorithm)
	}

	if summary.Reduction < 0 || summary.Reduction > 1 {
		t.Errorf("Expected reduction between 0 and 1, got %f", summary.Reduction)
	}
}

func TestTextRankSummarizer(t *testing.T) {
	summarizer := &TextRankSummarizer{}

	text := "The first sentence talks about cats. The second sentence also mentions cats. The third sentence discusses dogs. The fourth sentence is about dogs too."

	summary, err := summarizer.Summarize(context.Background(), text, SummarizeOptions{MaxLength: 100})
	if err != nil {
		t.Fatalf("Summarize failed: %v", err)
	}

	if summary.Algorithm != "textrank" {
		t.Errorf("Expected textrank, got %s", summary.Algorithm)
	}

	if summary.Confidence <= 0 || summary.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", summary.Confidence)
	}
}

func TestEngineStats(t *testing.T) {
	config := DefaultEngineConfig()
	engine := NewSummarizationEngine(config)

	text := "Test sentence one. Test sentence two. Test sentence three."

	engine.Summarize(context.Background(), text, SummarizeOptions{MaxLength: 50})

	stats := engine.GetStats()

	if stats.TotalSummaries != 1 {
		t.Errorf("Expected 1 summary, got %d", stats.TotalSummaries)
	}

	if stats.AlgorithmUsage["textrank"] != 1 {
		t.Errorf("Expected textrank to be used once")
	}
}
