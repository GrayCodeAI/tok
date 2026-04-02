package remotegain

import "testing"

func TestRemoteGain(t *testing.T) {
	g := NewRemoteGain()

	g.RecordLocal(100)
	g.RecordLocal(200)
	g.RecordLocalDaily("2026-01-01", 150)

	if g.LocalTotal != 300 {
		t.Errorf("Expected 300 local total, got %d", g.LocalTotal)
	}

	g.RegisterMachine(&RemoteMachine{ID: "m1", Name: "Laptop", Endpoint: "https://laptop.example.com"})
	g.UpdateRemote("m1", 500)

	total := g.GetAggregatedTotal()
	if total != 800 {
		t.Errorf("Expected 800 aggregated total, got %d", total)
	}
}

func TestRemoteGainBreakdown(t *testing.T) {
	g := NewRemoteGain()
	g.RecordLocal(100)
	g.RegisterMachine(&RemoteMachine{ID: "m1", Name: "Server", Endpoint: "https://server.example.com"})
	g.UpdateRemote("m1", 200)

	breakdown := g.GetMachineBreakdown()
	if breakdown["local"] != 100 {
		t.Errorf("Expected 100 local, got %d", breakdown["local"])
	}
	if breakdown["m1"] != 200 {
		t.Errorf("Expected 200 remote, got %d", breakdown["m1"])
	}
}

func TestRemoteGainExport(t *testing.T) {
	g := NewRemoteGain()
	g.RecordLocal(100)

	export, err := g.ExportJSON()
	if err != nil {
		t.Fatalf("ExportJSON error: %v", err)
	}
	if len(export) == 0 {
		t.Error("Expected non-empty export")
	}
}

func TestRemoteGainReset(t *testing.T) {
	g := NewRemoteGain()
	g.RecordLocal(100)
	g.Reset()

	if g.LocalTotal != 0 {
		t.Errorf("Expected 0 after reset, got %d", g.LocalTotal)
	}
}
