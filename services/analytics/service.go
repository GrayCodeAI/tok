// Package analytics provides the analytics service interface.
// This service handles token tracking, metrics, and economics calculations.
package analytics

import (
	"context"
	"time"

	"github.com/GrayCodeAI/tokman/internal/metrics"
)

// AnalyticsService defines the interface for the analytics service.
type AnalyticsService interface {
	// Record saves a command execution record.
	Record(ctx context.Context, req *RecordRequest) error

	// GetMetrics returns aggregated metrics.
	GetMetrics(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error)

	// GetEconomics returns cost analysis.
	GetEconomics(ctx context.Context) (*EconomicsResponse, error)

	// GetTopCommands returns the most used commands.
	GetTopCommands(ctx context.Context, limit int) ([]CommandStats, error)
}

// RecordRequest is the input for recording a command.
type RecordRequest struct {
	Command        string        // Command executed
	OriginalTokens int           // Tokens before compression
	FilteredTokens int           // Tokens after compression
	Duration       time.Duration // Execution duration
	ExitCode       int           // Command exit code
	FilterMode     string        // Filter mode used
	SessionID      string        // Session identifier
}

// MetricsRequest filters metrics by time range.
type MetricsRequest struct {
	StartTime time.Time
	EndTime   time.Time
	SessionID string
}

// MetricsResponse contains aggregated metrics.
type MetricsResponse struct {
	TotalCommands      int64
	TotalTokensSaved   int64
	TotalTokensIn      int64
	TotalTokensOut     int64
	AverageSavings     float64
	P50LatencyMs       float64
	P99LatencyMs       float64
	TopFilters         []FilterUsage
	HourlyDistribution []HourlyCount
}

// EconomicsResponse contains cost analysis.
type EconomicsResponse struct {
	TokensSaved        int64
	EstimatedCostSaved float64
	ModelUsed          string
	CostPerToken       float64
	QuotaUsed          float64
	QuotaRemaining     float64
}

// CommandStats represents statistics for a command.
type CommandStats struct {
	Command    string
	Count      int64
	AvgSavings float64
	TotalSaved int64
	LastUsed   time.Time
}

// FilterUsage represents filter usage statistics.
type FilterUsage struct {
	FilterName string
	Count      int64
	AvgSavings float64
}

// HourlyCount represents hourly distribution.
type HourlyCount struct {
	Hour  int
	Count int64
}

// Service implements AnalyticsService.
type Service struct {
	db AnalyticsRepository
}

// AnalyticsRepository defines the storage interface.
type AnalyticsRepository interface {
	Save(ctx context.Context, record *RecordRequest) error
	Query(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error)
	GetTopCommands(ctx context.Context, limit int) ([]CommandStats, error)
}

// NewService creates a new analytics service.
func NewService(db AnalyticsRepository) *Service {
	return &Service{db: db}
}

// Record implements AnalyticsService.
func (s *Service) Record(ctx context.Context, req *RecordRequest) error {
	// Record Prometheus metrics for tokens saved
	savedTokens := req.OriginalTokens - req.FilteredTokens
	if savedTokens > 0 {
		metrics.TokensSavedTotal.Add(float64(savedTokens))
	}
	metrics.TokensProcessedTotal.Add(float64(req.OriginalTokens))

	return s.db.Save(ctx, req)
}

// GetMetrics implements AnalyticsService.
func (s *Service) GetMetrics(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error) {
	return s.db.Query(ctx, req)
}

// GetEconomics implements AnalyticsService.
func (s *Service) GetEconomics(ctx context.Context) (*EconomicsResponse, error) {
	metrics, err := s.db.Query(ctx, &MetricsRequest{})
	if err != nil {
		return nil, err
	}

	// Default pricing: Claude 3.5 Sonnet
	costPerToken := 0.003 / 1000.0 // $0.003 per 1K input tokens

	return &EconomicsResponse{
		TokensSaved:        metrics.TotalTokensSaved,
		EstimatedCostSaved: float64(metrics.TotalTokensSaved) * costPerToken,
		ModelUsed:          "claude-3.5-sonnet",
		CostPerToken:       costPerToken,
	}, nil
}

// GetTopCommands implements AnalyticsService.
func (s *Service) GetTopCommands(ctx context.Context, limit int) ([]CommandStats, error) {
	return s.db.GetTopCommands(ctx, limit)
}
