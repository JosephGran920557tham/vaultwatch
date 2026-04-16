package tokenwatch

import (
	"sync"
	"time"
)

// DefaultEvictionConfig returns sensible eviction defaults.
func DefaultEvictionConfig() EvictionConfig {
	return EvictionConfig{
		MaxAge:          30 * time.Minute,
		SweepInterval:   5 * time.Minute,
	}
}

// EvictionConfig controls how stale token entries are evicted.
type EvictionConfig struct {
	MaxAge        time.Duration
	SweepInterval time.Duration
}

type evictionEntry struct {
	token     string
	seenAt    time.Time
}

// Eviction tracks token activity and evicts tokens unseen beyond MaxAge.
type Eviction struct {
	mu      sync.Mutex
	cfg     EvictionConfig
	entries map[string]time.Time
}

// NewEviction creates an Eviction with the given config.
// Zero values are replaced with defaults.
func NewEviction(cfg EvictionConfig) *Eviction {
	def := DefaultEvictionConfig()
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = def.MaxAge
	}
	if cfg.SweepInterval <= 0 {
		cfg.SweepInterval = def.SweepInterval
	}
	return &Eviction{
		cfg:     cfg,
		entries: make(map[string]time.Time),
	}
}

// Touch records the current time for the given token.
func (e *Eviction) Touch(token string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.entries[token] = time.Now()
}

// Sweep removes tokens not seen within MaxAge and returns evicted token IDs.
func (e *Eviction) Sweep() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	cutoff := time.Now().Add(-e.cfg.MaxAge)
	var evicted []string
	for token, seen := range e.entries {
		if seen.Before(cutoff) {
			delete(e.entries, token)
			evicted = append(evicted, token)
		}
	}
	return evicted
}

// Size returns the number of tracked tokens.
func (e *Eviction) Size() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.entries)
}
