package learning

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"
)

// AdaptiveEngine implements the adaptive learning algorithm.
type AdaptiveEngine struct {
	db     *sql.DB
	logger *slog.Logger
	config *TrainingConfig
}

// NewAdaptiveEngine creates a new adaptive learning engine.
func NewAdaptiveEngine(db *sql.DB, logger *slog.Logger, config *TrainingConfig) *AdaptiveEngine {
	if config == nil {
		config = DefaultTrainingConfig()
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &AdaptiveEngine{
		db:     db,
		logger: logger,
		config: config,
	}
}

// RecordFeedback records user feedback on a filter.
func (ae *AdaptiveEngine) RecordFeedback(feedback *FilterFeedback) error {
	if feedback.ID == "" {
		feedback.ID = generateID()
	}
	feedback.CreatedAt = time.Now()

	_, err := ae.db.Exec(`
		INSERT INTO filter_feedback (id, team_id, user_id, command_id, filter_name, quality, relevance, helpful, comment, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, feedback.ID, feedback.TeamID, feedback.UserID, feedback.CommandID, feedback.FilterName,
		feedback.Quality, feedback.Relevance, feedback.Helpful, feedback.Comment, feedback.CreatedAt)

	return err
}

// Train trains the adaptive model for a codebase using collected feedback.
func (ae *AdaptiveEngine) Train(teamID, codebasePath string) (*FilterWeights, error) {
	// Get feedback for this codebase
	feedbacks, err := ae.getFeedbacks(teamID, codebasePath)
	if err != nil {
		return nil, err
	}

	if len(feedbacks) < ae.config.MinSampleSize {
		ae.logger.Info("insufficient feedback for training",
			"team_id", teamID,
			"codebase", codebasePath,
			"feedback_count", len(feedbacks),
			"min_required", ae.config.MinSampleSize,
		)
		return nil, fmt.Errorf("insufficient feedback: have %d, need %d", len(feedbacks), ae.config.MinSampleSize)
	}

	// Get current weights as starting point
	currentWeights, err := ae.getLatestWeights(teamID, codebasePath)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if currentWeights == nil {
		// Initialize with uniform weights
		currentWeights = &FilterWeights{
			ID:           generateID(),
			TeamID:       teamID,
			CodebasePath: codebasePath,
			Epoch:        0,
			Weights:      initializeWeights(feedbacks),
			Confidence:   0.1,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
	}

	// Update codebase hash
	currentWeights.CodebaseHash = hashCodebasePath(codebasePath)

	// Train using SGD with momentum
	optimizer := &MomentumOptimizer{
		weights:      currentWeights.Weights,
		velocity:     make(map[string]float64),
		beta:         ae.config.MomentumBeta,
		learningRate: ae.config.LearningRate,
	}

	// Run multiple epochs
	losses := []float64{}
	for epoch := 0; epoch < ae.config.MaxEpochs; epoch++ {
		loss := ae.trainEpoch(optimizer, feedbacks, ae.config.RegularizationL2)
		losses = append(losses, loss)

		ae.logger.Debug("training epoch complete",
			"epoch", epoch,
			"loss", loss,
			"team_id", teamID,
		)

		// Early stopping if loss converges
		if epoch > 2 && isConverged(losses[len(losses)-3:]) {
			ae.logger.Info("training converged early",
				"epoch", epoch,
				"loss", loss,
			)
			break
		}
	}

	// Update weights
	currentWeights.Weights = optimizer.weights
	currentWeights.Epoch++
	currentWeights.UpdatedAt = time.Now()
	currentWeights.AverageFeedback = calculateAverageFeedback(feedbacks)
	currentWeights.SampleCount = len(feedbacks)
	currentWeights.Confidence = min(currentWeights.Confidence+0.05, 0.95) // Cap at 0.95

	// Save weights
	if err := ae.saveWeights(currentWeights); err != nil {
		return nil, err
	}

	ae.logger.Info("training completed successfully",
		"team_id", teamID,
		"codebase", codebasePath,
		"epoch", currentWeights.Epoch,
		"average_feedback", currentWeights.AverageFeedback,
		"confidence", currentWeights.Confidence,
	)

	return currentWeights, nil
}

// trainEpoch runs a single training epoch.
func (ae *AdaptiveEngine) trainEpoch(optimizer *MomentumOptimizer, feedbacks []*FilterFeedback, regularization float64) float64 {
	totalLoss := 0.0

	for _, feedback := range feedbacks {
		// Convert feedback to target value (quality + relevance) / 10
		target := float64(feedback.Quality+feedback.Relevance) / 10.0

		// Get current weight for this filter
		w := optimizer.weights[feedback.FilterName]

		// Compute prediction (simple linear model)
		prediction := w * 0.5 // simplified: weight contributes 50% to quality

		// Compute loss (MSE)
		error := prediction - target
		loss := error * error

		// Add L2 regularization
		loss += regularization * w * w

		totalLoss += loss

		// Compute gradient
		gradient := 2 * error * 0.5

		// Add regularization gradient
		gradient += 2 * regularization * w

		// Update weight using momentum
		optimizer.Update(feedback.FilterName, gradient)
	}

	return totalLoss / float64(len(feedbacks))
}

// GetWeights retrieves the current weights for a codebase.
func (ae *AdaptiveEngine) GetWeights(teamID, codebasePath string) (*FilterWeights, error) {
	return ae.getLatestWeights(teamID, codebasePath)
}

// getLatestWeights gets the latest weights (internal).
func (ae *AdaptiveEngine) getLatestWeights(teamID, codebasePath string) (*FilterWeights, error) {
	var weights FilterWeights
	var weightsJSON []byte

	row := ae.db.QueryRow(`
		SELECT id, team_id, codebase_path, codebase_hash, epoch, weights, confidence, average_feedback, sample_count, created_at, updated_at
		FROM filter_weights
		WHERE team_id = ? AND codebase_path = ?
		ORDER BY epoch DESC
		LIMIT 1
	`, teamID, codebasePath)

	err := row.Scan(&weights.ID, &weights.TeamID, &weights.CodebasePath, &weights.CodebaseHash,
		&weights.Epoch, &weightsJSON, &weights.Confidence, &weights.AverageFeedback,
		&weights.SampleCount, &weights.CreatedAt, &weights.UpdatedAt)

	if err != nil {
		return nil, err
	}

	// Unmarshal weights JSON
	weights.Weights = unmarshalWeights(weightsJSON)

	return &weights, nil
}

// getFeedbacks retrieves feedbacks for training (internal).
func (ae *AdaptiveEngine) getFeedbacks(teamID, codebasePath string) ([]*FilterFeedback, error) {
	// TODO: Implement feedback retrieval from database
	// For now, return empty slice
	return []*FilterFeedback{}, nil
}

// saveWeights saves trained weights to the database.
func (ae *AdaptiveEngine) saveWeights(weights *FilterWeights) error {
	weightsJSON := marshalWeights(weights.Weights)

	_, err := ae.db.Exec(`
		INSERT INTO filter_weights (id, team_id, codebase_path, codebase_hash, epoch, weights, confidence, average_feedback, sample_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, weights.ID, weights.TeamID, weights.CodebasePath, weights.CodebaseHash,
		weights.Epoch, weightsJSON, weights.Confidence, weights.AverageFeedback,
		weights.SampleCount, weights.CreatedAt, weights.UpdatedAt)

	return err
}

// MomentumOptimizer implements SGD with momentum.
type MomentumOptimizer struct {
	weights      map[string]float64
	velocity     map[string]float64
	beta         float64
	learningRate float64
}

// Update performs a weight update with momentum.
func (m *MomentumOptimizer) Update(paramName string, gradient float64) {
	// Initialize velocity if needed
	if _, ok := m.velocity[paramName]; !ok {
		m.velocity[paramName] = 0
	}

	// Update velocity: v = beta * v + (1 - beta) * gradient
	m.velocity[paramName] = m.beta*m.velocity[paramName] + (1-m.beta)*gradient

	// Update weight: w = w - learningRate * v
	if _, ok := m.weights[paramName]; !ok {
		m.weights[paramName] = 0.5 // default weight
	}
	m.weights[paramName] -= m.learningRate * m.velocity[paramName]

	// Clip to [0, 1]
	m.weights[paramName] = max(0, min(1, m.weights[paramName]))
}

// Helper functions

func initializeWeights(feedbacks []*FilterFeedback) map[string]float64 {
	weights := make(map[string]float64)
	filterCount := make(map[string]int)

	for _, fb := range feedbacks {
		weights[fb.FilterName] += float64(fb.Quality) / 5.0
		filterCount[fb.FilterName]++
	}

	for filter := range weights {
		weights[filter] /= float64(filterCount[filter])
	}

	return weights
}

func calculateAverageFeedback(feedbacks []*FilterFeedback) float64 {
	if len(feedbacks) == 0 {
		return 0
	}
	sum := 0
	for _, fb := range feedbacks {
		sum += fb.Quality + fb.Relevance
	}
	return float64(sum) / float64(len(feedbacks)*2)
}

func isConverged(losses []float64) bool {
	if len(losses) < 3 {
		return false
	}
	// Check if loss is decreasing at less than 0.1% per epoch
	improvement := (losses[0] - losses[len(losses)-1]) / losses[0]
	return improvement < 0.001
}

func hashCodebasePath(path string) string {
	h := sha256.Sum256([]byte(path))
	return hex.EncodeToString(h[:])
}

func marshalWeights(weights map[string]float64) []byte {
	// Simple JSON serialization
	// TODO: Use proper JSON marshaling
	return []byte{}
}

func unmarshalWeights(data []byte) map[string]float64 {
	// Simple JSON deserialization
	// TODO: Use proper JSON unmarshaling
	return make(map[string]float64)
}

func generateID() string {
	// TODO: Implement ID generation
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
