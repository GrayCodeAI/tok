package wsfeed

import (
	"encoding/json"
	"sync"
	"time"
)

type FeedEvent struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

type WebSocketFeed struct {
	subscribers map[string]chan FeedEvent
	mu          sync.RWMutex
}

func NewWebSocketFeed() *WebSocketFeed {
	return &WebSocketFeed{
		subscribers: make(map[string]chan FeedEvent),
	}
}

func (f *WebSocketFeed) Subscribe(id string) <-chan FeedEvent {
	f.mu.Lock()
	defer f.mu.Unlock()
	ch := make(chan FeedEvent, 100)
	f.subscribers[id] = ch
	return ch
}

func (f *WebSocketFeed) Unsubscribe(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if ch, ok := f.subscribers[id]; ok {
		close(ch)
		delete(f.subscribers, id)
	}
}

func (f *WebSocketFeed) Broadcast(event FeedEvent) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	event.Timestamp = time.Now()
	for _, ch := range f.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

func (f *WebSocketFeed) BroadcastCostAlert(message string, usd float64) {
	f.Broadcast(FeedEvent{
		Type: "cost_alert",
		Data: map[string]interface{}{
			"message": message,
			"usd":     usd,
		},
	})
}

func (f *WebSocketFeed) BroadcastRequest(source, model string, inputTokens, outputTokens int) {
	f.Broadcast(FeedEvent{
		Type: "request",
		Data: map[string]interface{}{
			"source":        source,
			"model":         model,
			"input_tokens":  inputTokens,
			"output_tokens": outputTokens,
		},
	})
}

func (f *WebSocketFeed) ExportJSON() ([]byte, error) {
	return json.MarshalIndent(f.subscribers, "", "  ")
}

func (f *WebSocketFeed) SubscriberCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.subscribers)
}

type BackgroundService struct {
	running bool
	mu      sync.RWMutex
}

func NewBackgroundService() *BackgroundService {
	return &BackgroundService{}
}

func (s *BackgroundService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = true
}

func (s *BackgroundService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
}

func (s *BackgroundService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *BackgroundService) Status() string {
	if s.IsRunning() {
		return "running"
	}
	return "stopped"
}
