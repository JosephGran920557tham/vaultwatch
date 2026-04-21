package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultCensusConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultCensusConfig()
	if cfg.MaxAge <= 0 {
		t.Fatal("expected positive MaxAge")
	}
}

func TestNewCensus_ZeroMaxAge_UsesDefault(t *testing.T) {
	c := NewCensus(CensusConfig{})
	if c.cfg.MaxAge != DefaultCensusConfig().MaxAge {
		t.Fatalf("expected default MaxAge, got %v", c.cfg.MaxAge)
	}
}

func TestCensus_Observe_And_Active(t *testing.T) {
	c := NewCensus(DefaultCensusConfig())
	c.Observe("tok-1", map[string]string{"env": "prod"})
	c.Observe("tok-2", nil)

	active := c.Active()
	if len(active) != 2 {
		t.Fatalf("expected 2 active tokens, got %d", len(active))
	}
}

func TestCensus_Active_ExcludesStale(t *testing.T) {
	cfg := CensusConfig{MaxAge: 50 * time.Millisecond}
	c := NewCensus(cfg)
	c.Observe("tok-old", nil)

	time.Sleep(80 * time.Millisecond)
	c.Observe("tok-new", nil)

	active := c.Active()
	if len(active) != 1 {
		t.Fatalf("expected 1 active token, got %d", len(active))
	}
	if active[0] != "tok-new" {
		t.Fatalf("expected tok-new, got %s", active[0])
	}
}

func TestCensus_Len_ReturnsActiveCount(t *testing.T) {
	c := NewCensus(DefaultCensusConfig())
	if c.Len() != 0 {
		t.Fatal("expected 0 initially")
	}
	c.Observe("tok-a", nil)
	if c.Len() != 1 {
		t.Fatal("expected 1 after observe")
	}
}

func TestCensus_Evict_RemovesStaleEntries(t *testing.T) {
	cfg := CensusConfig{MaxAge: 30 * time.Millisecond}
	c := NewCensus(cfg)
	c.Observe("stale", nil)

	time.Sleep(50 * time.Millisecond)
	c.Observe("fresh", nil)
	c.Evict()

	c.mu.RLock()
	_, hasStale := c.entries["stale"]
	_, hasFresh := c.entries["fresh"]
	c.mu.RUnlock()

	if hasStale {
		t.Error("expected stale entry to be evicted")
	}
	if !hasFresh {
		t.Error("expected fresh entry to remain")
	}
}

func TestCensus_Observe_RefreshesExisting(t *testing.T) {
	cfg := CensusConfig{MaxAge: 50 * time.Millisecond}
	c := NewCensus(cfg)
	c.Observe("tok", nil)

	time.Sleep(30 * time.Millisecond)
	c.Observe("tok", map[string]string{"refreshed": "true"})

	time.Sleep(30 * time.Millisecond)
	// 60ms total but entry was refreshed at 30ms, so still within 50ms window
	active := c.Active()
	if len(active) != 1 {
		t.Fatalf("expected refreshed entry to still be active, got %d", len(active))
	}
}
