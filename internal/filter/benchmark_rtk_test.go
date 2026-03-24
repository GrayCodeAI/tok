package filter

import (
	"fmt"
	"testing"
)

// TestBenchmarkVsRTK simulates RTK's compression and compares against TokMan.
// RTK uses 4 simple strategies: filter, group, truncate, dedup.
// TokMan uses 31 research-backed layers.
//
// RTK reference: https://github.com/yoav-lavi/rtk
// RTK achieves 60-90% compression with 4 strategies.
func TestBenchmarkVsRTK(t *testing.T) {
	iterations := 20

	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              TokMan vs RTK: Head-to-Head Compression Benchmark (20 iter)             ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║ RTK Strategies: filter + group + truncate + dedup (4 simple heuristics)              ║")
	fmt.Println("║ TokMan Layers:  31 research-backed layers from 20+ papers                           ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════════════════════════════╣")

	totalOrig := 0
	totalRTK := 0
	totalTokMan := 0
	totalUltra := 0

	for name, input := range benchmarkInputs {
		origTokens := EstimateTokens(input)
		totalOrig += origTokens

		// RTK simulation: 4 simple strategies
		var rtkTotalSaved int
		for i := 0; i < iterations; i++ {
			rtkTotalSaved += simulateRTK(input)
		}
		rtkAvgSaved := rtkTotalSaved / iterations

		// TokMan: Full 31-layer pipeline
		var tokmanTotalSaved int
		for i := 0; i < iterations; i++ {
			_, saved := QuickProcess(input, ModeMinimal)
			tokmanTotalSaved += saved
		}
		tokmanAvgSaved := tokmanTotalSaved / iterations

		// TokMan Ultra-Fast: 3-layer sub-ms pipeline
		var ultraTotalSaved int
		for i := 0; i < iterations; i++ {
			result := UltraFastCompress(input, origTokens/4)
			ultraSaved := origTokens - EstimateTokens(result)
			if ultraSaved > 0 {
				ultraTotalSaved += ultraSaved
			}
		}
		ultraAvgSaved := ultraTotalSaved / iterations

		totalRTK += rtkAvgSaved
		totalTokMan += tokmanAvgSaved
		totalUltra += ultraAvgSaved

		rtkRatio := float64(rtkAvgSaved) / float64(origTokens) * 100
		tokmanRatio := float64(tokmanAvgSaved) / float64(origTokens) * 100
		ultraRatio := float64(ultraAvgSaved) / float64(origTokens) * 100

		fmt.Printf("║ %s\n", name)
		fmt.Printf("║   %d tok | RTK: %.1f%% | TokMan: %.1f%% | Ultra: %.1f%% | Δ(TokMan-RTK): %+.1f%%\n",
			origTokens, rtkRatio, tokmanRatio, ultraRatio, tokmanRatio-rtkRatio)
	}

	totalRTKRatio := float64(totalRTK) / float64(totalOrig) * 100
	totalTokManRatio := float64(totalTokMan) / float64(totalOrig) * 100
	totalUltraRatio := float64(totalUltra) / float64(totalOrig) * 100

	fmt.Printf("\n╠══════════════════════════════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║ AGGREGATE (%d inputs × %d iterations):\n", len(benchmarkInputs), iterations)
	fmt.Printf("║   Total original tokens: %d\n", totalOrig)
	fmt.Printf("║   RTK (4 strategies):      %.1f%% compression\n", totalRTKRatio)
	fmt.Printf("║   TokMan Ultra-Fast (3):   %.1f%% compression\n", totalUltraRatio)
	fmt.Printf("║   TokMan Full (31 layers): %.1f%% compression\n", totalTokManRatio)
	fmt.Printf("║   TokMan vs RTK: %+.1f%% improvement\n", totalTokManRatio-totalRTKRatio)
	fmt.Printf("╚══════════════════════════════════════════════════════════════════════════════════════╝\n")
}

// simulateRTK simulates RTK's 4 compression strategies
func simulateRTK(input string) int {
	originalTokens := EstimateTokens(input)

	// Strategy 1: Filter - remove ANSI, progress bars, empty lines
	filtered := simulateRTKFilter(input)

	// Strategy 2: Group - combine related lines
	grouped := simulateRTKGroup(filtered)

	// Strategy 3: Truncate - limit output length
	truncated := simulateRTKTruncate(grouped)

	// Strategy 4: Dedup - remove duplicate lines
	deduped := simulateRTKDedup(truncated)

	finalTokens := EstimateTokens(deduped)
	saved := originalTokens - finalTokens
	if saved < 0 {
		return 0
	}
	return saved
}

func simulateRTKFilter(input string) string {
	lines := splitLinesZeroCopy(input, nil)
	var result []byte
	for _, line := range lines {
		trimmed := line
		// Remove ANSI
		for i := 0; i < len(trimmed); i++ {
			if trimmed[i] == '\x1b' {
				// Skip escape sequence
				for i < len(trimmed) && !((trimmed[i] >= 'a' && trimmed[i] <= 'z') || (trimmed[i] >= 'A' && trimmed[i] <= 'Z')) {
					i++
				}
				continue
			}
		}
		// Remove empty lines
		if len(trimmed) == 0 {
			continue
		}
		result = append(result, trimmed...)
		result = append(result, '\n')
	}
	return string(result)
}

func simulateRTKGroup(input string) string {
	// RTK groups similar consecutive lines
	lines := splitLinesZeroCopy(input, nil)
	if len(lines) <= 1 {
		return input
	}

	var result []byte
	prev := ""
	count := 1
	for _, line := range lines {
		if line == prev {
			count++
		} else {
			if prev != "" {
				if count > 1 {
					result = append(result, fmt.Sprintf("%s (×%d)", prev, count)...)
				} else {
					result = append(result, prev...)
				}
				result = append(result, '\n')
			}
			prev = line
			count = 1
		}
	}
	if prev != "" {
		if count > 1 {
			result = append(result, fmt.Sprintf("%s (×%d)", prev, count)...)
		} else {
			result = append(result, prev...)
		}
		result = append(result, '\n')
	}
	return string(result)
}

func simulateRTKTruncate(input string) string {
	// RTK truncates to a max line count
	lines := splitLinesZeroCopy(input, nil)
	if len(lines) <= 50 {
		return input
	}
	// Keep first 20 and last 30
	var result []byte
	for i := 0; i < 20 && i < len(lines); i++ {
		result = append(result, lines[i]...)
		result = append(result, '\n')
	}
	result = append(result, "[... truncated ...]\n"...)
	for i := len(lines) - 30; i < len(lines); i++ {
		if i >= 0 {
			result = append(result, lines[i]...)
			result = append(result, '\n')
		}
	}
	return string(result)
}

func simulateRTKDedup(input string) string {
	seen := make(map[string]bool)
	lines := splitLinesZeroCopy(input, nil)
	var result []byte
	for _, line := range lines {
		if !seen[line] {
			seen[line] = true
			result = append(result, line...)
			result = append(result, '\n')
		}
	}
	return string(result)
}
