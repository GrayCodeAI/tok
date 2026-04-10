package filter

import (
	"fmt"
	"strings"
)

// String returns a formatted summary of pipeline stats
func (s *PipelineStats) String() string {
	var sb strings.Builder

	sb.WriteString("╔════════════════════════════════════════════════════╗\n")
	sb.WriteString("║         Tokman 20-Layer Compression Stats          ║\n")
	sb.WriteString("╠════════════════════════════════════════════════════╣\n")
	sb.WriteString(fmt.Sprintf("║ Original:  %6d tokens                         ║\n", s.OriginalTokens))
	sb.WriteString(fmt.Sprintf("║ Final:     %6d tokens                         ║\n", s.FinalTokens))
	sb.WriteString(fmt.Sprintf("║ Saved:     %6d tokens (%.1f%%)                 ║\n", s.TotalSaved, s.ReductionPercent))
	sb.WriteString("╠════════════════════════════════════════════════════╣\n")
	sb.WriteString("║ Layer Breakdown:                                   ║\n")

	layerOrder := []string{
		"1_entropy", "2_perplexity", "3_goal_driven", "4_ast_preserve",
		"5_contrastive", "6_ngram", "7_evaluator", "8_gist", "9_hierarchical",
		"10_budget", "11_compaction", "12_attribution", "13_h2o", "14_attention_sink",
		"15_meta_token", "16_semantic_chunk", "17_sketch_store", "18_lazy_pruner",
		"19_semantic_anchor", "20_agent_memory",
	}

	for _, layer := range layerOrder {
		if stat, ok := s.LayerStats[layer]; ok && stat.TokensSaved > 0 {
			sb.WriteString(fmt.Sprintf("║   %-20s: %6d tokens saved     ║\n", layer, stat.TokensSaved))
		}
	}

	sb.WriteString("╚════════════════════════════════════════════════════╝\n")

	return sb.String()
}

// QuickProcess compresses input with default configuration
func QuickProcess(input string, mode Mode) (string, int) {
	cfg := PipelineConfig{
		Mode:                   mode,
		SessionTracking:        true,
		NgramEnabled:           true,
		EnableCompaction:       true,
		EnableAttribution:      true,
		EnableH2O:              true,
		EnableAttentionSink:    true,
		EnableMetaToken:        true,
		EnableSemanticChunk:    true,
		EnableSketchStore:      true,
		EnableLazyPruner:       true,
		EnableSemanticAnchor:   true,
		EnableAgentMemory:      true,
		EnableTFIDF:            true,
		EnableReasoningTrace:   true,
		EnableSymbolicCompress: true,
		EnablePhraseGrouping:   true,
		EnableNumericalQuant:   true,
		EnableDynamicRatio:     true,
		EnableHypernym:         true,
		EnableSemanticCache:    true,
		EnableScope:            true,
		EnableSmallKV:          true,
		EnableKVzip:            true,
	}
	p := NewPipelineCoordinator(cfg)
	output, stats := p.Process(input)
	return output, stats.TotalSaved
}
