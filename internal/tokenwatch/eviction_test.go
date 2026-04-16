package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultEvictionConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultEvictionConfig()
	if cfg.MaxAge <= 0 {
		t.Fatal("expected positive MaxAge")
	}
	if cfg.SweepInterval <= 0 {
		t.Fatal("expected positive SweepInterval")
	}
}

func TestNewEviction_ZeroValues_UsesDefaults(t *testing.T) {
	e := NewEviction(EvictionConfig{})
	def := DefaultEvictionConfig()
	if e.cfg.MaxAge != def.MaxAge {
		t.Errorf("expected MaxAge %v, got %v", def.MaxAge, e.cfg.MaxAge)
	}
	if e.cfg.SweepInterval != def.SweepInterval {
		t.Errorf("expected SweepInterval %v, got %v", def.SweepInterval, e.cfg.SweepInterval)
	}
}

func TestEviction_Touch_IncrementsSize(t *testing.T) {
	e := NewEviction(DefaultEvictionConfig())
	e.Touch("tok-a")
	e.Touch("tok-b")
	if e.Size() != 2 {
		t.Errorf("expected size 2, got %d", e.Size())
	}
}

func TestEviction_Touch_UpdatesExisting(t *testing.T) {
	e := NewEviction(DefaultEvictionConfig())
	e.Touch("tok-a")
	e.Touch("tok-a")
	if e.Size() != 1 {
		t.Errorf("expected size 1, got %d", e.Size())
	}
}

func TestEviction_Sweep_RemovesStale(t *testing.T) {
	e := NewEviction(EvictionConfig{MaxAge: 10 * time.Millisecond, SweepInterval: time.Minute})
	e.Touch("tok-old")
	time.Sleep(20 * time.Millisecond)
	e.Touch("tok-fresh")
	evicted := e.Sweep()
	if len(evicted) != 1 || evicted[0] != "tok-old" {
		t.Errorf("expected [tok-old] evicted, got %v", evicted)
	}
	if e.Size() != 1 {
		t.Errorf("expected size 1 after sweep, got %d", e.Size())
	}
}

func TestEviction_Sweep_NothingStale_ReturnsEmpty(t *testing.T) {
	e := NewEviction(EvictionConfig{MaxAge: time.Hour, SweepInterval: time.Minute})
	e.Touch("tok-a")
	evicted := e.Sweep()
	if len(evicted) != 0 {
		t.Errorf("expected no evictions, got %v", evicted)
	}
}

func TestEviction_Sweep_EmptyEntries_ReturnsEmpty(t *testing.T) {
	e := NewEviction(DefaultEvictionConfig())
	evicted := e.Sweep()
	if evicted != nil && len(evicted) != 0 {
		t.Errorf("expected empty eviction list, got %v", evicted)
	}
}
