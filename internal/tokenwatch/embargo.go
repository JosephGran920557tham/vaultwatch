package tokenwatch

import (
	"sync"
	"time"
)

// DefaultEmbargoConfig returns a sensible default EmbargoConfig.
func DefaultEmbargoConfig() EmbargoConfig {
	return EmbargoConfig{
		Window: 10 * time.Minute,
	}
}

// EmbargoConfig controls the embargo suppression window.
type EmbargoConfig struct {
	// Window is the duration during which alerts for a token are suppressed
	// after an embargo is placed.
	Window time.Duration
}

// Embargo suppresses alerts for a token for a configurable window after the
// token is explicitly embargoed. This is useful for planned maintenance or
// known-bad periods where alerts would be noise.
type Embargo struct {
	mu      sync.Mutex
	cfg     EmbargoConfig
	records map[string]time.Time
	now     func() time.Time
}

// NewEmbargo creates a new Embargo with the given config. Zero-value fields
// are replaced with defaults.
func NewEmbargo(cfg EmbargoConfig) *Embargo {
	def := DefaultEmbargoConfig()
	if cfg.Window <= 0 {
		cfg.Window = def.Window
	}
	return &Embargo{
		cfg:     cfg,
		records: make(map[string]time.Time),
		now:     time.Now,
	}
}

// Place sets an embargo on the given token ID, suppressing alerts for the
// configured window starting now.
func (e *Embargo) Place(tokenID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.records[tokenID] = e.now().Add(e.cfg.Window)
}

// Lift removes an active embargo for the given token ID before it expires.
func (e *Embargo) Lift(tokenID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.records, tokenID)
}

// IsSuppressed reports whether the token is currently under embargo.
func (e *Embargo) IsSuppressed(tokenID string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	expiry, ok := e.records[tokenID]
	if !ok {
		return false
	}
	if e.now().After(expiry) {
		delete(e.records, tokenID)
		return false
	}
	return true
}

// Len returns the number of currently active (non-expired) embargoes.
func (e *Embargo) Len() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := e.now()
	count := 0
	for id, expiry := range e.records {
		if now.After(expiry) {
			delete(e.records, id)
			continue
		}
		count++
	}
	return count
}
