package otelm

import (
	"fmt"
	"sync"
)

type Span struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type OTelManager struct {
	spans []Span
	mu    sync.RWMutex
}

func NewOTelManager() *OTelManager {
	return &OTelManager{}
}

func (m *OTelManager) StartSpan(name string) *Span {
	m.mu.Lock()
	defer m.mu.Unlock()
	span := &Span{Name: name, Status: "active"}
	m.spans = append(m.spans, *span)
	return span
}

func (m *OTelManager) EndSpan(span *Span) {
	m.mu.Lock()
	defer m.mu.Unlock()
	span.Status = "completed"
	for i := range m.spans {
		if m.spans[i].Name == span.Name {
			m.spans[i].Status = "completed"
		}
	}
}

func (m *OTelManager) GetSpans() []Span {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]Span{}, m.spans...)
}

func (m *OTelManager) Export() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result string
	for _, s := range m.spans {
		result += fmt.Sprintf("span: %s status: %s\n", s.Name, s.Status)
	}
	return result
}
