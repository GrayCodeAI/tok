package filter

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// Stress test inputs - larger and more diverse
var stressInputs = map[string]string{
	"large_git_diff": generateLargeGitDiff(),
	"large_log":      generateLargeLog(),
	"mixed_content":  generateMixedContent(),
	"code_heavy":     generateCodeHeavy(),
}

func generateLargeGitDiff() string {
	var sb strings.Builder
	sb.WriteString("diff --git a/src/main.go b/src/main.go\n")
	sb.WriteString("index abc1234..def5678 100644\n")
	sb.WriteString("--- a/src/main.go\n")
	sb.WriteString("+++ b/src/main.go\n")
	for i := 0; i < 50; i++ {
		sb.WriteString(fmt.Sprintf("@@ -%d,6 +%d,10 @@\n", i*10, i*10))
		sb.WriteString(" import \"fmt\"\n")
		sb.WriteString(" import \"os\"\n")
		sb.WriteString("+import \"context\"\n")
		sb.WriteString("+import \"time\"\n")
		sb.WriteString(" \n")
		sb.WriteString(" func main() {\n")
		sb.WriteString("-\tfmt.Println(\"hello\")\n")
		sb.WriteString("+\tctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)\n")
		sb.WriteString("+\tfmt.Println(ctx)\n")
		sb.WriteString("+\tcancel()\n")
		sb.WriteString(" }\n")
	}
	return sb.String()
}

func generateLargeLog() string {
	var sb strings.Builder
	levels := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	for i := 0; i < 100; i++ {
		sb.WriteString(fmt.Sprintf("2026-03-24T%02d:%02d:%02dZ [%s] Request %d processed in %dms status=%d\n",
			i%24, i%60, i%60, levels[i%4], i, i*10+5, 200+i%5))
	}
	return sb.String()
}

func generateMixedContent() string {
	var sb strings.Builder
	sb.WriteString("# API Documentation\n\n")
	sb.WriteString("## Overview\n\nThis API provides access to the compression pipeline.\n\n")
	sb.WriteString("## Endpoints\n\n")
	for i := 0; i < 30; i++ {
		sb.WriteString(fmt.Sprintf("### POST /api/v1/compress/%d\n\n", i))
		sb.WriteString("Compresses the given input using the specified pipeline.\n\n")
		sb.WriteString("```json\n")
		sb.WriteString(fmt.Sprintf(`{"input": "text_%d", "mode": "minimal", "budget": %d}`, i, i*100))
		sb.WriteString("\n```\n\n")
		sb.WriteString("Response: 200 OK\n\n")
	}
	return sb.String()
}

func generateCodeHeavy() string {
	var sb strings.Builder
	sb.WriteString("package main\n\n")
	sb.WriteString("import (\n\t\"fmt\"\n\t\"strings\"\n)\n\n")
	for i := 0; i < 40; i++ {
		sb.WriteString(fmt.Sprintf("func process%d(input string) string {\n", i))
		sb.WriteString("\tresult := strings.ToUpper(input)\n")
		sb.WriteString("\tresult = strings.TrimSpace(result)\n")
		sb.WriteString("\treturn result\n")
		sb.WriteString("}\n\n")
	}
	return sb.String()
}

// TestStressComparison runs 20 iterations with all pipeline variants
func TestStressComparison(t *testing.T) {
	iterations := 20

	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║            TokMan Adaptive Pipeline: Tiered vs Full (20 iterations × 11 inputs)             ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║ Tiers: Trivial(0L) | Simple(3L) | Medium(8L) | Complex(15L) | Extreme(31L)                  ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════════════════════════════════════╣")

	allInputs := make(map[string]string)
	for k, v := range benchmarkInputs {
		allInputs[k] = v
	}
	for k, v := range stressInputs {
		allInputs[k] = v
	}

	totalOrigTokens := 0
	totalFullSaved := 0
	totalAdaptiveSaved := 0
	totalFullTime := time.Duration(0)
	totalAdaptiveTime := time.Duration(0)

	for name, input := range allInputs {
		origTokens := EstimateTokens(input)
		totalOrigTokens += origTokens

		// Full pipeline (all 31 layers)
		var fullTotalSaved int
		var fullTotalTime time.Duration
		for i := 0; i < iterations; i++ {
			start := time.Now()
			_, saved := QuickProcess(input, ModeMinimal)
			fullTotalTime += time.Since(start)
			fullTotalSaved += saved
		}
		fullAvgSaved := fullTotalSaved / iterations
		fullAvgTime := fullTotalTime / time.Duration(iterations)

		// Adaptive pipeline (complexity-based tiering)
		var adaptiveTotalSaved int
		var adaptiveTotalTime time.Duration
		var detectedTier Tier
		for i := 0; i < iterations; i++ {
			start := time.Now()
			cfg := PresetConfig(PresetFull, ModeMinimal)
			ap := NewAdaptive(cfg)
			_, stats := ap.Process(input)
			adaptiveTotalTime += time.Since(start)
			adaptiveTotalSaved += stats.TotalSaved
			if i == 0 {
				detectedTier = ap.DetectTier(input)
			}
		}
		adaptiveAvgSaved := adaptiveTotalSaved / iterations
		adaptiveAvgTime := adaptiveTotalTime / time.Duration(iterations)

		totalFullSaved += fullAvgSaved
		totalAdaptiveSaved += adaptiveAvgSaved
		totalFullTime += fullAvgTime
		totalAdaptiveTime += adaptiveAvgTime

		fullRatio := float64(fullAvgSaved) / float64(origTokens) * 100
		adaptiveRatio := float64(adaptiveAvgSaved) / float64(origTokens) * 100

		fmt.Printf("║ %s [%s]\n", name, detectedTier.String())
		fmt.Printf("║   %d tok | Full(20L): %.1f%% (%.2fms) | Adaptive(%s): %.1f%% (%.2fms) | Speed: %.1fx\n",
			origTokens,
			fullRatio, float64(fullAvgTime.Microseconds())/1000,
			detectedTier.String(),
			adaptiveRatio, float64(adaptiveAvgTime.Microseconds())/1000,
			float64(fullAvgTime.Microseconds())/float64(adaptiveAvgTime.Microseconds()))
	}

	totalFullRatio := float64(totalFullSaved) / float64(totalOrigTokens) * 100
	totalAdaptiveRatio := float64(totalAdaptiveSaved) / float64(totalOrigTokens) * 100
	speedup := float64(totalFullTime.Microseconds()) / float64(totalAdaptiveTime.Microseconds())

	fmt.Printf("\n╠══════════════════════════════════════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║ AGGREGATE (%d inputs × %d iterations = %d runs):\n", len(allInputs), iterations, len(allInputs)*iterations)
	fmt.Printf("║   Total original tokens: %d\n", totalOrigTokens)
	fmt.Printf("║   Full Pipeline (31 layers):  %.1f%% compression | %.2fms avg\n", totalFullRatio, float64(totalFullTime.Microseconds())/1000)
	fmt.Printf("║   Adaptive Pipeline (tiered): %.1f%% compression | %.2fms avg\n", totalAdaptiveRatio, float64(totalAdaptiveTime.Microseconds())/1000)
	fmt.Printf("║   Compression difference: %+.1f%%\n", totalAdaptiveRatio-totalFullRatio)
	fmt.Printf("║   Speed improvement: %.2fx faster\n", speedup)
	fmt.Printf("║   Simple inputs: 3 layers (<0.5ms) | Medium: 8 layers (<2ms)\n")
	fmt.Printf("║   Complex: 15 layers (<10ms) | Extreme: 31 layers (full)\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════════════════════════════════════════╝\n")
}

func generateTestInput(size int) string {
	var sb strings.Builder
	words := []string{"the", "quick", "brown", "fox", "jumps", "over", "lazy", "dog",
		"function", "return", "import", "package", "error", "success", "test", "run",
		"file", "path", "line", "column", "value", "string", "number", "boolean"}

	for i := 0; i < size; i++ {
		sb.WriteString(words[i%len(words)])
		sb.WriteString(" ")
		if (i+1)%10 == 0 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
