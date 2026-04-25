package tokenwatch

import (
	"sync"
	"time"

	"github.com/vaultwatch/internal/alert"
)

// DefaultSignalConfig returns a SignalConfig with sensible defaults.
func DefaultSignalConfig() SignalConfig {
	return SignalConfig{
		MinStrength: 2,
		DecayWindow: 5 * time.Minute,
	}
}

// SignalConfig controls signal aggregation behaviour.
type SignalConfig struct {
	// MinStrength is the minimum number of corroborating alerts before
	// a signal is considered actionable.
	MinStrength int
	// DecayWindow is the duration after which accumulated signal strength
	// resets to zero.
	DecayWindow time.Duration
}

// signalEntry tracks accumulated evidence for a single token.
type signalEntry struct {
	strength  int
	resetAt   time.Time
	lastLevel alert.Level
}

// SignalAggregator accumulates alert signals per token and suppresses
// noise until a configurable strength threshold is reached.
type SignalAggregator struct {
	cfg     SignalConfig
	mu      sync.Mutex
	entries map[string]*signalEntry
}

// NewSignalAggregator creates a SignalAggregator with the given config.
// Zero-value fields fall back to defaults.
func NewSignalAggregator(cfg SignalConfig) *SignalAggregator {
	def := DefaultSignalConfig()
	if cfg.MinStrength <= 0 {
		cfg.MinStrength = def.MinStrength
	}
	if cfg.DecayWindow <= 0 {
		cfg.DecayWindow = def.DecayWindow
	}
	return &SignalAggregator{
		cfg:     cfg,
		entries: make(map[string]*signalEntry),
	}
}

// Observe records an alert signal for the given token ID and returns
// true when the accumulated strength meets or exceeds MinStrength.
func (s *SignalAggregator) Observe(tokenID string, lvl alert.Level) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	e, ok := s.entries[tokenID]
	if !ok || now.After(e.resetAt) {
		e = &signalEntry{resetAt: now.Add(s.cfg.DecayWindow)}
		s.entries[tokenID] = e
	}
	e.strength++
	e.lastLevel = lvl
	return e.strength >= s.cfg.MinStrength
}

// Reset clears accumulated signal for a token.
func (s *SignalAggregator) Reset(tokenID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, tokenID)
}

// Strength returns the current signal strength for a token.
func (s *SignalAggregator) Strength(tokenID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[tokenID]
	if !ok || time.Now().After(e.resetAt) {
		return 0
	}
	return e.strength
}
