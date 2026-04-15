package tokenwatch

import (
	"sync"
	"time"
)

// DefaultBaselineConfig returns a BaselineConfig with sensible defaults.
func DefaultBaselineConfig() BaselineConfig {
	return BaselineConfig{
		SampleWindow: 10,
		DeviationPct: 25.0,
	}
}

// BaselineConfig controls how the BaselineDetector identifies deviations.
type BaselineConfig struct {
	// SampleWindow is the number of historical TTL samples used to compute the baseline mean.
	SampleWindow int
	// DeviationPct is the percentage below the baseline mean that triggers an alert.
	DeviationPct float64
}

// BaselineDetector computes a rolling mean TTL for each token and alerts
// when the current TTL deviates significantly below that baseline.
type BaselineDetector struct {
	cfg     BaselineConfig
	mu      sync.Mutex
	samples map[string][]time.Duration
}

// NewBaselineDetector creates a BaselineDetector with the given config.
// Zero values are replaced with defaults.
func NewBaselineDetector(cfg BaselineConfig) *BaselineDetector {
	def := DefaultBaselineConfig()
	if cfg.SampleWindow <= 0 {
		cfg.SampleWindow = def.SampleWindow
	}
	if cfg.DeviationPct <= 0 {
		cfg.DeviationPct = def.DeviationPct
	}
	return &BaselineDetector{
		cfg:     cfg,
		samples: make(map[string][]time.Duration),
	}
}

// Record adds a TTL sample for the given token ID.
func (d *BaselineDetector) Record(tokenID string, ttl time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	buf := append(d.samples[tokenID], ttl)
	if len(buf) > d.cfg.SampleWindow {
		buf = buf[len(buf)-d.cfg.SampleWindow:]
	}
	d.samples[tokenID] = buf
}

// Check returns a non-nil alert if the current TTL is significantly below
// the rolling baseline mean. Returns nil when there are insufficient samples.
func (d *BaselineDetector) Check(tokenID string, current time.Duration) *Alert {
	d.mu.Lock()
	samples := d.samples[tokenID]
	d.mu.Unlock()

	if len(samples) < 2 {
		return nil
	}

	var sum time.Duration
	for _, s := range samples {
		sum += s
	}
	mean := sum / time.Duration(len(samples))
	threshold := time.Duration(float64(mean) * (1 - d.cfg.DeviationPct/100))

	if current >= threshold {
		return nil
	}
	return &Alert{
		LeaseID:   tokenID,
		Level:     LevelWarning,
		Message:   "token TTL is significantly below baseline mean",
		ExpiresAt: time.Now().Add(current),
	}
}
