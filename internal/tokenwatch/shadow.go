package tokenwatch

import (
	"sync"
	"time"
)

// ShadowConfig holds configuration for the shadow registry.
type ShadowConfig struct {
	TTL time.Duration
}

// DefaultShadowConfig returns sensible defaults.
func DefaultShadowConfig() ShadowConfig {
	return ShadowConfig{
		TTL: 5 * time.Minute,
	}
}

// ShadowEntry records a previously seen token state.
type ShadowEntry struct {
	TokenID   string
	TTL       time.Duration
	RecordedAt time.Time
}

// ShadowRegistry stores shadow copies of token states for comparison.
type ShadowRegistry struct {
	mu      sync.RWMutex
	entries map[string]ShadowEntry
	cfg     ShadowConfig
}

// NewShadowRegistry creates a new ShadowRegistry.
func NewShadowRegistry(cfg ShadowConfig) *ShadowRegistry {
	if cfg.TTL <= 0 {
		cfg.TTL = DefaultShadowConfig().TTL
	}
	return &ShadowRegistry{
		entries: make(map[string]ShadowEntry),
		cfg:     cfg,
	}
}

// Set stores or updates a shadow entry for a token.
func (s *ShadowRegistry) Set(tokenID string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[tokenID] = ShadowEntry{
		TokenID:    tokenID,
		TTL:        ttl,
		RecordedAt: time.Now(),
	}
}

// Get retrieves a shadow entry, returning false if missing or expired.
func (s *ShadowRegistry) Get(tokenID string) (ShadowEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[tokenID]
	if !ok {
		return ShadowEntry{}, false
	}
	if time.Since(e.RecordedAt) > s.cfg.TTL {
		return ShadowEntry{}, false
	}
	return e, true
}

// Delete removes a shadow entry.
func (s *ShadowRegistry) Delete(tokenID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, tokenID)
}

// Purge removes all expired entries.
func (s *ShadowRegistry) Purge() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, e := range s.entries {
		if now.Sub(e.RecordedAt) > s.cfg.TTL {
			delete(s.entries, id)
		}
	}
}
