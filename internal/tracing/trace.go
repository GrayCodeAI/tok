// Package tracing provides lightweight distributed tracing for the CLI execution path.
//
// Unlike full OTel, this uses a simple span-based model that flows through
// Cobra command context without external dependencies. Suitable for local
// debugging and production diagnostics.
package tracing

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ctxKey struct{}

// Span represents a single operation in a trace.
type Span struct {
	Name     string
	Start    time.Time
	End      time.Time
	Duration time.Duration
	Error    error
	Attrs    map[string]string
	Children []*Span
	mu       sync.Mutex
}

// SetAttr adds a key-value attribute to the span.
func (s *Span) SetAttr(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Attrs == nil {
		s.Attrs = make(map[string]string)
	}
	s.Attrs[key] = value
}

// Child creates a child span.
func (s *Span) Child(name string) *Span {
	child := &Span{
		Name:  name,
		Start: time.Now(),
		Attrs: make(map[string]string),
	}
	s.mu.Lock()
	s.Children = append(s.Children, child)
	s.mu.Unlock()
	return child
}

// Finish marks the span as complete and records duration.
func (s *Span) Finish(err ...error) {
	s.End = time.Now()
	s.Duration = s.End.Sub(s.Start)
	if len(err) > 0 && err[0] != nil {
		s.Error = err[0]
	}
}

// Trace is the root of a trace, containing one or more root spans.
type Trace struct {
	ID    string
	Spans []*Span
	mu    sync.Mutex
}

// StartSpan creates a new root span in the trace.
func (t *Trace) StartSpan(name string) *Span {
	s := &Span{
		Name:  name,
		Start: time.Now(),
		Attrs: make(map[string]string),
	}
	t.mu.Lock()
	t.Spans = append(t.Spans, s)
	t.mu.Unlock()
	return s
}

// Format returns a human-readable trace summary.
func (t *Trace) Format() string {
	var buf string
	buf += fmt.Sprintf("Trace %s:\n", t.ID)
	for _, s := range t.Spans {
		buf += formatSpan(s, 1)
	}
	return buf
}

func formatSpan(s *Span, depth int) string {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	line := fmt.Sprintf("%s%s: %s", indent, s.Name, s.Duration.Round(time.Millisecond))
	if s.Error != nil {
		line += fmt.Sprintf(" (error: %s)", s.Error)
	}
	if len(s.Attrs) > 0 {
		for k, v := range s.Attrs {
			line += fmt.Sprintf(" %s=%s", k, v)
		}
	}
	line += "\n"
	for _, c := range s.Children {
		line += formatSpan(c, depth+1)
	}
	return line
}

// NewTrace creates a new trace with a unique ID.
func NewTrace(id string) *Trace {
	if id == "" {
		id = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return &Trace{ID: id}
}

// NewContext returns a context with the trace attached.
func NewContext(ctx context.Context, trace *Trace) context.Context {
	return context.WithValue(ctx, ctxKey{}, trace)
}

// FromContext retrieves the trace from a context, or returns nil.
func FromContext(ctx context.Context) *Trace {
	t, _ := ctx.Value(ctxKey{}).(*Trace)
	return t
}

// StartSpanFromContext creates a root span, using the trace from context.
// If no trace is in context, creates and returns a detached span.
func StartSpanFromContext(ctx context.Context, name string) *Span {
	if t := FromContext(ctx); t != nil {
		return t.StartSpan(name)
	}
	return &Span{Name: name, Start: time.Now(), Attrs: make(map[string]string)}
}
