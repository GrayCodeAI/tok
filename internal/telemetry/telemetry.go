package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Event represents a telemetry event.
type Event struct {
	Type       string                 `json:"type"`
	Timestamp  time.Time              `json:"timestamp"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Telemetry collects and batches telemetry events for async sending.
type Telemetry struct {
	mu       sync.Mutex
	events   []Event
	maxBatch int
	maxTotal int // absolute cap to prevent OOM
	enabled  bool
	output   string // file path for local output (debug/testing)
}

// New creates a new Telemetry instance.
func New(enabled bool, maxBatch int) *Telemetry {
	const defaultMaxTotal = 10000 // cap total buffered events to prevent OOM
	maxTotal := maxBatch * 10
	if maxTotal < defaultMaxTotal {
		maxTotal = defaultMaxTotal
	}
	return &Telemetry{
		events:   make([]Event, 0, maxBatch),
		maxBatch: maxBatch,
		maxTotal: maxTotal,
		enabled:  enabled,
	}
}

// NewDefault creates a telemetry instance with default settings.
func NewDefault() *Telemetry {
	return New(os.Getenv("TOKMAN_TELEMETRY") != "false", 100)
}

// Record records a telemetry event.
// If the buffer is full, the event is dropped (no backpressure).
// This prevents OOM when the flush backend is slow or unavailable.
func (t *Telemetry) Record(eventType string, props map[string]interface{}) {
	if !t.enabled {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Drop events if buffer exceeds maxTotal (rate limiting via backpressure-free discard)
	if len(t.events) >= t.maxTotal {
		// Drop oldest 50% to make room for newer events
		half := len(t.events) / 2
		copy(t.events, t.events[half:])
		t.events = t.events[:half]
	}

	event := Event{
		Type:       eventType,
		Timestamp:  time.Now(),
		Properties: props,
	}
	t.events = append(t.events, event)

	if len(t.events) >= t.maxBatch {
		t.flushLocked()
	}
}

// Flush flushes pending events.
func (t *Telemetry) Flush() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.flushLocked()
}

func (t *Telemetry) flushLocked() {
	if len(t.events) == 0 {
		return
	}

	if t.output != "" {
		t.writeToFile()
	}

	t.events = t.events[:0]
}

func (t *Telemetry) writeToFile() {
	dir := filepath.Dir(t.output)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return
	}

	f, err := os.OpenFile(t.output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, event := range t.events {
		if err := enc.Encode(event); err != nil {
			return
		}
	}
}

// SetOutput sets the output file path for telemetry events.
func (t *Telemetry) SetOutput(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output = path
}

// EventCount returns the number of pending events.
func (t *Telemetry) EventCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.events)
}

// IsEnabled returns whether telemetry is enabled.
func (t *Telemetry) IsEnabled() bool {
	return t.enabled
}

// CommandTelemetryEvent records a command execution telemetry event.
func (t *Telemetry) CommandTelemetryEvent(command string, savedTokens int, execTimeMs int64) {
	t.Record("command_executed", map[string]interface{}{
		"command":      command,
		"saved_tokens": savedTokens,
		"exec_time_ms": execTimeMs,
	})
}

// FilterTelemetryEvent records a filter pipeline telemetry event.
func (t *Telemetry) FilterTelemetryEvent(originalTokens, filteredTokens int, layers map[string]int) {
	t.Record("filter_pipeline", map[string]interface{}{
		"original_tokens": originalTokens,
		"filtered_tokens": filteredTokens,
		"reduction_pct":   float64(originalTokens-filteredTokens) / float64(originalTokens) * 100,
		"layers":          layers,
	})
}

// String returns a summary of the telemetry state.
func (t *Telemetry) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return fmt.Sprintf("Telemetry{enabled=%v, pending=%d, maxBatch=%d}", t.enabled, len(t.events), t.maxBatch)
}
