package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultDriftConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultDriftConfig()
	if cfg.WarningThreshold <= 0 {
		t.Fatal("expected positive WarningThreshold")
	}
	if cfg.CriticalThreshold <= cfg.WarningThreshold {
		t.Fatal("expected CriticalThreshold > WarningThreshold")
	}
}

func TestNewDriftDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewDriftDetector(DriftConfig{})
	def := DefaultDriftConfig()
	if d.cfg.WarningThreshold != def.WarningThreshold {
		t.Errorf("want %s, got %s", def.WarningThreshold, d.cfg.WarningThreshold)
	}
	if d.cfg.CriticalThreshold != def.CriticalThreshold {
		t.Errorf("want %s, got %s", def.CriticalThreshold, d.cfg.CriticalThreshold)
	}
}

func TestDriftDetector_Check_NoBaseline_ReturnsNil(t *testing.T) {
	d := NewDriftDetector(DefaultDriftConfig())
	if got := d.Check("tok-1", 5*time.Minute); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestDriftDetector_Check_NoDrift_ReturnsNil(t *testing.T) {
	d := NewDriftDetector(DefaultDriftConfig())
	ttl := 5 * time.Minute
	d.Record("tok-1", ttl)
	// Immediately check — elapsed ≈ 0, so drift ≈ 0.
	if got := d.Check("tok-1", ttl); got != nil {
		t.Fatalf("expected nil for negligible drift, got %+v", got)
	}
}

func TestDriftDetector_Check_Warning(t *testing.T) {
	cfg := DriftConfig{
		WarningThreshold:  5 * time.Second,
		CriticalThreshold: 20 * time.Second,
	}
	d := NewDriftDetector(cfg)

	// Record a baseline of 10 minutes.
	d.Record("tok-2", 10*time.Minute)

	// Simulate drift by reporting a TTL that is 8 seconds lower than expected
	// (expected ≈ 10m since elapsed ≈ 0).
	currentTTL := 10*time.Minute - 8*time.Second
	got := d.Check("tok-2", currentTTL)
	if got == nil {
		t.Fatal("expected warning alert, got nil")
	}
	if got.Level != alert.Warning {
		t.Errorf("want Warning, got %v", got.Level)
	}
}

func TestDriftDetector_Check_Critical(t *testing.T) {
	cfg := DriftConfig{
		WarningThreshold:  5 * time.Second,
		CriticalThreshold: 20 * time.Second,
	}
	d := NewDriftDetector(cfg)

	d.Record("tok-3", 10*time.Minute)

	// Drift of 25 seconds — above critical threshold.
	currentTTL := 10*time.Minute - 25*time.Second
	got := d.Check("tok-3", currentTTL)
	if got == nil {
		t.Fatal("expected critical alert, got nil")
	}
	if got.Level != alert.Critical {
		t.Errorf("want Critical, got %v", got.Level)
	}
}

func TestDriftDetector_Record_OverwritesPrevious(t *testing.T) {
	d := NewDriftDetector(DefaultDriftConfig())
	d.Record("tok-4", 10*time.Minute)
	d.Record("tok-4", 5*time.Minute)

	// After overwrite the baseline is 5m; current is also 5m → no drift.
	if got := d.Check("tok-4", 5*time.Minute); got != nil {
		t.Fatalf("expected nil after overwrite, got %+v", got)
	}
}
