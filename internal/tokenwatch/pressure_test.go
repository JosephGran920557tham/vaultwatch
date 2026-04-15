package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultPressureConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultPressureConfig()
	if cfg.Window <= 0 {
		t.Error("expected positive Window")
	}
	if cfg.HighThreshold <= 0 || cfg.HighThreshold > 1 {
		t.Errorf("unexpected HighThreshold: %v", cfg.HighThreshold)
	}
	if cfg.LowThreshold <= 0 || cfg.LowThreshold >= cfg.HighThreshold {
		t.Errorf("unexpected LowThreshold: %v", cfg.LowThreshold)
	}
}

func TestNewPressureDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewPressureDetector(PressureConfig{})
	def := DefaultPressureConfig()
	if d.cfg.Window != def.Window {
		t.Errorf("expected default window %v, got %v", def.Window, d.cfg.Window)
	}
}

func TestPressureDetector_Check_NoSamples_ReturnsNil(t *testing.T) {
	d := NewPressureDetector(DefaultPressureConfig())
	if got := d.Check("tok-1"); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestPressureDetector_Check_LowPressure_ReturnsNil(t *testing.T) {
	d := NewPressureDetector(DefaultPressureConfig())
	for i := 0; i < 10; i++ {
		d.Record("tok-1", i < 2) // 20% pressed — below low threshold
	}
	if got := d.Check("tok-1"); got != nil {
		t.Errorf("expected nil for low pressure, got %v", got)
	}
}

func TestPressureDetector_Check_Warning(t *testing.T) {
	cfg := PressureConfig{
		Window:        time.Minute,
		HighThreshold: 0.75,
		LowThreshold:  0.40,
	}
	d := NewPressureDetector(cfg)
	for i := 0; i < 10; i++ {
		d.Record("tok-2", i < 5) // 50% pressed — between thresholds
	}
	got := d.Check("tok-2")
	if got == nil {
		t.Fatal("expected warning alert, got nil")
	}
	if got.Level != alert.Warning {
		t.Errorf("expected Warning, got %v", got.Level)
	}
}

func TestPressureDetector_Check_Critical(t *testing.T) {
	cfg := PressureConfig{
		Window:        time.Minute,
		HighThreshold: 0.75,
		LowThreshold:  0.40,
	}
	d := NewPressureDetector(cfg)
	for i := 0; i < 10; i++ {
		d.Record("tok-3", i < 9) // 90% pressed — above high threshold
	}
	got := d.Check("tok-3")
	if got == nil {
		t.Fatal("expected critical alert, got nil")
	}
	if got.Level != alert.Critical {
		t.Errorf("expected Critical, got %v", got.Level)
	}
	if got.LeaseID != "tok-3" {
		t.Errorf("expected LeaseID tok-3, got %v", got.LeaseID)
	}
}

func TestPressureDetector_Prune_RemovesOldSamples(t *testing.T) {
	cfg := PressureConfig{
		Window:        50 * time.Millisecond,
		HighThreshold: 0.75,
		LowThreshold:  0.40,
	}
	d := NewPressureDetector(cfg)
	for i := 0; i < 8; i++ {
		d.Record("tok-4", true) // all pressed
	}
	time.Sleep(80 * time.Millisecond)
	// add one non-pressed sample after window expires
	d.Record("tok-4", false)
	got := d.Check("tok-4")
	if got != nil {
		t.Errorf("expected nil after pruning old pressed samples, got %v", got)
	}
}
