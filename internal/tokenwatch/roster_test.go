package tokenwatch

import (
	"fmt"
	"testing"
	"time"
)

func TestDefaultRosterConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultRosterConfig()
	if cfg.MaxSize <= 0 {
		t.Errorf("MaxSize should be positive, got %d", cfg.MaxSize)
	}
	if cfg.EntryTTL <= 0 {
		t.Errorf("EntryTTL should be positive, got %v", cfg.EntryTTL)
	}
	if cfg.PrunePeriod <= 0 {
		t.Errorf("PrunePeriod should be positive, got %v", cfg.PrunePeriod)
	}
}

func TestNewRoster_ZeroValues_UsesDefaults(t *testing.T) {
	r := NewRoster(RosterConfig{})
	def := DefaultRosterConfig()
	if r.cfg.MaxSize != def.MaxSize {
		t.Errorf("expected MaxSize %d, got %d", def.MaxSize, r.cfg.MaxSize)
	}
	if r.cfg.EntryTTL != def.EntryTTL {
		t.Errorf("expected EntryTTL %v, got %v", def.EntryTTL, r.cfg.EntryTTL)
	}
}

func TestRoster_Touch_AddsEntry(t *testing.T) {
	r := NewRoster(DefaultRosterConfig())
	if err := r.Touch("tok-1", map[string]string{"env": "prod"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Len() != 1 {
		t.Errorf("expected Len 1, got %d", r.Len())
	}
}

func TestRoster_Touch_RefreshesExisting(t *testing.T) {
	r := NewRoster(DefaultRosterConfig())
	_ = r.Touch("tok-1", nil)
	_ = r.Touch("tok-1", nil)
	if r.Len() != 1 {
		t.Errorf("expected Len 1 after double-touch, got %d", r.Len())
	}
}

func TestRoster_Touch_CapacityError(t *testing.T) {
	r := NewRoster(RosterConfig{MaxSize: 2, EntryTTL: time.Hour, PrunePeriod: time.Minute})
	_ = r.Touch("tok-1", nil)
	_ = r.Touch("tok-2", nil)
	err := r.Touch("tok-3", nil)
	if err == nil {
		t.Fatal("expected capacity error, got nil")
	}
}

func TestRoster_Active_ExcludesStale(t *testing.T) {
	r := NewRoster(RosterConfig{MaxSize: 10, EntryTTL: 50 * time.Millisecond, PrunePeriod: time.Minute})
	_ = r.Touch("fresh", nil)
	// Inject a stale entry directly.
	r.mu.Lock()
	r.entries["stale"] = rosterEntry{TokenID: "stale", SeenAt: time.Now().Add(-time.Hour)}
	r.mu.Unlock()

	active := r.Active()
	if len(active) != 1 || active[0] != "fresh" {
		t.Errorf("expected only 'fresh', got %v", active)
	}
}

func TestRoster_Prune_RemovesStale(t *testing.T) {
	r := NewRoster(RosterConfig{MaxSize: 10, EntryTTL: 50 * time.Millisecond, PrunePeriod: time.Minute})
	for i := 0; i < 3; i++ {
		r.mu.Lock()
		r.entries[fmt.Sprintf("old-%d", i)] = rosterEntry{SeenAt: time.Now().Add(-time.Hour)}
		r.mu.Unlock()
	}
	_ = r.Touch("new", nil)

	removed := r.Prune()
	if removed != 3 {
		t.Errorf("expected 3 pruned, got %d", removed)
	}
	if r.Len() != 1 {
		t.Errorf("expected Len 1 after prune, got %d", r.Len())
	}
}

func TestRoster_Prune_NothingStale_ReturnsZero(t *testing.T) {
	r := NewRoster(DefaultRosterConfig())
	_ = r.Touch("tok-1", nil)
	if n := r.Prune(); n != 0 {
		t.Errorf("expected 0 pruned, got %d", n)
	}
}
