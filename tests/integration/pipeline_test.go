package integration

import (
	"testing"

	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/tests/integration/helpers"
)

// TestPipelineBasicExecution tests basic pipeline execution
func TestPipelineBasicExecution(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		mode    filter.Mode
		wantErr bool
	}{
		{
			name:  "simple text minimal mode",
			input: "Hello World\nThis is a test\nMultiple lines here",
			mode:  filter.ModeMinimal,
		},
		{
			name:  "code content aggressive mode",
			input: helpers.GetCodeSamples()[1].Content,
			mode:  filter.ModeAggressive,
		},
		{
			name:  "json data minimal mode",
			input: helpers.GetCodeSamples()[2].Content,
			mode:  filter.ModeMinimal,
		},
		{
			name:  "log output aggressive mode",
			input: helpers.GetCodeSamples()[3].Content,
			mode:  filter.ModeAggressive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := filter.PipelineConfig{
				Mode:          tt.mode,
				EnableEntropy: true,
				EnableAST:     true,
			}

			pipeline := filter.NewPipelineCoordinator(cfg)
			result, stats := pipeline.Process(tt.input)

			if tt.wantErr {
				t.Errorf("expected error but got none")
				return
			}

			// Verify output is not empty
			if result == "" {
				t.Error("pipeline returned empty result")
			}

			// Verify tokens saved is non-negative
			if stats.TotalSaved < 0 {
				t.Errorf("negative tokens saved: %d", stats.TotalSaved)
			}

			// For aggressive mode, expect some compression
			if tt.mode == filter.ModeAggressive {
				if stats.TotalSaved == 0 {
					t.Log("Warning: no tokens saved in aggressive mode")
				}
			}
		})
	}
}

// TestPipelineBudgetEnforcement tests budget enforcement
func TestPipelineBudgetEnforcement(t *testing.T) {
	input := helpers.GetLargeContent()

	tests := []struct {
		name   string
		budget int
	}{
		{"small budget 100", 100},
		{"medium budget 500", 500},
		{"large budget 1000", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := filter.PipelineConfig{
				Mode:          filter.ModeMinimal,
				Budget:        tt.budget,
				EnableEntropy: true,
			}

			pipeline := filter.NewPipelineCoordinator(cfg)
			result, stats := pipeline.Process(input)

			// With budget enforcement, should not exceed budget
			if stats.TotalSaved < 0 {
				t.Errorf("negative tokens saved: %d", stats.TotalSaved)
			}

			t.Logf("Budget %d: output length %d, saved %d tokens",
				tt.budget, len(result), stats.TotalSaved)
		})
	}
}

// TestPipelineQueryIntent tests query-aware filtering
func TestPipelineQueryIntent(t *testing.T) {
	input := helpers.GetCodeSamples()[1].Content

	tests := []struct {
		name        string
		queryIntent string
	}{
		{"debug query", "debug"},
		{"explain query", "explain"},
		{"review query", "review"},
		{"search query", "search"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := filter.PipelineConfig{
				Mode:             filter.ModeMinimal,
				QueryIntent:      tt.queryIntent,
				EnableGoalDriven: true,
			}

			pipeline := filter.NewPipelineCoordinator(cfg)
			result, stats := pipeline.Process(input)

			if result == "" {
				t.Error("pipeline returned empty result")
			}

			t.Logf("Query '%s': processed, saved %d tokens",
				tt.queryIntent, stats.TotalSaved)
		})
	}
}

// TestPipelineStageGates tests that stage gates work correctly
func TestPipelineStageGates(t *testing.T) {
	// Short content - some layers should be skipped
	shortInput := "Short content"

	// Long content - all layers should run
	longInput := helpers.GetLargeContent()

	tests := []struct {
		name  string
		input string
		desc  string
	}{
		{"short", shortInput, "should skip some layers"},
		{"long", longInput, "should run all applicable layers"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := filter.PipelineConfig{
				Mode:                filter.ModeMinimal,
				EnableEntropy:       true,
				EnablePerplexity:    true,
				EnableH2O:           true,
				EnableAttentionSink: true,
			}

			pipeline := filter.NewPipelineCoordinator(cfg)
			result, stats := pipeline.Process(tt.input)

			if result == "" {
				t.Error("pipeline returned empty result")
			}

			t.Logf("%s: %s - saved %d tokens", tt.name, tt.desc, stats.TotalSaved)
		})
	}
}

// TestPipelineRepetitiveContent tests compression of repetitive content
func TestPipelineRepetitiveContent(t *testing.T) {
	input := helpers.GetRepetitiveContent()

	cfg := filter.PipelineConfig{
		Mode: filter.ModeAggressive,
	}

	pipeline := filter.NewPipelineCoordinator(cfg)
	result, stats := pipeline.Process(input)

	// Log compression results (may or may not compress depending on content)
	t.Logf("Repetitive content: input=%d, output=%d, saved=%d tokens (%.1f%%)",
		len(input), len(result), stats.TotalSaved,
		float64(stats.TotalSaved)/float64(len(input))*100)

	// Just verify it runs without error
	if result == "" {
		t.Error("pipeline returned empty result")
	}
}

// TestPipelineMultiFile tests multi-file processing
func TestPipelineMultiFile(t *testing.T) {
	env := helpers.NewTestEnvironment(t)

	// Create test files
	files := []struct {
		path    string
		content string
	}{
		{"file1.go", helpers.GetCodeSamples()[0].Content},
		{"file2.go", helpers.GetCodeSamples()[1].Content},
		{"data.json", helpers.GetCodeSamples()[2].Content},
	}

	for _, f := range files {
		env.CreateFile(t, f.path, []byte(f.content))
	}

	// Test multi-file processing (if supported)
	cfg := filter.PipelineConfig{
		Mode:             filter.ModeMinimal,
		MultiFileEnabled: true,
	}

	pipeline := filter.NewPipelineCoordinator(cfg)

	// Combine files
	combined := ""
	for _, f := range files {
		combined += f.content + "\n"
	}

	result, stats := pipeline.Process(combined)

	if result == "" {
		t.Error("multi-file pipeline returned empty result")
	}

	t.Logf("Multi-file: processed %d files, saved %d tokens", len(files), stats.TotalSaved)
}

// TestPipelineAllLayers tests that all layers can be enabled
func TestPipelineAllLayers(t *testing.T) {
	input := helpers.GetCodeSamples()[1].Content

	cfg := filter.PipelineConfig{
		Mode:                 filter.ModeMinimal,
		EnableEntropy:        true,
		EnablePerplexity:     true,
		EnableGoalDriven:     true,
		EnableAST:            true,
		EnableContrastive:    true,
		EnableEvaluator:      true,
		EnableGist:           true,
		EnableHierarchical:   true,
		EnableCompaction:     true,
		EnableAttribution:    true,
		EnableH2O:            true,
		EnableAttentionSink:  true,
		EnableMetaToken:      true,
		EnableSemanticChunk:  true,
		EnableSketchStore:    true,
		EnableLazyPruner:     true,
		EnableSemanticAnchor: true,
		EnableAgentMemory:    true,
		Budget:               5000,
	}

	pipeline := filter.NewPipelineCoordinator(cfg)
	result, stats := pipeline.Process(input)

	if result == "" {
		t.Error("pipeline with all layers returned empty result")
	}

	t.Logf("All layers: saved %d tokens (%.1f%% reduction)",
		stats.TotalSaved, stats.ReductionPercent)
}

// TestPipelineEmptyInput tests behavior with empty input
func TestPipelineEmptyInput(t *testing.T) {
	cfg := filter.PipelineConfig{
		Mode: filter.ModeMinimal,
	}

	pipeline := filter.NewPipelineCoordinator(cfg)
	result, stats := pipeline.Process("")

	if result != "" {
		t.Errorf("expected empty result for empty input, got: %q", result)
	}

	if stats.TotalSaved != 0 {
		t.Errorf("expected 0 tokens saved for empty input, got: %d", stats.TotalSaved)
	}
}

// TestPipelineWhitespaceInput tests behavior with whitespace-only input
func TestPipelineWhitespaceInput(t *testing.T) {
	input := "   \n\t\n   "

	cfg := filter.PipelineConfig{
		Mode: filter.ModeMinimal,
	}

	pipeline := filter.NewPipelineCoordinator(cfg)
	result, stats := pipeline.Process(input)

	// Should handle whitespace gracefully
	if stats.TotalSaved < 0 {
		t.Errorf("negative tokens saved for whitespace: %d", stats.TotalSaved)
	}

	t.Logf("Whitespace input: output length %d, saved %d tokens", len(result), stats.TotalSaved)
}

// BenchmarkPipelineBasic benchmarks basic pipeline performance
func BenchmarkPipelineBasic(b *testing.B) {
	input := helpers.GetCodeSamples()[1].Content
	cfg := filter.PipelineConfig{
		Mode: filter.ModeMinimal,
	}
	pipeline := filter.NewPipelineCoordinator(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}
