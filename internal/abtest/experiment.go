// Package abtest provides A/B testing capabilities for TokMan
package abtest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// Experiment represents an A/B test experiment
type Experiment struct {
	ID                string
	Name              string
	Description       string
	Hypothesis        string
	Status            Status
	Type              ExperimentType
	Variants          []Variant
	TrafficAllocation float64
	PrimaryMetric     string
	SecondaryMetrics  []string
	SampleSize        int
	Duration          time.Duration
	StartTime         time.Time
	EndTime           time.Time
	Segments          []Segment
	Randomization     RandomizationMethod
	Results           *Results
	mu                sync.RWMutex
}

// Status represents experiment status
type Status string

const (
	StatusDraft     Status = "draft"
	StatusScheduled Status = "scheduled"
	StatusRunning   Status = "running"
	StatusPaused    Status = "paused"
	StatusCompleted Status = "completed"
	StatusStopped   Status = "stopped"
)

// ExperimentType defines the type of experiment
type ExperimentType string

const (
	TypeAB           ExperimentType = "ab"
	TypeMultivariate ExperimentType = "multivariate"
	TypeBandit       ExperimentType = "bandit"
	TypeSwitchback   ExperimentType = "switchback"
)

// Variant represents an experiment variant
type Variant struct {
	ID             string
	Name           string
	Description    string
	TrafficPercent float64
	Config         map[string]interface{}
	Metrics        VariantMetrics
	IsControl      bool
	IsWinner       bool
}

// VariantMetrics holds metrics for a variant
type VariantMetrics struct {
	Impressions    int64
	Conversions    int64
	ConversionRate float64
	Revenue        float64
	Engagement     float64
	CustomMetrics  map[string]float64
}

// Segment defines user segments
type Segment struct {
	Name              string
	Condition         SegmentCondition
	TrafficAllocation float64
}

// SegmentCondition defines a segment condition
type SegmentCondition struct {
	Attribute string
	Operator  string
	Value     interface{}
}

// RandomizationMethod defines how users are assigned to variants
type RandomizationMethod string

const (
	RandomizationRandom   RandomizationMethod = "random"
	RandomizationUserID   RandomizationMethod = "user_id"
	RandomizationSession  RandomizationMethod = "session"
	RandomizationDevice   RandomizationMethod = "device"
	RandomizationWeighted RandomizationMethod = "weighted"
)

// Results holds experiment results
type Results struct {
	StartTime           time.Time
	EndTime             time.Time
	TotalParticipants   int64
	Winner              *Variant
	ConfidenceLevel     float64
	StatisticalPower    float64
	PValues             map[string]float64
	EffectSizes         map[string]float64
	ConfidenceIntervals map[string]CI
	Recommendation      string
}

// CI represents a confidence interval
type CI struct {
	Lower float64
	Upper float64
}

// Manager manages A/B test experiments
type Manager struct {
	experiments map[string]*Experiment
	assignments map[string]string // userID -> variantID
	mu          sync.RWMutex
	storage     Storage
}

// Storage interface for persisting experiments
type Storage interface {
	Save(experiment *Experiment) error
	Load(id string) (*Experiment, error)
	List() ([]*Experiment, error)
}

// NewManager creates an A/B test manager
func NewManager(storage Storage) *Manager {
	return &Manager{
		experiments: make(map[string]*Experiment),
		assignments: make(map[string]string),
		storage:     storage,
	}
}

// CreateExperiment creates a new experiment
func (m *Manager) CreateExperiment(config ExperimentConfig) (*Experiment, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	exp := &Experiment{
		ID:                generateExperimentID(),
		Name:              config.Name,
		Description:       config.Description,
		Hypothesis:        config.Hypothesis,
		Status:            StatusDraft,
		Type:              config.Type,
		Variants:          make([]Variant, len(config.Variants)),
		TrafficAllocation: config.TrafficAllocation,
		PrimaryMetric:     config.PrimaryMetric,
		SecondaryMetrics:  config.SecondaryMetrics,
		SampleSize:        config.SampleSize,
		Duration:          config.Duration,
		Segments:          config.Segments,
		Randomization:     config.Randomization,
	}

	// Copy variants
	for i, v := range config.Variants {
		exp.Variants[i] = Variant{
			ID:             fmt.Sprintf("variant-%d", i),
			Name:           v.Name,
			Description:    v.Description,
			TrafficPercent: v.TrafficPercent,
			Config:         v.Config,
			IsControl:      v.IsControl,
			Metrics:        VariantMetrics{CustomMetrics: make(map[string]float64)},
		}
	}

	m.mu.Lock()
	m.experiments[exp.ID] = exp
	m.mu.Unlock()

	if m.storage != nil {
		_ = m.storage.Save(exp)
	}

	return exp, nil
}

// Start begins an experiment
func (m *Manager) Start(ctx context.Context, id string) error {
	m.mu.Lock()
	exp, exists := m.experiments[id]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("experiment %s not found", id)
	}

	exp.mu.Lock()
	defer exp.mu.Unlock()

	if exp.Status != StatusDraft && exp.Status != StatusScheduled {
		return fmt.Errorf("experiment cannot be started from status %s", exp.Status)
	}

	exp.Status = StatusRunning
	exp.StartTime = time.Now()
	exp.EndTime = exp.StartTime.Add(exp.Duration)

	if m.storage != nil {
		_ = m.storage.Save(exp)
	}

	return nil
}

// AssignVariant assigns a user to a variant
func (m *Manager) AssignVariant(ctx context.Context, experimentID, userID string) (*Variant, error) {
	m.mu.RLock()
	exp, exists := m.experiments[experimentID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	exp.mu.RLock()
	defer exp.mu.RUnlock()

	if exp.Status != StatusRunning {
		return nil, fmt.Errorf("experiment is not running")
	}

	// Check if user already assigned
	assignmentKey := fmt.Sprintf("%s:%s", experimentID, userID)
	m.mu.RLock()
	variantID, assigned := m.assignments[assignmentKey]
	m.mu.RUnlock()

	if assigned {
		return m.getVariant(exp, variantID)
	}

	// Assign based on randomization method
	variant := m.assignUser(exp, userID)

	m.mu.Lock()
	m.assignments[assignmentKey] = variant.ID
	m.mu.Unlock()

	return variant, nil
}

func (m *Manager) assignUser(exp *Experiment, userID string) *Variant {
	switch exp.Randomization {
	case RandomizationUserID:
		return m.assignByUserID(exp, userID)
	case RandomizationWeighted:
		return m.assignWeighted(exp)
	default:
		return m.assignRandom(exp)
	}
}

func (m *Manager) assignByUserID(exp *Experiment, userID string) *Variant {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", exp.ID, userID)))
	hashValue := hex.EncodeToString(hash[:])

	// Use first 8 chars as hex number
	var num uint64
	fmt.Sscanf(hashValue[:8], "%x", &num)

	randomValue := float64(num) / float64(^uint64(0))

	cumulative := 0.0
	for i := range exp.Variants {
		cumulative += exp.Variants[i].TrafficPercent / 100.0
		if randomValue <= cumulative {
			return &exp.Variants[i]
		}
	}

	return &exp.Variants[0]
}

func (m *Manager) assignRandom(exp *Experiment) *Variant {
	randomValue := rand.Float64()
	cumulative := 0.0

	for i := range exp.Variants {
		cumulative += exp.Variants[i].TrafficPercent / 100.0
		if randomValue <= cumulative {
			return &exp.Variants[i]
		}
	}

	return &exp.Variants[0]
}

func (m *Manager) assignWeighted(exp *Experiment) *Variant {
	// Thompson sampling for multi-armed bandit
	// Simplified implementation
	bestVariant := &exp.Variants[0]
	bestScore := 0.0

	for i := range exp.Variants {
		v := &exp.Variants[i]
		if v.Metrics.Impressions == 0 {
			return v // Explore new variants
		}

		// Upper confidence bound
		conversionRate := v.Metrics.ConversionRate
		confidence := math.Sqrt(2 * math.Log(float64(v.Metrics.Impressions)) / float64(v.Metrics.Impressions))
		score := conversionRate + confidence

		if score > bestScore {
			bestScore = score
			bestVariant = v
		}
	}

	return bestVariant
}

func (m *Manager) getVariant(exp *Experiment, variantID string) (*Variant, error) {
	for i := range exp.Variants {
		if exp.Variants[i].ID == variantID {
			return &exp.Variants[i], nil
		}
	}
	return nil, fmt.Errorf("variant %s not found", variantID)
}

// RecordEvent records an event for a user in an experiment
func (m *Manager) RecordEvent(ctx context.Context, experimentID, userID, eventType string, value float64) error {
	m.mu.RLock()
	exp, exists := m.experiments[experimentID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("experiment %s not found", experimentID)
	}

	// Get user's variant
	assignmentKey := fmt.Sprintf("%s:%s", experimentID, userID)
	m.mu.RLock()
	variantID, assigned := m.assignments[assignmentKey]
	m.mu.RUnlock()

	if !assigned {
		return fmt.Errorf("user %s not assigned to experiment", userID)
	}

	exp.mu.Lock()
	defer exp.mu.Unlock()

	// Find and update variant metrics
	for i := range exp.Variants {
		if exp.Variants[i].ID == variantID {
			v := &exp.Variants[i]
			v.Metrics.Impressions++

			switch eventType {
			case "conversion":
				v.Metrics.Conversions++
				v.Metrics.ConversionRate = float64(v.Metrics.Conversions) / float64(v.Metrics.Impressions)
			case "revenue":
				v.Metrics.Revenue += value
			case "engagement":
				v.Metrics.Engagement += value
			default:
				v.Metrics.CustomMetrics[eventType] += value
			}

			break
		}
	}

	return nil
}

// Analyze performs statistical analysis on experiment results
func (m *Manager) Analyze(ctx context.Context, id string) (*Results, error) {
	m.mu.RLock()
	exp, exists := m.experiments[id]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("experiment %s not found", id)
	}

	exp.mu.Lock()
	defer exp.mu.Unlock()

	results := &Results{
		StartTime:           exp.StartTime,
		EndTime:             time.Now(),
		PValues:             make(map[string]float64),
		EffectSizes:         make(map[string]float64),
		ConfidenceIntervals: make(map[string]CI),
	}

	// Find control and treatment variants
	var control, treatment *Variant
	for i := range exp.Variants {
		if exp.Variants[i].IsControl {
			control = &exp.Variants[i]
		} else if treatment == nil {
			treatment = &exp.Variants[i]
		}
	}

	if control == nil || treatment == nil {
		return nil, fmt.Errorf("experiment needs at least one control and one treatment variant")
	}

	// Calculate metrics
	results.TotalParticipants = control.Metrics.Impressions + treatment.Metrics.Impressions

	// Two-proportion z-test for conversion rate
	p1 := control.Metrics.ConversionRate
	p2 := treatment.Metrics.ConversionRate
	n1 := float64(control.Metrics.Impressions)
	n2 := float64(treatment.Metrics.Impressions)

	pooled := (p1*n1 + p2*n2) / (n1 + n2)
	se := math.Sqrt(pooled * (1 - pooled) * (1/n1 + 1/n2))

	if se > 0 {
		z := (p2 - p1) / se
		results.PValues["conversion_rate"] = calculatePValue(z)
		results.EffectSizes["conversion_rate"] = (p2 - p1) / math.Sqrt(pooled*(1-pooled))
	}

	// Determine winner
	if results.PValues["conversion_rate"] < 0.05 && p2 > p1 {
		results.Winner = treatment
		results.Recommendation = "Treatment variant shows statistically significant improvement"
	} else {
		results.Winner = control
		results.Recommendation = "No significant difference detected; stick with control"
	}

	results.ConfidenceLevel = 0.95
	exp.Results = results

	return results, nil
}

// Stop ends an experiment
func (m *Manager) Stop(ctx context.Context, id string) error {
	m.mu.Lock()
	exp, exists := m.experiments[id]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("experiment %s not found", id)
	}

	exp.mu.Lock()
	defer exp.mu.Unlock()

	if exp.Status != StatusRunning {
		return fmt.Errorf("experiment is not running")
	}

	exp.Status = StatusStopped
	exp.EndTime = time.Now()

	if m.storage != nil {
		_ = m.storage.Save(exp)
	}

	return nil
}

// GetExperiment returns experiment details
func (m *Manager) GetExperiment(id string) (*Experiment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	exp, exists := m.experiments[id]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", id)
	}

	return exp, nil
}

// ListExperiments returns all experiments
func (m *Manager) ListExperiments() []*Experiment {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Experiment, 0, len(m.experiments))
	for _, exp := range m.experiments {
		result = append(result, exp)
	}

	return result
}

// ExperimentConfig holds experiment configuration
type ExperimentConfig struct {
	Name              string
	Description       string
	Hypothesis        string
	Type              ExperimentType
	Variants          []VariantConfig
	TrafficAllocation float64
	PrimaryMetric     string
	SecondaryMetrics  []string
	SampleSize        int
	Duration          time.Duration
	Segments          []Segment
	Randomization     RandomizationMethod
}

// VariantConfig holds variant configuration
type VariantConfig struct {
	Name           string
	Description    string
	TrafficPercent float64
	Config         map[string]interface{}
	IsControl      bool
}

func validateConfig(config ExperimentConfig) error {
	if config.Name == "" {
		return fmt.Errorf("experiment name is required")
	}
	if len(config.Variants) < 2 {
		return fmt.Errorf("at least 2 variants are required")
	}

	totalTraffic := 0.0
	hasControl := false
	for _, v := range config.Variants {
		totalTraffic += v.TrafficPercent
		if v.IsControl {
			hasControl = true
		}
	}

	if math.Abs(totalTraffic-100.0) > 0.01 {
		return fmt.Errorf("variant traffic percentages must sum to 100%%")
	}

	if !hasControl {
		return fmt.Errorf("at least one control variant is required")
	}

	return nil
}

func generateExperimentID() string {
	return fmt.Sprintf("exp-%d", time.Now().Unix())
}

func calculatePValue(z float64) float64 {
	// Simplified p-value calculation
	// In production, use a proper statistical library
	return 2 * (1 - normalCDF(math.Abs(z)))
}

func normalCDF(x float64) float64 {
	// Approximation of normal CDF
	return 0.5 * (1 + math.Erf(x/math.Sqrt(2)))
}

// InMemoryStorage provides in-memory storage for experiments
type InMemoryStorage struct {
	data map[string]*Experiment
	mu   sync.RWMutex
}

// NewInMemoryStorage creates in-memory storage
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]*Experiment),
	}
}

// Save saves an experiment
func (s *InMemoryStorage) Save(exp *Experiment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[exp.ID] = exp
	return nil
}

// Load loads an experiment
func (s *InMemoryStorage) Load(id string) (*Experiment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	exp, exists := s.data[id]
	if !exists {
		return nil, fmt.Errorf("experiment not found")
	}
	return exp, nil
}

// List lists all experiments
func (s *InMemoryStorage) List() ([]*Experiment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Experiment, 0, len(s.data))
	for _, exp := range s.data {
		result = append(result, exp)
	}
	return result, nil
}
