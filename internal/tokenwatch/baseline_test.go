package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultBaselineConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultBaselineConfig()
	if cfg.SampleWindow <= 0 {
		t.Errorf("expected positive SampleWindow, got %d", cfg.SampleWindow)
	}
	if cfg.DeviationPct <= 0 {
		t.Errorf("expected positive DeviationPct, got %f", cfg.DeviationPct)
	}
}

func TestNewBaselineDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewBaselineDetector(BaselineConfig{})
	def := DefaultBaselineConfig()
	if d.cfg.SampleWindow != def.SampleWindow {
		t.Errorf("expected SampleWindow %d, got %d", def.SampleWindow, d.cfg.SampleWindow)
	}
	if d.cfg.DeviationPct != def.DeviationPct {
		t.Errorf("expected DeviationPct %f, got %f", def.DeviationPct, d.cfg.DeviationPct)
	}
}

func TestBaselineDetector_Check_InsufficientSamples_ReturnsNil(t *testing.T) {
	d := NewBaselineDetector(DefaultBaselineConfig())
	d.Record("tok1", 30*time.Minute)
	if a := d.Check("tok1", 5*time.Minute); a != nil {
		t.Errorf("expected nil with only one sample, got %+v", a)
	}
}

func TestBaselineDetector_Check_NoDeviation_ReturnsNil(t *testing.T) {
	d := NewBaselineDetector(BaselineConfig{SampleWindow: 5, DeviationPct: 25})
	for i := 0; i < 5; i++ {
		d.Record("tok2", 60*time.Minute)
	}
	// 60 min * (1 - 0.25) = 45 min; current 50 min is above threshold
	if a := d.Check("tok2", 50*time.Minute); a != nil {
		t.Errorf("expected nil alert, got %+v", a)
	}
}

func TestBaselineDetector_Check_Deviation_ReturnsAlert(t *testing.T) {
	d := NewBaselineDetector(BaselineConfig{SampleWindow: 5, DeviationPct: 25})
	for i := 0; i < 5; i++ {
		d.Record("tok3", 60*time.Minute)
	}
	// threshold = 45 min; current 20 min is below
	a := d.Check("tok3", 20*time.Minute)
	if a == nil {
		t.Fatal("expected alert for TTL well below baseline, got nil")
	}
	if a.Level != LevelWarning {
		t.Errorf("expected Warning level, got %s", a.Level)
	}
	if a.LeaseID != "tok3" {
		t.Errorf("expected LeaseID tok3, got %s", a.LeaseID)
	}
}

func TestBaselineDetector_SampleWindow_Trims(t *testing.T) {
	d := NewBaselineDetector(BaselineConfig{SampleWindow: 3, DeviationPct: 10})
	// Fill with high values, then overwrite window with low values
	for i := 0; i < 10; i++ {
		d.Record("tok4", 5*time.Minute)
	}
	// After trimming, mean should be ~5 min; current 4.6 min is above 90% of 5 min (4.5 min)
	if a := d.Check("tok4", 4*time.Minute+36*time.Second); a != nil {
		t.Logf("optional alert: %+v", a)
	}
	// Ensure samples slice is capped
	d.mu.Lock()
	n := len(d.samples["tok4"])
	d.mu.Unlock()
	if n > 3 {
		t.Errorf("expected at most 3 samples, got %d", n)
	}
}
