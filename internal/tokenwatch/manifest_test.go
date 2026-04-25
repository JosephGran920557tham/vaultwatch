package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultManifestConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultManifestConfig()
	if cfg.MaxAge <= 0 {
		t.Fatalf("expected positive MaxAge, got %v", cfg.MaxAge)
	}
}

func TestNewManifest_ZeroMaxAge_UsesDefault(t *testing.T) {
	m := NewManifest(ManifestConfig{})
	if m.cfg.MaxAge <= 0 {
		t.Fatalf("expected default MaxAge, got %v", m.cfg.MaxAge)
	}
}

func TestManifest_Register_And_Get(t *testing.T) {
	m := NewManifest(DefaultManifestConfig())
	meta := map[string]string{"env": "prod"}
	if err := m.Register("tok-1", meta); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := m.Get("tok-1")
	if !ok {
		t.Fatal("expected entry to be present")
	}
	if e.TokenID != "tok-1" {
		t.Errorf("expected tok-1, got %s", e.TokenID)
	}
	if e.Meta["env"] != "prod" {
		t.Errorf("expected meta env=prod, got %v", e.Meta)
	}
}

func TestManifest_Get_Missing_ReturnsFalse(t *testing.T) {
	m := NewManifest(DefaultManifestConfig())
	_, ok := m.Get("nonexistent")
	if ok {
		t.Fatal("expected false for missing token")
	}
}

func TestManifest_Get_Expired_ReturnsFalse(t *testing.T) {
	m := NewManifest(ManifestConfig{MaxAge: time.Millisecond})
	_ = m.Register("tok-exp", nil)
	time.Sleep(5 * time.Millisecond)
	_, ok := m.Get("tok-exp")
	if ok {
		t.Fatal("expected expired entry to be absent")
	}
}

func TestManifest_Register_EmptyID_ReturnsError(t *testing.T) {
	m := NewManifest(DefaultManifestConfig())
	if err := m.Register("", nil); err == nil {
		t.Fatal("expected error for empty token id")
	}
}

func TestManifest_List_SortedByRegistration(t *testing.T) {
	m := NewManifest(DefaultManifestConfig())
	for _, id := range []string{"c", "a", "b"} {
		_ = m.Register(id, nil)
		time.Sleep(time.Millisecond)
	}
	list := m.List()
	if len(list) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(list))
	}
	for i := 1; i < len(list); i++ {
		if list[i].RegisteredAt.Before(list[i-1].RegisteredAt) {
			t.Errorf("list not sorted at index %d", i)
		}
	}
}

func TestManifest_Len_CountsNonExpired(t *testing.T) {
	m := NewManifest(DefaultManifestConfig())
	_ = m.Register("x", nil)
	_ = m.Register("y", nil)
	if m.Len() != 2 {
		t.Errorf("expected 2, got %d", m.Len())
	}
}
