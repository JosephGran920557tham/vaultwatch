package tokenwatch

import (
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultCorrelationConfig returns sensible defaults for the correlation detector.
func DefaultCorrelationConfig() CorrelationConfig {
	return CorrelationConfig{
		Window:       5 * time.Minute,
		MinEvents:    3,
		ScoreThreshold: 0.75,
	}
}

// CorrelationConfig controls how the CorrelationDetector groups related alerts.
type CorrelationConfig struct {
	Window         time.Duration
	MinEvents      int
	ScoreThreshold float64
}

// CorrelationDetector groups alerts that share a common token or path pattern
// within a sliding time window and emits a single correlated alert when the
// event density exceeds the configured threshold.
type CorrelationDetector struct {
	mu     sync.Mutex
	cfg    CorrelationConfig
	bucket map[string][]time.Time
	now    func() time.Time
}

// NewCorrelationDetector creates a CorrelationDetector with the given config.
// Zero values are replaced with defaults.
func NewCorrelationDetector(cfg CorrelationConfig) *CorrelationDetector {
	def := DefaultCorrelationConfig()
	if cfg.Window <= 0 {
		cfg.Window = def.Window
	}
	if cfg.MinEvents <= 0 {
		cfg.MinEvents = def.MinEvents
	}
	if cfg.ScoreThreshold <= 0 {
		cfg.ScoreThreshold = def.ScoreThreshold
	}
	return &CorrelationDetector{
		cfg:    cfg,
		bucket: make(map[string][]time.Time),
		now:    time.Now,
	}
}

// Record registers an alert event for the given correlation key.
func (d *CorrelationDetector) Record(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.now()
	d.bucket[key] = append(d.prune(key, now), now)
}

// Check returns a correlated alert if the event density for key exceeds the
// configured threshold, otherwise nil.
func (d *CorrelationDetector) Check(key string, level alert.Level) *alert.Alert {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.now()
	events := d.prune(key, now)
	d.bucket[key] = events
	if len(events) < d.cfg.MinEvents {
		return nil
	}
	score := float64(len(events)) / float64(d.cfg.MinEvents)
	if score < d.cfg.ScoreThreshold {
		return nil
	}
	return &alert.Alert{
		LeaseID: key,
		Level:   level,
		Message: "correlated alert burst detected",
		Labels:  map[string]string{"detector": "correlation"},
	}
}

func (d *CorrelationDetector) prune(key string, now time.Time) []time.Time {
	cutoff := now.Add(-d.cfg.Window)
	var kept []time.Time
	for _, t := range d.bucket[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	return kept
}
