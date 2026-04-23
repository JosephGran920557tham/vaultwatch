package tokenwatch

import (
	"sync"
	"time"
)

// DefaultAffinityConfig returns a sensible AffinityConfig.
func DefaultAffinityConfig() AffinityConfig {
	return AffinityConfig{
		MaxAge:    30 * time.Minute,
		MinHits:   3,
		DecayRate: 0.5,
	}
}

// AffinityConfig controls how token affinity is tracked.
type AffinityConfig struct {
	// MaxAge is how long an affinity entry is retained.
	MaxAge time.Duration
	// MinHits is the minimum number of observations before affinity is established.
	MinHits int
	// DecayRate is the fraction by which the affinity score decays each cycle (0–1).
	DecayRate float64
}

type affinityEntry struct {
	hits      int
	score     float64
	lastSeen  time.Time
}

// AffinityDetector tracks how frequently each token is observed and
// computes a normalised affinity score that decays over time.
type AffinityDetector struct {
	mu     sync.Mutex
	cfg    AffinityConfig
	entries map[string]*affinityEntry
}

// NewAffinityDetector creates an AffinityDetector with the given config.
// Zero-value fields are replaced with defaults.
func NewAffinityDetector(cfg AffinityConfig) *AffinityDetector {
	def := DefaultAffinityConfig()
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = def.MaxAge
	}
	if cfg.MinHits <= 0 {
		cfg.MinHits = def.MinHits
	}
	if cfg.DecayRate <= 0 || cfg.DecayRate > 1 {
		cfg.DecayRate = def.DecayRate
	}
	return &AffinityDetector{
		cfg:     cfg,
		entries: make(map[string]*affinityEntry),
	}
}

// Observe records a new observation for the given token ID and
// returns the current affinity score (0.0–1.0).
func (a *AffinityDetector) Observe(tokenID string) float64 {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	e, ok := a.entries[tokenID]
	if !ok {
		e = &affinityEntry{}
		a.entries[tokenID] = e
	}

	// Decay score if entry already exists.
	if ok && !e.lastSeen.IsZero() {
		elapsed := now.Sub(e.lastSeen).Seconds() / a.cfg.MaxAge.Seconds()
		e.score *= (1.0 - a.cfg.DecayRate*elapsed)
		if e.score < 0 {
			e.score = 0
		}
	}

	e.hits++
	e.lastSeen = now

	if e.hits >= a.cfg.MinHits {
		e.score += 1.0 / float64(a.cfg.MinHits)
		if e.score > 1.0 {
			e.score = 1.0
		}
	}

	return e.score
}

// Evict removes entries that have exceeded MaxAge.
func (a *AffinityDetector) Evict() {
	a.mu.Lock()
	defer a.mu.Unlock()

	cutoff := time.Now().Add(-a.cfg.MaxAge)
	for id, e := range a.entries {
		if e.lastSeen.Before(cutoff) {
			delete(a.entries, id)
		}
	}
}

// Score returns the current affinity score for a token without recording an observation.
func (a *AffinityDetector) Score(tokenID string) float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	if e, ok := a.entries[tokenID]; ok {
		return e.score
	}
	return 0
}
