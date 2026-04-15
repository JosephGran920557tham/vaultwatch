package tokenwatch

import (
	"math/rand"
	"time"
)

// JitterConfig holds configuration for jitter applied to intervals.
type JitterConfig struct {
	// Factor is the maximum fraction of the base duration to add as jitter.
	// For example, 0.2 means up to 20% of the base duration is added.
	// Must be in the range [0.0, 1.0].
	Factor float64
}

// DefaultJitterConfig returns a JitterConfig with sensible defaults.
func DefaultJitterConfig() JitterConfig {
	return JitterConfig{
		Factor: 0.2,
	}
}

// Jitter applies randomised jitter to a base duration.
type Jitter struct {
	cfg JitterConfig
	rng *rand.Rand
}

// NewJitter creates a Jitter with the given config.
// If cfg.Factor is outside [0.0, 1.0] it is clamped to the nearest boundary.
func NewJitter(cfg JitterConfig) *Jitter {
	if cfg.Factor < 0 {
		cfg.Factor = 0
	}
	if cfg.Factor > 1 {
		cfg.Factor = 1
	}
	return &Jitter{
		cfg: cfg,
		//nolint:gosec // non-cryptographic randomness is intentional here
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Apply returns base plus a random jitter up to Factor*base.
func (j *Jitter) Apply(base time.Duration) time.Duration {
	if j.cfg.Factor == 0 || base <= 0 {
		return base
	}
	maxJitter := float64(base) * j.cfg.Factor
	added := time.Duration(j.rng.Float64() * maxJitter)
	return base + added
}

// Factor returns the configured jitter factor.
func (j *Jitter) Factor() float64 {
	return j.cfg.Factor
}
