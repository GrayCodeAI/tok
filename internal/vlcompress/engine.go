package vlcompress

import (
	"context"
	"math"
	"sort"
	"sync"
)

type VLCompressionEngine struct {
	mu         sync.RWMutex
	config     EngineConfig
	strategies map[string]CompressionStrategy
	stats      EngineStats
}

type EngineConfig struct {
	EnableTokenPruning   bool
	EnableDualPivot      bool
	EnableDensityCluster bool
	FlashAttention       bool
	MaxTokens            int
	QualityThreshold     float64
}

type EngineStats struct {
	TotalCompressions int64
	TokensPruned      int64
	TokensMerged      int64
	AvgReduction      float64
	StrategyUsage     map[string]int64
}

type CompressionStrategy interface {
	Compress(ctx context.Context, input *VLInput) (*VLOutput, error)
	Name() string
}

type VLInput struct {
	Images    []ImageToken
	Text      string
	Attention [][]float64
	Positions []int
}

type VLOutput struct {
	Images        []ImageToken
	Text          string
	TokensKept    int
	TokensRemoved int
	Reduction     float64
	Quality       float64
}

type ImageToken struct {
	ID        string
	Data      []byte
	Features  []float64
	Region    Region
	Attention float64
}

type Region struct {
	X, Y       int
	Width      int
	Height     int
	Confidence float64
}

type DualPivotClusterer struct {
	pivotsA []int
	pivotsB []int
}

func NewVLCompressionEngine(config EngineConfig) *VLCompressionEngine {
	e := &VLCompressionEngine{
		config:     config,
		strategies: make(map[string]CompressionStrategy),
		stats: EngineStats{
			StrategyUsage: make(map[string]int64),
		},
	}

	e.registerStrategies()

	return e
}

func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		EnableTokenPruning:   true,
		EnableDualPivot:      true,
		EnableDensityCluster: true,
		FlashAttention:       true,
		MaxTokens:            2048,
		QualityThreshold:     0.8,
	}
}

func (e *VLCompressionEngine) registerStrategies() {
	e.RegisterStrategy(&TokenPruner{})
	e.RegisterStrategy(&DualPivotClustering{})
	e.RegisterStrategy(&DensityClustering{})
	e.RegisterStrategy(&AttentionOptimizer{})
}

func (e *VLCompressionEngine) RegisterStrategy(s CompressionStrategy) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.strategies[s.Name()] = s
}

func (e *VLCompressionEngine) Compress(ctx context.Context, input *VLInput) (*VLOutput, error) {
	var output *VLOutput
	var err error

	e.mu.RLock()
	pruner, hasPruner := e.strategies["token_pruner"]
	e.mu.RUnlock()

	if hasPruner && e.config.EnableTokenPruning {
		output, err = pruner.Compress(ctx, input)
		if err != nil {
			return nil, err
		}
	}

	e.mu.RLock()
	dualPivot, hasDualPivot := e.strategies["dual_pivot"]
	e.mu.RUnlock()

	if hasDualPivot && e.config.EnableDualPivot {
		dpOutput, err := dualPivot.Compress(ctx, &VLInput{
			Images:    output.Images,
			Text:      output.Text,
			Attention: input.Attention,
			Positions: input.Positions,
		})
		if err == nil {
			output = dpOutput
		}
	}

	e.mu.Lock()
	e.stats.TotalCompressions++
	e.stats.TokensPruned += int64(output.TokensRemoved)
	e.stats.TokensMerged += int64(len(input.Images) - len(output.Images))
	e.stats.AvgReduction = (e.stats.AvgReduction*float64(e.stats.TotalCompressions-1) + output.Reduction) / float64(e.stats.TotalCompressions)
	e.mu.Unlock()

	return output, nil
}

func (e *VLCompressionEngine) GetStats() EngineStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stats
}

type TokenPruner struct{}

func (p *TokenPruner) Name() string { return "token_pruner" }

func (p *TokenPruner) Compress(ctx context.Context, input *VLInput) (*VLOutput, error) {
	output := &VLOutput{
		Images:     make([]ImageToken, 0, len(input.Images)),
		Text:       input.Text,
		TokensKept: 0,
	}

	imageTokens := input.Images
	if len(imageTokens) > 100 {
		sorted := make([]ImageToken, len(imageTokens))
		copy(sorted, imageTokens)

		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Attention > sorted[j].Attention
		})

		keepCount := len(sorted) / 2
		imageTokens = sorted[:keepCount]
	}

	for _, token := range imageTokens {
		if token.Attention > 0.1 {
			output.Images = append(output.Images, token)
			output.TokensKept++
		}
	}

	output.TokensRemoved = len(input.Images) - len(output.Images)

	if len(input.Images) > 0 {
		output.Reduction = float64(output.TokensRemoved) / float64(len(input.Images))
	}

	output.Quality = calculateVLQuality(input, output)

	return output, nil
}

func calculateVLQuality(input *VLInput, output *VLOutput) float64 {
	if len(input.Images) == 0 {
		return 1.0
	}

	keptRatio := float64(len(output.Images)) / float64(len(input.Images))
	attentionPreserved := 0.0

	for _, token := range output.Images {
		attentionPreserved += token.Attention
	}

	for _, token := range input.Images {
		attentionPreserved -= token.Attention
	}

	attentionPreserved = math.Abs(attentionPreserved)
	quality := keptRatio * 0.7

	if attentionPreserved < 0.5 {
		quality += 0.3
	}

	return math.Min(1.0, quality)
}

type DualPivotClustering struct {
	numPivots int
}

func (d *DualPivotClustering) Name() string { return "dual_pivot" }

func (d *DualPivotClustering) Compress(ctx context.Context, input *VLInput) (*VLOutput, error) {
	output := &VLOutput{
		Images:     make([]ImageToken, 0),
		Text:       input.Text,
		TokensKept: 0,
	}

	if len(input.Images) <= 2 {
		output.Images = input.Images
		output.TokensKept = len(input.Images)
		return output, nil
	}

	pivotA := input.Images[0]
	pivotB := input.Images[len(input.Images)-1]

	clusterA := []ImageToken{pivotA}
	clusterB := []ImageToken{pivotB}

	for i := 1; i < len(input.Images)-1; i++ {
		token := input.Images[i]

		distA := cosineSimilarity(token.Features, pivotA.Features)
		distB := cosineSimilarity(token.Features, pivotB.Features)

		if distA < distB {
			clusterA = append(clusterA, token)
		} else {
			clusterB = append(clusterB, token)
		}
	}

	output.Images = append(output.Images, compressCluster(clusterA)...)
	output.Images = append(output.Images, compressCluster(clusterB)...)

	output.TokensKept = len(output.Images)
	output.TokensRemoved = len(input.Images) - output.TokensKept

	if len(input.Images) > 0 {
		output.Reduction = float64(output.TokensRemoved) / float64(len(input.Images))
	}

	output.Quality = 0.9

	return output, nil
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 1.0
	}

	dot := 0.0
	normA := 0.0
	normB := 0.0

	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 1.0
	}

	return 1.0 - dot/(math.Sqrt(normA)*math.Sqrt(normB))
}

func compressCluster(cluster []ImageToken) []ImageToken {
	if len(cluster) <= 1 {
		return cluster
	}

	result := []ImageToken{cluster[0]}

	if len(cluster) > 2 {
		result = append(result, cluster[len(cluster)/2])
	}

	result = append(result, cluster[len(cluster)-1])

	return result
}

type DensityClustering struct {
	epsilon float64
	minPts  int
}

func (d *DensityClustering) Name() string { return "density" }

func (d *DensityClustering) Compress(ctx context.Context, input *VLInput) (*VLOutput, error) {
	output := &VLOutput{
		Images:     make([]ImageToken, 0),
		Text:       input.Text,
		TokensKept: 0,
	}

	if len(input.Images) == 0 {
		return output, nil
	}

	densities := make([]float64, len(input.Images))
	for i := range input.Images {
		for j := range input.Images {
			if i == j {
				continue
			}
			dist := euclideanDist(input.Images[i].Features, input.Images[j].Features)
			if dist < 0.5 {
				densities[i]++
			}
		}
	}

	sorted := make([]struct {
		index     int
		density   float64
		attention float64
	}, len(input.Images))

	for i := range input.Images {
		sorted[i] = struct {
			index     int
			density   float64
			attention float64
		}{i, densities[i], input.Images[i].Attention}
	}

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].density != sorted[j].density {
			return sorted[i].density > sorted[j].density
		}
		return sorted[i].attention > sorted[j].attention
	})

	keepCount := len(sorted) / 2
	if keepCount < 1 {
		keepCount = 1
	}

	for i := 0; i < keepCount; i++ {
		output.Images = append(output.Images, input.Images[sorted[i].index])
	}

	output.TokensKept = len(output.Images)
	output.TokensRemoved = len(input.Images) - output.TokensKept

	if len(input.Images) > 0 {
		output.Reduction = float64(output.TokensRemoved) / float64(len(input.Images))
	}

	output.Quality = 0.85

	return output, nil
}

func euclideanDist(a, b []float64) float64 {
	if len(a) != len(b) {
		return 1.0
	}

	sum := 0.0
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}

	return math.Sqrt(sum)
}

type AttentionOptimizer struct {
	windowSize int
}

func (a *AttentionOptimizer) Name() string { return "attention" }

func (a *AttentionOptimizer) Compress(ctx context.Context, input *VLInput) (*VLOutput, error) {
	output := &VLOutput{
		Images:     make([]ImageToken, 0, len(input.Images)),
		Text:       input.Text,
		TokensKept: 0,
	}

	if len(input.Attention) == 0 || len(input.Images) == 0 {
		output.Images = input.Images
		output.TokensKept = len(input.Images)
		return output, nil
	}

	attentionMatrix := input.Attention

	for i, token := range input.Images {
		rowSum := 0.0
		if i < len(attentionMatrix) {
			for j := 0; j < len(attentionMatrix[i]); j++ {
				rowSum += attentionMatrix[i][j]
			}
		}

		if rowSum > 0.01 || i < 5 || i >= len(input.Images)-3 {
			output.Images = append(output.Images, token)
			output.TokensKept++
		}
	}

	output.TokensRemoved = len(input.Images) - output.TokensKept

	if len(input.Images) > 0 {
		output.Reduction = float64(output.TokensRemoved) / float64(len(input.Images))
	}

	output.Quality = 0.88

	return output, nil
}
