package tracking

import (
	"path/filepath"
	"testing"
)

func TestCheckpointTelemetryAutoRecord(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tok.db")
	tr, err := NewTracker(dbPath)
	if err != nil {
		t.Fatalf("new tracker: %v", err)
	}
	defer tr.Close()

	rec := &CommandRecord{
		Command:        "git commit -m test",
		OriginalTokens: 60000,
		FilteredTokens: 30000,
		SavedTokens:    30000,
		ProjectPath:    t.TempDir(),
		SessionID:      "session-1",
		ParseSuccess:   true,
	}
	if err := tr.Record(rec); err != nil {
		t.Fatalf("record: %v", err)
	}

	tel, err := tr.GetCheckpointTelemetry(7)
	if err != nil {
		t.Fatalf("telemetry: %v", err)
	}
	if tel.TotalEvents == 0 {
		t.Fatal("expected checkpoint events to be recorded")
	}
	if tel.ByTrigger["progressive-20"] == 0 {
		t.Fatal("expected progressive-20 trigger")
	}
}
