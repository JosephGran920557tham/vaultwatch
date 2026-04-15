package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultBurstConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultBurstConfig()
	if cfg.Window <= 0 {
		t.Errorf("expected positive Window, got %s", cfg.Window)
	}
	if cfg.MaxEvents <= 0 {
		t.Errorf("expected positive MaxEvents, got %d", cfg.MaxEvents)
	}
}

func TestNewBurstDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d, err := NewBurstDetector(BurstConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	def := DefaultBurstConfig()
	if d.cfg.Window != def.Window {
		t.Errorf("expected window %s, got %s", def.Window, d.cfg.Window)
	}
	if d.cfg.MaxEvents != def.MaxEvents {
		t.Errorf("expected maxEvents %d, got %d", def.MaxEvents, d.cfg.MaxEvents)
	}
}

func TestBurstDetector_NotBursting_WhenUnderLimit(t *testing.T) {
	d, _ := NewBurstDetector(BurstConfig{Window: time.Minute, MaxEvents: 5})
	now := time.Now()
	for i := 0; i < 5; i++ {
		d.Record("tok", now)
	}
	if d.IsBursting("tok", now) {
		t.Error("expected not bursting when at limit")
	}
}

func TestBurstDetector_Bursting_WhenOverLimit(t *testing.T) {
	d, _ := NewBurstDetector(BurstConfig{Window: time.Minute, MaxEvents: 3})
	now := time.Now()
	for i := 0; i < 4; i++ {
		d.Record("tok", now)
	}
	if !d.IsBursting("tok", now) {
		t.Error("expected bursting when over limit")
	}
}

func TestBurstDetector_PrunesOldEvents(t *testing.T) {
	d, _ := NewBurstDetector(BurstConfig{Window: time.Second, MaxEvents: 2})
	old := time.Now().Add(-2 * time.Second)
	now := time.Now()
	for i := 0; i < 5; i++ {
		d.Record("tok", old)
	}
	if d.IsBursting("tok", now) {
		t.Error("expected old events to be pruned")
	}
}

func TestBurstDetector_Count_ReturnsWindowCount(t *testing.T) {
	d, _ := NewBurstDetector(BurstConfig{Window: time.Minute, MaxEvents: 100})
	now := time.Now()
	for i := 0; i < 7; i++ {
		d.Record("tok", now)
	}
	if c := d.Count("tok", now); c != 7 {
		t.Errorf("expected count 7, got %d", c)
	}
}

func TestBurstDetector_DifferentKeys_Independent(t *testing.T) {
	d, _ := NewBurstDetector(BurstConfig{Window: time.Minute, MaxEvents: 2})
	now := time.Now()
	for i := 0; i < 3; i++ {
		d.Record("a", now)
	}
	if d.IsBursting("b", now) {
		t.Error("key b should not be bursting")
	}
}
