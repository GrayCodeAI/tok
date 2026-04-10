package audit

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/GrayCodeAI/tokman/internal/tracking"
)

func TestSaveLoadAndCompareSnapshots(t *testing.T) {
	dir := t.TempDir()
	base := &Report{
		GeneratedAt: time.Now().UTC(),
		Days:        7,
		Summary: Summary{
			Saved:        1000,
			ReductionPct: 42.0,
		},
		Quality: QualityReport{Score: 70},
	}
	candidate := &Report{
		GeneratedAt: time.Now().UTC(),
		Days:        7,
		Summary: Summary{
			Saved:         1800,
			ReductionPct:  49.0,
			ParseFailures: 1,
		},
		Quality: QualityReport{Score: 75},
	}

	basePath, err := SaveSnapshot(dir, "base", base)
	if err != nil {
		t.Fatalf("save base: %v", err)
	}
	candidatePath, err := SaveSnapshot(dir, "candidate", candidate)
	if err != nil {
		t.Fatalf("save candidate: %v", err)
	}

	baseSnap, err := LoadSnapshot(basePath)
	if err != nil {
		t.Fatalf("load base: %v", err)
	}
	candidateSnap, err := LoadSnapshot(candidatePath)
	if err != nil {
		t.Fatalf("load candidate: %v", err)
	}

	comp := Compare(baseSnap, candidateSnap)
	if comp.DeltaSavedTokens <= 0 {
		t.Fatalf("expected positive saved delta, got %d", comp.DeltaSavedTokens)
	}
	if comp.Verdict == "" {
		t.Fatal("expected non-empty verdict")
	}
}

func TestGenerateWithTrackerData(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tokman.db")
	tr, err := tracking.NewTracker(dbPath)
	if err != nil {
		t.Fatalf("new tracker: %v", err)
	}
	defer tr.Close()

	record := &tracking.CommandRecord{
		Command:             "git status",
		OriginalTokens:      5000,
		FilteredTokens:      4900,
		SavedTokens:         100,
		ProjectPath:         t.TempDir(),
		ExecTimeMs:          400,
		ParseSuccess:        true,
		ContextKind:         "read",
		ContextRelatedFiles: 3,
	}
	if err := tr.Record(record); err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := tr.RecordLayerStats(record.ID, []tracking.LayerStatRecord{
		{LayerName: "11_compaction", TokensSaved: 50, DurationUs: 1000},
	}); err != nil {
		t.Fatalf("record layer stats: %v", err)
	}

	report, err := Generate(tr, 30)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if report.Summary.CommandCount == 0 {
		t.Fatal("expected non-zero command count")
	}
	if len(report.TopLayers) == 0 {
		t.Fatal("expected top layer entries")
	}
}
