package tracking

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestNewTrackerMigratesLegacySchemaAndCreatesCheckpointEvents(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}

	// Simulate older schema: commands only, no checkpoint_events table.
	// user_version is left at 0 to mimic a pre-versioning database.
	if _, err := db.Exec(CreateCommandsTable); err != nil {
		t.Fatalf("create commands table: %v", err)
	}
	db.Close()

	tr, err := NewTracker(dbPath)
	if err != nil {
		t.Fatalf("new tracker migration: %v", err)
	}
	defer tr.Close()

	var count int
	err = tr.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='checkpoint_events'").Scan(&count)
	if err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected checkpoint_events table to exist after migration, count=%d", count)
	}
}

func TestNewTrackerLegacyDBCanRecordAndReadCheckpointTelemetry(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy2.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	if _, err := db.Exec(CreateCommandsTable); err != nil {
		t.Fatalf("create commands table: %v", err)
	}
	_ = db.Close()

	tr, err := NewTracker(dbPath)
	if err != nil {
		t.Fatalf("new tracker migration: %v", err)
	}
	defer tr.Close()

	rec := &CommandRecord{
		Command:        "git commit -m test",
		OriginalTokens: 60000,
		FilteredTokens: 30000,
		SavedTokens:    30000,
		ProjectPath:    t.TempDir(),
		SessionID:      "legacy-session",
		ParseSuccess:   true,
	}
	if err := tr.Record(rec); err != nil {
		t.Fatalf("record command: %v", err)
	}

	tel, err := tr.GetCheckpointTelemetry(7)
	if err != nil {
		t.Fatalf("get checkpoint telemetry: %v", err)
	}
	if tel.TotalEvents == 0 {
		t.Fatal("expected checkpoint telemetry events after record")
	}
}
