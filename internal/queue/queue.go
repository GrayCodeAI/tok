package queue

import (
	"context"
	"sync"
)

type Job struct {
	ID      string
	Payload interface{}
	Status  string
}

type Queue struct {
	mu        sync.RWMutex
	jobs      map[string]*Job
	pending   []*Job
	processed map[string]bool
}

func NewQueue() *Queue {
	return &Queue{
		jobs:      make(map[string]*Job),
		pending:   make([]*Job, 0),
		processed: make(map[string]bool),
	}
}

func (q *Queue) Enqueue(ctx context.Context, job *Job) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs[job.ID] = job
	q.pending = append(q.pending, job)
}

func (q *Queue) Dequeue(ctx context.Context) *Job {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.pending) == 0 {
		return nil
	}

	job := q.pending[0]
	q.pending = q.pending[1:]
	return job
}

func (q *Queue) Complete(ctx context.Context, jobID string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.processed[jobID] = true
}

func (q *Queue) GetStats() (pending, processed int) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.pending), len(q.processed)
}
