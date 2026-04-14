// Package snapshot provides point-in-time capture and comparison of lease states.
package snapshot

import (
	"sync"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

// Entry holds a captured lease alert at a specific moment.
type Entry struct {
	Alert     alert.Alert
	CapturedAt time.Time
}

// Snapshot is an immutable collection of lease entries captured at one point in time.
type Snapshot struct {
	TakenAt time.Time
	Entries []Entry
}

// Store holds the most recent snapshot and supports atomic replacement.
type Store struct {
	mu       sync.RWMutex
	current  *Snapshot
}

// NewStore returns an initialised snapshot Store.
func NewStore() *Store {
	return &Store{}
}

// Save atomically replaces the current snapshot.
func (s *Store) Save(snap *Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = snap
}

// Latest returns the most recently saved snapshot, or nil if none exists.
func (s *Store) Latest() *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// Capture creates a new Snapshot from a slice of alerts.
func Capture(alerts []alert.Alert) *Snapshot {
	now := time.Now().UTC()
	entries := make([]Entry, len(alerts))
	for i, a := range alerts {
		entries[i] = Entry{Alert: a, CapturedAt: now}
	}
	return &Snapshot{TakenAt: now, Entries: entries}
}

// Diff returns alerts present in next but absent (by LeaseID) from prev.
// A nil prev snapshot treats all entries in next as new.
func Diff(prev, next *Snapshot) []alert.Alert {
	if prev == nil {
		result := make([]alert.Alert, len(next.Entries))
		for i, e := range next.Entries {
			result[i] = e.Alert
		}
		return result
	}
	existing := make(map[string]struct{}, len(prev.Entries))
	for _, e := range prev.Entries {
		existing[e.Alert.LeaseID] = struct{}{}
	}
	var diff []alert.Alert
	for _, e := range next.Entries {
		if _, found := existing[e.Alert.LeaseID]; !found {
			diff = append(diff, e.Alert)
		}
	}
	return diff
}
