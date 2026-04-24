package tokenwatch

import (
	"math"
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultEntropyConfig returns a sensible default configuration.
func DefaultEntropyConfig() EntropyConfig {
	return EntropyConfig{
		MinSamples:      5,
		WarningThreshold: 0.6,
		CriticalThreshold: 0.3,
		Window:          10 * time.Minute,
	}
}

// EntropyConfig controls entropy detection thresholds.
type EntropyConfig struct {
	MinSamples        int
	WarningThreshold  float64
	CriticalThreshold float64
	Window            time.Duration
}

// EntropyDetector measures TTL entropy (diversity) across a token's samples.
// Low entropy indicates that TTLs are converging, which may signal abnormal
// renewal patterns or external interference.
type EntropyDetector struct {
	cfg     EntropyConfig
	mu      sync.Mutex
	samples map[string][]float64
}

// NewEntropyDetector constructs an EntropyDetector, applying defaults for zero values.
func NewEntropyDetector(cfg EntropyConfig) *EntropyDetector {
	def := DefaultEntropyConfig()
	if cfg.MinSamples <= 0 {
		cfg.MinSamples = def.MinSamples
	}
	if cfg.WarningThreshold <= 0 {
		cfg.WarningThreshold = def.WarningThreshold
	}
	if cfg.CriticalThreshold <= 0 {
		cfg.CriticalThreshold = def.CriticalThreshold
	}
	if cfg.Window <= 0 {
		cfg.Window = def.Window
	}
	return &EntropyDetector{
		cfg:     cfg,
		samples: make(map[string][]float64),
	}
}

// Record adds a TTL sample for the given token.
func (e *EntropyDetector) Record(tokenID string, ttl time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.samples[tokenID] = append(e.samples[tokenID], ttl.Seconds())
}

// Check evaluates entropy for the token and returns an alert if below threshold.
func (e *EntropyDetector) Check(tokenID string) *alert.Alert {
	e.mu.Lock()
	samples := append([]float64(nil), e.samples[tokenID]...)
	e.mu.Unlock()

	if len(samples) < e.cfg.MinSamples {
		return nil
	}

	entropyVal := shannonEntropy(samples)

	var level alert.Level
	switch {
	case entropyVal < e.cfg.CriticalThreshold:
		level = alert.LevelCritical
	case entropyVal < e.cfg.WarningThreshold:
		level = alert.LevelWarning
	default:
		return nil
	}

	return &alert.Alert{
		LeaseID:   tokenID,
		Level:     level,
		Message:   "low TTL entropy detected — renewal pattern may be abnormal",
		ExpiresAt: time.Now().Add(e.cfg.Window),
	}
}

// shannonEntropy computes normalised Shannon entropy over a slice of values.
func shannonEntropy(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	freq := make(map[int64]int)
	for _, v := range values {
		bucket := int64(v / 60) // 1-minute buckets
		freq[bucket]++
	}
	n := float64(len(values))
	var h float64
	for _, count := range freq {
		p := float64(count) / n
		if p > 0 {
			h -= p * math.Log2(p)
		}
	}
	maxH := math.Log2(float64(len(freq)))
	if maxH == 0 {
		return 0
	}
	return h / maxH
}
