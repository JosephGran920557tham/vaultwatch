package tokenwatch

import (
	"sync"
	"time"
)

// DefaultCensusConfig returns a CensusConfig with sensible defaults.
func DefaultCensusConfig() CensusConfig {
	return CensusConfig{
		MaxAge: 10 * time.Minute,
	}
}

// CensusConfig controls Census behaviour.
type CensusConfig struct {
	// MaxAge is how long a token entry is retained without a refresh.
	MaxAge time.Duration
}

// censusEntry holds a snapshot of a token's observed state.
type censusEntry struct {
	TokenID   string
	Labels    map[string]string
	LastSeen  time.Time
}

// Census tracks the set of active tokens observed during a scan cycle,
// allowing downstream components to reason about the full population.
type Census struct {
	mu      sync.RWMutex
	cfg     CensusConfig
	entries map[string]*censusEntry
}

// NewCensus creates a Census with the given config, applying defaults for
// zero-value fields.
func NewCensus(cfg CensusConfig) *Census {
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = DefaultCensusConfig().MaxAge
	}
	return &Census{
		cfg:     cfg,
		entries: make(map[string]*censusEntry),
	}
}

// Observe records or refreshes a token in the census.
func (c *Census) Observe(tokenID string, labels map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[tokenID] = &censusEntry{
		TokenID:  tokenID,
		Labels:   labels,
		LastSeen: time.Now(),
	}
}

// Active returns all token IDs seen within MaxAge.
func (c *Census) Active() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cutoff := time.Now().Add(-c.cfg.MaxAge)
	out := make([]string, 0, len(c.entries))
	for id, e := range c.entries {
		if e.LastSeen.After(cutoff) {
			out = append(out, id)
		}
	}
	return out
}

// Len returns the number of currently active entries.
func (c *Census) Len() int {
	return len(c.Active())
}

// Evict removes entries older than MaxAge.
func (c *Census) Evict() {
	c.mu.Lock()
	defer c.mu.Unlock()
	cutoff := time.Now().Add(-c.cfg.MaxAge)
	for id, e := range c.entries {
		if !e.LastSeen.After(cutoff) {
			delete(c.entries, id)
		}
	}
}
