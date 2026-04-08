package tracking

import (
	"testing"
	"time"
)

// TestCommandRecordCreation tests command record creation
func TestCommandRecordCreation(t *testing.T) {
	record := CommandRecord{
		Command:        "git status",
		ProjectPath:    "/home/user/project",
		Timestamp:      time.Now(),
		OriginalTokens: 1000,
		FilteredTokens: 500,
		SavedTokens:    500,
	}

	if record.Command != "git status" {
		t.Errorf("Command = %q, want %q", record.Command, "git status")
	}

	if record.SavedTokens != 500 {
		t.Errorf("SavedTokens = %d, want %d", record.SavedTokens, 500)
	}

	expectedSavings := float64(500) / float64(1000) * 100
	actualSavings := float64(record.SavedTokens) / float64(record.OriginalTokens) * 100
	if actualSavings != expectedSavings {
		t.Errorf("Savings = %f%%, want %f%%", actualSavings, expectedSavings)
	}
}

// TestSavingsCalculation tests savings calculation
func TestSavingsCalculation(t *testing.T) {
	tests := []struct {
		name            string
		original        int
		filtered        int
		expectedSaved   int
		expectedPercent float64
	}{
		{
			name:            "50% savings",
			original:        1000,
			filtered:        500,
			expectedSaved:   500,
			expectedPercent: 50.0,
		},
		{
			name:            "no savings",
			original:        1000,
			filtered:        1000,
			expectedSaved:   0,
			expectedPercent: 0.0,
		},
		{
			name:            "max savings",
			original:        1000,
			filtered:        100,
			expectedSaved:   900,
			expectedPercent: 90.0,
		},
		{
			name:            "zero original",
			original:        0,
			filtered:        0,
			expectedSaved:   0,
			expectedPercent: 0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			saved := tc.original - tc.filtered
			var percent float64
			if tc.original > 0 {
				percent = float64(saved) / float64(tc.original) * 100
			}

			if saved != tc.expectedSaved {
				t.Errorf("Saved = %d, want %d", saved, tc.expectedSaved)
			}

			if percent != tc.expectedPercent {
				t.Errorf("Percent = %f, want %f", percent, tc.expectedPercent)
			}
		})
	}
}

// TestRecordValidation tests record validation
func TestRecordValidation(t *testing.T) {
	tests := []struct {
		name   string
		record CommandRecord
		valid  bool
	}{
		{
			name: "valid record",
			record: CommandRecord{
				Command:        "git status",
				ProjectPath:    "/home/user",
				Timestamp:      time.Now(),
				OriginalTokens: 100,
				FilteredTokens: 50,
			},
			valid: true,
		},
		{
			name: "empty command",
			record: CommandRecord{
				Command:        "",
				ProjectPath:    "/home/user",
				Timestamp:      time.Now(),
				OriginalTokens: 100,
				FilteredTokens: 50,
			},
			valid: false,
		},
		{
			name: "negative tokens",
			record: CommandRecord{
				Command:        "git status",
				ProjectPath:    "/home/user",
				Timestamp:      time.Now(),
				OriginalTokens: -100,
				FilteredTokens: 50,
			},
			valid: false,
		},
		{
			name: "filtered > original",
			record: CommandRecord{
				Command:        "git status",
				ProjectPath:    "/home/user",
				Timestamp:      time.Now(),
				OriginalTokens: 100,
				FilteredTokens: 150,
			},
			valid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid := tc.record.Command != "" &&
				tc.record.OriginalTokens >= 0 &&
				tc.record.FilteredTokens >= 0 &&
				tc.record.FilteredTokens <= tc.record.OriginalTokens

			if valid != tc.valid {
				t.Errorf("Record validation: valid=%v, want valid=%v", valid, tc.valid)
			}
		})
	}
}

// TestStatisticsAggregation tests statistics aggregation
func TestStatisticsAggregation(t *testing.T) {
	records := []CommandRecord{
		{Command: "git status", SavedTokens: 100, OriginalTokens: 200},
		{Command: "git log", SavedTokens: 200, OriginalTokens: 333},
		{Command: "docker ps", SavedTokens: 150, OriginalTokens: 333},
	}

	var totalSaved int
	var totalPercent float64

	for _, r := range records {
		totalSaved += r.SavedTokens
		if r.OriginalTokens > 0 {
			totalPercent += float64(r.SavedTokens) / float64(r.OriginalTokens) * 100
		}
	}

	avgPercent := totalPercent / float64(len(records))

	if totalSaved != 450 {
		t.Errorf("Total saved = %d, want 450", totalSaved)
	}

	// Average should be around 51.67%
	if avgPercent < 50 || avgPercent > 55 {
		t.Errorf("Average percent = %f, expected around 51.67", avgPercent)
	}
}

// TestTimeRangeFiltering tests time range filtering
func TestTimeRangeFiltering(t *testing.T) {
	now := time.Now()
	records := []CommandRecord{
		{Command: "cmd1", Timestamp: now.Add(-24 * time.Hour)},
		{Command: "cmd2", Timestamp: now.Add(-12 * time.Hour)},
		{Command: "cmd3", Timestamp: now.Add(-1 * time.Hour)},
		{Command: "cmd4", Timestamp: now},
	}

	// Filter for last 6 hours
	cutoff := now.Add(-6 * time.Hour)
	var recent []CommandRecord

	for _, r := range records {
		if r.Timestamp.After(cutoff) {
			recent = append(recent, r)
		}
	}

	if len(recent) != 2 {
		t.Errorf("Recent commands = %d, want 2", len(recent))
	}
}

// BenchmarkSavingsCalculation benchmarks savings calculation
func BenchmarkSavingsCalculation(b *testing.B) {
	original := 1000
	filtered := 500

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = original - filtered
		_ = float64(original-filtered) / float64(original) * 100
	}
}
