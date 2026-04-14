package tokenwatch

import (
	"sync"
	"time"
)

// ThrottleConfig controls how often repeat alerts are suppressed for a token.
type ThrottleConfig struct {
	// MinInterval is the minimum duration between repeated alerts for the same token.
	MinInterval time.Duration
}

// DefaultThrottleConfig returns a ThrottleConfig with sensible defaults.
func DefaultThrottleConfig() ThrottleConfig {
	return ThrottleConfig{
		MinInterval: 5 * time.Minute,
	}
}

// Throttle suppresses duplicate alerts for the same token within a time window.
type Throttle struct {
	cfg  ThrottleConfig
	mu   sync.Mutex
	last map[string]time.Time
}

// NewThrottle creates a new Throttle with the given config.
func NewThrottle(cfg ThrottleConfig) *Throttle {
	if cfg.MinInterval <= 0 {
		cfg.MinInterval = DefaultThrottleConfig().MinInterval
	}
	return &Throttle{
		cfg:  cfg,
		last: make(map[string]time.Time),
	}
}

// Allow returns true if an alert for the given token ID should be forwarded.
// It suppresses repeated alerts within MinInterval.
func (t *Throttle) Allow(tokenID string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if prev, ok := t.last[tokenID]; ok {
		if now.Sub(prev) < t.cfg.MinInterval {
			return false
		}
	}
	t.last[tokenID] = now
	return true
}

// Reset clears the throttle state for a specific token.
func (t *Throttle) Reset(tokenID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.last, tokenID)
}

// ResetAll clears all throttle state.
func (t *Throttle) ResetAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.last = make(map[string]time.Time)
}
