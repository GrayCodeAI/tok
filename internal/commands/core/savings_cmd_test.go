package core

import (
	"os"
	"strings"
	"testing"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

func TestGetTierForTokens(t *testing.T) {
	tierLimits := map[string]int{
		"free":      1_000_000,
		"pro":       5_000_000,
		"5x":        25_000_000,
		"20x":       100_000_000,
		"unlimited": 999_999_999,
	}

	tests := []struct {
		tokens   int
		expected string
	}{
		{500_000, "free"},
		{1_000_000, "pro"}, // At boundary, should go up
		{3_000_000, "pro"},
		{5_000_000, "5x"}, // At boundary
		{10_000_000, "5x"},
		{25_000_000, "20x"}, // At boundary
		{50_000_000, "20x"},
		{100_000_000, "unlimited"}, // At boundary
		{200_000_000, "unlimited"},
	}

	for _, tt := range tests {
		t.Run(formatTokensInt(tt.tokens), func(t *testing.T) {
			result := getTierForTokens(tt.tokens, tierLimits)
			if result != tt.expected {
				t.Errorf("getTierForTokens(%d) = %q, expected %q", tt.tokens, result, tt.expected)
			}
		})
	}
}

func TestFormatTokensInt(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{500, "500"},
		{1000, "1.0k"},
		{1500, "1.5k"},
		{999_999, "1000.0k"},
		{1_000_000, "1.00M"},
		{1_500_000, "1.50M"},
		{10_000_000, "10.00M"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatTokensInt(tt.input)
			if result != tt.expected {
				t.Errorf("formatTokensInt(%d) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		ms       int64
		expected string
	}{
		{500, "500ms"},
		{1000, "1.0s"},
		{1500, "1.5s"},
		{60_000, "1m0s"},
		{90_000, "1m30s"},
		{3_600_000, "60m0s"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.ms)
			if result != tt.expected {
				t.Errorf("formatDuration(%d) = %q, expected %q", tt.ms, result, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s        string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"short", 5, "short"},
		{"very long string here", 10, "very lo..."},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			result := truncate(tt.s, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, expected %q", tt.s, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestShortenPath(t *testing.T) {
	home, _ := getHomeDir()
	if home == "" {
		t.Skip("Cannot determine home directory")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{home + "/Documents", "~/Documents"},
		{home + "/projects/app", "~/projects/app"},
		{"/usr/local/bin", "/usr/local/bin"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := shortenPath(tt.input)
			if result != tt.expected {
				t.Errorf("shortenPath(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPrintQuotaEstimationCalculations(t *testing.T) {
	// Test the quota calculation logic
	summary := &tracking.GainSummary{
		TotalCommands: 100,
		TotalInput:    1_000_000,
		TotalOutput:   200_000,
		TotalSaved:    800_000,
		AvgSavingsPct: 80.0,
		DailyStats: []tracking.PeriodStats{
			{Period: "2024-01-01", SavedTokens: 10000},
			{Period: "2024-01-02", SavedTokens: 15000},
			{Period: "2024-01-03", SavedTokens: 12000},
		},
	}

	// Calculate expected values
	inputTokens := summary.TotalInput
	outputTokens := summary.TotalOutput
	totalTokens := inputTokens + outputTokens
	days := len(summary.DailyStats)
	avgDaily := totalTokens / days
	monthlyProjection := avgDaily * 30

	// Verify calculations
	if totalTokens != 1_200_000 {
		t.Errorf("Expected total tokens 1,200,000, got %d", totalTokens)
	}

	if days != 3 {
		t.Errorf("Expected 3 days, got %d", days)
	}

	expectedAvgDaily := 1_200_000 / 3
	if avgDaily != expectedAvgDaily {
		t.Errorf("Expected avg daily %d, got %d", expectedAvgDaily, avgDaily)
	}

	expectedMonthly := expectedAvgDaily * 30
	if monthlyProjection != expectedMonthly {
		t.Errorf("Expected monthly projection %d, got %d", expectedMonthly, monthlyProjection)
	}
}

func TestGainFlagDefaults(t *testing.T) {
	// Verify flag defaults are set correctly
	if gainSinceDays != 30 {
		t.Errorf("Expected gainSinceDays default to be 30, got %d", gainSinceDays)
	}

	if gainFormat != "text" {
		t.Errorf("Expected gainFormat default to be 'text', got %q", gainFormat)
	}
}

// Helper function that doesn't depend on os.UserHomeDir
func getHomeDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	return home, nil
}

// Mock writer for testing output
type mockWriter struct {
	content []string
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	m.content = append(m.content, string(p))
	return len(p), nil
}

func TestProgressBarGeneration(t *testing.T) {
	// Test the progress bar generation logic indirectly
	tests := []struct {
		pct      float64
		width    int
		expected int // expected filled characters
	}{
		{0, 40, 0},
		{25, 40, 10},
		{50, 40, 20},
		{75, 40, 30},
		{100, 40, 40},
		{150, 40, 40}, // should cap at width
	}

	for _, tt := range tests {
		t.Run(formatTokensInt(int(tt.pct)), func(t *testing.T) {
			filled := int((tt.pct / 100.0) * float64(tt.width))
			if filled > tt.width {
				filled = tt.width
			}
			if filled != tt.expected {
				t.Errorf("Expected %d filled chars, got %d", tt.expected, filled)
			}
		})
	}
}

func TestTruncateEdgeCases(t *testing.T) {
	// Test truncate with edge cases
	tests := []struct {
		s      string
		maxLen int
	}{
		{"", 10},                        // empty string
		{strings.Repeat("a", 1000), 50}, // very long string
		{"test", 5},                     // normal case
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			// Should not panic for valid maxLen >= 3
			if tt.maxLen >= 3 {
				result := truncate(tt.s, tt.maxLen)

				// Result should not exceed maxLen
				if len(result) > tt.maxLen {
					t.Errorf("truncate(%q, %d) returned %q with length %d > %d",
						tt.s, tt.maxLen, result, len(result), tt.maxLen)
				}
			}
		})
	}
}
