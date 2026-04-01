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
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "pkg", "helper"), 0755)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nimport \"example.com/demo/pkg/helper\"\n\nfunc main() { helper.Run() }"), 0644)
	os.WriteFile(filepath.Join(dir, "pkg", "helper", "helper.go"), []byte("package helper\n\nfunc Run() {}"), 0644)
	os.WriteFile(filepath.Join(dir, "utils.go"), []byte("package main\n\nfunc helper() {}"), 0644)

	g := NewProjectGraph(dir)
	g.Analyze()

	related := g.FindRelatedFiles("main.go", 5)
	if len(related) == 0 {
		t.Fatal("expected related files")
	}
	if related[0] != filepath.ToSlash(filepath.Join("pkg", "helper", "helper.go")) {
		t.Fatalf("expected helper.go to rank first, got %v", related)
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

func TestExtractGoSymbolsUsesAST(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	content := `package main

import "fmt"

type service struct{}

func (s *service) Run() {
	fmt.Println(helper())
}

func helper() string { return "ok" }
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	symbols, refs := extractSymbols(path, "go")
	if !contains(symbols, "Run") || !contains(symbols, "helper") || !contains(symbols, "service") {
		t.Fatalf("expected AST symbols, got %v", symbols)
	}
	if !contains(refs, "Println") {
		t.Fatalf("expected selector reference, got %v", refs)
	}
}

func TestExtractTypeScriptSymbols(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.ts")
	content := `export class Service {}
export async function boot() {}
const helper = () => "ok"
boot()
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	symbols, refs := extractSymbols(path, "typescript")
	if !contains(symbols, "Service") || !contains(symbols, "boot") || !contains(symbols, "helper") {
		t.Fatalf("expected TS symbols, got %v", symbols)
	}
	if !contains(refs, "boot") {
		t.Fatalf("expected TS reference, got %v", refs)
	}
}

func TestExtractPythonSymbols(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.py")
	content := `class Service:
    pass

async def run():
    return helper()

def helper():
    return "ok"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	symbols, refs := extractSymbols(path, "python")
	if !contains(symbols, "Service") || !contains(symbols, "run") || !contains(symbols, "helper") {
		t.Fatalf("expected Python symbols, got %v", symbols)
	}
	if !contains(refs, "helper") {
		t.Fatalf("expected Python reference, got %v", refs)
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
