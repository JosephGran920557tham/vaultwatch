package ratelimit

import (
	"context"
	"fmt"
)

// CheckFunc is a function that performs a single check operation.
type CheckFunc func(ctx context.Context) error

// Middleware wraps a CheckFunc with rate-limiting logic.
type Middleware struct {
	limiter *Limiter
	wait    bool
}

// NewMiddleware creates a Middleware backed by the given Limiter.
// If wait is true, calls block until a token is available; otherwise they
// return ErrRateLimited immediately when no token is available.
func NewMiddleware(l *Limiter, wait bool) (*Middleware, error) {
	if l == nil {
		return nil, fmt.Errorf("ratelimit: limiter must not be nil")
	}
	return &Middleware{limiter: l, wait: wait}, nil
}

// ErrRateLimited is returned when a token is unavailable and wait=false.
var ErrRateLimited = fmt.Errorf("ratelimit: rate limit exceeded")

// Wrap returns a new CheckFunc that enforces the rate limit before calling fn.
func (m *Middleware) Wrap(fn CheckFunc) CheckFunc {
	return func(ctx context.Context) error {
		if m.wait {
			if err := m.limiter.Wait(ctx); err != nil {
				return fmt.Errorf("ratelimit: wait interrupted: %w", err)
			}
		} else {
			if !m.limiter.Allow() {
				return ErrRateLimited
			}
		}
		return fn(ctx)
	}
}
