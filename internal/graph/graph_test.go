package graph

import (
	"testing"
)

func TestIsStub(t *testing.T) {
	if !IsStub() {
		t.Error("expected IsStub() to be true")
	}
}

func TestProjectGraph(t *testing.T) {
	g := NewProjectGraph("/tmp/test")
	if g.Path != "/tmp/test" {
		t.Errorf("expected path '/tmp/test', got %q", g.Path)
	}

	if err := g.Analyze("shallow"); err != nil {
		t.Errorf("unexpected analyze error: %v", err)
	}

	stats := g.Stats()
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}

	related := g.FindRelatedFiles("main.go", 5)
	if related == nil {
		t.Fatal("expected non-nil related files slice")
	}
}

func TestFormatGraphStats(t *testing.T) {
	stats := map[string]interface{}{"nodes": 10}
	out := FormatGraphStats(stats)
	if out == "" {
		t.Error("expected non-empty formatted stats")
	}
}

func TestVisualize(t *testing.T) {
	out, err := Visualize("/tmp/test")
	if err != nil {
		t.Errorf("unexpected visualize error: %v", err)
	}
	if out != "" {
		t.Error("expected empty string from stub visualize")
	}
}
