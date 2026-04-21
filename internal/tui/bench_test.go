package tui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/GrayCodeAI/tok/internal/session"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

// These benchmarks establish the performance envelope for the TUI's
// hot paths so regressions show up in CI before they ship. Targets:
//
//   BenchmarkBrailleLineChart_Wide       <  250 µs/op
//   BenchmarkTableRender_1000Rows        <  6 ms/op
//   BenchmarkModelView_FullFrame         < 16 ms/op (single frame budget)
//
// Numbers are indicative — adjust when the hardware or lipgloss
// version changes materially. The absolute numbers matter less than
// the delta between runs.

func BenchmarkBrailleLineChart_Wide(b *testing.B) {
	// 180 samples is the default 30-day window × ~6 points per day if
	// we ever switch to sub-daily buckets. Width 180 so samples are
	// 1:1 with columns — the worst case for the render loop.
	samples := make([]float64, 180)
	for i := range samples {
		samples[i] = float64(i*37%1000) + 100
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = BrailleLineChart(samples, 180, 8)
	}
}

func BenchmarkASCIILineChart_Wide(b *testing.B) {
	samples := make([]float64, 180)
	for i := range samples {
		samples[i] = float64(i*37%1000) + 100
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = asciiLineChart(samples, 180, 8)
	}
}

func BenchmarkSparkline(b *testing.B) {
	values := make([]int64, 90)
	for i := range values {
		values[i] = int64(i*13) % 500
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = sparklineGlyphs(values, true)
	}
}

func BenchmarkTableRender_1000Rows(b *testing.B) {
	tbl := NewTable([]Column{
		{Title: "Key", MinWidth: 20, Sortable: true},
		{Title: "Commands", MinWidth: 10, Numeric: true, Sortable: true, Align: AlignRight},
		{Title: "Saved", MinWidth: 10, Numeric: true, Sortable: true, Align: AlignRight, Accent: true},
		{Title: "Reduction", MinWidth: 10, Numeric: true, Sortable: true, Align: AlignRight},
	})
	rows := make([]Row, 1000)
	for i := range rows {
		rows[i] = Row{Cells: []string{
			fmt.Sprintf("command-%04d", i),
			fmt.Sprintf("%d", i*3),
			fmt.Sprintf("%d", (i*37)%9000),
			fmt.Sprintf("%.1f", float64(i%100)),
		}}
	}
	tbl.SetRows(rows)
	th := newTheme()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// View height 30 is typical for a terminal with padding.
		_ = tbl.View(th, 120, 30)
	}
}

func BenchmarkTableFilter_1000Rows(b *testing.B) {
	tbl := NewTable([]Column{
		{Title: "Key", MinWidth: 20, Sortable: true},
	})
	rows := make([]Row, 1000)
	for i := range rows {
		rows[i] = Row{Cells: []string{fmt.Sprintf("item-%04d-segment-%d", i, i%7)}}
	}
	tbl.SetRows(rows)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tbl.SetFilter(fmt.Sprintf("%d", i%9))
	}
}

// benchSnapshot builds a realistic WorkspaceDashboardSnapshot sized
// like a heavy-user account — enough rows to exercise every section's
// rendering path. Returned value is reusable across benchmarks.
func benchSnapshot() *tracking.WorkspaceDashboardSnapshot {
	daily := make([]tracking.DashboardTrendPoint, 30)
	weekly := make([]tracking.DashboardTrendPoint, 12)
	for i := range daily {
		daily[i] = tracking.DashboardTrendPoint{
			Period:       fmt.Sprintf("2026-04-%02d", i+1),
			Commands:     int64(i*7 + 20),
			SavedTokens:  int64(i*200 + 1000),
			ReductionPct: float64(i%50 + 30),
		}
	}
	for i := range weekly {
		weekly[i] = tracking.DashboardTrendPoint{
			Period: fmt.Sprintf("W%02d", i+1), Commands: int64(i * 40),
			SavedTokens: int64(i * 6000), ReductionPct: float64(i%20 + 50),
		}
	}
	providers := make([]tracking.DashboardBreakdown, 8)
	for i := range providers {
		providers[i] = tracking.DashboardBreakdown{
			Key: fmt.Sprintf("provider-%d", i), Commands: int64(i * 20),
			SavedTokens: int64(i * 1500), ReductionPct: float64(i*5 + 30),
		}
	}
	layers := make([]tracking.DashboardLayerSummary, 20)
	for i := range layers {
		layers[i] = tracking.DashboardLayerSummary{
			LayerName: fmt.Sprintf("layer-%02d", i), CallCount: int64(100 - i*3),
			TotalSaved: int64(i * 300), AvgSaved: float64(i * 5),
		}
	}
	return &tracking.WorkspaceDashboardSnapshot{
		Dashboard: &tracking.DashboardSnapshot{
			Overview:     tracking.DashboardOverview{TotalSavedTokens: 999_999, TotalCommands: 1234},
			DailyTrends:  daily,
			WeeklyTrends: weekly,
			TopProviders: providers,
			TopModels:    providers,
			TopAgents:    providers,
			TopCommands:  providers,
			TopLayers:    layers,
			Streaks:      tracking.DashboardStreaks{SavingsDays: 5, GoalDays: 7},
			Gamification: tracking.DashboardGamification{Points: 1000, Level: 3},
		},
		Sessions: &session.SessionAnalyticsSnapshot{
			StoreSummary: session.SessionStoreSummary{TotalSessions: 10, ActiveSessions: 1, TopAgent: "claude"},
		},
	}
}

func BenchmarkModelView_FullFrame(b *testing.B) {
	loader := &stubLoader{snapshot: benchSnapshot()}
	m := NewModelWithLoader(Options{Theme: ThemeDark}, loader).(model)
	// Prime the model the way tea.Program would: WindowSize, then a
	// snapshotLoadedMsg, then pretend a few frames ticked.
	next, _ := m.Update(tea.WindowSizeMsg{Width: 180, Height: 50})
	m = next.(model)
	next, _ = m.Update(snapshotLoadedMsg{snapshot: loader.snapshot, loadedAt: time.Now()})
	m = next.(model)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}

func BenchmarkModelView_Compact(b *testing.B) {
	loader := &stubLoader{snapshot: benchSnapshot()}
	m := NewModelWithLoader(Options{Theme: ThemeDark}, loader).(model)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 20})
	m = next.(model)
	next, _ = m.Update(snapshotLoadedMsg{snapshot: loader.snapshot, loadedAt: time.Now()})
	m = next.(model)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = m.View()
	}
}

func BenchmarkPaletteFuzzySearch(b *testing.B) {
	p := NewPalette(
		DefaultActionRegistry(ActionDeps{SectionCount: 12}),
		defaultSections(),
	)
	p.Open()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p.applyFilter("re")
	}
}
