package health_test

import (
	"context"
	"testing"

	"github.com/GrayCodeAI/tok/internal/config"
	"github.com/GrayCodeAI/tok/internal/health"
)

func TestChecker_Creation(t *testing.T) {
	cfg := config.Defaults()
	checker := health.NewChecker(cfg, nil, "test-version")

	if checker == nil {
		t.Fatal("expected checker to not be nil")
	}
}

func TestChecker_Check(t *testing.T) {
	cfg := config.Defaults()
	checker := health.NewChecker(cfg, nil, "test-version")

	check := checker.Check(context.Background())

	if check.Status == "" {
		t.Error("expected status to be set")
	}
	if check.Version != "test-version" {
		t.Errorf("expected version test-version, got %s", check.Version)
	}
	if len(check.Components) == 0 {
		t.Error("expected components to be set")
	}
}

func TestChecker_Liveness(t *testing.T) {
	cfg := config.Defaults()
	checker := health.NewChecker(cfg, nil, "test-version")

	comp := checker.CheckLiveness(context.Background())

	if comp.Name != "liveness" {
		t.Errorf("expected name liveness, got %s", comp.Name)
	}
	if comp.Status != health.StatusHealthy {
		t.Errorf("expected healthy status, got %s", comp.Status)
	}
}

func TestChecker_Readiness(t *testing.T) {
	cfg := config.Defaults()
	checker := health.NewChecker(cfg, nil, "test-version")

	comp := checker.CheckReadiness(context.Background())

	if comp.Name != "readiness" {
		t.Errorf("expected name readiness, got %s", comp.Name)
	}
}

func TestStatus_Values(t *testing.T) {
	// Just verify the constants exist and are different
	if health.StatusHealthy == health.StatusDegraded {
		t.Error("expected statuses to be different")
	}
	if health.StatusDegraded == health.StatusUnhealthy {
		t.Error("expected statuses to be different")
	}
}
