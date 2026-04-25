package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultEmbargoConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultEmbargoConfig()
	if cfg.Window <= 0 {
		t.Fatalf("expected positive Window, got %v", cfg.Window)
	}
}

func TestNewEmbargo_ZeroWindow_UsesDefault(t *testing.T) {
	e := NewEmbargo(EmbargoConfig{})
	if e.cfg.Window != DefaultEmbargoConfig().Window {
		t.Fatalf("expected default window, got %v", e.cfg.Window)
	}
}

func TestEmbargo_IsSuppressed_UnknownToken_ReturnsFalse(t *testing.T) {
	e := NewEmbargo(DefaultEmbargoConfig())
	if e.IsSuppressed("tok-unknown") {
		t.Fatal("expected false for unknown token")
	}
}

func TestEmbargo_Place_SuppressesToken(t *testing.T) {
	e := NewEmbargo(DefaultEmbargoConfig())
	e.Place("tok-1")
	if !e.IsSuppressed("tok-1") {
		t.Fatal("expected token to be suppressed after Place")
	}
}

func TestEmbargo_Lift_RemovesSuppression(t *testing.T) {
	e := NewEmbargo(DefaultEmbargoConfig())
	e.Place("tok-1")
	e.Lift("tok-1")
	if e.IsSuppressed("tok-1") {
		t.Fatal("expected token to be unsuppressed after Lift")
	}
}

func TestEmbargo_IsSuppressed_AfterWindowExpires_ReturnsFalse(t *testing.T) {
	e := NewEmbargo(EmbargoConfig{Window: 5 * time.Minute})
	now := time.Now()
	e.now = func() time.Time { return now }
	e.Place("tok-2")

	e.now = func() time.Time { return now.Add(6 * time.Minute) }
	if e.IsSuppressed("tok-2") {
		t.Fatal("expected suppression to expire after window")
	}
}

func TestEmbargo_Len_CountsActiveEmbargoes(t *testing.T) {
	e := NewEmbargo(DefaultEmbargoConfig())
	e.Place("tok-a")
	e.Place("tok-b")
	if got := e.Len(); got != 2 {
		t.Fatalf("expected 2 active embargoes, got %d", got)
	}
}

func TestEmbargo_Len_ExcludesExpired(t *testing.T) {
	e := NewEmbargo(EmbargoConfig{Window: 5 * time.Minute})
	now := time.Now()
	e.now = func() time.Time { return now }
	e.Place("tok-x")
	e.Place("tok-y")

	e.now = func() time.Time { return now.Add(10 * time.Minute) }
	if got := e.Len(); got != 0 {
		t.Fatalf("expected 0 active embargoes after expiry, got %d", got)
	}
}

func TestEmbargo_DifferentTokensAreIndependent(t *testing.T) {
	e := NewEmbargo(DefaultEmbargoConfig())
	e.Place("tok-1")
	if e.IsSuppressed("tok-2") {
		t.Fatal("embargo on tok-1 should not affect tok-2")
	}
}
