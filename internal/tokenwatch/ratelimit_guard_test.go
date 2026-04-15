package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultRateLimitGuardConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultRateLimitGuardConfig()
	if cfg.Window <= 0 {
		t.Errorf("expected positive window, got %v", cfg.Window)
	}
	if cfg.MaxOps < 1 {
		t.Errorf("expected MaxOps >= 1, got %d", cfg.MaxOps)
	}
}

func TestNewRateLimitGuard_InvalidMaxOps(t *testing.T) {
	_, err := NewRateLimitGuard(RateLimitGuardConfig{MaxOps: 0, Window: time.Minute})
	if err == nil {
		t.Fatal("expected error for MaxOps=0")
	}
}

func TestNewRateLimitGuard_InvalidWindow(t *testing.T) {
	_, err := NewRateLimitGuard(RateLimitGuardConfig{MaxOps: 5, Window: 0})
	if err == nil {
		t.Fatal("expected error for Window=0")
	}
}

func TestRateLimitGuard_Allow_FirstCallPermitted(t *testing.T) {
	g, err := NewRateLimitGuard(RateLimitGuardConfig{MaxOps: 3, Window: time.Minute})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !g.Allow("token-1") {
		t.Error("expected first call to be allowed")
	}
}

func TestRateLimitGuard_Allow_ExhaustsLimit(t *testing.T) {
	g, _ := NewRateLimitGuard(RateLimitGuardConfig{MaxOps: 2, Window: time.Minute})
	g.Allow("tok")
	g.Allow("tok")
	if g.Allow("tok") {
		t.Error("expected third call to be denied")
	}
}

func TestRateLimitGuard_Allow_DifferentTokensIndependent(t *testing.T) {
	g, _ := NewRateLimitGuard(RateLimitGuardConfig{MaxOps: 1, Window: time.Minute})
	if !g.Allow("a") {
		t.Error("expected token-a first call allowed")
	}
	if !g.Allow("b") {
		t.Error("expected token-b first call allowed")
	}
	if g.Allow("a") {
		t.Error("expected token-a second call denied")
	}
}

func TestRateLimitGuard_Allow_AfterWindowExpires(t *testing.T) {
	now := time.Now()
	g, _ := NewRateLimitGuard(RateLimitGuardConfig{MaxOps: 1, Window: 100 * time.Millisecond})
	g.now = func() time.Time { return now }
	g.Allow("tok")

	// Advance time past the window.
	g.now = func() time.Time { return now.Add(200 * time.Millisecond) }
	if !g.Allow("tok") {
		t.Error("expected call to be allowed after window expires")
	}
}

func TestRateLimitGuard_Reset_ClearsState(t *testing.T) {
	g, _ := NewRateLimitGuard(RateLimitGuardConfig{MaxOps: 1, Window: time.Minute})
	g.Allow("tok")
	if g.Allow("tok") {
		t.Error("expected second call to be denied before reset")
	}
	g.Reset("tok")
	if !g.Allow("tok") {
		t.Error("expected call to be allowed after reset")
	}
}
