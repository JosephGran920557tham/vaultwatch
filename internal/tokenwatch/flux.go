package tokenwatch

import (
	"sync"
	"time"
)

// DefaultFluxConfig returns a FluxConfig with sensible defaults.
func DefaultFluxConfig() FluxConfig {
	return FluxConfig{
		Window:     5 * time.Minute,
		MaxSamples: 20,
	}
}

// FluxConfig controls the behaviour of the FluxDetector.
type FluxConfig struct {
	// Window is the time range over which flux is measured.
	Window time.Duration
	// MaxSamples is the maximum number of TTL samples retained per token.
	MaxSamples int
}

// fluxSample holds a single TTL observation.
type fluxSample struct {
	at  time.Time
	ttl time.Duration
}

// FluxDetector measures the rate of TTL change (flux) for a token and
// returns a warning or critical alert when the flux exceeds configured
// thresholds.
type FluxDetector struct {
	cfg     FluxConfig
	mu      sync.Mutex
	samples map[string][]fluxSample
}

// NewFluxDetector creates a FluxDetector. Zero-value fields in cfg are
// replaced with defaults.
func NewFluxDetector(cfg FluxConfig) *FluxDetector {
	def := DefaultFluxConfig()
	if cfg.Window <= 0 {
		cfg.Window = def.Window
	}
	if cfg.MaxSamples <= 0 {
		cfg.MaxSamples = def.MaxSamples
	}
	return &FluxDetector{
		cfg:     cfg,
		samples: make(map[string][]fluxSample),
	}
}

// Record adds a new TTL observation for the given token.
func (d *FluxDetector) Record(tokenID string, ttl time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-d.cfg.Window)

	ss := d.samples[tokenID]
	// Prune old samples.
	filtered := ss[:0]
	for _, s := range ss {
		if s.at.After(cutoff) {
			filtered = append(filtered, s)
		}
	}
	filtered = append(filtered, fluxSample{at: now, ttl: ttl})
	if len(filtered) > d.cfg.MaxSamples {
		filtered = filtered[len(filtered)-d.cfg.MaxSamples:]
	}
	d.samples[tokenID] = filtered
}

// Flux returns the absolute difference in TTL between the oldest and newest
// samples within the window. Returns 0 if fewer than two samples exist.
func (d *FluxDetector) Flux(tokenID string) time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()

	ss := d.samples[tokenID]
	if len(ss) < 2 {
		return 0
	}
	delta := ss[0].ttl - ss[len(ss)-1].ttl
	if delta < 0 {
		delta = -delta
	}
	return delta
}
