package tokenwatch

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

func TestDefaultVelocityConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultVelocityConfig()
	if cfg.Window <= 0 {
		t.Error("expected positive Window")
	}
	if cfg.MinSamples <= 0 {
		t.Error("expected positive MinSamples")
	}
	if cfg.DropThreshold <= 0 {
		t.Error("expected positive DropThreshold")
	}
	if cfg.CritThreshold <= cfg.DropThreshold {
		t.Error("expected CritThreshold > DropThreshold")
	}
}

func TestNewVelocityDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewVelocityDetector(VelocityConfig{})
	def := DefaultVelocityConfig()
	if d.cfg.Window != def.Window {
		t.Errorf("expected Window %v, got %v", def.Window, d.cfg.Window)
	}
}

func TestVelocityDetector_Check_InsufficientSamples_ReturnsNil(t *testing.T) {
	d := NewVelocityDetector(VelocityConfig{
		Window: time.Minute, MinSamples: 3,
		DropThreshold: 0.1, CritThreshold: 0.3,
	})
	d.Record("tok1", 30*time.Minute)
	d.Record("tok1", 28*time.Minute)
	if a := d.Check("tok1"); a != nil {
		t.Errorf("expected nil alert with insufficient samples, got %v", a)
	}
}

func TestVelocityDetector_Check_SlowDrop_ReturnsNil(t *testing.T) {
	d := NewVelocityDetector(VelocityConfig{
		Window: 10 * time.Minute, MinSamples: 2,
		DropThreshold: 0.5, CritThreshold: 0.9,
	})
	// Tiny drop — well below threshold
	d.Record("tok2", 60*time.Minute)
	time.Sleep(2 * time.Millisecond)
	d.Record("tok2", 59*time.Minute)
	if a := d.Check("tok2"); a != nil {
		t.Errorf("expected nil for slow drop, got %+v", a)
	}
}

func TestVelocityDetector_Check_FastDrop_ReturnsWarning(t *testing.T) {
	d := NewVelocityDetector(VelocityConfig{
		Window: 10 * time.Minute, MinSamples: 2,
		DropThreshold: 0.001, CritThreshold: 0.9,
	})
	d.Record("tok3", 60*time.Minute)
	time.Sleep(5 * time.Millisecond)
	d.Record("tok3", 30*time.Minute) // 50% drop in ms → very high velocity
	a := d.Check("tok3")
	if a == nil {
		t.Fatal("expected alert for fast drop")
	}
	if a.Level != alert.LevelWarning && a.Level != alert.LevelCritical {
		t.Errorf("expected Warning or Critical, got %v", a.Level)
	}
}

func TestNewVelocityScanner_NilRegistry_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil registry")
		}
	}()
	NewVelocityScanner(nil, NewVelocityDetector(VelocityConfig{}))
}

func TestNewVelocityScanner_NilDetector_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil detector")
		}
	}()
	NewVelocityScanner(NewRegistry(), nil)
}

func TestVelocityScanner_Scan_NoAlerts_WhenSlow(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Add("tokA")
	det := NewVelocityDetector(VelocityConfig{
		Window: time.Minute, MinSamples: 10,
		DropThreshold: 0.5, CritThreshold: 0.9,
	})
	scanner := NewVelocityScanner(reg, det)
	lookup := func(id string) (TokenInfo, error) {
		return TokenInfo{ID: id, TTL: 30 * time.Minute}, nil
	}
	alerts := scanner.Scan(lookup)
	if len(alerts) != 0 {
		t.Errorf("expected no alerts, got %d", len(alerts))
	}
}
