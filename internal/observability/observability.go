package observability

import (
	"fmt"
	"sync"
	"time"
)

type LogLevel string

const (
	LogDebug LogLevel = "DEBUG"
	LogInfo  LogLevel = "INFO"
	LogWarn  LogLevel = "WARN"
	LogError LogLevel = "ERROR"
)

type LogEntry struct {
	Timestamp     time.Time              `json:"timestamp"`
	Level         LogLevel               `json:"level"`
	Message       string                 `json:"message"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Fields        map[string]interface{} `json:"fields,omitempty"`
}

type StructuredLogger struct {
	entries []LogEntry
	mu      sync.Mutex
}

func NewStructuredLogger() *StructuredLogger {
	return &StructuredLogger{}
}

func (l *StructuredLogger) Debug(msg string, fields map[string]interface{}) {
	l.log(LogDebug, msg, fields)
}

func (l *StructuredLogger) Info(msg string, fields map[string]interface{}) {
	l.log(LogInfo, msg, fields)
}

func (l *StructuredLogger) Warn(msg string, fields map[string]interface{}) {
	l.log(LogWarn, msg, fields)
}

func (l *StructuredLogger) Error(msg string, fields map[string]interface{}) {
	l.log(LogError, msg, fields)
}

func (l *StructuredLogger) log(level LogLevel, msg string, fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Fields:    fields,
	})
}

func (l *StructuredLogger) GetEntries() []LogEntry {
	return l.entries
}

func (l *StructuredLogger) Format(entry LogEntry) string {
	return fmt.Sprintf("[%s] %s %s",
		entry.Timestamp.Format(time.RFC3339),
		entry.Level,
		entry.Message)
}

type TraceSpan struct {
	ID         string            `json:"id"`
	ParentID   string            `json:"parent_id,omitempty"`
	Name       string            `json:"name"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Duration   time.Duration     `json:"duration"`
	Status     string            `json:"status"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type DistributedTracer struct {
	spans []*TraceSpan
	mu    sync.Mutex
}

func NewDistributedTracer() *DistributedTracer {
	return &DistributedTracer{}
}

func (t *DistributedTracer) StartSpan(id, name string, parentID string) *TraceSpan {
	span := &TraceSpan{
		ID:         id,
		ParentID:   parentID,
		Name:       name,
		StartTime:  time.Now(),
		Status:     "running",
		Attributes: make(map[string]string),
	}
	t.mu.Lock()
	t.spans = append(t.spans, span)
	t.mu.Unlock()
	return span
}

func (t *DistributedTracer) EndSpan(span *TraceSpan, status string) {
	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)
	span.Status = status
}

func (t *DistributedTracer) AddAttribute(span *TraceSpan, key, value string) {
	span.Attributes[key] = value
}

func (t *DistributedTracer) GetSpans() []*TraceSpan {
	return t.spans
}

type MetricCollector struct {
	counters   map[string]int64
	gauges     map[string]float64
	histograms map[string][]float64
	mu         sync.Mutex
}

func NewMetricCollector() *MetricCollector {
	return &MetricCollector{
		counters:   make(map[string]int64),
		gauges:     make(map[string]float64),
		histograms: make(map[string][]float64),
	}
}

func (m *MetricCollector) Inc(name string, value int64) {
	m.mu.Lock()
	m.counters[name] += value
	m.mu.Unlock()
}

func (m *MetricCollector) Set(name string, value float64) {
	m.mu.Lock()
	m.gauges[name] = value
	m.mu.Unlock()
}

func (m *MetricCollector) Observe(name string, value float64) {
	m.mu.Lock()
	m.histograms[name] = append(m.histograms[name], value)
	m.mu.Unlock()
}

func (m *MetricCollector) GetCounter(name string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.counters[name]
}

func (m *MetricCollector) GetGauge(name string) float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.gauges[name]
}

func (m *MetricCollector) ExportPrometheus() string {
	var result string
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, value := range m.counters {
		result += fmt.Sprintf("# TYPE %s counter\n%s %d\n", name, name, value)
	}
	for name, value := range m.gauges {
		result += fmt.Sprintf("# TYPE %s gauge\n%s %.2f\n", name, name, value)
	}
	return result
}

type SLAReporter struct {
	targetUptime   float64
	targetLatency  time.Duration
	currentUptime  float64
	currentLatency time.Duration
}

func NewSLAReporter(targetUptime float64, targetLatency time.Duration) *SLAReporter {
	return &SLAReporter{
		targetUptime:  targetUptime,
		targetLatency: targetLatency,
	}
}

func (s *SLAReporter) Update(uptime float64, latency time.Duration) {
	s.currentUptime = uptime
	s.currentLatency = latency
}

func (s *SLAReporter) IsHealthy() bool {
	return s.currentUptime >= s.targetUptime && s.currentLatency <= s.targetLatency
}

func (s *SLAReporter) Report() map[string]interface{} {
	return map[string]interface{}{
		"target_uptime":      s.targetUptime,
		"current_uptime":     s.currentUptime,
		"target_latency_ms":  s.targetLatency.Milliseconds(),
		"current_latency_ms": s.currentLatency.Milliseconds(),
		"healthy":            s.IsHealthy(),
	}
}
