package remotegain

import (
	"encoding/json"
	"sync"
)

type RemoteMachine struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
}

type RemoteGain struct {
	LocalTotal   int64            `json:"local_total"`
	LocalDaily   map[string]int64 `json:"local_daily"`
	RemoteTotals map[string]int64 `json:"remote_totals"`
	machines     map[string]*RemoteMachine
	mu           sync.RWMutex
}

func NewRemoteGain() *RemoteGain {
	return &RemoteGain{
		LocalDaily:   make(map[string]int64),
		RemoteTotals: make(map[string]int64),
		machines:     make(map[string]*RemoteMachine),
	}
}

func (g *RemoteGain) RecordLocal(savedTokens int64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.LocalTotal += savedTokens
}

func (g *RemoteGain) RecordLocalDaily(date string, savedTokens int64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.LocalDaily[date] += savedTokens
}

func (g *RemoteGain) RegisterMachine(machine *RemoteMachine) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.machines[machine.ID] = machine
}

func (g *RemoteGain) UpdateRemote(machineID string, total int64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.RemoteTotals[machineID] = total
}

func (g *RemoteGain) GetAggregatedTotal() int64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	total := g.LocalTotal
	for _, v := range g.RemoteTotals {
		total += v
	}
	return total
}

func (g *RemoteGain) GetMachineBreakdown() map[string]int64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make(map[string]int64)
	result["local"] = g.LocalTotal
	for id, total := range g.RemoteTotals {
		result[id] = total
	}
	return result
}

func (g *RemoteGain) ExportJSON() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return json.MarshalIndent(struct {
		LocalTotal   int64            `json:"local_total"`
		LocalDaily   map[string]int64 `json:"local_daily"`
		RemoteTotals map[string]int64 `json:"remote_totals"`
		Aggregated   int64            `json:"aggregated_total"`
	}{
		LocalTotal:   g.LocalTotal,
		LocalDaily:   g.LocalDaily,
		RemoteTotals: g.RemoteTotals,
		Aggregated:   g.GetAggregatedTotal(),
	}, "", "  ")
}

func (g *RemoteGain) ListMachines() []*RemoteMachine {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var result []*RemoteMachine
	for _, m := range g.machines {
		result = append(result, m)
	}
	return result
}

func (g *RemoteGain) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.LocalTotal = 0
	g.LocalDaily = make(map[string]int64)
	g.RemoteTotals = make(map[string]int64)
}
