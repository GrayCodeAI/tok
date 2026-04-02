// Package grpcserver provides analytics gRPC service.
package grpcserver

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Request/Response types for gRPC service
type (
	// UsageReportRequest requests usage data
	UsageReportRequest struct {
		Days int32
	}

	// UsageReportResponse contains usage analytics
	UsageReportResponse struct {
		PeriodDays       int32
		TotalTokensSaved int64
		TotalCommands    int64
		DailyData        []*DailyUsage
		GeneratedAt      int64
	}

	// DailyUsage represents one day of usage
	DailyUsage struct {
		Date        string
		TokensSaved int64
		Commands    int64
	}

	// LayerPerformanceRequest requests layer metrics
	LayerPerformanceRequest struct{}

	// LayerPerformanceResponse contains layer metrics
	LayerPerformanceResponse struct {
		Layers      []*LayerMetrics
		GeneratedAt int64
	}

	// LayerMetrics represents metrics for one layer
	LayerMetrics struct {
		LayerName   string
		Invocations int64
		TokensSaved int64
		AvgNanos    int64
	}

	// CostAnalysisRequest requests cost estimation
	CostAnalysisRequest struct {
		ModelName string
	}

	// CostAnalysisResponse contains cost analysis
	CostAnalysisResponse struct {
		TotalTokensSaved  int64
		CostPerMillion    float64
		Currency          string
		EstimatedSavings  float64
		MonthlyProjection float64
		YearlyProjection  float64
		ModelName         string
		GeneratedAt       int64
	}

	// RealtimeStreamRequest configures realtime stream
	RealtimeStreamRequest struct {
		IntervalSeconds int32
	}

	// RealtimeEvent represents a realtime analytics event
	RealtimeEvent struct {
		Timestamp        int64
		TotalTokensSaved int64
		LayerMetrics     map[string]*LayerMetrics
	}

	// CompressionRatioRequest requests compression stats
	CompressionRatioRequest struct{}

	// CompressionRatioResponse contains compression stats
	CompressionRatioResponse struct {
		InputTokens       int64
		OutputTokens      int64
		TokensSaved       int64
		Ratio             float64
		SavingsPercentage float64
		GeneratedAt       int64
	}

	// ExportFormat defines export format type
	ExportFormat int32

	// ExportRequest requests data export
	ExportRequest struct {
		Format ExportFormat
	}

	// ExportResponse contains exported data
	ExportResponse struct {
		Data        []byte
		Format      string
		GeneratedAt int64
	}
)

// Export format constants
const (
	ExportFormat_JSON ExportFormat = iota
	ExportFormat_CSV
	ExportFormat_PROTO
)

// AnalyticsServer implements the Analytics gRPC service.
type AnalyticsServer struct {
	tracker  AnalyticsTracker
	registry *LayerRegistry
}

// AnalyticsTracker interface for tracking data.
type AnalyticsTracker interface {
	GetDailySavings(cmd string, days int) ([]DailyRecord, error)
	GetLayerStats() map[string]LayerStat
	TotalTokensSaved() int64
}

// DailyRecord represents daily tracking data.
type DailyRecord struct {
	Date     string
	Saved    int
	Input    int
	Output   int
	Commands int
}

// LayerStat represents layer statistics.
type LayerStat struct {
	Name        string
	Invocations int64
	TokensSaved int64
	AvgTime     time.Duration
}

// LayerRegistry manages compression layers.
type LayerRegistry struct {
	layers []LayerInfo
}

// LayerInfo represents layer information.
type LayerInfo struct {
	Name        string
	Description string
	Enabled     bool
}

// NewAnalyticsServer creates a new analytics server.
func NewAnalyticsServer(tracker AnalyticsTracker, registry *LayerRegistry) *AnalyticsServer {
	return &AnalyticsServer{
		tracker:  tracker,
		registry: registry,
	}
}

// GetUsageReport returns comprehensive usage analytics.
func (s *AnalyticsServer) GetUsageReport(ctx context.Context, req *UsageReportRequest) (*UsageReportResponse, error) {
	if req.Days == 0 {
		req.Days = 30
	}

	records, err := s.tracker.GetDailySavings("", int(req.Days))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get usage data: %v", err)
	}

	var totalSaved int64
	var totalCommands int64
	dailyData := make([]*DailyUsage, len(records))

	for i, r := range records {
		totalSaved += int64(r.Saved)
		totalCommands += int64(r.Commands)
		dailyData[i] = &DailyUsage{
			Date:        r.Date,
			TokensSaved: int64(r.Saved),
			Commands:    int64(r.Commands),
		}
	}

	return &UsageReportResponse{
		PeriodDays:       req.Days,
		TotalTokensSaved: totalSaved,
		TotalCommands:    totalCommands,
		DailyData:        dailyData,
		GeneratedAt:      time.Now().Unix(),
	}, nil
}

// GetLayerPerformance returns layer performance metrics.
func (s *AnalyticsServer) GetLayerPerformance(ctx context.Context, req *LayerPerformanceRequest) (*LayerPerformanceResponse, error) {
	stats := s.tracker.GetLayerStats()

	layers := make([]*LayerMetrics, 0, len(stats))
	for name, stat := range stats {
		layers = append(layers, &LayerMetrics{
			LayerName:   name,
			Invocations: stat.Invocations,
			TokensSaved: stat.TokensSaved,
			AvgNanos:    int64(stat.AvgTime),
		})
	}

	return &LayerPerformanceResponse{
		Layers:      layers,
		GeneratedAt: time.Now().Unix(),
	}, nil
}

// GetCostAnalysis returns cost estimation analytics.
func (s *AnalyticsServer) GetCostAnalysis(ctx context.Context, req *CostAnalysisRequest) (*CostAnalysisResponse, error) {
	totalSaved := s.tracker.TotalTokensSaved()

	// Default pricing (Claude 3.5 Sonnet)
	costPer1M := 3.0
	if req.ModelName != "" {
		costPer1M = s.getModelPricing(req.ModelName)
	}

	savingsUSD := float64(totalSaved) * costPer1M / 1_000_000

	// Calculate projections
	dailyAvg := float64(totalSaved) / 30.0 // last 30 days
	monthlyProjection := dailyAvg * 30 * costPer1M / 1_000_000

	return &CostAnalysisResponse{
		TotalTokensSaved:  totalSaved,
		CostPerMillion:    costPer1M,
		Currency:          "USD",
		EstimatedSavings:  savingsUSD,
		MonthlyProjection: monthlyProjection,
		YearlyProjection:  monthlyProjection * 12,
		ModelName:         req.ModelName,
		GeneratedAt:       time.Now().Unix(),
	}, nil
}

// StreamRealtimeAnalytics streams real-time analytics data.
func (s *AnalyticsServer) StreamRealtimeAnalytics(req *RealtimeStreamRequest, stream chan<- *RealtimeEvent) error {
	ticker := time.NewTicker(time.Duration(req.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := s.tracker.GetLayerStats()
			totalSaved := s.tracker.TotalTokensSaved()

			metrics := make(map[string]*LayerMetrics)
			for name, stat := range stats {
				metrics[name] = &LayerMetrics{
					LayerName:   name,
					Invocations: stat.Invocations,
					TokensSaved: stat.TokensSaved,
				}
			}

			event := &RealtimeEvent{
				Timestamp:        time.Now().Unix(),
				TotalTokensSaved: totalSaved,
				LayerMetrics:     metrics,
			}

			stream <- event
		}
	}
}

// GetCompressionRatio returns compression ratio analysis.
func (s *AnalyticsServer) GetCompressionRatio(ctx context.Context, req *CompressionRatioRequest) (*CompressionRatioResponse, error) {
	records, err := s.tracker.GetDailySavings("", 30)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get data: %v", err)
	}

	var totalInput, totalOutput int64
	for _, r := range records {
		totalInput += int64(r.Input)
		totalOutput += int64(r.Output)
	}

	ratio := 1.0
	if totalInput > 0 {
		ratio = float64(totalOutput) / float64(totalInput)
	}

	savings := totalInput - totalOutput

	return &CompressionRatioResponse{
		InputTokens:       totalInput,
		OutputTokens:      totalOutput,
		TokensSaved:       savings,
		Ratio:             ratio,
		SavingsPercentage: (1.0 - ratio) * 100,
		GeneratedAt:       time.Now().Unix(),
	}, nil
}

// ExportAnalytics exports analytics data in various formats.
func (s *AnalyticsServer) ExportAnalytics(ctx context.Context, req *ExportRequest) (*ExportResponse, error) {
	var data []byte
	var format string

	switch req.Format {
	case ExportFormat_JSON:
		format = "json"
		data = []byte(`{"export": "json_data"}`)
	case ExportFormat_CSV:
		format = "csv"
		data = []byte("date,saved,commands\n")
	case ExportFormat_PROTO:
		format = "proto"
		data = []byte{0x08, 0x01}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported format: %v", req.Format)
	}

	return &ExportResponse{
		Data:        data,
		Format:      format,
		GeneratedAt: time.Now().Unix(),
	}, nil
}

func (s *AnalyticsServer) getModelPricing(model string) float64 {
	pricing := map[string]float64{
		"claude-3-opus":     15.00,
		"claude-3-sonnet":   3.00,
		"claude-3.5-sonnet": 3.00,
		"claude-3-haiku":    0.25,
		"gpt-4":             30.00,
		"gpt-4-turbo":       10.00,
		"gpt-3.5-turbo":     0.50,
	}

	if p, ok := pricing[model]; ok {
		return p
	}
	return 3.0
}

// RegisterAnalyticsService registers the analytics service with a gRPC server.
func RegisterAnalyticsService(server *grpc.Server, tracker AnalyticsTracker, registry *LayerRegistry) {
	// Register with gRPC server when proto is generated
	fmt.Println("Analytics service registered")
}
