package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultDebounceConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultDebounceConfig()
	if cfg.Wait <= 0 {
		t.Fatal("expected positive Wait")
	}
}

func TestNewDebounce_ZeroWait_UsesDefault(t *testing.T) {
	d := NewDebounce(DebounceConfig{})
	if d.cfg.Wait != DefaultDebounceConfig().Wait {
		t.Fatalf("expected default wait, got %v", d.cfg.Wait)
	}
}

func TestDebounce_Allow_FirstCallPermitted(t *testing.T) {
	d := NewDebounce(DebounceConfig{Wait: time.Second})
	if !d.Allow("tok1") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestDebounce_Allow_SecondCallSuppressed(t *testing.T) {
	d := NewDebounce(DebounceConfig{Wait: time.Second})
	d.Allow("tok1")
	if d.Allow("tok1") {
		t.Fatal("expected second call within wait to be suppressed")
	}
}

func TestDebounce_Allow_AfterWait_Permitted(t *testing.T) {
	now := time.Now()
	d := NewDebounce(DebounceConfig{Wait: 5 * time.Second})
	d.now = func() time.Time { return now }
	d.Allow("tok1")
	d.now = func() time.Time { return now.Add(6 * time.Second) }
	if !d.Allow("tok1") {
		t.Fatal("expected call after wait to be allowed")
	}
}

func TestDebounce_Allow_DifferentKeysIndependent(t *testing.T) {
	d := NewDebounce(DebounceConfig{Wait: time.Second})
	d.Allow("a")
	if !d.Allow("b") {
		t.Fatal("expected different key to be allowed")
	}
}

func TestDebounce_Reset_ClearsKey(t *testing.T) {
	d := NewDebounce(DebounceConfig{Wait: time.Minute})
	d.Allow("tok1")
	d.Reset("tok1")
	if !d.Allow("tok1") {
		t.Fatal("expected allow after reset")
	}
}

func TestDebounce_Len_TracksKeys(t *testing.T) {
	d := NewDebounce(DebounceConfig{Wait: time.Minute})
	d.Allow("a")
	d.Allow("b")
	if d.Len() != 2 {
		t.Fatalf("expected 2 keys, got %d", d.Len())
	}
}
