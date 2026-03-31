package core

import (
	"sync"
	"time"
)

type LiveCommand struct {
	Command      string
	PID          int
	StartedAt    time.Time
	InputTokens  int
	OutputTokens int
	SavedTokens  int
	Status       string
}

type LiveMonitor struct {
	mu       sync.RWMutex
	commands map[int]*LiveCommand
}

func NewLiveMonitor() *LiveMonitor {
	return &LiveMonitor{commands: make(map[int]*LiveCommand)}
}

func (m *LiveMonitor) StartCommand(pid int, cmd string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commands[pid] = &LiveCommand{
		Command:   cmd,
		PID:       pid,
		StartedAt: time.Now(),
		Status:    "running",
	}
}

func (m *LiveMonitor) EndCommand(pid int, inputTokens, outputTokens int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.commands[pid]; ok {
		c.InputTokens = inputTokens
		c.OutputTokens = outputTokens
		c.SavedTokens = inputTokens - outputTokens
		c.Status = "done"
	}
}

func (m *LiveMonitor) GetActive() []*LiveCommand {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var active []*LiveCommand
	for _, c := range m.commands {
		if c.Status == "running" {
			active = append(active, c)
		}
	}
	return active
}

func (m *LiveMonitor) GetRecent(n int) []*LiveCommand {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var all []*LiveCommand
	for _, c := range m.commands {
		all = append(all, c)
	}
	if len(all) > n {
		all = all[len(all)-n:]
	}
	return all
}

func (m *LiveMonitor) Cleanup(maxAge time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().Add(-maxAge)
	for pid, c := range m.commands {
		if c.Status == "done" && c.StartedAt.Before(cutoff) {
			delete(m.commands, pid)
		}
	}
}
