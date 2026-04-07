package predictive

import (
	"context"
	"sync"
	"time"
)

type Prediction struct {
	Key        string
	Confidence float64
	ExpiresAt  time.Time
	Reason     string
}

type PredictiveCache interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{})
	Prefetch(ctx context.Context, keys []string)
	GetStats() CacheStats
}

type CacheStats struct {
	PredictionsMade  int64
	PredictionsHit   int64
	PrefetchRequests int64
	PrefetchHitRate  float64
}

type Predictor struct {
	mu         sync.RWMutex
	history    *CommandHistory
	model      PredictionModel
	threshold  float64
	windowSize int
	prefetcher *Prefetcher
	stats      CacheStats
}

type PredictionModel interface {
	Predict(ctx context.Context, features []float64) (float64, error)
	PredictNext(ctx context.Context, features []float64, count int) []Prediction
	Train(ctx context.Context, features []float64, labels []float64) error
}

type CommandHistory struct {
	mu       sync.RWMutex
	commands []CommandEntry
	maxSize  int
}

type CommandEntry struct {
	Timestamp   time.Time
	Command     string
	Args        []string
	Directory   string
	DurationMs  int64
	ResultSize  int
	AccessCount int
}

type Prefetcher struct {
	mu          sync.RWMutex
	predictions map[string]Prediction
	enabled     bool
	maxPrefetch int
	workerCount int
	inputChan   chan string
	resultChan  chan string
}

func NewPredictor(config PredictorConfig) *Predictor {
	p := &Predictor{
		history:    NewCommandHistory(config.HistorySize),
		model:      NewMLModel(config.ModelType),
		threshold:  config.Threshold,
		windowSize: config.WindowSize,
		stats:      CacheStats{},
	}

	if config.PrefetchEnabled {
		p.prefetcher = NewPrefetcher(config.PrefetchWorkers, config.MaxPrefetch)
	}

	return p
}

type PredictorConfig struct {
	HistorySize     int
	WindowSize      int
	Threshold       float64
	ModelType       string
	PrefetchEnabled bool
	PrefetchWorkers int
	MaxPrefetch     int
}

func DefaultPredictorConfig() PredictorConfig {
	return PredictorConfig{
		HistorySize:     1000,
		WindowSize:      50,
		Threshold:       0.7,
		ModelType:       "markov",
		PrefetchEnabled: true,
		PrefetchWorkers: 4,
		MaxPrefetch:     10,
	}
}

func (p *Predictor) PredictNextCommands(ctx context.Context, count int) []Prediction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	features := p.extractFeatures()
	predictions := p.model.PredictNext(ctx, features, count)

	p.stats.PredictionsMade += int64(len(predictions))

	return predictions
}

func (p *Predictor) extractFeatures() []float64 {
	p.history.mu.RLock()
	defer p.history.mu.RUnlock()

	var features []float64

	window := p.history.commands
	if len(window) > p.windowSize {
		window = window[len(window)-p.windowSize:]
	}

	commandFreq := make(map[string]int)
	for _, e := range window {
		commandFreq[e.Command]++
	}

	for _, freq := range commandFreq {
		features = append(features, float64(freq)/float64(len(window)))
	}

	timeOfDay := float64(time.Now().Hour())
	features = append(features, timeOfDay/24.0)

	dayOfWeek := float64(time.Now().Weekday())
	features = append(features, dayOfWeek/7.0)

	return features
}

func (p *Predictor) RecordCommand(entry CommandEntry) {
	p.history.Add(entry)

	if p.prefetcher != nil {
		p.prefetcher.Suggest(entry.Command)
	}
}

func (p *Predictor) GetStats() CacheStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := p.stats
	if stats.PredictionsMade > 0 {
		stats.PrefetchHitRate = float64(stats.PredictionsHit) / float64(stats.PredictionsMade)
	}
	return stats
}

func (p *Predictor) Close() error {
	if p.prefetcher != nil {
		p.prefetcher.Close()
	}
	return nil
}

func NewCommandHistory(maxSize int) *CommandHistory {
	return &CommandHistory{
		commands: make([]CommandEntry, 0, maxSize),
		maxSize:  maxSize,
	}
}

func (h *CommandHistory) Add(entry CommandEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.commands = append(h.commands, entry)

	if len(h.commands) > h.maxSize {
		h.commands = h.commands[1:]
	}
}

func (h *CommandHistory) GetRecent(n int) []CommandEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if n > len(h.commands) {
		n = len(h.commands)
	}

	result := make([]CommandEntry, n)
	copy(result, h.commands[len(h.commands)-n:])
	return result
}

func NewPrefetcher(workers, maxPrefetch int) *Prefetcher {
	p := &Prefetcher{
		predictions: make(map[string]Prediction),
		enabled:     true,
		maxPrefetch: maxPrefetch,
		workerCount: workers,
		inputChan:   make(chan string, 100),
		resultChan:  make(chan string, 100),
	}

	for i := 0; i < workers; i++ {
		go p.worker()
	}

	return p
}

func (p *Prefetcher) Suggest(command string) {
	if !p.enabled {
		return
	}

	select {
	case p.inputChan <- command:
	default:
	}
}

func (p *Prefetcher) worker() {
	for {
		select {
		case cmd := <-p.inputChan:
			predictions := p.generatePredictions(cmd)
			p.mu.Lock()
			for _, pred := range predictions {
				if len(p.predictions) < p.maxPrefetch {
					p.predictions[pred.Key] = pred
				}
			}
			p.mu.Unlock()
		}
	}
}

func (p *Prefetcher) generatePredictions(cmd string) []Prediction {
	return []Prediction{
		{Key: cmd + " --help", Confidence: 0.8, Reason: "common suffix"},
		{Key: cmd + " -v", Confidence: 0.6, Reason: "common flag"},
	}
}

func (p *Prefetcher) Close() {
	p.enabled = false
	close(p.inputChan)
	close(p.resultChan)
}
