package tracking

import (
	"path/filepath"
	"testing"
	"time"
)

func newTestTracker(t *testing.T) *Tracker {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "tokman.db")
	tr, err := NewTracker(dbPath)
	if err != nil {
		t.Fatalf("new tracker: %v", err)
	}
	t.Cleanup(func() {
		_ = tr.Close()
	})
	return tr
}

func insertCommandAt(t *testing.T, tr *Tracker, record CommandRecord, ts time.Time) {
	t.Helper()

	_, err := tr.db.Exec(`
		INSERT INTO commands (
			command, original_output, filtered_output,
			original_tokens, filtered_tokens, saved_tokens,
			project_path, session_id, exec_time_ms, timestamp, parse_success,
			agent_name, model_name, provider, model_family,
			context_kind, context_mode, context_resolved_mode,
			context_target, context_related_files, context_bundle
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		record.Command,
		record.OriginalOutput,
		record.FilteredOutput,
		record.OriginalTokens,
		record.FilteredTokens,
		record.SavedTokens,
		normalizeProjectPath(record.ProjectPath),
		record.SessionID,
		record.ExecTimeMs,
		ts.Format(time.RFC3339),
		record.ParseSuccess,
		record.AgentName,
		record.ModelName,
		record.Provider,
		record.ModelFamily,
		record.ContextKind,
		record.ContextMode,
		record.ContextResolvedMode,
		record.ContextTarget,
		record.ContextRelatedFiles,
		record.ContextBundle,
	)
	if err != nil {
		t.Fatalf("insert command: %v", err)
	}
}

func TestAnalyticsUseCommandsTableAndProjectScope(t *testing.T) {
	tr := newTestTracker(t)

	projectA := filepath.Join(t.TempDir(), "project-a")
	projectB := filepath.Join(t.TempDir(), "project-b")

	now := time.Now()
	insertCommandAt(t, tr, CommandRecord{
		Command:        "git status",
		OriginalTokens: 1000,
		FilteredTokens: 400,
		SavedTokens:    600,
		ProjectPath:    projectA,
		ExecTimeMs:     120,
		ParseSuccess:   true,
	}, now)
	insertCommandAt(t, tr, CommandRecord{
		Command:        "git diff",
		OriginalTokens: 800,
		FilteredTokens: 300,
		SavedTokens:    500,
		ProjectPath:    projectA,
		ExecTimeMs:     80,
		ParseSuccess:   true,
	}, now.Add(-24*time.Hour))
	insertCommandAt(t, tr, CommandRecord{
		Command:        "docker ps",
		OriginalTokens: 600,
		FilteredTokens: 500,
		SavedTokens:    100,
		ProjectPath:    projectB,
		ExecTimeMs:     50,
		ParseSuccess:   true,
	}, now)

	daily, err := tr.GetDailyStats(30, projectA)
	if err != nil {
		t.Fatalf("GetDailyStats: %v", err)
	}
	if len(daily) != 2 {
		t.Fatalf("expected 2 daily rows for project A, got %d", len(daily))
	}

	breakdown, err := tr.GetCommandBreakdown(10, projectA)
	if err != nil {
		t.Fatalf("GetCommandBreakdown: %v", err)
	}
	if len(breakdown) != 2 {
		t.Fatalf("expected 2 command rows for project A, got %d", len(breakdown))
	}

	summary, err := tr.GetFullGainSummary(GainSummaryOptions{
		ProjectPath:    projectA,
		IncludeDaily:   true,
		IncludeWeekly:  true,
		IncludeMonthly: true,
		IncludeHistory: true,
	})
	if err != nil {
		t.Fatalf("GetFullGainSummary: %v", err)
	}
	if summary.TotalCommands != 2 {
		t.Fatalf("TotalCommands = %d, want 2", summary.TotalCommands)
	}
	if summary.TotalSaved != 1100 {
		t.Fatalf("TotalSaved = %d, want 1100", summary.TotalSaved)
	}
	if summary.TotalExecTimeMs != 200 {
		t.Fatalf("TotalExecTimeMs = %d, want 200", summary.TotalExecTimeMs)
	}
	if len(summary.RecentCommands) != 2 {
		t.Fatalf("RecentCommands = %d, want 2", len(summary.RecentCommands))
	}
}

func TestGetDailySavingsWithoutProjectFilterIncludesAllProjects(t *testing.T) {
	tr := newTestTracker(t)

	now := time.Now()
	insertCommandAt(t, tr, CommandRecord{
		Command:        "git status",
		OriginalTokens: 100,
		FilteredTokens: 40,
		SavedTokens:    60,
		ProjectPath:    filepath.Join(t.TempDir(), "a"),
		ParseSuccess:   true,
	}, now)
	insertCommandAt(t, tr, CommandRecord{
		Command:        "git diff",
		OriginalTokens: 200,
		FilteredTokens: 50,
		SavedTokens:    150,
		ProjectPath:    filepath.Join(t.TempDir(), "b"),
		ParseSuccess:   true,
	}, now)

	daily, err := tr.GetDailySavings("", 7)
	if err != nil {
		t.Fatalf("GetDailySavings: %v", err)
	}
	if len(daily) == 0 {
		t.Fatal("expected at least one daily savings row")
	}
	if daily[0].Saved != 210 {
		t.Fatalf("Saved = %d, want 210", daily[0].Saved)
	}
	if daily[0].Commands != 2 {
		t.Fatalf("Commands = %d, want 2", daily[0].Commands)
	}
}

func TestGenerateCostReportAndAlertsHandleEdgeCases(t *testing.T) {
	tr := newTestTracker(t)

	report, err := tr.GenerateCostReport("default", 0)
	if err != nil {
		t.Fatalf("GenerateCostReport on empty db: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil cost report")
	}
	if report.TotalTokensSaved != 0 || report.EstimatedSavings != 0 {
		t.Fatalf("unexpected non-zero empty report: %+v", report)
	}
	if report.Projections.MonthlyEstimate != 0 || report.Projections.YearlyEstimate != 0 {
		t.Fatalf("unexpected non-zero projections: %+v", report.Projections)
	}

	insertCommandAt(t, tr, CommandRecord{
		Command:        "git status",
		OriginalTokens: 1000,
		FilteredTokens: 500,
		SavedTokens:    500,
		ProjectPath:    filepath.Join(t.TempDir(), "alerts"),
		ParseSuccess:   true,
	}, time.Now())

	alerts, err := tr.CheckAlert(AlertThreshold{
		DailyTokenLimit:  100,
		WeeklyTokenLimit: 1000,
	})
	if err != nil {
		t.Fatalf("CheckAlert: %v", err)
	}
	if len(alerts) == 0 {
		t.Fatal("expected at least one alert")
	}
	foundDaily := false
	for _, alert := range alerts {
		if alert.Type == "daily_tokens" {
			foundDaily = true
		}
	}
	if !foundDaily {
		t.Fatal("expected daily_tokens alert")
	}
}
