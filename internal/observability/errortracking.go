package observability

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

// ErrorEvent represents a tracked error.
type ErrorEvent struct {
	ID              string
	ErrorType       string
	Message         string
	StackTrace      []StackFrame
	Context         map[string]interface{}
	FirstOccurrence time.Time
	LastOccurrence  time.Time
	Occurrences     int
	Fingerprint     string // SHA256 hash for deduplication
	Severity        string
	Source          string // File and line
}

// StackFrame represents a single frame in a stack trace.
type StackFrame struct {
	File     string
	Function string
	Line     int
}

// ErrorTracker tracks and aggregates errors.
type ErrorTracker struct {
	mu             sync.RWMutex
	events         map[string]*ErrorEvent // keyed by fingerprint
	eventHistory   []*ErrorEvent
	maxHistorySize int
	logger         *slog.Logger
	errorChannels  map[string]chan ErrorEvent
}

// NewErrorTracker creates a new error tracker.
func NewErrorTracker(logger *slog.Logger) *ErrorTracker {
	if logger == nil {
		logger = slog.Default()
	}

	return &ErrorTracker{
		events:         make(map[string]*ErrorEvent),
		eventHistory:   make([]*ErrorEvent, 0),
		maxHistorySize: 10000,
		logger:         logger,
		errorChannels:  make(map[string]chan ErrorEvent),
	}
}

// TrackError tracks an error occurrence.
func (et *ErrorTracker) TrackError(err error, context map[string]interface{}, severity string) string {
	if err == nil {
		return ""
	}

	// Get stack trace
	stackTrace := getStackTrace(2)

	// Get source location
	source := getSourceLocation(stackTrace)

	// Create fingerprint
	fingerprint := generateFingerprint(err.Error(), source)

	event := &ErrorEvent{
		ID:              generateErrorID(),
		ErrorType:       fmt.Sprintf("%T", err),
		Message:         err.Error(),
		StackTrace:      stackTrace,
		Context:         context,
		FirstOccurrence: time.Now(),
		LastOccurrence:  time.Now(),
		Occurrences:     1,
		Fingerprint:     fingerprint,
		Severity:        severity,
		Source:          source,
	}

	et.mu.Lock()
	defer et.mu.Unlock()

	// Update or create event
	if existing, exists := et.events[fingerprint]; exists {
		existing.Occurrences++
		existing.LastOccurrence = time.Now()
		existing.Context = context // Update context
		event = existing
	} else {
		et.events[fingerprint] = event
		et.eventHistory = append(et.eventHistory, event)

		// Trim history
		if len(et.eventHistory) > et.maxHistorySize {
			et.eventHistory = et.eventHistory[1:]
		}
	}

	// Log the error
	et.logger.Error("error tracked",
		slog.String("error_id", event.ID),
		slog.String("fingerprint", fingerprint),
		slog.String("message", event.Message),
		slog.String("severity", severity),
		slog.Int("occurrences", event.Occurrences),
	)

	// Send to channel if exists
	if ch, ok := et.errorChannels[event.ErrorType]; ok {
		select {
		case ch <- *event:
		default:
			et.logger.Warn("error channel full", slog.String("error_type", event.ErrorType))
		}
	}

	return event.ID
}

// GetErrorEvents returns tracked error events.
func (et *ErrorTracker) GetErrorEvents(limit int, severityFilter string) []*ErrorEvent {
	et.mu.RLock()
	defer et.mu.RUnlock()

	var events []*ErrorEvent
	for _, event := range et.events {
		if severityFilter == "" || event.Severity == severityFilter {
			events = append(events, event)
		}
	}

	if limit > 0 && len(events) > limit {
		return events[:limit]
	}

	return events
}

// GetErrorHistory returns error history.
func (et *ErrorTracker) GetErrorHistory(limit int) []*ErrorEvent {
	et.mu.RLock()
	defer et.mu.RUnlock()

	if limit > len(et.eventHistory) || limit <= 0 {
		limit = len(et.eventHistory)
	}

	return et.eventHistory[len(et.eventHistory)-limit:]
}

// GetErrorStats returns error statistics.
func (et *ErrorTracker) GetErrorStats() map[string]interface{} {
	et.mu.RLock()
	defer et.mu.RUnlock()

	totalErrors := 0
	totalOccurrences := 0
	bySeverity := make(map[string]int)

	for _, event := range et.events {
		totalErrors++
		totalOccurrences += event.Occurrences
		bySeverity[event.Severity]++
	}

	return map[string]interface{}{
		"unique_errors":     totalErrors,
		"total_occurrences": totalOccurrences,
		"by_severity":       bySeverity,
		"history_size":      len(et.eventHistory),
	}
}

// WatchErrorType subscribes to errors of a specific type.
func (et *ErrorTracker) WatchErrorType(errorType string) <-chan ErrorEvent {
	et.mu.Lock()
	defer et.mu.Unlock()

	if _, exists := et.errorChannels[errorType]; !exists {
		et.errorChannels[errorType] = make(chan ErrorEvent, 10)
	}

	return et.errorChannels[errorType]
}

// getStackTrace captures a stack trace.
func getStackTrace(skip int) []StackFrame {
	var frames []StackFrame
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip+1, pcs)

	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pc)
		frames = append(frames, StackFrame{
			File:     file,
			Function: fn.Name(),
			Line:     line,
		})
	}

	return frames
}

// getSourceLocation extracts source location from stack trace.
func getSourceLocation(stack []StackFrame) string {
	if len(stack) == 0 {
		return "unknown"
	}

	// Find first non-observability frame
	for _, frame := range stack {
		// Skip observability frames
		if frame.Function != "getStackTrace" && frame.Function != "TrackError" {
			return fmt.Sprintf("%s:%d in %s", frame.File, frame.Line, frame.Function)
		}
	}

	return fmt.Sprintf("%s:%d", stack[0].File, stack[0].Line)
}

// generateFingerprint creates a fingerprint for error deduplication.
func generateFingerprint(message string, source string) string {
	h := sha256.New()
	h.Write([]byte(message + source))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func generateErrorID() string {
	return fmt.Sprintf("err_%d", time.Now().UnixNano())
}
