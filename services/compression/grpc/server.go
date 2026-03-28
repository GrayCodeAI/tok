// Package grpc provides gRPC server implementation for the compression service.
package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/metrics"
	pb "github.com/GrayCodeAI/tokman/pkg/api/proto/compressionv1"
	"github.com/GrayCodeAI/tokman/services/compression"
)

// CompressionServer implements the gRPC CompressionService.
type CompressionServer struct {
	pb.UnimplementedCompressionServiceServer
	svc *compression.Service
}

// NewServer creates a new gRPC compression server.
func NewServer(svc *compression.Service) *CompressionServer {
	return &CompressionServer{svc: svc}
}

// Compress applies the compression pipeline to input text.
func (s *CompressionServer) Compress(ctx context.Context, req *pb.CompressRequest) (*pb.CompressResponse, error) {
	start := time.Now()
	method := "CompressionService.Compress"

	if req.Input == "" {
		metrics.RecordGRPCRequest(method, false, 0)
		return nil, status.Error(codes.InvalidArgument, "input is required")
	}

	// Map mode string to filter mode
	mode := compressionModeFromString(req.Mode)

	resp, err := s.svc.Compress(ctx, &compression.CompressRequest{
		Input:       req.Input,
		Mode:        mode,
		QueryIntent: req.QueryIntent,
		Budget:      int(req.Budget),
		Preset:      req.Preset,
	})
	if err != nil {
		durationMs := float64(time.Since(start).Nanoseconds()) / 1e6
		metrics.RecordGRPCRequest(method, false, durationMs)
		return nil, status.Errorf(codes.Internal, "compression failed: %v", err)
	}

	durationMs := float64(time.Since(start).Nanoseconds()) / 1e6
	metrics.RecordGRPCRequest(method, true, durationMs)

	return &pb.CompressResponse{
		Output:           resp.Output,
		OriginalTokens:   int32(resp.OriginalSize),
		CompressedTokens: int32(resp.CompressedSize),
		SavingsPercent:   resp.SavingsPercent,
		LayersApplied:    resp.LayersApplied,
	}, nil
}

// GetStats returns compression statistics.
func (s *CompressionServer) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	stats, err := s.svc.GetStats(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get stats: %v", err)
	}

	return &pb.GetStatsResponse{
		TotalCompressions:  stats.TotalCompressions,
		TotalTokensSaved:   stats.TotalTokensSaved,
		AverageCompression: stats.AverageCompression,
		P99LatencyMs:       stats.P99LatencyMs,
	}, nil
}

// GetLayers returns information about all compression layers.
func (s *CompressionServer) GetLayers(ctx context.Context, req *pb.GetLayersRequest) (*pb.GetLayersResponse, error) {
	layers, err := s.svc.GetLayers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get layers: %v", err)
	}

	pbLayers := make([]*pb.LayerInfo, len(layers))
	for i, l := range layers {
		pbLayers[i] = &pb.LayerInfo{
			Number:      int32(l.Number),
			Name:        l.Name,
			Research:    l.Research,
			Compression: l.Compression,
		}
	}

	return &pb.GetLayersResponse{Layers: pbLayers}, nil
}

// compressionModeFromString converts a string mode to filter mode.
func compressionModeFromString(mode string) filter.Mode {
	switch mode {
	case "aggressive":
		return filter.ModeAggressive
	case "minimal":
		return filter.ModeMinimal
	default:
		return filter.ModeNone
	}
}
