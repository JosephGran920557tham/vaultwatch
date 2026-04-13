// Package cache provides an in-memory store for Vault lease metadata,
// reducing redundant API calls during repeated monitor runs.
package cache

import (
	"sync"
	"time"
)

// Entry holds a cached lease record with an expiry timestamp.
type Entry struct {
	LeaseID   string
	ExpiresAt time.Time
	Meta      map[string]string
	CachedAt  time.Time
}

// IsExpired reports whether the cache entry itself has gone stale.
func (e Entry) IsExpired(ttl time.Duration) bool {
	return time.Since(e.CachedAt) > ttl
}

// Store is a thread-safe in-memory lease cache.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
	ttl     time.Duration
}

// New creates a Store with the given entry TTL.
func New(ttl time.Duration) *Store {
	return &Store{
		entries: make(map[string]Entry),
		ttl:     ttl,
	}
}

// Set inserts or replaces a lease entry.
func (s *Store) Set(leaseID string, e Entry) {
	e.CachedAt = time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[leaseID] = e
}

// Get retrieves a lease entry. The second return value is false when the
// entry is absent or has exceeded the store TTL.
func (s *Store) Get(leaseID string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[leaseID]
	if !ok || e.IsExpired(s.ttl) {
		return Entry{}, false
	}
	return e, true
}

// Delete removes a single entry.
func (s *Store) Delete(leaseID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, leaseID)
}

// Purge removes all entries that have exceeded the store TTL.
func (s *Store) Purge() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := 0
	for id, e := range s.entries {
		if e.IsExpired(s.ttl) {
			delete(s.entries, id)
			removed++
		}
	}
	return removed
}

// Len returns the current number of entries (including potentially stale ones).
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}
