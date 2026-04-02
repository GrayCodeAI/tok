package canary

import (
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil || m.deployments == nil {
		t.Fatal("expected manager")
	}
}

func TestCreateDeployment(t *testing.T) {
	m := NewManager()
	d, err := m.CreateDeployment(DeploymentConfig{
		Name: "Test", Service: "api", Strategy: StrategyStepped,
	})
	if err != nil {
		t.Fatal(err)
	}
	if d.Name != "Test" || d.Status != StatusPending || len(d.Phases) == 0 {
		t.Error("deployment not created correctly")
	}
}

func TestGetDeployment(t *testing.T) {
	m := NewManager()
	d, _ := m.CreateDeployment(DeploymentConfig{Name: "Test", Service: "s"})
	found, err := m.GetDeployment(d.ID)
	if err != nil || found.ID != d.ID {
		t.Error("failed to get deployment")
	}
}

func TestGetDeploymentNotFound(t *testing.T) {
	m := NewManager()
	_, err := m.GetDeployment("missing")
	if err == nil {
		t.Error("expected error")
	}
}

func TestDeploymentPromote(t *testing.T) {
	m := NewManager()
	d, _ := m.CreateDeployment(DeploymentConfig{Name: "T", Service: "s"})
	d.Status = StatusRunning
	d.CurrentPhase = len(d.Phases)
	d.promote()
	if d.Status != StatusPromoted || d.TrafficSplit.Canary != 100.0 {
		t.Error("promotion failed")
	}
}

func TestDeploymentRollback(t *testing.T) {
	m := NewManager()
	d, _ := m.CreateDeployment(DeploymentConfig{Name: "T", Service: "s"})
	d.Status = StatusRunning
	d.Rollback()
	if d.Status != StatusRolledBack || d.TrafficSplit.Stable != 100.0 {
		t.Error("rollback failed")
	}
}

func TestDeploymentAbort(t *testing.T) {
	m := NewManager()
	d, _ := m.CreateDeployment(DeploymentConfig{Name: "T", Service: "s"})
	d.Status = StatusPending
	d.Abort("test")
	if d.Status != StatusAborted {
		t.Error("abort failed")
	}
}

func TestGeneratePhasesLinear(t *testing.T) {
	phases := generatePhases(DeploymentConfig{Strategy: StrategyLinear, Steps: 5})
	if len(phases) != 5 || phases[0].TrafficWeight != 20.0 || phases[4].TrafficWeight != 100.0 {
		t.Error("linear phases incorrect")
	}
}

func TestGeneratePhasesStepped(t *testing.T) {
	phases := generatePhases(DeploymentConfig{Strategy: StrategyStepped})
	if len(phases) != 6 {
		t.Errorf("expected 6 phases, got %d", len(phases))
	}
	expected := []float64{5, 10, 25, 50, 75, 100}
	for i, exp := range expected {
		if phases[i].TrafficWeight != exp {
			t.Errorf("phase %d: expected %.0f, got %.2f", i, exp, phases[i].TrafficWeight)
		}
	}
}

func TestPhaseAnalysis(t *testing.T) {
	d := &Deployment{Metrics: Metrics{ErrorRateThreshold: 5.0, LatencyP99Threshold: 500 * time.Millisecond, LatencyP95Threshold: 200 * time.Millisecond}}
	if !d.analyzePhase(&Phase{Metrics: PhaseMetrics{ErrorRate: 2.0, LatencyP95: 100 * time.Millisecond, LatencyP99: 400 * time.Millisecond}}) {
		t.Error("expected pass")
	}
	if d.analyzePhase(&Phase{Metrics: PhaseMetrics{ErrorRate: 10.0}}) {
		t.Error("expected fail on error rate")
	}
	if d.analyzePhase(&Phase{Metrics: PhaseMetrics{ErrorRate: 2.0, LatencyP95: 300 * time.Millisecond}}) {
		t.Error("expected fail on latency")
	}
}

func TestDeploymentEvents(t *testing.T) {
	m := NewManager()
	d, _ := m.CreateDeployment(DeploymentConfig{Name: "T", Service: "s"})
	events := make([]Event, 0)
	d.OnEvent(func(e Event) { events = append(events, e) })
	d.emit(Event{Type: EventStarted, Timestamp: time.Now(), Message: "test"})
	time.Sleep(10 * time.Millisecond)
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}

func TestRollbackAlreadyRolledBack(t *testing.T) {
	m := NewManager()
	d, _ := m.CreateDeployment(DeploymentConfig{Name: "T", Service: "s"})
	d.Status = StatusRolledBack
	if d.Rollback() == nil {
		t.Error("expected error")
	}
}

func BenchmarkCreateDeployment(b *testing.B) {
	m := NewManager()
	cfg := DeploymentConfig{Name: "Bench", Service: "test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CreateDeployment(cfg)
	}
}
