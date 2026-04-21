package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/GrayCodeAI/tok/internal/filter"
)

type sample struct {
	name    string
	content string
}

type result struct {
	origTokens  int
	finalTokens int
	savedPct    float64
	latency     time.Duration
}

func main() {
	iterations := flag.Int("iterations", 5, "number of runs per sample")
	flag.Parse()

	samples := buildSamples()
	baselineCfg := filter.TierConfig(filter.TierTrim, filter.ModeMinimal)
	adaptiveCfg := filter.TierConfig(filter.TierAdaptive, filter.ModeMinimal)

	fmt.Println("# Adaptive Evaluation Report")
	fmt.Printf("\nGenerated: %s\n\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Printf("Iterations per sample: %d\n\n", *iterations)
	fmt.Println("| Sample | Baseline Saved % | Adaptive Saved % | Delta Saved % | Baseline Latency | Adaptive Latency | Latency Ratio |")
	fmt.Println("|---|---:|---:|---:|---:|---:|---:|")

	var sumBaseSaved float64
	var sumAdpSaved float64
	var sumBaseLatency time.Duration
	var sumAdpLatency time.Duration

	for _, s := range samples {
		base := runSample(s.content, baselineCfg, *iterations)
		adp := runSample(s.content, adaptiveCfg, *iterations)
		delta := adp.savedPct - base.savedPct
		latRatio := ratio(adp.latency, base.latency)

		sumBaseSaved += base.savedPct
		sumAdpSaved += adp.savedPct
		sumBaseLatency += base.latency
		sumAdpLatency += adp.latency

		fmt.Printf("| %s | %.2f | %.2f | %.2f | %s | %s | %.2fx |\n",
			s.name, base.savedPct, adp.savedPct, delta,
			base.latency.Round(time.Microsecond), adp.latency.Round(time.Microsecond), latRatio,
		)
	}

	n := float64(len(samples))
	avgBaseSaved := sumBaseSaved / n
	avgAdpSaved := sumAdpSaved / n
	avgDelta := avgAdpSaved - avgBaseSaved
	avgBaseLatency := sumBaseLatency / time.Duration(len(samples))
	avgAdpLatency := sumAdpLatency / time.Duration(len(samples))
	avgLatRatio := ratio(avgAdpLatency, avgBaseLatency)

	fmt.Println("\n## Summary")
	fmt.Printf("- Average baseline saved: %.2f%%\n", avgBaseSaved)
	fmt.Printf("- Average adaptive saved: %.2f%%\n", avgAdpSaved)
	fmt.Printf("- Average saved delta: %.2f%%\n", avgDelta)
	fmt.Printf("- Average latency ratio (adaptive/baseline): %.2fx\n", avgLatRatio)

	fmt.Println("\n## Suggested Defaults")
	maxLines := 400
	headLines := 80
	tailLines := 60
	signalLines := 120

	if avgLatRatio > 1.8 && avgDelta < 4.0 {
		maxLines = 550
		headLines = 70
		tailLines = 50
		signalLines = 110
	}
	if avgDelta > 8.0 {
		signalLines = 140
	}

	fmt.Printf("- `extractive_max_lines=%d`\n", maxLines)
	fmt.Printf("- `extractive_head_lines=%d`\n", headLines)
	fmt.Printf("- `extractive_tail_lines=%d`\n", tailLines)
	fmt.Printf("- `extractive_signal_lines=%d`\n", signalLines)
	fmt.Println("- `enable_quality_guardrail=true`")
	fmt.Println("- Keep `profile=adaptive` for mixed logs/diff/test sessions")
}

func buildSamples() []sample {
	return []sample{
		{
			name: "build_failure",
			content: strings.Repeat("INFO compiling package...\n", 220) +
				"ERROR: linker failed at cmd/main.go:42\n" +
				strings.Repeat("INFO retry...\n", 100),
		},
		{
			name: "test_failure",
			content: strings.Repeat("=== RUN TestX\n", 90) +
				"--- FAIL: TestX (0.01s)\nexpected 1 got 2\n" +
				strings.Repeat("PASS helper case\n", 80),
		},
		{
			name: "diff_review",
			content: "diff --git a/internal/a.go b/internal/a.go\n@@ -10,7 +10,9 @@\n" +
				strings.Repeat("- old line\n+ new line\n", 160),
		},
		{
			name: "ops_rollout",
			content: strings.Repeat("kubectl get pods -A\n", 180) +
				"rollout status deployment/api\n" +
				strings.Repeat("Warning: backoff restarting failed container\n", 50),
		},
	}
}

func runSample(input string, cfg filter.PipelineConfig, iterations int) result {
	var finalTokens int
	var saved float64
	start := time.Now()
	for i := 0; i < iterations; i++ {
		p := filter.NewPipelineCoordinator(cfg)
		_, stats := p.Process(input)
		finalTokens = stats.FinalTokens
		saved = stats.ReductionPercent
	}
	total := time.Since(start)

	return result{
		origTokens:  filter.EstimateTokens(input),
		finalTokens: finalTokens,
		savedPct:    saved,
		latency:     total / time.Duration(iterations),
	}
}

func ratio(a, b time.Duration) float64 {
	if b <= 0 {
		return 0
	}
	return float64(a) / float64(b)
}
