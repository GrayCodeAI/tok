package tracking

import (
	"fmt"
	"log/slog"
	"time"
)

func (t *Tracker) RecordParseFailure(rawCommand string, errorMessage string, fallbackSucceeded bool) error {
	_, err := t.db.Exec(
		`INSERT INTO parse_failures (timestamp, raw_command, error_message, fallback_succeeded)
		 VALUES (?, ?, ?, ?)`,
		time.Now().Format(time.RFC3339),
		rawCommand,
		errorMessage,
		fallbackSucceeded,
	)
	if err != nil {
		return fmt.Errorf("failed to record parse failure: %w", err)
	}

	// Cleanup old records (throttled)
	if !t.closed.Load() {
		select {
		case t.cleanupCh <- struct{}{}:
		default:
		}
	}

	return nil
}

// GetParseFailureSummary returns aggregated parse failure analytics.
func (t *Tracker) GetParseFailureSummary() (*ParseFailureSummary, error) {
	summary := &ParseFailureSummary{}

	// Get total count
	err := t.db.QueryRow("SELECT COUNT(*) FROM parse_failures").Scan(&summary.Total)
	if err != nil {
		return nil, fmt.Errorf("failed to get parse failure count: %w", err)
	}

	if summary.Total == 0 {
		return summary, nil
	}

	// Get recovery rate
	var succeeded int64
	err = t.db.QueryRow(
		"SELECT COUNT(*) FROM parse_failures WHERE fallback_succeeded = 1",
	).Scan(&succeeded)
	if err == nil {
		summary.RecoveryRate = float64(succeeded) / float64(summary.Total) * 100
	}

	// Get top 10 failing commands
	topRows, err := t.db.Query(
		`SELECT raw_command, COUNT(*) as cnt
		 FROM parse_failures
		 GROUP BY raw_command
		 ORDER BY cnt DESC
		 LIMIT 10`,
	)
	if err != nil {
		return summary, nil // return partial results on query failure
	}
	defer topRows.Close()
	for topRows.Next() {
		var cfc CommandFailureCount
		if err := topRows.Scan(&cfc.Command, &cfc.Count); err == nil {
			summary.TopCommands = append(summary.TopCommands, cfc)
		}
	}

	// Get recent 10 failures
	recentRows, err := t.db.Query(
		`SELECT id, timestamp, raw_command, error_message, fallback_succeeded
		 FROM parse_failures
		 ORDER BY timestamp DESC
		 LIMIT 10`,
	)
	if err != nil {
		return summary, nil // return partial results on query failure
	}
	defer recentRows.Close()
	for recentRows.Next() {
		var pfr ParseFailureRecord
		var ts string
		var fb int
		if err := recentRows.Scan(&pfr.ID, &ts, &pfr.RawCommand, &pfr.ErrorMessage, &fb); err == nil {
			parsed, parseErr := time.Parse(time.RFC3339, ts)
			if parseErr != nil {
				slog.Warn("failed to parse timestamp", "timestamp", ts, "error", parseErr)
			}
			pfr.Timestamp = parsed
			pfr.FallbackSucceeded = fb == 1
			summary.RecentFailures = append(summary.RecentFailures, pfr)
		}
	}
	if err := recentRows.Err(); err != nil {
		slog.Error("error iterating parse failure rows", "error", err)
	}

	return summary, nil
}
