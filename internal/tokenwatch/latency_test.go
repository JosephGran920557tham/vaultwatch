package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultLatencyConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultLatencyConfig()
	if cfg.WarningThreshold <= 0 {
		t.Fatal("expected positive warning threshold")
	}
	if cfg.CriticalThreshold <= cfg.WarningThreshold {
		t.Fatal("critical threshold must exceed warning threshold")
	}
	if cfg.MinSamples <= 0 {
		t.Fatal("expected positive min samples")
	}
}

func TestNewLatencyDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewLatencyDetector(LatencyConfig{})
	def := DefaultLatencyConfig()
	if d.cfg.WarningThreshold != def.WarningThreshold {
		t.Errorf("expected default warning threshold, got %s", d.cfg.WarningThreshold)
	}
	if d.cfg.MinSamples != def.MinSamples {
		t.Errorf("expected default min samples, got %d", d.cfg.MinSamples)
	}
}

func TestLatencyDetector_Check_InsufficientSamples_ReturnsNil(t *testing.T) {
	d := NewLatencyDetector(LatencyConfig{MinSamples: 3})
	d.Record("tok1", 300*time.Millisecond)
	d.Record("tok1", 400*time.Millisecond)
	if got := d.Check("tok1"); got != nil {
		t.Fatalf("expected nil with insufficient samples, got %+v", got)
	}
}

func TestLatencyDetector_Check_LowLatency_ReturnsNil(t *testing.T) {
	d := NewLatencyDetector(LatencyConfig{
		WarningThreshold:  200 * time.Millisecond,
		CriticalThreshold: 500 * time.Millisecond,
		MinSamples:        2,
	})
	d.Record("tok2", 50*time.Millisecond)
	d.Record("tok2", 60*time.Millisecond)
	if got := d.Check("tok2"); got != nil {
		t.Fatalf("expected nil for low latency, got %+v", got)
	}
}

func TestLatencyDetector_Check_Warning(t *testing.T) {
	d := NewLatencyDetector(LatencyConfig{
		WarningThreshold:  200 * time.Millisecond,
		CriticalThreshold: 500 * time.Millisecond,
		MinSamples:        2,
	})
	d.Record("tok3", 250*time.Millisecond)
	d.Record("tok3", 300*time.Millisecond)
	a := d.Check("tok3")
	if a == nil {
		t.Fatal("expected warning alert")
	}
	if a.Level != alert.LevelWarning {
		t.Errorf("expected warning, got %v", a.Level)
	}
}

func TestLatencyDetector_Check_Critical(t *testing.T) {
	d := NewLatencyDetector(LatencyConfig{
		WarningThreshold:  200 * time.Millisecond,
		CriticalThreshold: 500 * time.Millisecond,
		MinSamples:        2,
	})
	d.Record("tok4", 600*time.Millisecond)
	d.Record("tok4", 700*time.Millisecond)
	a := d.Check("tok4")
	if a == nil {
		t.Fatal("expected critical alert")
	}
	if a.Level != alert.LevelCritical {
		t.Errorf("expected critical, got %v", a.Level)
	}
	if a.LeaseID != "tok4" {
		t.Errorf("expected token id tok4, got %s", a.LeaseID)
	}
}
