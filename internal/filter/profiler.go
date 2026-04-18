package filter

import (
	"sync"
	"time"
)

// Profiler tracks layer performance
type Profiler struct {
	mu     sync.Mutex
	layers map[string]*LayerMetrics
}

type LayerMetrics struct {
	Calls     int64
	TotalTime time.Duration
	AvgTime   time.Duration
	MaxTime   time.Duration
}

func NewProfiler() *Profiler {
	return &Profiler{layers: make(map[string]*LayerMetrics)}
}

func (p *Profiler) Track(name string, fn func()) {
	start := time.Now()
	fn()
	elapsed := time.Since(start)
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	m, ok := p.layers[name]
	if !ok {
		m = &LayerMetrics{}
		p.layers[name] = m
	}
	
	m.Calls++
	m.TotalTime += elapsed
	m.AvgTime = m.TotalTime / time.Duration(m.Calls)
	if elapsed > m.MaxTime {
		m.MaxTime = elapsed
	}
}

func (p *Profiler) GetMetrics() map[string]*LayerMetrics {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	result := make(map[string]*LayerMetrics, len(p.layers))
	for k, v := range p.layers {
		result[k] = v
	}
	return result
}
