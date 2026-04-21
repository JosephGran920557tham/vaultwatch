package tokenwatch

import (
	"sync"
	"time"
)

// DefaultMirrorConfig returns a MirrorConfig with sensible defaults.
func DefaultMirrorConfig() MirrorConfig {
	return MirrorConfig{
		TTL:      5 * time.Minute,
		MaxItems: 512,
	}
}

// MirrorConfig controls the behaviour of a Mirror.
type MirrorConfig struct {
	TTL      time.Duration
	MaxItems int
}

// mirrorEntry holds a mirrored token TTL snapshot.
type mirrorEntry struct {
	ttl       time.Duration
	recordedAt time.Time
}

// Mirror stores a read-only shadow copy of observed token TTLs, evicting
// entries that exceed the configured TTL or item cap.
type Mirror struct {
	mu      sync.RWMutex
	cfg     MirrorConfig
	entries map[string]mirrorEntry
}

// NewMirror creates a Mirror with the provided config. Zero values are
// replaced with defaults.
func NewMirror(cfg MirrorConfig) *Mirror {
	def := DefaultMirrorConfig()
	if cfg.TTL <= 0 {
		cfg.TTL = def.TTL
	}
	if cfg.MaxItems <= 0 {
		cfg.MaxItems = def.MaxItems
	}
	return &Mirror{
		cfg:     cfg,
		entries: make(map[string]mirrorEntry, cfg.MaxItems),
	}
}

// Observe records the current TTL for a token.
func (m *Mirror) Observe(tokenID string, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.entries) >= m.cfg.MaxItems {
		m.evictOldestLocked()
	}
	m.entries[tokenID] = mirrorEntry{ttl: ttl, recordedAt: time.Now()}
}

// Get returns the last observed TTL for a token and whether it is still valid.
func (m *Mirror) Get(tokenID string) (time.Duration, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.entries[tokenID]
	if !ok {
		return 0, false
	}
	if time.Since(e.recordedAt) > m.cfg.TTL {
		return 0, false
	}
	return e.ttl, true
}

// Len returns the number of live (non-expired) entries.
func (m *Mirror) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, e := range m.entries {
		if time.Since(e.recordedAt) <= m.cfg.TTL {
			count++
		}
	}
	return count
}

// evictOldestLocked removes the entry with the oldest recordedAt timestamp.
// Must be called with m.mu held for writing.
func (m *Mirror) evictOldestLocked() {
	var oldest string
	var oldestTime time.Time
	for id, e := range m.entries {
		if oldest == "" || e.recordedAt.Before(oldestTime) {
			oldest = id
			oldestTime = e.recordedAt
		}
	}
	if oldest != "" {
		delete(m.entries, oldest)
	}
}
