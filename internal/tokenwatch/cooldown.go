package tokenwatch

import (
	"sync"
	"time"
)

// CooldownConfig holds configuration for the cooldown tracker.
type CooldownConfig struct {
	// Window is the minimum duration between repeated actions for the same key.
	Window time.Duration
}

// DefaultCooldownConfig returns a CooldownConfig with sensible defaults.
func DefaultCooldownConfig() CooldownConfig {
	return CooldownConfig{
		Window: 5 * time.Minute,
	}
}

// Cooldown tracks per-key cooldown windows to suppress repeated actions.
type Cooldown struct {
	mu      sync.Mutex
	window  time.Duration
	lastAct map[string]time.Time
	now     func() time.Time
}

// NewCooldown creates a Cooldown with the given window duration.
// If window is zero or negative, DefaultCooldownConfig().Window is used.
func NewCooldown(window time.Duration) *Cooldown {
	if window <= 0 {
		window = DefaultCooldownConfig().Window
	}
	return &Cooldown{
		window:  window,
		lastAct: make(map[string]time.Time),
		now:     time.Now,
	}
}

// Allow returns true and records the action time if the key is not within
// its cooldown window. Returns false if the key was acted on recently.
func (c *Cooldown) Allow(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	if last, ok := c.lastAct[key]; ok {
		if now.Sub(last) < c.window {
			return false
		}
	}
	c.lastAct[key] = now
	return true
}

// Reset clears the cooldown record for the given key.
func (c *Cooldown) Reset(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.lastAct, key)
}

// Len returns the number of keys currently tracked.
func (c *Cooldown) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.lastAct)
}
