package tokenwatch

import (
	"sync"
	"time"
)

// DefaultStencilConfig returns a StencilConfig with sensible defaults.
func DefaultStencilConfig() StencilConfig {
	return StencilConfig{
		MaxAge:   30 * time.Minute,
		MaxItems: 512,
	}
}

// StencilConfig controls the behaviour of a Stencil.
type StencilConfig struct {
	MaxAge   time.Duration
	MaxItems int
}

// stencilEntry holds a template string and the time it was recorded.
type stencilEntry struct {
	template  string
	recordedAt time.Time
}

// Stencil stores per-token alert template strings and evicts entries that
// exceed MaxAge or cause the store to exceed MaxItems.
type Stencil struct {
	mu      sync.Mutex
	cfg     StencilConfig
	entries map[string]stencilEntry
}

// NewStencil creates a Stencil. Zero-value fields in cfg are replaced with
// defaults.
func NewStencil(cfg StencilConfig) *Stencil {
	def := DefaultStencilConfig()
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = def.MaxAge
	}
	if cfg.MaxItems <= 0 {
		cfg.MaxItems = def.MaxItems
	}
	return &Stencil{
		cfg:     cfg,
		entries: make(map[string]stencilEntry, cfg.MaxItems),
	}
}

// Set stores a template for the given token ID, evicting stale entries first.
func (s *Stencil) Set(tokenID, template string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.evict()
	if len(s.entries) >= s.cfg.MaxItems {
		return
	}
	s.entries[tokenID] = stencilEntry{template: template, recordedAt: time.Now()}
}

// Get retrieves the template for tokenID. Returns ("", false) if absent or
// expired.
func (s *Stencil) Get(tokenID string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[tokenID]
	if !ok {
		return "", false
	}
	if time.Since(e.recordedAt) > s.cfg.MaxAge {
		delete(s.entries, tokenID)
		return "", false
	}
	return e.template, true
}

// Len returns the number of non-expired entries currently stored.
func (s *Stencil) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.evict()
	return len(s.entries)
}

// evict removes all entries older than MaxAge. Must be called with s.mu held.
func (s *Stencil) evict() {
	now := time.Now()
	for id, e := range s.entries {
		if now.Sub(e.recordedAt) > s.cfg.MaxAge {
			delete(s.entries, id)
		}
	}
}
