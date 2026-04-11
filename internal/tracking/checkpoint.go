package tracking

import (
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

func (t *Tracker) RecordCheckpointEvent(event *CheckpointEventRecord) error {
	if event == nil {
		return nil
	}
	_, err := t.db.Exec(
		`INSERT INTO checkpoint_events
		 (command_id, session_id, trigger, reason, fill_pct, quality_score, cooldown_sec)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.CommandID,
		event.SessionID,
		event.Trigger,
		event.Reason,
		event.FillPct,
		event.Quality,
		event.CooldownSec,
	)
	return err
}

// GetCheckpointTelemetry returns trigger event telemetry for the last N days.
func (t *Tracker) GetCheckpointTelemetry(days int) (*CheckpointTelemetry, error) {
	if days <= 0 {
		days = 7
	}
	telemetry := &CheckpointTelemetry{
		Days:      days,
		ByTrigger: map[string]int64{},
	}
	window := fmt.Sprintf("-%d day", days)

	if err := t.db.QueryRow(
		"SELECT COALESCE(COUNT(*),0) FROM checkpoint_events WHERE created_at >= datetime('now', ?)",
		window,
	).Scan(&telemetry.TotalEvents); err != nil {
		return nil, err
	}

	rows, err := t.db.Query(
		`SELECT trigger, COALESCE(COUNT(*),0) FROM checkpoint_events
		 WHERE created_at >= datetime('now', ?)
		 GROUP BY trigger ORDER BY COUNT(*) DESC`,
		window,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var trig string
		var count int64
		if err := rows.Scan(&trig, &count); err != nil {
			return nil, err
		}
		telemetry.ByTrigger[trig] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var last CheckpointEventRecord
	var lastEpoch int64
	err = t.db.QueryRow(
		`SELECT id, command_id, COALESCE(session_id,''), trigger, COALESCE(reason,''),
		        COALESCE(fill_pct,0), COALESCE(quality_score,0), COALESCE(cooldown_sec,0),
				COALESCE(CAST(strftime('%s', created_at) AS INTEGER), 0)
		   FROM checkpoint_events ORDER BY created_at DESC LIMIT 1`,
	).Scan(&last.ID, &last.CommandID, &last.SessionID, &last.Trigger, &last.Reason, &last.FillPct, &last.Quality, &last.CooldownSec, &lastEpoch)
	if err == nil {
		last.CreatedAt = time.Unix(lastEpoch, 0).UTC()
		telemetry.LastEvent = &last
	}
	return telemetry, nil
}

// LayerStatRecord holds per-layer statistics for database recording.
type LayerStatRecord struct {
	LayerName   string
	TokensSaved int
	DurationUs  int64
}

// RecordLayerStats saves per-layer statistics for a command.
// T184: Per-layer savings tracking.
func (t *Tracker) RecordLayerStats(commandID int64, stats []LayerStatRecord) error {
	if len(stats) == 0 {
		return nil
	}

	tx, err := t.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		"INSERT INTO layer_stats (command_id, layer_name, tokens_saved, duration_us) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, s := range stats {
		if _, err := stmt.Exec(commandID, s.LayerName, s.TokensSaved, s.DurationUs); err != nil {
			return fmt.Errorf("failed to insert layer stat: %w", err)
		}
	}

	return tx.Commit()
}

// GetLayerStats returns per-layer statistics for a command.
func (t *Tracker) GetLayerStats(commandID int64) ([]LayerStatRecord, error) {
	rows, err := t.db.Query(
		"SELECT layer_name, tokens_saved, duration_us FROM layer_stats WHERE command_id = ? ORDER BY id",
		commandID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query layer stats: %w", err)
	}
	defer rows.Close()

	var stats []LayerStatRecord
	for rows.Next() {
		var s LayerStatRecord
		if err := rows.Scan(&s.LayerName, &s.TokensSaved, &s.DurationUs); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

// GetTopLayers returns the most effective compression layers.
func (t *Tracker) GetTopLayers(limit int) ([]struct {
	LayerName  string
	TotalSaved int64
	AvgSaved   float64
	CallCount  int64
}, error) {
	query := `
		SELECT layer_name, SUM(tokens_saved) as total_saved,
		       AVG(tokens_saved) as avg_saved, COUNT(*) as call_count
		FROM layer_stats
		GROUP BY layer_name
		ORDER BY total_saved DESC
		LIMIT ?
	`
	rows, err := t.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		LayerName  string
		TotalSaved int64
		AvgSaved   float64
		CallCount  int64
	}
	for rows.Next() {
		var r struct {
			LayerName  string
			TotalSaved int64
			AvgSaved   float64
			CallCount  int64
		}
		if err := rows.Scan(&r.LayerName, &r.TotalSaved, &r.AvgSaved, &r.CallCount); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// cleanupOld removes records older than HistoryRetentionDays.
// This is called automatically after each Record operation.
func (t *Tracker) cleanupOld() {
	// Throttle: at most one cleanup per 60 seconds
	now := time.Now().UnixMilli()
	last := atomic.LoadInt64(&t.lastCleanupMs)
	if now-last < 60000 {
		return
	}
	if !atomic.CompareAndSwapInt64(&t.lastCleanupMs, last, now) {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -HistoryRetentionDays)
	if _, err := t.db.Exec(
		"DELETE FROM commands WHERE timestamp < ?",
		cutoff.Format(time.RFC3339),
	); err != nil {
		slog.Error("tracking cleanup failed", "error", err)
	}
}

// CleanupOld manually triggers cleanup of old records.
// Returns the number of records deleted.
