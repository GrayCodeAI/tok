package core

import (
	"fmt"
	"testing"

	"github.com/GrayCodeAI/tokman/internal/tracking"
)

func BenchmarkFormatTokensInt(b *testing.B) {
	tests := []int{
		500,
		1000,
		15000,
		999999,
		1000000,
		15000000,
		100000000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, n := range tests {
			formatTokensInt(n)
		}
	}
}

func BenchmarkFormatDuration(b *testing.B) {
	tests := []int64{
		500,
		1000,
		15000,
		60000,
		90000,
		3600000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ms := range tests {
			formatDuration(ms)
		}
	}
}

func BenchmarkTruncate(b *testing.B) {
	tests := []struct {
		s      string
		maxLen int
	}{
		{"short", 100},
		{"this is a medium length string", 20},
		{"this is a very long string that needs to be truncated", 30},
		{string(make([]byte, 10000)), 50},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tt := range tests {
			truncate(tt.s, tt.maxLen)
		}
	}
}

func BenchmarkGetTierForTokens(b *testing.B) {
	tierLimits := map[string]int{
		"free":      1_000_000,
		"pro":       5_000_000,
		"5x":        25_000_000,
		"20x":       100_000_000,
		"unlimited": 999_999_999,
	}

	tests := []int{
		500_000,
		3_000_000,
		10_000_000,
		50_000_000,
		200_000_000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tokens := range tests {
			getTierForTokens(tokens, tierLimits)
		}
	}
}

func BenchmarkQuotaEstimationCalculations(b *testing.B) {
	summary := &tracking.GainSummary{
		TotalCommands: 1000,
		TotalInput:    10_000_000,
		TotalOutput:   2_000_000,
		TotalSaved:    8_000_000,
		AvgSavingsPct: 80.0,
		DailyStats:    make([]tracking.PeriodStats, 30),
	}

	// Populate daily stats
	for i := 0; i < 30; i++ {
		summary.DailyStats[i] = tracking.PeriodStats{
			Period:      fmt.Sprintf("2024-01-%02d", i+1),
			SavedTokens: 100000 + i*1000,
		}
	}

	tierLimits := map[string]int{
		"free":      1_000_000,
		"pro":       5_000_000,
		"5x":        25_000_000,
		"20x":       100_000_000,
		"unlimited": 999_999_999,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Calculate all the quota metrics
		inputTokens := summary.TotalInput
		outputTokens := summary.TotalOutput
		totalTokens := inputTokens + outputTokens
		days := len(summary.DailyStats)
		avgDaily := totalTokens / days
		monthlyProjection := avgDaily * 30

		for tier, limit := range tierLimits {
			usagePct := float64(monthlyProjection) / float64(limit) * 100
			_ = usagePct
			_ = tier
		}
	}
}

func BenchmarkShortenPath(b *testing.B) {
	tests := []string{
		"/home/user/Documents/project/file.txt",
		"/Users/username/projects/app/src/main.go",
		"/usr/local/bin/myapp",
		"relative/path/to/file",
		"file.txt",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range tests {
			shortenPath(path)
		}
	}
}

func BenchmarkProgressBarGeneration(b *testing.B) {
	barWidth := 40
	percentages := []float64{0, 25, 50, 75, 100, 150}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pct := range percentages {
			filled := int((pct / 100.0) * float64(barWidth))
			if filled > barWidth {
				filled = barWidth
			}
			_ = filled
		}
	}
}

func BenchmarkQuotaEstimationWithDifferentDataSizes(b *testing.B) {
	sizes := []int{7, 30, 90, 365}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("days-%d", size), func(b *testing.B) {
			summary := &tracking.GainSummary{
				TotalCommands: 1000,
				TotalInput:    10_000_000,
				TotalOutput:   2_000_000,
				TotalSaved:    8_000_000,
				AvgSavingsPct: 80.0,
				DailyStats:    make([]tracking.PeriodStats, size),
			}

			for i := 0; i < size; i++ {
				summary.DailyStats[i] = tracking.PeriodStats{
					Period:      fmt.Sprintf("day-%d", i),
					SavedTokens: 100000,
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				totalTokens := summary.TotalInput + summary.TotalOutput
				days := len(summary.DailyStats)
				avgDaily := totalTokens / days
				_ = avgDaily * 30
			}
		})
	}
}
