package tracking

import (
	"fmt"
	"strings"
	"time"
)

// DashboardQueryOptions controls the window and filters used by dashboard data queries.
type DashboardQueryOptions struct {
	Days                 int
	ProjectPath          string
	AgentName            string
	Provider             string
	ModelName            string
	SessionID            string
	Limit                int
	ReductionGoalPct     float64
	DailyTokenBudget     int64
	WeeklyTokenBudget    int64
	MonthlyTokenBudget   int64
	DailyCostBudgetUSD   float64
	WeeklyCostBudgetUSD  float64
	MonthlyCostBudgetUSD float64
}

// DashboardOverview contains top-line metrics for a dashboard window.
type DashboardOverview struct {
	WindowDays               int     `json:"window_days"`
	TotalCommands            int64   `json:"total_commands"`
	TotalOriginalTokens      int64   `json:"total_original_tokens"`
	TotalFilteredTokens      int64   `json:"total_filtered_tokens"`
	TotalSavedTokens         int64   `json:"total_saved_tokens"`
	ReductionPct             float64 `json:"reduction_pct"`
	AvgExecTimeMs            float64 `json:"avg_exec_time_ms"`
	ParseSuccessRatePct      float64 `json:"parse_success_rate_pct"`
	UniqueAgents             int64   `json:"unique_agents"`
	UniqueProviders          int64   `json:"unique_providers"`
	UniqueModels             int64   `json:"unique_models"`
	UniqueSessions           int64   `json:"unique_sessions"`
	EstimatedOriginalCostUSD float64 `json:"estimated_original_cost_usd"`
	EstimatedFilteredCostUSD float64 `json:"estimated_filtered_cost_usd"`
	EstimatedSavingsUSD      float64 `json:"estimated_savings_usd"`
}

// DashboardTrendPoint contains one time-bucket of dashboard trend data.
type DashboardTrendPoint struct {
	Period                   string  `json:"period"`
	Commands                 int64   `json:"commands"`
	OriginalTokens           int64   `json:"original_tokens"`
	FilteredTokens           int64   `json:"filtered_tokens"`
	SavedTokens              int64   `json:"saved_tokens"`
	ReductionPct             float64 `json:"reduction_pct"`
	EstimatedOriginalCostUSD float64 `json:"estimated_original_cost_usd"`
	EstimatedFilteredCostUSD float64 `json:"estimated_filtered_cost_usd"`
	EstimatedSavingsUSD      float64 `json:"estimated_savings_usd"`
}

// DashboardBreakdown contains aggregate metrics for one analytics dimension.
type DashboardBreakdown struct {
	Key                      string  `json:"key"`
	Commands                 int64   `json:"commands"`
	OriginalTokens           int64   `json:"original_tokens"`
	FilteredTokens           int64   `json:"filtered_tokens"`
	SavedTokens              int64   `json:"saved_tokens"`
	ReductionPct             float64 `json:"reduction_pct"`
	EstimatedOriginalCostUSD float64 `json:"estimated_original_cost_usd"`
	EstimatedFilteredCostUSD float64 `json:"estimated_filtered_cost_usd"`
	EstimatedSavingsUSD      float64 `json:"estimated_savings_usd"`
}

// DashboardLayerSummary contains layer-level token savings.
type DashboardLayerSummary struct {
	LayerName  string  `json:"layer_name"`
	CallCount  int64   `json:"call_count"`
	TotalSaved int64   `json:"total_saved"`
	AvgSaved   float64 `json:"avg_saved"`
}

// DashboardBudgetWindow contains budget/usage state for a fixed planning period.
type DashboardBudgetWindow struct {
	Window                   string  `json:"window"`
	Days                     int     `json:"days"`
	OriginalTokens           int64   `json:"original_tokens"`
	FilteredTokens           int64   `json:"filtered_tokens"`
	SavedTokens              int64   `json:"saved_tokens"`
	EstimatedOriginalCostUSD float64 `json:"estimated_original_cost_usd"`
	EstimatedFilteredCostUSD float64 `json:"estimated_filtered_cost_usd"`
	EstimatedSavingsUSD      float64 `json:"estimated_savings_usd"`
	TokenBudget              int64   `json:"token_budget"`
	CostBudgetUSD            float64 `json:"cost_budget_usd"`
	TokenRemaining           int64   `json:"token_remaining"`
	CostRemainingUSD         float64 `json:"cost_remaining_usd"`
	TokenUtilizationPct      float64 `json:"token_utilization_pct"`
	CostUtilizationPct       float64 `json:"cost_utilization_pct"`
	OverTokenBudget          bool    `json:"over_token_budget"`
	OverCostBudget           bool    `json:"over_cost_budget"`
}

// DashboardBudgetStatus groups token/cost budget windows.
type DashboardBudgetStatus struct {
	Daily   DashboardBudgetWindow `json:"daily"`
	Weekly  DashboardBudgetWindow `json:"weekly"`
	Monthly DashboardBudgetWindow `json:"monthly"`
}

// DashboardStreaks captures consecutive saving/efficiency days.
type DashboardStreaks struct {
	SavingsDays         int     `json:"savings_days"`
	GoalDays            int     `json:"goal_days"`
	GoalReductionPct    float64 `json:"goal_reduction_pct"`
	BestDay             string  `json:"best_day"`
	BestDaySavedTokens  int64   `json:"best_day_saved_tokens"`
	BestDayReductionPct float64 `json:"best_day_reduction_pct"`
}

// DashboardLifecycle captures retention and history depth signals.
type DashboardLifecycle struct {
	FirstSeenDate         string  `json:"first_seen_date"`
	DaysSinceFirstUse     int     `json:"days_since_first_use"`
	ActiveDays30d         int     `json:"active_days_30d"`
	CommandsTotal         int64   `json:"commands_total"`
	ProjectsCount         int64   `json:"projects_count"`
	AvgSavedTokensPerExec float64 `json:"avg_saved_tokens_per_exec"`
}

// DashboardGamification summarizes points, level, and badges.
type DashboardGamification struct {
	Points          int64    `json:"points"`
	Level           int      `json:"level"`
	NextLevelPoints int64    `json:"next_level_points"`
	Badges          []string `json:"badges"`
}

// DashboardSnapshot is the canonical aggregate payload for the future TUI.
type DashboardSnapshot struct {
	Overview           DashboardOverview       `json:"overview"`
	DailyTrends        []DashboardTrendPoint   `json:"daily_trends"`
	WeeklyTrends       []DashboardTrendPoint   `json:"weekly_trends"`
	TopAgents          []DashboardBreakdown    `json:"top_agents"`
	TopProviders       []DashboardBreakdown    `json:"top_providers"`
	TopModels          []DashboardBreakdown    `json:"top_models"`
	TopProviderModels  []DashboardBreakdown    `json:"top_provider_models"`
	TopProjects        []DashboardBreakdown    `json:"top_projects"`
	TopCommands        []DashboardBreakdown    `json:"top_commands"`
	TopSessions        []DashboardBreakdown    `json:"top_sessions"`
	ContextKinds       []DashboardBreakdown    `json:"context_kinds"`
	TopLayers          []DashboardLayerSummary `json:"top_layers"`
	LowSavingsCommands []DashboardBreakdown    `json:"low_savings_commands"`
	Budgets            DashboardBudgetStatus   `json:"budgets"`
	Streaks            DashboardStreaks        `json:"streaks"`
	Lifecycle          DashboardLifecycle      `json:"lifecycle"`
	Gamification       DashboardGamification   `json:"gamification"`
}

// GetDashboardSnapshot returns a stable aggregate view suitable for TUI consumption.
func (t *Tracker) GetDashboardSnapshot(opts DashboardQueryOptions) (*DashboardSnapshot, error) {
	overview, err := t.GetDashboardOverview(opts)
	if err != nil {
		return nil, err
	}

	daily, err := t.GetDashboardTrends("day", opts)
	if err != nil {
		return nil, err
	}
	weekly, err := t.GetDashboardTrends("week", opts)
	if err != nil {
		return nil, err
	}

	agents, err := t.GetDashboardBreakdown("agent", opts)
	if err != nil {
		return nil, err
	}
	providers, err := t.GetDashboardBreakdown("provider", opts)
	if err != nil {
		return nil, err
	}
	models, err := t.GetDashboardBreakdown("model", opts)
	if err != nil {
		return nil, err
	}
	providerModels, err := t.GetDashboardBreakdown("provider_model", opts)
	if err != nil {
		return nil, err
	}
	projects, err := t.GetDashboardBreakdown("project", opts)
	if err != nil {
		return nil, err
	}
	commands, err := t.GetDashboardBreakdown("command", opts)
	if err != nil {
		return nil, err
	}
	sessions, err := t.GetDashboardBreakdown("session", opts)
	if err != nil {
		return nil, err
	}
	contextKinds, err := t.GetDashboardBreakdown("context_kind", opts)
	if err != nil {
		return nil, err
	}
	layers, err := t.GetDashboardTopLayers(opts)
	if err != nil {
		return nil, err
	}
	lowSavings, err := t.GetDashboardLowSavingsCommands(opts)
	if err != nil {
		return nil, err
	}
	budgets, err := t.GetDashboardBudgets(opts)
	if err != nil {
		return nil, err
	}
	streaks, err := t.GetDashboardStreaks(opts)
	if err != nil {
		return nil, err
	}
	lifecycle, err := t.GetDashboardLifecycle(opts)
	if err != nil {
		return nil, err
	}
	gamification := buildDashboardGamification(overview, streaks)

	return &DashboardSnapshot{
		Overview:           overview,
		DailyTrends:        daily,
		WeeklyTrends:       weekly,
		TopAgents:          agents,
		TopProviders:       providers,
		TopModels:          models,
		TopProviderModels:  providerModels,
		TopProjects:        projects,
		TopCommands:        commands,
		TopSessions:        sessions,
		ContextKinds:       contextKinds,
		TopLayers:          layers,
		LowSavingsCommands: lowSavings,
		Budgets:            budgets,
		Streaks:            streaks,
		Lifecycle:          lifecycle,
		Gamification:       gamification,
	}, nil
}

// GetDashboardOverview returns dashboard KPIs for the requested window.
func (t *Tracker) GetDashboardOverview(opts DashboardQueryOptions) (DashboardOverview, error) {
	days := normalizeDashboardDays(opts.Days)
	where, args := buildDashboardFilters(opts, days)

	query := `
		SELECT
			COUNT(*) as total_commands,
			COALESCE(SUM(original_tokens), 0) as total_original_tokens,
			COALESCE(SUM(filtered_tokens), 0) as total_filtered_tokens,
			COALESCE(SUM(saved_tokens), 0) as total_saved_tokens,
			COALESCE(AVG(exec_time_ms), 0) as avg_exec_time_ms,
			COALESCE(SUM(CASE WHEN parse_success = 1 THEN 1 ELSE 0 END), 0) as parse_success_count,
			COUNT(DISTINCT NULLIF(TRIM(agent_name), '')) as unique_agents,
			COUNT(DISTINCT NULLIF(TRIM(provider), '')) as unique_providers,
			COUNT(DISTINCT NULLIF(TRIM(model_name), '')) as unique_models,
			COUNT(DISTINCT NULLIF(TRIM(session_id), '')) as unique_sessions
		FROM commands
	`
	if where != "" {
		query += " WHERE " + where
	}

	var overview DashboardOverview
	var parseSuccessCount int64
	if err := t.db.QueryRow(query, args...).Scan(
		&overview.TotalCommands,
		&overview.TotalOriginalTokens,
		&overview.TotalFilteredTokens,
		&overview.TotalSavedTokens,
		&overview.AvgExecTimeMs,
		&parseSuccessCount,
		&overview.UniqueAgents,
		&overview.UniqueProviders,
		&overview.UniqueModels,
		&overview.UniqueSessions,
	); err != nil {
		return DashboardOverview{}, fmt.Errorf("dashboard overview query: %w", err)
	}

	overview.WindowDays = days
	if overview.TotalOriginalTokens > 0 {
		overview.ReductionPct = float64(overview.TotalSavedTokens) / float64(overview.TotalOriginalTokens) * 100
	}
	if overview.TotalCommands > 0 {
		overview.ParseSuccessRatePct = float64(parseSuccessCount) / float64(overview.TotalCommands) * 100
	}

	costs, err := t.getDashboardCostSummary(opts)
	if err != nil {
		return DashboardOverview{}, err
	}
	overview.EstimatedOriginalCostUSD = costs.original
	overview.EstimatedFilteredCostUSD = costs.filtered
	overview.EstimatedSavingsUSD = costs.saved

	return overview, nil
}

// GetDashboardTrends returns grouped time-series analytics.
func (t *Tracker) GetDashboardTrends(granularity string, opts DashboardQueryOptions) ([]DashboardTrendPoint, error) {
	periodExpr, err := dashboardPeriodExpr(granularity)
	if err != nil {
		return nil, err
	}

	days := normalizeDashboardDays(opts.Days)
	where, args := buildDashboardFilters(opts, days)
	query := fmt.Sprintf(`
		SELECT
			%s as period,
			COALESCE(model_name, '') as model_name,
			COUNT(*) as commands,
			COALESCE(SUM(original_tokens), 0) as original_tokens,
			COALESCE(SUM(filtered_tokens), 0) as filtered_tokens,
			COALESCE(SUM(saved_tokens), 0) as saved_tokens
		FROM commands
	`, periodExpr)
	if where != "" {
		query += " WHERE " + where
	}
	query += fmt.Sprintf(`
		GROUP BY %s, COALESCE(model_name, '')
		ORDER BY period ASC
	`, periodExpr)

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("dashboard trend query: %w", err)
	}
	defer rows.Close()

	type bucket struct {
		DashboardTrendPoint
	}
	buckets := make(map[string]*bucket)
	order := make([]string, 0)

	for rows.Next() {
		var period, modelName string
		var commands, original, filtered, saved int64
		if err := rows.Scan(&period, &modelName, &commands, &original, &filtered, &saved); err != nil {
			return nil, fmt.Errorf("dashboard trend scan: %w", err)
		}
		if _, ok := buckets[period]; !ok {
			buckets[period] = &bucket{
				DashboardTrendPoint: DashboardTrendPoint{Period: period},
			}
			order = append(order, period)
		}
		entry := buckets[period]
		entry.Commands += commands
		entry.OriginalTokens += original
		entry.FilteredTokens += filtered
		entry.SavedTokens += saved
		entry.EstimatedOriginalCostUSD += estimateCostForModel(modelName, original)
		entry.EstimatedFilteredCostUSD += estimateCostForModel(modelName, filtered)
		entry.EstimatedSavingsUSD += estimateCostForModel(modelName, saved)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard trend iteration: %w", err)
	}

	points := make([]DashboardTrendPoint, 0, len(order))
	for _, period := range order {
		point := buckets[period].DashboardTrendPoint
		if point.OriginalTokens > 0 {
			point.ReductionPct = float64(point.SavedTokens) / float64(point.OriginalTokens) * 100
		}
		points = append(points, point)
	}
	return points, nil
}

// GetDashboardBreakdown returns grouped analytics for a supported dimension.
func (t *Tracker) GetDashboardBreakdown(dimension string, opts DashboardQueryOptions) ([]DashboardBreakdown, error) {
	keyExpr, err := dashboardDimensionExpr(dimension)
	if err != nil {
		return nil, err
	}

	days := normalizeDashboardDays(opts.Days)
	limit := normalizeDashboardLimit(opts.Limit)
	where, args := buildDashboardFilters(opts, days)

	query := fmt.Sprintf(`
		SELECT
			%s as breakdown_key,
			COALESCE(model_name, '') as model_name,
			COUNT(*) as commands,
			COALESCE(SUM(original_tokens), 0) as original_tokens,
			COALESCE(SUM(filtered_tokens), 0) as filtered_tokens,
			COALESCE(SUM(saved_tokens), 0) as saved_tokens
		FROM commands
	`, keyExpr)
	if where != "" {
		query += " WHERE " + where
	}
	query += fmt.Sprintf(`
		GROUP BY %s, COALESCE(model_name, '')
		ORDER BY saved_tokens DESC
	`, keyExpr)

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("dashboard breakdown query: %w", err)
	}
	defer rows.Close()

	items := make(map[string]*DashboardBreakdown)
	order := make([]string, 0)

	for rows.Next() {
		var key, modelName string
		var commands, original, filtered, saved int64
		if err := rows.Scan(&key, &modelName, &commands, &original, &filtered, &saved); err != nil {
			return nil, fmt.Errorf("dashboard breakdown scan: %w", err)
		}
		key = normalizeDashboardKey(key)
		if _, ok := items[key]; !ok {
			items[key] = &DashboardBreakdown{Key: key}
			order = append(order, key)
		}
		item := items[key]
		item.Commands += commands
		item.OriginalTokens += original
		item.FilteredTokens += filtered
		item.SavedTokens += saved
		item.EstimatedOriginalCostUSD += estimateCostForModel(modelName, original)
		item.EstimatedFilteredCostUSD += estimateCostForModel(modelName, filtered)
		item.EstimatedSavingsUSD += estimateCostForModel(modelName, saved)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard breakdown iteration: %w", err)
	}

	out := make([]DashboardBreakdown, 0, len(order))
	for _, key := range order {
		item := *items[key]
		if item.OriginalTokens > 0 {
			item.ReductionPct = float64(item.SavedTokens) / float64(item.OriginalTokens) * 100
		}
		out = append(out, item)
		if len(out) >= limit {
			break
		}
	}

	return out, nil
}

// GetDashboardTopLayers returns layer effectiveness in the current window.
func (t *Tracker) GetDashboardTopLayers(opts DashboardQueryOptions) ([]DashboardLayerSummary, error) {
	days := normalizeDashboardDays(opts.Days)
	where, args := buildDashboardFilters(opts, days)

	query := `
		SELECT
			ls.layer_name,
			COUNT(*) as call_count,
			COALESCE(SUM(ls.tokens_saved), 0) as total_saved,
			COALESCE(AVG(ls.tokens_saved), 0) as avg_saved
		FROM layer_stats ls
		JOIN commands c ON c.id = ls.command_id
	`
	if where != "" {
		query += " WHERE " + strings.ReplaceAll(where, "timestamp", "c.timestamp")
		query = strings.ReplaceAll(query, "project_path", "c.project_path")
		query = strings.ReplaceAll(query, "agent_name", "c.agent_name")
		query = strings.ReplaceAll(query, "provider", "c.provider")
		query = strings.ReplaceAll(query, "model_name", "c.model_name")
		query = strings.ReplaceAll(query, "session_id", "c.session_id")
	}
	query += `
		GROUP BY ls.layer_name
		ORDER BY total_saved DESC
		LIMIT 10
	`

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("dashboard top layers query: %w", err)
	}
	defer rows.Close()

	var layers []DashboardLayerSummary
	for rows.Next() {
		var item DashboardLayerSummary
		if err := rows.Scan(&item.LayerName, &item.CallCount, &item.TotalSaved, &item.AvgSaved); err != nil {
			return nil, fmt.Errorf("dashboard top layers scan: %w", err)
		}
		layers = append(layers, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard top layers iteration: %w", err)
	}

	return layers, nil
}

// GetDashboardLowSavingsCommands returns repeated commands with the weakest savings performance.
func (t *Tracker) GetDashboardLowSavingsCommands(opts DashboardQueryOptions) ([]DashboardBreakdown, error) {
	days := normalizeDashboardDays(opts.Days)
	limit := normalizeDashboardLimit(opts.Limit)
	where, args := buildDashboardFilters(opts, days)

	query := `
		SELECT
			COALESCE(NULLIF(TRIM(command), ''), '(unknown)') as breakdown_key,
			COALESCE(model_name, '') as model_name,
			COUNT(*) as commands,
			COALESCE(SUM(original_tokens), 0) as original_tokens,
			COALESCE(SUM(filtered_tokens), 0) as filtered_tokens,
			COALESCE(SUM(saved_tokens), 0) as saved_tokens
		FROM commands
	`
	if where != "" {
		query += " WHERE " + where
	}
	query += `
		GROUP BY COALESCE(NULLIF(TRIM(command), ''), '(unknown)'), COALESCE(model_name, '')
		HAVING COUNT(*) >= 1 AND COALESCE(SUM(original_tokens), 0) > 0
		ORDER BY
			CAST(COALESCE(SUM(saved_tokens), 0) AS REAL) / NULLIF(SUM(original_tokens), 0) ASC,
			COALESCE(SUM(original_tokens), 0) DESC
		LIMIT ?
	`
	args = append(args, limit)

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("dashboard low savings query: %w", err)
	}
	defer rows.Close()

	var items []DashboardBreakdown
	for rows.Next() {
		var item DashboardBreakdown
		var modelName string
		if err := rows.Scan(
			&item.Key,
			&modelName,
			&item.Commands,
			&item.OriginalTokens,
			&item.FilteredTokens,
			&item.SavedTokens,
		); err != nil {
			return nil, fmt.Errorf("dashboard low savings scan: %w", err)
		}
		if item.OriginalTokens > 0 {
			item.ReductionPct = float64(item.SavedTokens) / float64(item.OriginalTokens) * 100
		}
		item.EstimatedOriginalCostUSD = estimateCostForModel(modelName, item.OriginalTokens)
		item.EstimatedFilteredCostUSD = estimateCostForModel(modelName, item.FilteredTokens)
		item.EstimatedSavingsUSD = estimateCostForModel(modelName, item.SavedTokens)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard low savings iteration: %w", err)
	}

	return items, nil
}

// GetDashboardBudgets returns daily/weekly/monthly token and cost windows for management views.
func (t *Tracker) GetDashboardBudgets(opts DashboardQueryOptions) (DashboardBudgetStatus, error) {
	daily, err := t.getDashboardBudgetWindow("daily", 1, opts.DailyTokenBudget, opts.DailyCostBudgetUSD, opts)
	if err != nil {
		return DashboardBudgetStatus{}, err
	}
	weekly, err := t.getDashboardBudgetWindow("weekly", 7, opts.WeeklyTokenBudget, opts.WeeklyCostBudgetUSD, opts)
	if err != nil {
		return DashboardBudgetStatus{}, err
	}
	monthly, err := t.getDashboardBudgetWindow("monthly", 30, opts.MonthlyTokenBudget, opts.MonthlyCostBudgetUSD, opts)
	if err != nil {
		return DashboardBudgetStatus{}, err
	}

	return DashboardBudgetStatus{
		Daily:   daily,
		Weekly:  weekly,
		Monthly: monthly,
	}, nil
}

// GetDashboardStreaks returns consecutive day streaks for savings and efficiency goals.
func (t *Tracker) GetDashboardStreaks(opts DashboardQueryOptions) (DashboardStreaks, error) {
	goal := opts.ReductionGoalPct
	if goal <= 0 {
		goal = 30
	}

	where, args := buildDashboardFilters(opts, 365)
	query := `
		SELECT
			DATE(timestamp) as period,
			COALESCE(SUM(original_tokens), 0) as original_tokens,
			COALESCE(SUM(saved_tokens), 0) as saved_tokens
		FROM commands
	`
	if where != "" {
		query += " WHERE " + where
	}
	query += `
		GROUP BY DATE(timestamp)
		ORDER BY period DESC
	`

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return DashboardStreaks{}, fmt.Errorf("dashboard streak query: %w", err)
	}
	defer rows.Close()

	type streakDay struct {
		period   string
		original int64
		saved    int64
	}
	var days []streakDay
	streaks := DashboardStreaks{GoalReductionPct: goal}

	for rows.Next() {
		var item streakDay
		if err := rows.Scan(&item.period, &item.original, &item.saved); err != nil {
			return DashboardStreaks{}, fmt.Errorf("dashboard streak scan: %w", err)
		}
		days = append(days, item)

		reduction := 0.0
		if item.original > 0 {
			reduction = float64(item.saved) / float64(item.original) * 100
		}
		if item.saved > streaks.BestDaySavedTokens {
			streaks.BestDay = item.period
			streaks.BestDaySavedTokens = item.saved
			streaks.BestDayReductionPct = reduction
		}
	}
	if err := rows.Err(); err != nil {
		return DashboardStreaks{}, fmt.Errorf("dashboard streak iteration: %w", err)
	}
	if len(days) == 0 {
		return streaks, nil
	}

	prevDate, err := parseDashboardDate(days[0].period)
	if err != nil {
		return DashboardStreaks{}, err
	}
	savingsActive := true
	goalActive := true
	for idx, item := range days {
		currentDate, err := parseDashboardDate(item.period)
		if err != nil {
			return DashboardStreaks{}, err
		}
		if idx > 0 {
			expected := prevDate.AddDate(0, 0, -1)
			if !currentDate.Equal(expected) {
				break
			}
			prevDate = currentDate
		}

		if savingsActive {
			if item.saved > 0 {
				streaks.SavingsDays++
			} else {
				savingsActive = false
			}
		}

		reduction := 0.0
		if item.original > 0 {
			reduction = float64(item.saved) / float64(item.original) * 100
		}
		if goalActive {
			if item.saved > 0 && reduction >= goal {
				streaks.GoalDays++
			} else {
				goalActive = false
			}
		}
		if !savingsActive && !goalActive {
			break
		}
	}

	return streaks, nil
}

// GetDashboardLifecycle returns retention and history-depth signals for dashboard analytics.
func (t *Tracker) GetDashboardLifecycle(opts DashboardQueryOptions) (DashboardLifecycle, error) {
	scopeFilters, scopeArgs := buildDashboardScopeFilters(opts)
	query := `
		SELECT
			COALESCE(MIN(DATE(timestamp)), '') as first_seen_date,
			COUNT(*) as commands_total,
			COUNT(DISTINCT NULLIF(TRIM(project_path), '')) as projects_count,
			COALESCE(SUM(saved_tokens), 0) as total_saved_tokens
		FROM commands
	`
	if len(scopeFilters) > 0 {
		query += " WHERE " + strings.Join(scopeFilters, " AND ")
	}

	var lifecycle DashboardLifecycle
	var totalSavedTokens int64
	if err := t.db.QueryRow(query, scopeArgs...).Scan(
		&lifecycle.FirstSeenDate,
		&lifecycle.CommandsTotal,
		&lifecycle.ProjectsCount,
		&totalSavedTokens,
	); err != nil {
		return DashboardLifecycle{}, fmt.Errorf("dashboard lifecycle query: %w", err)
	}

	if lifecycle.FirstSeenDate != "" {
		firstSeen, err := parseDashboardDate(lifecycle.FirstSeenDate)
		if err != nil {
			return DashboardLifecycle{}, err
		}
		lifecycle.DaysSinceFirstUse = int(time.Since(firstSeen).Hours() / 24)
	}
	if lifecycle.CommandsTotal > 0 {
		lifecycle.AvgSavedTokensPerExec = float64(totalSavedTokens) / float64(lifecycle.CommandsTotal)
	}

	activeWhere, activeArgs := buildDashboardFilters(opts, 30)
	activeQuery := `SELECT COUNT(DISTINCT DATE(timestamp)) FROM commands`
	if activeWhere != "" {
		activeQuery += " WHERE " + activeWhere
	}
	if err := t.db.QueryRow(activeQuery, activeArgs...).Scan(&lifecycle.ActiveDays30d); err != nil {
		return DashboardLifecycle{}, fmt.Errorf("dashboard lifecycle active days query: %w", err)
	}

	return lifecycle, nil
}

type dashboardCostSummary struct {
	original float64
	filtered float64
	saved    float64
}

func (t *Tracker) getDashboardCostSummary(opts DashboardQueryOptions) (dashboardCostSummary, error) {
	days := normalizeDashboardDays(opts.Days)
	where, args := buildDashboardFilters(opts, days)
	query := `
		SELECT
			COALESCE(model_name, '') as model_name,
			COALESCE(SUM(original_tokens), 0) as original_tokens,
			COALESCE(SUM(filtered_tokens), 0) as filtered_tokens,
			COALESCE(SUM(saved_tokens), 0) as saved_tokens
		FROM commands
	`
	if where != "" {
		query += " WHERE " + where
	}
	query += ` GROUP BY COALESCE(model_name, '')`

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return dashboardCostSummary{}, fmt.Errorf("dashboard cost summary query: %w", err)
	}
	defer rows.Close()

	var summary dashboardCostSummary
	for rows.Next() {
		var modelName string
		var original, filtered, saved int64
		if err := rows.Scan(&modelName, &original, &filtered, &saved); err != nil {
			return dashboardCostSummary{}, fmt.Errorf("dashboard cost summary scan: %w", err)
		}
		summary.original += estimateCostForModel(modelName, original)
		summary.filtered += estimateCostForModel(modelName, filtered)
		summary.saved += estimateCostForModel(modelName, saved)
	}
	if err := rows.Err(); err != nil {
		return dashboardCostSummary{}, fmt.Errorf("dashboard cost summary iteration: %w", err)
	}
	return summary, nil
}

func (t *Tracker) getDashboardBudgetWindow(window string, days int, tokenBudget int64, costBudget float64, opts DashboardQueryOptions) (DashboardBudgetWindow, error) {
	windowOpts := opts
	windowOpts.Days = days
	overview, err := t.GetDashboardOverview(windowOpts)
	if err != nil {
		return DashboardBudgetWindow{}, err
	}

	item := DashboardBudgetWindow{
		Window:                   window,
		Days:                     days,
		OriginalTokens:           overview.TotalOriginalTokens,
		FilteredTokens:           overview.TotalFilteredTokens,
		SavedTokens:              overview.TotalSavedTokens,
		EstimatedOriginalCostUSD: overview.EstimatedOriginalCostUSD,
		EstimatedFilteredCostUSD: overview.EstimatedFilteredCostUSD,
		EstimatedSavingsUSD:      overview.EstimatedSavingsUSD,
		TokenBudget:              tokenBudget,
		CostBudgetUSD:            costBudget,
	}
	if tokenBudget > 0 {
		item.TokenRemaining = tokenBudget - item.FilteredTokens
		item.TokenUtilizationPct = (float64(item.FilteredTokens) / float64(tokenBudget)) * 100
		item.OverTokenBudget = item.FilteredTokens > tokenBudget
	}
	if costBudget > 0 {
		item.CostRemainingUSD = costBudget - item.EstimatedFilteredCostUSD
		item.CostUtilizationPct = (item.EstimatedFilteredCostUSD / costBudget) * 100
		item.OverCostBudget = item.EstimatedFilteredCostUSD > costBudget
	}
	return item, nil
}

func buildDashboardFilters(opts DashboardQueryOptions, days int) (string, []any) {
	var filters []string
	var args []any

	filters = append(filters, "timestamp >= datetime('now', ?)")
	args = append(args, fmt.Sprintf("-%d days", days))

	scopeFilters, scopeArgs := buildDashboardScopeFilters(opts)
	filters = append(filters, scopeFilters...)
	args = append(args, scopeArgs...)

	return strings.Join(filters, " AND "), args
}

func buildDashboardScopeFilters(opts DashboardQueryOptions) ([]string, []any) {
	var filters []string
	var args []any

	if projectPath := normalizeProjectPath(opts.ProjectPath); projectPath != "" {
		filters = append(filters, "(project_path GLOB ? OR project_path = ?)")
		pattern := escapeGLOB(projectPath) + "/%"
		args = append(args, pattern, projectPath)
	}
	if value := strings.TrimSpace(opts.AgentName); value != "" {
		filters = append(filters, "agent_name = ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(opts.Provider); value != "" {
		filters = append(filters, "provider = ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(opts.ModelName); value != "" {
		filters = append(filters, "model_name = ?")
		args = append(args, value)
	}
	if value := strings.TrimSpace(opts.SessionID); value != "" {
		filters = append(filters, "session_id = ?")
		args = append(args, value)
	}

	return filters, args
}

func dashboardPeriodExpr(granularity string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(granularity)) {
	case "day", "daily":
		return "DATE(timestamp)", nil
	case "week", "weekly":
		return "strftime('%Y-W%W', timestamp)", nil
	case "month", "monthly":
		return "strftime('%Y-%m', timestamp)", nil
	default:
		return "", fmt.Errorf("unsupported dashboard trend granularity: %s", granularity)
	}
}

func dashboardDimensionExpr(dimension string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(dimension)) {
	case "agent":
		return "COALESCE(NULLIF(TRIM(agent_name), ''), '(unknown)')", nil
	case "provider":
		return "COALESCE(NULLIF(TRIM(provider), ''), '(unknown)')", nil
	case "model":
		return "COALESCE(NULLIF(TRIM(model_name), ''), '(unknown)')", nil
	case "provider_model":
		return "COALESCE(NULLIF(TRIM(provider), ''), '(unknown)') || ' / ' || COALESCE(NULLIF(TRIM(model_name), ''), '(unknown)')", nil
	case "project":
		return "COALESCE(NULLIF(TRIM(project_path), ''), '(unknown)')", nil
	case "command":
		return "COALESCE(NULLIF(TRIM(command), ''), '(unknown)')", nil
	case "session":
		return "COALESCE(NULLIF(TRIM(session_id), ''), '(unknown)')", nil
	case "context_kind":
		return "COALESCE(NULLIF(TRIM(context_kind), ''), '(unknown)')", nil
	default:
		return "", fmt.Errorf("unsupported dashboard breakdown dimension: %s", dimension)
	}
}

func normalizeDashboardDays(days int) int {
	if days <= 0 {
		return 30
	}
	return days
}

func normalizeDashboardLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func normalizeDashboardKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "(unknown)"
	}
	return key
}

func estimateCostForModel(modelName string, tokens int64) float64 {
	if tokens <= 0 {
		return 0
	}
	estimator := NewCostEstimator(modelName)
	return estimator.EstimateCost(int(tokens))
}

func parseDashboardDate(value string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse dashboard date %q: %w", value, err)
	}
	return date, nil
}

func buildDashboardGamification(overview DashboardOverview, streaks DashboardStreaks) DashboardGamification {
	points := overview.TotalSavedTokens/100 + int64(streaks.SavingsDays*25) + int64(streaks.GoalDays*50)
	level := int(points/1000) + 1
	nextLevel := int64(level * 1000)

	var badges []string
	if overview.TotalSavedTokens >= 1_000 {
		badges = append(badges, "first-1k-saved")
	}
	if overview.TotalSavedTokens >= 100_000 {
		badges = append(badges, "hundred-k-saver")
	}
	if streaks.SavingsDays >= 7 {
		badges = append(badges, "seven-day-streak")
	}
	if streaks.GoalDays >= 7 {
		badges = append(badges, "efficiency-week")
	}
	if overview.ReductionPct >= 50 {
		badges = append(badges, "fifty-percent-reducer")
	}

	return DashboardGamification{
		Points:          points,
		Level:           level,
		NextLevelPoints: nextLevel,
		Badges:          badges,
	}
}
