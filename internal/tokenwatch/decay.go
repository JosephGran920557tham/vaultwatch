package tokenwatch

import (
	"sync"
	"time"
)

// DefaultDecayConfig returns sensible defaults for DecayDetector.
func DefaultDecayConfig() DecayConfig {
	return DecayConfig{
		HalfLife:        30 * time.Minute,
		WarningScore:    0.4,
		CriticalScore:   0.2,
	}
}

// DecayConfig controls decay detection thresholds.
type DecayConfig struct {
	HalfLife      time.Duration
	WarningScore  float64
	CriticalScore float64
}

// DecayDetector tracks exponential TTL decay and emits alerts when the
// normalised score drops below configured thresholds.
type DecayDetector struct {
	cfg    DecayConfig
	mu     sync.Mutex
	scores map[string]float64
	last   map[string]time.Time
}

// NewDecayDetector creates a DecayDetector. Zero-value fields in cfg are
// replaced with defaults.
func NewDecayDetector(cfg DecayConfig) *DecayDetector {
	def := DefaultDecayConfig()
	if cfg.HalfLife <= 0 {
		cfg.HalfLife = def.HalfLife
	}
	if cfg.WarningScore <= 0 {
		cfg.WarningScore = def.WarningScore
	}
	if cfg.CriticalScore <= 0 {
		cfg.CriticalScore = def.CriticalScore
	}
	return &DecayDetector{
		cfg:    cfg,
		scores: make(map[string]float64),
		last:   make(map[string]time.Time),
	}
}

// Record updates the decay score for tokenID using the provided currentTTL and
// initialTTL. It must be called each time a fresh TTL observation is available.
func (d *DecayDetector) Record(tokenID string, currentTTL, initialTTL time.Duration) {
	if initialTTL <= 0 {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.scores[tokenID] = float64(currentTTL) / float64(initialTTL)
	d.last[tokenID] = time.Now()
}

// Check returns a warning or critical alert if the decay score for tokenID is
// below the configured thresholds. Returns nil when healthy.
func (d *DecayDetector) Check(tokenID string) *Alert {
	d.mu.Lock()
	score, ok := d.scores[tokenID]
	d.mu.Unlock()
	if !ok {
		return nil
	}
	switch {
	case score <= d.cfg.CriticalScore:
		a := buildTokenAlert(tokenID, LevelCritical, "token TTL decay critical")
		return &a
	case score <= d.cfg.WarningScore:
		a := buildTokenAlert(tokenID, LevelWarning, "token TTL decay warning")
		return &a
	}
	return nil
}
