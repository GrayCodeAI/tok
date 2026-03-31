package analytics

import (
	"context"
	"testing"
	"time"
)

// mockRepository implements AnalyticsRepository for testing
type mockRepository struct {
	records     []*RecordRequest
	metrics     *MetricsResponse
	topCommands []CommandStats
}

func (m *mockRepository) Save(ctx context.Context, record *RecordRequest) error {
	m.records = append(m.records, record)
	return nil
}

func (m *mockRepository) Query(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error) {
	if m.metrics != nil {
		return m.metrics, nil
	}
	return &MetricsResponse{}, nil
}

func (m *mockRepository) GetTopCommands(ctx context.Context, limit int) ([]CommandStats, error) {
	return m.topCommands, nil
}

func TestServiceRecord(t *testing.T) {
	mock := &mockRepository{}
	svc := NewService(mock)
	ctx := context.Background()

	record := &RecordRequest{
		Command:        "git status",
		OriginalTokens: 1000,
		FilteredTokens: 400,
		Duration:       50 * time.Millisecond,
		ExitCode:       0,
		FilterMode:     "minimal",
		SessionID:      "test-session",
	}

	err := svc.Record(ctx, record)
	if err != nil {
		t.Fatalf("Record returned error: %v", err)
	}

	if len(mock.records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(mock.records))
	}

	if mock.records[0].Command != "git status" {
		t.Errorf("Expected command 'git status', got '%s'", mock.records[0].Command)
	}
}

func TestServiceGetMetrics(t *testing.T) {
	mock := &mockRepository{
		metrics: &MetricsResponse{
			TotalCommands:    100,
			TotalTokensSaved: 50000,
			AverageSavings:   45.5,
		},
	}
	svc := NewService(mock)
	ctx := context.Background()

	metrics, err := svc.GetMetrics(ctx, &MetricsRequest{})
	if err != nil {
		t.Fatalf("GetMetrics returned error: %v", err)
	}

	if metrics.TotalCommands != 100 {
		t.Errorf("Expected 100 commands, got %d", metrics.TotalCommands)
	}
}

func TestServiceGetEconomics(t *testing.T) {
	mock := &mockRepository{
		metrics: &MetricsResponse{
			TotalTokensSaved: 100000,
		},
	}
	svc := NewService(mock)
	ctx := context.Background()

	econ, err := svc.GetEconomics(ctx)
	if err != nil {
		t.Fatalf("GetEconomics returned error: %v", err)
	}

	if econ == nil {
		t.Fatal("GetEconomics returned nil")
	}

	if econ.TokensSaved != 100000 {
		t.Errorf("Expected 100000 tokens saved, got %d", econ.TokensSaved)
	}

	// Cost should be calculated
	if econ.EstimatedCostSaved <= 0 {
		t.Error("Expected positive cost savings")
	}
}

func TestServiceGetTopCommands(t *testing.T) {
	mock := &mockRepository{
		topCommands: []CommandStats{
			{Command: "git status", Count: 50, TotalSaved: 10000},
			{Command: "npm install", Count: 30, TotalSaved: 5000},
		},
	}
	svc := NewService(mock)
	ctx := context.Background()

	commands, err := svc.GetTopCommands(ctx, 10)
	if err != nil {
		t.Fatalf("GetTopCommands returned error: %v", err)
	}

	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}
}
