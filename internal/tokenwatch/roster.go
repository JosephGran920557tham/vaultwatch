package tokenwatch

import (
	"fmt"
	"sync"
	"time"
)

// DefaultRosterConfig returns sensible defaults for the Roster.
func DefaultRosterConfig() RosterConfig {
	return RosterConfig{
		MaxSize:    1000,
		EntryTTL:   24 * time.Hour,
		PrunePeriod: 5 * time.Minute,
	}
}

// RosterConfig controls the behaviour of Roster.
type RosterConfig struct {
	MaxSize     int
	EntryTTL    time.Duration
	PrunePeriod time.Duration
}

// rosterEntry holds a token's last-seen time and metadata labels.
type rosterEntry struct {
	TokenID string
	Labels  map[string]string
	SeenAt  time.Time
}

// Roster maintains a bounded, time-limited set of known token IDs and their
// associated labels. It is safe for concurrent use.
type Roster struct {
	mu      sync.RWMutex
	cfg     RosterConfig
	entries map[string]rosterEntry
}

// NewRoster creates a Roster with the supplied config. Zero-value fields are
// replaced with defaults.
func NewRoster(cfg RosterConfig) *Roster {
	def := DefaultRosterConfig()
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = def.MaxSize
	}
	if cfg.EntryTTL <= 0 {
		cfg.EntryTTL = def.EntryTTL
	}
	if cfg.PrunePeriod <= 0 {
		cfg.PrunePeriod = def.PrunePeriod
	}
	return &Roster{
		cfg:     cfg,
		entries: make(map[string]rosterEntry, cfg.MaxSize),
	}
}

// Touch records or refreshes a token in the roster. It returns an error when
// the roster is already at capacity and the token is not already present.
func (r *Roster) Touch(tokenID string, labels map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.entries[tokenID]; !exists && len(r.entries) >= r.cfg.MaxSize {
		return fmt.Errorf("roster: capacity %d exceeded", r.cfg.MaxSize)
	}
	r.entries[tokenID] = rosterEntry{
		TokenID: tokenID,
		Labels:  labels,
		SeenAt:  time.Now(),
	}
	return nil
}

// Active returns all token IDs that have been seen within the configured TTL.
func (r *Roster) Active() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cutoff := time.Now().Add(-r.cfg.EntryTTL)
	out := make([]string, 0, len(r.entries))
	for id, e := range r.entries {
		if e.SeenAt.After(cutoff) {
			out = append(out, id)
		}
	}
	return out
}

// Prune removes entries older than the configured TTL. It returns the number
// of entries removed.
func (r *Roster) Prune() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-r.cfg.EntryTTL)
	removed := 0
	for id, e := range r.entries {
		if !e.SeenAt.After(cutoff) {
			delete(r.entries, id)
			removed++
		}
	}
	return removed
}

// Len returns the current number of entries regardless of staleness.
func (r *Roster) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}
