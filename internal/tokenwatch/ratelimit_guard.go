package tokenwatch

import (
	"errors"
	"sync"
	"time"
)

// RateLimitGuard enforces a maximum number of token check operations
// per time window across all tokens, providing a global check-rate ceiling.
type RateLimitGuard struct {
	mu       sync.Mutex
	window   time.Duration
	maxOps   int
	counts   map[string][]time.Time
	now      func() time.Time
}

// RateLimitGuardConfig holds configuration for a RateLimitGuard.
type RateLimitGuardConfig struct {
	// Window is the rolling time window for rate limiting.
	Window time.Duration
	// MaxOps is the maximum number of operations allowed per token per window.
	MaxOps int
}

// DefaultRateLimitGuardConfig returns a sensible default configuration.
func DefaultRateLimitGuardConfig() RateLimitGuardConfig {
	return RateLimitGuardConfig{
		Window: 1 * time.Minute,
		MaxOps: 10,
	}
}

// NewRateLimitGuard creates a new RateLimitGuard with the given config.
// Returns an error if MaxOps < 1 or Window <= 0.
func NewRateLimitGuard(cfg RateLimitGuardConfig) (*RateLimitGuard, error) {
	if cfg.MaxOps < 1 {
		return nil, errors.New("ratelimitguard: MaxOps must be at least 1")
	}
	if cfg.Window <= 0 {
		return nil, errors.New("ratelimitguard: Window must be positive")
	}
	return &RateLimitGuard{
		window: cfg.Window,
		maxOps: cfg.MaxOps,
		counts: make(map[string][]time.Time),
		now:    time.Now,
	}, nil
}

// Allow returns true if the given token ID is permitted to proceed,
// false if it has exceeded its rate limit for the current window.
func (g *RateLimitGuard) Allow(tokenID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := g.now()
	cutoff := now.Add(-g.window)

	times := g.counts[tokenID]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= g.maxOps {
		g.counts[tokenID] = filtered
		return false
	}

	g.counts[tokenID] = append(filtered, now)
	return true
}

// Reset clears the rate limit state for a specific token ID.
func (g *RateLimitGuard) Reset(tokenID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.counts, tokenID)
}
