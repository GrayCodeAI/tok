package templatepipes

import "testing"

func TestPipeEngine(t *testing.T) {
	engine := NewPipeEngine()

	result := engine.Execute("hello world", "upper")
	if result != "HELLO WORLD" {
		t.Errorf("Expected 'HELLO WORLD', got '%s'", result)
	}

	result = engine.Execute("hello world", "lower")
	if result != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", result)
	}

	result = engine.Execute("  hello  ", "trim")
	if result != "hello" {
		t.Errorf("Expected 'hello', got '%s'", result)
	}
}

func TestPipeEnginePipeline(t *testing.T) {
	engine := NewPipeEngine()

	result := engine.Execute("hello world", "upper | trim")
	if result != "HELLO WORLD" {
		t.Errorf("Expected 'HELLO WORLD', got '%s'", result)
	}
}

func TestPipeEngineListStages(t *testing.T) {
	engine := NewPipeEngine()
	stages := engine.ListStages()
	if len(stages) < 10 {
		t.Errorf("Expected at least 10 stages, got %d", len(stages))
	}
}
