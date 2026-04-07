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

func (m *mockRepository) Save(ctx context.Context, record *RecordRequest) (int64, error) {
	m.records = append(m.records, record)
	return int64(len(m.records)), nil
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

func (m *mockRepository) GetTeamStats(ctx context.Context, teamID string) (*TeamStatsResponse, error) {
	return &TeamStatsResponse{}, nil
}

func (m *mockRepository) GetUserStats(ctx context.Context, teamID, userID string) (*UserStatsResponse, error) {
	return &UserStatsResponse{}, nil
}

func (m *mockRepository) GetFilterEffectiveness(ctx context.Context, teamID string, limit int) ([]FilterMetricResponse, error) {
	return nil, nil
}

func (m *mockRepository) GetTrendAnalysis(ctx context.Context, teamID string, granularity string, periods int) ([]TrendPoint, error) {
	return nil, nil
}

func (m *mockRepository) GetCostProjection(ctx context.Context, teamID string, daysToProject int) (*CostProjectionResponse, error) {
	return &CostProjectionResponse{}, nil
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

	_, err := svc.Record(ctx, record)
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
			TotalCommands:         100,
			TotalTokensSaved:      5000,
			AverageSavingsPercent: 50.0,
			P99LatencyMs:          25.0,
		},
	}
	svc := NewService(mock)
	ctx := context.Background()

	metrics, err := svc.GetMetrics(ctx, &MetricsRequest{TeamID: "test"})
	if err != nil {
		t.Fatalf("GetMetrics returned error: %v", err)
	}

	if metrics.TotalCommands != 100 {
		t.Errorf("Expected 100 commands, got %d", metrics.TotalCommands)
	}

	if metrics.AverageSavingsPercent != 50.0 {
		t.Errorf("Expected 50.0 savings, got %f", metrics.AverageSavingsPercent)
	}
}

func TestServiceGetEconomics(t *testing.T) {
	mock := &mockRepository{
		metrics: &MetricsResponse{
			TotalTokensSaved:    10000,
			EstimatedSavingsUSD: 50.0,
		},
	}
	svc := NewService(mock)
	ctx := context.Background()

	econ, err := svc.GetEconomics(ctx, "test-team")
	if err != nil {
		t.Fatalf("GetEconomics returned error: %v", err)
	}

	if econ.TokensSaved != 10000 {
		t.Errorf("Expected 10000 tokens saved, got %d", econ.TokensSaved)
	}
}
