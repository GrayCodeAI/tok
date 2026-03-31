package benchmarks

import (
	"strings"
	"testing"

	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/graph"
	"github.com/GrayCodeAI/tokman/internal/memory"
	"github.com/GrayCodeAI/tokman/internal/security"
)

func BenchmarkTDD(b *testing.B) {
	tdd := filter.NewTokenDenseDialect(filter.DefaultTDDConfig())
	input := strings.Repeat("function hello() { return true; }\n", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tdd.Encode(input)
	}
}

func BenchmarkTOON(b *testing.B) {
	encoder := filter.NewTOONEncoder(filter.DefaultTOONConfig())
	input := `[{"name":"file.go","size":100},{"name":"file.go","size":200},{"name":"file.go","size":300},{"name":"file.go","size":400},{"name":"file.go","size":500}]`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoder.Encode(input)
	}
}

func BenchmarkPhoton(b *testing.B) {
	pf := filter.NewPhotonFilter(filter.DefaultPhotonConfig())
	content := strings.Repeat(`<img src="data:image/png;base64,aGVsbG8gd29ybGQ=">`, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pf.Process(content)
	}
}

func BenchmarkLogCrunch(b *testing.B) {
	lc := filter.NewLogCrunch(filter.DefaultLogCrunchConfig())
	content := strings.Repeat("INFO: starting\n", 100) + "ERROR: failed"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Process(content)
	}
}

func BenchmarkDiffCrunch(b *testing.B) {
	dc := filter.NewDiffCrunch(filter.DefaultDiffCrunchConfig())
	content := strings.Repeat(" context line\n", 50) + "+new line\n" + strings.Repeat(" context line\n", 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dc.Process(content)
	}
}

func BenchmarkDictionaryEncoding(b *testing.B) {
	de := filter.NewDictionaryEncoding()
	content := strings.Repeat("This is a test. This is a test. ", 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		de.Encode(content)
	}
}

func BenchmarkInformationBottleneck(b *testing.B) {
	ib := filter.NewInformationBottleneck(filter.DefaultIBConfig())
	content := strings.Repeat("High entropy line with diverse characters and tokens.\n", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ib.Process(content, "diverse tokens")
	}
}

func BenchmarkFeedbackLoop(b *testing.B) {
	fl := filter.NewFeedbackLoop()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fl.Record("go", 0.8)
		fl.GetThreshold("go", 0.5)
	}
}

func BenchmarkEngramMemory(b *testing.B) {
	em := filter.NewEngramMemory(0.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em.Observe("observation", 0.8)
		em.TieredSummary()
	}
}

func BenchmarkReadModes(b *testing.B) {
	content := strings.Repeat("func main() {}\n\n// comment\ntype Foo struct{}\n", 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.ReadContent(content, filter.ReadOptions{Mode: filter.ReadMap})
	}
}

func BenchmarkTemplatePipe(b *testing.B) {
	pipe := filter.NewTemplatePipe("join | truncate 100 | lines 10")
	input := strings.Repeat("line of text\n", 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipe.Process(input)
	}
}

func BenchmarkKVCacheAlignment(b *testing.B) {
	aligner := filter.NewKVCacheAligner(filter.DefaultKVCacheConfig())
	content := strings.Repeat("You are a helpful assistant.\n", 10) + strings.Repeat("timestamp: 2024-01-01\n", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		aligner.AlignPrefix(content)
	}
}

func BenchmarkCrossMessageDedup(b *testing.B) {
	dedup := filter.NewCrossMessageDedup()
	content := strings.Repeat("This is repeated content across messages.\n", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dedup.DedupMessage(content)
	}
}

func BenchmarkSecurityScanner(b *testing.B) {
	s := security.NewScanner()
	content := "SELECT * FROM users WHERE id = 1; <script>alert('xss')</script>"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Scan(content)
	}
}

func BenchmarkSecretsDetection(b *testing.B) {
	content := "api_key = 'AKIAIOSFODNN7EXAMPLE123'\npassword = 'secret123'"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		security.DetectSecrets(content)
	}
}

func BenchmarkPIIRedaction(b *testing.B) {
	content := "Contact user@example.com or call 123-456-7890. SSN: 123-45-6789."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		security.RedactPII(content)
	}
}


func BenchmarkGraphAnalysis(b *testing.B) {
	g := graph.NewProjectGraph(".")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Stats()
	}
}

func BenchmarkMemoryStore(b *testing.B) {
	store := memory.NewMemoryStore("")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.AddTask("benchmark task", "tag")
		store.Query("task")
	}
}
