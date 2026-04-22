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
				return NewPipelineCoordinator(&config)
			},
		},
	}
}

// Get retrieves a coordinator from pool.
// If the pool contains an unexpected type, a fresh coordinator is created.
func (p *CoordinatorPool) Get() *PipelineCoordinator {
	v := p.pool.Get()
	coord, ok := v.(*PipelineCoordinator)
	if !ok {
		// Pool corruption or nil value: allocate fresh
		return NewPipelineCoordinator(&p.config)
	}
	return coord
}

// Put returns coordinator to pool
func (p *CoordinatorPool) Put(coord *PipelineCoordinator) {
	// Reset coordinator state before returning to pool
	coord.reset()
	p.pool.Put(coord)
}

// reset clears coordinator state for reuse
func (p *PipelineCoordinator) reset() {
	p.processedLayers = 0
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

// ProcessWithPool processes input using a caller-provided pool.
// Use NewCoordinatorPool to create a pool for a fixed config,
// or call GetDefaultPool for the global default pool.
func ProcessWithPool(input string, pool *CoordinatorPool) (string, *PipelineStats) {
	coord := pool.Get()
	output, stats := coord.Process(input)
	pool.Put(coord)
	return output, stats
}
