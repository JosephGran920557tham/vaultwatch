package tokenwatch

import (
	"sync"
	"time"
)

// DefaultTombstoneConfig returns a TombstoneConfig with sensible defaults.
func DefaultTombstoneConfig() TombstoneConfig {
	return TombstoneConfig{
		RetentionWindow: 24 * time.Hour,
	}
}

// TombstoneConfig controls how long revoked token records are retained.
type TombstoneConfig struct {
	RetentionWindow time.Duration
}

// TombstoneEntry records when a token was revoked.
type TombstoneEntry struct {
	TokenID  string
	RevokedAt time.Time
}

// Tombstone tracks revoked tokens within a retention window so that
// downstream scanners can suppress alerts for already-revoked tokens.
type Tombstone struct {
	mu      sync.RWMutex
	entries map[string]TombstoneEntry
	cfg     TombstoneConfig
	now     func() time.Time
}

// NewTombstone creates a Tombstone with the given config.
// Zero values in cfg are replaced with defaults.
func NewTombstone(cfg TombstoneConfig) *Tombstone {
	defaults := DefaultTombstoneConfig()
	if cfg.RetentionWindow <= 0 {
		cfg.RetentionWindow = defaults.RetentionWindow
	}
	return &Tombstone{
		entries: make(map[string]TombstoneEntry),
		cfg:     cfg,
		now:     time.Now,
	}
}

// Revoke marks a token as revoked at the current time.
func (t *Tombstone) Revoke(tokenID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries[tokenID] = TombstoneEntry{
		TokenID:   tokenID,
		RevokedAt: t.now(),
	}
}

// IsRevoked reports whether the token has a tombstone entry within the
// retention window. Expired tombstones are pruned lazily on read.
func (t *Tombstone) IsRevoked(tokenID string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	entry, ok := t.entries[tokenID]
	if !ok {
		return false
	}
	if t.now().Sub(entry.RevokedAt) > t.cfg.RetentionWindow {
		delete(t.entries, tokenID)
		return false
	}
	return true
}

// Purge removes all tombstone entries whose retention window has elapsed.
func (t *Tombstone) Purge() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.now()
	removed := 0
	for id, entry := range t.entries {
		if now.Sub(entry.RevokedAt) > t.cfg.RetentionWindow {
			delete(t.entries, id)
			removed++
		}
	}
	return removed
}

// Len returns the number of currently tracked tombstone entries.
func (t *Tombstone) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.entries)
}
