package tokenwatch

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
)

func TestDefaultSentinelConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultSentinelConfig()
	if cfg.MissWindow <= 0 {
		t.Errorf("expected positive MissWindow, got %v", cfg.MissWindow)
	}
	if cfg.MissThreshold <= 0 {
		t.Errorf("expected positive MissThreshold, got %d", cfg.MissThreshold)
	}
}

func TestNewSentinelDetector_ZeroValues_UsesDefaults(t *testing.T) {
	d := NewSentinelDetector(SentinelConfig{})
	def := DefaultSentinelConfig()
	if d.cfg.MissWindow != def.MissWindow {
		t.Errorf("expected MissWindow %v, got %v", def.MissWindow, d.cfg.MissWindow)
	}
	if d.cfg.MissThreshold != def.MissThreshold {
		t.Errorf("expected MissThreshold %d, got %d", def.MissThreshold, d.cfg.MissThreshold)
	}
}

func TestSentinelDetector_Check_NeverSeen_ReturnsNil(t *testing.T) {
	d := NewSentinelDetector(SentinelConfig{MissWindow: time.Minute, MissThreshold: 2})
	result := d.Check("tok-1", time.Now())
	if result != nil {
		t.Errorf("expected nil on first check, got %v", result)
	}
}

func TestSentinelDetector_Ping_ResetsMisses(t *testing.T) {
	d := NewSentinelDetector(SentinelConfig{MissWindow: time.Millisecond, MissThreshold: 1})
	now := time.Now()
	// accumulate a miss
	d.Check("tok-2", now)
	d.Check("tok-2", now.Add(2*time.Millisecond))
	// ping resets
	d.Ping("tok-2")
	result := d.Check("tok-2", now.Add(3*time.Millisecond))
	if result != nil {
		t.Errorf("expected nil after ping, got %v", result)
	}
}

func TestSentinelDetector_Check_ExceedsThreshold_ReturnsCritical(t *testing.T) {
	d := NewSentinelDetector(SentinelConfig{MissWindow: time.Millisecond, MissThreshold: 2})
	now := time.Now()
	d.Check("tok-3", now)
	d.Check("tok-3", now.Add(2*time.Millisecond))
	result := d.Check("tok-3", now.Add(4*time.Millisecond))
	if result == nil {
		t.Fatal("expected critical alert, got nil")
	}
	if result.Level != alert.Critical {
		t.Errorf("expected Critical, got %v", result.Level)
	}
	if result.LeaseID != "tok-3" {
		t.Errorf("expected leaseID tok-3, got %s", result.LeaseID)
	}
}

func TestSentinelDetector_BelowThreshold_ReturnsNil(t *testing.T) {
	d := NewSentinelDetector(SentinelConfig{MissWindow: time.Millisecond, MissThreshold: 5})
	now := time.Now()
	for i := 0; i < 3; i++ {
		d.Check("tok-4", now.Add(time.Duration(i*2)*time.Millisecond))
	}
	result := d.Check("tok-4", now.Add(8*time.Millisecond))
	if result != nil {
		t.Errorf("expected nil below threshold, got %v", result)
	}
}
