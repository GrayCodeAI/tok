package filter

import (
	"testing"
)

func TestLayerBitset_RoundTrip(t *testing.T) {
	cfg := PipelineConfig{
		EnableEntropy:          true,
		EnablePerplexity:       false,
		EnableGoalDriven:       true,
		EnableAST:              true,
		EnableContrastive:      false,
		EnableEvaluator:        true,
		EnableGist:             false,
		EnableHierarchical:     true,
		EnableCompaction:       true,
		EnableAttribution:      false,
		EnableH2O:              true,
		EnableAttentionSink:    false,
		EnableMetaToken:        true,
		EnableSemanticChunk:    false,
		EnableSketchStore:      true,
		EnableLazyPruner:       false,
		EnableSemanticAnchor:   true,
		EnableAgentMemory:      false,
		EnableEdgeCase:         true,
		EnableReasoning:        false,
		EnableAdvanced:         true,
		EnableQuantumLock:      false,
		EnablePhoton:           true,
		EnableAdaptiveLearning: false,
		EnableCrunchBench:      true,
	}

	bits := cfg.ToLayerBitset()
	restored := bits.ToConfig()

	if restored.EnableEntropy != cfg.EnableEntropy {
		t.Error("EnableEntropy mismatch")
	}
	if restored.EnablePerplexity != cfg.EnablePerplexity {
		t.Error("EnablePerplexity mismatch")
	}
	if restored.EnableGoalDriven != cfg.EnableGoalDriven {
		t.Error("EnableGoalDriven mismatch")
	}
	if restored.EnableAST != cfg.EnableAST {
		t.Error("EnableAST mismatch")
	}
	if restored.EnableEvaluator != cfg.EnableEvaluator {
		t.Error("EnableEvaluator mismatch")
	}
	if restored.EnableHierarchical != cfg.EnableHierarchical {
		t.Error("EnableHierarchical mismatch")
	}
	if restored.EnableCompaction != cfg.EnableCompaction {
		t.Error("EnableCompaction mismatch")
	}
	if restored.EnableH2O != cfg.EnableH2O {
		t.Error("EnableH2O mismatch")
	}
	if restored.EnableMetaToken != cfg.EnableMetaToken {
		t.Error("EnableMetaToken mismatch")
	}
	if restored.EnableSketchStore != cfg.EnableSketchStore {
		t.Error("EnableSketchStore mismatch")
	}
	if restored.EnableSemanticAnchor != cfg.EnableSemanticAnchor {
		t.Error("EnableSemanticAnchor mismatch")
	}
	if restored.EnableEdgeCase != cfg.EnableEdgeCase {
		t.Error("EnableEdgeCase mismatch")
	}
	if restored.EnableAdvanced != cfg.EnableAdvanced {
		t.Error("EnableAdvanced mismatch")
	}
	if restored.EnablePhoton != cfg.EnablePhoton {
		t.Error("EnablePhoton mismatch")
	}
	if restored.EnableCrunchBench != cfg.EnableCrunchBench {
		t.Error("EnableCrunchBench mismatch")
	}
}

func TestLayerBitset_Empty(t *testing.T) {
	cfg := PipelineConfig{}
	bits := cfg.ToLayerBitset()
	if bits != 0 {
		t.Errorf("expected empty bitset to be 0, got %d", bits)
	}

	restored := bits.ToConfig()
	if restored.EnableEntropy {
		t.Error("expected no flags set")
	}
}

func TestLayerBitset_AllEnabled(t *testing.T) {
	cfg := PipelineConfig{
		EnableEntropy:          true,
		EnablePerplexity:       true,
		EnableGoalDriven:       true,
		EnableAST:              true,
		EnableContrastive:      true,
		EnableEvaluator:        true,
		EnableGist:             true,
		EnableHierarchical:     true,
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
		EnableEdgeCase:         true,
		EnableReasoning:        true,
		EnableAdvanced:         true,
		EnableQuantumLock:      true,
		EnablePhoton:           true,
		EnableAdaptiveLearning: true,
		EnableCrunchBench:      true,
	}

	bits := cfg.ToLayerBitset()
	if bits == 0 {
		t.Error("expected non-zero bitset")
	}

	restored := bits.ToConfig()
	if !restored.EnableEntropy {
		t.Error("expected all flags restored")
	}
	if !restored.EnableCrunchBench {
		t.Error("expected CrunchBench restored")
	}
}
