package iteragent

import (
	"context"
	"testing"
	"time"
)

func TestNewAgent(t *testing.T) {
	config := AgentConfig{
		Name:          "TestAgent",
		MaxIterations: 5,
		Timeout:       30 * time.Second,
	}

	agent := NewAgent(config)
	if agent == nil {
		t.Fatal("expected agent to be created")
	}

	if agent.name != "TestAgent" {
		t.Errorf("expected name 'TestAgent', got %s", agent.name)
	}

	if agent.config.MaxIterations != 5 {
		t.Errorf("expected 5 max iterations, got %d", agent.config.MaxIterations)
	}

	if agent.state.Status != StatusIdle {
		t.Errorf("expected status 'idle', got %s", agent.state.Status)
	}
}

func TestAgentID(t *testing.T) {
	config := AgentConfig{Name: "Test"}
	agent := NewAgent(config)

	if agent.ID() == "" {
		t.Error("expected agent ID to be generated")
	}
}

func TestAgentName(t *testing.T) {
	config := AgentConfig{Name: "TestAgent"}
	agent := NewAgent(config)

	if agent.Name() != "TestAgent" {
		t.Errorf("expected 'TestAgent', got %s", agent.Name())
	}
}

func TestRegisterTool(t *testing.T) {
	config := AgentConfig{Name: "Test"}
	agent := NewAgent(config)

	tool := Tool{
		Name:        "test-tool",
		Description: "A test tool",
		Execute: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "result", nil
		},
	}

	agent.RegisterTool(tool)

	found, exists := agent.GetTool("test-tool")
	if !exists {
		t.Error("expected tool to be registered")
	}

	if found.Name != "test-tool" {
		t.Errorf("expected tool name 'test-tool', got %s", found.Name)
	}

	// Check non-existent tool
	_, exists = agent.GetTool("missing")
	if exists {
		t.Error("expected tool to not exist")
	}
}

func TestAgentExecute(t *testing.T) {
	config := AgentConfig{
		Name:          "Test",
		MaxIterations: 3,
		Timeout:       5 * time.Second,
	}
	agent := NewAgent(config)

	events := make([]Event, 0)
	agent.SetEventHandler(func(e Event) {
		events = append(events, e)
	})

	// Register a simple tool
	agent.RegisterTool(Tool{
		Name: "search",
		Execute: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "search result", nil
		},
	})

	ctx := context.Background()
	result, err := agent.Execute(ctx, "Test goal")
	if err != nil {
		t.Fatalf("failed to execute: %v", err)
	}

	if result == nil {
		t.Fatal("expected result")
	}

	if result.Success != true {
		t.Error("expected success")
	}

	if len(events) == 0 {
		t.Error("expected events to be emitted")
	}
}

func TestAgentExecuteContextCancelled(t *testing.T) {
	config := AgentConfig{
		Name:          "Test",
		MaxIterations: 100,
		Timeout:       30 * time.Second,
	}
	agent := NewAgent(config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := agent.Execute(ctx, "Test goal")
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	if result.Success {
		t.Error("expected failure")
	}
}

func TestAgentPauseResume(t *testing.T) {
	config := AgentConfig{Name: "Test"}
	agent := NewAgent(config)

	// Try to pause when not running
	err := agent.Pause()
	if err == nil {
		t.Error("expected error when pausing non-running agent")
	}

	// Try to resume when not paused
	err = agent.Resume(context.Background())
	if err == nil {
		t.Error("expected error when resuming non-paused agent")
	}
}

func TestAgentGetState(t *testing.T) {
	config := AgentConfig{Name: "Test"}
	agent := NewAgent(config)

	state := agent.GetState()
	if state.Status != StatusIdle {
		t.Errorf("expected status 'idle', got %s", state.Status)
	}
}

func TestAgentGetIterations(t *testing.T) {
	config := AgentConfig{Name: "Test"}
	agent := NewAgent(config)

	iterations := agent.GetIterations()
	if len(iterations) != 0 {
		t.Errorf("expected 0 iterations, got %d", len(iterations))
	}
}

func TestAgentClearMemory(t *testing.T) {
	config := AgentConfig{Name: "Test"}
	agent := NewAgent(config)

	// Add some memory
	agent.memory.ShortTerm = append(agent.memory.ShortTerm, MemoryEntry{
		Content: "test",
	})
	agent.memory.Working["key"] = "value"

	agent.ClearMemory()

	if len(agent.memory.ShortTerm) != 0 {
		t.Error("expected short-term memory to be cleared")
	}

	if len(agent.memory.Working) != 0 {
		t.Error("expected working memory to be cleared")
	}
}

func TestManagerCreateAgent(t *testing.T) {
	manager := NewManager()

	config := AgentConfig{Name: "Test"}
	agent := manager.CreateAgent(config)

	if agent == nil {
		t.Fatal("expected agent to be created")
	}

	found, err := manager.GetAgent(agent.ID())
	if err != nil {
		t.Fatalf("failed to get agent: %v", err)
	}

	if found.ID() != agent.ID() {
		t.Error("expected to find agent")
	}
}

func TestManagerListAgents(t *testing.T) {
	manager := NewManager()

	manager.CreateAgent(AgentConfig{Name: "Agent1"})
	time.Sleep(time.Millisecond)
	manager.CreateAgent(AgentConfig{Name: "Agent2"})

	agents := manager.ListAgents()

	if len(agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(agents))
	}
}

func TestManagerDeleteAgent(t *testing.T) {
	manager := NewManager()

	agent := manager.CreateAgent(AgentConfig{Name: "Test"})

	err := manager.DeleteAgent(agent.ID())
	if err != nil {
		t.Fatalf("failed to delete agent: %v", err)
	}

	_, err = manager.GetAgent(agent.ID())
	if err == nil {
		t.Error("expected error after deletion")
	}

	// Delete non-existent
	err = manager.DeleteAgent("missing")
	if err == nil {
		t.Error("expected error for non-existent agent")
	}
}

func TestManagerGetAgentNotFound(t *testing.T) {
	manager := NewManager()

	_, err := manager.GetAgent("missing")
	if err == nil {
		t.Error("expected error for non-existent agent")
	}
}

func TestAgentStatusTransitions(t *testing.T) {
	config := AgentConfig{Name: "Test"}
	agent := NewAgent(config)

	// Initial state
	if agent.GetState().Status != StatusIdle {
		t.Errorf("expected 'idle', got %s", agent.GetState().Status)
	}
}

func TestAgentConfig(t *testing.T) {
	config := AgentConfig{
		Name:             "Test",
		MaxIterations:    10,
		Timeout:          60 * time.Second,
		Temperature:      0.7,
		SystemPrompt:     "You are a test agent",
		EnableReflection: true,
		RequireApproval:  false,
	}

	agent := NewAgent(config)

	if agent.config.MaxIterations != 10 {
		t.Errorf("expected 10 iterations, got %d", agent.config.MaxIterations)
	}

	if agent.config.Temperature != 0.7 {
		t.Errorf("expected temperature 0.7, got %.2f", agent.config.Temperature)
	}

	if !agent.config.EnableReflection {
		t.Error("expected reflection enabled")
	}
}

func BenchmarkAgentExecute(b *testing.B) {
	config := AgentConfig{
		Name:          "Bench",
		MaxIterations: 1,
		Timeout:       5 * time.Second,
	}
	agent := NewAgent(config)

	agent.RegisterTool(Tool{
		Name: "noop",
		Execute: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	})

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent.Execute(ctx, "Test")
	}
}
