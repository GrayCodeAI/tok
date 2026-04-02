package filterejection

import (
	"strings"
	"sync"
)

type EjectionReason string

const (
	EjectionSecurity    EjectionReason = "security"
	EjectionPerformance EjectionReason = "performance"
	EjectionQuality     EjectionReason = "quality"
	EjectionManual      EjectionReason = "manual"
	EjectionConflict    EjectionReason = "conflict"
)

type EjectedFilter struct {
	FilterName string         `json:"filter_name"`
	Reason     EjectionReason `json:"reason"`
	Details    string         `json:"details"`
	EjectedAt  string         `json:"ejected_at"`
}

type EjectionEngine struct {
	ejected  map[string]*EjectedFilter
	registry map[string]bool
	mu       sync.RWMutex
}

func NewEjectionEngine() *EjectionEngine {
	return &EjectionEngine{
		ejected:  make(map[string]*EjectedFilter),
		registry: make(map[string]bool),
	}
}

func (e *EjectionEngine) Eject(filterName string, reason EjectionReason, details string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.ejected[filterName] = &EjectedFilter{
		FilterName: filterName,
		Reason:     reason,
		Details:    details,
	}
}

func (e *EjectionEngine) IsEjected(filterName string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, ejected := e.ejected[filterName]
	return ejected
}

func (e *EjectionEngine) Restore(filterName string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.ejected, filterName)
}

func (e *EjectionEngine) GetEjected() []*EjectedFilter {
	e.mu.RLock()
	defer e.mu.RUnlock()
	var result []*EjectedFilter
	for _, ef := range e.ejected {
		result = append(result, ef)
	}
	return result
}

func (e *EjectionEngine) EjectByPattern(pattern string, reason EjectionReason) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	count := 0
	for name := range e.registry {
		if strings.Contains(name, pattern) {
			e.ejected[name] = &EjectedFilter{
				FilterName: name,
				Reason:     reason,
				Details:    "Ejected by pattern: " + pattern,
			}
			count++
		}
	}
	return count
}

func (e *EjectionEngine) FilterActive(filters []string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	var active []string
	for _, f := range filters {
		if _, ejected := e.ejected[f]; !ejected {
			active = append(active, f)
		}
	}
	return active
}

func (e *EjectionEngine) Register(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.registry[name] = true
}

func (e *EjectionEngine) Stats() map[string]int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return map[string]int{
		"registered": len(e.registry),
		"ejected":    len(e.ejected),
		"active":     len(e.registry) - len(e.ejected),
	}
}
