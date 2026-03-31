// Package grpc provides gRPC server implementation for the analytics service.
package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

	err := s.svc.Record(ctx, &analytics.RecordRequest{
		Command:        req.Command,
		OriginalTokens: int(req.OriginalTokens),
		FilteredTokens: int(req.FilteredTokens),
		FilterMode:     req.FilterMode,
		SessionID:      req.SessionId,
	})
	if err != nil {
		durationMs := float64(time.Since(start).Nanoseconds()) / 1e6
		metrics.RecordGRPCRequest(method, false, durationMs)
		return nil, status.Errorf(codes.Internal, "failed to record: %v", err)
	}

	durationMs := float64(time.Since(start).Nanoseconds()) / 1e6
	metrics.RecordGRPCRequest(method, true, durationMs)

	return &pb.RecordResponse{Success: true}, nil
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
			FilterName: f.FilterName,
			Count:      f.Count,
			AvgSavings: f.AvgSavings,
		}
	}

	durationMs := float64(time.Since(start).Nanoseconds()) / 1e6
	metrics.RecordGRPCRequest(method, true, durationMs)

	return &pb.GetMetricsResponse{
		TotalCommands:    svcMetrics.TotalCommands,
		TotalTokensSaved: svcMetrics.TotalTokensSaved,
		TotalTokensIn:    svcMetrics.TotalTokensIn,
		TotalTokensOut:   svcMetrics.TotalTokensOut,
		AverageSavings:   svcMetrics.AverageSavings,
		P50LatencyMs:     svcMetrics.P50LatencyMs,
		P99LatencyMs:     svcMetrics.P99LatencyMs,
		TopFilters:       pbFilters,
	}, nil
}

// GetEconomics returns cost analysis.
func (s *AnalyticsServer) GetEconomics(ctx context.Context, req *pb.GetEconomicsRequest) (*pb.GetEconomicsResponse, error) {
	econ, err := s.svc.GetEconomics(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get economics: %v", err)
	}

	return &pb.GetEconomicsResponse{
		TokensSaved:        econ.TokensSaved,
		EstimatedCostSaved: econ.EstimatedCostSaved,
		ModelUsed:          econ.ModelUsed,
		CostPerToken:       econ.CostPerToken,
		QuotaUsed:          econ.QuotaUsed,
		QuotaRemaining:     econ.QuotaRemaining,
	}, nil
}
