package tracking

import (
	"math"
	"path/filepath"
	"testing"
	"time"
)

func insertCommandForDashboard(t *testing.T, tr *Tracker, record CommandRecord, ts time.Time) int64 {
	t.Helper()

	result, err := tr.db.Exec(`
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
		t.Fatalf("insert dashboard command: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("dashboard command last insert id: %v", err)
	}
	return id
}

func insertLayerStat(t *testing.T, tr *Tracker, commandID int64, layerName string, saved int) {
	t.Helper()
	_, err := tr.db.Exec(
		`INSERT INTO layer_stats (command_id, layer_name, tokens_saved, duration_us) VALUES (?, ?, ?, ?)`,
		commandID, layerName, saved, 100,
	)
	if err != nil {
		t.Fatalf("insert layer stat: %v", err)
	}
}

func TestDashboardSnapshotAggregatesCanonicalData(t *testing.T) {
	tr := newTestTracker(t)

	projectA := filepath.Join(t.TempDir(), "project-a")
	projectB := filepath.Join(t.TempDir(), "project-b")
	now := time.Now()

	id1 := insertCommandForDashboard(t, tr, CommandRecord{
		Command:        "git status",
		OriginalTokens: 1000,
		FilteredTokens: 400,
		SavedTokens:    600,
		ProjectPath:    projectA,
		SessionID:      "session-a",
		ExecTimeMs:     100,
		ParseSuccess:   true,
		AgentName:      "Claude Code",
		ModelName:      "claude-3-sonnet",
		Provider:       "Anthropic",
		ModelFamily:    "claude",
		ContextKind:    "read",
	}, now)
	id2 := insertCommandForDashboard(t, tr, CommandRecord{
		Command:        "git diff",
		OriginalTokens: 2000,
		FilteredTokens: 1000,
		SavedTokens:    1000,
		ProjectPath:    projectA,
		SessionID:      "session-b",
		ExecTimeMs:     200,
		ParseSuccess:   true,
		AgentName:      "Codex",
		ModelName:      "gpt-4-turbo",
		Provider:       "OpenAI",
		ModelFamily:    "gpt",
		ContextKind:    "mcp",
	}, now)
	_ = insertCommandForDashboard(t, tr, CommandRecord{
		Command:        "docker ps",
		OriginalTokens: 500,
		FilteredTokens: 250,
		SavedTokens:    250,
		ProjectPath:    projectB,
		SessionID:      "session-a",
		ExecTimeMs:     50,
		ParseSuccess:   false,
		AgentName:      "Claude Code",
		ModelName:      "claude-3-sonnet",
		Provider:       "Anthropic",
		ModelFamily:    "claude",
		ContextKind:    "read",
	}, now)

	insertLayerStat(t, tr, id1, "entropy", 200)
	insertLayerStat(t, tr, id2, "budget", 500)

	snapshot, err := tr.GetDashboardSnapshot(DashboardQueryOptions{
		Days:               30,
		ProjectPath:        projectA,
		Limit:              5,
		ReductionGoalPct:   40,
		DailyTokenBudget:   1200,
		WeeklyTokenBudget:  5000,
		MonthlyTokenBudget: 10000,
	})
	if err != nil {
		t.Fatalf("GetDashboardSnapshot: %v", err)
	}

	if snapshot.Overview.TotalCommands != 2 {
		t.Fatalf("TotalCommands = %d, want 2", snapshot.Overview.TotalCommands)
	}
	if snapshot.Overview.TotalSavedTokens != 1600 {
		t.Fatalf("TotalSavedTokens = %d, want 1600", snapshot.Overview.TotalSavedTokens)
	}
	if snapshot.Overview.UniqueAgents != 2 || snapshot.Overview.UniqueProviders != 2 || snapshot.Overview.UniqueModels != 2 {
		t.Fatalf("unexpected uniqueness counts: %+v", snapshot.Overview)
	}
	if snapshot.Overview.UniqueSessions != 2 {
		t.Fatalf("UniqueSessions = %d, want 2", snapshot.Overview.UniqueSessions)
	}
	if snapshot.Overview.ParseSuccessRatePct != 100 {
		t.Fatalf("ParseSuccessRatePct = %v, want 100", snapshot.Overview.ParseSuccessRatePct)
	}
	if math.Abs(snapshot.Overview.EstimatedSavingsUSD-0.0118) > 0.00001 {
		t.Fatalf("EstimatedSavingsUSD = %v, want about 0.0118", snapshot.Overview.EstimatedSavingsUSD)
	}

	if len(snapshot.DailyTrends) != 1 {
		t.Fatalf("len(DailyTrends) = %d, want 1", len(snapshot.DailyTrends))
	}
	if snapshot.DailyTrends[0].SavedTokens != 1600 {
		t.Fatalf("DailyTrends[0].SavedTokens = %d, want 1600", snapshot.DailyTrends[0].SavedTokens)
	}

	if len(snapshot.TopProviders) != 2 {
		t.Fatalf("len(TopProviders) = %d, want 2", len(snapshot.TopProviders))
	}
	if snapshot.TopProviders[0].Key != "OpenAI" {
		t.Fatalf("TopProviders[0].Key = %q, want OpenAI", snapshot.TopProviders[0].Key)
	}
	if len(snapshot.TopProviderModels) != 2 {
		t.Fatalf("len(TopProviderModels) = %d, want 2", len(snapshot.TopProviderModels))
	}
	if snapshot.TopProviderModels[0].Key != "OpenAI / gpt-4-turbo" {
		t.Fatalf("TopProviderModels[0].Key = %q, want OpenAI / gpt-4-turbo", snapshot.TopProviderModels[0].Key)
	}

	if len(snapshot.TopLayers) != 2 {
		t.Fatalf("len(TopLayers) = %d, want 2", len(snapshot.TopLayers))
	}
	if snapshot.TopLayers[0].LayerName != "budget" {
		t.Fatalf("TopLayers[0].LayerName = %q, want budget", snapshot.TopLayers[0].LayerName)
	}
	if snapshot.Budgets.Daily.FilteredTokens != 1400 {
		t.Fatalf("Daily.FilteredTokens = %d, want 1400", snapshot.Budgets.Daily.FilteredTokens)
	}
	if !snapshot.Budgets.Daily.OverTokenBudget {
		t.Fatal("expected daily token budget to be exceeded")
	}
	if snapshot.Streaks.SavingsDays != 1 || snapshot.Streaks.GoalDays != 1 {
		t.Fatalf("unexpected streaks: %+v", snapshot.Streaks)
	}
	if snapshot.Lifecycle.CommandsTotal != 2 {
		t.Fatalf("Lifecycle.CommandsTotal = %d, want 2", snapshot.Lifecycle.CommandsTotal)
	}
	if snapshot.Lifecycle.ProjectsCount != 1 {
		t.Fatalf("Lifecycle.ProjectsCount = %d, want 1", snapshot.Lifecycle.ProjectsCount)
	}
	if snapshot.Lifecycle.ActiveDays30d != 1 {
		t.Fatalf("Lifecycle.ActiveDays30d = %d, want 1", snapshot.Lifecycle.ActiveDays30d)
	}
	if math.Abs(snapshot.Lifecycle.AvgSavedTokensPerExec-800) > 0.00001 {
		t.Fatalf("Lifecycle.AvgSavedTokensPerExec = %v, want 800", snapshot.Lifecycle.AvgSavedTokensPerExec)
	}
	if len(snapshot.LowSavingsCommands) != 2 {
		t.Fatalf("len(LowSavingsCommands) = %d, want 2", len(snapshot.LowSavingsCommands))
	}
	if snapshot.LowSavingsCommands[0].Key != "git diff" {
		t.Fatalf("LowSavingsCommands[0].Key = %q, want git diff", snapshot.LowSavingsCommands[0].Key)
	}
	if math.Abs(snapshot.LowSavingsCommands[0].ReductionPct-50) > 0.00001 {
		t.Fatalf("LowSavingsCommands[0].ReductionPct = %v, want 50", snapshot.LowSavingsCommands[0].ReductionPct)
	}
	if snapshot.Gamification.Points == 0 || snapshot.Gamification.Level < 1 {
		t.Fatalf("unexpected gamification: %+v", snapshot.Gamification)
	}
}

func TestDashboardBreakdownSupportsFiltersAndRejectsUnknownDimensions(t *testing.T) {
	tr := newTestTracker(t)
	now := time.Now()
	project := filepath.Join(t.TempDir(), "project")

	insertCommandForDashboard(t, tr, CommandRecord{
		Command:        "git status",
		OriginalTokens: 100,
		FilteredTokens: 40,
		SavedTokens:    60,
		ProjectPath:    project,
		SessionID:      "session-a",
		ParseSuccess:   true,
		AgentName:      "Claude Code",
		ModelName:      "claude-3-sonnet",
		Provider:       "Anthropic",
		ContextKind:    "read",
	}, now)
	insertCommandForDashboard(t, tr, CommandRecord{
		Command:        "git diff",
		OriginalTokens: 100,
		FilteredTokens: 10,
		SavedTokens:    90,
		ProjectPath:    project,
		SessionID:      "session-b",
		ParseSuccess:   true,
		AgentName:      "Codex",
		ModelName:      "gpt-4-turbo",
		Provider:       "OpenAI",
		ContextKind:    "mcp",
	}, now)

	items, err := tr.GetDashboardBreakdown("agent", DashboardQueryOptions{
		Days:        30,
		ProjectPath: project,
		AgentName:   "Claude Code",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("GetDashboardBreakdown(agent): %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].Key != "Claude Code" || items[0].SavedTokens != 60 {
		t.Fatalf("unexpected filtered breakdown row: %+v", items[0])
	}

	contexts, err := tr.GetDashboardBreakdown("context_kind", DashboardQueryOptions{
		Days:        30,
		ProjectPath: project,
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("GetDashboardBreakdown(context_kind): %v", err)
	}
	if len(contexts) != 2 {
		t.Fatalf("len(contexts) = %d, want 2", len(contexts))
	}

	if _, err := tr.GetDashboardBreakdown("bad-dimension", DashboardQueryOptions{}); err == nil {
		t.Fatal("expected invalid breakdown dimension to fail")
	}
	if _, err := tr.GetDashboardTrends("quarter", DashboardQueryOptions{}); err == nil {
		t.Fatal("expected invalid trend granularity to fail")
	}
}
