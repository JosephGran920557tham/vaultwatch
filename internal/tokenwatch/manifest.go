package tokenwatch

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// DefaultManifestConfig returns a ManifestConfig with sensible defaults.
func DefaultManifestConfig() ManifestConfig {
	return ManifestConfig{
		MaxAge: 24 * time.Hour,
	}
}

// ManifestConfig controls the behaviour of the Manifest store.
type ManifestConfig struct {
	MaxAge time.Duration
}

// ManifestEntry records a single token registration event.
type ManifestEntry struct {
	TokenID     string
	RegisteredAt time.Time
	Meta        map[string]string
}

// Manifest tracks all tokens that have ever been registered, with optional
// expiry of old entries based on MaxAge.
type Manifest struct {
	mu      sync.RWMutex
	cfg     ManifestConfig
	entries map[string]ManifestEntry
}

// NewManifest creates a new Manifest. Zero-value config fields are replaced
// with defaults.
func NewManifest(cfg ManifestConfig) *Manifest {
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = DefaultManifestConfig().MaxAge
	}
	return &Manifest{
		cfg:     cfg,
		entries: make(map[string]ManifestEntry),
	}
}

// Register adds or updates a token entry in the manifest.
func (m *Manifest) Register(id string, meta map[string]string) error {
	if id == "" {
		return fmt.Errorf("manifest: token id must not be empty")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[id] = ManifestEntry{
		TokenID:      id,
		RegisteredAt: time.Now(),
		Meta:         meta,
	}
	return nil
}

// Get returns the ManifestEntry for the given token id, and whether it exists
// and has not expired.
func (m *Manifest) Get(id string) (ManifestEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.entries[id]
	if !ok {
		return ManifestEntry{}, false
	}
	if time.Since(e.RegisteredAt) > m.cfg.MaxAge {
		return ManifestEntry{}, false
	}
	return e, true
}

// List returns all non-expired entries sorted by registration time (oldest first).
func (m *Manifest) List() []ManifestEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cutoff := time.Now().Add(-m.cfg.MaxAge)
	out := make([]ManifestEntry, 0, len(m.entries))
	for _, e := range m.entries {
		if e.RegisteredAt.After(cutoff) {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].RegisteredAt.Before(out[j].RegisteredAt)
	})
	return out
}

// Len returns the number of non-expired entries.
func (m *Manifest) Len() int {
	return len(m.List())
}
