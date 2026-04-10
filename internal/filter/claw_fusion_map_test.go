package filter

import "testing"

func TestClawFusionStageCoverage_Has14Stages(t *testing.T) {
	m := ClawFusionStageCoverage()
	if len(m) != 14 {
		t.Fatalf("expected 14 stages, got %d", len(m))
	}
	for _, s := range m {
		if s.Stage == "" {
			t.Fatalf("empty stage name")
		}
		if len(s.LayerIDs) == 0 {
			t.Fatalf("stage %s has no mapped layers", s.Stage)
		}
	}
}
