// Package ratelimit provides a token-bucket rate limiter for controlling
// how frequently alerts and Vault API calls are dispatched.
package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Limiter controls the rate of operations using a token-bucket algorithm.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTick time.Time
	clock    func() time.Time
}

// Config holds parameters for constructing a Limiter.
type Config struct {
	// Rate is the number of tokens replenished per second.
	Rate float64
	// Burst is the maximum number of tokens the bucket can hold.
	Burst float64
}

// New creates a Limiter from cfg. Returns an error if Rate or Burst are non-positive.
func New(cfg Config) (*Limiter, error) {
	if cfg.Rate <= 0 {
		return nil, fmt.Errorf("ratelimit: rate must be positive, got %v", cfg.Rate)
	}
	if cfg.Burst <= 0 {
		return nil, fmt.Errorf("ratelimit: burst must be positive, got %v", cfg.Burst)
	}
	return &Limiter{
		tokens:   cfg.Burst,
		max:      cfg.Burst,
		rate:     cfg.Rate,
		lastTick: time.Now(),
		clock:    time.Now,
	}, nil
}

// Allow reports whether one token is available and consumes it.
// It is safe for concurrent use.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

// Wait blocks until a token is available or ctx is cancelled.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		if l.Allow() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(float64(time.Second) / l.rate)):
		}
	}
}

// refill adds tokens based on elapsed time. Must be called with l.mu held.
func (l *Limiter) refill() {
	now := l.clock()
	elapsed := now.Sub(l.lastTick).Seconds()
	l.tokens = min(l.max, l.tokens+elapsed*l.rate)
	l.lastTick = now
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
