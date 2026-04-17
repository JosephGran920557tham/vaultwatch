package tokenwatch

import (
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

func TestDefaultGraceConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultGraceConfig()
	if cfg.WarningBefore <= 0 {
		t.Fatal("expected positive WarningBefore")
	}
	if cfg.CriticalBefore <= 0 {
		t.Fatal("expected positive CriticalBefore")
	}
	if cfg.CriticalBefore >= cfg.WarningBefore {
		t.Fatal("expected CriticalBefore < WarningBefore")
	}
}

func TestNewGraceDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewGraceDetector(GraceConfig{})
	def := DefaultGraceConfig()
	if d.cfg.WarningBefore != def.WarningBefore {
		t.Errorf("expected default WarningBefore")
	}
}

func TestGraceDetector_Check_HighTTL_ReturnsNil(t *testing.T) {
	d := NewGraceDetector(DefaultGraceConfig())
	if a := d.Check("tok1", time.Hour); a != nil {
		t.Errorf("expected nil for high TTL, got %v", a)
	}
}

func TestGraceDetector_Check_Warning(t *testing.T) {
	d := NewGraceDetector(GraceConfig{
		WarningBefore:  10 * time.Minute,
		CriticalBefore: 2 * time.Minute,
	})
	a := d.Check("tok2", 5*time.Minute)
	if a == nil {
		t.Fatal("expected warning alert")
	}
	if a.Level != alert.Warning {
		t.Errorf("expected Warning, got %v", a.Level)
	}
}

func TestGraceDetector_Check_Critical(t *testing.T) {
	d := NewGraceDetector(GraceConfig{
		WarningBefore:  10 * time.Minute,
		CriticalBefore: 2 * time.Minute,
	})
	a := d.Check("tok3", 90*time.Second)
	if a == nil {
		t.Fatal("expected critical alert")
	}
	if a.Level != alert.Critical {
		t.Errorf("expected Critical, got %v", a.Level)
	}
}

func TestGraceDetector_Check_ExactBoundary_Warning(t *testing.T) {
	d := NewGraceDetector(GraceConfig{
		WarningBefore:  10 * time.Minute,
		CriticalBefore: 2 * time.Minute,
	})
	a := d.Check("tok4", 10*time.Minute)
	if a == nil || a.Level != alert.Warning {
		t.Errorf("expected Warning at exact boundary, got %v", a)
	}
}
