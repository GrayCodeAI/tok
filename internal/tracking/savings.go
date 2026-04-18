package tracking

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/GrayCodeAI/tokman/internal/contextread"
)

func (t *Tracker) GetSavings(projectPath string) (*SavingsSummary, error) {
	return t.GetSavingsForCommands(projectPath, nil)
}

// GetSavingsForContextReads returns smart-read savings using structured metadata
// when available, with command-pattern fallback for older records.
func (t *Tracker) GetSavingsForContextReads(projectPath, kind, mode string) (*SavingsSummary, error) {
	projectPath = normalizeProjectPath(projectPath)

	var args []any
	var filters []string

	query := `
		SELECT
			COUNT(*) as total_commands,
			COALESCE(SUM(saved_tokens), 0) as total_saved,
			COALESCE(SUM(original_tokens), 0) as total_original,
			COALESCE(SUM(filtered_tokens), 0) as total_filtered
		FROM commands
	`

	if projectPath != "" {
		filters = append(filters, "(project_path GLOB ? OR project_path = ?)")
		pattern := escapeGLOB(projectPath) + "/%"
		args = append(args, pattern, projectPath)
	}

	contextFilter, contextArgs := buildContextReadFilter(kind, mode)
	if contextFilter != "" {
		filters = append(filters, contextFilter)
		args = append(args, contextArgs...)
	}

	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}

	summary := &SavingsSummary{}
	if err := t.db.QueryRow(query, args...).Scan(
		&summary.TotalCommands,
		&summary.TotalSaved,
		&summary.TotalOriginal,
		&summary.TotalFiltered,
	); err != nil {
		return nil, fmt.Errorf("failed to get context-read savings: %w", err)
	}
	if summary.TotalOriginal > 0 {
		summary.ReductionPct = float64(summary.TotalSaved) / float64(summary.TotalOriginal) * 100
	}
	return summary, nil
}

// GetSavingsForCommands returns token savings for commands matching any of the
// provided GLOB patterns. When commandPatterns is empty, it returns all records.
func (t *Tracker) GetSavingsForCommands(projectPath string, commandPatterns []string) (*SavingsSummary, error) {
	projectPath = normalizeProjectPath(projectPath)

	var query string
	var args []any
	var filters []string

	query = `
		SELECT 
			COUNT(*) as total_commands,
			COALESCE(SUM(saved_tokens), 0) as total_saved,
			COALESCE(SUM(original_tokens), 0) as total_original,
			COALESCE(SUM(filtered_tokens), 0) as total_filtered
		FROM commands
	`

	if projectPath != "" {
		filters = append(filters, "(project_path GLOB ? OR project_path = ?)")
		pattern := escapeGLOB(projectPath) + "/%"
		args = append(args, pattern, projectPath)
	}

	if len(commandPatterns) > 0 {
		var commandFilters []string
		for _, pattern := range commandPatterns {
			commandFilters = append(commandFilters, "command GLOB ?")
			args = append(args, pattern)
		}
		filters = append(filters, "("+strings.Join(commandFilters, " OR ")+")")
	}

	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}

	summary := &SavingsSummary{}

	err := t.db.QueryRow(query, args...).Scan(
		&summary.TotalCommands,
		&summary.TotalSaved,
		&summary.TotalOriginal,
		&summary.TotalFiltered,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get savings: %w", err)
	}

	if summary.TotalOriginal > 0 {
		summary.ReductionPct = float64(summary.TotalSaved) / float64(summary.TotalOriginal) * 100
	}

	return summary, nil
}

// CountCommandsSince returns the count of commands executed since the given time.
func (t *Tracker) CountCommandsSince(since time.Time) (int64, error) {
	var count int64
	err := t.db.QueryRow(
		"SELECT COUNT(*) FROM commands WHERE timestamp >= ?",
		since.Format(time.RFC3339),
	).Scan(&count)
	return count, err
}

// ParseFailureRecord represents a single parse failure event.
type ParseFailureRecord struct {
	ID                int64     `json:"id"`
	Timestamp         time.Time `json:"timestamp"`
	RawCommand        string    `json:"raw_command"`
	ErrorMessage      string    `json:"error_message"`
	FallbackSucceeded bool      `json:"fallback_succeeded"`
}

// ParseFailureSummary represents aggregated parse failure analytics.
type ParseFailureSummary struct {
	Total          int64                 `json:"total"`
	RecoveryRate   float64               `json:"recovery_rate"`
	TopCommands    []CommandFailureCount `json:"top_commands"`
	RecentFailures []ParseFailureRecord  `json:"recent_failures"`
}

// CommandFailureCount represents a command and its failure count.
type CommandFailureCount struct {
	Command string `json:"command"`
	Count   int    `json:"count"`
}

// TopCommands returns the top N commands by execution count.
func (t *Tracker) TopCommands(limit int) ([]string, error) {
	rows, err := t.db.Query(
		`SELECT command FROM commands
		 GROUP BY command
		 ORDER BY COUNT(*) DESC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []string
	for rows.Next() {
		var cmd string
		if err := rows.Scan(&cmd); err != nil {
			continue
		}
		commands = append(commands, cmd)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return commands, nil
}

// OverallSavingsPct returns the overall savings percentage across all commands.
func (t *Tracker) OverallSavingsPct() (float64, error) {
	var saved, original int64
	err := t.db.QueryRow(
		"SELECT COALESCE(SUM(saved_tokens), 0), COALESCE(SUM(original_tokens), 0) FROM commands",
	).Scan(&saved, &original)
	if err != nil {
		return 0, err
	}
	if original == 0 {
		return 0, nil
	}
	return float64(saved) / float64(original) * 100, nil
}

// TokensSaved24h returns tokens saved in the last 24 hours.
func (t *Tracker) TokensSaved24h() (int64, error) {
	var saved int64
	err := t.db.QueryRow(
		"SELECT COALESCE(SUM(saved_tokens), 0) FROM commands WHERE timestamp >= ?",
		time.Now().Add(-24*time.Hour).Format(time.RFC3339),
	).Scan(&saved)
	return saved, err
}

// TokensSavedTotal returns total tokens saved across all time.
func (t *Tracker) TokensSavedTotal() (int64, error) {
	var saved int64
	err := t.db.QueryRow(
		"SELECT COALESCE(SUM(saved_tokens), 0) FROM commands",
	).Scan(&saved)
	return saved, err
}

// GetCommandStats returns statistics grouped by command.
// When projectPath is empty, returns all commands without filtering.
func (t *Tracker) GetCommandStats(projectPath string) ([]CommandStats, error) {
	projectPath = normalizeProjectPath(projectPath)

	var query string
	var rows *sql.Rows
	var err error

	if projectPath == "" {
		query = `
			SELECT 
				command,
				COUNT(*) as execution_count,
				COALESCE(SUM(saved_tokens), 0) as total_saved,
				COALESCE(SUM(original_tokens), 0) as total_original
			FROM commands
			GROUP BY command
			ORDER BY total_saved DESC
		`
		rows, err = t.db.Query(query)
	} else {
		query = `
			SELECT 
				command,
				COUNT(*) as execution_count,
				COALESCE(SUM(saved_tokens), 0) as total_saved,
				COALESCE(SUM(original_tokens), 0) as total_original
			FROM commands
			WHERE project_path GLOB ? OR project_path = ?
			GROUP BY command
			ORDER BY total_saved DESC
		`
		pattern := escapeGLOB(projectPath) + "/%"
		rows, err = t.db.Query(query, pattern, projectPath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get command stats: %w", err)
	}
	defer rows.Close()

	var stats []CommandStats
	for rows.Next() {
		var s CommandStats
		if err := rows.Scan(&s.Command, &s.ExecutionCount, &s.TotalSaved, &s.TotalOriginal); err != nil {
			return nil, err
		}
		if s.TotalOriginal > 0 {
			s.ReductionPct = float64(s.TotalSaved) / float64(s.TotalOriginal) * 100
		}
		stats = append(stats, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

// GetRecentCommands returns the most recent command executions.
// When projectPath is empty, returns all recent commands without filtering.
func (t *Tracker) GetRecentCommands(projectPath string, limit int) ([]CommandRecord, error) {
	return t.GetRecentCommandsForPatterns(projectPath, limit, nil)
}

// GetRecentContextReads returns recent smart-read records using structured
// metadata when available, with legacy command fallback for older rows.
func (t *Tracker) GetRecentContextReads(projectPath, kind, mode string, limit int) ([]CommandRecord, error) {
	projectPath = normalizeProjectPath(projectPath)

	var args []any
	var filters []string

	query := `
		SELECT id, command, original_tokens, filtered_tokens, saved_tokens,
		       project_path, session_id, exec_time_ms, timestamp, parse_success,
		       context_kind, context_mode, context_resolved_mode,
		       context_target, context_related_files, context_bundle
		FROM commands
	`

	if projectPath != "" {
		filters = append(filters, "(project_path GLOB ? OR project_path = ?)")
		pattern := escapeGLOB(projectPath) + "/%"
		args = append(args, pattern, projectPath)
	}

	contextFilter, contextArgs := buildContextReadFilter(kind, mode)
	if contextFilter != "" {
		filters = append(filters, contextFilter)
		args = append(args, contextArgs...)
	}

	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent context reads: %w", err)
	}
	defer rows.Close()

	var records []CommandRecord
	for rows.Next() {
		var r CommandRecord
		var parseSuccess int
		if err := rows.Scan(
			&r.ID, &r.Command, &r.OriginalTokens, &r.FilteredTokens, &r.SavedTokens,
			&r.ProjectPath, &r.SessionID, &r.ExecTimeMs, &r.Timestamp, &parseSuccess,
			&r.ContextKind, &r.ContextMode, &r.ContextResolvedMode,
			&r.ContextTarget, &r.ContextRelatedFiles, &r.ContextBundle,
		); err != nil {
			return nil, err
		}
		r.ParseSuccess = parseSuccess == 1
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

// GetRecentCommandsForPatterns returns recent commands optionally filtered by
// command GLOB patterns. When commandPatterns is empty, it returns all commands.
func (t *Tracker) GetRecentCommandsForPatterns(projectPath string, limit int, commandPatterns []string) ([]CommandRecord, error) {
	projectPath = normalizeProjectPath(projectPath)

	var query string
	var args []any
	var filters []string

	query = `
		SELECT id, command, original_tokens, filtered_tokens, saved_tokens,
		       project_path, session_id, exec_time_ms, timestamp, parse_success,
		       context_kind, context_mode, context_resolved_mode,
		       context_target, context_related_files, context_bundle
		FROM commands
	`

	if projectPath != "" {
		filters = append(filters, "(project_path GLOB ? OR project_path = ?)")
		pattern := escapeGLOB(projectPath) + "/%"
		args = append(args, pattern, projectPath)
	}
	if len(commandPatterns) > 0 {
		var commandFilters []string
		for _, pattern := range commandPatterns {
			commandFilters = append(commandFilters, "command GLOB ?")
			args = append(args, pattern)
		}
		filters = append(filters, "("+strings.Join(commandFilters, " OR ")+")")
	}
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent commands: %w", err)
	}
	defer rows.Close()

	var records []CommandRecord
	for rows.Next() {
		var r CommandRecord
		var parseSuccess int
		if err := rows.Scan(
			&r.ID, &r.Command, &r.OriginalTokens, &r.FilteredTokens, &r.SavedTokens,
			&r.ProjectPath, &r.SessionID, &r.ExecTimeMs, &r.Timestamp, &parseSuccess,
			&r.ContextKind, &r.ContextMode, &r.ContextResolvedMode,
			&r.ContextTarget, &r.ContextRelatedFiles, &r.ContextBundle,
		); err != nil {
			return nil, err
		}
		r.ParseSuccess = parseSuccess == 1
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func buildContextReadFilter(kind, mode string) (string, []any) {
	var filters []string
	var args []any

	if strings.TrimSpace(kind) != "" {
		legacyPatterns := contextread.TrackedCommandPatternsForKind(kind)
		var legacy []string
		for _, pattern := range legacyPatterns {
			legacy = append(legacy, "command GLOB ?")
			args = append(args, pattern)
		}
		if len(legacy) > 0 {
			filters = append(filters, "(context_kind = ? OR (COALESCE(context_kind, '') = '' AND ("+strings.Join(legacy, " OR ")+")))")
			args = append([]any{strings.ToLower(kind)}, args...)
		} else {
			filters = append(filters, "context_kind = ?")
			args = append(args, strings.ToLower(kind))
		}
	} else {
		var legacy []string
		for _, pattern := range contextread.TrackedCommandPatterns() {
			legacy = append(legacy, "command GLOB ?")
			args = append(args, pattern)
		}
		filters = append(filters, "(COALESCE(context_kind, '') != '' OR ("+strings.Join(legacy, " OR ")+"))")
	}

	if strings.TrimSpace(mode) != "" {
		filters = append(filters, "(context_mode = ? OR context_resolved_mode = ?)")
		mode = strings.ToLower(mode)
		args = append(args, mode, mode)
	}

	return strings.Join(filters, " AND "), args
}

// GetDailySavings returns token savings grouped by day.
func (t *Tracker) GetDailySavings(projectPath string, days int) ([]struct {
	Date     string
	Saved    int
	Original int
	Commands int
}, error) {
	projectPath = normalizeProjectPath(projectPath)

	query := `
		SELECT 
			DATE(timestamp) as date,
			COALESCE(SUM(saved_tokens), 0) as saved,
			COALESCE(SUM(original_tokens), 0) as original,
			COUNT(*) as commands
		FROM commands
	`
	args := []any{}
	filters := []string{"timestamp >= DATE('now', ?)"}
	daysStr := fmt.Sprintf("-%d days", days)
	args = append(args, daysStr)

	if projectPath != "" {
		filters = append(filters, "(project_path GLOB ? OR project_path = ?)")
		pattern := escapeGLOB(projectPath) + "/%"
		args = append(args, pattern, projectPath)
	}

	query += " WHERE " + strings.Join(filters, " AND ")
	query += `
		GROUP BY DATE(timestamp)
		ORDER BY date DESC
	`

	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily savings: %w", err)
	}
	defer rows.Close()

	var results []struct {
		Date     string
		Saved    int
		Original int
		Commands int
	}
	for rows.Next() {
		var r struct {
			Date     string
			Saved    int
			Original int
			Commands int
		}
		if err := rows.Scan(&r.Date, &r.Saved, &r.Original, &r.Commands); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func normalizeProjectPath(projectPath string) string {
	projectPath = strings.TrimSpace(projectPath)
	if projectPath == "" {
		return ""
	}

	if !filepath.IsAbs(projectPath) {
		if absPath, err := filepath.Abs(projectPath); err == nil {
			projectPath = absPath
		}
	}

	if canonicalPath, err := filepath.EvalSymlinks(projectPath); err == nil && canonicalPath != "" {
		return canonicalPath
	}

	return filepath.Clean(projectPath)
}

// escapeGLOB escapes SQLite GLOB metacharacters (*, ?, [, ])
// so user-controlled strings are treated as literals, not wildcards.
// Uses a single-pass approach to avoid corrupting already-escaped sequences.
func escapeGLOB(pattern string) string {
	var b strings.Builder
	b.Grow(len(pattern))
	for _, r := range pattern {
		switch r {
		case '[':
			b.WriteString("[[]")
		case ']':
			b.WriteString("[]]")
		case '*':
			b.WriteString("[*]")
		case '?':
			b.WriteString("[?]")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// RecordParseFailure records a parse failure event.
