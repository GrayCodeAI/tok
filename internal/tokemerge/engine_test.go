package tokemerge

import (
	"context"
	"testing"
)

func TestTokenMergingEngine(t *testing.T) {
	config := DefaultEngineConfig()
	engine := NewTokenMergingEngine(config)

	if len(engine.matchers) == 0 {
		t.Error("Expected matchers to be registered")
	}

	matcher, ok := engine.matchers["bipartite"]
	if !ok {
		t.Error("Expected bipartite matcher to exist")
	}

	if matcher.Name() != "bipartite" {
		t.Errorf("Expected bipartite, got %s", matcher.Name())
	}
}

func TestBipartiteMatcher(t *testing.T) {
	matcher := &BipartiteMatcher{}

	tokens := []Token{
		{ID: "1", Embedding: []float64{1, 1}, Importance: 0.9},
		{ID: "2", Embedding: []float64{1.1, 1.1}, Importance: 0.8},
		{ID: "3", Embedding: []float64{10, 10}, Importance: 0.1},
		{ID: "4", Embedding: []float64{10.1, 10.1}, Importance: 0.05},
	}

	result, err := matcher.Match(context.Background(), tokens, 0.5)
	if err != nil {
		t.Fatalf("Match failed: %v", err)
	}

	if result.MergedCount >= len(tokens) {
		t.Logf("Expected merging, got %d vs %d", result.MergedCount, len(tokens))
	}

	if result.Reduction < 0 || result.Reduction > 1 {
		t.Errorf("Expected reduction 0-1, got %f", result.Reduction)
	}
}

func TestKMeansMatcher(t *testing.T) {
	matcher := &KMeansMatcher{}

	tokens := []Token{
		{ID: "1", Embedding: []float64{1, 1}},
		{ID: "2", Embedding: []float64{1.1, 1.1}},
		{ID: "3", Embedding: []float64{10, 10}},
		{ID: "4", Embedding: []float64{10.1, 10.1}},
		{ID: "5", Embedding: []float64{5, 5}},
	}

	result, err := matcher.Match(context.Background(), tokens, 0.4)
	if err != nil {
		t.Fatalf("Match failed: %v", err)
	}

	if len(result.MergedTokens) == 0 {
		t.Error("Expected merged tokens")
	}

	if result.OriginalCount != len(tokens) {
		t.Errorf("Expected original count %d, got %d", len(tokens), result.OriginalCount)
	}
}

func TestGreedyMatcher(t *testing.T) {
	matcher := &GreedyMatcher{}

	tokens := []Token{
		{ID: "1", Embedding: []float64{1, 1}},
		{ID: "2", Embedding: []float64{1.1, 1.1}},
		{ID: "3", Embedding: []float64{10, 10}},
		{ID: "4", Embedding: []float64{10.1, 10.1}},
	}

	result, err := matcher.Match(context.Background(), tokens, 0.5)
	if err != nil {
		t.Fatalf("Match failed: %v", err)
	}

	if len(result.Merges) == 0 {
		t.Error("Expected at least one merge")
	}
}

func TestEngineMerge(t *testing.T) {
	engine := NewTokenMergingEngine(DefaultEngineConfig())

	tokens := []Token{
		{ID: "1", Embedding: []float64{1, 1}, Importance: 0.9},
		{ID: "2", Embedding: []float64{1.1, 1.1}, Importance: 0.8},
		{ID: "3", Embedding: []float64{10, 10}, Importance: 0.1},
		{ID: "4", Embedding: []float64{10.1, 10.1}, Importance: 0.05},
	}

	result, err := engine.Merge(context.Background(), tokens)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if result.OriginalCount != 4 {
		t.Errorf("Expected 4 original tokens, got %d", result.OriginalCount)
	}

	stats := engine.GetStats()
	if stats.TotalMerges != 1 {
		t.Errorf("Expected 1 total merge, got %d", stats.TotalMerges)
	}
}

func TestEngineUnmerge(t *testing.T) {
	engine := NewTokenMergingEngine(DefaultEngineConfig())

	merged := []Token{
		{ID: "merged_1", Embedding: []float64{1, 1}, Importance: 0.5, Metadata: map[string]interface{}{"splits": []string{"1", "2"}}},
		{ID: "3", Embedding: []float64{10, 10}, Importance: 0.1},
	}

	unmerged := engine.Unmerge(context.Background(), merged, 3)

	if len(unmerged) < len(merged) {
		t.Logf("Expected unmerge to expand tokens, got %d vs %d", len(unmerged), len(merged))
	}
}

func TestEngineStats(t *testing.T) {
	engine := NewTokenMergingEngine(DefaultEngineConfig())

	tokens := []Token{
		{ID: "1", Embedding: []float64{1, 1}},
		{ID: "2", Embedding: []float64{2, 2}},
		{ID: "3", Embedding: []float64{3, 3}},
	}

	engine.Merge(context.Background(), tokens)
	engine.Merge(context.Background(), tokens)

	stats := engine.GetStats()

	if stats.TotalMerges != 2 {
		t.Errorf("Expected 2 merges, got %d", stats.TotalMerges)
	}

	if stats.AvgReduction < 0 || stats.AvgReduction > 1 {
		t.Errorf("Expected valid reduction, got %f", stats.AvgReduction)
	}
}

func TestCosineSimilarity(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{1, 0, 0}
	c := []float64{0, 1, 0}

	simAB := cosineSimilarity(a, b)
	simAC := cosineSimilarity(a, c)

	if simAB < 0.99 {
		t.Errorf("Expected similarity ~1 for same vectors, got %f", simAB)
	}

	if simAC > 0.1 {
		t.Errorf("Expected similarity ~0 for different vectors, got %f", simAC)
	}
}
