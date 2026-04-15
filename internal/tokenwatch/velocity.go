package tokenwatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

// DefaultVelocityConfig returns sensible defaults for velocity detection.
func DefaultVelocityConfig() VelocityConfig {
	return VelocityConfig{
		Window:        5 * time.Minute,
		MinSamples:    3,
		DropThreshold: 0.20, // 20% drop per minute triggers warning
		CritThreshold: 0.40, // 40% drop per minute triggers critical
	}
}

// VelocityConfig controls how TTL velocity (rate of change) is evaluated.
type VelocityConfig struct {
	Window        time.Duration
	MinSamples    int
	DropThreshold float64 // fraction per minute
	CritThreshold float64 // fraction per minute
}

// velocitySample holds a TTL observation at a point in time.
type velocitySample struct {
	at  time.Time
	ttl time.Duration
}

// VelocityDetector tracks the rate at which a token's TTL is decreasing
// and raises alerts when the drop rate exceeds configured thresholds.
type VelocityDetector struct {
	cfg     VelocityConfig
	mu      sync.Mutex
	samples map[string][]velocitySample
}

// NewVelocityDetector creates a VelocityDetector with the given config.
// Zero-value fields are replaced with defaults.
func NewVelocityDetector(cfg VelocityConfig) *VelocityDetector {
	def := DefaultVelocityConfig()
	if cfg.Window <= 0 {
		cfg.Window = def.Window
	}
	if cfg.MinSamples <= 0 {
		cfg.MinSamples = def.MinSamples
	}
	if cfg.DropThreshold <= 0 {
		cfg.DropThreshold = def.DropThreshold
	}
	if cfg.CritThreshold <= 0 {
		cfg.CritThreshold = def.CritThreshold
	}
	return &VelocityDetector{
		cfg:     cfg,
		samples: make(map[string][]velocitySample),
	}
}

// Record adds a TTL observation for the given token ID.
func (v *VelocityDetector) Record(tokenID string, ttl time.Duration) {
	v.mu.Lock()
	defer v.mu.Unlock()
	cutoff := time.Now().Add(-v.cfg.Window)
	prev := v.samples[tokenID]
	filtered := prev[:0]
	for _, s := range prev {
		if s.at.After(cutoff) {
			filtered = append(filtered, s)
		}
	}
	v.samples[tokenID] = append(filtered, velocitySample{at: time.Now(), ttl: ttl})
}

// Check returns an alert if the TTL drop velocity for tokenID exceeds thresholds.
// Returns nil when there are insufficient samples or velocity is acceptable.
func (v *VelocityDetector) Check(tokenID string) *alert.Alert {
	v.mu.Lock()
	defer v.mu.Unlock()
	samples := v.samples[tokenID]
	if len(samples) < v.cfg.MinSamples {
		return nil
	}
	first := samples[0]
	last := samples[len(samples)-1]
	elapsed := last.at.Sub(first.at).Minutes()
	if elapsed <= 0 {
		return nil
	}
	drop := (first.ttl - last.ttl).Minutes()
	velocity := drop / elapsed // fraction-of-initial drop per minute not needed; raw minutes/min
	// Normalise against initial TTL to get a fraction
	if first.ttl <= 0 {
		return nil
	}
	normVelocity := (drop / first.ttl.Minutes()) / elapsed
	var level alert.Level
	switch {
	case normVelocity >= v.cfg.CritThreshold:
		level = alert.LevelCritical
	case normVelocity >= v.cfg.DropThreshold:
		level = alert.LevelWarning
	default:
		_ = velocity
		return nil
	}
	return &alert.Alert{
		LeaseID: tokenID,
		Level:   level,
		Message: fmt.Sprintf("token %s TTL dropping at %.2f%% per minute", tokenID, normVelocity*100),
	}
}
