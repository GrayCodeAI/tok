package tracing

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewTrace(t *testing.T) {
	tr := NewTrace("test-1")
	if tr.ID != "test-1" {
		t.Errorf("expected test-1, got %s", tr.ID)
	}
}

func TestNewTraceEmptyID(t *testing.T) {
	tr := NewTrace("")
	if tr.ID == "" {
		t.Error("expected non-empty auto-generated ID")
	}
}

func TestSpanBasic(t *testing.T) {
	tr := NewTrace("t1")
	span := tr.StartSpan("test-op")
	span.SetAttr("key", "value")

	time.Sleep(1 * time.Millisecond)
	span.Finish()

	if span.Duration == 0 {
		t.Error("expected non-zero duration")
	}
	if span.Attrs["key"] != "value" {
		t.Errorf("expected key=value, got %s", span.Attrs["key"])
	}
}

func TestSpanWithError(t *testing.T) {
	tr := NewTrace("t1")
	span := tr.StartSpan("failing-op")
	span.Finish(errors.New("something failed"))

	if span.Error == nil {
		t.Error("expected error on span")
	}
}

func TestChildSpan(t *testing.T) {
	tr := NewTrace("t1")
	parent := tr.StartSpan("parent")
	child := parent.Child("child")
	child.Finish()
	parent.Finish()

	if len(parent.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(parent.Children))
	}
	if parent.Children[0].Name != "child" {
		t.Errorf("expected child span named 'child', got %s", parent.Children[0].Name)
	}
}

func TestContextPropagation(t *testing.T) {
	tr := NewTrace("ctx-test")
	ctx := NewContext(context.Background(), tr)

	retrieved := FromContext(ctx)
	if retrieved != tr {
		t.Error("expected same trace from context")
	}

	// Non-trace context returns nil
	emptyCtx := context.Background()
	if FromContext(emptyCtx) != nil {
		t.Error("expected nil from empty context")
	}
}

func TestStartSpanFromContext(t *testing.T) {
	tr := NewTrace("ctx-test")
	ctx := NewContext(context.Background(), tr)

	span := StartSpanFromContext(ctx, "from-ctx")
	if span.Name != "from-ctx" {
		t.Errorf("unexpected span name: %s", span.Name)
	}

	// Without a trace in context, still gets a valid span
	emptyCtx := context.Background()
	span2 := StartSpanFromContext(emptyCtx, "detached")
	if span2.Name != "detached" {
		t.Errorf("unexpected detached span name: %s", span2.Name)
	}
}

func TestTraceFormat(t *testing.T) {
	tr := NewTrace("format-test")
	s := tr.StartSpan("op1")
	s.SetAttr("status", "ok")
	s.Finish()

	output := tr.Format()
	if output == "" {
		t.Error("expected non-empty trace format")
	}
}

func TestConcurrentSpanAccess(t *testing.T) {
	tr := NewTrace("concurrent")
	span := tr.StartSpan("parent")

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			child := span.Child("child")
			child.SetAttr("id", string(rune('0'+id)))
			child.Finish()
			done <- true
		}(i)
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	span.Finish()

	if len(span.Children) != 10 {
		t.Errorf("expected 10 children, got %d", len(span.Children))
	}
}
