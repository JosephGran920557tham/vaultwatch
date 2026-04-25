package tokenwatch

import (
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultArchiveConfig returns a sensible default configuration for the Archive.
func DefaultArchiveConfig() ArchiveConfig {
	return ArchiveConfig{
		MaxEntries: 500,
		MaxAge:     24 * time.Hour,
	}
}

// ArchiveConfig controls retention behaviour of the alert archive.
type ArchiveConfig struct {
	MaxEntries int
	MaxAge     time.Duration
}

// archiveEntry wraps an alert with its recorded timestamp.
type archiveEntry struct {
	Alert     alert.Alert
	RecordedAt time.Time
}

// Archive stores a bounded, time-limited history of fired alerts for
// post-hoc inspection and reporting.
type Archive struct {
	mu      sync.Mutex
	cfg     ArchiveConfig
	entries []archiveEntry
}

// NewArchive creates an Archive with the given config. Zero values are
// replaced with defaults.
func NewArchive(cfg ArchiveConfig) *Archive {
	def := DefaultArchiveConfig()
	if cfg.MaxEntries <= 0 {
		cfg.MaxEntries = def.MaxEntries
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = def.MaxAge
	}
	return &Archive{cfg: cfg}
}

// Record appends an alert to the archive, evicting stale or excess entries.
func (a *Archive) Record(al alert.Alert) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.evict()
	if len(a.entries) >= a.cfg.MaxEntries {
		a.entries = a.entries[1:]
	}
	a.entries = append(a.entries, archiveEntry{Alert: al, RecordedAt: time.Now()})
}

// List returns all non-stale archived alerts.
func (a *Archive) List() []alert.Alert {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.evict()
	out := make([]alert.Alert, len(a.entries))
	for i, e := range a.entries {
		out[i] = e.Alert
	}
	return out
}

// Len returns the number of currently stored entries.
func (a *Archive) Len() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.evict()
	return len(a.entries)
}

// evict removes entries older than MaxAge. Must be called with mu held.
func (a *Archive) evict() {
	cutoff := time.Now().Add(-a.cfg.MaxAge)
	i := 0
	for i < len(a.entries) && a.entries[i].RecordedAt.Before(cutoff) {
		i++
	}
	a.entries = a.entries[i:]
}
