package quality

import (
	"context"
	"math"
	"sync"
	"time"
)

type MetricsPipeline struct {
	mu         sync.RWMutex
	collector  *MetricsCollector
	analyzer   *QualityAnalyzer
	thresholds ThresholdConfig
	stats      PipelineStats
}

type PipelineStats struct {
	TotalProcessed     int64
	TotalHighQuality   int64
	TotalMediumQuality int64
	TotalLowQuality    int64
	AverageQuality     float64
	AverageLatencyMs   float64
}

type ThresholdConfig struct {
	HighQualityMin   float64
	MediumQualityMin float64
	MaxLatencyMs     float64
	MinCompression   float64
	MaxCompression   float64
}

type MetricsCollector struct {
	mu      sync.RWMutex
	samples []CompressionSample
	maxSize int
}

type CompressionSample struct {
	Timestamp        time.Time
	OriginalSize     int
	CompressedSize   int
	OriginalTokens   int
	CompressedTokens int
	Algorithm        string
	DurationMs       float64
	Command          string
	Agent            string
}

type QualityAnalyzer struct {
	mu         sync.RWMutex
	windowSize int
	weights    QualityWeights
	history    []QualityMetrics
}

type QualityWeights struct {
	CompressionWeight  float64
	LatencyWeight      float64
	FidelityWeight     float64
	TokenSavingsWeight float64
}

type QualityMetrics struct {
	Timestamp    time.Time
	OverallScore float64
	Compression  float64
	Latency      float64
	Fidelity     float64
	TokenSavings float64
}

type QualityReport struct {
	OverallScore     float64
	CompressionRatio float64
	LatencyMs        float64
	FidelityScore    float64
	TokenSavingsPct  float64
	QualityTier      string
	Issues           []string
	Recommendations  []string
}

func NewMetricsPipeline(config PipelineConfig) *MetricsPipeline {
	return &MetricsPipeline{
		collector:  NewMetricsCollector(config.SampleBufferSize),
		analyzer:   NewQualityAnalyzer(config.WindowSize),
		thresholds: config.Thresholds,
		stats:      PipelineStats{},
	}
}

type PipelineConfig struct {
	SampleBufferSize int
	WindowSize       int
	Thresholds       ThresholdConfig
}

func DefaultPipelineConfig() PipelineConfig {
	return PipelineConfig{
		SampleBufferSize: 1000,
		WindowSize:       100,
		Thresholds: ThresholdConfig{
			HighQualityMin:   0.85,
			MediumQualityMin: 0.70,
			MaxLatencyMs:     100,
			MinCompression:   0.3,
			MaxCompression:   0.95,
		},
	}
}

func (mp *MetricsPipeline) RecordCompression(ctx context.Context, sample CompressionSample) {
	mp.collector.Add(sample)

	score := mp.analyzer.Analyze(sample)

	mp.mu.Lock()
	mp.stats.TotalProcessed++

	switch {
	case score.OverallScore >= mp.thresholds.HighQualityMin:
		mp.stats.TotalHighQuality++
	case score.OverallScore >= mp.thresholds.MediumQualityMin:
		mp.stats.TotalMediumQuality++
	default:
		mp.stats.TotalLowQuality++
	}

	if mp.stats.TotalProcessed > 0 {
		mp.stats.AverageQuality = (mp.stats.AverageQuality*float64(mp.stats.TotalProcessed-1) + score.OverallScore) / float64(mp.stats.TotalProcessed)
	}

	mp.stats.AverageLatencyMs = (mp.stats.AverageLatencyMs*float64(mp.stats.TotalProcessed-1) + sample.DurationMs) / float64(mp.stats.TotalProcessed)
	mp.mu.Unlock()
}

func (mp *MetricsPipeline) GetQualityReport(ctx context.Context) QualityReport {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	samples := mp.collector.GetRecent(100)

	if len(samples) == 0 {
		return QualityReport{
			QualityTier: "unknown",
			Issues:      []string{"No samples collected yet"},
		}
	}

	var totalCompression, totalLatency, totalFidelity, totalSavings float64
	for _, s := range samples {
		compRatio := 1.0 - float64(s.CompressedSize)/float64(s.OriginalSize)
		latencyScore := mp.thresholds.MaxLatencyMs - s.DurationMs
		if latencyScore < 0 {
			latencyScore = 0
		}
		latencyScore = latencyScore / mp.thresholds.MaxLatencyMs

		tokenSavings := float64(s.OriginalTokens-s.CompressedTokens) / float64(s.OriginalTokens)

		totalCompression += compRatio
		totalLatency += latencyScore
		totalFidelity += 0.9
		totalSavings += tokenSavings
	}

	n := float64(len(samples))
	avgCompression := totalCompression / n
	avgLatency := totalLatency / n
	avgFidelity := totalFidelity / n
	avgSavings := totalSavings / n

	weights := QualityWeights{
		CompressionWeight:  0.3,
		LatencyWeight:      0.2,
		FidelityWeight:     0.3,
		TokenSavingsWeight: 0.2,
	}

	overall := avgCompression*weights.CompressionWeight +
		avgLatency*weights.LatencyWeight +
		avgFidelity*weights.FidelityWeight +
		avgSavings*weights.TokenSavingsWeight

	report := QualityReport{
		OverallScore:     overall,
		CompressionRatio: avgCompression,
		LatencyMs:        totalLatency / n,
		FidelityScore:    avgFidelity,
		TokenSavingsPct:  avgSavings * 100,
	}

	switch {
	case overall >= mp.thresholds.HighQualityMin:
		report.QualityTier = "high"
	case overall >= mp.thresholds.MediumQualityMin:
		report.QualityTier = "medium"
	default:
		report.QualityTier = "low"
	}

	if avgCompression < mp.thresholds.MinCompression {
		report.Issues = append(report.Issues, "Compression ratio below minimum threshold")
		report.Recommendations = append(report.Recommendations, "Consider using stronger compression or different algorithm")
	}

	if avgLatency > mp.thresholds.MaxLatencyMs {
		report.Issues = append(report.Issues, "Latency exceeds maximum threshold")
		report.Recommendations = append(report.Recommendations, "Consider async compression or caching")
	}

	return report
}

func (mp *MetricsPipeline) GetStats() PipelineStats {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.stats
}

func NewMetricsCollector(maxSize int) *MetricsCollector {
	return &MetricsCollector{
		samples: make([]CompressionSample, 0, maxSize),
		maxSize: maxSize,
	}
}

func (mc *MetricsCollector) Add(sample CompressionSample) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.samples = append(mc.samples, sample)

	if len(mc.samples) > mc.maxSize {
		mc.samples = mc.samples[1:]
	}
}

func (mc *MetricsCollector) GetRecent(n int) []CompressionSample {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if n > len(mc.samples) {
		n = len(mc.samples)
	}

	result := make([]CompressionSample, n)
	copy(result, mc.samples[len(mc.samples)-n:])
	return result
}

func NewQualityAnalyzer(windowSize int) *QualityAnalyzer {
	return &QualityAnalyzer{
		windowSize: windowSize,
		weights: QualityWeights{
			CompressionWeight:  0.3,
			LatencyWeight:      0.2,
			FidelityWeight:     0.3,
			TokenSavingsWeight: 0.2,
		},
		history: make([]QualityMetrics, 0, windowSize),
	}
}

func (qa *QualityAnalyzer) Analyze(sample CompressionSample) QualityMetrics {
	compressionRatio := 1.0 - float64(sample.CompressedSize)/float64(sample.OriginalSize)

	latencyScore := 100.0 - sample.DurationMs
	if latencyScore < 0 {
		latencyScore = 0
	}

	tokenSavings := 0.0
	if sample.OriginalTokens > 0 {
		tokenSavings = float64(sample.OriginalTokens-sample.CompressedTokens) / float64(sample.OriginalTokens)
	}

	fidelityScore := qa.estimateFidelity(sample)

	overall := compressionRatio*qa.weights.CompressionWeight +
		latencyScore/100*qa.weights.LatencyWeight +
		fidelityScore*qa.weights.FidelityWeight +
		tokenSavings*qa.weights.TokenSavingsWeight

	score := QualityMetrics{
		Timestamp:    time.Now(),
		OverallScore: overall,
		Compression:  compressionRatio,
		Latency:      latencyScore / 100,
		Fidelity:     fidelityScore,
		TokenSavings: tokenSavings,
	}

	qa.mu.Lock()
	qa.history = append(qa.history, score)
	if len(qa.history) > qa.windowSize {
		qa.history = qa.history[1:]
	}
	qa.mu.Unlock()

	return score
}

func (qa *QualityAnalyzer) estimateFidelity(sample CompressionSample) float64 {
	ratio := float64(sample.CompressedSize) / float64(sample.OriginalSize)

	fidelity := 1.0 - ratio
	if fidelity < 0 {
		fidelity = 0
	}

	fidelity = math.Min(1.0, fidelity+0.2)

	return fidelity
}
