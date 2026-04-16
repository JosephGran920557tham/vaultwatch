package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultShadowConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultShadowConfig()
	if cfg.TTL <= 0 {
		t.Fatal("expected positive TTL")
	}
}

func TestNewShadowRegistry_ZeroTTL_UsesDefault(t *testing.T) {
	r := NewShadowRegistry(ShadowConfig{})
	if r.cfg.TTL != DefaultShadowConfig().TTL {
		t.Errorf("expected default TTL, got %v", r.cfg.TTL)
	}
}

func TestShadowRegistry_SetAndGet(t *testing.T) {
	r := NewShadowRegistry(DefaultShadowConfig())
	r.Set("tok1", 30*time.Second)
	e, ok := r.Get("tok1")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.TTL != 30*time.Second {
		t.Errorf("expected 30s, got %v", e.TTL)
	}
}

func TestShadowRegistry_Get_Missing(t *testing.T) {
	r := NewShadowRegistry(DefaultShadowConfig())
	_, ok := r.Get("missing")
	if ok {
		t.Fatal("expected missing entry")
	}
}

func TestShadowRegistry_Get_Expired(t *testing.T) {
	r := NewShadowRegistry(ShadowConfig{TTL: time.Millisecond})
	r.Set("tok1", 10*time.Second)
	time.Sleep(5 * time.Millisecond)
	_, ok := r.Get("tok1")
	if ok {
		t.Fatal("expected expired entry to be missing")
	}
}

func TestShadowRegistry_Delete(t *testing.T) {
	r := NewShadowRegistry(DefaultShadowConfig())
	r.Set("tok1", 10*time.Second)
	r.Delete("tok1")
	_, ok := r.Get("tok1")
	if ok {
		t.Fatal("expected deleted entry to be missing")
	}
}

func TestShadowRegistry_Purge_RemovesExpired(t *testing.T) {
	r := NewShadowRegistry(ShadowConfig{TTL: time.Millisecond})
	r.Set("tok1", 10*time.Second)
	r.Set("tok2", 10*time.Second)
	time.Sleep(5 * time.Millisecond)
	r.Purge()
	if len(r.entries) != 0 {
		t.Errorf("expected 0 entries after purge, got %d", len(r.entries))
	}
}

func TestShadowRegistry_Purge_KeepsFresh(t *testing.T) {
	r := NewShadowRegistry(ShadowConfig{TTL: time.Hour})
	r.Set("tok1", 10*time.Second)
	r.Purge()
	if len(r.entries) != 1 {
		t.Errorf("expected 1 entry after purge, got %d", len(r.entries))
	}
}
