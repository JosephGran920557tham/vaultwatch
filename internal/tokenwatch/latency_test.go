package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultLatencyConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultLatencyConfig()
	if cfg.MinSamples <= 0 {
		t.Errorf("MinSamples should be positive, got %d", cfg.MinSamples)
	}
	if cfg.WarnThreshold <= 0 {
		t.Errorf("WarnThreshold should be positive, got %v", cfg.WarnThreshold)
	}
	if cfg.CritThreshold <= cfg.WarnThreshold {
		t.Errorf("CritThreshold (%v) should exceed WarnThreshold (%v)", cfg.CritThreshold, cfg.WarnThreshold)
	}
}

func TestNewLatencyDetector_ZeroValues_UsesDefaults(t *testing.T) {
	det := NewLatencyDetector(LatencyConfig{})
	if det == nil {
		t.Fatal("expected non-nil detector")
	}
}

func TestLatencyDetector_Check_InsufficientSamples_ReturnsNil(t *testing.T) {
	cfg := DefaultLatencyConfig()
	cfg.MinSamples = 5
	det := NewLatencyDetector(cfg)
	result := det.Check("tok-1", 500*time.Millisecond)
	if result != nil {
		t.Fatalf("expected nil before MinSamples reached, got %+v", result)
	}
}

func TestLatencyDetector_Check_LowLatency_ReturnsNil(t *testing.T) {
	cfg := DefaultLatencyConfig()
	cfg.MinSamples = 1
	cfg.WarnThreshold = 200 * time.Millisecond
	cfg.CritThreshold = 500 * time.Millisecond
	det := NewLatencyDetector(cfg)
	result := det.Check("tok-ok", 10*time.Millisecond)
	if result != nil {
		t.Fatalf("expected nil for low latency, got %+v", result)
	}
}

func TestLatencyDetector_Check_HighLatency_Warning(t *testing.T) {
	cfg := DefaultLatencyConfig()
	cfg.MinSamples = 1
	cfg.WarnThreshold = 50 * time.Millisecond
	cfg.CritThreshold = 500 * time.Millisecond
	det := NewLatencyDetector(cfg)
	result := det.Check("tok-slow", 100*time.Millisecond)
	if result == nil {
		t.Fatal("expected a warning alert")
	}
	if result.Level != alert.Warning {
		t.Errorf("expected Warning, got %v", result.Level)
	}
}

func TestLatencyDetector_Check_VeryHighLatency_Critical(t *testing.T) {
	cfg := DefaultLatencyConfig()
	cfg.MinSamples = 1
	cfg.WarnThreshold = 50 * time.Millisecond
	cfg.CritThreshold = 200 * time.Millisecond
	det := NewLatencyDetector(cfg)
	result := det.Check("tok-crit", 300*time.Millisecond)
	if result == nil {
		t.Fatal("expected a critical alert")
	}
	if result.Level != alert.Critical {
		t.Errorf("expected Critical, got %v", result.Level)
	}
}
