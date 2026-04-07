package queue

import (
	"testing"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue()
	if q == nil {
		t.Error("Expected non-nil queue")
	}
}

func TestQueueEnqueueDequeue(t *testing.T) {
	q := NewQueue()

	q.Enqueue(nil, &Job{ID: "job1", Payload: "test"})
	q.Enqueue(nil, &Job{ID: "job2", Payload: "test2"})

	job := q.Dequeue(nil)
	if job == nil {
		t.Error("Expected non-nil job")
	}
	if job.ID != "job1" {
		t.Errorf("Expected job1, got %s", job.ID)
	}
}

func TestQueueComplete(t *testing.T) {
	q := NewQueue()

	q.Enqueue(nil, &Job{ID: "job1"})
	q.Dequeue(nil)
	q.Complete(nil, "job1")

	_, processed := q.GetStats()
	if processed != 1 {
		t.Errorf("Expected 1 processed, got %d", processed)
	}
}

func TestQueueGetStats(t *testing.T) {
	q := NewQueue()

	q.Enqueue(nil, &Job{ID: "job1"})
	q.Enqueue(nil, &Job{ID: "job2"})

	_, processed := q.GetStats()
	if processed != 0 {
		t.Errorf("Expected 0 processed, got %d", processed)
	}
}

func TestQueueEmpty(t *testing.T) {
	q := NewQueue()
	job := q.Dequeue(nil)
	if job != nil {
		t.Error("Expected nil from empty queue")
	}
}
