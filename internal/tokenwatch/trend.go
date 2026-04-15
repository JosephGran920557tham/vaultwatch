package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultTrendConfig returns a TrendConfig with sensible defaults.
func DefaultTrendConfig() TrendConfig {
	return TrendConfig{
		SampleWindow: 5 * time.Minute,
		MinSamples:   3,
		DropThreshold: 0.25, // 25% TTL drop triggers warning
	}
}

// TrendConfig controls TTL trend detection behaviour.
type TrendConfig struct {
	SampleWindow  time.Duration
	MinSamples    int
	DropThreshold float64
}

// trendSample is a single TTL observation.
type trendSample struct {
	at  time.Time
	ttl time.Duration
}

// TrendDetector tracks TTL samples per token and detects declining trends.
type TrendDetector struct {
	mu      sync.Mutex
	cfg     TrendConfig
	samples map[string][]trendSample
}

// NewTrendDetector creates a TrendDetector with the given config.
// Zero values in cfg are replaced with defaults.
func NewTrendDetector(cfg TrendConfig) *TrendDetector {
	def := DefaultTrendConfig()
	if cfg.SampleWindow <= 0 {
		cfg.SampleWindow = def.SampleWindow
	}
	if cfg.MinSamples <= 0 {
		cfg.MinSamples = def.MinSamples
	}
	if cfg.DropThreshold <= 0 {
		cfg.DropThreshold = def.DropThreshold
	}
	return &TrendDetector{
		cfg:     cfg,
		samples: make(map[string][]trendSample),
	}
}

// Record adds a TTL observation for the given token.
func (d *TrendDetector) Record(tokenID string, ttl time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-d.cfg.SampleWindow)
	prev := d.samples[tokenID]
	filtered := prev[:0]
	for _, s := range prev {
		if s.at.After(cutoff) {
			filtered = append(filtered, s)
		}
	}
	d.samples[tokenID] = append(filtered, trendSample{at: now, ttl: ttl})
}

// Check returns a warning alert if the token's TTL is trending downward
// beyond the configured threshold, or nil otherwise.
func (d *TrendDetector) Check(tokenID string) *alert.Alert {
	d.mu.Lock()
	defer d.mu.Unlock()
	samples := d.samples[tokenID]
	if len(samples) < d.cfg.MinSamples {
		return nil
	}
	first := samples[0].ttl
	last := samples[len(samples)-1].ttl
	if first <= 0 {
		return nil
	}
	drop := float64(first-last) / float64(first)
	if drop < d.cfg.DropThreshold {
		return nil
	}
	return &alert.Alert{
		LeaseID: tokenID,
		Level:   alert.Warning,
		Message: fmt.Sprintf("token TTL dropped %.0f%% over last %s", drop*100, d.cfg.SampleWindow),
	}
}
