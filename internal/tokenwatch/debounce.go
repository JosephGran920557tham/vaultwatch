package tokenwatch

import (
	"sync"
	"time"
)

// DefaultDebounceConfig returns sensible debounce defaults.
func DefaultDebounceConfig() DebounceConfig {
	return DebounceConfig{
		Wait: 10 * time.Second,
	}
}

// DebounceConfig controls debounce behaviour.
type DebounceConfig struct {
	Wait time.Duration
}

// Debounce suppresses repeated alerts for the same token until the wait
// period has elapsed since the last suppressed call.
type Debounce struct {
	cfg   DebounceConfig
	mu    sync.Mutex
	last  map[string]time.Time
	now   func() time.Time
}

// NewDebounce creates a Debounce. Zero Wait defaults to 10 s.
func NewDebounce(cfg DebounceConfig) *Debounce {
	if cfg.Wait <= 0 {
		cfg.Wait = DefaultDebounceConfig().Wait
	}
	return &Debounce{
		cfg:  cfg,
		last: make(map[string]time.Time),
		now:  time.Now,
	}
}

// Allow returns true the first time a token key is seen, and again only
// after the wait period has elapsed since it was last allowed.
func (d *Debounce) Allow(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.now()
	if t, ok := d.last[key]; ok && now.Sub(t) < d.cfg.Wait {
		return false
	}
	d.last[key] = now
	return true
}

// Reset clears the debounce state for a specific key.
func (d *Debounce) Reset(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.last, key)
}

// Len returns the number of tracked keys.
func (d *Debounce) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.last)
}
