package filter

import (
	"testing"
)

func TestModeConstants(t *testing.T) {
	if ModeNone != "none" {
		t.Error("ModeNone mismatch")
	}
	if ModeMinimal != "minimal" {
		t.Error("ModeMinimal mismatch")
	}
	if ModeAggressive != "aggressive" {
		t.Error("ModeAggressive mismatch")
	}
}

func TestLanguageConstants(t *testing.T) {
	langs := []Language{LangRust, LangPython, LangJavaScript, LangTypeScript, LangGo, LangC, LangCpp, LangJava, LangRuby, LangShell, LangSQL, LangUnknown}
	for _, l := range langs {
		if l == "" {
			t.Error("expected non-empty language constant")
		}
	}
}

func TestNewEngine(t *testing.T) {
	engine := NewEngine(ModeMinimal)
	if engine == nil {
		t.Fatal("expected non-nil Engine")
	}
}

func TestNewEngineWithQuery(t *testing.T) {
	engine := NewEngineWithQuery(ModeMinimal, "debug")
	if engine == nil {
		t.Fatal("expected non-nil Engine")
	}
}

func TestNewEngineWithConfig(t *testing.T) {
	cfg := EngineConfig{
		Mode:             ModeMinimal,
		QueryIntent:      "review",
		LLMEnabled:       false,
		MultiFileEnabled: true,
		PromptTemplate:   "default",
	}
	engine := NewEngineWithConfig(cfg)
	if engine == nil {
		t.Fatal("expected non-nil Engine")
	}
}

func TestEngine_Process(t *testing.T) {
	engine := NewEngine(ModeMinimal)
	input := "hello world\nthis is a test\n"
	output, saved := engine.Process(input)
	if output == "" {
		t.Error("expected non-empty output")
	}
	if saved < 0 {
		t.Error("expected non-negative saved")
	}
}

func TestEngine_Process_WithQuery(t *testing.T) {
	engine := NewEngineWithQuery(ModeMinimal, "debug")
	input := "func main() {\n\treturn 42\n}\n"
	output, saved := engine.Process(input)
	if output == "" {
		t.Error("expected non-empty output")
	}
	if saved < 0 {
		t.Error("expected non-negative saved")
	}
}

func TestEngine_SetMode(t *testing.T) {
	engine := NewEngine(ModeMinimal)
	engine.SetMode(ModeAggressive)
}

func TestIsCode(t *testing.T) {
	if !IsCode("func main() {}") {
		t.Error("expected true for Go code")
	}
	if !IsCode("def hello():\n    pass") {
		t.Error("expected true for Python code")
	}
	if IsCode("hello world") {
		t.Error("expected false for plain text")
	}
}

func TestEstimateTokens(t *testing.T) {
	if EstimateTokens("") != 0 {
		t.Error("expected 0 tokens for empty string")
	}
	if EstimateTokens("hello world") == 0 {
		t.Error("expected non-zero tokens for non-empty string")
	}
}
