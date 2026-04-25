package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultPledgeConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultPledgeConfig()
	if cfg.MaxAge <= 0 {
		t.Fatalf("expected positive MaxAge, got %v", cfg.MaxAge)
	}
}

func TestNewPledge_ZeroMaxAge_UsesDefault(t *testing.T) {
	p := NewPledge(PledgeConfig{})
	if p.cfg.MaxAge != DefaultPledgeConfig().MaxAge {
		t.Fatalf("expected default MaxAge, got %v", p.cfg.MaxAge)
	}
}

func TestPledge_Register_And_Get(t *testing.T) {
	p := NewPledge(DefaultPledgeConfig())
	p.Register("tok-1", "primary token")
	e, ok := p.Get("tok-1")
	if !ok {
		t.Fatal("expected pledge to be found")
	}
	if e.Token != "tok-1" {
		t.Errorf("expected token tok-1, got %q", e.Token)
	}
	if e.Note != "primary token" {
		t.Errorf("expected note 'primary token', got %q", e.Note)
	}
}

func TestPledge_Get_Missing_ReturnsFalse(t *testing.T) {
	p := NewPledge(DefaultPledgeConfig())
	_, ok := p.Get("nonexistent")
	if ok {
		t.Fatal("expected false for missing token")
	}
}

func TestPledge_Get_Expired_ReturnsFalse(t *testing.T) {
	p := NewPledge(PledgeConfig{MaxAge: time.Millisecond})
	p.Register("tok-exp", "expiring")
	time.Sleep(5 * time.Millisecond)
	_, ok := p.Get("tok-exp")
	if ok {
		t.Fatal("expected pledge to be expired")
	}
}

func TestPledge_Revoke_RemovesEntry(t *testing.T) {
	p := NewPledge(DefaultPledgeConfig())
	p.Register("tok-2", "secondary")
	p.Revoke("tok-2")
	_, ok := p.Get("tok-2")
	if ok {
		t.Fatal("expected pledge to be revoked")
	}
}

func TestPledge_Len_CountsActiveEntries(t *testing.T) {
	p := NewPledge(DefaultPledgeConfig())
	p.Register("a", "")
	p.Register("b", "")
	p.Register("c", "")
	if l := p.Len(); l != 3 {
		t.Fatalf("expected 3 entries, got %d", l)
	}
}

func TestPledge_Len_ExcludesExpired(t *testing.T) {
	p := NewPledge(PledgeConfig{MaxAge: time.Millisecond})
	p.Register("x", "")
	p.Register("y", "")
	time.Sleep(5 * time.Millisecond)
	if l := p.Len(); l != 0 {
		t.Fatalf("expected 0 active entries after expiry, got %d", l)
	}
}
