package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultEntropyConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultEntropyConfig()
	if cfg.MinSamples <= 0 {
		t.Fatal("expected positive MinSamples")
	}
	if cfg.WarningThreshold <= 0 || cfg.WarningThreshold > 1 {
		t.Fatalf("unexpected WarningThreshold: %v", cfg.WarningThreshold)
	}
	if cfg.CriticalThreshold <= 0 || cfg.CriticalThreshold >= cfg.WarningThreshold {
		t.Fatalf("CriticalThreshold should be below WarningThreshold, got %v", cfg.CriticalThreshold)
	}
	if cfg.Window <= 0 {
		t.Fatal("expected positive Window")
	}
}

func TestNewEntropyDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewEntropyDetector(EntropyConfig{})
	def := DefaultEntropyConfig()
	if d.cfg.MinSamples != def.MinSamples {
		t.Errorf("expected MinSamples %d, got %d", def.MinSamples, d.cfg.MinSamples)
	}
}

func TestEntropyDetector_Check_InsufficientSamples_ReturnsNil(t *testing.T) {
	d := NewEntropyDetector(EntropyConfig{MinSamples: 5})
	d.Record("tok1", 30*time.Minute)
	d.Record("tok1", 29*time.Minute)
	if got := d.Check("tok1"); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestEntropyDetector_Check_HighEntropy_ReturnsNil(t *testing.T) {
	d := NewEntropyDetector(DefaultEntropyConfig())
	// Spread samples across many distinct buckets → high entropy
	for i := 0; i < 10; i++ {
		d.Record("tok2", time.Duration(i*5)*time.Minute)
	}
	if got := d.Check("tok2"); got != nil {
		t.Fatalf("expected nil for high entropy, got %+v", got)
	}
}

func TestEntropyDetector_Check_LowEntropy_Warning(t *testing.T) {
	cfg := DefaultEntropyConfig()
	cfg.MinSamples = 3
	cfg.WarningThreshold = 0.99 // almost anything triggers warning
	cfg.CriticalThreshold = 0.01
	d := NewEntropyDetector(cfg)
	// All identical → zero entropy
	for i := 0; i < 5; i++ {
		d.Record("tok3", 30*time.Minute)
	}
	got := d.Check("tok3")
	if got == nil {
		t.Fatal("expected an alert for low entropy")
	}
	if got.Level != alert.LevelWarning && got.Level != alert.LevelCritical {
		t.Errorf("unexpected level: %v", got.Level)
	}
}

func TestEntropyDetector_Check_ZeroEntropy_Critical(t *testing.T) {
	cfg := DefaultEntropyConfig()
	cfg.MinSamples = 3
	cfg.CriticalThreshold = 0.99 // triggers critical for any low entropy
	d := NewEntropyDetector(cfg)
	for i := 0; i < 5; i++ {
		d.Record("tok4", 30*time.Minute)
	}
	got := d.Check("tok4")
	if got == nil {
		t.Fatal("expected a critical alert")
	}
	if got.Level != alert.LevelCritical {
		t.Errorf("expected critical, got %v", got.Level)
	}
}

func TestEntropyDetector_Check_LeaseIDSet(t *testing.T) {
	cfg := DefaultEntropyConfig()
	cfg.MinSamples = 2
	cfg.CriticalThreshold = 0.99
	d := NewEntropyDetector(cfg)
	d.Record("mytoken", 10*time.Minute)
	d.Record("mytoken", 10*time.Minute)
	got := d.Check("mytoken")
	if got == nil {
		t.Fatal("expected alert")
	}
	if got.LeaseID != "mytoken" {
		t.Errorf("expected LeaseID 'mytoken', got %q", got.LeaseID)
	}
}
