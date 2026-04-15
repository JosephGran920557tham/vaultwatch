package tokenwatch

import (
	"context"
	"testing"
	"time"
)

func TestDefaultDrainConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultDrainConfig()
	if cfg.SampleWindow <= 0 {
		t.Fatal("expected positive SampleWindow")
	}
	if cfg.DrainThreshold <= 0 {
		t.Fatal("expected positive DrainThreshold")
	}
	if cfg.MinSamples <= 0 {
		t.Fatal("expected positive MinSamples")
	}
}

func TestNewDrainDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d, err := NewDrainDetector(DrainConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	def := DefaultDrainConfig()
	if d.cfg.SampleWindow != def.SampleWindow {
		t.Errorf("want %v got %v", def.SampleWindow, d.cfg.SampleWindow)
	}
}

func TestIsDraining_InsufficientSamples(t *testing.T) {
	d, _ := NewDrainDetector(DrainConfig{MinSamples: 3})
	ctx := context.Background()
	d.Record(ctx, "tok1", 300*time.Second)
	d.Record(ctx, "tok1", 290*time.Second)
	draining, _, err := d.IsDraining(ctx, "tok1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draining {
		t.Fatal("expected not draining with insufficient samples")
	}
}

func TestIsDraining_SlowLoss_NotDraining(t *testing.T) {
	cfg := DrainConfig{SampleWindow: time.Minute, DrainThreshold: 5.0, MinSamples: 2}
	d, _ := NewDrainDetector(cfg)
	ctx := context.Background()
	// Simulate ~1 s/s loss over 10 s window by manipulating samples directly.
	now := time.Now()
	d.mu.Lock()
	d.samples["tok2"] = []sample{
		{at: now.Add(-10 * time.Second), ttl: 310 * time.Second},
		{at: now, ttl: 300 * time.Second},
	}
	d.mu.Unlock()
	draining, rate, err := d.IsDraining(ctx, "tok2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draining {
		t.Errorf("expected not draining, rate=%.2f", rate)
	}
}

func TestIsDraining_FastLoss_IsDraining(t *testing.T) {
	cfg := DrainConfig{SampleWindow: time.Minute, DrainThreshold: 5.0, MinSamples: 2}
	d, _ := NewDrainDetector(cfg)
	ctx := context.Background()
	now := time.Now()
	d.mu.Lock()
	d.samples["tok3"] = []sample{
		{at: now.Add(-10 * time.Second), ttl: 400 * time.Second},
		{at: now, ttl: 300 * time.Second},
	}
	d.mu.Unlock()
	draining, rate, err := d.IsDraining(ctx, "tok3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !draining {
		t.Errorf("expected draining, rate=%.2f", rate)
	}
}

func TestIsDraining_UnknownToken_ReturnsFalse(t *testing.T) {
	d, _ := NewDrainDetector(DefaultDrainConfig())
	draining, _, err := d.IsDraining(context.Background(), "unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draining {
		t.Fatal("expected false for unknown token")
	}
}

func TestRecord_PrunesOldSamples(t *testing.T) {
	cfg := DrainConfig{SampleWindow: 5 * time.Second, DrainThreshold: 5.0, MinSamples: 2}
	d, _ := NewDrainDetector(cfg)
	ctx := context.Background()
	old := time.Now().Add(-10 * time.Second)
	d.mu.Lock()
	d.samples["tok4"] = []sample{{at: old, ttl: 500 * time.Second}}
	d.mu.Unlock()
	d.Record(ctx, "tok4", 300*time.Second)
	d.mu.Lock()
	n := len(d.samples["tok4"])
	d.mu.Unlock()
	if n != 1 {
		t.Errorf("expected old sample pruned, got %d samples", n)
	}
}
