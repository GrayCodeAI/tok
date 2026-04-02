package multiclient

import (
	"strings"
	"sync"
)

type ClientType string

const (
	ClientOpenCode   ClientType = "opencode"
	ClientClaudeCode ClientType = "claude-code"
	ClientCodex      ClientType = "codex"
	ClientGemini     ClientType = "gemini"
	ClientCursor     ClientType = "cursor"
	ClientAmp        ClientType = "amp"
	ClientDroid      ClientType = "droid"
	ClientOpenClaw   ClientType = "openclaw"
	ClientPi         ClientType = "pi"
	ClientKimi       ClientType = "kimi"
	ClientQwen       ClientType = "qwen"
	ClientRooCode    ClientType = "roo-code"
	ClientKilo       ClientType = "kilo"
	ClientMux        ClientType = "mux"
	ClientCrush      ClientType = "crush"
	ClientSynthetic  ClientType = "synthetic"
)

type ClientSession struct {
	Client    ClientType `json:"client"`
	SessionID string     `json:"session_id"`
	Tokens    int64      `json:"tokens"`
	Cost      float64    `json:"cost"`
}

type MultiClientTracker struct {
	sessions map[string]*ClientSession
	mu       sync.RWMutex
}

func NewMultiClientTracker() *MultiClientTracker {
	return &MultiClientTracker{
		sessions: make(map[string]*ClientSession),
	}
}

func (t *MultiClientTracker) Record(client ClientType, sessionID string, tokens int64, cost float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := string(client) + ":" + sessionID
	if session, ok := t.sessions[key]; ok {
		session.Tokens += tokens
		session.Cost += cost
	} else {
		t.sessions[key] = &ClientSession{
			Client:    client,
			SessionID: sessionID,
			Tokens:    tokens,
			Cost:      cost,
		}
	}
}

func (t *MultiClientTracker) GetByClient(client ClientType) []*ClientSession {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var result []*ClientSession
	for _, s := range t.sessions {
		if s.Client == client {
			result = append(result, s)
		}
	}
	return result
}

func (t *MultiClientTracker) GetAll() []*ClientSession {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var result []*ClientSession
	for _, s := range t.sessions {
		result = append(result, s)
	}
	return result
}

func (t *MultiClientTracker) GetClients() []ClientType {
	t.mu.RLock()
	defer t.mu.RUnlock()
	seen := make(map[ClientType]bool)
	var result []ClientType
	for _, s := range t.sessions {
		if !seen[s.Client] {
			seen[s.Client] = true
			result = append(result, s.Client)
		}
	}
	return result
}

func (t *MultiClientTracker) FilterByClient(clients []ClientType) []*ClientSession {
	t.mu.RLock()
	defer t.mu.RUnlock()
	clientMap := make(map[ClientType]bool)
	for _, c := range clients {
		clientMap[c] = true
	}
	var result []*ClientSession
	for _, s := range t.sessions {
		if clientMap[s.Client] {
			result = append(result, s)
		}
	}
	return result
}

type DateFilter struct {
	Today bool
	Week  bool
	Month bool
	Since string
	Until string
}

type GroupByStrategy string

const (
	GroupByModel       GroupByStrategy = "model"
	GroupByClientModel GroupByStrategy = "client+model"
	GroupByClient      GroupByStrategy = "client"
	GroupByDate        GroupByStrategy = "date"
)

type Grouper struct {
	strategy GroupByStrategy
}

func NewGrouper(strategy GroupByStrategy) *Grouper {
	return &Grouper{strategy: strategy}
}

func (g *Grouper) SetStrategy(strategy GroupByStrategy) {
	g.strategy = strategy
}

func (g *Grouper) GetStrategy() GroupByStrategy {
	return g.strategy
}

func ParseClientType(s string) ClientType {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "opencode", "open-code":
		return ClientOpenCode
	case "claude-code", "claudecode":
		return ClientClaudeCode
	case "codex", "codex-cli":
		return ClientCodex
	case "gemini", "gemini-cli":
		return ClientGemini
	case "cursor":
		return ClientCursor
	case "amp":
		return ClientAmp
	case "droid":
		return ClientDroid
	case "openclaw", "open-claw":
		return ClientOpenClaw
	case "pi":
		return ClientPi
	case "kimi":
		return ClientKimi
	case "qwen":
		return ClientQwen
	case "roo-code", "roocode":
		return ClientRooCode
	case "kilo":
		return ClientKilo
	case "mux":
		return ClientMux
	case "crush":
		return ClientCrush
	case "synthetic":
		return ClientSynthetic
	default:
		return ClientOpenCode
	}
}
