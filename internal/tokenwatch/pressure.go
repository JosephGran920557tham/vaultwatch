package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultPressureConfig returns a PressureConfig with sensible defaults.
func DefaultPressureConfig() PressureConfig {
	return PressureConfig{
		Window:        5 * time.Minute,
		HighThreshold: 0.75,
		LowThreshold:  0.40,
	}
}

// PressureConfig controls how the PressureDetector evaluates token TTL pressure.
type PressureConfig struct {
	// Window is the time window over which samples are evaluated.
	Window time.Duration
	// HighThreshold is the fraction of tokens near expiry that triggers a critical alert.
	HighThreshold float64
	// LowThreshold is the fraction that triggers a warning alert.
	LowThreshold float64
}

// PressureDetector tracks the ratio of tokens under TTL pressure within a sliding window.
type PressureDetector struct {
	cfg     PressureConfig
	mu      sync.Mutex
	samples []pressureSample
}

type pressureSample struct {
	at      time.Time
	pressed bool
}

// NewPressureDetector creates a PressureDetector with the given config.
// Zero-value fields fall back to defaults.
func NewPressureDetector(cfg PressureConfig) *PressureDetector {
	def := DefaultPressureConfig()
	if cfg.Window <= 0 {
		cfg.Window = def.Window
	}
	if cfg.HighThreshold <= 0 {
		cfg.HighThreshold = def.HighThreshold
	}
	if cfg.LowThreshold <= 0 {
		cfg.LowThreshold = def.LowThreshold
	}
	return &PressureDetector{cfg: cfg}
}

// Record adds a sample indicating whether the token is under pressure.
func (d *PressureDetector) Record(tokenID string, pressed bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.prune(time.Now())
	d.samples = append(d.samples, pressureSample{at: time.Now(), pressed: pressed})
}

// Check evaluates current pressure and returns an alert if thresholds are breached.
func (d *PressureDetector) Check(tokenID string) *alert.Alert {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	d.prune(now)
	if len(d.samples) == 0 {
		return nil
	}
	var pressed int
	for _, s := range d.samples {
		if s.pressed {
			pressed++
		}
	}
	ratio := float64(pressed) / float64(len(d.samples))
	switch {
	case ratio >= d.cfg.HighThreshold:
		return &alert.Alert{
			LeaseID:   tokenID,
			Level:     alert.Critical,
			Message:   fmt.Sprintf("token pressure critical: %.0f%% of samples near expiry", ratio*100),
			Timestamp: now,
		}
	case ratio >= d.cfg.LowThreshold:
		return &alert.Alert{
			LeaseID:   tokenID,
			Level:     alert.Warning,
			Message:   fmt.Sprintf("token pressure elevated: %.0f%% of samples near expiry", ratio*100),
			Timestamp: now,
		}
	}
	return nil
}

func (d *PressureDetector) prune(now time.Time) {
	cutoff := now.Add(-d.cfg.Window)
	i := 0
	for i < len(d.samples) && d.samples[i].at.Before(cutoff) {
		i++
	}
	d.samples = d.samples[i:]
}
