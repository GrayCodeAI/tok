package fusionpipeline

import "testing"

func TestFusionEngine(t *testing.T) {
	e := NewFusionEngine()

	input := "system prompt here\nuser query here\nassistant response here\nfunction foo() { return bar }"
	output, result := e.Compress(input)

	if output == "" {
		t.Error("Expected non-empty output")
	}
	if result.OriginalTokens == 0 {
		t.Error("Expected non-zero original tokens")
	}
	if result.StagesRun < 2 {
		t.Errorf("Expected at least 2 stages to run, got %d", result.StagesRun)
	}
}

func TestFusionEngineEmpty(t *testing.T) {
	e := NewFusionEngine()
	_, result := e.Compress("")
	if result.StagesRun != 0 {
		t.Errorf("Expected 0 stages for empty input, got %d", result.StagesRun)
	}
}

func TestQuantumLockStage(t *testing.T) {
	s := NewQuantumLockStage()
	input := "system prompt\nsystem prompt\nduplicate line"
	output, saved := s.Apply(input)
	if saved <= 0 {
		t.Error("Expected some savings from deduplication")
	}
	_ = output
}

func TestCortexStage(t *testing.T) {
	s := NewCortexStage()
	longInput := "this is a very long string with many characters that should definitely exceed the minimum threshold of one hundred characters for the cortex stage to apply"
	if !s.ShouldApply(longInput) {
		t.Error("Should apply for long input")
	}
	if s.ShouldApply("short") {
		t.Error("Should not apply for short input")
	}
}

func TestPhotonStage(t *testing.T) {
	s := NewPhotonStage()
	if !s.ShouldApply("data:image/png;base64,abc") {
		t.Error("Should apply for base64 content")
	}
	if s.ShouldApply("regular text") {
		t.Error("Should not apply for regular text")
	}
}

func TestLogCrunchStage(t *testing.T) {
	s := NewLogCrunchStage()
	if !s.ShouldApply("INFO: something happened") {
		t.Error("Should apply for log output")
	}
}

func TestNeurosyntaxStage(t *testing.T) {
	s := NewNeurosyntaxStage()
	if !s.ShouldApply("func main() { return 0 }") {
		t.Error("Should apply for code")
	}
	if !s.ShouldApply("def foo(): pass") {
		t.Error("Should apply for Python code")
	}
}

func TestAllStages(t *testing.T) {
	stages := []FusionStage{
		NewQuantumLockStage(),
		NewCortexStage(),
		NewPhotonStage(),
		NewRLEStage(),
		NewSemanticDedupStage(),
		NewIonizerStage(),
		NewLogCrunchStage(),
		NewSearchCrunchStage(),
		NewDiffCrunchStage(),
		NewStructuralCollapseStage(),
		NewNeurosyntaxStage(),
		NewNexusStage(),
		NewTokenOptStage(),
		NewAbbrevStage(),
	}

	for _, stage := range stages {
		name := stage.Name()
		if name == "" {
			t.Error("Stage name should not be empty")
		}
	}
}
