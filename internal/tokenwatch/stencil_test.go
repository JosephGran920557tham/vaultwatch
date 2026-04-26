package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultStencilConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultStencilConfig()
	if cfg.MaxAge <= 0 {
		t.Fatalf("expected positive MaxAge, got %v", cfg.MaxAge)
	}
	if cfg.MaxItems <= 0 {
		t.Fatalf("expected positive MaxItems, got %d", cfg.MaxItems)
	}
}

func TestNewStencil_ZeroValues_UsesDefaults(t *testing.T) {
	s := NewStencil(StencilConfig{})
	def := DefaultStencilConfig()
	if s.cfg.MaxAge != def.MaxAge {
		t.Errorf("expected MaxAge %v, got %v", def.MaxAge, s.cfg.MaxAge)
	}
	if s.cfg.MaxItems != def.MaxItems {
		t.Errorf("expected MaxItems %d, got %d", def.MaxItems, s.cfg.MaxItems)
	}
}

func TestStencil_Set_And_Get(t *testing.T) {
	s := NewStencil(DefaultStencilConfig())
	s.Set("tok-1", "alert: {{.LeaseID}} expiring")

	tmpl, ok := s.Get("tok-1")
	if !ok {
		t.Fatal("expected entry to be present")
	}
	if tmpl != "alert: {{.LeaseID}} expiring" {
		t.Errorf("unexpected template: %q", tmpl)
	}
}

func TestStencil_Get_Missing_ReturnsFalse(t *testing.T) {
	s := NewStencil(DefaultStencilConfig())
	_, ok := s.Get("does-not-exist")
	if ok {
		t.Fatal("expected false for missing key")
	}
}

func TestStencil_Get_Expired_ReturnsFalse(t *testing.T) {
	s := NewStencil(StencilConfig{MaxAge: 10 * time.Millisecond, MaxItems: 16})
	s.Set("tok-exp", "some template")
	time.Sleep(20 * time.Millisecond)

	_, ok := s.Get("tok-exp")
	if ok {
		t.Fatal("expected expired entry to return false")
	}
}

func TestStencil_Len_ExcludesExpired(t *testing.T) {
	s := NewStencil(StencilConfig{MaxAge: 10 * time.Millisecond, MaxItems: 16})
	s.Set("tok-a", "tmpl-a")
	s.Set("tok-b", "tmpl-b")
	time.Sleep(20 * time.Millisecond)
	s.Set("tok-c", "tmpl-c")

	if n := s.Len(); n != 1 {
		t.Errorf("expected Len 1 after expiry, got %d", n)
	}
}

func TestStencil_Set_RespectsMaxItems(t *testing.T) {
	s := NewStencil(StencilConfig{MaxAge: time.Hour, MaxItems: 2})
	s.Set("tok-1", "t1")
	s.Set("tok-2", "t2")
	s.Set("tok-3", "t3") // should be silently dropped

	if n := s.Len(); n > 2 {
		t.Errorf("expected at most 2 entries, got %d", n)
	}
}
