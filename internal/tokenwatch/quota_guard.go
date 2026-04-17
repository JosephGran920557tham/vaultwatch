package tokenwatch

import (
	"fmt"
	"sync"
	"time"
)

// DefaultQuotaGuardConfig returns sensible defaults for QuotaGuard.
func DefaultQuotaGuardConfig() QuotaGuardConfig {
	return QuotaGuardConfig{
		MaxAlerts: 10,
		Window:    time.Minute,
	}
}

// QuotaGuardConfig controls per-token alert quotas within a sliding window.
type QuotaGuardConfig struct {
	MaxAlerts int
	Window    time.Duration
}

type quotaEntry struct {
	count     int
	windowEnd time.Time
}

// QuotaGuard limits the number of alerts emitted per token within a time window.
type QuotaGuard struct {
	mu      sync.Mutex
	cfg     QuotaGuardConfig
	entries map[string]*quotaEntry
}

// NewQuotaGuard creates a QuotaGuard with the given config.
// Zero values are replaced with defaults.
func NewQuotaGuard(cfg QuotaGuardConfig) *QuotaGuard {
	if cfg.MaxAlerts <= 0 {
		cfg.MaxAlerts = DefaultQuotaGuardConfig().MaxAlerts
	}
	if cfg.Window <= 0 {
		cfg.Window = DefaultQuotaGuardConfig().Window
	}
	return &QuotaGuard{cfg: cfg, entries: make(map[string]*quotaEntry)}
}

// Allow returns true if the token has not exceeded its quota in the current window.
func (q *QuotaGuard) Allow(tokenID string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	now := time.Now()
	e, ok := q.entries[tokenID]
	if !ok || now.After(e.windowEnd) {
		q.entries[tokenID] = &quotaEntry{count: 1, windowEnd: now.Add(q.cfg.Window)}
		return true
	}
	if e.count >= q.cfg.MaxAlerts {
		return false
	}
	e.count++
	return true
}

// Reset clears the quota state for a token.
func (q *QuotaGuard) Reset(tokenID string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.entries, tokenID)
}

// String returns a human-readable description of the guard config.
func (q *QuotaGuard) String() string {
	return fmt.Sprintf("QuotaGuard(max=%d window=%s)", q.cfg.MaxAlerts, q.cfg.Window)
}
