package tracking

import (
	"database/sql"
	"fmt"
)

// PeriodStats represents statistics for a specific time period (day/week/month).
type PeriodStats struct {
	Period       string  `json:"period"`        // Date string (YYYY-MM-DD, YYYY-W##, YYYY-MM)
	Commands     int     `json:"commands"`      // Number of commands
	InputTokens  int     `json:"input_tokens"`  // Total input tokens
	OutputTokens int     `json:"output_tokens"` // Total output tokens (after filtering)
	SavedTokens  int     `json:"saved_tokens"`  // Tokens saved
	SavingsPct   float64 `json:"savings_pct"`   // Percentage saved
	ExecTimeMs   int64   `json:"exec_time_ms"`  // Total execution time
}

// CommandBreakdown represents stats for a specific command.
type CommandBreakdown struct {
	Command      string  `json:"command"`
	Count        int     `json:"count"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	SavedTokens  int     `json:"saved_tokens"`
	SavingsPct   float64 `json:"savings_pct"`
}

// GainSummary represents the full gain output summary.
type GainSummary struct {
	TotalCommands   int                `json:"total_commands"`
	TotalInput      int                `json:"total_input"`
	TotalOutput     int                `json:"total_output"`
	TotalSaved      int                `json:"total_saved"`
	AvgSavingsPct   float64            `json:"avg_savings_pct"`
	TotalExecTimeMs int64              `json:"total_exec_time_ms"`
	AvgExecTimeMs   int64              `json:"avg_exec_time_ms"`
	ByCommand       []CommandBreakdown `json:"by_command"`
	DailyStats      []PeriodStats      `json:"daily_stats,omitempty"`
	WeeklyStats     []PeriodStats      `json:"weekly_stats,omitempty"`
	MonthlyStats    []PeriodStats      `json:"monthly_stats,omitempty"`
	RecentCommands  []CommandRecord    `json:"recent_commands,omitempty"`
}

// GetDailyStats returns daily statistics for the last N days.
func (t *Tracker) GetDailyStats(days int, projectPath string) ([]PeriodStats, error) {
	projectPath = normalizeProjectPath(projectPath)

	query := `
		SELECT 
			DATE(timestamp) as period,
			COUNT(*) as commands,
			COALESCE(SUM(original_tokens), 0) as input_tokens,
			COALESCE(SUM(filtered_tokens), 0) as output_tokens,
			COALESCE(SUM(saved_tokens), 0) as saved_tokens,
			COALESCE(SUM(exec_time_ms), 0) as exec_time_ms
		FROM commands
		WHERE timestamp >= DATE('now', '-%d days')
	`
	args := []interface{}{}

	if projectPath != "" {
		query += ` AND (project_path GLOB ? OR project_path = ?)`
		pattern := escapeGLOB(projectPath) + "/%"
		args = append(args, pattern, projectPath)
	}

	query += `
		GROUP BY DATE(timestamp)
		ORDER BY period DESC
	`

	query = fmt.Sprintf(query, days)
	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPeriodStats(rows)
}

// GetWeeklyStats returns weekly statistics for the last N weeks.
func (t *Tracker) GetWeeklyStats(weeks int, projectPath string) ([]PeriodStats, error) {
	projectPath = normalizeProjectPath(projectPath)

	query := `
		SELECT 
			strftime('%Y-W%W', timestamp) as period,
			COUNT(*) as commands,
			COALESCE(SUM(original_tokens), 0) as input_tokens,
			COALESCE(SUM(filtered_tokens), 0) as output_tokens,
			COALESCE(SUM(saved_tokens), 0) as saved_tokens,
			COALESCE(SUM(exec_time_ms), 0) as exec_time_ms
		FROM commands
		WHERE timestamp >= DATE('now', '-%d days')
	`
	args := []interface{}{}

	if projectPath != "" {
		query += ` AND (project_path GLOB ? OR project_path = ?)`
		pattern := escapeGLOB(projectPath) + "/%"
		args = append(args, pattern, projectPath)
	}

	query += `
		GROUP BY strftime('%Y-W%W', timestamp)
		ORDER BY period DESC
	`

	query = fmt.Sprintf(query, weeks*7)
	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPeriodStats(rows)
}

// GetMonthlyStats returns monthly statistics for the last N months.
func (t *Tracker) GetMonthlyStats(months int, projectPath string) ([]PeriodStats, error) {
	projectPath = normalizeProjectPath(projectPath)

	query := `
		SELECT 
			strftime('%Y-%m', timestamp) as period,
			COUNT(*) as commands,
			COALESCE(SUM(original_tokens), 0) as input_tokens,
			COALESCE(SUM(filtered_tokens), 0) as output_tokens,
			COALESCE(SUM(saved_tokens), 0) as saved_tokens,
			COALESCE(SUM(exec_time_ms), 0) as exec_time_ms
		FROM commands
		WHERE timestamp >= DATE('now', '-%d months')
	`
	args := []interface{}{}

	if projectPath != "" {
		query += ` AND (project_path GLOB ? OR project_path = ?)`
		pattern := escapeGLOB(projectPath) + "/%"
		args = append(args, pattern, projectPath)
	}

	query += `
		GROUP BY strftime('%Y-%m', timestamp)
		ORDER BY period DESC
	`

	query = fmt.Sprintf(query, months)
	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPeriodStats(rows)
}

// GetCommandBreakdown returns statistics grouped by command.
func (t *Tracker) GetCommandBreakdown(limit int, projectPath string) ([]CommandBreakdown, error) {
	projectPath = normalizeProjectPath(projectPath)

	query := `
		SELECT 
			command,
			COUNT(*) as count,
			COALESCE(SUM(original_tokens), 0) as input_tokens,
			COALESCE(SUM(filtered_tokens), 0) as output_tokens,
			COALESCE(SUM(saved_tokens), 0) as saved_tokens
		FROM commands
		WHERE 1=1
	`
	args := []interface{}{}

	if projectPath != "" {
		query += ` AND (project_path GLOB ? OR project_path = ?)`
		pattern := escapeGLOB(projectPath) + "/%"
		args = append(args, pattern, projectPath)
	}

	query += `
		GROUP BY command
		ORDER BY saved_tokens DESC
		LIMIT ?
	`
	args = append(args, limit)

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CommandBreakdown
	for rows.Next() {
		var cb CommandBreakdown
		var inputTokens, outputTokens, savedTokens int
		err := rows.Scan(&cb.Command, &cb.Count, &inputTokens, &outputTokens, &savedTokens)
		if err != nil {
			continue
		}
		cb.InputTokens = inputTokens
		cb.OutputTokens = outputTokens
		cb.SavedTokens = savedTokens
		if inputTokens > 0 {
			cb.SavingsPct = float64(savedTokens) / float64(inputTokens) * 100
		}
		results = append(results, cb)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// GetFullGainSummary returns a comprehensive gain summary with all requested data.
func (t *Tracker) GetFullGainSummary(opts GainSummaryOptions) (*GainSummary, error) {
	summary := &GainSummary{}

	// Get base savings
	savings, err := t.GetSavings(opts.ProjectPath)
	if err != nil {
		return nil, err
	}

	summary.TotalCommands = savings.TotalCommands
	summary.TotalInput = savings.TotalOriginal
	summary.TotalOutput = savings.TotalFiltered
	summary.TotalSaved = savings.TotalSaved
	summary.AvgSavingsPct = savings.ReductionPct

	// Get total execution time
	query := `SELECT COALESCE(SUM(exec_time_ms), 0) FROM commands WHERE 1=1`
	args := []interface{}{}

	if opts.ProjectPath != "" {
		projectPath := normalizeProjectPath(opts.ProjectPath)
		query += ` AND (project_path GLOB ? OR project_path = ?)`
		pattern := escapeGLOB(projectPath) + "/%"
		args = append(args, pattern, projectPath)
	}

	var totalExecTime int64
	err = t.db.QueryRow(query, args...).Scan(&totalExecTime)
	if err != nil {
		totalExecTime = 0
	}
	summary.TotalExecTimeMs = totalExecTime
	if summary.TotalCommands > 0 {
		summary.AvgExecTimeMs = totalExecTime / int64(summary.TotalCommands)
	}

	// Get command breakdown
	summary.ByCommand, err = t.GetCommandBreakdown(10, opts.ProjectPath)
	if err != nil {
		summary.ByCommand = []CommandBreakdown{}
	}

	// Get period stats if requested
	if opts.IncludeDaily {
		summary.DailyStats, _ = t.GetDailyStats(30, opts.ProjectPath)
	}
	if opts.IncludeWeekly {
		summary.WeeklyStats, _ = t.GetWeeklyStats(12, opts.ProjectPath)
	}
	if opts.IncludeMonthly {
		summary.MonthlyStats, _ = t.GetMonthlyStats(12, opts.ProjectPath)
	}
	if opts.IncludeHistory {
		summary.RecentCommands, _ = t.GetRecentCommands(opts.ProjectPath, 20)
	}

	return summary, nil
}

// GainSummaryOptions controls what data is included in the gain summary.
type GainSummaryOptions struct {
	ProjectPath    string
	IncludeDaily   bool
	IncludeWeekly  bool
	IncludeMonthly bool
	IncludeHistory bool
}

// scanPeriodStats scans period stats from SQL rows.
func scanPeriodStats(rows *sql.Rows) ([]PeriodStats, error) {
	var results []PeriodStats
	for rows.Next() {
		var ps PeriodStats
		var inputTokens, outputTokens, savedTokens int
		err := rows.Scan(&ps.Period, &ps.Commands, &inputTokens, &outputTokens, &savedTokens, &ps.ExecTimeMs)
		if err != nil {
			continue
		}
		ps.InputTokens = inputTokens
		ps.OutputTokens = outputTokens
		ps.SavedTokens = savedTokens
		if inputTokens > 0 {
			ps.SavingsPct = float64(savedTokens) / float64(inputTokens) * 100
		}
		results = append(results, ps)
	}
	return results, nil
}
