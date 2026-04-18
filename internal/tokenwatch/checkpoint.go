package tokenwatch

import (
	"sync"
	"time"
)

// CheckpointConfig holds configuration for the checkpoint tracker.
type CheckpointConfig struct {
	// MaxAge is how long a checkpoint is considered valid.
	MaxAge time.Duration
}

// DefaultCheckpointConfig returns sensible defaults.
func DefaultCheckpointConfig() CheckpointConfig {
	return CheckpointConfig{
		MaxAge: 10 * time.Minute,
	}
}

type checkpointEntry struct {
	recordedAt time.Time
	ttl        time.Duration
}

// Checkpoint tracks the last known TTL snapshot for each token.
type Checkpoint struct {
	mu      sync.Mutex
	cfg     CheckpointConfig
	entries map[string]checkpointEntry
}

// NewCheckpoint creates a new Checkpoint with the given config.
// Zero values fall back to defaults.
func NewCheckpoint(cfg CheckpointConfig) *Checkpoint {
	def := DefaultCheckpointConfig()
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = def.MaxAge
	}
	return &Checkpoint{
		cfg:     cfg,
		entries: make(map[string]checkpointEntry),
	}
}

// Record stores the current TTL for a token.
func (c *Checkpoint) Record(tokenID string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[tokenID] = checkpointEntry{
		recordedAt: time.Now(),
		ttl:        ttl,
	}
}

// Get returns the last recorded TTL and whether it is still valid.
func (c *Checkpoint) Get(tokenID string) (time.Duration, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[tokenID]
	if !ok {
		return 0, false
	}
	if time.Since(e.recordedAt) > c.cfg.MaxAge {
		delete(c.entries, tokenID)
		return 0, false
	}
	return e.ttl, true
}

// Delete removes the checkpoint entry for a token.
func (c *Checkpoint) Delete(tokenID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, tokenID)
}

// Len returns the number of active checkpoint entries.
func (c *Checkpoint) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}
