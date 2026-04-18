package tokenwatch

import (
	"testing"
	"time"
)

func TestDefaultDecayConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultDecayConfig()
	if cfg.HalfLife <= 0 {
		t.Fatal("expected positive HalfLife")
	}
	if cfg.WarningScore <= 0 || cfg.WarningScore >= 1 {
		t.Fatalf("unexpected WarningScore: %v", cfg.WarningScore)
	}
	if cfg.CriticalScore <= 0 || cfg.CriticalScore >= cfg.WarningScore {
		t.Fatalf("CriticalScore should be below WarningScore, got %v", cfg.CriticalScore)
	}
}

func TestNewDecayDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewDecayDetector(DecayConfig{})
	def := DefaultDecayConfig()
	if d.cfg.HalfLife != def.HalfLife {
		t.Errorf("expected default HalfLife")
	}
}

func TestDecayDetector_Check_NoRecord_ReturnsNil(t *testing.T) {
	d := NewDecayDetector(DefaultDecayConfig())
	if got := d.Check("tok1"); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestDecayDetector_Check_HealthyScore_ReturnsNil(t *testing.T) {
	d := NewDecayDetector(DefaultDecayConfig())
	d.Record("tok1", 50*time.Minute, 60*time.Minute) // score ~0.83
	if got := d.Check("tok1"); got != nil {
		t.Fatalf("expected nil for healthy score, got %v", got)
	}
}

func TestDecayDetector_Check_Warning(t *testing.T) {
	cfg := DefaultDecayConfig()
	d := NewDecayDetector(cfg)
	// score = 0.35 which is below WarningScore(0.4) but above CriticalScore(0.2)
	d.Record("tok1", 35*time.Minute, 100*time.Minute)
	a := d.Check("tok1")
	if a == nil {
		t.Fatal("expected warning alert")
	}
	if a.Level != LevelWarning {
		t.Errorf("expected warning, got %v", a.Level)
	}
}

func TestDecayDetector_Check_Critical(t *testing.T) {
	d := NewDecayDetector(DefaultDecayConfig())
	// score = 0.1 which is below CriticalScore(0.2)
	d.Record("tok1", 10*time.Minute, 100*time.Minute)
	a := d.Check("tok1")
	if a == nil {
		t.Fatal("expected critical alert")
	}
	if a.Level != LevelCritical {
		t.Errorf("expected critical, got %v", a.Level)
	}
}

func TestDecayScannerNewPanicsOnNil(t *testing.T) {
	defer func() { recover() }()
	NewDecayScanner(nil, nil, nil)
	t.Fatal("expected panic")
}
