package filter

import (
	"sync"
)

// CoordinatorPool manages reusable pipeline coordinators
type CoordinatorPool struct {
	pool   sync.Pool
	config PipelineConfig
}

// NewCoordinatorPool creates a coordinator pool
func NewCoordinatorPool(config PipelineConfig) *CoordinatorPool {
	return &CoordinatorPool{
		config: config,
		pool: sync.Pool{
			New: func() any {
				return NewPipelineCoordinator(config)
			},
		},
	}
}

// Get retrieves a coordinator from pool
func (p *CoordinatorPool) Get() *PipelineCoordinator {
	return p.pool.Get().(*PipelineCoordinator)
}

// Put returns coordinator to pool
func (p *CoordinatorPool) Put(coord *PipelineCoordinator) {
	// Reset coordinator state before returning to pool
	coord.reset()
	p.pool.Put(coord)
}

// reset clears coordinator state for reuse
func (p *PipelineCoordinator) reset() {
	// Clear caches but keep filter instances
	if p.layerCache != nil {
		p.layerCache.Clear()
	}
}

// Global pool with default config
var defaultPool *CoordinatorPool
var poolOnce sync.Once

// GetDefaultPool returns the global coordinator pool
func GetDefaultPool() *CoordinatorPool {
	poolOnce.Do(func() {
		defaultPool = NewCoordinatorPool(PipelineConfig{
			Mode:          ModeMinimal,
			EnableEntropy: true,
			EnableAST:     true,
			Budget:        1000, // Enable budget enforcer with 1000 token limit
		})
	})
	return defaultPool
}

// ProcessWithPool processes input using pooled coordinator
func ProcessWithPool(input string, config PipelineConfig) (string, *PipelineStats) {
	pool := NewCoordinatorPool(config)
	coord := pool.Get()
	defer pool.Put(coord)
	output, stats, err := coord.Process(input)
	if err != nil {
		return input, nil
	}
	return output, stats
}
