package tokenwatch

import (
	"fmt"
	"sync"
	"time"
)

// DefaultBurstConfig returns a BurstDetector with sensible defaults.
func DefaultBurstConfig() BurstConfig {
	return BurstConfig{
		Window:    30 * time.Second,
		MaxEvents: 10,
	}
}

// BurstConfig controls burst detection behaviour.
type BurstConfig struct {
	Window    time.Duration
	MaxEvents int
}

// BurstDetector tracks event counts within a sliding window and flags
// when the rate exceeds MaxEvents within Window.
type BurstDetector struct {
	cfg    BurstConfig
	mu     sync.Mutex
	events map[string][]time.Time
}

// NewBurstDetector constructs a BurstDetector. Zero-value config fields
// fall back to DefaultBurstConfig.
func NewBurstDetector(cfg BurstConfig) (*BurstDetector, error) {
	def := DefaultBurstConfig()
	if cfg.Window <= 0 {
		cfg.Window = def.Window
	}
	if cfg.MaxEvents <= 0 {
		cfg.MaxEvents = def.MaxEvents
	}
	return &BurstDetector{
		cfg:    cfg,
		events: make(map[string][]time.Time),
	}, nil
}

// Record registers an event for the given key at now.
func (b *BurstDetector) Record(key string, now time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.prune(key, now)
	b.events[key] = append(b.events[key], now)
}

// IsBursting returns true when the number of events recorded for key
// within the configured window exceeds MaxEvents.
func (b *BurstDetector) IsBursting(key string, now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.prune(key, now)
	return len(b.events[key]) > b.cfg.MaxEvents
}

// Count returns the current event count for key within the window.
func (b *BurstDetector) Count(key string, now time.Time) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.prune(key, now)
	return len(b.events[key])
}

func (b *BurstDetector) prune(key string, now time.Time) {
	cutoff := now.Add(-b.cfg.Window)
	evs := b.events[key]
	var kept []time.Time
	for _, t := range evs {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	b.events[key] = kept
}

// String returns a human-readable description of the config.
func (b *BurstDetector) String() string {
	return fmt.Sprintf("BurstDetector(window=%s, maxEvents=%d)", b.cfg.Window, b.cfg.MaxEvents)
}
