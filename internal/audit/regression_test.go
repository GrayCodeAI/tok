package audit

import (
	"path/filepath"
	"testing"

	"github.com/lakshmanpatel/tok/internal/tracking"
)

func TestAuditRegressionFixture(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "audit-fixture.db")
	tr, err := tracking.NewTracker(dbPath)
	if err != nil {
		t.Fatalf("new tracker: %v", err)
	}
	defer tr.Close()

	project := t.TempDir()
	records := []*tracking.CommandRecord{
		{
			Command:        "ls -la",
			OriginalTokens: 25000,
			FilteredTokens: 24500,
			SavedTokens:    500,
			ProjectPath:    project,
			SessionID:      "s1",
			ParseSuccess:   true,
			ModelName:      "gpt-5.4",
		},
		{
			Command:        "ls -la",
			OriginalTokens: 21000,
			FilteredTokens: 20600,
			SavedTokens:    400,
			ProjectPath:    project,
			SessionID:      "s2",
			ParseSuccess:   true,
			ModelName:      "gpt-5.4",
		},
		{
			Command:        "git status",
			OriginalTokens: 18000,
			FilteredTokens: 17700,
			SavedTokens:    300,
			ProjectPath:    project,
			SessionID:      "s3",
			ParseSuccess:   true,
			ModelName:      "gpt-5.4",
		},
		{
			Command:        "npm test",
			OriginalTokens: 8000,
			FilteredTokens: 7600,
			SavedTokens:    400,
			ProjectPath:    project,
			SessionID:      "s4",
			ParseSuccess:   false,
			ModelName:      "sonnet",
		},
		{
			Command:        "go test ./...",
			OriginalTokens: 12000,
			FilteredTokens: 6000,
			SavedTokens:    6000,
			ProjectPath:    project,
			SessionID:      "s5",
			ParseSuccess:   true,
			ModelName:      "haiku",
		},
	}

	for _, rec := range records {
		if err := tr.Record(rec); err != nil {
			t.Fatalf("record fixture command: %v", err)
		}
		if rec.ID > 0 {
			_ = tr.RecordLayerStats(rec.ID, []tracking.LayerStatRecord{
				{LayerName: "11_compaction", TokensSaved: rec.SavedTokens / 2, DurationUs: 1000},
				{LayerName: "10_budget", TokensSaved: rec.SavedTokens / 3, DurationUs: 500},
			})
		}
	}

	report, err := GenerateWithOptions(tr, 30, GenerateOptions{})
	if err != nil {
		t.Fatalf("generate report: %v", err)
	}

	if report.Summary.CommandCount != int64(len(records)) {
		t.Fatalf("summary command count mismatch: got=%d want=%d", report.Summary.CommandCount, len(records))
	}
	if len(report.WasteFindings) == 0 {
		t.Fatal("expected waste findings in regression fixture")
	}
	if len(report.TopLayers) == 0 {
		t.Fatal("expected top layer stats in report")
	}
	if len(report.CostlyPrompts) == 0 {
		t.Fatal("expected costly prompts in report")
	}
	if len(report.TurnAnalytics) != len(records) {
		t.Fatalf("turn analytics mismatch: got=%d want=%d", len(report.TurnAnalytics), len(records))
	}
	if report.Quality.Score <= 0 || report.Quality.Score > 100 {
		t.Fatalf("quality score out of bounds: %.2f", report.Quality.Score)
	}
	if report.BudgetController.RecommendedMode == "" {
		t.Fatal("expected budget controller recommendation")
	}
	if report.AnchorRetention.Grade == "" {
		t.Fatal("expected anchor retention grade")
	}
	if len(report.IntentProfiles) == 0 {
		t.Fatal("expected intent profiles in report")
	}

	ids := map[string]bool{}
	for _, f := range report.WasteFindings {
		ids[f.ID] = true
	}
	if !ids["empty_runs"] {
		t.Fatal("expected empty_runs finding in regression fixture")
	}
	if !ids["model_routing"] {
		t.Fatal("expected model_routing finding in regression fixture")
	}
}
