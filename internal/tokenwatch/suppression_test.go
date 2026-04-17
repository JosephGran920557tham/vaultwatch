package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultSuppressionConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultSuppressionConfig()
	if cfg.Window <= 0 {
		t.Fatal("expected positive window")
	}
}

func TestNewSuppression_ZeroWindow_UsesDefault(t *testing.T) {
	s := NewSuppression(SuppressionConfig{})
	if s.cfg.Window != DefaultSuppressionConfig().Window {
		t.Fatalf("expected default window, got %v", s.cfg.Window)
	}
}

func TestSuppression_Allow_FirstCallPermitted(t *testing.T) {
	s := NewSuppression(SuppressionConfig{Window: time.Minute})
	if !s.Allow("tok1:critical") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestSuppression_Allow_SecondCallSuppressed(t *testing.T) {
	s := NewSuppression(SuppressionConfig{Window: time.Minute})
	s.Allow("tok1:critical")
	if s.Allow("tok1:critical") {
		t.Fatal("expected second call within window to be suppressed")
	}
}

func TestSuppression_Allow_AfterWindowExpires_Permitted(t *testing.T) {
	now := time.Now()
	s := NewSuppression(SuppressionConfig{Window: time.Minute})
	s.nowFn = func() time.Time { return now }
	s.Allow("tok1:warning")
	s.nowFn = func() time.Time { return now.Add(2 * time.Minute) }
	if !s.Allow("tok1:warning") {
		t.Fatal("expected call after window expiry to be allowed")
	}
}

func TestSuppression_Allow_DifferentKeysIndependent(t *testing.T) {
	s := NewSuppression(SuppressionConfig{Window: time.Minute})
	s.Allow("tok1:critical")
	if !s.Allow("tok2:critical") {
		t.Fatal("expected different key to be allowed")
	}
}

func TestSuppression_Reset_ClearsKey(t *testing.T) {
	s := NewSuppression(SuppressionConfig{Window: time.Minute})
	s.Allow("tok1:critical")
	s.Reset("tok1:critical")
	if !s.Allow("tok1:critical") {
		t.Fatal("expected allow after reset")
	}
}

func TestSuppression_Len_TracksKeys(t *testing.T) {
	s := NewSuppression(SuppressionConfig{Window: time.Minute})
	s.Allow("a")
	s.Allow("b")
	s.Allow("a") // suppressed, not re-added
	if s.Len() != 2 {
		t.Fatalf("expected 2 keys, got %d", s.Len())
	}
}
