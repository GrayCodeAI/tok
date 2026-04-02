package abtest

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	storage := NewInMemoryStorage()
	manager := NewManager(storage)

	if manager == nil {
		t.Fatal("expected manager to be created")
	}

	if manager.experiments == nil {
		t.Error("expected experiments map to be initialized")
	}
}

func TestCreateExperiment(t *testing.T) {
	storage := NewInMemoryStorage()
	manager := NewManager(storage)

	config := ExperimentConfig{
		Name:        "Test Experiment",
		Description: "A test experiment",
		Hypothesis:  "Variant B will perform better",
		Type:        TypeAB,
		Variants: []VariantConfig{
			{
				Name:           "Control",
				TrafficPercent: 50,
				IsControl:      true,
			},
			{
				Name:           "Treatment",
				TrafficPercent: 50,
				IsControl:      false,
			},
		},
		TrafficAllocation: 100,
		PrimaryMetric:     "conversion_rate",
		SampleSize:        1000,
		Duration:          7 * 24 * time.Hour,
		Randomization:     RandomizationRandom,
	}

	experiment, err := manager.CreateExperiment(config)
	if err != nil {
		t.Fatalf("failed to create experiment: %v", err)
	}

	if experiment == nil {
		t.Fatal("expected experiment to be created")
	}

	if experiment.Name != "Test Experiment" {
		t.Errorf("expected name 'Test Experiment', got %s", experiment.Name)
	}

	if len(experiment.Variants) != 2 {
		t.Errorf("expected 2 variants, got %d", len(experiment.Variants))
	}
}

func TestCreateExperimentValidation(t *testing.T) {
	storage := NewInMemoryStorage()
	manager := NewManager(storage)

	// Test missing name
	config := ExperimentConfig{
		Name: "",
	}
	_, err := manager.CreateExperiment(config)
	if err == nil {
		t.Error("expected error for missing name")
	}

	// Test insufficient variants
	config.Name = "Test"
	config.Variants = []VariantConfig{
		{Name: "OnlyOne", TrafficPercent: 100, IsControl: true},
	}
	_, err = manager.CreateExperiment(config)
	if err == nil {
		t.Error("expected error for insufficient variants")
	}

	// Test traffic percentage not summing to 100
	config.Variants = []VariantConfig{
		{Name: "Control", TrafficPercent: 30, IsControl: true},
		{Name: "Treatment", TrafficPercent: 30, IsControl: false},
	}
	_, err = manager.CreateExperiment(config)
	if err == nil {
		t.Error("expected error for traffic not summing to 100")
	}

	// Test no control variant
	config.Variants = []VariantConfig{
		{Name: "VariantA", TrafficPercent: 50, IsControl: false},
		{Name: "VariantB", TrafficPercent: 50, IsControl: false},
	}
	_, err = manager.CreateExperiment(config)
	if err == nil {
		t.Error("expected error for no control variant")
	}
}

func TestStartExperiment(t *testing.T) {
	storage := NewInMemoryStorage()
	manager := NewManager(storage)

	config := ExperimentConfig{
		Name: "Test",
		Variants: []VariantConfig{
			{Name: "Control", TrafficPercent: 50, IsControl: true},
			{Name: "Treatment", TrafficPercent: 50, IsControl: false},
		},
	}

	experiment, _ := manager.CreateExperiment(config)

	ctx := context.Background()
	err := manager.Start(ctx, experiment.ID)
	if err != nil {
		t.Errorf("failed to start experiment: %v", err)
	}

	if experiment.Status != StatusRunning {
		t.Errorf("expected status 'running', got %s", experiment.Status)
	}

	if experiment.StartTime.IsZero() {
		t.Error("expected start time to be set")
	}
}

func TestAssignVariant(t *testing.T) {
	storage := NewInMemoryStorage()
	manager := NewManager(storage)

	config := ExperimentConfig{
		Name:          "Test",
		Randomization: RandomizationUserID,
		Variants: []VariantConfig{
			{Name: "Control", TrafficPercent: 50, IsControl: true},
			{Name: "Treatment", TrafficPercent: 50, IsControl: false},
		},
	}

	experiment, _ := manager.CreateExperiment(config)

	ctx := context.Background()
	manager.Start(ctx, experiment.ID)

	// Assign variant to user
	variant, err := manager.AssignVariant(ctx, experiment.ID, "user-123")
	if err != nil {
		t.Fatalf("failed to assign variant: %v", err)
	}

	if variant == nil {
		t.Fatal("expected variant to be assigned")
	}

	// Same user should get same variant
	variant2, _ := manager.AssignVariant(ctx, experiment.ID, "user-123")
	if variant2.ID != variant.ID {
		t.Error("expected same variant for same user")
	}
}

func TestAssignVariantNotRunning(t *testing.T) {
	storage := NewInMemoryStorage()
	manager := NewManager(storage)

	config := ExperimentConfig{
		Name: "Test",
		Variants: []VariantConfig{
			{Name: "Control", TrafficPercent: 50, IsControl: true},
			{Name: "Treatment", TrafficPercent: 50, IsControl: false},
		},
	}

	experiment, _ := manager.CreateExperiment(config)

	ctx := context.Background()
	_, err := manager.AssignVariant(ctx, experiment.ID, "user-123")

	if err == nil {
		t.Error("expected error for experiment not running")
	}
}

func TestRecordEvent(t *testing.T) {
	storage := NewInMemoryStorage()
	manager := NewManager(storage)

	config := ExperimentConfig{
		Name: "Test",
		Variants: []VariantConfig{
			{Name: "Control", TrafficPercent: 50, IsControl: true},
			{Name: "Treatment", TrafficPercent: 50, IsControl: false},
		},
	}

	experiment, _ := manager.CreateExperiment(config)

	ctx := context.Background()
	manager.Start(ctx, experiment.ID)
	manager.AssignVariant(ctx, experiment.ID, "user-123")

	// Record conversion event
	err := manager.RecordEvent(ctx, experiment.ID, "user-123", "conversion", 1.0)
	if err != nil {
		t.Errorf("failed to record event: %v", err)
	}

	// Verify metrics were updated
	experiment.mu.RLock()
	defer experiment.mu.RUnlock()

	found := false
	for _, v := range experiment.Variants {
		if v.Metrics.Conversions > 0 {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected conversion to be recorded")
	}
}

func TestAnalyze(t *testing.T) {
	storage := NewInMemoryStorage()
	manager := NewManager(storage)

	config := ExperimentConfig{
		Name: "Test",
		Variants: []VariantConfig{
			{Name: "Control", TrafficPercent: 50, IsControl: true},
			{Name: "Treatment", TrafficPercent: 50, IsControl: false},
		},
	}

	experiment, _ := manager.CreateExperiment(config)

	ctx := context.Background()
	manager.Start(ctx, experiment.ID)

	// Assign users and record events
	for i := 0; i < 100; i++ {
		userID := fmt.Sprintf("user-%d", i)
		manager.AssignVariant(ctx, experiment.ID, userID)

		// Control: 50% conversion rate
		if i < 50 {
			manager.RecordEvent(ctx, experiment.ID, userID, "impression", 1)
			if i < 25 {
				manager.RecordEvent(ctx, experiment.ID, userID, "conversion", 1)
			}
		} else { // Treatment: 70% conversion rate
			manager.RecordEvent(ctx, experiment.ID, userID, "impression", 1)
			if i < 85 {
				manager.RecordEvent(ctx, experiment.ID, userID, "conversion", 1)
			}
		}
	}

	// Analyze results
	results, err := manager.Analyze(ctx, experiment.ID)
	if err != nil {
		t.Fatalf("failed to analyze: %v", err)
	}

	if results == nil {
		t.Fatal("expected results")
	}

	// Check that a winner was determined
	if results.Winner == nil {
		t.Error("expected a winner to be determined")
	}
}

func TestStop(t *testing.T) {
	storage := NewInMemoryStorage()
	manager := NewManager(storage)

	config := ExperimentConfig{
		Name: "Test",
		Variants: []VariantConfig{
			{Name: "Control", TrafficPercent: 50, IsControl: true},
			{Name: "Treatment", TrafficPercent: 50, IsControl: false},
		},
	}

	experiment, _ := manager.CreateExperiment(config)

	ctx := context.Background()
	manager.Start(ctx, experiment.ID)
	err := manager.Stop(ctx, experiment.ID)

	if err != nil {
		t.Errorf("failed to stop: %v", err)
	}

	if experiment.Status != StatusStopped {
		t.Errorf("expected status 'stopped', got %s", experiment.Status)
	}

	if experiment.EndTime.IsZero() {
		t.Error("expected end time to be set")
	}
}

func TestInMemoryStorage(t *testing.T) {
	storage := NewInMemoryStorage()

	// Test save
	experiment := &Experiment{
		ID:     "exp-123",
		Name:   "Test",
		Status: StatusRunning,
	}

	err := storage.Save(experiment)
	if err != nil {
		t.Errorf("failed to save: %v", err)
	}

	// Test load
	loaded, err := storage.Load("exp-123")
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if loaded.Name != "Test" {
		t.Errorf("expected name 'Test', got %s", loaded.Name)
	}

	// Test list
	list, err := storage.List()
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("expected 1 experiment, got %d", len(list))
	}
}

func TestCalculatePValue(t *testing.T) {
	// Test that p-value is calculated correctly
	z := 1.96 // 95% confidence
	p := calculatePValue(z)

	// For z=1.96, p should be approximately 0.05
	if p < 0.04 || p > 0.06 {
		t.Errorf("expected p-value around 0.05, got %.4f", p)
	}
}

func BenchmarkAssignVariant(b *testing.B) {
	storage := NewInMemoryStorage()
	manager := NewManager(storage)

	config := ExperimentConfig{
		Name: "Bench",
		Variants: []VariantConfig{
			{Name: "Control", TrafficPercent: 50, IsControl: true},
			{Name: "Treatment", TrafficPercent: 50, IsControl: false},
		},
	}

	experiment, _ := manager.CreateExperiment(config)
	manager.Start(context.Background(), experiment.ID)

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		userID := fmt.Sprintf("user-%d", i)
		manager.AssignVariant(ctx, experiment.ID, userID)
	}
}
