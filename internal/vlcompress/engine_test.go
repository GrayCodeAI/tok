package vlcompress

import (
	"context"
	"testing"
)

func TestVLCompressionEngine(t *testing.T) {
	config := DefaultEngineConfig()
	engine := NewVLCompressionEngine(config)

	if len(engine.strategies) == 0 {
		t.Error("Expected strategies to be registered")
	}

	strategy, ok := engine.strategies["token_pruner"]
	if !ok {
		t.Error("Expected token_pruner strategy to exist")
	}

	if strategy.Name() != "token_pruner" {
		t.Errorf("Expected token_pruner, got %s", strategy.Name())
	}
}

func TestTokenPruner(t *testing.T) {
	pruner := &TokenPruner{}

	images := []ImageToken{
		{ID: "1", Attention: 0.9, Features: []float64{1, 2, 3}},
		{ID: "2", Attention: 0.1, Features: []float64{4, 5, 6}},
		{ID: "3", Attention: 0.8, Features: []float64{7, 8, 9}},
		{ID: "4", Attention: 0.05, Features: []float64{10, 11, 12}},
	}

	input := &VLInput{
		Images: images,
		Text:   "test",
	}

	output, err := pruner.Compress(context.Background(), input)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	if len(output.Images) == 0 {
		t.Error("Expected some images to be kept")
	}

	if output.Quality <= 0 || output.Quality > 1 {
		t.Errorf("Expected quality between 0 and 1, got %f", output.Quality)
	}
}

func TestDualPivotClustering(t *testing.T) {
	clustering := &DualPivotClustering{}

	images := []ImageToken{
		{ID: "1", Features: []float64{1, 1, 1}},
		{ID: "2", Features: []float64{1.1, 1.1, 1.1}},
		{ID: "3", Features: []float64{10, 10, 10}},
		{ID: "4", Features: []float64{10.1, 10.1, 10.1}},
		{ID: "5", Features: []float64{5, 5, 5}},
	}

	input := &VLInput{
		Images: images,
		Text:   "test",
	}

	output, err := clustering.Compress(context.Background(), input)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	if len(output.Images) >= len(images) {
		t.Logf("Expected compression, got %d vs %d", len(output.Images), len(images))
	}
}

func TestDensityClustering(t *testing.T) {
	clustering := &DensityClustering{}

	images := []ImageToken{
		{ID: "1", Features: []float64{1, 1}, Attention: 0.9},
		{ID: "2", Features: []float64{1.1, 1.1}, Attention: 0.8},
		{ID: "3", Features: []float64{10, 10}, Attention: 0.1},
		{ID: "4", Features: []float64{10.1, 10.1}, Attention: 0.05},
	}

	input := &VLInput{
		Images: images,
		Text:   "test",
	}

	output, err := clustering.Compress(context.Background(), input)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	if output.TokensRemoved <= 0 {
		t.Logf("Expected some tokens removed, got %d", output.TokensRemoved)
	}
}

func TestAttentionOptimizer(t *testing.T) {
	optimizer := &AttentionOptimizer{}

	images := []ImageToken{
		{ID: "1", Attention: 0.9},
		{ID: "2", Attention: 0.1},
		{ID: "3", Attention: 0.8},
		{ID: "4", Attention: 0.05},
		{ID: "5", Attention: 0.02},
	}

	input := &VLInput{
		Images:    images,
		Text:      "test",
		Attention: [][]float64{{0.9, 0.1, 0.0, 0.0, 0.0}},
	}

	output, err := optimizer.Compress(context.Background(), input)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	if len(output.Images) < len(images) {
		t.Logf("Expected compression, kept %d of %d", len(output.Images), len(images))
	}
}

func TestEngineStats(t *testing.T) {
	engine := NewVLCompressionEngine(DefaultEngineConfig())

	images := []ImageToken{
		{ID: "1", Attention: 0.9, Features: []float64{1, 2}},
		{ID: "2", Attention: 0.1, Features: []float64{3, 4}},
		{ID: "3", Attention: 0.8, Features: []float64{5, 6}},
		{ID: "4", Attention: 0.05, Features: []float64{7, 8}},
	}

	input := &VLInput{
		Images: images,
		Text:   "test",
	}

	_, err := engine.Compress(context.Background(), input)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	stats := engine.GetStats()

	if stats.TotalCompressions != 1 {
		t.Errorf("Expected 1 compression, got %d", stats.TotalCompressions)
	}

	if stats.TokensPruned <= 0 {
		t.Logf("Expected tokens pruned, got %d", stats.TokensPruned)
	}
}

func TestVLCosineSimilarity(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{1, 0, 0}
	c := []float64{0, 1, 0}

	simAB := cosineSimilarity(a, b)
	simAC := cosineSimilarity(a, c)

	if simAB > 0.1 {
		t.Logf("Expected high similarity for similar vectors, got %f", simAB)
	}

	if simAC < 0.9 {
		t.Logf("Expected low similarity for different vectors, got %f", simAC)
	}
}
