package tokenwatch

import (
	"sync"
	"time"
)

// DefaultSuppressionConfig returns a SuppressionConfig with sensible defaults.
func DefaultSuppressionConfig() SuppressionConfig {
	return SuppressionConfig{
		Window: 10 * time.Minute,
	}
}

// SuppressionConfig controls how long a token alert is suppressed after first fire.
type SuppressionConfig struct {
	Window time.Duration
}

// Suppression prevents repeated alerts for the same token+level key within a window.
type Suppression struct {
	mu     sync.Mutex
	cfg    SuppressionConfig
	seen   map[string]time.Time
	nowFn  func() time.Time
}

// NewSuppression creates a Suppression with the given config.
// Zero-value window is replaced with the default.
func NewSuppression(cfg SuppressionConfig) *Suppression {
	if cfg.Window <= 0 {
		cfg.Window = DefaultSuppressionConfig().Window
	}
	return &Suppression{
		cfg:   cfg,
		seen:  make(map[string]time.Time),
		nowFn: time.Now,
	}
}

// Allow returns true if the key has not been seen within the suppression window.
// It records the key if allowed.
func (s *Suppression) Allow(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.nowFn()
	if t, ok := s.seen[key]; ok && now.Sub(t) < s.cfg.Window {
		return false
	}
	s.seen[key] = now
	return true
}

// Reset clears the suppression state for a specific key.
func (s *Suppression) Reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.seen, key)
}

// Len returns the number of currently tracked keys.
func (s *Suppression) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.seen)
}
