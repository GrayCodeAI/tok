package tracking

import (
	"fmt"
	"log/slog"
	"time"
)

func (t *Tracker) CleanupOld() (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -HistoryRetentionDays)
	result, err := t.db.Exec(
		"DELETE FROM commands WHERE timestamp < ?",
		cutoff.Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old records: %w", err)
	}
	return result.RowsAffected()
}

// CleanupWithRetention removes records older than specified days.
// T183: Configurable data retention policy.
func (t *Tracker) CleanupWithRetention(days int) (int64, error) {
	if days <= 0 {
		days = HistoryRetentionDays
	}
	cutoff := time.Now().AddDate(0, 0, -days)

	tx, err := t.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete layer stats for old commands
	if _, err := tx.Exec(
		"DELETE FROM layer_stats WHERE command_id IN (SELECT id FROM commands WHERE timestamp < ?)",
		cutoff.Format(time.RFC3339),
	); err != nil {
		slog.Warn("layer_stats cleanup failed", "error", err)
	}

	// Delete old commands
	result, err := tx.Exec(
		"DELETE FROM commands WHERE timestamp < ?",
		cutoff.Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup: %w", err)
	}

	// Delete old parse failures
	if _, err := tx.Exec(
		"DELETE FROM parse_failures WHERE timestamp < ?",
		cutoff.Format(time.RFC3339),
	); err != nil {
		slog.Warn("parse_failures cleanup failed", "error", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit cleanup: %w", err)
	}

	return result.RowsAffected()
}

// DatabaseSize returns the size of the tracking database in bytes.
func (t *Tracker) DatabaseSize() (int64, error) {
	var pageCount, pageSize int64
	err := t.db.QueryRow("PRAGMA page_count").Scan(&pageCount)
	if err != nil {
		return 0, err
	}
	err = t.db.QueryRow("PRAGMA page_size").Scan(&pageSize)
	if err != nil {
		return 0, err
	}
	return pageCount * pageSize, nil
}

// Vacuum reclaims unused space in the database.
func (t *Tracker) Vacuum() error {
	_, err := t.db.Exec("VACUUM")
	return err
}

// GetSavings returns the total token savings for a project path.
// Uses GLOB matching for case-sensitive path comparison.
// When projectPath is empty, returns all records without filtering.
