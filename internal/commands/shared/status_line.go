package shared

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// StatusEvent represents a real-time status update during command execution.
type StatusEvent struct {
	Command      string
	Stage        string // "executing", "compressing", "finalizing", "done"
	Layer        string // current layer name (if in compression)
	InputTokens  int
	OutputTokens int
	ProgressPct  float64 // 0-100
	ETA          time.Duration
	Timestamp    time.Time
}

// StatusLine manages real-time status output to stderr.
type StatusLine struct {
	mu        sync.RWMutex
	lastEvent *StatusEvent
	enabled   bool
	verbose   bool
	writer    io.Writer
	cancel    chan struct{}
	wg        sync.WaitGroup
	shutdown  sync.Once
}

var (
	globalStatus *StatusLine
	statusOnce   sync.Once
)

// GetStatusLine returns the singleton status line instance.
func GetStatusLine() *StatusLine {
	statusOnce.Do(func() {
		globalStatus = &StatusLine{
			enabled: isTerminal(), // auto-enable if stdout is terminal
			verbose: false,
			writer:  os.Stderr,
			cancel:  make(chan struct{}),
		}
	})
	return globalStatus
}

// SetEnabled enables or disables the status line.
func (s *StatusLine) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled && isTerminal()
}

// SetVerbose enables or disables verbose status updates.
func (s *StatusLine) SetVerbose(verbose bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.verbose = verbose
}

// SetStatusEnabled enables or disables the status line (package-level helper).
func SetStatusEnabled(enabled bool) {
	GetStatusLine().SetEnabled(enabled)
}

// SetVerbose enables or disables verbose status updates (package-level helper).
func SetVerboseStatus(verbose bool) {
	GetStatusLine().SetVerbose(verbose)
}

// isTerminal checks if stderr is a terminal.
func isTerminal() bool {
	fileInfo, _ := os.Stderr.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// Publish sends a status event (non-blocking).
func (s *StatusLine) Publish(event StatusEvent) {
	s.mu.RLock()
	enabled := s.enabled
	s.mu.RUnlock()
	if !enabled {
		return
	}

	s.mu.Lock()
	s.lastEvent = &event
	s.mu.Unlock()

	s.render(event)
}

// render displays the status line.
func (s *StatusLine) render(event StatusEvent) {
	select {
	case <-s.cancel:
		return
	default:
	}

	// Build status string
	var sb strings.Builder

	// Stage icon
	icon := "⟳"
	switch event.Stage {
	case "executing":
		icon = "⚡"
	case "compressing":
		icon = "🗜️"
	case "finalizing":
		icon = "✨"
	case "done":
		icon = "✓"
	}

	sb.WriteString(fmt.Sprintf("\r%s %s", icon, event.Command))

	// Add layer name if compressing
	if event.Layer != "" {
		sb.WriteString(fmt.Sprintf(" [%s]", event.Layer))
	}

	// Token counts
	if event.InputTokens > 0 {
		sb.WriteString(fmt.Sprintf(" %d→%d", event.InputTokens, event.OutputTokens))

		// Show reduction if we have both
		if event.InputTokens > event.OutputTokens {
			reduction := 100.0 * float64(event.InputTokens-event.OutputTokens) / float64(event.InputTokens)
			sb.WriteString(fmt.Sprintf(" (%.0f%%)", reduction))
		}
	}

	// Progress bar
	if event.ProgressPct > 0 && event.ProgressPct < 100 {
		barWidth := 10
		filled := int(float64(barWidth) * event.ProgressPct / 100.0)
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		sb.WriteString(fmt.Sprintf(" [%s] %.0f%%", bar, event.ProgressPct))
	}

	// ETA
	if event.ETA > 0 && event.Stage != "done" {
		sb.WriteString(fmt.Sprintf(" eta %s", event.ETA.Round(time.Millisecond)))
	}

	// Clear rest of line
	sb.WriteString("    ")

	fmt.Fprint(s.writer, sb.String())
}

// Start begins status line updates for a command.
func (s *StatusLine) Start(command string) {
	s.Publish(StatusEvent{
		Command:   command,
		Stage:     "executing",
		Timestamp: time.Now(),
	})
}

// CompressionProgress updates status during pipeline processing.
func (s *StatusEvent) CompressionProgress(layer string, inputTokens, outputTokens int, progress float64) {
	event := StatusEvent{
		Command:      s.Command,
		Stage:        "compressing",
		Layer:        layer,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		ProgressPct:  progress,
		Timestamp:    time.Now(),
	}
	GetStatusLine().Publish(event)
}

// Done clears the status line.
func (s *StatusLine) Done() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.enabled {
		return
	}
	// Clear line
	fmt.Fprint(s.writer, "\r\x1b[K")
}

// Shutdown gracefully stops the status line and waits for in-flight renders.
func (s *StatusLine) Shutdown() {
	s.shutdown.Do(func() {
		close(s.cancel)
		// Wait for active renders with a timeout to avoid hanging.
		done := make(chan struct{})
		go func() {
			s.wg.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
	})
}

// GetLastEvent returns the most recent status event.
func (s *StatusLine) GetLastEvent() *StatusEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.lastEvent == nil {
		return nil
	}
	// Return a copy
	ev := *s.lastEvent
	return &ev
}

// EnableStatusLine enables status line if terminal is available.
func EnableStatusLine() {
	GetStatusLine().SetEnabled(true)
}

// DisableStatusLine disables status line output.
func DisableStatusLine() {
	GetStatusLine().SetEnabled(false)
}

// IsEnabled returns whether status line is currently enabled.
func IsEnabled() bool {
	if globalStatus == nil {
		return false
	}
	globalStatus.mu.RLock()
	defer globalStatus.mu.RUnlock()
	return globalStatus.enabled
}
