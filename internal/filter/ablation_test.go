package filter

import "testing"

func TestLayerAblationBasic(t *testing.T) {
	input := "ERROR: test failed in handler.go:21\n" +
		"diff --git a/handler.go b/handler.go\n@@ -1,2 +1,2 @@\n"

	baseCfg := TierConfig(TierAdaptive, ModeMinimal)
	baseCfg.EnableQualityGuardrail = true

	run := func(cfg PipelineConfig) int {
		p := NewPipelineCoordinator(cfg)
		_, stats := p.Process(input)
		return stats.TotalSaved
	}

	baseSaved := run(baseCfg)
	if baseSaved < 0 {
		t.Fatalf("unexpected negative savings")
	}

	candidates := []struct {
		name string
		off  func(*PipelineConfig)
	}{
		{"tfidf", func(c *PipelineConfig) { c.EnableTFIDF = false }},
		{"h2o", func(c *PipelineConfig) { c.EnableH2O = false }},
		{"extractive", func(c *PipelineConfig) { c.EnableExtractivePrefilter = false }},
	}

	for _, c := range candidates {
		t.Run(c.name, func(t *testing.T) {
			cfg := baseCfg
			c.off(&cfg)
			saved := run(cfg)
			_ = baseSaved - saved // kept for future threshold-based assertions
		})
	}
}
