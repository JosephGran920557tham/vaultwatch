package tokenwatch

import (
	"errors"
	"sync"
	"time"
)

// WindowConfig holds configuration for a sliding time window counter.
type WindowConfig struct {
	// Size is the duration of the sliding window.
	Size time.Duration
	// MaxEvents is the maximum number of events allowed within the window.
	MaxEvents int
}

// DefaultWindowConfig returns a WindowConfig with sensible defaults.
func DefaultWindowConfig() WindowConfig {
	return WindowConfig{
		Size:      5 * time.Minute,
		MaxEvents: 10,
	}
}

// Window is a thread-safe sliding window event counter.
type Window struct {
	cfg    WindowConfig
	mu     sync.Mutex
	events map[string][]time.Time
	now    func() time.Time
}

// NewWindow creates a new Window. Returns an error if the config is invalid.
func NewWindow(cfg WindowConfig) (*Window, error) {
	if cfg.Size <= 0 {
		return nil, errors.New("window: Size must be positive")
	}
	if cfg.MaxEvents <= 0 {
		return nil, errors.New("window: MaxEvents must be positive")
	}
	return &Window{
		cfg:    cfg,
		events: make(map[string][]time.Time),
		now:    time.Now,
	}, nil
}

// Allow records an event for key and returns true if the count is within the
// configured MaxEvents for the sliding window. Expired events are pruned first.
func (w *Window) Allow(key string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.now()
	cutoff := now.Add(-w.cfg.Size)

	filtered := w.events[key][:0]
	for _, t := range w.events[key] {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= w.cfg.MaxEvents {
		w.events[key] = filtered
		return false
	}

	w.events[key] = append(filtered, now)
	return true
}

// Count returns the number of events recorded for key within the current window.
func (w *Window) Count(key string) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.now()
	cutoff := now.Add(-w.cfg.Size)
	count := 0
	for _, t := range w.events[key] {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}

// Reset clears all recorded events for key.
func (w *Window) Reset(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.events, key)
}
