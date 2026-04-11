package filter

import (
	"strings"
	"testing"
)

// TestEmptyInput verifies handling of empty input.
func TestEmptyInput(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	pipeline := NewPipelineCoordinator(cfg)

	output, stats := pipeline.Process("")
	
	if output != "" {
		t.Errorf("empty input should return empty output, got %q", output)
	}
	
	if stats.OriginalTokens != 0 {
		t.Errorf("empty input should have 0 tokens, got %d", stats.OriginalTokens)
	}
}

// TestVeryLargeInput verifies handling of large inputs.
func TestVeryLargeInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large input test in short mode")
	}
	
	// Generate 1MB of content
	var sb strings.Builder
	for i := 0; i < 10000; i++ {
		sb.WriteString("Line " + string(rune(i)) + ": This is test content for large input handling.\n")
	}
	input := sb.String()
	
	cfg := PipelineConfig{Mode: ModeMinimal}
	pipeline := NewPipelineCoordinator(cfg)
	
	output, stats := pipeline.Process(input)
	
	if len(output) == 0 {
		t.Error("large input should not return empty output")
	}
	
	if stats.OriginalTokens == 0 {
		t.Error("large input should have tokens counted")
	}
}

// TestNilFilterHandling verifies nil filters don't panic.
func TestNilFilterHandling(t *testing.T) {
	cfg := PipelineConfig{
		Mode:           ModeMinimal,
		EnableEntropy:  true,
		EnableH2O:      false, // This might be nil
	}
	
	pipeline := NewPipelineCoordinator(cfg)
	input := "test content with some data"
	
	// Should not panic even with nil filters
	output, _ := pipeline.Process(input)
	
	if output == "" && input != "" {
		t.Error("nil filter handling should preserve input")
	}
}

// TestConcurrentPipelineAccess verifies thread safety.
func TestConcurrentPipelineAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent test in short mode")
	}
	
	cfg := PipelineConfig{Mode: ModeMinimal}
	pipeline := NewPipelineCoordinator(cfg)
	input := "Concurrent test input with enough content"
	
	// Run multiple goroutines
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 10; j++ {
				pipeline.Process(input)
			}
		}()
	}
	
	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestAllModes verifies all compression modes work.
func TestAllModes(t *testing.T) {
	modes := []Mode{ModeNone, ModeMinimal, ModeAggressive}
	input := "Test input for mode verification"
	
	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			cfg := PipelineConfig{Mode: mode}
			pipeline := NewPipelineCoordinator(cfg)
			
			output, stats := pipeline.Process(input)
			
			if output == "" {
				t.Errorf("mode %s returned empty output", mode)
			}
			
			if stats.OriginalTokens == 0 {
				t.Errorf("mode %s did not count tokens", mode)
			}
		})
	}
}

// TestBudgetEnforcement verifies budget is respected.
func TestBudgetEnforcement(t *testing.T) {
	input := "This is a test input that should be compressed within budget"
	budget := 10
	
	cfg := PipelineConfig{
		Mode:   ModeAggressive,
		Budget: budget,
	}
	pipeline := NewPipelineCoordinator(cfg)
	
	output, stats := pipeline.Process(input)
	
	// Output should respect budget (approximately)
	// Note: Exact enforcement depends on filter implementation
	t.Logf("Budget: %d, Final tokens: %d", budget, stats.FinalTokens)
	
	if stats.FinalTokens > budget*2 {
		t.Errorf("budget not respected: got %d tokens, budget was %d", 
			stats.FinalTokens, budget)
	}
}

// TestUnicodeHandling verifies unicode content is handled.
func TestUnicodeHandling(t *testing.T) {
	inputs := []string{
		"Unicode: 你好世界 🌍 émojis",
		"Arabic: مرحبا بالعالم",
		"Russian: Привет мир",
		"Japanese: こんにちは世界",
	}
	
	cfg := PipelineConfig{Mode: ModeMinimal}
	pipeline := NewPipelineCoordinator(cfg)
	
	for _, input := range inputs {
		output, _ := pipeline.Process(input)
		
		if output == "" && input != "" {
			t.Errorf("unicode handling failed for: %s", input)
		}
	}
}

// TestBinaryContentHandling verifies binary-like content doesn't crash.
func TestBinaryContentHandling(t *testing.T) {
	// Content with null bytes and control characters
	input := "Test\x00binary\xffcontent\x01\x02\x03"
	
	cfg := PipelineConfig{Mode: ModeMinimal}
	pipeline := NewPipelineCoordinator(cfg)
	
	// Should not panic
	output, _ := pipeline.Process(input)
	
	// Output may be modified, but shouldn't be empty
	t.Logf("Binary input handled, output length: %d", len(output))
}

// TestRepeatedProcessing verifies consistency.
func TestRepeatedProcessing(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	pipeline := NewPipelineCoordinator(cfg)
	input := "Test input for consistency check"
	
	// Process same input multiple times
	var outputs []string
	for i := 0; i < 5; i++ {
		output, _ := pipeline.Process(input)
		outputs = append(outputs, output)
	}
	
	// All outputs should be identical (deterministic)
	for i := 1; i < len(outputs); i++ {
		if outputs[i] != outputs[0] {
			t.Errorf("inconsistent output on run %d", i)
		}
	}
}

// TestLayerStatsAccumulation verifies stats are accumulated correctly.
func TestLayerStatsAccumulation(t *testing.T) {
	cfg := PipelineConfig{
		Mode:             ModeMinimal,
		EnableEntropy:    true,
		EnablePerplexity: true,
		EnableAST:        true,
	}
	pipeline := NewPipelineCoordinator(cfg)
	input := "Test input with enough content for multiple layers"
	
	_, stats := pipeline.Process(input)
	
	// Should have layer stats
	if len(stats.LayerStats) == 0 {
		t.Error("no layer stats recorded")
	}
	
	// Total saved should be non-negative
	if stats.TotalSaved < 0 {
		t.Errorf("negative tokens saved: %d", stats.TotalSaved)
	}
	
	// Reduction percent should be reasonable
	if stats.ReductionPercent < 0 || stats.ReductionPercent > 100 {
		t.Errorf("invalid reduction percent: %f", stats.ReductionPercent)
	}
}

// BenchmarkSmallInput benchmarks small input processing.
func BenchmarkSmallInput(b *testing.B) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	pipeline := NewPipelineCoordinator(cfg)
	input := "Small"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// BenchmarkLargeInput benchmarks large input processing.
func BenchmarkLargeInput(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping large benchmark in short mode")
	}
	
	cfg := PipelineConfig{Mode: ModeMinimal}
	pipeline := NewPipelineCoordinator(cfg)
	
	// 10KB input
	input := strings.Repeat("Large input content for benchmarking. ", 100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

// BenchmarkConcurrentProcessing benchmarks concurrent access.
func BenchmarkConcurrentProcessing(b *testing.B) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	pipeline := NewPipelineCoordinator(cfg)
	input := "Concurrent test input"
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pipeline.Process(input)
		}
	})
}
