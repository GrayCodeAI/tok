package rightsizing

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestComplexityScorer(t *testing.T) {
	s := NewComplexityScorer()

	factors := ComplexityFactors{
		CodeDepth:         0.5,
		DomainSpecificity: 0.3,
		ReasoningDepth:    0.4,
		ContextSize:       0.5,
		ToolUsage:         0.2,
		Ambiguity:         0.3,
	}

	level := s.Score(factors)
	if level < 0 || level > 9 {
		t.Errorf("Invalid complexity level: %d", level)
	}
}

func TestScoreFromText(t *testing.T) {
	s := NewComplexityScorer()

	simple := "fix this"
	complex := "Analyze the performance bottleneck in the distributed caching layer. Consider the impact of cache invalidation on consistency. What are the trade-offs between strong consistency and eventual consistency in this context? How would you implement a circuit breaker pattern?"

	simpleLevel := s.ScoreFromText(simple)
	complexLevel := s.ScoreFromText(complex)

	if complexLevel <= simpleLevel {
		t.Errorf("Complex text should score higher: simple=%d, complex=%d", simpleLevel, complexLevel)
	}
}

func TestModelRecommendation(t *testing.T) {
	e := NewModelRecommendationEngine()

	rec := e.Recommend(ComplexitySimple, "claude-3-opus")
	if rec.RecommendedModel == "" {
		t.Error("Expected recommendation")
	}
	if rec.EstimatedSavings <= 0 {
		t.Error("Expected positive savings")
	}

	rec = e.Recommend(ComplexityFrontier, "gpt-4o-mini")
	if rec.RecommendedModel == rec.CurrentModel {
		t.Error("Should recommend upgrade for frontier complexity")
	}
}

func TestRightSizingStore(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skip("SQLite not available")
	}
	defer db.Close()

	store := NewRightSizingStore(db)
	if err := store.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}

	rec := &RightSizingRecord{
		Command:          "git status",
		CurrentModel:     "claude-3-opus",
		RecommendedModel: "gpt-4o-mini",
		ComplexityScore:  2,
		EstimatedSavings: 50.0,
	}

	if err := store.Record(rec); err != nil {
		t.Fatalf("Record error: %v", err)
	}

	records, err := store.GetRecent(10)
	if err != nil {
		t.Fatalf("GetRecent error: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}

	if err := store.AcceptRecommendation(records[0].ID); err != nil {
		t.Fatalf("AcceptRecommendation error: %v", err)
	}

	accuracy, err := store.GetAccuracy()
	if err != nil {
		t.Fatalf("GetAccuracy error: %v", err)
	}
	if accuracy != 100.0 {
		t.Errorf("Expected 100%% accuracy, got %.2f", accuracy)
	}
}
