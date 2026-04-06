package analysis

import (
	"strings"
	"testing"

	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/toml"
)

func BenchmarkPipelineProcess(b *testing.B) {
	input := strings.Repeat(
		"   Compiling package\n    Checking module\n    Finished in 2.5s\nerror[E0425]: cannot find function `foo`\nwarning: unused variable x\n",
		100)

	cfg := filter.PipelineConfig{
		Mode: filter.ModeMinimal,
	}
	pipeline := filter.NewPipelineCoordinator(cfg)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		pipeline.Process(input)
	}
}

func BenchmarkTOMLFilterApply(b *testing.B) {
	input := `diff --git a/main.go b/main.go
index abc123..def456 100644
--- a/main.go
+++ b/main.go
@@ -1,5 +1,8 @@
 package main

+import "fmt"
+
 func main() {
-	fmt.Println("hello")
+	fmt.Println("hello world")
+	fmt.Println("added line")
 }
`
	rule := toml.TOMLFilterRule{
		StripLinesMatching: []string{"^index ", "^--- a/", "^\\+\\+\\+ b/"},
	}

	engine := toml.NewTOMLFilterEngine(&rule)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		engine.Apply(input, filter.ModeNone)
	}
}

func BenchmarkTOMLFilterLargeInput(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString("   Compiling myapp v0.1.0\n    Checking serde v1.0.0\n    Finished dev [unoptimized] in 2.15s\nerror[E0425]: cannot find function `foo`\nwarning: unused import `std::io`\n")
	}
	largeInput := sb.String()

	rule := toml.TOMLFilterRule{
		StripLinesMatching: []string{"^   Compiling ", "^    Checking ", "^    Finished "},
		KeepLinesMatching:  []string{"^error", "^warning"},
	}

	engine := toml.NewTOMLFilterEngine(&rule)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		engine.Apply(largeInput, filter.ModeNone)
	}
}

func BenchmarkFilterLoadAll(b *testing.B) {
	parser := toml.NewParser()
	for i := 0; i < b.N; i++ {
		// Parse a single filter file
		_, err := parser.ParseFile("builtin/docker.toml")
		if err != nil {
			b.Fatal(err)
		}
	}
}
