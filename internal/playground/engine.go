package playground

import (
	"fmt"
	"strings"
	"sync"
)

type Session struct {
	ID          string    `json:"id"`
	Model       string    `json:"model"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	History     []Message `json:"history"`
	Source      string    `json:"source"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Tokens  int    `json:"tokens"`
}

type PlaygroundEngine struct {
	sessions map[string]*Session
	counter  int
	mu       sync.RWMutex
}

func NewPlaygroundEngine() *PlaygroundEngine {
	return &PlaygroundEngine{
		sessions: make(map[string]*Session),
	}
}

func (e *PlaygroundEngine) CreateSession(model string) *Session {
	e.counter++
	session := &Session{
		ID:          fmt.Sprintf("session-%d", e.counter),
		Model:       model,
		Temperature: 0.7,
		MaxTokens:   4096,
	}
	e.mu.Lock()
	e.sessions[session.ID] = session
	e.mu.Unlock()
	return session
}

func (e *PlaygroundEngine) AddMessage(sessionID, role, content string) *Message {
	e.mu.Lock()
	defer e.mu.Unlock()
	session, ok := e.sessions[sessionID]
	if !ok {
		return nil
	}
	msg := Message{
		Role:    role,
		Content: content,
		Tokens:  len(content) / 4,
	}
	session.History = append(session.History, msg)
	return &msg
}

func (e *PlaygroundEngine) GetSession(sessionID string) *Session {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.sessions[sessionID]
}

func (e *PlaygroundEngine) EstimateCost(sessionID string, inputCost, outputCost float64) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	session, ok := e.sessions[sessionID]
	if !ok {
		return 0
	}
	totalInput := 0
	totalOutput := 0
	for _, msg := range session.History {
		if msg.Role == "user" {
			totalInput += msg.Tokens
		} else {
			totalOutput += msg.Tokens
		}
	}
	return float64(totalInput)/1000*inputCost + float64(totalOutput)/1000*outputCost
}

func (e *PlaygroundEngine) CompareModels(sessionID string, models []string) map[string]float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	results := make(map[string]float64)
	session, ok := e.sessions[sessionID]
	if !ok {
		return results
	}
	totalTokens := 0
	for _, msg := range session.History {
		totalTokens += msg.Tokens
	}
	for _, model := range models {
		results[model] = float64(totalTokens) * 0.001
	}
	return results
}

func (e *PlaygroundEngine) ExportSession(sessionID string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	session, ok := e.sessions[sessionID]
	if !ok {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("# Playground Session\n\n")
	sb.WriteString("Model: " + session.Model + "\n\n")
	for _, msg := range session.History {
		sb.WriteString("## " + msg.Role + "\n")
		sb.WriteString(msg.Content + "\n\n")
	}
	return sb.String()
}

func (e *PlaygroundEngine) ListSessions() []*Session {
	e.mu.RLock()
	defer e.mu.RUnlock()
	var sessions []*Session
	for _, s := range e.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}
