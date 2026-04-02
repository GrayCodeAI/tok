package benchmarking

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewDSLParser(t *testing.T) {
	parser := NewDSLParser()
	if parser == nil {
		t.Fatal("expected parser to be created")
	}

	if parser.benchmarks == nil {
		t.Error("expected benchmarks slice to be initialized")
	}
}

func TestDSLParserParse(t *testing.T) {
	dsl := `
benchmark test-compression
  type = compression
  description = "Test compression benchmark"
  iterations = 50
  duration = 10s
  warmup = 2
  size = 2048
`

	parser := NewDSLParser()
	definitions, err := parser.Parse(dsl)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(definitions) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(definitions))
	}

	def := definitions[0]
	if def.Name != "test-compression" {
		t.Errorf("expected name 'test-compression', got %s", def.Name)
	}

	if def.Type != "compression" {
		t.Errorf("expected type 'compression', got %s", def.Type)
	}

	if def.Description != "Test compression benchmark" {
		t.Errorf("expected description 'Test compression benchmark', got %s", def.Description)
	}

	if def.Iterations != 50 {
		t.Errorf("expected 50 iterations, got %d", def.Iterations)
	}

	if def.Warmup != 2 {
		t.Errorf("expected 2 warmup, got %d", def.Warmup)
	}

	if def.Parameters["size"] != 2048 {
		t.Errorf("expected size 2048, got %v", def.Parameters["size"])
	}
}

func TestDSLParserParseMultiple(t *testing.T) {
	dsl := `
benchmark first
  type = compression
  size = 1024

benchmark second
  type = memory
  size = 2048
`

	parser := NewDSLParser()
	definitions, err := parser.Parse(dsl)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(definitions) != 2 {
		t.Fatalf("expected 2 definitions, got %d", len(definitions))
	}

	if definitions[0].Name != "first" {
		t.Errorf("expected first name 'first', got %s", definitions[0].Name)
	}

	if definitions[1].Name != "second" {
		t.Errorf("expected second name 'second', got %s", definitions[1].Name)
	}
}

func TestDSLParserParseComments(t *testing.T) {
	dsl := `
# This is a comment
benchmark test
  type = compression
  # Another comment
  size = 1024
`

	parser := NewDSLParser()
	definitions, err := parser.Parse(dsl)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(definitions) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(definitions))
	}
}

func TestDSLParserParseError(t *testing.T) {
	dsl := `
  type = compression
`

	parser := NewDSLParser()
	_, err := parser.Parse(dsl)

	if err == nil {
		t.Error("expected error for property without benchmark")
	}
}

func TestNewDSLBuilder(t *testing.T) {
	builder := NewDSLBuilder()
	if builder == nil {
		t.Fatal("expected builder to be created")
	}
}

func TestDSLBuilderAddBenchmark(t *testing.T) {
	builder := NewDSLBuilder()
	builder.AddBenchmark("test-bench").
		WithType("compression").
		WithDescription("Test benchmark").
		WithIterations(50).
		WithDuration(10*time.Second).
		WithWarmup(2).
		WithParameter("size", 2048)

	dsl := builder.Build()

	if dsl == "" {
		t.Error("expected non-empty DSL")
	}

	if !contains(dsl, "benchmark test-bench") {
		t.Error("expected benchmark declaration")
	}

	if !contains(dsl, "type = compression") {
		t.Error("expected type property")
	}

	if !contains(dsl, `"Test benchmark"`) {
		t.Error("expected description property")
	}
}

func TestDSLBuilderBuildMultiple(t *testing.T) {
	builder := NewDSLBuilder()
	builder.AddBenchmark("first").WithType("compression")
	builder.AddBenchmark("second").WithType("memory")

	dsl := builder.Build()

	if !contains(dsl, "benchmark first") {
		t.Error("expected first benchmark")
	}

	if !contains(dsl, "benchmark second") {
		t.Error("expected second benchmark")
	}
}

func TestBenchmarkDefinitionToBenchmark(t *testing.T) {
	tests := []struct {
		name      string
		def       BenchmarkDefinition
		wantType  string
		wantError bool
	}{
		{
			name: "compression",
			def: BenchmarkDefinition{
				Name:       "comp",
				Type:       "compression",
				Parameters: map[string]interface{}{"size": 1024},
			},
			wantType:  "*benchmarking.CompressionBenchmark",
			wantError: false,
		},
		{
			name: "memory",
			def: BenchmarkDefinition{
				Name:       "mem",
				Type:       "memory",
				Parameters: map[string]interface{}{"size": 1024},
			},
			wantType:  "*benchmarking.MemoryBenchmark",
			wantError: false,
		},
		{
			name: "concurrency",
			def: BenchmarkDefinition{
				Name:       "conc",
				Type:       "concurrency",
				Parameters: map[string]interface{}{"workers": 10, "tasks": 100},
			},
			wantType:  "*benchmarking.ConcurrencyBenchmark",
			wantError: false,
		},
		{
			name: "unknown",
			def: BenchmarkDefinition{
				Name: "unknown",
				Type: "unknown",
			},
			wantType:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			benchmark, err := tt.def.ToBenchmark()

			if tt.wantError {
				if err == nil {
					t.Error("expected error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if benchmark == nil {
				t.Error("expected benchmark")
			}
		})
	}
}

func TestStandardDSLTemplates(t *testing.T) {
	templates := StandardDSLTemplates()

	if len(templates) == 0 {
		t.Error("expected templates")
	}

	if _, ok := templates["quick"]; !ok {
		t.Error("expected 'quick' template")
	}

	if _, ok := templates["thorough"]; !ok {
		t.Error("expected 'thorough' template")
	}
}

func TestParseDSLFile(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "benchmark.dsl")

	content := `
benchmark test
  type = compression
  size = 1024
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	definitions, err := ParseDSLFile(tmpFile)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(definitions) != 1 {
		t.Errorf("expected 1 definition, got %d", len(definitions))
	}
}

func TestParseDSLFileNotFound(t *testing.T) {
	_, err := ParseDSLFile("/nonexistent/file.dsl")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func BenchmarkDSLParserParse(b *testing.B) {
	dsl := `
benchmark test-compression
  type = compression
  description = "Test compression benchmark"
  iterations = 50
  duration = 10s
  warmup = 2
  size = 2048
`

	parser := NewDSLParser()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(dsl)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsInternal(s, substr))
}

func containsInternal(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
