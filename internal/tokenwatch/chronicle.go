package tokenwatch

import (
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultChronicleConfig returns sensible defaults for Chronicle.
func DefaultChronicleConfig() ChronicleConfig {
	return ChronicleConfig{
		MaxEntries: 500,
		MaxAge:     24 * time.Hour,
	}
}

// ChronicleConfig controls retention for the Chronicle store.
type ChronicleConfig struct {
	MaxEntries int
	MaxAge     time.Duration
}

// ChronicleEntry is a single recorded alert event.
type ChronicleEntry struct {
	Alert     alert.Alert
	RecordedAt time.Time
}

// Chronicle maintains a bounded, time-limited history of alerts per token.
type Chronicle struct {
	mu      sync.Mutex
	cfg     ChronicleConfig
	records map[string][]ChronicleEntry
}

// NewChronicle creates a Chronicle with the given config, applying defaults for zero values.
func NewChronicle(cfg ChronicleConfig) *Chronicle {
	def := DefaultChronicleConfig()
	if cfg.MaxEntries <= 0 {
		cfg.MaxEntries = def.MaxEntries
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = def.MaxAge
	}
	return &Chronicle{
		cfg:     cfg,
		records: make(map[string][]ChronicleEntry),
	}
}

// Record appends an alert to the history for the given token ID.
func (c *Chronicle) Record(tokenID string, a alert.Alert) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	entries := c.prune(c.records[tokenID], now)
	entries = append(entries, ChronicleEntry{Alert: a, RecordedAt: now})
	if len(entries) > c.cfg.MaxEntries {
		entries = entries[len(entries)-c.cfg.MaxEntries:]
	}
	c.records[tokenID] = entries
}

// List returns all non-expired entries for the given token ID.
func (c *Chronicle) List(tokenID string) []ChronicleEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	pruned := c.prune(c.records[tokenID], time.Now())
	c.records[tokenID] = pruned
	out := make([]ChronicleEntry, len(pruned))
	copy(out, pruned)
	return out
}

// Len returns the number of stored entries for a token.
func (c *Chronicle) Len(tokenID string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.prune(c.records[tokenID], time.Now()))
}

func (c *Chronicle) prune(entries []ChronicleEntry, now time.Time) []ChronicleEntry {
	cutoff := now.Add(-c.cfg.MaxAge)
	var out []ChronicleEntry
	for _, e := range entries {
		if e.RecordedAt.After(cutoff) {
			out = append(out, e)
		}
	}
	return out
}
