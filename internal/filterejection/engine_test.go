package filterejection

import "testing"

func TestEjectionEngine(t *testing.T) {
	engine := NewEjectionEngine()

	engine.Register("filter1")
	engine.Register("filter2")
	engine.Register("filter3")

	if engine.IsEjected("filter1") {
		t.Error("filter1 should not be ejected")
	}

	engine.Eject("filter2", EjectionSecurity, "detected PII exposure")
	if !engine.IsEjected("filter2") {
		t.Error("filter2 should be ejected")
	}

	engine.Restore("filter2")
	if engine.IsEjected("filter2") {
		t.Error("filter2 should be restored")
	}
}

func TestEjectionEngineStats(t *testing.T) {
	engine := NewEjectionEngine()
	engine.Register("a")
	engine.Register("b")
	engine.Eject("b", EjectionPerformance, "too slow")

	stats := engine.Stats()
	if stats["registered"] != 2 {
		t.Errorf("Expected 2 registered, got %d", stats["registered"])
	}
	if stats["ejected"] != 1 {
		t.Errorf("Expected 1 ejected, got %d", stats["ejected"])
	}
}

func TestFilterActive(t *testing.T) {
	engine := NewEjectionEngine()
	engine.Register("a")
	engine.Register("b")
	engine.Register("c")
	engine.Eject("b", EjectionManual, "")

	active := engine.FilterActive([]string{"a", "b", "c"})
	if len(active) != 2 {
		t.Errorf("Expected 2 active, got %d", len(active))
	}
}
