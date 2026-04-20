package session

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	out "github.com/lakshmanpatel/tok/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/config"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var (
	sessionAdoptionDays int
	sessionAdoptionJSON bool
)

func init() {
	adoptionCmd := &cobra.Command{
		Use:   "adoption",
		Short: "Show tok adoption across recent sessions",
		Long: `Analyze command execution history to show tok adoption statistics
across recent sessions.

This command groups commands by session and calculates:
- Percentage of commands routed through tok
- Token savings by session
- Adoption trends over time

Examples:
  tok session adoption              # Show last 30 days
  tok session adoption --since 7    # Last 7 days only
  tok session adoption --format json # JSON output`,
		RunE: runSessionAdoption,
	}
	adoptionCmd.Flags().IntVarP(&sessionAdoptionDays, "since", "s", 30, "Limit to last N days")
	adoptionCmd.Flags().BoolVarP(&sessionAdoptionJSON, "format", "f", false, "Output as JSON")

	sessionCmd.AddCommand(adoptionCmd)
}

type SessionAdoptionRow struct {
	SessionID   string    `json:"session_id"`
	StartTime   time.Time `json:"start_time"`
	TotalCmds   int       `json:"total_commands"`
	TokCmds     int       `json:"tok_commands"`
	TokensSaved int       `json:"tokens_saved"`
	AdoptionPct float64   `json:"adoption_pct"`
}

type SessionAdoptionResult struct {
	Sessions         []SessionAdoptionRow `json:"sessions"`
	AvgAdoption      float64              `json:"average_adoption_pct"`
	TotalCmds        int                  `json:"total_commands"`
	TotalTokCmds     int                  `json:"total_tok_commands"`
	TotalTokensSaved int                  `json:"total_tokens_saved"`
	PeriodDays       int                  `json:"period_days"`
}

func runSessionAdoption(cmd *cobra.Command, args []string) error {
	dbPath := config.DatabasePath()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		out.Global().Println("No tracking data found.")
		out.Global().Println("Run some commands through tok to start tracking!")
		return nil
	}

	tracker, err := tracking.NewTracker(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize tracker: %w", err)
	}
	defer tracker.Close()

	sessions, err := querySessionAdoption(tracker, sessionAdoptionDays)
	if err != nil {
		return fmt.Errorf("failed to query session data: %w", err)
	}

	if len(sessions) == 0 {
		out.Global().Println("No session data found for the specified period.")
		return nil
	}

	result := buildSessionAdoptionResult(sessions, sessionAdoptionDays)

	if sessionAdoptionJSON {
		return outputSessionAdoptionJSON(result)
	}

	return printSessionAdoptionText(result)
}

func querySessionAdoption(t *tracking.Tracker, days int) ([]SessionAdoptionRow, error) {
	query := `
		SELECT
			COALESCE(session_id, 'standalone') as session_id,
			MIN(timestamp) as start_time,
			COUNT(*) as total_cmds,
			COALESCE(SUM(CASE WHEN command LIKE 'tok %' THEN 1 ELSE 0 END), 0) as tok_cmds,
			COALESCE(SUM(saved_tokens), 0) as tokens_saved
		FROM commands
		WHERE timestamp >= datetime('now', ?)
		GROUP BY COALESCE(session_id, 'standalone')
		ORDER BY start_time DESC
	`

	rows, err := t.Query(query, fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionAdoptionRow
	for rows.Next() {
		var s SessionAdoptionRow
		var startTimeStr string
		if err := rows.Scan(&s.SessionID, &startTimeStr, &s.TotalCmds, &s.TokCmds, &s.TokensSaved); err != nil {
			continue
		}

		parsed, err := time.Parse("2006-01-02 15:04:05", startTimeStr)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339, startTimeStr)
			if err != nil {
				continue
			}
		}
		s.StartTime = parsed

		if s.TotalCmds > 0 {
			s.AdoptionPct = float64(s.TokCmds) / float64(s.TotalCmds) * 100
		}

		sessions = append(sessions, s)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.After(sessions[j].StartTime)
	})

	return sessions, nil
}

func buildSessionAdoptionResult(sessions []SessionAdoptionRow, days int) *SessionAdoptionResult {
	result := &SessionAdoptionResult{
		Sessions:   sessions,
		PeriodDays: days,
	}

	for _, s := range sessions {
		result.TotalCmds += s.TotalCmds
		result.TotalTokCmds += s.TokCmds
		result.TotalTokensSaved += s.TokensSaved
	}

	if result.TotalCmds > 0 {
		result.AvgAdoption = float64(result.TotalTokCmds) / float64(result.TotalCmds) * 100
	}

	return result
}

func printSessionAdoptionText(result *SessionAdoptionResult) error {
	out.Global().Println()
	out.Global().Println(color.New(color.Bold).Sprint(fmt.Sprintf("Session Adoption (Last %d Days)", result.PeriodDays)))
	out.Global().Println("┌" + strings.Repeat("─", 20) + "┬" + strings.Repeat("─", 10) + "┬" + strings.Repeat("─", 10) + "┬" + strings.Repeat("─", 10) + "┐")
	out.Global().Printf("│ %-18s │ %8s │ %8s │ %8s │\n", "Session", "Commands", "tok Used", "Adoption")
	out.Global().Println("├" + strings.Repeat("─", 20) + "┼" + strings.Repeat("─", 10) + "┼" + strings.Repeat("─", 10) + "┼" + strings.Repeat("─", 10) + "┤")

	for _, s := range result.Sessions {
		sessionLabel := s.StartTime.Format("2006-01-02 15:04")
		if len(sessionLabel) > 18 {
			sessionLabel = sessionLabel[:18]
		}

		out.Global().Printf("│ %-18s │ %8d │ %8d │ %7.1f%% │\n",
			sessionLabel,
			s.TotalCmds,
			s.TokCmds,
			s.AdoptionPct,
		)
	}
	out.Global().Println("└" + strings.Repeat("─", 20) + "┴" + strings.Repeat("─", 10) + "┴" + strings.Repeat("─", 10) + "┴" + strings.Repeat("─", 10) + "┘")

	out.Global().Printf("Average adoption: %.1f%%\n", result.AvgAdoption)
	out.Global().Println()

	return nil
}

func outputSessionAdoptionJSON(result *SessionAdoptionResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
