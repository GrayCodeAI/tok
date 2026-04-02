package routing

import (
	"math/rand"
	"sync"
	"time"
)

type FallbackChain struct {
	Primary   string   `json:"primary"`
	Fallbacks []string `json:"fallbacks"`
	Current   int      `json:"current"`
}

type FallbackManager struct {
	chains map[string]*FallbackChain
	mu     sync.RWMutex
}

func NewFallbackManager() *FallbackManager {
	return &FallbackManager{
		chains: make(map[string]*FallbackChain),
	}
}

func (f *FallbackManager) Register(source, primary string, fallbacks ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.chains[source] = &FallbackChain{
		Primary:   primary,
		Fallbacks: fallbacks,
		Current:   0,
	}
}

func (f *FallbackManager) GetModel(source string, statusCode int) string {
	f.mu.Lock()
	defer f.mu.Unlock()

	chain, ok := f.chains[source]
	if !ok {
		return source
	}

	if statusCode == 200 || statusCode == 0 {
		return chain.Primary
	}

	if statusCode == 429 || statusCode == 500 || statusCode == 502 || statusCode == 503 || statusCode == 504 {
		chain.Current++
		if chain.Current > len(chain.Fallbacks) {
			chain.Current = 0
		}
		if chain.Current == 0 {
			return chain.Primary
		}
		return chain.Fallbacks[chain.Current-1]
	}

	return chain.Primary
}

func (f *FallbackManager) Reset(source string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if chain, ok := f.chains[source]; ok {
		chain.Current = 0
	}
}

type WeightedRoute struct {
	Model  string `json:"model"`
	Weight int    `json:"weight"`
}

type WeightedBalancer struct {
	routes []WeightedRoute
	mu     sync.RWMutex
}

func NewWeightedBalancer() *WeightedBalancer {
	return &WeightedBalancer{}
}

func (b *WeightedBalancer) AddRoute(model string, weight int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.routes = append(b.routes, WeightedRoute{Model: model, Weight: weight})
}

func (b *WeightedBalancer) Select() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	totalWeight := 0
	for _, r := range b.routes {
		totalWeight += r.Weight
	}

	if totalWeight == 0 {
		return ""
	}

	r := rand.Intn(totalWeight)
	cumulative := 0
	for _, route := range b.routes {
		cumulative += route.Weight
		if r < cumulative {
			return route.Model
		}
	}

	return b.routes[0].Model
}

type LatencyTracker struct {
	latencies map[string][]time.Duration
	mu        sync.RWMutex
}

func NewLatencyTracker() *LatencyTracker {
	return &LatencyTracker{
		latencies: make(map[string][]time.Duration),
	}
}

func (t *LatencyTracker) Record(model string, latency time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.latencies[model] = append(t.latencies[model], latency)
	if len(t.latencies[model]) > 100 {
		t.latencies[model] = t.latencies[model][1:]
	}
}

func (t *LatencyTracker) GetP50(model string) time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.percentile(model, 0.50)
}

func (t *LatencyTracker) GetP95(model string) time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.percentile(model, 0.95)
}

func (t *LatencyTracker) GetP99(model string) time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.percentile(model, 0.99)
}

func (t *LatencyTracker) percentile(model string, p float64) time.Duration {
	latencies := t.latencies[model]
	if len(latencies) == 0 {
		return 0
	}

	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	idx := int(float64(len(sorted)) * p)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func (t *LatencyTracker) FastestModel() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var best string
	var bestP50 time.Duration
	for model := range t.latencies {
		p50 := t.percentile(model, 0.50)
		if best == "" || p50 < bestP50 {
			best = model
			bestP50 = p50
		}
	}
	return best
}
