package filter

import (
	"testing"
)

func TestNewCoreFilters(t *testing.T) {
	cfg := PipelineConfig{
		EnableEntropy:      true,
		EnablePerplexity:   true,
		EnableGoalDriven:   true,
		EnableAST:          true,
		EnableContrastive:  true,
		EnableEvaluator:    true,
		EnableGist:         true,
		EnableHierarchical: true,
		NgramEnabled:       true,
		QueryIntent:        "debug",
	}
	cf := NewCoreFilters(cfg)
	if cf == nil {
		t.Fatal("expected non-nil CoreFilters")
	}
	if cf.entropy == nil {
		t.Error("expected entropy filter initialized")
	}
	if cf.perplexity == nil {
		t.Error("expected perplexity filter initialized")
	}
	if cf.goalDriven == nil {
		t.Error("expected goalDriven filter initialized")
	}
	if cf.ast == nil {
		t.Error("expected ast filter initialized")
	}
	if cf.contrastive == nil {
		t.Error("expected contrastive filter initialized")
	}
	if cf.ngram == nil {
		t.Error("expected ngram filter initialized")
	}
	if cf.evaluator == nil {
		t.Error("expected evaluator filter initialized")
	}
	if cf.gist == nil {
		t.Error("expected gist filter initialized")
	}
	if cf.hierarchical == nil {
		t.Error("expected hierarchical filter initialized")
	}
}

func TestNewCoreFilters_Disabled(t *testing.T) {
	cfg := PipelineConfig{}
	cf := NewCoreFilters(cfg)
	if cf == nil {
		t.Fatal("expected non-nil CoreFilters")
	}
	if cf.entropy != nil {
		t.Error("expected entropy filter nil when disabled")
	}
	if cf.goalDriven != nil {
		t.Error("expected goalDriven nil when disabled and no query intent")
	}
}

func TestCoreFiltersApply(t *testing.T) {
	cfg := PipelineConfig{
		EnableEntropy: true,
		EnableAST:     true,
	}
	cf := NewCoreFilters(cfg)
	stats := &PipelineStats{LayerStats: make(map[string]LayerStat)}

	input := "func main() {\n\treturn 42\n}\n"
	output := cf.Apply(input, ModeMinimal, stats)
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestNewSemanticFilters(t *testing.T) {
	cfg := PipelineConfig{
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
	}
	sf := NewSemanticFilters(cfg)
	if sf == nil {
		t.Fatal("expected non-nil SemanticFilters")
	}
}

func TestSemanticFiltersApply(t *testing.T) {
	cfg := PipelineConfig{EnableH2O: true}
	sf := NewSemanticFilters(cfg)
	stats := &PipelineStats{LayerStats: make(map[string]LayerStat)}

	input := "line1\nline2\nline3\n"
	output := sf.Apply(input, ModeMinimal, stats)
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestRefactoredCoordinatorProcess(t *testing.T) {
	cfg := PipelineConfig{Mode: ModeMinimal}
	rc := NewRefactoredCoordinator(cfg)
	if rc == nil {
		t.Fatal("expected non-nil RefactoredCoordinator")
	}

	input := "hello world"
	output, stats := rc.Process(input)
	if output == "" {
		t.Error("expected non-empty output")
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
}
