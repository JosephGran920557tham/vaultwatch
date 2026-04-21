package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultMirrorConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultMirrorConfig()
	if cfg.TTL <= 0 {
		t.Fatalf("expected positive TTL, got %s", cfg.TTL)
	}
	if cfg.MaxItems <= 0 {
		t.Fatalf("expected positive MaxItems, got %d", cfg.MaxItems)
	}
}

func TestNewMirror_ZeroValues_UsesDefaults(t *testing.T) {
	m := NewMirror(MirrorConfig{})
	def := DefaultMirrorConfig()
	if m.cfg.TTL != def.TTL {
		t.Fatalf("expected TTL %s, got %s", def.TTL, m.cfg.TTL)
	}
	if m.cfg.MaxItems != def.MaxItems {
		t.Fatalf("expected MaxItems %d, got %d", def.MaxItems, m.cfg.MaxItems)
	}
}

func TestMirror_Observe_And_Get(t *testing.T) {
	m := NewMirror(DefaultMirrorConfig())
	m.Observe("tok-1", 10*time.Minute)
	ttl, ok := m.Get("tok-1")
	if !ok {
		t.Fatal("expected entry to be present")
	}
	if ttl != 10*time.Minute {
		t.Fatalf("expected 10m, got %s", ttl)
	}
}

func TestMirror_Get_Missing(t *testing.T) {
	m := NewMirror(DefaultMirrorConfig())
	_, ok := m.Get("nonexistent")
	if ok {
		t.Fatal("expected missing entry")
	}
}

func TestMirror_Get_Expired(t *testing.T) {
	m := NewMirror(MirrorConfig{TTL: 1 * time.Millisecond, MaxItems: 10})
	m.Observe("tok-exp", 5*time.Minute)
	time.Sleep(5 * time.Millisecond)
	_, ok := m.Get("tok-exp")
	if ok {
		t.Fatal("expected expired entry to be absent")
	}
}

func TestMirror_Len_CountsLiveEntries(t *testing.T) {
	m := NewMirror(DefaultMirrorConfig())
	m.Observe("a", time.Minute)
	m.Observe("b", time.Minute)
	if m.Len() != 2 {
		t.Fatalf("expected 2, got %d", m.Len())
	}
}

func TestMirror_EvictsOldestWhenFull(t *testing.T) {
	m := NewMirror(MirrorConfig{TTL: time.Hour, MaxItems: 2})
	m.Observe("first", time.Minute)
	time.Sleep(time.Millisecond)
	m.Observe("second", time.Minute)
	m.Observe("third", time.Minute) // triggers eviction of "first"
	_, ok := m.Get("second")
	if !ok {
		t.Fatal("expected 'second' to survive eviction")
	}
	_, ok = m.Get("third")
	if !ok {
		t.Fatal("expected 'third' to be present")
	}
}
