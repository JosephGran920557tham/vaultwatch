package tokenwatch

import (
	"context"
	"sync"
)

// Queue is a bounded, thread-safe FIFO queue of Envelopes.
type Queue struct {
	mu      sync.Mutex
	items   []Envelope
	cap     int
	notify  chan struct{}
}

// NewQueue creates a Queue with the given capacity.
// If cap is <= 0 it defaults to 64.
func NewQueue(cap int) *Queue {
	if cap <= 0 {
		cap = 64
	}
	return &Queue{
		cap:    cap,
		notify: make(chan struct{}, 1),
	}
}

// Push enqueues an envelope. Returns false if the queue is full.
func (q *Queue) Push(env Envelope) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) >= q.cap {
		return false
	}
	q.items = append(q.items, env)
	select {
	case q.notify <- struct{}{}:
	default:
	}
	return true
}

// Pop removes and returns the oldest envelope.
// Returns (Envelope{}, false) when the queue is empty.
func (q *Queue) Pop() (Envelope, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return Envelope{}, false
	}
	env := q.items[0]
	q.items = q.items[1:]
	return env, true
}

// Len returns the current number of items in the queue.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// Wait blocks until an item is available or the context is cancelled.
func (q *Queue) Wait(ctx context.Context) error {
	for {
		q.mu.Lock()
		ready := len(q.items) > 0
		q.mu.Unlock()
		if ready {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-q.notify:
		}
	}
}
