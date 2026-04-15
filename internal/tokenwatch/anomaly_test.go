package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultAnomalyConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultAnomalyConfig()
	if cfg.MinTTL <= 0 {
		t.Fatal("expected positive MinTTL")
	}
	if cfg.MaxTTL <= 0 {
		t.Fatal("expected positive MaxTTL")
	}
	if cfg.MinTTL >= cfg.MaxTTL {
		t.Fatal("expected MinTTL < MaxTTL")
	}
}

func TestNewAnomalyDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewAnomalyDetector(AnomalyConfig{})
	def := DefaultAnomalyConfig()
	if d.cfg.MinTTL != def.MinTTL {
		t.Errorf("MinTTL: got %s, want %s", d.cfg.MinTTL, def.MinTTL)
	}
	if d.cfg.MaxTTL != def.MaxTTL {
		t.Errorf("MaxTTL: got %s, want %s", d.cfg.MaxTTL, def.MaxTTL)
	}
}

func TestAnomalyDetector_Check_NormalTTL_ReturnsNil(t *testing.T) {
	d := NewAnomalyDetector(AnomalyConfig{
		MinTTL: 1 * time.Minute,
		MaxTTL: 24 * time.Hour,
	})
	result := d.Check("tok-abc", 6*time.Hour)
	if result != nil {
		t.Fatalf("expected nil alert, got %+v", result)
	}
}

func TestAnomalyDetector_Check_BelowMin_ReturnsCritical(t *testing.T) {
	d := NewAnomalyDetector(AnomalyConfig{
		MinTTL: 1 * time.Minute,
		MaxTTL: 24 * time.Hour,
	})
	a := d.Check("tok-low", 10*time.Second)
	if a == nil {
		t.Fatal("expected alert, got nil")
	}
	if a.Level != alert.LevelCritical {
		t.Errorf("expected Critical, got %s", a.Level)
	}
	if a.LeaseID != "tok-low" {
		t.Errorf("unexpected LeaseID: %s", a.LeaseID)
	}
}

func TestAnomalyDetector_Check_AboveMax_ReturnsWarning(t *testing.T) {
	d := NewAnomalyDetector(AnomalyConfig{
		MinTTL: 1 * time.Minute,
		MaxTTL: 24 * time.Hour,
	})
	a := d.Check("tok-high", 48*time.Hour)
	if a == nil {
		t.Fatal("expected alert, got nil")
	}
	if a.Level != alert.LevelWarning {
		t.Errorf("expected Warning, got %s", a.Level)
	}
}

func TestAnomalyDetector_Check_ExactMinBoundary_ReturnsNil(t *testing.T) {
	d := NewAnomalyDetector(AnomalyConfig{
		MinTTL: 1 * time.Minute,
		MaxTTL: 24 * time.Hour,
	})
	result := d.Check("tok-boundary", 1*time.Minute)
	if result != nil {
		t.Fatalf("exact min boundary should not alert, got %+v", result)
	}
}
