package filter

import "testing"

func TestFeedbackLoop_Record(t *testing.T) {
	fl := NewFeedbackLoop()
	fl.Record("go", 0.8)
	fl.Record("go", 0.6)
	fl.Record("go", 0.9)
	threshold := fl.GetThreshold("go", 0.5)
	if threshold == 0 {
		t.Error("expected non-zero threshold")
	}
}

func TestFeedbackLoop_GetThreshold(t *testing.T) {
	fl := NewFeedbackLoop()
	base := 0.5
	threshold := fl.GetThreshold("unknown", base)
	if threshold != base {
		t.Errorf("expected base threshold %f, got %f", base, threshold)
	}
}

func TestInformationBottleneck_Process(t *testing.T) {
	ib := NewInformationBottleneck(DefaultIBConfig())
	content := "This is a high entropy line with diverse characters.\nLow.\nAnother diverse line with many unique tokens."
	result := ib.Process(content, "diverse")
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestInformationBottleneck_Disabled(t *testing.T) {
	cfg := DefaultIBConfig()
	cfg.Enabled = false
	ib := NewInformationBottleneck(cfg)
	content := "test content"
	result := ib.Process(content, "")
	if result != content {
		t.Error("expected unchanged content when disabled")
	}
}

func TestLineEntropy(t *testing.T) {
	highEntropy := "The quick brown fox jumps over the lazy dog"
	lowEntropy := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if lineEntropy(highEntropy) <= lineEntropy(lowEntropy) {
		t.Error("expected high entropy line to have higher entropy")
	}
}

func TestLineRelevance(t *testing.T) {
	line := "This line contains the word debug and error"
	relevance := lineRelevance(line, "debug error")
	if relevance <= 0 {
		t.Error("expected positive relevance")
	}
}
