package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultSpikeConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultSpikeConfig()
	if cfg.Window <= 0 {
		t.Error("expected positive Window")
	}
	if cfg.MinSamples < 2 {
		t.Error("expected MinSamples >= 2")
	}
	if cfg.MaxIncrease <= 0 || cfg.MaxIncrease > 1 {
		t.Errorf("unexpected MaxIncrease: %v", cfg.MaxIncrease)
	}
}

func TestNewSpikeDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewSpikeDetector(SpikeConfig{})
	def := DefaultSpikeConfig()
	if d.cfg.Window != def.Window {
		t.Errorf("want Window=%v got %v", def.Window, d.cfg.Window)
	}
	if d.cfg.MinSamples != def.MinSamples {
		t.Errorf("want MinSamples=%d got %d", def.MinSamples, d.cfg.MinSamples)
	}
}

func TestSpikeDetector_Check_InsufficientSamples_ReturnsNil(t *testing.T) {
	d := NewSpikeDetector(SpikeConfig{Window: time.Minute, MinSamples: 3, MaxIncrease: 0.5})
	d.Record("tok1", 10*time.Minute)
	d.Record("tok1", 11*time.Minute)
	// only 2 samples, need 3
	if a := d.Check("tok1", 11*time.Minute); a != nil {
		t.Errorf("expected nil alert, got %v", a)
	}
}

func TestSpikeDetector_Check_NoSpike_ReturnsNil(t *testing.T) {
	d := NewSpikeDetector(SpikeConfig{Window: time.Minute, MinSamples: 3, MaxIncrease: 0.5})
	for i := 0; i < 3; i++ {
		d.Record("tok2", 10*time.Minute)
	}
	if a := d.Check("tok2", 10*time.Minute); a != nil {
		t.Errorf("expected nil alert, got %v", a)
	}
}

func TestSpikeDetector_Check_SpikeDetected_ReturnsWarning(t *testing.T) {
	d := NewSpikeDetector(SpikeConfig{Window: time.Minute, MinSamples: 3, MaxIncrease: 0.5})
	// seed with stable samples
	d.Record("tok3", 10*time.Minute)
	d.Record("tok3", 10*time.Minute)
	// sudden spike: 200% increase
	d.Record("tok3", 30*time.Minute)
	a := d.Check("tok3", 30*time.Minute)
	if a == nil {
		t.Fatal("expected spike alert, got nil")
	}
	if a.LeaseID != "tok3" {
		t.Errorf("unexpected LeaseID: %s", a.LeaseID)
	}
}

func TestSpikeDetector_OldSamplesEvicted(t *testing.T) {
	d := NewSpikeDetector(SpikeConfig{Window: 1 * time.Millisecond, MinSamples: 2, MaxIncrease: 0.5})
	d.Record("tok4", 10*time.Minute)
	time.Sleep(5 * time.Millisecond)
	// old sample evicted, only 1 new sample
	d.Record("tok4", 20*time.Minute)
	if a := d.Check("tok4", 20*time.Minute); a != nil {
		t.Errorf("expected nil after eviction, got %v", a)
	}
}
