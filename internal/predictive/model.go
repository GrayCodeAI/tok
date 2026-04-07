package predictive

import (
	"context"
	"math"
	"sync"
)

type MLModel struct {
	mu           sync.RWMutex
	modelType    string
	markovChain  *MarkovChain
	coefficients []float64
	intercept    float64
	trained      bool
}

type MarkovChain struct {
	mu          sync.RWMutex
	transitions map[string]map[string]int
	order       int
	defaultProb float64
}

func NewMLModel(modelType string) *MLModel {
	m := &MLModel{
		modelType: modelType,
	}

	switch modelType {
	case "markov":
		m.markovChain = NewMarkovChain(2)
	case "linear":
		m.coefficients = make([]float64, 10)
		m.intercept = 0
	}

	return m
}

func (m *MLModel) Predict(ctx context.Context, features []float64) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch m.modelType {
	case "markov":
		return 0.5, nil
	case "linear":
		return m.predictLinear(features), nil
	default:
		return 0.5, nil
	}
}

func (m *MLModel) PredictNext(ctx context.Context, features []float64, count int) []Prediction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch m.modelType {
	case "markov":
		return m.markovChain.PredictNext(features, count)
	case "linear":
		return m.predictLinearMultiple(features, count)
	default:
		return []Prediction{}
	}
}

func (m *MLModel) Train(ctx context.Context, features []float64, labels []float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(features) != len(labels) {
		return nil
	}

	switch m.modelType {
	case "markov":
		m.markovChain.Train(features, labels)
	case "linear":
		m.trainLinear(features, labels)
	}

	m.trained = true
	return nil
}

func (m *MLModel) predictLinear(features []float64) float64 {
	result := m.intercept
	for i, f := range features {
		if i < len(m.coefficients) {
			result += f * m.coefficients[i]
		}
	}

	result = math.Max(0, math.Min(1, result))
	return result
}

func (m *MLModel) predictLinearMultiple(features []float64, count int) []Prediction {
	basePred := m.predictLinear(features)

	predictions := make([]Prediction, 0, count)
	for i := 0; i < count; i++ {
		conf := basePred * (1.0 - float64(i)*0.1)
		predictions = append(predictions, Prediction{
			Key:        "predicted_command",
			Confidence: conf,
		})
	}

	return predictions
}

func (m *MLModel) trainLinear(features []float64, labels []float64) {
	n := float64(len(labels))
	meanX := 0.0
	meanY := 0.0

	for _, x := range features {
		meanX += x
	}
	meanX /= n

	for _, y := range labels {
		meanY += y
	}
	meanY /= n

	numerator := 0.0
	denominator := 0.0
	for i := range labels {
		diffX := features[i] - meanX
		diffY := labels[i] - meanY
		numerator += diffX * diffY
		denominator += diffX * diffX
	}

	if denominator > 0 {
		slope := numerator / denominator
		m.intercept = meanY - slope*meanX
		for i := range m.coefficients {
			m.coefficients[i] = slope
		}
	}
}

func NewMarkovChain(order int) *MarkovChain {
	return &MarkovChain{
		transitions: make(map[string]map[string]int),
		order:       order,
		defaultProb: 0.1,
	}
}

func (mc *MarkovChain) Train(features []float64, labels []float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	for i := range labels {
		if i < len(features) {
			state := stateFromFeature(features[i])
			mc.transitions[state] = mc.transitions[state]
			if mc.transitions[state] == nil {
				mc.transitions[state] = make(map[string]int)
			}
			mc.transitions[state]["next"]++
		}
	}
}

func (mc *MarkovChain) PredictNext(features []float64, count int) []Prediction {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	predictions := make([]Prediction, 0, count)

	for i, f := range features {
		state := stateFromFeature(f)

		if trans, ok := mc.transitions[state]; ok {
			for nextState, count := range trans {
				conf := float64(count) / 10.0
				if conf > 1 {
					conf = 1
				}
				predictions = append(predictions, Prediction{
					Key:        nextState,
					Confidence: conf,
					Reason:     "markov transition",
				})
			}
		}

		if i >= count {
			break
		}
	}

	if len(predictions) == 0 {
		predictions = append(predictions, Prediction{
			Key:        "git status",
			Confidence: 0.5,
			Reason:     "default prediction",
		})
	}

	return predictions[:count]
}

func stateFromFeature(f float64) string {
	bucket := int(f * 10)
	return string(rune('a' + bucket))
}
