// Package grpc provides gRPC server implementation for the analytics service.
package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/GrayCodeAI/tokman/internal/metrics"
	pb "github.com/GrayCodeAI/tokman/pkg/api/proto/analyticsv1"
	"github.com/GrayCodeAI/tokman/services/analytics"
)

// AnalyticsServer implements the gRPC AnalyticsService.
type AnalyticsServer struct {
	pb.UnimplementedAnalyticsServiceServer
	svc *analytics.Service
}

// NewServer creates a new gRPC analytics server.
func NewServer(svc *analytics.Service) *AnalyticsServer {
	return &AnalyticsServer{svc: svc}
}

// Record saves a command execution record.
func (s *AnalyticsServer) Record(ctx context.Context, req *pb.RecordRequest) (*pb.RecordResponse, error) {
	start := time.Now()
	method := "AnalyticsService.Record"

	if req.Command == "" {
		metrics.RecordGRPCRequest(method, false, 0)
		return nil, status.Error(codes.InvalidArgument, "command is required")
	}

	recordID, err := s.svc.Record(ctx, &analytics.RecordRequest{
		TeamID:         req.TeamId,
		UserID:         req.UserId,
		Command:        req.Command,
		OriginalTokens: int(req.OriginalTokens),
		FilteredTokens: int(req.FilteredTokens),
		Duration:       time.Duration(req.DurationMs) * time.Millisecond,
		ExitCode:       int(req.DurationMs),
		FilterMode:     req.FilterMode,
		SessionID:      req.SessionId,
		AgentName:      req.AgentName,
		ModelName:      req.ModelName,
		Provider:       req.Provider,
	})
	if err != nil {
		durationMs := float64(time.Since(start).Nanoseconds()) / 1e6
		metrics.RecordGRPCRequest(method, false, durationMs)
		return nil, status.Errorf(codes.Internal, "failed to record: %v", err)
	}

	durationMs := float64(time.Since(start).Nanoseconds()) / 1e6
	metrics.RecordGRPCRequest(method, true, durationMs)

	return &pb.RecordResponse{Success: true, RecordId: recordID}, nil
}

// GetMetrics returns aggregated metrics.
func (s *AnalyticsServer) GetMetrics(ctx context.Context, req *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {
	start := time.Now()
	method := "AnalyticsService.GetMetrics"

	svcMetrics, err := s.svc.GetMetrics(ctx, &analytics.MetricsRequest{
		SessionID: req.SessionId,
	})
	if err != nil {
		durationMs := float64(time.Since(start).Nanoseconds()) / 1e6
		metrics.RecordGRPCRequest(method, false, durationMs)
		return nil, status.Errorf(codes.Internal, "failed to get metrics: %v", err)
	}

	pbFilters := make([]*pb.FilterUsage, len(svcMetrics.TopFilters))
	for i, f := range svcMetrics.TopFilters {
		pbFilters[i] = &pb.FilterUsage{
			FilterName:         f.FilterName,
			Count:              f.Count,
			AvgSavingsPercent:  f.AvgSavingPercent,
			TotalTokensSaved:   f.TotalTokensSaved,
			EffectivenessScore: f.EffectivenessScore,
		}
	}

	durationMs := float64(time.Since(start).Nanoseconds()) / 1e6
	metrics.RecordGRPCRequest(method, true, durationMs)

	return &pb.GetMetricsResponse{
		TotalCommands:         svcMetrics.TotalCommands,
		TotalTokensSaved:      svcMetrics.TotalTokensSaved,
		TotalTokensIn:         svcMetrics.TotalTokensIn,
		TotalTokensOut:        svcMetrics.TotalTokensOut,
		AverageSavingsPercent: svcMetrics.AverageSavingsPercent,
		P50LatencyMs:          svcMetrics.P50LatencyMs,
		P99LatencyMs:          svcMetrics.P99LatencyMs,
		TopFilters:            pbFilters,
		EstimatedSavingsUsd:   svcMetrics.EstimatedSavingsUSD,
	}, nil
}

// GetEconomics returns cost analysis.
func (s *AnalyticsServer) GetEconomics(ctx context.Context, req *pb.GetEconomicsRequest) (*pb.GetEconomicsResponse, error) {
	econ, err := s.svc.GetEconomics(ctx, req.TeamId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get economics: %v", err)
	}

	return &pb.GetEconomicsResponse{
		TokensSaved:                econ.TokensSaved,
		EstimatedCostSaved:         econ.EstimatedCostSaved,
		EstimatedCostWithoutTokman: econ.EstimatedCostWithoutTok,
		RoiPercent:                 econ.ROIPercent,
		MonthlyCommands:            int32(econ.MonthlyCommands),
		AvgTokensPerCommand:        econ.AvgTokensPerCommand,
		PrimaryModel:               econ.PrimaryModel,
		CostPerMtok:                econ.CostPerMTok,
		QuotaUsedPercent:           econ.QuotaUsedPercent,
	}, nil
}

// GetFilterEffectiveness returns filter performance rankings.
func (s *AnalyticsServer) GetFilterEffectiveness(ctx context.Context, req *pb.FilterEffectivenessRequest) (*pb.FilterEffectivenessResponse, error) {
	filters, err := s.svc.GetFilterEffectiveness(ctx, req.TeamId, int(req.Limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get filter effectiveness: %v", err)
	}

	pbFilters := make([]*pb.FilterMetric, len(filters))
	for i, f := range filters {
		pbFilters[i] = &pb.FilterMetric{
			FilterName:          f.FilterName,
			UsageCount:          f.UsageCount,
			AvgEffectiveness:    f.AvgEffectiveness,
			TotalTokensSaved:    f.TotalTokensSaved,
			AvgProcessingTimeMs: f.AvgProcessingTimeMs,
			FirstUsed:           timestampPtr(f.FirstUsed),
			LastUsed:            timestampPtr(f.LastUsed),
		}
	}

	return &pb.FilterEffectivenessResponse{Filters: pbFilters}, nil
}

// GetTrendAnalysis returns historical trends for dashboard.
func (s *AnalyticsServer) GetTrendAnalysis(ctx context.Context, req *pb.TrendAnalysisRequest) (*pb.TrendAnalysisResponse, error) {
	trends, err := s.svc.GetTrendAnalysis(ctx, req.TeamId, req.Granularity, int(req.Periods))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get trends: %v", err)
	}

	pbTrends := make([]*pb.TrendPoint, len(trends))
	for i, t := range trends {
		pbTrends[i] = &pb.TrendPoint{
			Timestamp:           timestampPtr(t.Timestamp),
			TotalCommands:       t.TotalCommands,
			TotalOriginalTokens: t.TotalOriginalTokens,
			TotalFilteredTokens: t.TotalFilteredTokens,
			TotalSavedTokens:    t.TotalSavedTokens,
			AvgReductionPercent: t.AvgReductionPercent,
			EstimatedCostUsd:    t.EstimatedCostUSD,
			EstimatedSavingsUsd: t.EstimatedSavingsUSD,
		}
	}

	return &pb.TrendAnalysisResponse{Points: pbTrends}, nil
}

// GetTeamStats returns aggregated stats for a team.
func (s *AnalyticsServer) GetTeamStats(ctx context.Context, req *pb.TeamStatsRequest) (*pb.TeamStatsResponse, error) {
	stats, err := s.svc.GetTeamStats(ctx, req.TeamId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get team stats: %v", err)
	}

	return &pb.TeamStatsResponse{
		TeamId:               stats.TeamID,
		TeamName:             stats.TeamName,
		MemberCount:          int32(stats.MemberCount),
		TotalCommands:        stats.TotalCommands,
		TotalTokensSaved:     stats.TotalTokensSaved,
		AvgReductionPercent:  stats.AvgReductionPercent,
		MonthlyTokenBudget:   int32(stats.MonthlyTokenBudget),
		TokensUsedThisMonth:  int32(stats.TokensUsedThisMonth),
		EstimatedMonthlyCost: stats.EstimatedMonthlyCost,
	}, nil
}

// GetUserStats returns stats for an individual user.
func (s *AnalyticsServer) GetUserStats(ctx context.Context, req *pb.UserStatsRequest) (*pb.UserStatsResponse, error) {
	stats, err := s.svc.GetUserStats(ctx, req.TeamId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user stats: %v", err)
	}

	return &pb.UserStatsResponse{
		UserId:              stats.UserID,
		Email:               stats.Email,
		TotalCommands:       stats.TotalCommands,
		TotalTokensSaved:    stats.TotalTokensSaved,
		AvgReductionPercent: stats.AvgReductionPercent,
		LastCommand:         timestampPtr(stats.LastCommand),
		MostUsedFilter:      stats.MostUsedFilter,
	}, nil
}

// GetCostProjection returns projected monthly costs.
func (s *AnalyticsServer) GetCostProjection(ctx context.Context, req *pb.CostProjectionRequest) (*pb.CostProjectionResponse, error) {
	proj, err := s.svc.GetCostProjection(ctx, req.TeamId, int(req.DaysToProject))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cost projection: %v", err)
	}

	return &pb.CostProjectionResponse{
		ProjectedMonthlyCost:    proj.ProjectedMonthlyCost,
		ProjectedMonthlySavings: proj.ProjectedMonthlySavings,
		ActualMonthlyCost:       proj.ActualMonthlyCost,
		ActualMonthlySavings:    proj.ActualMonthlySavings,
		ProjectionBasis:         proj.ProjectionBasis,
	}, nil
}

// GetAuditLogs returns audit trail for compliance.
func (s *AnalyticsServer) GetAuditLogs(ctx context.Context, req *pb.AuditLogsRequest) (*pb.AuditLogsResponse, error) {
	// TODO: Implement audit log retrieval
	return &pb.AuditLogsResponse{}, nil
}

// CreateTeam creates a new team/organization.
func (s *AnalyticsServer) CreateTeam(ctx context.Context, req *pb.CreateTeamRequest) (*pb.TeamResponse, error) {
	// TODO: Implement team creation
	return &pb.TeamResponse{}, nil
}

// GetTeam retrieves team details.
func (s *AnalyticsServer) GetTeam(ctx context.Context, req *pb.GetTeamRequest) (*pb.TeamResponse, error) {
	// TODO: Implement get team
	return &pb.TeamResponse{}, nil
}

// UpdateTeamBudget updates the monthly token budget.
func (s *AnalyticsServer) UpdateTeamBudget(ctx context.Context, req *pb.UpdateTeamBudgetRequest) (*pb.TeamResponse, error) {
	// TODO: Implement team budget update
	return &pb.TeamResponse{}, nil
}

// ListUsers lists team members.
func (s *AnalyticsServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// TODO: Implement list users
	return &pb.ListUsersResponse{}, nil
}

// GetDashboardData returns combined data for dashboard rendering.
func (s *AnalyticsServer) GetDashboardData(ctx context.Context, req *pb.DashboardRequest) (*pb.DashboardResponse, error) {
	dashboard, err := s.svc.GetDashboardData(ctx, req.TeamId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get dashboard data: %v", err)
	}

	// Build team stats
	pbTeamStats := &pb.TeamStats{
		TotalCommands:        dashboard.TeamStats.TotalCommands,
		TotalTokensSaved:     dashboard.TeamStats.TotalTokensSaved,
		AvgReductionPercent:  dashboard.TeamStats.AvgReductionPercent,
		MemberCount:          int32(dashboard.TeamStats.MemberCount),
		EstimatedMonthlyCost: dashboard.TeamStats.EstimatedMonthlyCost,
	}

	// Build economics
	pbEconomics := &pb.GetEconomicsResponse{
		TokensSaved:                dashboard.Economics.TokensSaved,
		EstimatedCostSaved:         dashboard.Economics.EstimatedCostSaved,
		EstimatedCostWithoutTokman: dashboard.Economics.EstimatedCostWithoutTok,
		RoiPercent:                 dashboard.Economics.ROIPercent,
	}

	// Build trends
	pbTrends := make([]*pb.TrendPoint, len(dashboard.Trends))
	for i, t := range dashboard.Trends {
		pbTrends[i] = &pb.TrendPoint{
			Timestamp:           timestampPtr(t.Timestamp),
			TotalCommands:       t.TotalCommands,
			TotalOriginalTokens: t.TotalOriginalTokens,
			TotalFilteredTokens: t.TotalFilteredTokens,
			TotalSavedTokens:    t.TotalSavedTokens,
			AvgReductionPercent: t.AvgReductionPercent,
			EstimatedCostUsd:    t.EstimatedCostUSD,
			EstimatedSavingsUsd: t.EstimatedSavingsUSD,
		}
	}

	// Build top filters
	pbFilters := make([]*pb.FilterMetric, len(dashboard.TopFilters))
	for i, f := range dashboard.TopFilters {
		pbFilters[i] = &pb.FilterMetric{
			FilterName:          f.FilterName,
			UsageCount:          f.UsageCount,
			AvgEffectiveness:    f.AvgEffectiveness,
			TotalTokensSaved:    f.TotalTokensSaved,
			AvgProcessingTimeMs: f.AvgProcessingTimeMs,
		}
	}

	// Build user stats
	pbUserStats := make([]*pb.UserStats, 0)

	// Build projection
	pbProjection := &pb.CostProjectionResponse{
		ProjectedMonthlyCost:    dashboard.CostProjection.ProjectedMonthlyCost,
		ProjectedMonthlySavings: dashboard.CostProjection.ProjectedMonthlySavings,
		ActualMonthlyCost:       dashboard.CostProjection.ActualMonthlyCost,
		ActualMonthlySavings:    dashboard.CostProjection.ActualMonthlySavings,
	}

	return &pb.DashboardResponse{
		TeamStats:  pbTeamStats,
		Economics:  pbEconomics,
		Trends:     pbTrends,
		TopFilters: pbFilters,
		UserStats:  pbUserStats,
		Projection: pbProjection,
	}, nil
}

// Helper function to convert time.Time to *timestamppb.Timestamp
func timestampPtr(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
