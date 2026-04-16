package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

// DefaultSpikeConfig returns sensible defaults for spike detection.
func DefaultSpikeConfig() SpikeConfig {
	return SpikeConfig{
		Window:      5 * time.Minute,
		MinSamples:  3,
		MaxIncrease: 0.50, // 50% sudden increase triggers alert
	}
}

// SpikeConfig controls spike detection sensitivity.
type SpikeConfig struct {
	Window      time.Duration
	MinSamples  int
	MaxIncrease float64
}

// SpikeDetector detects sudden upward spikes in token TTL,
// which may indicate unexpected token renewal or rotation.
type SpikeDetector struct {
	mu      sync.Mutex
	cfg     SpikeConfig
	samples map[string][]spikePoint
}

type spikePoint struct {
	at  time.Time
	ttl time.Duration
}

// NewSpikeDetector creates a SpikeDetector with the given config.
// Zero values fall back to defaults.
func NewSpikeDetector(cfg SpikeConfig) *SpikeDetector {
	def := DefaultSpikeConfig()
	if cfg.Window <= 0 {
		cfg.Window = def.Window
	}
	if cfg.MinSamples <= 0 {
		cfg.MinSamples = def.MinSamples
	}
	if cfg.MaxIncrease <= 0 {
		cfg.MaxIncrease = def.MaxIncrease
	}
	return &SpikeDetector{cfg: cfg, samples: make(map[string][]spikePoint)}
}

// Record adds a TTL observation for the given token ID.
func (d *SpikeDetector) Record(tokenID string, ttl time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-d.cfg.Window)
	pts := d.samples[tokenID]
	filtered := pts[:0]
	for _, p := range pts {
		if p.at.After(cutoff) {
			filtered = append(filtered, p)
		}
	}
	d.samples[tokenID] = append(filtered, spikePoint{at: now, ttl: ttl})
}

// Check returns a warning alert if a TTL spike is detected, nil otherwise.
func (d *SpikeDetector) Check(tokenID string, current time.Duration) *alert.Alert {
	d.mu.Lock()
	defer d.mu.Unlock()
	pts := d.samples[tokenID]
	if len(pts) < d.cfg.MinSamples {
		return nil
	}
	prev := pts[len(pts)-2].ttl
	if prev <= 0 {
		return nil
	}
	increase := float64(current-prev) / float64(prev)
	if increase < d.cfg.MaxIncrease {
		return nil
	}
	return &alert.Alert{
		LeaseID:   tokenID,
		Level:     alert.Warning,
		Message:   fmt.Sprintf("token TTL spike detected: %.0f%% increase (prev=%s current=%s)", increase*100, prev.Round(time.Second), current.Round(time.Second)),
		ExpiresAt: time.Now().Add(current),
	}
}
