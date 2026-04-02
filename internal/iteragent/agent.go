// Package iteragent provides iterative agent framework integration for TokMan
package iteragent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Agent represents an iterative agent capable of multi-step reasoning
type Agent struct {
	id          string
	name        string
	config      AgentConfig
	state       AgentState
	memory      Memory
	tools       map[string]Tool
	iterations  []Iteration
	currentStep int
	mu          sync.RWMutex
	handler     EventHandler
}

// AgentConfig holds agent configuration
type AgentConfig struct {
	Name             string
	Description      string
	MaxIterations    int
	Timeout          time.Duration
	Temperature      float64
	SystemPrompt     string
	AllowedTools     []string
	RequireApproval  bool
	EnableReflection bool
}

// AgentState represents the agent's current state
type AgentState struct {
	Status      AgentStatus
	CurrentGoal string
	Context     map[string]interface{}
	Variables   map[string]string
	StartTime   time.Time
	EndTime     *time.Time
}

// AgentStatus represents agent status
type AgentStatus string

const (
	StatusIdle      AgentStatus = "idle"
	StatusRunning   AgentStatus = "running"
	StatusPaused    AgentStatus = "paused"
	StatusCompleted AgentStatus = "completed"
	StatusFailed    AgentStatus = "failed"
	StatusAborted   AgentStatus = "aborted"
)

// Memory provides agent memory capabilities
type Memory struct {
	ShortTerm []MemoryEntry
	LongTerm  []MemoryEntry
	Working   map[string]interface{}
	mu        sync.RWMutex
}

// MemoryEntry represents a memory entry
type MemoryEntry struct {
	Timestamp  time.Time
	Type       string
	Content    string
	Importance float64
	Metadata   map[string]string
}

// Tool represents a tool the agent can use
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]Parameter
	Execute     func(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// Parameter represents a tool parameter
type Parameter struct {
	Type        string
	Description string
	Required    bool
	Default     interface{}
}

// Iteration represents a single iteration of the agent
type Iteration struct {
	Number      int
	Timestamp   time.Time
	Thought     string
	Action      Action
	Observation string
	Reflection  string
	Duration    time.Duration
}

// Action represents an action taken by the agent
type Action struct {
	Type       string
	Tool       string
	Parameters map[string]interface{}
	Result     interface{}
	Error      error
}

// EventHandler handles agent events
type EventHandler func(event Event)

// Event represents an agent event
type Event struct {
	Type      EventType
	Timestamp time.Time
	AgentID   string
	Iteration int
	Message   string
	Data      interface{}
}

// EventType represents event types
type EventType string

const (
	EventStarted     EventType = "started"
	EventIteration   EventType = "iteration"
	EventThinking    EventType = "thinking"
	EventAction      EventType = "action"
	EventObservation EventType = "observation"
	EventReflection  EventType = "reflection"
	EventCompleted   EventType = "completed"
	EventFailed      EventType = "failed"
	EventAborted     EventType = "aborted"
)

// Result holds the final result of agent execution
type Result struct {
	Success    bool
	Output     string
	Iterations int
	Duration   time.Duration
	ToolsUsed  []string
	TokenUsage TokenUsage
	FinalState map[string]interface{}
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	Input  int
	Output int
	Total  int
}

// NewAgent creates a new iterative agent
func NewAgent(config AgentConfig) *Agent {
	return &Agent{
		id:     generateAgentID(),
		name:   config.Name,
		config: config,
		state: AgentState{
			Status:    StatusIdle,
			Context:   make(map[string]interface{}),
			Variables: make(map[string]string),
		},
		memory: Memory{
			ShortTerm: make([]MemoryEntry, 0),
			LongTerm:  make([]MemoryEntry, 0),
			Working:   make(map[string]interface{}),
		},
		tools:      make(map[string]Tool),
		iterations: make([]Iteration, 0),
	}
}

// ID returns the agent ID
func (a *Agent) ID() string {
	return a.id
}

// Name returns the agent name
func (a *Agent) Name() string {
	return a.name
}

// SetEventHandler sets the event handler
func (a *Agent) SetEventHandler(handler EventHandler) {
	a.handler = handler
}

func (a *Agent) emit(event Event) {
	if a.handler != nil {
		a.handler(event)
	}
}

// RegisterTool registers a tool for the agent
func (a *Agent) RegisterTool(tool Tool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tools[tool.Name] = tool
}

// GetTool returns a tool by name
func (a *Agent) GetTool(name string) (Tool, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	tool, exists := a.tools[name]
	return tool, exists
}

// Execute runs the agent on a goal
func (a *Agent) Execute(ctx context.Context, goal string) (*Result, error) {
	a.mu.Lock()
	if a.state.Status == StatusRunning {
		a.mu.Unlock()
		return nil, fmt.Errorf("agent is already running")
	}

	a.state.Status = StatusRunning
	a.state.CurrentGoal = goal
	a.state.StartTime = time.Now()
	a.iterations = make([]Iteration, 0)
	a.currentStep = 0
	a.mu.Unlock()

	a.emit(Event{
		Type:      EventStarted,
		Timestamp: time.Now(),
		AgentID:   a.id,
		Message:   fmt.Sprintf("Starting execution with goal: %s", goal),
	})

	// Main execution loop
	result := &Result{
		Success:    true,
		ToolsUsed:  make([]string, 0),
		FinalState: make(map[string]interface{}),
	}

	startTime := time.Now()

	for a.currentStep < a.config.MaxIterations {
		select {
		case <-ctx.Done():
			a.abort()
			result.Success = false
			result.Output = "Execution aborted: context cancelled"
			return result, ctx.Err()

		default:
			iteration, err := a.executeIteration(ctx, goal)
			if err != nil {
				a.fail(err)
				result.Success = false
				result.Output = fmt.Sprintf("Execution failed: %v", err)
				break
			}

			a.mu.Lock()
			a.iterations = append(a.iterations, *iteration)
			a.currentStep++
			a.mu.Unlock()

			// Check if goal is achieved
			if a.isGoalAchieved(iteration) {
				break
			}

			// Check for completion conditions
			if iteration.Action.Type == "complete" {
				break
			}
		}
	}

	duration := time.Since(startTime)

	a.mu.Lock()
	defer a.mu.Unlock()

	if result.Success {
		a.state.Status = StatusCompleted
		result.Output = a.generateFinalOutput()
	}

	result.Iterations = len(a.iterations)
	result.Duration = duration
	result.TokenUsage = a.calculateTokenUsage()

	endTime := time.Now()
	a.state.EndTime = &endTime

	a.emit(Event{
		Type:      EventCompleted,
		Timestamp: time.Now(),
		AgentID:   a.id,
		Message:   fmt.Sprintf("Execution completed in %d iterations", result.Iterations),
		Data:      result,
	})

	return result, nil
}

func (a *Agent) executeIteration(ctx context.Context, goal string) (*Iteration, error) {
	start := time.Now()
	iteration := &Iteration{
		Number:    a.currentStep + 1,
		Timestamp: start,
	}

	// Think step
	thought, err := a.think(ctx, goal)
	if err != nil {
		return nil, err
	}
	iteration.Thought = thought

	a.emit(Event{
		Type:      EventThinking,
		Timestamp: time.Now(),
		AgentID:   a.id,
		Iteration: iteration.Number,
		Message:   thought,
	})

	// Decide action
	action, err := a.decideAction(ctx, thought)
	if err != nil {
		return nil, err
	}
	iteration.Action = *action

	a.emit(Event{
		Type:      EventAction,
		Timestamp: time.Now(),
		AgentID:   a.id,
		Iteration: iteration.Number,
		Message:   fmt.Sprintf("Executing action: %s", action.Type),
		Data:      action,
	})

	// Execute action
	observation, err := a.executeAction(ctx, action)
	if err != nil {
		action.Error = err
	}
	iteration.Observation = observation

	a.emit(Event{
		Type:      EventObservation,
		Timestamp: time.Now(),
		AgentID:   a.id,
		Iteration: iteration.Number,
		Message:   observation,
	})

	// Reflection step (if enabled)
	if a.config.EnableReflection {
		reflection, _ := a.reflect(ctx, thought, action, observation)
		iteration.Reflection = reflection

		a.emit(Event{
			Type:      EventReflection,
			Timestamp: time.Now(),
			AgentID:   a.id,
			Iteration: iteration.Number,
			Message:   reflection,
		})
	}

	iteration.Duration = time.Since(start)

	// Store in memory
	a.memory.mu.Lock()
	a.memory.ShortTerm = append(a.memory.ShortTerm, MemoryEntry{
		Timestamp:  time.Now(),
		Type:       "iteration",
		Content:    fmt.Sprintf("Thought: %s, Action: %s, Observation: %s", thought, action.Type, observation),
		Importance: 0.7,
	})
	a.memory.mu.Unlock()

	return iteration, nil
}

func (a *Agent) think(ctx context.Context, goal string) (string, error) {
	// This would integrate with an LLM for reasoning
	// Simplified implementation
	if len(a.iterations) == 0 {
		return fmt.Sprintf("I need to understand and work towards the goal: %s", goal), nil
	}

	lastIteration := a.iterations[len(a.iterations)-1]
	return fmt.Sprintf("Based on the previous observation: %s, I should continue working toward the goal.",
		lastIteration.Observation), nil
}

func (a *Agent) decideAction(ctx context.Context, thought string) (*Action, error) {
	// This would use LLM to decide next action
	// Simplified implementation - returns a complete action after 3 iterations
	if len(a.iterations) >= 3 {
		return &Action{
			Type:   "complete",
			Result: "Task completed successfully",
		}, nil
	}

	// Otherwise, use a tool
	return &Action{
		Type:       "tool",
		Tool:       "search",
		Parameters: map[string]interface{}{"query": thought},
	}, nil
}

func (a *Agent) executeAction(ctx context.Context, action *Action) (string, error) {
	switch action.Type {
	case "tool":
		tool, exists := a.tools[action.Tool]
		if !exists {
			return "", fmt.Errorf("tool %s not found", action.Tool)
		}

		result, err := tool.Execute(ctx, action.Parameters)
		if err != nil {
			return "", err
		}

		action.Result = result
		return fmt.Sprintf("Tool %s executed successfully with result: %v", action.Tool, result), nil

	case "complete":
		return "Task marked as complete", nil

	default:
		return "", fmt.Errorf("unknown action type: %s", action.Type)
	}
}

func (a *Agent) reflect(ctx context.Context, thought string, action *Action, observation string) (string, error) {
	// Reflection on the iteration
	return fmt.Sprintf("Reflection: The action %s was appropriate for the thought '%s'. Observation confirms progress.",
		action.Type, thought), nil
}

func (a *Agent) isGoalAchieved(iteration *Iteration) bool {
	// Check if goal is achieved
	return iteration.Action.Type == "complete"
}

func (a *Agent) generateFinalOutput() string {
	if len(a.iterations) == 0 {
		return "No iterations completed"
	}

	lastIteration := a.iterations[len(a.iterations)-1]
	if lastIteration.Action.Type == "complete" {
		if result, ok := lastIteration.Action.Result.(string); ok {
			return result
		}
	}

	return "Execution completed"
}

func (a *Agent) calculateTokenUsage() TokenUsage {
	// Simplified token calculation
	totalTokens := 0
	for _, iter := range a.iterations {
		totalTokens += len(iter.Thought) + len(iter.Observation)
	}

	return TokenUsage{
		Input:  totalTokens / 2,
		Output: totalTokens / 2,
		Total:  totalTokens,
	}
}

func (a *Agent) abort() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.Status = StatusAborted

	endTime := time.Now()
	a.state.EndTime = &endTime

	a.emit(Event{
		Type:      EventAborted,
		Timestamp: time.Now(),
		AgentID:   a.id,
		Message:   "Execution aborted",
	})
}

func (a *Agent) fail(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.Status = StatusFailed

	endTime := time.Now()
	a.state.EndTime = &endTime

	a.emit(Event{
		Type:      EventFailed,
		Timestamp: time.Now(),
		AgentID:   a.id,
		Message:   err.Error(),
	})
}

func (a *Agent) Pause() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.state.Status != StatusRunning {
		return fmt.Errorf("agent is not running")
	}

	a.state.Status = StatusPaused
	return nil
}

// Resume resumes agent execution
func (a *Agent) Resume(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.state.Status != StatusPaused {
		return fmt.Errorf("agent is not paused")
	}

	a.state.Status = StatusRunning
	return nil
}

// GetState returns current agent state
func (a *Agent) GetState() AgentState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// GetIterations returns all iterations
func (a *Agent) GetIterations() []Iteration {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]Iteration, len(a.iterations))
	copy(result, a.iterations)
	return result
}

// ClearMemory clears agent memory
func (a *Agent) ClearMemory() {
	a.memory.mu.Lock()
	defer a.memory.mu.Unlock()

	a.memory.ShortTerm = make([]MemoryEntry, 0)
	a.memory.Working = make(map[string]interface{})
}

// Manager manages multiple agents
type Manager struct {
	agents map[string]*Agent
	mu     sync.RWMutex
}

// NewManager creates an agent manager
func NewManager() *Manager {
	return &Manager{
		agents: make(map[string]*Agent),
	}
}

// CreateAgent creates a new agent
func (m *Manager) CreateAgent(config AgentConfig) *Agent {
	agent := NewAgent(config)

	m.mu.Lock()
	m.agents[agent.ID()] = agent
	m.mu.Unlock()

	return agent
}

// GetAgent returns an agent by ID
func (m *Manager) GetAgent(id string) (*Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, exists := m.agents[id]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", id)
	}

	return agent, nil
}

// ListAgents returns all agents
func (m *Manager) ListAgents() []*Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		result = append(result, agent)
	}

	return result
}

// DeleteAgent deletes an agent
func (m *Manager) DeleteAgent(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.agents[id]; !exists {
		return fmt.Errorf("agent %s not found", id)
	}

	delete(m.agents, id)
	return nil
}

func generateAgentID() string {
	return fmt.Sprintf("agent-%d", time.Now().UnixNano())
}
