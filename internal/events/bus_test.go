package events

import (
	"testing"
	"time"
)

func TestNewEventBus(t *testing.T) {
	eb := NewEventBus()
	if eb == nil {
		t.Error("Expected non-nil event bus")
	}
}

func TestEventBusSubscribe(t *testing.T) {
	eb := NewEventBus()
	ch := eb.Subscribe("test-event")

	if ch == nil {
		t.Error("Expected non-nil channel")
	}
}

func TestEventBusPublish(t *testing.T) {
	eb := NewEventBus()
	ch := eb.Subscribe("test-event")

	event := Event{
		ID:        "e1",
		Type:      "test-event",
		Payload:   "test",
		Timestamp: time.Now(),
	}

	eb.Publish(nil, event)

	select {
	case received := <-ch:
		if received.ID != "e1" {
			t.Errorf("Expected e1, got %s", received.ID)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for event")
	}
}

func TestEventBusUnsubscribe(t *testing.T) {
	eb := NewEventBus()
	ch := eb.Subscribe("test-event")

	eb.Unsubscribe("test-event", ch)
}
