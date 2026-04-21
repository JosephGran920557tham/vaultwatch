package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultHorizonConfig returns a HorizonConfig with sensible defaults.
func DefaultHorizonConfig() HorizonConfig {
	return HorizonConfig{
		WarningWindow:  6 * time.Hour,
		CriticalWindow: 1 * time.Hour,
	}
}

// HorizonConfig controls the look-ahead windows for expiry horizon detection.
type HorizonConfig struct {
	WarningWindow  time.Duration
	CriticalWindow time.Duration
}

func (c *HorizonConfig) applyDefaults() {
	d := DefaultHorizonConfig()
	if c.WarningWindow <= 0 {
		c.WarningWindow = d.WarningWindow
	}
	if c.CriticalWindow <= 0 {
		c.CriticalWindow = d.CriticalWindow
	}
}

// HorizonDetector raises alerts when a token's TTL falls within a look-ahead
// window, giving operators advance notice before the expiry threshold is hit.
type HorizonDetector struct {
	cfg  HorizonConfig
	mu   sync.Mutex
	seen map[string]alert.Level
}

// NewHorizonDetector constructs a HorizonDetector. Zero-value config fields
// are replaced with defaults.
func NewHorizonDetector(cfg HorizonConfig) *HorizonDetector {
	cfg.applyDefaults()
	return &HorizonDetector{
		cfg:  cfg,
		seen: make(map[string]alert.Level),
	}
}

// Check returns an alert if the token's remaining TTL falls within a horizon
// window. It suppresses duplicate alerts at the same level.
func (h *HorizonDetector) Check(tokenID string, ttl time.Duration) *alert.Alert {
	var level alert.Level
	switch {
	case ttl <= h.cfg.CriticalWindow:
		level = alert.LevelCritical
	case ttl <= h.cfg.WarningWindow:
		level = alert.LevelWarning
	default:
		h.mu.Lock()
		delete(h.seen, tokenID)
		h.mu.Unlock()
		return nil
	}

	h.mu.Lock()
	prev, ok := h.seen[tokenID]
	if ok && prev == level {
		h.mu.Unlock()
		return nil
	}
	h.seen[tokenID] = level
	h.mu.Unlock()

	return &alert.Alert{
		LeaseID: tokenID,
		Level:   level,
		Message: fmt.Sprintf("token %s expires within %s (TTL: %s)", tokenID, windowLabel(ttl, h.cfg), ttl.Round(time.Second)),
	}
}

func windowLabel(ttl time.Duration, cfg HorizonConfig) string {
	if ttl <= cfg.CriticalWindow {
		return cfg.CriticalWindow.String()
	}
	return cfg.WarningWindow.String()
}
