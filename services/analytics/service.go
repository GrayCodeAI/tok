// Package analytics provides the analytics service interface.
// This service handles token tracking, metrics, and economics calculations
// with enterprise-grade multi-tenant support.
package analytics

import (
	"context"
	"time"

	"github.com/GrayCodeAI/tokman/internal/metrics"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

// AnalyticsService defines the interface for the analytics service.
type AnalyticsService interface {
	// Record saves a command execution record.
	Record(ctx context.Context, req *RecordRequest) (int64, error)

	// GetMetrics returns aggregated metrics.
	GetMetrics(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error)

	// GetEconomics returns cost analysis.
	GetEconomics(ctx context.Context, teamID string) (*EconomicsResponse, error)

	// GetTopCommands returns the most used commands.
	GetTopCommands(ctx context.Context, limit int) ([]CommandStats, error)

	// GetTeamStats returns aggregated stats for a team.
	GetTeamStats(ctx context.Context, teamID string) (*TeamStatsResponse, error)

	// GetUserStats returns stats for an individual user.
	GetUserStats(ctx context.Context, teamID, userID string) (*UserStatsResponse, error)

	// GetFilterEffectiveness ranks filters by effectiveness.
	GetFilterEffectiveness(ctx context.Context, teamID string, limit int) ([]FilterMetricResponse, error)

	// GetTrendAnalysis returns historical trend data.
	GetTrendAnalysis(ctx context.Context, teamID string, granularity string, periods int) ([]TrendPoint, error)

	// GetCostProjection projects monthly costs.
	GetCostProjection(ctx context.Context, teamID string, daysToProject int) (*CostProjectionResponse, error)

	// GetDashboardData returns combined data for dashboard rendering.
	GetDashboardData(ctx context.Context, teamID string) (*DashboardResponse, error)
}

// RecordRequest is the input for recording a command.
type RecordRequest struct {
	TeamID         string        // Team identifier
	UserID         string        // User identifier
	Command        string        // Command executed
	OriginalTokens int           // Tokens before compression
	FilteredTokens int           // Tokens after compression
	Duration       time.Duration // Execution duration
	ExitCode       int           // Command exit code
	FilterMode     string        // Filter mode used
	SessionID      string        // Session identifier
	AgentName      string        // AI agent name
	ModelName      string        // Model name
	Provider       string        // Provider name
}

// MetricsRequest filters metrics by time range.
type MetricsRequest struct {
	TeamID    string
	UserID    string
	StartTime time.Time
	EndTime   time.Time
	SessionID string
}

// MetricsResponse contains aggregated metrics.
type MetricsResponse struct {
	TotalCommands        int64
	TotalTokensSaved     int64
	TotalTokensIn        int64
	TotalTokensOut       int64
	AverageSavingsPercent float64
	P50LatencyMs         float64
	P99LatencyMs         float64
	TopFilters           []FilterUsage
	EstimatedSavingsUSD  float64
	HourlyDistribution   []HourlyCount
}

// EconomicsResponse contains cost analysis.
type EconomicsResponse struct {
	TokensSaved              int64
	EstimatedCostSaved      float64
	EstimatedCostWithoutTok float64
	ROIPercent              float64
	MonthlyCommands         int
	AvgTokensPerCommand     float64
	PrimaryModel            string
	CostPerMTok             float64
	QuotaUsedPercent        float64
}

// CommandStats represents statistics for a command.
type CommandStats struct {
	Command         string
	Count           int64
	AvgSavings      float64
	TotalSaved      int64
	LastUsed        time.Time
	AverageSavingsPct float64
}

// FilterUsage represents filter usage statistics.
type FilterUsage struct {
	FilterName           string
	Count                int64
	AvgSavingPercent     float64
	TotalTokensSaved     int64
	EffectivenessScore   float64
}

// FilterMetricResponse represents filter performance metrics.
type FilterMetricResponse struct {
	FilterName             string
	UsageCount             int64
	AvgEffectiveness       float64
	TotalTokensSaved       int64
	AvgProcessingTimeMs    float64
	FirstUsed              time.Time
	LastUsed               time.Time
}

// HourlyCount represents hourly distribution.
type HourlyCount struct {
	Hour  int
	Count int64
}

// TeamStatsResponse contains team-level statistics.
type TeamStatsResponse struct {
	TeamID              string
	TeamName            string
	MemberCount         int
	TotalCommands       int64
	TotalTokensSaved    int64
	AvgReductionPercent float64
	MonthlyTokenBudget  int
	TokensUsedThisMonth int
	EstimatedMonthlyCost float64
}

// UserStatsResponse contains user-level statistics.
type UserStatsResponse struct {
	UserID              string
	Email               string
	TotalCommands       int64
	TotalTokensSaved    int64
	AvgReductionPercent float64
	LastCommand         time.Time
	MostUsedFilter      string
}

// TrendPoint represents a single trend data point.
type TrendPoint struct {
	Timestamp              time.Time
	TotalCommands          int64
	TotalOriginalTokens    int64
	TotalFilteredTokens    int64
	TotalSavedTokens       int64
	AvgReductionPercent    float64
	EstimatedCostUSD       float64
	EstimatedSavingsUSD    float64
}

// CostProjectionResponse contains cost projections.
type CostProjectionResponse struct {
	ProjectedMonthlyCost    float64
	ProjectedMonthlySavings float64
	ActualMonthlyCost       float64
	ActualMonthlySavings    float64
	ProjectionBasis         string
}

// DashboardResponse contains all data for dashboard rendering.
type DashboardResponse struct {
	TeamStats       *TeamStatsResponse
	Economics       *EconomicsResponse
	Trends          []TrendPoint
	TopFilters      []FilterMetricResponse
	UserStats       []UserStatsResponse
	CostProjection  *CostProjectionResponse
}

// Service implements AnalyticsService.
type Service struct {
	db AnalyticsRepository
	tracker *tracking.Tracker
}

// AnalyticsRepository defines the storage interface.
type AnalyticsRepository interface {
	Save(ctx context.Context, record *RecordRequest) (int64, error)
	Query(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error)
	GetTopCommands(ctx context.Context, limit int) ([]CommandStats, error)
	GetTeamStats(ctx context.Context, teamID string) (*TeamStatsResponse, error)
	GetUserStats(ctx context.Context, teamID, userID string) (*UserStatsResponse, error)
	GetFilterEffectiveness(ctx context.Context, teamID string, limit int) ([]FilterMetricResponse, error)
	GetTrendAnalysis(ctx context.Context, teamID string, granularity string, periods int) ([]TrendPoint, error)
	GetCostProjection(ctx context.Context, teamID string, daysToProject int) (*CostProjectionResponse, error)
}

// NewService creates a new analytics service.
func NewService(db AnalyticsRepository) *Service {
	return &Service{db: db}
}

// Record implements AnalyticsService.
func (s *Service) Record(ctx context.Context, req *RecordRequest) (int64, error) {
	// Record Prometheus metrics for tokens saved
	savedTokens := req.OriginalTokens - req.FilteredTokens
	if savedTokens > 0 {
		metrics.TokensSavedTotal.Add(float64(savedTokens))
	}
	metrics.TokensProcessedTotal.Add(float64(req.OriginalTokens))

	recordID, err := s.db.Save(ctx, req)
	return recordID, err
}

// GetMetrics implements AnalyticsService.
func (s *Service) GetMetrics(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error) {
	return s.db.Query(ctx, req)
}

// GetEconomics implements AnalyticsService.
func (s *Service) GetEconomics(ctx context.Context, teamID string) (*EconomicsResponse, error) {
	metrics, err := s.db.Query(ctx, &MetricsRequest{TeamID: teamID})
	if err != nil {
		return nil, err
	}

	// Default pricing: Claude 3.5 Sonnet
	costPerMTok := 3.0 // $3 per 1M input tokens

	totalCost := float64(metrics.TotalTokensIn) * costPerMTok / 1_000_000
	savingsCost := float64(metrics.TotalTokensSaved) * costPerMTok / 1_000_000

	costWithout := totalCost + savingsCost
	roi := 0.0
	if costWithout > 0 {
		roi = (savingsCost / costWithout) * 100
	}

	return &EconomicsResponse{
		TokensSaved:              metrics.TotalTokensSaved,
		EstimatedCostSaved:       savingsCost,
		EstimatedCostWithoutTok:  costWithout,
		ROIPercent:               roi,
		MonthlyCommands:          int(metrics.TotalCommands),
		AvgTokensPerCommand:      float64(metrics.TotalTokensIn) / float64(metrics.TotalCommands),
		PrimaryModel:             "claude-3.5-sonnet",
		CostPerMTok:              costPerMTok,
		QuotaUsedPercent:         0,
	}, nil
}

// GetTopCommands implements AnalyticsService.
func (s *Service) GetTopCommands(ctx context.Context, limit int) ([]CommandStats, error) {
	return s.db.GetTopCommands(ctx, limit)
}

// GetTeamStats implements AnalyticsService.
func (s *Service) GetTeamStats(ctx context.Context, teamID string) (*TeamStatsResponse, error) {
	return s.db.GetTeamStats(ctx, teamID)
}

// GetUserStats implements AnalyticsService.
func (s *Service) GetUserStats(ctx context.Context, teamID, userID string) (*UserStatsResponse, error) {
	return s.db.GetUserStats(ctx, teamID, userID)
}

// GetFilterEffectiveness implements AnalyticsService.
func (s *Service) GetFilterEffectiveness(ctx context.Context, teamID string, limit int) ([]FilterMetricResponse, error) {
	return s.db.GetFilterEffectiveness(ctx, teamID, limit)
}

// GetTrendAnalysis implements AnalyticsService.
func (s *Service) GetTrendAnalysis(ctx context.Context, teamID string, granularity string, periods int) ([]TrendPoint, error) {
	return s.db.GetTrendAnalysis(ctx, teamID, granularity, periods)
}

// GetCostProjection implements AnalyticsService.
func (s *Service) GetCostProjection(ctx context.Context, teamID string, daysToProject int) (*CostProjectionResponse, error) {
	return s.db.GetCostProjection(ctx, teamID, daysToProject)
}

// GetDashboardData implements AnalyticsService.
func (s *Service) GetDashboardData(ctx context.Context, teamID string) (*DashboardResponse, error) {
	teamStats, err := s.GetTeamStats(ctx, teamID)
	if err != nil {
		return nil, err
	}

	economics, err := s.GetEconomics(ctx, teamID)
	if err != nil {
		return nil, err
	}

	trends, err := s.GetTrendAnalysis(ctx, teamID, "daily", 30)
	if err != nil {
		return nil, err
	}

	topFilters, err := s.GetFilterEffectiveness(ctx, teamID, 10)
	if err != nil {
		return nil, err
	}

	projection, err := s.GetCostProjection(ctx, teamID, 30)
	if err != nil {
		return nil, err
	}

	return &DashboardResponse{
		TeamStats:      teamStats,
		Economics:      economics,
		Trends:         trends,
		TopFilters:     topFilters,
		CostProjection: projection,
	}, nil
}
