package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultTenureConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultTenureConfig()
	if cfg.WarningAge <= 0 {
		t.Fatal("expected positive WarningAge")
	}
	if cfg.CriticalAge <= cfg.WarningAge {
		t.Fatal("expected CriticalAge > WarningAge")
	}
}

func TestNewTenureDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewTenureDetector(TenureConfig{})
	def := DefaultTenureConfig()
	if d.cfg.WarningAge != def.WarningAge {
		t.Errorf("expected WarningAge %v, got %v", def.WarningAge, d.cfg.WarningAge)
	}
	if d.cfg.CriticalAge != def.CriticalAge {
		t.Errorf("expected CriticalAge %v, got %v", def.CriticalAge, d.cfg.CriticalAge)
	}
}

func TestTenureDetector_Check_YoungToken_ReturnsNil(t *testing.T) {
	d := NewTenureDetector(DefaultTenureConfig())
	now := time.Now()
	d.Track("tok-young", now.Add(-1*time.Hour))
	if a := d.Check("tok-young", now); a != nil {
		t.Fatalf("expected nil alert, got %+v", a)
	}
}

func TestTenureDetector_Check_Warning(t *testing.T) {
	cfg := TenureConfig{
		WarningAge:  2 * time.Hour,
		CriticalAge: 10 * time.Hour,
	}
	d := NewTenureDetector(cfg)
	now := time.Now()
	d.Track("tok-warn", now.Add(-3*time.Hour))
	a := d.Check("tok-warn", now)
	if a == nil {
		t.Fatal("expected warning alert")
	}
	if a.Level != alert.Warning {
		t.Errorf("expected Warning, got %v", a.Level)
	}
}

func TestTenureDetector_Check_Critical(t *testing.T) {
	cfg := TenureConfig{
		WarningAge:  2 * time.Hour,
		CriticalAge: 5 * time.Hour,
	}
	d := NewTenureDetector(cfg)
	now := time.Now()
	d.Track("tok-crit", now.Add(-6*time.Hour))
	a := d.Check("tok-crit", now)
	if a == nil {
		t.Fatal("expected critical alert")
	}
	if a.Level != alert.Critical {
		t.Errorf("expected Critical, got %v", a.Level)
	}
}

func TestTenureDetector_Check_UnknownToken_ReturnsNil(t *testing.T) {
	d := NewTenureDetector(DefaultTenureConfig())
	if a := d.Check("unknown", time.Now()); a != nil {
		t.Fatalf("expected nil for unknown token, got %+v", a)
	}
}

func TestTenureDetector_Track_PreservesOriginalIssueTime(t *testing.T) {
	d := NewTenureDetector(DefaultTenureConfig())
	original := time.Now().Add(-48 * time.Hour)
	d.Track("tok-dup", original)
	d.Track("tok-dup", time.Now()) // should be ignored
	d.mu.Lock()
	got := d.issued["tok-dup"]
	d.mu.Unlock()
	if !got.Equal(original) {
		t.Errorf("expected original issue time to be preserved")
	}
}
