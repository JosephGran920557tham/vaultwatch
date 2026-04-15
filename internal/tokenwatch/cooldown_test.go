package tokenwatch

import (
	"testing"
	"time"
)

func TestNewCooldown_DefaultWindowOnZero(t *testing.T) {
	c := NewCooldown(0)
	if c.window != DefaultCooldownConfig().Window {
		t.Errorf("expected default window %v, got %v", DefaultCooldownConfig().Window, c.window)
	}
}

func TestNewCooldown_DefaultWindowOnNegative(t *testing.T) {
	c := NewCooldown(-1 * time.Second)
	if c.window != DefaultCooldownConfig().Window {
		t.Errorf("expected default window, got %v", c.window)
	}
}

func TestCooldown_Allow_FirstCallPermitted(t *testing.T) {
	c := NewCooldown(1 * time.Minute)
	if !c.Allow("token-a") {
		t.Error("expected first call to be allowed")
	}
}

func TestCooldown_Allow_SecondCallSuppressed(t *testing.T) {
	c := NewCooldown(1 * time.Minute)
	c.Allow("token-a")
	if c.Allow("token-a") {
		t.Error("expected second call within window to be suppressed")
	}
}

func TestCooldown_Allow_AfterWindowExpires_Permitted(t *testing.T) {
	now := time.Now()
	c := NewCooldown(1 * time.Minute)
	c.now = func() time.Time { return now }
	c.Allow("token-a")

	// Advance time past the window.
	c.now = func() time.Time { return now.Add(2 * time.Minute) }
	if !c.Allow("token-a") {
		t.Error("expected call after window expiry to be allowed")
	}
}

func TestCooldown_Allow_DifferentKeysIndependent(t *testing.T) {
	c := NewCooldown(1 * time.Minute)
	c.Allow("token-a")
	if !c.Allow("token-b") {
		t.Error("expected different key to be allowed independently")
	}
}

func TestCooldown_Reset_ClearsKey(t *testing.T) {
	c := NewCooldown(1 * time.Minute)
	c.Allow("token-a")
	c.Reset("token-a")
	if !c.Allow("token-a") {
		t.Error("expected allow after reset")
	}
}

func TestCooldown_Len_TracksEntries(t *testing.T) {
	c := NewCooldown(1 * time.Minute)
	if c.Len() != 0 {
		t.Errorf("expected 0 entries, got %d", c.Len())
	}
	c.Allow("token-a")
	c.Allow("token-b")
	if c.Len() != 2 {
		t.Errorf("expected 2 entries, got %d", c.Len())
	}
	c.Reset("token-a")
	if c.Len() != 1 {
		t.Errorf("expected 1 entry after reset, got %d", c.Len())
	}
}
