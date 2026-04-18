package core

import "testing"

func TestBuildCcusageArgs(t *testing.T) {
	tests := []struct {
		name       string
		invocation ccusageInvocation
		want       []string
	}{
		{
			name:       "direct binary",
			invocation: ccusageInvocation{Path: "ccusage"},
			want:       []string{"daily", "--json", "--since", "20250101"},
		},
		{
			name:       "npx wrapper",
			invocation: ccusageInvocation{Path: "npx", BaseArgs: []string{"ccusage"}},
			want:       []string{"ccusage", "daily", "--json", "--since", "20250101"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildCcusageArgs(tc.invocation, "daily")
			if len(got) != len(tc.want) {
				t.Fatalf("len(args) = %d, want %d (%v)", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("args[%d] = %q, want %q (full: %v)", i, got[i], tc.want[i], got)
				}
			}
		})
	}
}

func TestGenerateEconomicsReportsWithoutCcusageFallsBackToTokmanData(t *testing.T) {
	reports := generateEconomicsReports(nil, map[string][]TokManSavings{
		"daily": {
			{Date: "2026-04-18", Commands: 2, SavedTokens: 500, OriginalSize: 1000, FilteredSize: 500},
		},
	})

	if len(reports) != 1 {
		t.Fatalf("len(reports) = %d, want 1", len(reports))
	}
	if reports[0].Period != "2026-04-18" {
		t.Fatalf("Period = %q, want 2026-04-18", reports[0].Period)
	}
	if reports[0].SavingsPercent != 50 {
		t.Fatalf("SavingsPercent = %v, want 50", reports[0].SavingsPercent)
	}
}
