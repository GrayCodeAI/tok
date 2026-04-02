package chaos

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine()

	if engine == nil {
		t.Fatal("expected engine to be created")
	}

	if engine.experiments == nil {
		t.Error("expected experiments map to be initialized")
	}

	if engine.handlers == nil {
		t.Error("expected handlers map to be initialized")
	}
}

func TestNewExperiment(t *testing.T) {
	exp := NewExperiment("test-experiment", TypeLatency)

	if exp == nil {
		t.Fatal("expected experiment to be created")
	}

	if exp.Name != "test-experiment" {
		t.Errorf("expected name 'test-experiment', got %s", exp.Name)
	}

	if exp.Type != TypeLatency {
		t.Errorf("expected type 'latency', got %s", exp.Type)
	}

	if exp.Config.Probability != 0.1 {
		t.Errorf("expected default probability 0.1, got %f", exp.Config.Probability)
	}
}

func TestRegisterExperiment(t *testing.T) {
	engine := NewEngine()
	exp := NewExperiment("test", TypeError)

	err := engine.RegisterExperiment(exp)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Registering duplicate should fail
	err = engine.RegisterExperiment(exp)
	if err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestGetExperiment(t *testing.T) {
	engine := NewEngine()
	exp := NewExperiment("test", TypeCPU)
	exp.ID = "exp-123"

	_ = engine.RegisterExperiment(exp)

	retrieved, err := engine.GetExperiment("exp-123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if retrieved.Name != "test" {
		t.Errorf("expected name 'test', got %s", retrieved.Name)
	}

	// Non-existent experiment
	_, err = engine.GetExperiment("non-existent")
	if err == nil {
		t.Error("expected error for non-existent experiment")
	}
}

func TestListExperiments(t *testing.T) {
	engine := NewEngine()

	// Create and register experiments with unique IDs
	for i := 0; i < 3; i++ {
		exp := NewExperiment(fmt.Sprintf("test-%d", i), TypeMemory)
		exp.ID = fmt.Sprintf("exp-%d", i) // Ensure unique IDs
		_ = engine.RegisterExperiment(exp)
		time.Sleep(time.Millisecond) // Small delay to avoid timestamp collisions
	}

	experiments := engine.ListExperiments()
	if len(experiments) != 3 {
		t.Errorf("expected 3 experiments, got %d", len(experiments))
	}
}

func TestRegisterHandler(t *testing.T) {
	engine := NewEngine()

	handler := func(ctx context.Context, config FaultConfig) error {
		return nil
	}

	engine.RegisterHandler(TypeLatency, handler)

	if _, exists := engine.handlers[TypeLatency]; !exists {
		t.Error("expected handler to be registered")
	}
}

func TestStandardFaultHandlers(t *testing.T) {
	handlers := StandardFaultHandlers()

	if len(handlers) == 0 {
		t.Error("expected standard handlers to be available")
	}

	// Test latency handler
	if handler, exists := handlers[TypeLatency]; !exists {
		t.Error("expected latency handler")
	} else {
		ctx := context.Background()
		config := FaultConfig{
			Duration:  100 * time.Millisecond,
			Intensity: 0.5,
		}
		start := time.Now()
		err := handler(ctx, config)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Should have added some delay
		if duration < 40*time.Millisecond {
			t.Errorf("expected delay, got %v", duration)
		}
	}

	// Test error handler
	if handler, exists := handlers[TypeError]; !exists {
		t.Error("expected error handler")
	} else {
		ctx := context.Background()
		config := FaultConfig{Target: "test-target"}
		err := handler(ctx, config)

		if err == nil {
			t.Error("expected error from error handler")
		}
	}
}

func TestExperimentStatus(t *testing.T) {
	exp := NewExperiment("test", TypeLatency)

	if exp.status.State != StatePending {
		t.Errorf("expected initial state 'pending', got %s", exp.status.State)
	}
}

func TestSafetyConfig(t *testing.T) {
	exp := NewExperiment("test", TypeLatency)

	if !exp.Safety.AutoRollback {
		t.Error("expected auto-rollback to be enabled by default")
	}

	if exp.Safety.MaxDuration != 30*time.Minute {
		t.Errorf("expected default max duration 30m, got %v", exp.Safety.MaxDuration)
	}
}
