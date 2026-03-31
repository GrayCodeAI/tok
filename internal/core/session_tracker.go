package core

import (
	"sync"
	"time"
)

type SessionInfo struct {
	ID           string
	StartedAt    time.Time
	LastActive   time.Time
	CommandCount int
	TokensSaved  int
	Agent        string
}

type SessionTracker struct {
	mu       sync.RWMutex
	sessions map[string]*SessionInfo
	current  string
}

func NewSessionTracker() *SessionTracker {
	return &SessionTracker{
		sessions: make(map[string]*SessionInfo),
	}
}

func (s *SessionTracker) StartSession(id, agent string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = id
	s.sessions[id] = &SessionInfo{
		ID:         id,
		StartedAt:  time.Now(),
		LastActive: time.Now(),
		Agent:      agent,
	}
}

func (s *SessionTracker) RecordCommand(id string, tokensSaved int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[id]; ok {
		sess.CommandCount++
		sess.TokensSaved += tokensSaved
		sess.LastActive = time.Now()
	}
}

func (s *SessionTracker) GetActiveSessions() []*SessionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var active []*SessionInfo
	cutoff := time.Now().Add(-30 * time.Minute)
	for _, sess := range s.sessions {
		if sess.LastActive.After(cutoff) {
			active = append(active, sess)
		}
	}
	return active
}

func (s *SessionTracker) GetAllSessions() []*SessionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var all []*SessionInfo
	for _, sess := range s.sessions {
		all = append(all, sess)
	}
	return all
}

func (s *SessionTracker) GetAdoptionRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.sessions) == 0 {
		return 0
	}
	var adopted int
	for _, sess := range s.sessions {
		if sess.CommandCount > 0 {
			adopted++
		}
	}
	return float64(adopted) / float64(len(s.sessions)) * 100
}
