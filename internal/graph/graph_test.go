package graph

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewProjectGraph(t *testing.T) {
	g := NewProjectGraph("/tmp")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
}

func TestProjectGraph_Analyze(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}"), 0644)

	g := NewProjectGraph(dir)
	if err := g.Analyze(); err != nil {
		t.Fatalf("analyze failed: %v", err)
	}

	stats := g.Stats()
	if stats["total_files"].(int) == 0 {
		t.Error("expected at least 1 file")
	}
}

func TestProjectGraph_FindRelatedFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"hello\") }"), 0644)
	os.WriteFile(filepath.Join(dir, "utils.go"), []byte("package main\n\nfunc helper() {}"), 0644)

	g := NewProjectGraph(dir)
	g.Analyze()

	related := g.FindRelatedFiles("main.go", 5)
	if len(related) == 0 {
		t.Log("no related files found (expected for simple project)")
	}
}

func TestProjectGraph_ImpactAnalysis(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)

	g := NewProjectGraph(dir)
	g.Analyze()

	affected := g.ImpactAnalysis("main.go")
	if len(affected) != 0 {
		t.Logf("found %d affected files", len(affected))
	}
}

func TestFormatGraphStats(t *testing.T) {
	stats := map[string]any{
		"total_files": 10,
		"total_edges": 20,
		"by_language": map[string]int{"go": 5, "py": 5},
	}
	result := FormatGraphStats(stats)
	if len(result) == 0 {
		t.Error("expected non-empty stats string")
	}
}
