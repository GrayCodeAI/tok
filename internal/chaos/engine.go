// Package chaos provides chaos engineering capabilities for TokMan
package chaos

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ExperimentType defines the type of chaos experiment
type ExperimentType string

const (
	TypeLatency    ExperimentType = "latency"
	TypeError      ExperimentType = "error"
	TypeMemory     ExperimentType = "memory"
	TypeCPU        ExperimentType = "cpu"
	TypeNetwork    ExperimentType = "network"
	TypeDisk       ExperimentType = "disk"
	TypeKill       ExperimentType = "kill"
	TypeTimeDrift  ExperimentType = "time_drift"
	TypeDependency ExperimentType = "dependency"
)

// Engine manages chaos experiments
type Engine struct {
	experiments map[string]*Experiment
	running     map[string]bool
	mu          sync.RWMutex
	handlers    map[ExperimentType]FaultHandler
}

// FaultHandler is a function that injects a specific type of fault
type FaultHandler func(ctx context.Context, config FaultConfig) error

// Experiment defines a chaos experiment
type Experiment struct {
	ID          string
	Name        string
	Description string
	Type        ExperimentType
	Config      FaultConfig
	Scope       ScopeConfig
	Schedule    ScheduleConfig
	Safety      SafetyConfig
	status      Status
}

// FaultConfig holds fault injection configuration
type FaultConfig struct {
	Probability float64 // 0.0 to 1.0
	Duration    time.Duration
	Intensity   float64 // 0.0 to 1.0
	Target      string  // specific target identifier
	Attributes  map[string]interface{}
}

// ScopeConfig defines the blast radius
type ScopeConfig struct {
	Services   []string
	Instances  []string
	Percentage float64 // percentage of targets affected
	Tags       map[string]string
}

// ScheduleConfig defines when to run the experiment
type ScheduleConfig struct {
	StartTime  time.Time
	EndTime    time.Time
	Cron       string
	Duration   time.Duration
	Continuous bool
}

// SafetyConfig defines safety mechanisms
type SafetyConfig struct {
	AutoRollback bool
	MaxDuration  time.Duration
	AbortOnError bool
	HealthCheck  func() bool
	KillSwitch   chan struct{}
}

// Status represents experiment status
type Status struct {
	State     State
	StartTime time.Time
	EndTime   time.Time
	Runs      int
	Successes int
	Failures  int
	LastError error
}

// State represents experiment state
type State string

const (
	StatePending   State = "pending"
	StateRunning   State = "running"
	StatePaused    State = "paused"
	StateCompleted State = "completed"
	StateFailed    State = "failed"
	StateAborted   State = "aborted"
)

// NewEngine creates a chaos engineering engine
func NewEngine() *Engine {
	return &Engine{
		experiments: make(map[string]*Experiment),
		running:     make(map[string]bool),
		handlers:    make(map[ExperimentType]FaultHandler),
	}
}

// RegisterHandler registers a fault handler
func (e *Engine) RegisterHandler(t ExperimentType, h FaultHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[t] = h
}

// RegisterExperiment adds an experiment
func (e *Engine) RegisterExperiment(ex *Experiment) error {
	if ex.ID == "" {
		ex.ID = generateID()
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.experiments[ex.ID]; exists {
		return fmt.Errorf("experiment %s already exists", ex.ID)
	}

	ex.status = Status{State: StatePending}
	e.experiments[ex.ID] = ex

	return nil
}

// StartExperiment begins a chaos experiment
func (e *Engine) StartExperiment(ctx context.Context, id string) error {
	e.mu.Lock()
	ex, exists := e.experiments[id]
	handler, handlerExists := e.handlers[ex.Type]
	e.running[id] = true
	e.mu.Unlock()

	if !exists {
		return fmt.Errorf("experiment %s not found", id)
	}

	if !handlerExists {
		return fmt.Errorf("no handler registered for type %s", ex.Type)
	}

	ex.status.State = StateRunning
	ex.status.StartTime = time.Now()

	// Run safety health check
	if ex.Safety.HealthCheck != nil && !ex.Safety.HealthCheck() {
		ex.status.State = StateAborted
		return fmt.Errorf("health check failed, aborting experiment")
	}

	// Set up kill switch
	killSwitch := ex.Safety.KillSwitch
	if killSwitch == nil {
		killSwitch = make(chan struct{})
	}

	// Create experiment context with timeout
	expCtx, cancel := context.WithTimeout(ctx, ex.Safety.MaxDuration)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- e.runExperiment(expCtx, ex, handler)
	}()

	select {
	case err := <-done:
		ex.status.EndTime = time.Now()
		if err != nil {
			ex.status.State = StateFailed
			ex.status.LastError = err
			if ex.Safety.AbortOnError {
				return err
			}
		} else {
			ex.status.State = StateCompleted
		}
	case <-killSwitch:
		ex.status.State = StateAborted
		ex.status.EndTime = time.Now()
		return fmt.Errorf("experiment aborted via kill switch")
	}

	return nil
}

func (e *Engine) runExperiment(ctx context.Context, ex *Experiment, handler FaultHandler) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Check if we should inject fault based on probability
			if rand.Float64() <= ex.Config.Probability {
				ex.status.Runs++
				if err := handler(ctx, ex.Config); err != nil {
					ex.status.Failures++
					if ex.Safety.AbortOnError {
						return err
					}
				} else {
					ex.status.Successes++
				}
			}

			// Check if experiment should end
			if !ex.Schedule.Continuous && time.Now().After(ex.Schedule.EndTime) {
				return nil
			}
		}
	}
}

// StopExperiment stops a running experiment
func (e *Engine) StopExperiment(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ex, exists := e.experiments[id]
	if !exists {
		return fmt.Errorf("experiment %s not found", id)
	}

	if ex.status.State != StateRunning {
		return fmt.Errorf("experiment %s is not running", id)
	}

	e.running[id] = false
	ex.status.State = StateAborted
	ex.status.EndTime = time.Now()

	return nil
}

// GetExperiment returns experiment details
func (e *Engine) GetExperiment(id string) (*Experiment, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ex, exists := e.experiments[id]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", id)
	}

	return ex, nil
}

// ListExperiments returns all experiments
func (e *Engine) ListExperiments() []*Experiment {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]*Experiment, 0, len(e.experiments))
	for _, ex := range e.experiments {
		result = append(result, ex)
	}

	return result
}

// GetStatus returns experiment status
func (e *Engine) GetStatus(id string) (Status, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ex, exists := e.experiments[id]
	if !exists {
		return Status{}, fmt.Errorf("experiment %s not found", id)
	}

	return ex.status, nil
}

// StandardFaultHandlers returns default fault handlers
func StandardFaultHandlers() map[ExperimentType]FaultHandler {
	return map[ExperimentType]FaultHandler{
		TypeLatency: func(ctx context.Context, config FaultConfig) error {
			delay := time.Duration(float64(config.Duration) * config.Intensity)
			time.Sleep(delay)
			return nil
		},
		TypeError: func(ctx context.Context, config FaultConfig) error {
			return fmt.Errorf("injected error: %s", config.Target)
		},
		TypeMemory: func(ctx context.Context, config FaultConfig) error {
			size := int(config.Intensity * 100 * 1024 * 1024) // MB
			_ = make([]byte, size)
			return nil
		},
		TypeCPU: func(ctx context.Context, config FaultConfig) error {
			duration := time.Duration(float64(config.Duration) * config.Intensity)
			end := time.Now().Add(duration)
			for time.Now().Before(end) {
				// Busy loop to consume CPU
				_ = rand.Float64() * rand.Float64()
			}
			return nil
		},
	}
}

// NewExperiment creates a new experiment with sensible defaults
func NewExperiment(name string, t ExperimentType) *Experiment {
	return &Experiment{
		ID:     generateID(),
		Name:   name,
		Type:   t,
		status: Status{State: StatePending},
		Config: FaultConfig{
			Probability: 0.1,
			Duration:    5 * time.Second,
			Intensity:   0.5,
			Attributes:  make(map[string]interface{}),
		},
		Scope: ScopeConfig{
			Percentage: 10.0,
			Tags:       make(map[string]string),
		},
		Schedule: ScheduleConfig{
			Duration:   5 * time.Minute,
			Continuous: false,
		},
		Safety: SafetyConfig{
			AutoRollback: true,
			MaxDuration:  30 * time.Minute,
			AbortOnError: true,
			KillSwitch:   make(chan struct{}),
		},
	}
}

func generateID() string {
	return fmt.Sprintf("exp-%d", time.Now().UnixNano())
}
