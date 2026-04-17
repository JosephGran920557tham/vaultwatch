package tokenwatch

import (
	"sync"
	"time"
)

// DefaultHeartbeatConfig returns sensible defaults for HeartbeatDetector.
func DefaultHeartbeatConfig() HeartbeatConfig {
	return HeartbeatConfig{
		StaleAfter:   5 * time.Minute,
		CriticalAfter: 15 * time.Minute,
	}
}

// HeartbeatConfig controls thresholds for missed heartbeats.
type HeartbeatConfig struct {
	StaleAfter    time.Duration
	CriticalAfter time.Duration
}

// HeartbeatDetector tracks the last seen time for each token and raises
// alerts when a token has not been observed within the configured window.
type HeartbeatDetector struct {
	cfg  HeartbeatConfig
	mu   sync.Mutex
	seen map[string]time.Time
}

// NewHeartbeatDetector creates a HeartbeatDetector with the given config.
// Zero values are replaced with defaults.
func NewHeartbeatDetector(cfg HeartbeatConfig) *HeartbeatDetector {
	def := DefaultHeartbeatConfig()
	if cfg.StaleAfter <= 0 {
		cfg.StaleAfter = def.StaleAfter
	}
	if cfg.CriticalAfter <= 0 {
		cfg.CriticalAfter = def.CriticalAfter
	}
	return &HeartbeatDetector{
		cfg:  cfg,
		seen: make(map[string]time.Time),
	}
}

// Beat records that a token was observed at the current time.
func (h *HeartbeatDetector) Beat(tokenID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seen[tokenID] = time.Now()
}

// Check returns an alert if the token has missed its heartbeat window,
// or nil if the token is current.
func (h *HeartbeatDetector) Check(tokenID string, now time.Time) *Alert {
	h.mu.Lock()
	last, ok := h.seen[tokenID]
	h.mu.Unlock()

	if !ok {
		return buildHeartbeatAlert(tokenID, "never seen", LevelCritical)
	}
	since := now.Sub(last)
	switch {
	case since >= h.cfg.CriticalAfter:
		return buildHeartbeatAlert(tokenID, "heartbeat critical", LevelCritical)
	case since >= h.cfg.StaleAfter:
		return buildHeartbeatAlert(tokenID, "heartbeat warning", LevelWarning)
	default:
		return nil
	}
}

func buildHeartbeatAlert(tokenID, msg string, level Level) *Alert {
	return &Alert{
		LeaseID: tokenID,
		Message: msg,
		Level:   level,
	}
}
