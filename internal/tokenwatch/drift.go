package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultDriftConfig returns a DriftConfig with sensible defaults.
func DefaultDriftConfig() DriftConfig {
	return DriftConfig{
		WarningThreshold:  10 * time.Second,
		CriticalThreshold: 30 * time.Second,
	}
}

// DriftConfig controls how much clock drift is tolerated before alerting.
type DriftConfig struct {
	WarningThreshold  time.Duration
	CriticalThreshold time.Duration
}

// DriftDetector compares the observed TTL against an expected TTL to detect
// clock drift or unexpected token mutations between successive checks.
type DriftDetector struct {
	cfg      DriftConfig
	mu       sync.Mutex
	baseline map[string]driftEntry
}

type driftEntry struct {
	recordedAt  time.Time
	observedTTL time.Duration
}

// NewDriftDetector creates a DriftDetector. Zero values in cfg are replaced
// with the defaults from DefaultDriftConfig.
func NewDriftDetector(cfg DriftConfig) *DriftDetector {
	def := DefaultDriftConfig()
	if cfg.WarningThreshold <= 0 {
		cfg.WarningThreshold = def.WarningThreshold
	}
	if cfg.CriticalThreshold <= 0 {
		cfg.CriticalThreshold = def.CriticalThreshold
	}
	return &DriftDetector{
		cfg:      cfg,
		baseline: make(map[string]driftEntry),
	}
}

// Record stores the current observed TTL for a token so that the next call to
// Check can compute drift.
func (d *DriftDetector) Record(tokenID string, ttl time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.baseline[tokenID] = driftEntry{recordedAt: time.Now(), observedTTL: ttl}
}

// Check computes the drift between the expected TTL (baseline minus elapsed
// time) and the current observed TTL. It returns nil when drift is within the
// warning threshold.
func (d *DriftDetector) Check(tokenID string, currentTTL time.Duration) *alert.Alert {
	d.mu.Lock()
	entry, ok := d.baseline[tokenID]
	d.mu.Unlock()
	if !ok {
		return nil
	}

	elapsed := time.Since(entry.recordedAt)
	expected := entry.observedTTL - elapsed
	drift := currentTTL - expected
	if drift < 0 {
		drift = -drift
	}

	switch {
	case drift >= d.cfg.CriticalThreshold:
		return &alert.Alert{
			LeaseID:  tokenID,
			Level:    alert.Critical,
			Message:  fmt.Sprintf("token TTL drift of %s exceeds critical threshold (%s)", drift.Round(time.Millisecond), d.cfg.CriticalThreshold),
			ExpireAt: time.Now().Add(currentTTL),
		}
	case drift >= d.cfg.WarningThreshold:
		return &alert.Alert{
			LeaseID:  tokenID,
			Level:    alert.Warning,
			Message:  fmt.Sprintf("token TTL drift of %s exceeds warning threshold (%s)", drift.Round(time.Millisecond), d.cfg.WarningThreshold),
			ExpireAt: time.Now().Add(currentTTL),
		}
	default:
		return nil
	}
}
