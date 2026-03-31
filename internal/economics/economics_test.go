package economics

import (
	"testing"
	"time"

	"github.com/GrayCodeAI/tokman/internal/ccusage"
)

// ── Weight Constants ───────────────────────────────────────────

func TestWeightConstants(t *testing.T) {
	if WeightOutput != 5.0 {
		t.Errorf("WeightOutput = %f, want 5.0", WeightOutput)
	}
	if WeightCacheCreate != 1.25 {
		t.Errorf("WeightCacheCreate = %f, want 1.25", WeightCacheCreate)
	}
	if WeightCacheRead != 0.1 {
		t.Errorf("WeightCacheRead = %f, want 0.1", WeightCacheRead)
	}
}

// ── PeriodEconomics.computeWeightedMetrics ─────────────────────

func TestPeriodEconomics_ComputeWeightedMetrics_NilValues(t *testing.T) {
	p := &PeriodEconomics{}
	p.computeWeightedMetrics()
	if p.WeightedInputCPT != nil {
		t.Error("WeightedInputCPT should be nil with nil inputs")
	}
	if p.SavingsWeighted != nil {
		t.Error("SavingsWeighted should be nil with nil inputs")
	}
}

func TestPeriodEconomics_ComputeWeightedMetrics_PartialNil(t *testing.T) {
	cost := 10.0
	saved := 1000
	p := &PeriodEconomics{
		CCCost:        &cost,
		TMSavedTokens: &saved,
	}
	p.computeWeightedMetrics()
	if p.WeightedInputCPT != nil {
		t.Error("WeightedInputCPT should be nil when token breakdown is missing")
	}
}

func TestPeriodEconomics_ComputeWeightedMetrics_Valid(t *testing.T) {
	cost := 10.0
	saved := 1000
	input := uint64(1000)
	output := uint64(500)
	cacheCreate := uint64(100)
	cacheRead := uint64(200)

	p := &PeriodEconomics{
		CCCost:              &cost,
		TMSavedTokens:       &saved,
		CCInputTokens:       &input,
		CCOutputTokens:      &output,
		CCCacheCreateTokens: &cacheCreate,
		CCCacheReadTokens:   &cacheRead,
	}
	p.computeWeightedMetrics()

	if p.WeightedInputCPT == nil {
		t.Fatal("WeightedInputCPT should not be nil")
	}
	if p.SavingsWeighted == nil {
		t.Fatal("SavingsWeighted should not be nil")
	}

	// weightedUnits = 1000 + 5*500 + 1.25*100 + 0.1*200 = 1000 + 2500 + 125 + 20 = 3645
	// inputCPT = 10.0 / 3645
	expectedCPT := 10.0 / 3645.0
	if *p.WeightedInputCPT != expectedCPT {
		t.Errorf("WeightedInputCPT = %f, want %f", *p.WeightedInputCPT, expectedCPT)
	}

	// savings = 1000 * inputCPT
	expectedSavings := 1000.0 * expectedCPT
	if *p.SavingsWeighted != expectedSavings {
		t.Errorf("SavingsWeighted = %f, want %f", *p.SavingsWeighted, expectedSavings)
	}
}

func TestPeriodEconomics_ComputeWeightedMetrics_ZeroUnits(t *testing.T) {
	cost := 10.0
	saved := 1000
	zero := uint64(0)

	p := &PeriodEconomics{
		CCCost:              &cost,
		TMSavedTokens:       &saved,
		CCInputTokens:       &zero,
		CCOutputTokens:      &zero,
		CCCacheCreateTokens: &zero,
		CCCacheReadTokens:   &zero,
	}
	p.computeWeightedMetrics()

	if p.WeightedInputCPT != nil {
		t.Error("WeightedInputCPT should be nil when weightedUnits is 0")
	}
}

// ── PeriodEconomics.computeDualMetrics ─────────────────────────

func TestPeriodEconomics_ComputeDualMetrics_NilValues(t *testing.T) {
	p := &PeriodEconomics{}
	p.computeDualMetrics()
	if p.BlendedCPT != nil {
		t.Error("BlendedCPT should be nil with nil inputs")
	}
	if p.ActiveCPT != nil {
		t.Error("ActiveCPT should be nil with nil inputs")
	}
}

func TestPeriodEconomics_ComputeDualMetrics_PartialNil(t *testing.T) {
	cost := 10.0
	saved := 1000
	p := &PeriodEconomics{
		CCCost:        &cost,
		TMSavedTokens: &saved,
	}
	p.computeDualMetrics()
	if p.BlendedCPT != nil {
		t.Error("BlendedCPT should be nil when total tokens missing")
	}
	if p.ActiveCPT != nil {
		t.Error("ActiveCPT should be nil when active tokens missing")
	}
}

func TestPeriodEconomics_ComputeDualMetrics_Valid(t *testing.T) {
	cost := 10.0
	saved := 1000
	total := uint64(5000)
	active := uint64(3000)

	p := &PeriodEconomics{
		CCCost:         &cost,
		TMSavedTokens:  &saved,
		CCTotalTokens:  &total,
		CCActiveTokens: &active,
	}
	p.computeDualMetrics()

	if p.BlendedCPT == nil {
		t.Fatal("BlendedCPT should not be nil")
	}
	if p.ActiveCPT == nil {
		t.Fatal("ActiveCPT should not be nil")
	}
	if p.SavingsBlended == nil {
		t.Fatal("SavingsBlended should not be nil")
	}
	if p.SavingsActive == nil {
		t.Fatal("SavingsActive should not be nil")
	}

	expectedBlended := 10.0 / 5000.0
	if *p.BlendedCPT != expectedBlended {
		t.Errorf("BlendedCPT = %f, want %f", *p.BlendedCPT, expectedBlended)
	}

	expectedActive := 10.0 / 3000.0
	if *p.ActiveCPT != expectedActive {
		t.Errorf("ActiveCPT = %f, want %f", *p.ActiveCPT, expectedActive)
	}
}

func TestPeriodEconomics_ComputeDualMetrics_ZeroTokens(t *testing.T) {
	cost := 10.0
	saved := 1000
	zero := uint64(0)

	p := &PeriodEconomics{
		CCCost:         &cost,
		TMSavedTokens:  &saved,
		CCTotalTokens:  &zero,
		CCActiveTokens: &zero,
	}
	p.computeDualMetrics()

	if p.BlendedCPT != nil {
		t.Error("BlendedCPT should be nil when total tokens is 0")
	}
	if p.ActiveCPT != nil {
		t.Error("ActiveCPT should be nil when active tokens is 0")
	}
}

// ── computeTotals ──────────────────────────────────────────────

func TestComputeTotals_Empty(t *testing.T) {
	totals := computeTotals([]PeriodEconomics{})
	if totals.CCCost != 0 {
		t.Errorf("CCCost = %f, want 0", totals.CCCost)
	}
	if totals.TMSavedTokens != 0 {
		t.Errorf("TMSavedTokens = %d, want 0", totals.TMSavedTokens)
	}
	if totals.TMCommands != 0 {
		t.Errorf("TMCommands = %d, want 0", totals.TMCommands)
	}
	if totals.TMAvgSavingsPct != 0 {
		t.Errorf("TMAvgSavingsPct = %f, want 0", totals.TMAvgSavingsPct)
	}
}

func TestComputeTotals_SinglePeriod(t *testing.T) {
	cost := 10.0
	saved := 1000
	cmds := 50
	tokens := uint64(5000)
	input := uint64(2000)
	output := uint64(1000)
	cacheCreate := uint64(500)
	cacheRead := uint64(1500)
	pct := 50.0

	periods := []PeriodEconomics{
		{
			CCCost:              &cost,
			CCTotalTokens:       &tokens,
			CCActiveTokens:      &tokens,
			CCInputTokens:       &input,
			CCOutputTokens:      &output,
			CCCacheCreateTokens: &cacheCreate,
			CCCacheReadTokens:   &cacheRead,
			TMCommands:          &cmds,
			TMSavedTokens:       &saved,
			TMSavingsPct:        &pct,
		},
	}

	totals := computeTotals(periods)

	if totals.CCCost != cost {
		t.Errorf("CCCost = %f, want %f", totals.CCCost, cost)
	}
	if totals.TMSavedTokens != saved {
		t.Errorf("TMSavedTokens = %d, want %d", totals.TMSavedTokens, saved)
	}
	if totals.TMCommands != cmds {
		t.Errorf("TMCommands = %d, want %d", totals.TMCommands, cmds)
	}
	if totals.TMAvgSavingsPct != pct {
		t.Errorf("TMAvgSavingsPct = %f, want %f", totals.TMAvgSavingsPct, pct)
	}
}

func TestComputeTotals_MultiplePeriods(t *testing.T) {
	cost1 := 10.0
	cost2 := 20.0
	saved1 := 1000
	saved2 := 2000
	cmds1 := 50
	cmds2 := 100
	tokens := uint64(5000)
	input := uint64(2000)
	output := uint64(1000)
	cacheCreate := uint64(500)
	cacheRead := uint64(1500)
	pct1 := 50.0
	pct2 := 60.0

	periods := []PeriodEconomics{
		{
			CCCost:              &cost1,
			CCTotalTokens:       &tokens,
			CCActiveTokens:      &tokens,
			CCInputTokens:       &input,
			CCOutputTokens:      &output,
			CCCacheCreateTokens: &cacheCreate,
			CCCacheReadTokens:   &cacheRead,
			TMCommands:          &cmds1,
			TMSavedTokens:       &saved1,
			TMSavingsPct:        &pct1,
		},
		{
			CCCost:              &cost2,
			CCTotalTokens:       &tokens,
			CCActiveTokens:      &tokens,
			CCInputTokens:       &input,
			CCOutputTokens:      &output,
			CCCacheCreateTokens: &cacheCreate,
			CCCacheReadTokens:   &cacheRead,
			TMCommands:          &cmds2,
			TMSavedTokens:       &saved2,
			TMSavingsPct:        &pct2,
		},
	}

	totals := computeTotals(periods)

	if totals.CCCost != 30.0 {
		t.Errorf("CCCost = %f, want 30.0", totals.CCCost)
	}
	if totals.TMSavedTokens != 3000 {
		t.Errorf("TMSavedTokens = %d, want 3000", totals.TMSavedTokens)
	}
	if totals.TMCommands != 150 {
		t.Errorf("TMCommands = %d, want 150", totals.TMCommands)
	}
	expectedAvgPct := (50.0 + 60.0) / 2.0
	if totals.TMAvgSavingsPct != expectedAvgPct {
		t.Errorf("TMAvgSavingsPct = %f, want %f", totals.TMAvgSavingsPct, expectedAvgPct)
	}
}

func TestComputeTotals_NilFields(t *testing.T) {
	periods := []PeriodEconomics{
		{Label: "empty"},
	}
	totals := computeTotals(periods)
	if totals.CCCost != 0 {
		t.Errorf("CCCost should be 0 for nil fields, got %f", totals.CCCost)
	}
}

func TestComputeTotals_WeightedMetrics(t *testing.T) {
	cost := 10.0
	saved := 1000
	input := uint64(2000)
	output := uint64(1000)
	cacheCreate := uint64(500)
	cacheRead := uint64(1500)
	tokens := uint64(5000)
	pct := 50.0

	periods := []PeriodEconomics{
		{
			CCCost:              &cost,
			CCTotalTokens:       &tokens,
			CCActiveTokens:      &tokens,
			CCInputTokens:       &input,
			CCOutputTokens:      &output,
			CCCacheCreateTokens: &cacheCreate,
			CCCacheReadTokens:   &cacheRead,
			TMSavedTokens:       &saved,
			TMSavingsPct:        &pct,
		},
	}

	totals := computeTotals(periods)

	if totals.WeightedInputCPT == nil {
		t.Error("WeightedInputCPT should not be nil")
	}
	if totals.SavingsWeighted == nil {
		t.Error("SavingsWeighted should not be nil")
	}
	if totals.BlendedCPT == nil {
		t.Error("BlendedCPT should not be nil")
	}
	if totals.ActiveCPT == nil {
		t.Error("ActiveCPT should not be nil")
	}
}

// ── mergeMonthly ───────────────────────────────────────────────

func TestMergeMonthly_Empty(t *testing.T) {
	result := mergeMonthly(nil, nil)
	if len(result) != 0 {
		t.Errorf("mergeMonthly(nil, nil) returned %d periods, want 0", len(result))
	}
}

func TestMergeMonthly_CCOnly(t *testing.T) {
	cc := []ccusage.Period{
		{
			Key: "2026-01",
			Metrics: ccusage.Metrics{
				TotalCost:           10.0,
				TotalTokens:         5000,
				InputTokens:         2000,
				OutputTokens:        1000,
				CacheCreationTokens: 500,
				CacheReadTokens:     1500,
			},
		},
	}
	result := mergeMonthly(cc, nil)
	if len(result) != 1 {
		t.Fatalf("mergeMonthly returned %d periods, want 1", len(result))
	}
	if result[0].Label != "2026-01" {
		t.Errorf("Label = %q, want '2026-01'", result[0].Label)
	}
	if result[0].CCCost == nil || *result[0].CCCost != 10.0 {
		t.Errorf("CCCost = %v, want 10.0", result[0].CCCost)
	}
}

func TestMergeMonthly_TMOnly(t *testing.T) {
	tm := []MonthStats{
		{Month: "2026-01", Commands: 50, SavedTokens: 1000, SavingsPct: 50.0},
	}
	result := mergeMonthly(nil, tm)
	if len(result) != 1 {
		t.Fatalf("mergeMonthly returned %d periods, want 1", len(result))
	}
	if result[0].Label != "2026-01" {
		t.Errorf("Label = %q, want '2026-01'", result[0].Label)
	}
	if result[0].TMCommands == nil || *result[0].TMCommands != 50 {
		t.Errorf("TMCommands = %v, want 50", result[0].TMCommands)
	}
}

func TestMergeMonthly_Merged(t *testing.T) {
	cc := []ccusage.Period{
		{
			Key: "2026-01",
			Metrics: ccusage.Metrics{
				TotalCost:           10.0,
				TotalTokens:         5000,
				InputTokens:         2000,
				OutputTokens:        1000,
				CacheCreationTokens: 500,
				CacheReadTokens:     1500,
			},
		},
	}
	tm := []MonthStats{
		{Month: "2026-01", Commands: 50, SavedTokens: 1000, SavingsPct: 50.0},
	}
	result := mergeMonthly(cc, tm)
	if len(result) != 1 {
		t.Fatalf("mergeMonthly returned %d periods, want 1", len(result))
	}
	if result[0].CCCost == nil || *result[0].CCCost != 10.0 {
		t.Errorf("CCCost = %v, want 10.0", result[0].CCCost)
	}
	if result[0].TMCommands == nil || *result[0].TMCommands != 50 {
		t.Errorf("TMCommands = %v, want 50", result[0].TMCommands)
	}
}

func TestMergeMonthly_Sorted(t *testing.T) {
	cc := []ccusage.Period{
		{Key: "2026-03", Metrics: ccusage.Metrics{TotalCost: 30.0, TotalTokens: 3000, InputTokens: 1000, OutputTokens: 1000, CacheCreationTokens: 500, CacheReadTokens: 500}},
		{Key: "2026-01", Metrics: ccusage.Metrics{TotalCost: 10.0, TotalTokens: 1000, InputTokens: 500, OutputTokens: 200, CacheCreationTokens: 100, CacheReadTokens: 200}},
		{Key: "2026-02", Metrics: ccusage.Metrics{TotalCost: 20.0, TotalTokens: 2000, InputTokens: 800, OutputTokens: 500, CacheCreationTokens: 300, CacheReadTokens: 400}},
	}
	result := mergeMonthly(cc, nil)
	if len(result) != 3 {
		t.Fatalf("mergeMonthly returned %d periods, want 3", len(result))
	}
	if result[0].Label != "2026-01" || result[1].Label != "2026-02" || result[2].Label != "2026-03" {
		t.Errorf("mergeMonthly not sorted: %v", result)
	}
}

// ── mergeDaily ─────────────────────────────────────────────────

func TestMergeDaily_Empty(t *testing.T) {
	result := mergeDaily(nil, nil)
	if len(result) != 0 {
		t.Errorf("mergeDaily(nil, nil) returned %d periods, want 0", len(result))
	}
}

func TestMergeDaily_Merged(t *testing.T) {
	cc := []ccusage.Period{
		{
			Key: "2026-01-15",
			Metrics: ccusage.Metrics{
				TotalCost:           5.0,
				TotalTokens:         2000,
				InputTokens:         1000,
				OutputTokens:        500,
				CacheCreationTokens: 200,
				CacheReadTokens:     300,
			},
		},
	}
	tm := []DayStats{
		{Date: "2026-01-15", Commands: 25, SavedTokens: 500, SavingsPct: 40.0},
	}
	result := mergeDaily(cc, tm)
	if len(result) != 1 {
		t.Fatalf("mergeDaily returned %d periods, want 1", len(result))
	}
	if result[0].Label != "2026-01-15" {
		t.Errorf("Label = %q, want '2026-01-15'", result[0].Label)
	}
}

// ── mergeWeekly ────────────────────────────────────────────────

func TestMergeWeekly_Empty(t *testing.T) {
	result := mergeWeekly(nil, nil)
	if len(result) != 0 {
		t.Errorf("mergeWeekly(nil, nil) returned %d periods, want 0", len(result))
	}
}

func TestMergeWeekly_Merged(t *testing.T) {
	cc := []ccusage.Period{
		{
			Key: "2026-01-19",
			Metrics: ccusage.Metrics{
				TotalCost:           15.0,
				TotalTokens:         6000,
				InputTokens:         3000,
				OutputTokens:        1500,
				CacheCreationTokens: 600,
				CacheReadTokens:     900,
			},
		},
	}
	tm := []WeekStats{
		{WeekStart: "2026-01-17", Commands: 75, SavedTokens: 1500, SavingsPct: 55.0},
	}
	result := mergeWeekly(cc, tm)
	if len(result) != 1 {
		t.Fatalf("mergeWeekly returned %d periods, want 1", len(result))
	}
}

// ── formatUSD ──────────────────────────────────────────────────

func TestFormatUSD(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		expected string
	}{
		{"zero", 0, "$0.00"},
		{"small", 0.01, "$0.01"},
		{"medium", 1.5, "$1.50"},
		{"large", 100.0, "$100.00"},
		{"negative", -5.0, "$-5.00"},
		{"fractional", 0.123, "$0.12"},
		{"big", 9999.99, "$9999.99"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUSD(tt.amount)
			if result != tt.expected {
				t.Errorf("formatUSD(%.2f) = %q, want %q", tt.amount, result, tt.expected)
			}
		})
	}
}

// ── formatCPT ──────────────────────────────────────────────────

func TestFormatCPT(t *testing.T) {
	tests := []struct {
		name     string
		cpt      float64
		expected string
	}{
		{"zero", 0, "$0.000000/tok"},
		{"small", 0.000001, "$0.000001/tok"},
		{"typical", 0.000015, "$0.000015/tok"},
		{"larger", 0.002000, "$0.002000/tok"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCPT(tt.cpt)
			if result != tt.expected {
				t.Errorf("formatCPT(%f) = %q, want %q", tt.cpt, result, tt.expected)
			}
		})
	}
}

// ── formatOptionalFloat ────────────────────────────────────────

func TestFormatOptionalFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    *float64
		expected string
	}{
		{"nil", nil, ""},
		{"zero", floatPtr(0), "0.000000"},
		{"positive", floatPtr(1.5), "1.500000"},
		{"negative", floatPtr(-0.5), "-0.500000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatOptionalFloat(tt.input)
			if result != tt.expected {
				t.Errorf("formatOptionalFloat(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ── formatOptionalUint ─────────────────────────────────────────

func TestFormatOptionalUint(t *testing.T) {
	tests := []struct {
		name     string
		input    *uint64
		expected string
	}{
		{"nil", nil, ""},
		{"zero", uintPtr(0), "0"},
		{"positive", uintPtr(1000), "1000"},
		{"large", uintPtr(9999999999), "9999999999"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatOptionalUint(tt.input)
			if result != tt.expected {
				t.Errorf("formatOptionalUint(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ── formatOptionalInt ──────────────────────────────────────────

func TestFormatOptionalInt(t *testing.T) {
	tests := []struct {
		name     string
		input    *int
		expected string
	}{
		{"nil", nil, ""},
		{"zero", intPtr(0), "0"},
		{"positive", intPtr(100), "100"},
		{"negative", intPtr(-50), "-50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatOptionalInt(tt.input)
			if result != tt.expected {
				t.Errorf("formatOptionalInt(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ── alignWeekStart ─────────────────────────────────────────────

func TestAlignWeekStart(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"saturday to monday", "2026-01-17", "2026-01-19"},
		{"sunday to monday", "2026-01-18", "2026-01-19"},
		{"monday unchanged", "2026-01-19", "2026-01-19"},
		{"tuesday unchanged", "2026-01-20", "2026-01-20"},
		{"invalid date passthrough", "not-a-date", "not-a-date"},
		{"empty passthrough", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := alignWeekStart(tt.input)
			if result != tt.expected {
				t.Errorf("alignWeekStart(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ── RunOptions ─────────────────────────────────────────────────

func TestRunOptions_Defaults(t *testing.T) {
	opts := RunOptions{}
	if opts.Daily {
		t.Error("Daily should default to false")
	}
	if opts.Weekly {
		t.Error("Weekly should default to false")
	}
	if opts.Monthly {
		t.Error("Monthly should default to false")
	}
	if opts.All {
		t.Error("All should default to false")
	}
	if opts.Format != "" {
		t.Errorf("Format should default to empty, got %q", opts.Format)
	}
	if opts.Verbose {
		t.Error("Verbose should default to false")
	}
}

// ── PeriodEconomics ────────────────────────────────────────────

func TestPeriodEconomics_ZeroValue(t *testing.T) {
	p := PeriodEconomics{}
	if p.Label != "" {
		t.Errorf("Label = %q, want ''", p.Label)
	}
	if p.CCCost != nil {
		t.Error("CCCost should be nil")
	}
	if p.CCTotalTokens != nil {
		t.Error("CCTotalTokens should be nil")
	}
	if p.CCActiveTokens != nil {
		t.Error("CCActiveTokens should be nil")
	}
	if p.CCInputTokens != nil {
		t.Error("CCInputTokens should be nil")
	}
	if p.CCOutputTokens != nil {
		t.Error("CCOutputTokens should be nil")
	}
	if p.CCCacheCreateTokens != nil {
		t.Error("CCCacheCreateTokens should be nil")
	}
	if p.CCCacheReadTokens != nil {
		t.Error("CCCacheReadTokens should be nil")
	}
	if p.TMCommands != nil {
		t.Error("TMCommands should be nil")
	}
	if p.TMSavedTokens != nil {
		t.Error("TMSavedTokens should be nil")
	}
	if p.TMSavingsPct != nil {
		t.Error("TMSavingsPct should be nil")
	}
	if p.WeightedInputCPT != nil {
		t.Error("WeightedInputCPT should be nil")
	}
	if p.SavingsWeighted != nil {
		t.Error("SavingsWeighted should be nil")
	}
	if p.BlendedCPT != nil {
		t.Error("BlendedCPT should be nil")
	}
	if p.ActiveCPT != nil {
		t.Error("ActiveCPT should be nil")
	}
	if p.SavingsBlended != nil {
		t.Error("SavingsBlended should be nil")
	}
	if p.SavingsActive != nil {
		t.Error("SavingsActive should be nil")
	}
}

// ── Totals ─────────────────────────────────────────────────────

func TestTotals_ZeroValue(t *testing.T) {
	totals := Totals{}
	if totals.CCCost != 0 {
		t.Errorf("CCCost = %f, want 0", totals.CCCost)
	}
	if totals.CCTotalTokens != 0 {
		t.Errorf("CCTotalTokens = %d, want 0", totals.CCTotalTokens)
	}
	if totals.CCActiveTokens != 0 {
		t.Errorf("CCActiveTokens = %d, want 0", totals.CCActiveTokens)
	}
	if totals.CCInputTokens != 0 {
		t.Errorf("CCInputTokens = %d, want 0", totals.CCInputTokens)
	}
	if totals.CCOutputTokens != 0 {
		t.Errorf("CCOutputTokens = %d, want 0", totals.CCOutputTokens)
	}
	if totals.CCCacheCreate != 0 {
		t.Errorf("CCCacheCreate = %d, want 0", totals.CCCacheCreate)
	}
	if totals.CCCacheRead != 0 {
		t.Errorf("CCCacheRead = %d, want 0", totals.CCCacheRead)
	}
	if totals.TMCommands != 0 {
		t.Errorf("TMCommands = %d, want 0", totals.TMCommands)
	}
	if totals.TMSavedTokens != 0 {
		t.Errorf("TMSavedTokens = %d, want 0", totals.TMSavedTokens)
	}
	if totals.TMAvgSavingsPct != 0 {
		t.Errorf("TMAvgSavingsPct = %f, want 0", totals.TMAvgSavingsPct)
	}
}

// ── DayStats ───────────────────────────────────────────────────

func TestDayStats(t *testing.T) {
	ds := DayStats{
		Date:        "2026-01-15",
		Commands:    25,
		SavedTokens: 500,
		SavingsPct:  40.0,
	}
	if ds.Date != "2026-01-15" {
		t.Errorf("Date = %q, want '2026-01-15'", ds.Date)
	}
	if ds.Commands != 25 {
		t.Errorf("Commands = %d, want 25", ds.Commands)
	}
	if ds.SavedTokens != 500 {
		t.Errorf("SavedTokens = %d, want 500", ds.SavedTokens)
	}
	if ds.SavingsPct != 40.0 {
		t.Errorf("SavingsPct = %f, want 40.0", ds.SavingsPct)
	}
}

// ── MonthStats ─────────────────────────────────────────────────

func TestMonthStats(t *testing.T) {
	ms := MonthStats{
		Month:       "2026-01",
		Commands:    100,
		SavedTokens: 2000,
		SavingsPct:  50.0,
	}
	if ms.Month != "2026-01" {
		t.Errorf("Month = %q, want '2026-01'", ms.Month)
	}
	if ms.Commands != 100 {
		t.Errorf("Commands = %d, want 100", ms.Commands)
	}
}

// ── WeekStats ──────────────────────────────────────────────────

func TestWeekStats(t *testing.T) {
	ws := WeekStats{
		WeekStart:   "2026-01-19",
		Commands:    75,
		SavedTokens: 1500,
		SavingsPct:  55.0,
	}
	if ws.WeekStart != "2026-01-19" {
		t.Errorf("WeekStart = %q, want '2026-01-19'", ws.WeekStart)
	}
	if ws.Commands != 75 {
		t.Errorf("Commands = %d, want 75", ws.Commands)
	}
}

// ── computeWeightedMetrics Edge Cases ──────────────────────────

func TestComputeWeightedMetrics_Correctness(t *testing.T) {
	cost := 100.0
	saved := 10000
	input := uint64(100000)
	output := uint64(50000)
	cacheCreate := uint64(20000)
	cacheRead := uint64(30000)

	p := &PeriodEconomics{
		CCCost:              &cost,
		TMSavedTokens:       &saved,
		CCInputTokens:       &input,
		CCOutputTokens:      &output,
		CCCacheCreateTokens: &cacheCreate,
		CCCacheReadTokens:   &cacheRead,
	}
	p.computeWeightedMetrics()

	// weightedUnits = 100000 + 5*50000 + 1.25*20000 + 0.1*30000
	// = 100000 + 250000 + 25000 + 3000 = 378000
	// inputCPT = 100 / 378000
	expectedCPT := 100.0 / 378000.0
	if *p.WeightedInputCPT != expectedCPT {
		t.Errorf("WeightedInputCPT = %.10f, want %.10f", *p.WeightedInputCPT, expectedCPT)
	}

	expectedSavings := 10000.0 * expectedCPT
	if *p.SavingsWeighted != expectedSavings {
		t.Errorf("SavingsWeighted = %.10f, want %.10f", *p.SavingsWeighted, expectedSavings)
	}
}

// ── computeDualMetrics Edge Cases ──────────────────────────────

func TestComputeDualMetrics_Correctness(t *testing.T) {
	cost := 50.0
	saved := 5000
	total := uint64(100000)
	active := uint64(60000)

	p := &PeriodEconomics{
		CCCost:         &cost,
		TMSavedTokens:  &saved,
		CCTotalTokens:  &total,
		CCActiveTokens: &active,
	}
	p.computeDualMetrics()

	expectedBlended := 50.0 / 100000.0
	if *p.BlendedCPT != expectedBlended {
		t.Errorf("BlendedCPT = %.10f, want %.10f", *p.BlendedCPT, expectedBlended)
	}

	expectedActive := 50.0 / 60000.0
	if *p.ActiveCPT != expectedActive {
		t.Errorf("ActiveCPT = %.10f, want %.10f", *p.ActiveCPT, expectedActive)
	}

	expectedSavingsBlended := 5000.0 * expectedBlended
	if *p.SavingsBlended != expectedSavingsBlended {
		t.Errorf("SavingsBlended = %.10f, want %.10f", *p.SavingsBlended, expectedSavingsBlended)
	}

	expectedSavingsActive := 5000.0 * expectedActive
	if *p.SavingsActive != expectedSavingsActive {
		t.Errorf("SavingsActive = %.10f, want %.10f", *p.SavingsActive, expectedSavingsActive)
	}
}

// ── Benchmarks ─────────────────────────────────────────────────

func BenchmarkComputeWeightedMetrics(b *testing.B) {
	cost := 10.0
	saved := 1000
	input := uint64(1000)
	output := uint64(500)
	cacheCreate := uint64(100)
	cacheRead := uint64(200)

	p := &PeriodEconomics{
		CCCost:              &cost,
		TMSavedTokens:       &saved,
		CCInputTokens:       &input,
		CCOutputTokens:      &output,
		CCCacheCreateTokens: &cacheCreate,
		CCCacheReadTokens:   &cacheRead,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.computeWeightedMetrics()
	}
}

func BenchmarkComputeDualMetrics(b *testing.B) {
	cost := 10.0
	saved := 1000
	total := uint64(5000)
	active := uint64(3000)

	p := &PeriodEconomics{
		CCCost:         &cost,
		TMSavedTokens:  &saved,
		CCTotalTokens:  &total,
		CCActiveTokens: &active,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.computeDualMetrics()
	}
}

func BenchmarkComputeTotals(b *testing.B) {
	periods := make([]PeriodEconomics, 12)
	for i := range periods {
		cost := float64(i+1) * 10.0
		saved := (i + 1) * 1000
		cmds := (i + 1) * 50
		tokens := uint64((i + 1) * 5000)
		input := uint64((i + 1) * 2000)
		output := uint64((i + 1) * 1000)
		cacheCreate := uint64((i + 1) * 500)
		cacheRead := uint64((i + 1) * 1500)
		pct := 50.0
		periods[i] = PeriodEconomics{
			CCCost:              &cost,
			CCTotalTokens:       &tokens,
			CCActiveTokens:      &tokens,
			CCInputTokens:       &input,
			CCOutputTokens:      &output,
			CCCacheCreateTokens: &cacheCreate,
			CCCacheReadTokens:   &cacheRead,
			TMCommands:          &cmds,
			TMSavedTokens:       &saved,
			TMSavingsPct:        &pct,
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		computeTotals(periods)
	}
}

func BenchmarkFormatUSD(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatUSD(123.45)
	}
}

func BenchmarkFormatOptionalFloat(b *testing.B) {
	v := 3.14159
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatOptionalFloat(&v)
	}
}

func BenchmarkAlignWeekStart(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alignWeekStart("2026-01-17")
	}
}

// ── Helper functions ───────────────────────────────────────────

func floatPtr(f float64) *float64 { return &f }
func uintPtr(u uint64) *uint64    { return &u }
func intPtr(i int) *int           { return &i }

// ── mergeWeekly with ccusage.Period ────────────────────────────

func TestMergeWeekly_Sorted(t *testing.T) {
	cc := []ccusage.Period{
		{Key: "2026-01-26", Metrics: ccusage.Metrics{TotalCost: 30.0, TotalTokens: 3000, InputTokens: 1000, OutputTokens: 1000, CacheCreationTokens: 500, CacheReadTokens: 500}},
		{Key: "2026-01-12", Metrics: ccusage.Metrics{TotalCost: 10.0, TotalTokens: 1000, InputTokens: 500, OutputTokens: 200, CacheCreationTokens: 100, CacheReadTokens: 200}},
		{Key: "2026-01-19", Metrics: ccusage.Metrics{TotalCost: 20.0, TotalTokens: 2000, InputTokens: 800, OutputTokens: 500, CacheCreationTokens: 300, CacheReadTokens: 400}},
	}
	result := mergeWeekly(cc, nil)
	if len(result) != 3 {
		t.Fatalf("mergeWeekly returned %d periods, want 3", len(result))
	}
	if result[0].Label != "2026-01-12" || result[1].Label != "2026-01-19" || result[2].Label != "2026-01-26" {
		t.Errorf("mergeWeekly not sorted: %v", result)
	}
}

func TestMergeDaily_Sorted(t *testing.T) {
	cc := []ccusage.Period{
		{Key: "2026-01-17", Metrics: ccusage.Metrics{TotalCost: 30.0, TotalTokens: 3000, InputTokens: 1000, OutputTokens: 1000, CacheCreationTokens: 500, CacheReadTokens: 500}},
		{Key: "2026-01-15", Metrics: ccusage.Metrics{TotalCost: 10.0, TotalTokens: 1000, InputTokens: 500, OutputTokens: 200, CacheCreationTokens: 100, CacheReadTokens: 200}},
		{Key: "2026-01-16", Metrics: ccusage.Metrics{TotalCost: 20.0, TotalTokens: 2000, InputTokens: 800, OutputTokens: 500, CacheCreationTokens: 300, CacheReadTokens: 400}},
	}
	result := mergeDaily(cc, nil)
	if len(result) != 3 {
		t.Fatalf("mergeDaily returned %d periods, want 3", len(result))
	}
	if result[0].Label != "2026-01-15" || result[1].Label != "2026-01-16" || result[2].Label != "2026-01-17" {
		t.Errorf("mergeDaily not sorted: %v", result)
	}
}

// ── alignWeekStart more days ───────────────────────────────────

func TestAlignWeekStart_AllDays(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"wednesday", "2026-01-21", "2026-01-21"},
		{"thursday", "2026-01-22", "2026-01-22"},
		{"friday", "2026-01-23", "2026-01-23"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := alignWeekStart(tt.input)
			if result != tt.expected {
				t.Errorf("alignWeekStart(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ── Run function (with no DB) ──────────────────────────────────

func TestRun_DefaultFormat(t *testing.T) {
	// Run with default format (text) should not panic even without DB
	opts := RunOptions{}
	err := Run(opts)
	// It will fail because no tracking DB exists, but should not panic
	if err == nil {
		t.Log("Run succeeded (tracking DB exists)")
	} else {
		t.Logf("Run returned error (expected without DB): %v", err)
	}
}

func TestRun_JSONFormat(t *testing.T) {
	opts := RunOptions{Format: "json"}
	err := Run(opts)
	if err == nil {
		t.Log("Run JSON succeeded")
	} else {
		t.Logf("Run JSON returned error (expected without DB): %v", err)
	}
}

func TestRun_CSVFormat(t *testing.T) {
	opts := RunOptions{Format: "csv"}
	err := Run(opts)
	if err == nil {
		t.Log("Run CSV succeeded")
	} else {
		t.Logf("Run CSV returned error (expected without DB): %v", err)
	}
}

// ── ComputeTotals with weighted metrics ────────────────────────

func TestComputeTotals_WeightedMetricsCorrectness(t *testing.T) {
	cost := 100.0
	saved := 10000
	input := uint64(100000)
	output := uint64(50000)
	cacheCreate := uint64(20000)
	cacheRead := uint64(30000)
	tokens := uint64(200000)
	active := uint64(150000)
	pct := 50.0

	periods := []PeriodEconomics{
		{
			CCCost:              &cost,
			CCTotalTokens:       &tokens,
			CCActiveTokens:      &active,
			CCInputTokens:       &input,
			CCOutputTokens:      &output,
			CCCacheCreateTokens: &cacheCreate,
			CCCacheReadTokens:   &cacheRead,
			TMSavedTokens:       &saved,
			TMSavingsPct:        &pct,
		},
	}

	totals := computeTotals(periods)

	// weightedUnits = 100000 + 5*50000 + 1.25*20000 + 0.1*30000 = 378000
	expectedCPT := 100.0 / 378000.0
	if totals.WeightedInputCPT == nil {
		t.Fatal("WeightedInputCPT should not be nil")
	}
	if *totals.WeightedInputCPT != expectedCPT {
		t.Errorf("WeightedInputCPT = %.10f, want %.10f", *totals.WeightedInputCPT, expectedCPT)
	}

	expectedSavings := 10000.0 * expectedCPT
	if totals.SavingsWeighted == nil {
		t.Fatal("SavingsWeighted should not be nil")
	}
	if *totals.SavingsWeighted != expectedSavings {
		t.Errorf("SavingsWeighted = %.10f, want %.10f", *totals.SavingsWeighted, expectedSavings)
	}
}

// ── Time-related tests ─────────────────────────────────────────

func TestAlignWeekStart_FutureDate(t *testing.T) {
	// Test a future Saturday
	futureSat := time.Date(2027, 6, 12, 0, 0, 0, 0, time.UTC) // Saturday
	result := alignWeekStart(futureSat.Format("2006-01-02"))
	expected := futureSat.AddDate(0, 0, 2).Format("2006-01-02")
	if result != expected {
		t.Errorf("alignWeekStart(%q) = %q, want %q", futureSat.Format("2006-01-02"), result, expected)
	}
}

func TestAlignWeekStart_PastDate(t *testing.T) {
	// Test a past Sunday
	pastSun := time.Date(2025, 3, 2, 0, 0, 0, 0, time.UTC) // Sunday
	result := alignWeekStart(pastSun.Format("2006-01-02"))
	expected := pastSun.AddDate(0, 0, 1).Format("2006-01-02")
	if result != expected {
		t.Errorf("alignWeekStart(%q) = %q, want %q", pastSun.Format("2006-01-02"), result, expected)
	}
}
