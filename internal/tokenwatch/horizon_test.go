package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultHorizonConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultHorizonConfig()
	if cfg.WarningWindow <= 0 {
		t.Fatal("expected positive WarningWindow")
	}
	if cfg.CriticalWindow <= 0 {
		t.Fatal("expected positive CriticalWindow")
	}
	if cfg.CriticalWindow >= cfg.WarningWindow {
		t.Fatal("CriticalWindow should be shorter than WarningWindow")
	}
}

func TestNewHorizonDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewHorizonDetector(HorizonConfig{})
	def := DefaultHorizonConfig()
	if d.cfg.WarningWindow != def.WarningWindow {
		t.Errorf("got WarningWindow %v, want %v", d.cfg.WarningWindow, def.WarningWindow)
	}
	if d.cfg.CriticalWindow != def.CriticalWindow {
		t.Errorf("got CriticalWindow %v, want %v", d.cfg.CriticalWindow, def.CriticalWindow)
	}
}

func TestHorizonDetector_Check_HighTTL_ReturnsNil(t *testing.T) {
	d := NewHorizonDetector(DefaultHorizonConfig())
	if got := d.Check("tok-1", 24*time.Hour); got != nil {
		t.Errorf("expected nil alert for high TTL, got %+v", got)
	}
}

func TestHorizonDetector_Check_Warning(t *testing.T) {
	cfg := HorizonConfig{WarningWindow: 4 * time.Hour, CriticalWindow: 30 * time.Minute}
	d := NewHorizonDetector(cfg)
	a := d.Check("tok-2", 2*time.Hour)
	if a == nil {
		t.Fatal("expected warning alert, got nil")
	}
	if a.Level != alert.LevelWarning {
		t.Errorf("expected LevelWarning, got %v", a.Level)
	}
}

func TestHorizonDetector_Check_Critical(t *testing.T) {
	cfg := HorizonConfig{WarningWindow: 4 * time.Hour, CriticalWindow: 30 * time.Minute}
	d := NewHorizonDetector(cfg)
	a := d.Check("tok-3", 10*time.Minute)
	if a == nil {
		t.Fatal("expected critical alert, got nil")
	}
	if a.Level != alert.LevelCritical {
		t.Errorf("expected LevelCritical, got %v", a.Level)
	}
}

func TestHorizonDetector_Check_SuppressesDuplicateLevel(t *testing.T) {
	cfg := HorizonConfig{WarningWindow: 4 * time.Hour, CriticalWindow: 30 * time.Minute}
	d := NewHorizonDetector(cfg)
	first := d.Check("tok-4", 2*time.Hour)
	if first == nil {
		t.Fatal("expected first alert")
	}
	second := d.Check("tok-4", 90*time.Minute)
	if second != nil {
		t.Errorf("expected suppressed duplicate, got %+v", second)
	}
}

func TestHorizonDetector_Check_ClearsOnRecovery(t *testing.T) {
	cfg := HorizonConfig{WarningWindow: 4 * time.Hour, CriticalWindow: 30 * time.Minute}
	d := NewHorizonDetector(cfg)
	_ = d.Check("tok-5", 2*time.Hour) // trigger warning
	_ = d.Check("tok-5", 24*time.Hour) // recover — clears seen state
	a := d.Check("tok-5", 2*time.Hour) // should fire again
	if a == nil {
		t.Fatal("expected alert after recovery, got nil")
	}
}
