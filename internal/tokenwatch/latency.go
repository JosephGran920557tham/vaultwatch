package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultLatencyConfig returns sensible defaults for latency detection.
func DefaultLatencyConfig() LatencyConfig {
	return LatencyConfig{
		WarningThreshold:  200 * time.Millisecond,
		CriticalThreshold: 500 * time.Millisecond,
		MinSamples:        3,
	}
}

// LatencyConfig controls thresholds for token lookup latency alerts.
type LatencyConfig struct {
	WarningThreshold  time.Duration
	CriticalThreshold time.Duration
	MinSamples        int
}

// LatencyDetector tracks round-trip lookup durations per token and
// emits alerts when average latency exceeds configured thresholds.
type LatencyDetector struct {
	cfg     LatencyConfig
	mu      sync.Mutex
	samples map[string][]time.Duration
}

// NewLatencyDetector creates a LatencyDetector with the given config,
// applying defaults for any zero values.
func NewLatencyDetector(cfg LatencyConfig) *LatencyDetector {
	def := DefaultLatencyConfig()
	if cfg.WarningThreshold <= 0 {
		cfg.WarningThreshold = def.WarningThreshold
	}
	if cfg.CriticalThreshold <= 0 {
		cfg.CriticalThreshold = def.CriticalThreshold
	}
	if cfg.MinSamples <= 0 {
		cfg.MinSamples = def.MinSamples
	}
	return &LatencyDetector{
		cfg:     cfg,
		samples: make(map[string][]time.Duration),
	}
}

// Record appends a latency sample for the given token ID.
func (d *LatencyDetector) Record(tokenID string, latency time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.samples[tokenID] = append(d.samples[tokenID], latency)
}

// Check returns an alert if the average latency for tokenID exceeds
// a threshold, or nil if everything looks healthy.
func (d *LatencyDetector) Check(tokenID string) *alert.Alert {
	d.mu.Lock()
	defer d.mu.Unlock()

	samples := d.samples[tokenID]
	if len(samples) < d.cfg.MinSamples {
		return nil
	}

	var total time.Duration
	for _, s := range samples {
		total += s
	}
	avg := total / time.Duration(len(samples))

	var level alert.Level
	switch {
	case avg >= d.cfg.CriticalThreshold:
		level = alert.LevelCritical
	case avg >= d.cfg.WarningThreshold:
		level = alert.LevelWarning
	default:
		return nil
	}

	return &alert.Alert{
		LeaseID:  tokenID,
		Level:    level,
		Message:  fmt.Sprintf("high token lookup latency: avg %s", avg.Round(time.Millisecond)),
		Labels:   map[string]string{"detector": "latency"},
	}
}
