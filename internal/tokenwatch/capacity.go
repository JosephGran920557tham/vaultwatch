package tokenwatch

import (
	"sync"
	"time"
)

// DefaultCapacityConfig returns sensible defaults for capacity tracking.
func DefaultCapacityConfig() CapacityConfig {
	return CapacityConfig{
		MaxTokens:      500,
		WarnThreshold:  0.75,
		CritThreshold:  0.90,
		Window:         5 * time.Minute,
	}
}

// CapacityConfig controls thresholds for the capacity detector.
type CapacityConfig struct {
	MaxTokens     int
	WarnThreshold float64
	CritThreshold float64
	Window        time.Duration
}

// CapacityDetector tracks how full the token registry is and emits alerts
// when utilisation crosses configured thresholds.
type CapacityDetector struct {
	cfg CapacityConfig
	mu  sync.Mutex
	last time.Time
}

// NewCapacityDetector creates a CapacityDetector. Zero-value fields fall back
// to defaults.
func NewCapacityDetector(cfg CapacityConfig) *CapacityDetector {
	def := DefaultCapacityConfig()
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = def.MaxTokens
	}
	if cfg.WarnThreshold <= 0 {
		cfg.WarnThreshold = def.WarnThreshold
	}
	if cfg.CritThreshold <= 0 {
		cfg.CritThreshold = def.CritThreshold
	}
	if cfg.Window <= 0 {
		cfg.Window = def.Window
	}
	return &CapacityDetector{cfg: cfg}
}

// CapacityResult holds the outcome of a capacity check.
type CapacityResult struct {
	Count       int
	Utilisation float64
	Level       string // "ok", "warning", "critical"
}

// Check evaluates current token count against capacity limits.
func (d *CapacityDetector) Check(count int) CapacityResult {
	d.mu.Lock()
	defer d.mu.Unlock()

	util := float64(count) / float64(d.cfg.MaxTokens)
	level := "ok"
	switch {
	case util >= d.cfg.CritThreshold:
		level = "critical"
	case util >= d.cfg.WarnThreshold:
		level = "warning"
	}
	d.last = time.Now()
	return CapacityResult{Count: count, Utilisation: util, Level: level}
}

// LastChecked returns the time of the most recent Check call.
func (d *CapacityDetector) LastChecked() time.Time {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.last
}
