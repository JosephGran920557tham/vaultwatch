package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultWindowConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultWindowConfig()
	if cfg.Size <= 0 {
		t.Errorf("expected positive Size, got %v", cfg.Size)
	}
	if cfg.MaxEvents <= 0 {
		t.Errorf("expected positive MaxEvents, got %d", cfg.MaxEvents)
	}
}

func TestNewWindow_InvalidSize(t *testing.T) {
	_, err := NewWindow(WindowConfig{Size: 0, MaxEvents: 5})
	if err == nil {
		t.Fatal("expected error for zero Size")
	}
}

func TestNewWindow_InvalidMaxEvents(t *testing.T) {
	_, err := NewWindow(WindowConfig{Size: time.Minute, MaxEvents: 0})
	if err == nil {
		t.Fatal("expected error for zero MaxEvents")
	}
}

func TestWindow_Allow_FirstCallPermitted(t *testing.T) {
	w, _ := NewWindow(WindowConfig{Size: time.Minute, MaxEvents: 3})
	if !w.Allow("tok1") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestWindow_Allow_ExhaustsLimit(t *testing.T) {
	w, _ := NewWindow(WindowConfig{Size: time.Minute, MaxEvents: 2})
	if !w.Allow("tok") {
		t.Fatal("expected first allow")
	}
	if !w.Allow("tok") {
		t.Fatal("expected second allow")
	}
	if w.Allow("tok") {
		t.Fatal("expected third call to be denied")
	}
}

func TestWindow_Allow_DifferentKeysIndependent(t *testing.T) {
	w, _ := NewWindow(WindowConfig{Size: time.Minute, MaxEvents: 1})
	w.Allow("a")
	if !w.Allow("b") {
		t.Fatal("expected key b to be independent of key a")
	}
}

func TestWindow_Count_ReflectsRecordedEvents(t *testing.T) {
	w, _ := NewWindow(WindowConfig{Size: time.Minute, MaxEvents: 10})
	w.Allow("x")
	w.Allow("x")
	if got := w.Count("x"); got != 2 {
		t.Errorf("expected count 2, got %d", got)
	}
}

func TestWindow_Count_PrunesExpiredEvents(t *testing.T) {
	w, _ := NewWindow(WindowConfig{Size: time.Second, MaxEvents: 10})

	past := time.Now().Add(-2 * time.Second)
	w.mu.Lock()
	w.events["y"] = []time.Time{past, past}
	w.mu.Unlock()

	if got := w.Count("y"); got != 0 {
		t.Errorf("expected 0 after expiry, got %d", got)
	}
}

func TestWindow_Reset_ClearsKey(t *testing.T) {
	w, _ := NewWindow(WindowConfig{Size: time.Minute, MaxEvents: 2})
	w.Allow("z")
	w.Allow("z")
	w.Reset("z")
	if got := w.Count("z"); got != 0 {
		t.Errorf("expected 0 after reset, got %d", got)
	}
	if !w.Allow("z") {
		t.Fatal("expected allow after reset")
	}
}
