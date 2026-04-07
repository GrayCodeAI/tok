package learning

import "time"

// FilterFeedback represents user feedback on a filter's output quality.
type FilterFeedback struct {
	ID          string
	TeamID      string
	UserID      string
	CommandID   int64
	FilterName  string
	Quality     int // 1-5 scale: 1=poor, 5=excellent
	Relevance   int // 1-5 scale: 1=irrelevant, 5=perfect
	Helpful     bool
	Comment     string
	CreatedAt   time.Time
}

// FilterWeights represents learned weights for filters in a codebase.
type FilterWeights struct {
	ID              string
	TeamID          string
	CodebasePath    string
	CodebaseHash    string // SHA256 of codebase structure
	Epoch           int
	Weights         map[string]float64
	Confidence      float64 // 0-1: confidence in weights
	AverageFeedback float64 // Average feedback score
	SampleCount     int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// WeightUpdate represents a single weight update during training.
type WeightUpdate struct {
	FilterName    string
	OldWeight     float64
	NewWeight     float64
	Delta         float64
	LearningRate  float64
	ConfidenceGain float64
}

// ABTestVariant represents a variant in an A/B test.
type ABTestVariant struct {
	ID        string
	TestID    string
	Name      string // "control", "variant_a", "variant_b", etc.
	Weights   map[string]float64
	Rollout   float64 // 0-1: percentage of traffic
	IsControl bool
	Metrics   *ABTestMetrics
}

// ABTestMetrics aggregates metrics for an A/B test variant.
type ABTestMetrics struct {
	Conversions      int
	Improvements     int
	AverageFeedback  float64
	SampleSize       int
	ConfidenceLevel  float64
	Winner           bool
}

// AdaptiveSession represents a single session's learning data.
type AdaptiveSession struct {
	ID              string
	TeamID          string
	UserID          string
	CodebasePath    string
	CodebaseHash    string
	FilterFeedbacks []*FilterFeedback
	CommandRecords  []CommandRecord
	SessionWeights  map[string]float64
	LearningGain    float64 // Improvement in compression ratio
	CreatedAt       time.Time
}

// CommandRecord represents data from command execution.
type CommandRecord struct {
	ID             int64
	Command        string
	OriginalTokens int
	FilteredTokens int
	SavedTokens    int
	FilterMode     string
	ExecTimeMs     int64
}

// TrainingConfig represents configuration for the learning algorithm.
type TrainingConfig struct {
	LearningRate      float64 // How aggressive weight updates are
	MinSampleSize     int     // Minimum feedback samples before training
	ConfidenceThreshold float64 // Minimum confidence to deploy weights
	EpochSize         int     // Samples per training epoch
	MaxEpochs         int     // Maximum training epochs
	RegularizationL2  float64 // L2 regularization strength
	MomentumBeta      float64 // Momentum for gradient updates
}

// LearningProgress tracks overall learning progress.
type LearningProgress struct {
	TeamID            string
	CodebasePath      string
	FeedbackCount     int
	WeightUpdateCount int
	AverageImprovement float64
	LatestEpoch       int
	LastTrainedAt     time.Time
	NextTrainDueAt    time.Time
}

// DefaultTrainingConfig returns sensible defaults.
func DefaultTrainingConfig() *TrainingConfig {
	return &TrainingConfig{
		LearningRate:        0.01,
		MinSampleSize:       50,
		ConfidenceThreshold: 0.75,
		EpochSize:           100,
		MaxEpochs:           10,
		RegularizationL2:    0.001,
		MomentumBeta:        0.9,
	}
}
