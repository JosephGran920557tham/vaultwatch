package tokenwatch

import (
	"fmt"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// ForecastConfig holds configuration for the TTL forecast detector.
type ForecastConfig struct {
	// SampleWindow is how many TTL samples to use for projection.
	SampleWindow int
	// ProjectionHorizon is how far ahead to forecast expiry.
	ProjectionHorizon time.Duration
	// CriticalThreshold triggers a critical alert if projected TTL is below this.
	CriticalThreshold time.Duration
	// WarningThreshold triggers a warning alert if projected TTL is below this.
	WarningThreshold time.Duration
}

// DefaultForecastConfig returns a ForecastConfig with sensible defaults.
func DefaultForecastConfig() ForecastConfig {
	return ForecastConfig{
		SampleWindow:      5,
		ProjectionHorizon: 30 * time.Minute,
		CriticalThreshold: 10 * time.Minute,
		WarningThreshold:  20 * time.Minute,
	}
}

// ForecastDetector projects future TTL based on observed decay samples.
type ForecastDetector struct {
	cfg     ForecastConfig
	samples []time.Duration
}

// NewForecastDetector creates a ForecastDetector with the given config.
// Zero-value fields fall back to defaults.
func NewForecastDetector(cfg ForecastConfig) *ForecastDetector {
	def := DefaultForecastConfig()
	if cfg.SampleWindow <= 0 {
		cfg.SampleWindow = def.SampleWindow
	}
	if cfg.ProjectionHorizon <= 0 {
		cfg.ProjectionHorizon = def.ProjectionHorizon
	}
	if cfg.CriticalThreshold <= 0 {
		cfg.CriticalThreshold = def.CriticalThreshold
	}
	if cfg.WarningThreshold <= 0 {
		cfg.WarningThreshold = def.WarningThreshold
	}
	return &ForecastDetector{cfg: cfg}
}

// Record adds a TTL observation to the sample buffer.
func (f *ForecastDetector) Record(ttl time.Duration) {
	f.samples = append(f.samples, ttl)
	if len(f.samples) > f.cfg.SampleWindow {
		f.samples = f.samples[len(f.samples)-f.cfg.SampleWindow:]
	}
}

// Check returns an alert if the projected TTL breaches a threshold, or nil.
func (f *ForecastDetector) Check(tokenID string) *alert.Alert {
	if len(f.samples) < 2 {
		return nil
	}
	decay := f.averageDecay()
	if decay <= 0 {
		return nil
	}
	latest := f.samples[len(f.samples)-1]
	steps := float64(f.cfg.ProjectionHorizon) / float64(decay)
	projected := time.Duration(float64(latest) - steps*float64(decay))
	if projected < 0 {
		projected = 0
	}
	var level alert.Level
	switch {
	case projected <= f.cfg.CriticalThreshold:
		level = alert.LevelCritical
	case projected <= f.cfg.WarningThreshold:
		level = alert.LevelWarning
	default:
		return nil
	}
	return &alert.Alert{
		LeaseID:   tokenID,
		Level:     level,
		Message:   fmt.Sprintf("projected TTL in %s: %s", f.cfg.ProjectionHorizon, projected.Round(time.Second)),
		ExpiresIn: projected,
	}
}

func (f *ForecastDetector) averageDecay() time.Duration {
	var total time.Duration
	for i := 1; i < len(f.samples); i++ {
		diff := f.samples[i-1] - f.samples[i]
		if diff > 0 {
			total += diff
		}
	}
	return total / time.Duration(len(f.samples)-1)
}
