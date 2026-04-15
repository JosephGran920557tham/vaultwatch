package tokenwatch

import (
	"errors"
	"sync"
	"time"
)

// QuotaConfig defines limits for token alert emission per time window.
type QuotaConfig struct {
	// MaxAlertsPerWindow is the maximum number of alerts allowed in the window.
	MaxAlertsPerWindow int
	// Window is the rolling time window duration.
	Window time.Duration
}

// DefaultQuotaConfig returns a sensible default quota configuration.
func DefaultQuotaConfig() QuotaConfig {
	return QuotaConfig{
		MaxAlertsPerWindow: 20,
		Window:             time.Minute,
	}
}

// Quota enforces a per-token alert emission rate limit using a sliding window.
type Quota struct {
	mu     sync.Mutex
	cfg    QuotaConfig
	bucket map[string][]time.Time
}

// NewQuota creates a new Quota enforcer. Returns an error for invalid config.
func NewQuota(cfg QuotaConfig) (*Quota, error) {
	if cfg.MaxAlertsPerWindow <= 0 {
		return nil, errors.New("quota: MaxAlertsPerWindow must be positive")
	}
	if cfg.Window <= 0 {
		return nil, errors.New("quota: Window must be positive")
	}
	return &Quota{
		cfg:    cfg,
		bucket: make(map[string][]time.Time),
	}, nil
}

// Allow returns true if the given token is permitted to emit another alert.
// It prunes expired timestamps and records the current emission if allowed.
func (q *Quota) Allow(tokenID string) bool {
	now := time.Now()
	q.mu.Lock()
	defer q.mu.Unlock()

	times := q.prune(q.bucket[tokenID], now)
	if len(times) >= q.cfg.MaxAlertsPerWindow {
		q.bucket[tokenID] = times
		return false
	}
	q.bucket[tokenID] = append(times, now)
	return true
}

// Remaining returns how many more alerts the token may emit in the current window.
func (q *Quota) Remaining(tokenID string) int {
	now := time.Now()
	q.mu.Lock()
	defer q.mu.Unlock()

	times := q.prune(q.bucket[tokenID], now)
	q.bucket[tokenID] = times
	remaining := q.cfg.MaxAlertsPerWindow - len(times)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (q *Quota) prune(times []time.Time, now time.Time) []time.Time {
	cutoff := now.Add(-q.cfg.Window)
	result := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			result = append(result, t)
		}
	}
	return result
}
