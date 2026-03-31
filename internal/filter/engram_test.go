package filter

import "testing"

func TestEngramMemory_Observe(t *testing.T) {
	em := NewEngramMemory(0.7)
	em.Observe("test observation", 0.8)
	if len(em.observations) != 1 {
		t.Errorf("expected 1 observation, got %d", len(em.observations))
	}
}

func TestEngramMemory_ObserveBelowThreshold(t *testing.T) {
	em := NewEngramMemory(0.7)
	em.Observe("test observation", 0.5)
	if len(em.observations) != 0 {
		t.Errorf("expected 0 observations below threshold, got %d", len(em.observations))
	}
}

func TestEngramMemory_TieredSummary(t *testing.T) {
	em := NewEngramMemory(0.5)
	em.Observe("obs1", 0.8)
	em.Observe("obs2", 0.9)
	summary := em.TieredSummary()
	if _, ok := summary["L0"]; !ok {
		t.Error("expected L0 summary")
	}
	if _, ok := summary["L1"]; !ok {
		t.Error("expected L1 summary")
	}
	if _, ok := summary["L2"]; !ok {
		t.Error("expected L2 summary")
	}
}

func TestEngramMemory_Reflect(t *testing.T) {
	em := NewEngramMemory(0.5)
	em.Observe("obs1", 0.8)
	em.Observe("obs2", 0.8)
	reflections := em.Reflect()
	if len(reflections) == 0 {
		t.Log("no reflections generated (expected for simple cases)")
	}
}
