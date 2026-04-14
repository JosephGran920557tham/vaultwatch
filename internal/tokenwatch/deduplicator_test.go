package tokenwatch

import (
	"testing"
	"time"
)

func TestNewDeduplicator_DefaultWindow(t *testing.T) {
	d := NewDeduplicator(0)
	if d.window != 5*time.Minute {
		t.Fatalf("expected default window 5m, got %v", d.window)
	}
}

func TestDeduplicator_Allow_FirstCallPermitted(t *testing.T) {
	d := NewDeduplicator(1 * time.Minute)
	if !d.Allow("token-a") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestDeduplicator_Allow_SecondCallSuppressed(t *testing.T) {
	d := NewDeduplicator(1 * time.Minute)
	d.Allow("token-a")
	if d.Allow("token-a") {
		t.Fatal("expected second call within window to be suppressed")
	}
}

func TestDeduplicator_Allow_AfterWindowExpires(t *testing.T) {
	base := time.Now()
	d := NewDeduplicator(1 * time.Minute)
	d.now = func() time.Time { return base }

	d.Allow("token-a")

	// advance past the window
	d.now = func() time.Time { return base.Add(2 * time.Minute) }
	if !d.Allow("token-a") {
		t.Fatal("expected call after window expiry to be allowed")
	}
}

func TestDeduplicator_Allow_DifferentKeysIndependent(t *testing.T) {
	d := NewDeduplicator(1 * time.Minute)
	d.Allow("token-a")
	if !d.Allow("token-b") {
		t.Fatal("expected different key to be allowed independently")
	}
}

func TestDeduplicator_SuppressedCount(t *testing.T) {
	d := NewDeduplicator(1 * time.Minute)
	d.Allow("token-a") // allowed, count = 1
	d.Allow("token-a") // suppressed, count = 2
	d.Allow("token-a") // suppressed, count = 3

	if got := d.SuppressedCount("token-a"); got != 2 {
		t.Fatalf("expected 2 suppressed, got %d", got)
	}
}

func TestDeduplicator_Reset_AllowsImmediately(t *testing.T) {
	d := NewDeduplicator(1 * time.Minute)
	d.Allow("token-a")
	d.Reset("token-a")
	if !d.Allow("token-a") {
		t.Fatal("expected allow after reset")
	}
}

func TestDeduplicator_Purge_RemovesStaleEntries(t *testing.T) {
	base := time.Now()
	d := NewDeduplicator(1 * time.Minute)
	d.now = func() time.Time { return base }

	d.Allow("token-stale")
	d.now = func() time.Time { return base.Add(2 * time.Minute) }
	d.Purge()

	// After purge the key should be gone; Allow should return true again.
	if !d.Allow("token-stale") {
		t.Fatal("expected allow after purge of stale entry")
	}
}
