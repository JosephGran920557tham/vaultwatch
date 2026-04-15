package tokenwatch

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DrainConfig holds configuration for the token drain detector.
type DrainConfig struct {
	// SampleWindow is how far back to look when computing drain rate.
	SampleWindow time.Duration
	// DrainThreshold is the TTL-loss-per-second rate that triggers a drain alert.
	DrainThreshold float64
	// MinSamples is the minimum number of observations required before alerting.
	MinSamples int
}

// DefaultDrainConfig returns sensible defaults for DrainConfig.
func DefaultDrainConfig() DrainConfig {
	return DrainConfig{
		SampleWindow:   5 * time.Minute,
		DrainThreshold: 2.0,
		MinSamples:     3,
	}
}

type sample struct {
	at  time.Time
	ttl time.Duration
}

// DrainDetector tracks TTL observations for a token and detects abnormal drain.
type DrainDetector struct {
	cfg     DrainConfig
	mu      sync.Mutex
	samples map[string][]sample
}

// NewDrainDetector constructs a DrainDetector. Zero-value fields in cfg are
// replaced with defaults.
func NewDrainDetector(cfg DrainConfig) (*DrainDetector, error) {
	def := DefaultDrainConfig()
	if cfg.SampleWindow <= 0 {
		cfg.SampleWindow = def.SampleWindow
	}
	if cfg.DrainThreshold <= 0 {
		cfg.DrainThreshold = def.DrainThreshold
	}
	if cfg.MinSamples <= 0 {
		cfg.MinSamples = def.MinSamples
	}
	return &DrainDetector{
		cfg:     cfg,
		samples: make(map[string][]sample),
	}, nil
}

// Record adds a TTL observation for the given token ID.
func (d *DrainDetector) Record(_ context.Context, tokenID string, ttl time.Duration) {
	now := time.Now()
	d.mu.Lock()
	defer d.mu.Unlock()
	cutoff := now.Add(-d.cfg.SampleWindow)
	prev := d.samples[tokenID]
	filtered := prev[:0]
	for _, s := range prev {
		if s.at.After(cutoff) {
			filtered = append(filtered, s)
		}
	}
	filtered = append(filtered, sample{at: now, ttl: ttl})
	d.samples[tokenID] = filtered
}

// IsDraining returns true when the observed TTL loss rate exceeds the threshold.
func (d *DrainDetector) IsDraining(_ context.Context, tokenID string) (bool, float64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	ss := d.samples[tokenID]
	if len(ss) < d.cfg.MinSamples {
		return false, 0, nil
	}
	first, last := ss[0], ss[len(ss)-1]
	elapsed := last.at.Sub(first.at).Seconds()
	if elapsed <= 0 {
		return false, 0, fmt.Errorf("drain: non-positive elapsed time for token %s", tokenID)
	}
	ttlLoss := first.ttl.Seconds() - last.ttl.Seconds()
	rate := ttlLoss / elapsed
	return rate > d.cfg.DrainThreshold, rate, nil
}
