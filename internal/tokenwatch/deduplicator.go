package tokenwatch

import (
	"sync"
	"time"
)

// dedupEntry tracks the last time an alert was emitted for a given key.
type dedupEntry struct {
	lastSeen time.Time
	count    int
}

// Deduplicator suppresses repeated alerts for the same token within a
// configurable window, reducing noise when a token is persistently expiring.
type Deduplicator struct {
	mu      sync.Mutex
	entries map[string]*dedupEntry
	window  time.Duration
	now     func() time.Time
}

// NewDeduplicator creates a Deduplicator that suppresses duplicate alerts
// within the given window. A zero or negative window disables deduplication.
func NewDeduplicator(window time.Duration) *Deduplicator {
	if window <= 0 {
		window = 5 * time.Minute
	}
	return &Deduplicator{
		entries: make(map[string]*dedupEntry),
		window:  window,
		now:     time.Now,
	}
}

// Allow returns true if an alert for key should be forwarded. It returns false
// when an alert for the same key was already forwarded within the window.
func (d *Deduplicator) Allow(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	if e, ok := d.entries[key]; ok {
		if now.Sub(e.lastSeen) < d.window {
			e.count++
			return false
		}
	}
	d.entries[key] = &dedupEntry{lastSeen: now, count: 1}
	return true
}

// Reset clears the deduplication state for a specific key.
func (d *Deduplicator) Reset(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.entries, key)
}

// Purge removes all entries whose last-seen time is older than the window,
// keeping memory usage bounded during long-running processes.
func (d *Deduplicator) Purge() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	for k, e := range d.entries {
		if now.Sub(e.lastSeen) >= d.window {
			delete(d.entries, k)
		}
	}
}

// SuppressedCount returns how many times an alert for key was suppressed since
// it was last allowed through.
func (d *Deduplicator) SuppressedCount(key string) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	if e, ok := d.entries[key]; ok {
		return e.count - 1
	}
	return 0
}
